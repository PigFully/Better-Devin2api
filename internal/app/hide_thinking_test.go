package app

import (
	"testing"

	"github.com/WncFht/devin2api/internal/api/common"
	"github.com/WncFht/devin2api/internal/llm"
)

type recordingEncoder struct{ types []llm.ResponseEventType }

func (r *recordingEncoder) Encode(event llm.ResponseEvent) ([]common.SSEEvent, error) {
	r.types = append(r.types, event.Type)
	return nil, nil
}

func TestThinkingHidingEncoderDropsThinkingEvents(t *testing.T) {
	inner := &recordingEncoder{}
	encoder := thinkingHidingEncoder{inner: inner}
	events := []llm.ResponseEventType{
		llm.ResponseEventThinkingStart, llm.ResponseEventThinkingDelta,
		llm.ResponseEventThinkingEnd, llm.ResponseEventSignature,
		llm.ResponseEventTextStart, llm.ResponseEventTextDelta, llm.ResponseEventTextEnd,
	}
	for _, typ := range events {
		if _, err := encoder.Encode(llm.ResponseEvent{Type: typ}); err != nil {
			t.Fatal(err)
		}
	}
	want := []llm.ResponseEventType{llm.ResponseEventTextStart, llm.ResponseEventTextDelta, llm.ResponseEventTextEnd}
	if len(inner.types) != len(want) {
		t.Fatalf("forwarded %v, want %v", inner.types, want)
	}
	for i := range want {
		if inner.types[i] != want[i] {
			t.Fatalf("forwarded %v, want %v", inner.types, want)
		}
	}
}

func TestWithoutThinkingKeepsTextAndToolCalls(t *testing.T) {
	original := &llm.AssistantMessage{Content: []llm.Content{
		llm.ThinkingContent{Thinking: "plan", Signature: "sig"},
		llm.TextContent{Text: "4"},
		llm.ThinkingContent{Redacted: true, Signature: "x"},
	}}
	got := withoutThinking(original)
	if len(got.Content) != 1 {
		t.Fatalf("content = %+v, want only text", got.Content)
	}
	if got.Content[0].(llm.TextContent).Text != "4" {
		t.Fatalf("text lost: %+v", got.Content[0])
	}
	if len(original.Content) != 3 {
		t.Fatal("input message was mutated")
	}
	if withoutThinking(nil) != nil {
		t.Fatal("nil must stay nil")
	}
}
