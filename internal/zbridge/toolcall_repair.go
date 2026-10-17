// toolcall_repair.go
//
// Tolerant tool-call parsing shared by both agent-mode shims.
//
//   - repairToolJSON / looseDecodeObject: repair JSON the model produced by
//     hand — invalid escapes (raw Windows paths such as "C:\Users\x"), raw
//     control characters inside strings, missing closing quotes/braces at the
//     end of the payload, trailing garbage after the first complete object.
//   - parseGLMNativeToolCall: GLM's native tool-call template, which the
//     model sometimes falls back to instead of the <<<TOOL_CALL>>> contract:
//       <tool_call>name<arg_key>k</arg_key><arg_value>v</arg_value></tool_call>
//     including the observed hybrid
//       <tool_call>bash command="ls"</arg_value><arg_key>timeout</arg_key>...
//   - toolSchemas / normalizeToolArgs: argument values are coerced to the
//     types declared by the client's tool schema (e.g. "30000" -> 30000 for an
//     integer parameter) and ALWAYS re-marshalled as one valid JSON object, so
//     clients never receive unparsable or truncated arguments.

package zbridge

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// ── JSON repair ──────────────────────────────────────────────────────────────

func isHexByte(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// rawWindowsPathRe detects a JSON string body that contains an UNESCAPED
// Windows drive path ("C:\Users" — a single backslash after the drive colon).
var rawWindowsPathRe = regexp.MustCompile(`(?:^|[^A-Za-z0-9])[A-Za-z]:\\(?:[^\\"]|$)`)

// repairJSONEscapes fixes string-literal problems that make otherwise sound
// JSON unparsable:
//   - a backslash not followed by a valid JSON escape is doubled
//     ("C:\Users" -> "C:\\Users");
//   - inside a string that clearly holds a raw Windows path (drive letter
//     followed by a single backslash) every lone backslash is taken literally,
//     so "C:\new\temp" keeps its \n and \t as path separators;
//   - raw control characters (newline, tab, ...) inside strings are escaped.
//
// It returns the input unchanged when nothing needed fixing.
func repairJSONEscapes(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 16)
	inStr := false
	rawPath := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !inStr {
			b.WriteByte(c)
			if c == '"' {
				inStr = true
				rawPath = rawWindowsPathRe.MatchString(stringBodyAhead(s, i+1))
			}
			continue
		}
		switch {
		case c == '\\':
			if i+1 >= len(s) {
				b.WriteString(`\\`)
				continue
			}
			n := s[i+1]
			valid := false
			switch n {
			case '"', '\\':
				valid = true
			case '/', 'b', 'f', 'n', 'r', 't':
				valid = !rawPath
			case 'u':
				valid = !rawPath && i+5 < len(s) && isHexByte(s[i+2]) && isHexByte(s[i+3]) && isHexByte(s[i+4]) && isHexByte(s[i+5])
			}
			if n == '"' && rawPath && closesStringAt(s, i+1) {
				// "C:\dir\" — the backslash is the path's trailing separator
				// and the quote really terminates the string.
				valid = false
			}
			if valid {
				b.WriteByte(c)
				b.WriteByte(n)
				i++
			} else {
				b.WriteString(`\\`)
			}
		case c == '"':
			inStr = false
			b.WriteByte(c)
		case c < 0x20:
			switch c {
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\t':
				b.WriteString(`\t`)
			default:
				b.WriteString(`\u00`)
				b.WriteByte("0123456789abcdef"[c>>4])
				b.WriteByte("0123456789abcdef"[c&0xf])
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// stringBodyAhead returns the raw bytes of the string literal starting at i
// (just after its opening quote) up to the first quote that is not preceded
// by a backslash — a best-effort view used only for heuristics.
func stringBodyAhead(s string, i int) string {
	for j := i; j < len(s); j++ {
		if s[j] == '"' && (j == i || s[j-1] != '\\') {
			return s[i:j]
		}
	}
	return s[i:]
}

// closesStringAt reports whether the quote at s[q] is plausibly the end of a
// string literal: it is followed (after spaces) by ',', '}', ']', ':' or EOF.
func closesStringAt(s string, q int) bool {
	j := q + 1
	for j < len(s) && isASCIISpace(s[j]) {
		j++
	}
	return j >= len(s) || s[j] == ',' || s[j] == '}' || s[j] == ']' || s[j] == ':'
}

// repairJSONClosers completes a JSON value that was cut off at the end:
// an unterminated string is closed, a dangling ',' is dropped and missing
// '}' / ']' are appended in the right order. When the open string ends with
// closer characters ("ls}}") they are treated as the intended closers rather
// than string content. ok is false when the text is not merely truncated
// (e.g. mismatched brackets) or nothing needed fixing.
func repairJSONClosers(s string) (string, bool) {
	var stack []byte
	inStr, esc := false, false
	strStart := -1
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr, strStart = true, i
		case '{':
			stack = append(stack, '}')
		case '[':
			stack = append(stack, ']')
		case '}', ']':
			if len(stack) == 0 || stack[len(stack)-1] != c {
				return s, false
			}
			stack = stack[:len(stack)-1]
		}
	}
	if !inStr && len(stack) == 0 {
		return s, false
	}
	out := s
	if inStr {
		// Peel closer characters the model typed inside the open string.
		body := out[strStart+1:]
		trimmed := strings.TrimRight(body, "}] \t\r\n")
		peeled := body[len(trimmed):]
		if esc { // dangling backslash
			trimmed += `\`
		}
		out = out[:strStart+1] + trimmed + `"`
		// Re-apply the peeled closers that match the stack top-down.
		for _, pc := range strings.TrimSpace(peeled) {
			if pc == ' ' || pc == '\t' || pc == '\r' || pc == '\n' {
				continue
			}
			if len(stack) > 0 && stack[len(stack)-1] == byte(pc) {
				out += string(pc)
				stack = stack[:len(stack)-1]
			}
		}
	}
	out = strings.TrimRight(out, " \t\r\n")
	out = strings.TrimSuffix(out, ",")
	if strings.HasSuffix(out, ":") {
		return s, false // a key without a value: ambiguous
	}
	for i := len(stack) - 1; i >= 0; i-- {
		out += string(stack[i])
	}
	return out, true
}

// decodeFirstObject decodes the first JSON object in s, tolerating trailing
// stray closers / whitespace after it (e.g. one '}' too many).
func decodeFirstObject(s string) (map[string]json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(s), &obj); err == nil {
		return obj, obj != nil
	}
	dec := json.NewDecoder(strings.NewReader(s))
	if err := dec.Decode(&obj); err != nil || obj == nil {
		return nil, false
	}
	rest := strings.Trim(s[dec.InputOffset():], " \t\r\n}]")
	if rest != "" {
		return nil, false
	}
	return obj, true
}

// looseDecodeObject parses s as a JSON object, applying the repairs above in
// increasing order of invasiveness. The first variant that parses wins.
func looseDecodeObject(s string) (map[string]json.RawMessage, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}
	esc := repairJSONEscapes(s)
	if esc != s && rawWindowsPathRe.MatchString(s) {
		// "C:\new\temp" is technically valid JSON but means a path, not
		// a newline and a tab: prefer the repaired reading.
		if obj, ok := decodeFirstObject(esc); ok {
			return obj, true
		}
	}
	if obj, ok := decodeFirstObject(s); ok {
		return obj, true
	}
	if esc != s {
		if obj, ok := decodeFirstObject(esc); ok {
			return obj, true
		}
	}
	if fixed, ok := repairJSONClosers(s); ok {
		if obj, ok := decodeFirstObject(fixed); ok {
			return obj, true
		}
	}
	if fixed, ok := repairJSONClosers(esc); ok {
		if obj, ok := decodeFirstObject(fixed); ok {
			return obj, true
		}
	}
	return nil, false
}

