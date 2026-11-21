package messages

import "testing"

// TestDecodeRequestCapturesOutputConfigEffort 验证 output_config.effort 进入
// 中间请求的 Effort 字段，且不再被记为未消费（丢弃）字段。
func TestDecodeRequestCapturesOutputConfigEffort(t *testing.T) {
	data := []byte(`{"model":"claude-opus-5","max_tokens":16,"messages":[{"role":"user","content":"hi"}],"output_config":{"effort":"low"}}`)
	adapted, err := DecodeRequest(data, true)
	if err != nil {
		t.Fatal(err)
	}
	if adapted.Context.Effort != "low" {
		t.Fatalf("Effort = %q, want low", adapted.Context.Effort)
	}
	for _, dropped := range adapted.Context.Dropped {
		if dropped == "field:output_config" {
			t.Fatalf("output_config reported as dropped: %v", adapted.Context.Dropped)
		}
	}
}

// TestDecodeRequestToleratesMalformedOutputConfig 验证非对象的 output_config
// 不会让整个请求解码失败，只是没有 effort。
func TestDecodeRequestToleratesMalformedOutputConfig(t *testing.T) {
	data := []byte(`{"model":"claude-opus-5","max_tokens":16,"messages":[{"role":"user","content":"hi"}],"output_config":"high"}`)
	adapted, err := DecodeRequest(data, false)
	if err != nil {
		t.Fatal(err)
	}
	if adapted.Context.Effort != "" {
		t.Fatalf("Effort = %q, want empty", adapted.Context.Effort)
	}
}
