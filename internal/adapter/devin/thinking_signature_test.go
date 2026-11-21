package devin

import (
	"errors"
	"testing"

	"github.com/WncFht/devin2api/internal/llm"
)

func TestStripHistoryThinkingSignaturesKeepsTextAndDoesNotMutateInput(t *testing.T) {
	original := []llm.Content{
		llm.ThinkingContent{Thinking: "plan", Signature: "forged-sig", SignatureType: "anthropic"},
		llm.TextContent{Text: "4"},
	}
	request := llm.RequestMessages{Messages: []llm.Message{
		llm.UserMessage{Content: []llm.Content{llm.TextContent{Text: "2+2?"}}},
		llm.AssistantMessage{Content: original},
	}}
	got, n := stripHistoryThinkingSignatures(request)
	if n != 1 {
		t.Fatalf("stripped = %d, want 1", n)
	}
	assistant := got.Messages[1].(llm.AssistantMessage)
	thinking := assistant.Content[0].(llm.ThinkingContent)
	if thinking.Thinking != "plan" || thinking.Signature != "" || thinking.SignatureType != "" {
		t.Fatalf("thinking after strip = %+v", thinking)
	}
	if assistant.Content[1].(llm.TextContent).Text != "4" {
		t.Fatalf("text block changed: %+v", assistant.Content[1])
	}
	// 入参不得被改写
	if original[0].(llm.ThinkingContent).Signature != "forged-sig" {
		t.Fatal("input signature was mutated")
	}
}

func TestStripHistoryThinkingSignaturesNoSignatureIsNoop(t *testing.T) {
	request := llm.RequestMessages{Messages: []llm.Message{
		llm.AssistantMessage{Content: []llm.Content{llm.ThinkingContent{Thinking: "x"}}},
	}}
	if _, n := stripHistoryThinkingSignatures(request); n != 0 {
		t.Fatalf("stripped = %d, want 0", n)
	}
}

func TestIsInvalidArgumentCause(t *testing.T) {
	if isInvalidArgumentCause(nil) {
		t.Fatal("nil cause must not match")
	}
	if isInvalidArgumentCause(errors.New("connection reset")) {
		t.Fatal("transport error must not match")
	}
	if !isInvalidArgumentCause(&llm.Failure{Code: "invalid_argument", Message: "x"}) {
		t.Fatal("invalid_argument failure must match")
	}
}
