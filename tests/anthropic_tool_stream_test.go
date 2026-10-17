package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Godde3s/glm-free-api/internal/zbridge"
)

// TestAnthropicStreamToolUseNotInterruptedByReasoning reproduces a tool call
// whose JSON is split by a reasoning (<details>) chunk, plus reasoning after
// the call. The client must receive every input_json_delta on the tool_use
// block's own index and the concatenated input must be complete, valid JSON.
// Previously the reasoning chunk closed the open tool_use block, and the
// remaining argument fragments were sent with the thinking block's index
// (tool input truncated to {"command":"ls","timeout":50).
func TestAnthropicStreamToolUseNotInterruptedByReasoning(t *testing.T) {
	for _, variant := range []string{"legacy", "modern"} {
		t.Run(variant, func(t *testing.T) {
			testAnthropicToolStream(t, variant)
		})
	}
}

func testAnthropicToolStream(t *testing.T, variant string) {
	events := []string{
		`{"data":{"delta_content":"<<<TOOL_CALL>>>\n{\"name\":\"bash\",\"arguments\":{\"command\":\"dir C:\\Users\\me\\Temp\",\"timeout\":50"}}`,
		`{"data":{"delta_content":"<details>still thinking</details>"}}`,
		`{"data":{"delta_content":"00}}\n<<<END_TOOL_CALL>>>"}}`,
		`{"data":{"delta_content":"<details>more thoughts</details>"}}`,
		`{"data":{"phase":"done"}}`,
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		for _, e := range events {
			fmt.Fprintf(w, "data: %s\n\n", e)
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer upstream.Close()

	oldBase := zbridge.BASE_URL
	zbridge.BASE_URL = upstream.URL
	defer func() { zbridge.BASE_URL = oldBase }()
	defer zbridge.OverrideSessionState("test-token", "test-user", true)()

	cfg := zbridge.GetConfig()
	oldMode, oldVariant := cfg.AgentMode, cfg.AgentModeVariant
	cfg.AgentMode, cfg.AgentModeVariant = true, variant
	defer func() { cfg.AgentMode, cfg.AgentModeVariant = oldMode, oldVariant }()
	zbridge.SeedCaptchaParam("test-captcha-param")

	body := `{"model":"glm-4.7","stream":true,"max_tokens":256,
	  "messages":[{"role":"user","content":"list temp"}],
	  "tools":[{"name":"bash","input_schema":{"type":"object","properties":{
	    "command":{"type":"string"},"timeout":{"type":"number"}}}}]}`
	req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.Auth.Token)
	rec := httptest.NewRecorder()
	zbridge.NewHandler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	blockType := map[int]string{}
	toolInput := map[int]string{}
	stopReason := ""
	for _, line := range strings.Split(rec.Body.String(), "\n") {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var ev struct {
			Type         string `json:"type"`
			Index        int    `json:"index"`
			ContentBlock struct {
				Type string `json:"type"`
			} `json:"content_block"`
			Delta struct {
				Type        string `json:"type"`
				PartialJSON string `json:"partial_json"`
				StopReason  string `json:"stop_reason"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(line[len("data: "):]), &ev); err != nil {
			t.Fatalf("bad event %q: %v", line, err)
		}
		switch ev.Type {
		case "content_block_start":
			blockType[ev.Index] = ev.ContentBlock.Type
		case "content_block_delta":
			if ev.Delta.Type == "input_json_delta" {
				if blockType[ev.Index] != "tool_use" {
					t.Errorf("input_json_delta sent to %q block %d", blockType[ev.Index], ev.Index)
				}
				toolInput[ev.Index] += ev.Delta.PartialJSON
			} else if blockType[ev.Index] == "tool_use" {
				t.Errorf("%s delta sent to tool_use block %d", ev.Delta.Type, ev.Index)
			}
		case "message_delta":
			stopReason = ev.Delta.StopReason
		}
	}

	if len(toolInput) != 1 {
		t.Fatalf("got %d tool_use inputs, want 1: %v\n%s", len(toolInput), toolInput, rec.Body.String())
	}
	for _, in := range toolInput {
		var args struct {
			Command string  `json:"command"`
			Timeout float64 `json:"timeout"`
		}
		if err := json.Unmarshal([]byte(in), &args); err != nil {
			t.Fatalf("tool input is not valid JSON: %q: %v", in, err)
		}
		if args.Command != `dir C:\Users\me\Temp` || args.Timeout != 5000 {
			t.Errorf("tool input = %+v", args)
		}
	}
	if stopReason != "tool_use" {
		t.Errorf("stop_reason = %q, want tool_use", stopReason)
	}
}