// repairToolJSON returns s as valid, compact JSON if it is (or can be repaired
// into) a JSON object; ok is false otherwise.
func repairToolJSON(s string) (string, bool) {
	obj, ok := looseDecodeObject(s)
	if !ok {
		return "", false
	}
	out, err := marshalNoHTML(obj)
	if err != nil {
		return "", false
	}
	return out, true
}

func marshalNoHTML(v interface{}) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimRight(buf.String(), "\n"), nil
}

// ── tool schemas & argument normalization ────────────────────────────────────

// toolSchemas maps tool name -> parameter name -> JSON-schema type
// ("string", "integer", "number", "boolean", "object", "array").
type toolSchemas map[string]map[string]string

// buildToolSchemas indexes the parameter types of an OpenAI-format tools
// array (nested {type,function:{name,parameters}} or flat definitions).
func buildToolSchemas(toolsRaw json.RawMessage) toolSchemas {
	if len(bytes.TrimSpace(toolsRaw)) == 0 {
		return nil
	}
	var tools []openAITool
	if err := json.Unmarshal(toolsRaw, &tools); err != nil {
		return nil
	}
	out := toolSchemas{}
	for i := range tools {
		name := tools[i].fnName()
		if name == "" {
			continue
		}
		var params struct {
			Properties map[string]struct {
				Type json.RawMessage `json:"type"`
			} `json:"properties"`
		}
		props := map[string]string{}
		if json.Unmarshal(tools[i].fnParameters(), &params) == nil {
			for p, def := range params.Properties {
				props[p] = schemaTypeOf(def.Type)
			}
		}
		out[name] = props
	}
	return out
}

