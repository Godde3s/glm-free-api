package zbridge

import (
	"encoding/json"
	"strings"
	"testing"
)

const testToolsJSON = `[
 {"type":"function","function":{"name":"bash","parameters":{"type":"object","properties":{
   "command":{"type":"string"},"timeout":{"type":"number"},"description":{"type":"string"}}}}},
 {"type":"function","function":{"name":"read","parameters":{"type":"object","properties":{
   "filePath":{"type":"string"},"offset":{"type":"integer"},"limit":{"type":["integer","null"]}}}}}
]`

func testSchemas(t *testing.T) toolSchemas {
	t.Helper()
	ts := buildToolSchemas(json.RawMessage(testToolsJSON))
	if ts.paramType("bash", "timeout") != "number" || ts.paramType("read", "limit") != "integer" {
		t.Fatalf("schema index wrong: %#v", ts)
	}
	return ts
}

// argsMap decodes an arguments string, failing the test if it is not a
// valid JSON object.
func argsMap(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("arguments are not valid JSON: %q: %v", s, err)
	}
	return m
}

func TestRepairJSONEscapesWindowsPaths(t *testing.T) {
	cases := map[string]string{
		`{"p":"C:\Users\alice\AppData\Local\Temp\x"}`: `C:\Users\alice\AppData\Local\Temp\x`,
		// \n and \t would be valid escapes, but inside a raw drive path
		// they are separators.
		`{"p":"C:\new\temp\file.txt"}`: `C:\new\temp\file.txt`,
		`{"p":"D:\dir\"}`:              `D:\dir\`,
		// Non-path strings keep their valid escapes.
		`{"p":"a\nb \d+"}`: "a\nb \\d+",
		// Raw control characters are escaped.
		"{\"p\":\"line1\nline2\"}": "line1\nline2",
	}
	for in, want := range cases {
		obj, ok := looseDecodeObject(in)
		if !ok {
			t.Errorf("looseDecodeObject(%q) failed", in)
			continue
		}
		var got string
		json.Unmarshal(obj["p"], &got)
		if got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
	// Already-valid JSON is untouched.
	valid := `{"p":"C:\\Users\\x","q":"a\nb"}`
	if got := repairJSONEscapes(valid); got != valid {
		t.Errorf("valid JSON modified: %q", got)
	}
}

func TestRepairJSONClosers(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"name":"bash","arguments":{"command":"ls"}`, `{"arguments":{"command":"ls"},"name":"bash"}`},
		{`{"name":"bash","arguments":{"command":"ls","timeout":5000`, `{"arguments":{"command":"ls","timeout":5000},"name":"bash"}`},
		{`{"name":"bash","arguments":{"command":"ls}}`, `{"arguments":{"command":"ls"},"name":"bash"}`},
		{`{"name":"bash","arguments":{"command":"ls"},`, `{"arguments":{"command":"ls"},"name":"bash"}`},
		{`{"name":"bash","arguments":{"command":"ls"}}}`, `{"arguments":{"command":"ls"},"name":"bash"}`},
	}
	for _, c := range cases {
		got, ok := repairToolJSON(c.in)
		if !ok || got != c.want {
			t.Errorf("repairToolJSON(%q) = %q, %v; want %q", c.in, got, ok, c.want)
		}
	}
	for _, bad := range []string{`{"a":`, `{"a":1]`, `not json`} {
		if got, ok := repairToolJSON(bad); ok {
			t.Errorf("repairToolJSON(%q) = %q, want failure", bad, got)
		}
	}
}

func TestParseGLMNativeToolCall(t *testing.T) {
	ts := testSchemas(t)
	cases := []struct {
		body string
		name string
		want map[string]interface{}
	}{
		{
			"bash<arg_key>command</arg_key><arg_value>ls -la</arg_value><arg_key>timeout</arg_key><arg_value>30000</arg_value>",
			"bash", map[string]interface{}{"command": "ls -la", "timeout": 30000.0},
		},
		{
			"bash\n<arg_key>command</arg_key>\n<arg_value>echo \"hi\"</arg_value>\n",
			"bash", map[string]interface{}{"command": `echo "hi"`},
		},
		{
			// Hybrid attribute form observed from GLM.
			`bash command="curl.exe -s http://localhost:8080/x"</arg_value><arg_key>timeout</arg_key><arg_value>30000</arg_value>`,
			"bash", map[string]interface{}{"command": "curl.exe -s http://localhost:8080/x", "timeout": 30000.0},
		},
		{
			`read filePath="C:\Users\me\a.kt" offset=75 limit=16`,
			"read", map[string]interface{}{"filePath": `C:\Users\me\a.kt`, "offset": 75.0, "limit": 16.0},
		},
		{
			`{"name":"bash","arguments":{"command":"pwd"}}`,
			"bash", map[string]interface{}{"command": "pwd"},
		},
	}
	for _, c := range cases {
		name, args, ok := parseGLMNativeToolCall(c.body, ts)
		if !ok || name != c.name {
			t.Errorf("parse(%q) = %q, %v", c.body, name, ok)
			continue
		}
		got := argsMap(t, args)
		if len(got) != len(c.want) {
			t.Errorf("parse(%q) args = %v, want %v", c.body, got, c.want)
		}
		for k, v := range c.want {
			if got[k] != v {
				t.Errorf("parse(%q) %s = %#v, want %#v", c.body, k, got[k], v)
			}
		}
	}
	for _, bad := range []string{"", "   ", "!!", "bash some prose here"} {
		if _, _, ok := parseGLMNativeToolCall(bad, ts); ok {
			t.Errorf("parse(%q) unexpectedly succeeded", bad)
		}
	}
}

