// 本文件实现 Anthropic 响应的「隐藏思考」：上游 Claude 系列默认会输出思考块，
// 开启 server.hide_thinking 后在协议层剔除，客户端只看到文本与工具调用。
// 过滤发生在编码前（流式逐事件、非流式整条消息），编码器本身不感知此开关。
package app

import (
	"github.com/WncFht/devin2api/internal/api/common"
	"github.com/WncFht/devin2api/internal/llm"
)

// thinkingHidingEncoder 包装 Anthropic 流式编码器：丢弃思考相关事件，
// 其余事件原样交给内层编码器。思考块的 content_block 序号留空洞——
// Anthropic 流规范允许 index 不连续，编码器本身已按此处理。
type thinkingHidingEncoder struct {
	inner streamEncoder
}

func (encoder thinkingHidingEncoder) Encode(event llm.ResponseEvent) ([]common.SSEEvent, error) {
	switch event.Type {
	case llm.ResponseEventThinkingStart, llm.ResponseEventThinkingDelta,
		llm.ResponseEventThinkingEnd, llm.ResponseEventSignature:
		return nil, nil
	}
	return encoder.inner.Encode(event)
}

// withoutThinking 返回去掉思考块（含 redacted）的消息副本；不改写入参。
func withoutThinking(message *llm.AssistantMessage) *llm.AssistantMessage {
	if message == nil {
		return nil
	}
	kept := make([]llm.Content, 0, len(message.Content))
	for _, block := range message.Content {
		if _, isThinking := block.(llm.ThinkingContent); isThinking {
			continue
		}
		kept = append(kept, block)
	}
	copied := *message
	copied.Content = kept
	return &copied
}