// schemaTypeOf reads a JSON-schema "type" that may be a string or an array
// (["integer","null"]); the first non-null entry wins.
func schemaTypeOf(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var arr []string
	if json.Unmarshal(raw, &arr) == nil {
		for _, t := range arr {
			if t != "null" {
				return t
			}
		}
	}
	return ""
}

// paramType returns the declared type of tool.param ("" when unknown).
func (ts toolSchemas) paramType(tool, param string) string {
	if ts == nil {
		return ""
	}
	return ts[tool][param]
}

// coerceStringValue converts a string argument to the schema type when the
// conversion is lossless and unambiguous; otherwise the string is kept.
// With no schema type, only JSON objects/arrays are decoded.
func coerceStringValue(v, typ string) interface{} {
	t := strings.TrimSpace(v)
	switch typ {
	case "integer":
		if n, err := strconv.ParseInt(t, 10, 64); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(t, 64); err == nil && f == float64(int64(f)) {
			return int64(f)
		}
	case "number":
		if _, err := strconv.ParseFloat(t, 64); err == nil {
			return json.Number(t)
		}
	case "boolean":
		if bv, err := strconv.ParseBool(strings.ToLower(t)); err == nil {
			return bv
		}
	case "object", "array", "":
		if (strings.HasPrefix(t, "{") && typ != "array") || (strings.HasPrefix(t, "[") && typ != "object") {
			var x interface{}
			dec := json.NewDecoder(strings.NewReader(t))
			dec.UseNumber()
			if dec.Decode(&x) == nil && !dec.More() {
				return x
			}
			if typ == "object" {
				if obj, ok := looseDecodeObject(t); ok {
					return obj
				}
			}
		}
	}
	return v
}

// normalizeToolArgs turns whatever the model supplied as "arguments" into
// one valid compact JSON object string, coercing top-level values to the
// tool's schema. Never returns invalid JSON.
func normalizeToolArgs(tool string, raw json.RawMessage, ts toolSchemas) string {
	t := bytes.TrimSpace(raw)
	if len(t) == 0 || bytes.Equal(t, []byte("null")) {
		return "{}"
	}
	if t[0] == '"' {
		var s string
		if json.Unmarshal(t, &s) != nil {
			return "{}"
		}
		obj, ok := looseDecodeObject(s)
		if !ok {
			return "{}"
		}
		b, _ := json.Marshal(obj)
		t = b
	}
	if t[0] != '{' {
		return "{}"
	}
	var m map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(t))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil || m == nil {
		obj, ok := looseDecodeObject(string(t))
		if !ok {
			return "{}"
		}
		b, _ := json.Marshal(obj)
		dec = json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		if dec.Decode(&m) != nil || m == nil {
			return "{}"
		}
	}
	changed := false
	for k, v := range m {
		if s, ok := v.(string); ok {
			if typ := ts.paramType(tool, k); typ != "" && typ != "string" {
				if nv := coerceStringValue(s, typ); !isSameString(nv, s) {
					m[k] = nv
					changed = true
				}
			}
		}
	}
	if !changed {
		// Keep the model's key order when the object is already valid.
		var c bytes.Buffer
		if json.Valid(t) && json.Compact(&c, t) == nil {
			return c.String()
		}
	}
	out, err := marshalNoHTML(m)
	if err != nil {
		return "{}"
	}
	return out
}

// ── GLM-native tool-call format ──────────────────────────────────────────────

const (
	glmToolOpen  = "<tool_call>"
	glmToolClose = "</tool_call>"
	glmKeyOpen   = "<arg_key>"
	glmKeyClose  = "</arg_key>"
	glmValOpen   = "<arg_value>"
	glmValClose  = "</arg_value>"
)

var (
	glmNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.:\-]*`)
	glmAttrRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_\-]*)\s*=\s*`)
)

// glmAttrTerminator reports whether rest (the text right after a closing
// quote) ends an attribute value: end of body, an arg tag, or another attr.
func glmAttrTerminator(rest string) bool {
	r := strings.TrimLeft(rest, " \t\r\n")
	return r == "" || strings.HasPrefix(r, glmValClose) || strings.HasPrefix(r, glmKeyOpen) ||
		strings.HasPrefix(r, glmToolClose) || glmAttrRe.MatchString(r)
}

