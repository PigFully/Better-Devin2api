// 本文件实现「思考签名降级」：历史 assistant 思考块的签名在上游验签失败时
// （伪造、错配或来自别的供应商体制），整单以 invalid_argument 拒绝，且
// 同一会话每轮都会复现。上游实测接受无签名的思考块，因此在首发被拒后
// 剥掉历史签名（保留思考正文）重发一次。
package devin

import (
	"github.com/WncFht/devin2api/internal/llm"
)

// isInvalidArgumentCause 判定流建立期的失败是否为上游 invalid_argument。
func isInvalidArgumentCause(cause error) bool {
	if cause == nil {
		return false
	}
	failure := llm.Classify(cause)
	return failure != nil && failure.Code == "invalid_argument"
}

// stripHistoryThinkingSignatures 返回去掉历史 assistant 思考块签名后的请求副本
// 与清掉的块数；不改写入参的消息或内容切片。
func stripHistoryThinkingSignatures(request llm.RequestMessages) (llm.RequestMessages, int) {
	stripped := 0
	messages := make([]llm.Message, len(request.Messages))
	for i, message := range request.Messages {
		assistant, ok := message.(llm.AssistantMessage)
		if !ok {
			messages[i] = message
			continue
		}
		content := make([]llm.Content, len(assistant.Content))
		for j, block := range assistant.Content {
			if thinking, isThinking := block.(llm.ThinkingContent); isThinking && thinking.Signature != "" {
				thinking.Signature = ""
				thinking.SignatureType = ""
				stripped++
				content[j] = thinking
				continue
			}
			content[j] = block
		}
		assistant.Content = content
		messages[i] = assistant
	}
	if stripped == 0 {
		return request, 0
	}
	request.Messages = messages
	return request, stripped
}