func TestNormalizeToolArgs(t *testing.T) {
	ts := testSchemas(t)
	cases := []struct{ tool, in, want string }{
		{"bash", `{"command":"ls","timeout":"5000"}`, `{"command":"ls","timeout":5000}`},
		{"bash", `{"command":"ls","timeout":5000}`, `{"command":"ls","timeout":5000}`}, // order kept
		{"bash", `"{\"command\":\"ls\"}"`, `{"command":"ls"}`},
		{"bash", `null`, `{}`},
		{"bash", `"not json"`, `{}`},
		{"read", `{"filePath":"a","offset":"75","limit":"x"}`, `{"filePath":"a","limit":"x","offset":75}`},
		{"unknown", `{"a":"1"}`, `{"a":"1"}`},
	}
	for _, c := range cases {
		if got := normalizeToolArgs(c.tool, json.RawMessage(c.in), ts); got != c.want {
			t.Errorf("normalizeToolArgs(%s, %s) = %s, want %s", c.tool, c.in, got, c.want)
		}
	}
}

// TestInterceptorToleratesModelFormats feeds each observed model output
// through the streaming interceptor in small chunks and checks that exactly
// one tool call with valid, complete arguments comes out and no protocol
// text leaks.
func TestInterceptorToleratesModelFormats(t *testing.T) {
	cases := []struct {
		name, stream, tool, key string
		want                    interface{}
	}{
		{"canonical", "<<<TOOL_CALL>>>\n{\"name\":\"bash\",\"arguments\":{\"command\":\"ls\",\"timeout\":5000}}\n<<<END_TOOL_CALL>>>", "bash", "timeout", 5000.0},
		{"missing brace", "Listing.\n<<<TOOL_CALL>>>\n{\"name\":\"bash\",\"arguments\":{\"command\":\"ls\"}\n<<<END_TOOL_CALL>>>", "bash", "command", "ls"},
		{"windows path", "<<<TOOL_CALL>>>{\"name\":\"read\",\"arguments\":{\"filePath\":\"C:\\Users\\alice\\AppData\\Local\\Temp\\x\"}}<<<END_TOOL_CALL>>>", "read", "filePath", `C:\Users\alice\AppData\Local\Temp\x`},
		{"glm native", "I'll run it.\n<tool_call>bash<arg_key>command</arg_key><arg_value>ls</arg_value><arg_key>timeout</arg_key><arg_value>30000</arg_value></tool_call>", "bash", "timeout", 30000.0},
		{"glm hybrid", "<tool_call>bash command=\"curl.exe -s http://x\"</arg_value><arg_key>timeout</arg_key><arg_value>30000</arg_value></tool_call>", "bash", "command", "curl.exe -s http://x"},
		{"unterminated at EOF", "<<<TOOL_CALL>>>\n{\"name\":\"bash\",\"arguments\":{\"command\":\"ls\"}}", "bash", "command", "ls"},
		{"glm in contract markers", "<<<TOOL_CALL>>>bash<arg_key>command</arg_key><arg_value>pwd</arg_value><<<END_TOOL_CALL>>>", "bash", "command", "pwd"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := newAgentInterceptorWith(json.RawMessage(testToolsJSON))
			var content string
			var calls []map[string]interface{}
			for i := 0; i < len(c.stream); i += 3 {
				end := min(i+3, len(c.stream))
				txt, tcs := in.feed(c.stream[i:end])
				content += txt
				calls = append(calls, tcs...)
			}
			txt, tcs := in.finish()
			content += txt
			calls = append(calls, tcs...)

			if len(calls) != 1 {
				t.Fatalf("%d calls, want 1 (content %q)", len(calls), content)
			}
			fn := calls[0]["function"].(map[string]interface{})
			if fn["name"] != c.tool {
				t.Errorf("name = %v, want %s", fn["name"], c.tool)
			}
			args := argsMap(t, fn["arguments"].(string))
			if args[c.key] != c.want {
				t.Errorf("%s = %#v, want %#v (args %v)", c.key, args[c.key], c.want, args)
			}
			for _, leak := range []string{"TOOL_CALL", "tool_call", "arg_key", "arg_value"} {
				if strings.Contains(content, leak) {
					t.Errorf("protocol text leaked: %q", content)
				}
			}
		})
	}
}

// Prose that merely mentions the GLM tag must stay text.
func TestInterceptorIgnoresTagMentions(t *testing.T) {
	in := newAgentInterceptorWith(json.RawMessage(testToolsJSON))
	text := "GLM wraps calls in <tool_call> tags"
	got, calls := in.feed(text)
	rest, more := in.finish()
	if len(calls)+len(more) != 0 {
		t.Fatalf("unexpected tool call from prose: %v %v", calls, more)
	}
	if got+rest != text {
		t.Errorf("content = %q, want %q", got+rest, text)
	}
}

func TestExtractAndStripGLMNative(t *testing.T) {
	text := "Checking.\n<tool_call>bash<arg_key>command</arg_key><arg_value>ls</arg_value></tool_call>"
	calls := agentExtractToolCallsWith(text, json.RawMessage(testToolsJSON))
	if len(calls) != 1 {
		t.Fatalf("%d calls, want 1", len(calls))
	}
	if got := agentStripToolCallsWith(text, json.RawMessage(testToolsJSON)); got != "Checking." {
		t.Errorf("strip = %q", got)
	}
}