// parseGLMNativeToolCall parses the body of a <tool_call>…</tool_call> block
// (markers excluded). It accepts
//
//	name<arg_key>k</arg_key><arg_value>v</arg_value>…
//	name k="v" k2=v2 …   (attribute style, optionally mixed with the above
//	                      and with stray </arg_value> tags)
//	{"name":…,"arguments":{…}}   (JSON inside the GLM tags)
//	name {…json args…}
//
// Values are coerced to the schema type of the named tool. The returned args
// are a valid compact JSON object.
func parseGLMNativeToolCall(body string, ts toolSchemas) (name, args string, ok bool) {
	body = strings.TrimSpace(body)
	body = strings.TrimSuffix(body, glmToolClose)
	body = strings.TrimSpace(body)
	if body == "" {
		return "", "", false
	}
	if body[0] == '{' {
		n, a, ok := agentLooseParse(body)
		if !ok || n == "" {
			return "", "", false
		}
		return n, normalizeToolArgs(n, a, ts), true
	}
	name = glmNameRe.FindString(body)
	if name == "" {
		return "", "", false
	}
	rest := body[len(name):]
	vals := map[string]interface{}{}
	set := func(k, v string) {
		k = strings.TrimSpace(k)
		if k == "" {
			return
		}
		typ := ts.paramType(name, k)
		if typ == "string" {
			vals[k] = v
			return
		}
		vals[k] = coerceStringValue(v, typ)
	}
	for {
		rest = strings.TrimLeft(rest, " \t\r\n")
		if rest == "" {
			break
		}
		switch {
		case strings.HasPrefix(rest, glmValClose):
			rest = rest[len(glmValClose):]
		case strings.HasPrefix(rest, glmKeyOpen):
			rest = rest[len(glmKeyOpen):]
			ke := strings.Index(rest, glmKeyClose)
			if ke < 0 {
				return "", "", false
			}
			key := rest[:ke]
			rest = strings.TrimLeft(rest[ke+len(glmKeyClose):], " \t\r\n")
			val := ""
			if strings.HasPrefix(rest, glmValOpen) {
				rest = rest[len(glmValOpen):]
				ve := strings.Index(rest, glmValClose)
				if ve < 0 {
					// Unterminated value: take up to the next key or the end.
					if nk := strings.Index(rest, glmKeyOpen); nk >= 0 {
						val, rest = rest[:nk], rest[nk:]
					} else {
						val, rest = rest, ""
					}
				} else {
					val, rest = rest[:ve], rest[ve+len(glmValClose):]
				}
			}
			set(key, val)
		case rest[0] == '{':
			obj, ok := looseDecodeObject(rest)
			if !ok {
				return "", "", false
			}
			for k, v := range obj {
				var x interface{}
				dec := json.NewDecoder(bytes.NewReader(v))
				dec.UseNumber()
				if dec.Decode(&x) == nil {
					if s, isStr := x.(string); isStr {
						set(k, s)
					} else {
						vals[k] = x
					}
				}
			}
			rest = ""
		default:
			m := glmAttrRe.FindStringSubmatch(rest)
			if m == nil {
				return "", "", false
			}
			key := m[1]
			rest = rest[len(m[0]):]
			var val string
			if rest != "" && (rest[0] == '"' || rest[0] == '\'') {
				q := rest[0]
				found := -1
				for j := 1; j < len(rest); j++ {
					if rest[j] == q && glmAttrTerminator(rest[j+1:]) {
						found = j
						break
					}
				}
				if found < 0 {
					// Unterminated quote: value runs to the next tag / end.
					v := rest[1:]
					if e := strings.Index(v, "</arg_"); e >= 0 {
						v = v[:e]
					} else if e := strings.Index(v, glmKeyOpen); e >= 0 {
						v = v[:e]
					}
					val = strings.TrimSuffix(strings.TrimRight(v, " \t\r\n"), string(q))
					rest = rest[1+len(v):]
				} else {
					val, rest = rest[1:found], rest[found+1:]
				}
			} else {
				e := strings.IndexAny(rest, " \t\r\n<")
				if e < 0 {
					e = len(rest)
				}
				val, rest = rest[:e], rest[e:]
			}
			set(key, val)
		}
	}
	out, err := marshalNoHTML(vals)
	if err != nil {
		return "", "", false
	}
	return name, out, true
}

func isSameString(v interface{}, s string) bool {
	vs, ok := v.(string)
	return ok && vs == s
}
