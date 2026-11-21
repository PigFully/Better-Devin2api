// 本文件处理 Claude 系列模型的 effort 档位：上游目录把推理强度编码进 uid
// 后缀（claude-opus-5-high、claude-opus-5-high-fast），对外统一为标准名
// claude-opus-5，档位由客户端请求携带的 effort 决定。

package adapter

import "strings"

// ClaudeEffortLevels 按强度从低到高排列，路由回退时据此向上/向下查找。
var ClaudeEffortLevels = []string{"low", "medium", "high", "xhigh", "max"}

// DefaultClaudeEffort 是客户端未声明 effort 时使用的档位，与 Anthropic 默认一致。
const DefaultClaudeEffort = "high"

// NormalizeClaudeEffort 规范化客户端声明的 effort；空串或未知档位返回 ok=false。
func NormalizeClaudeEffort(effort string) (level string, ok bool) {
	level = strings.ToLower(strings.TrimSpace(effort))
	for _, known := range ClaudeEffortLevels {
		if level == known {
			return level, true
		}
	}
	return "", false
}

// ClaudeEffortIndex 返回档位在 ClaudeEffortLevels 中的下标；未知返回 -1。
func ClaudeEffortIndex(level string) int {
	for i, known := range ClaudeEffortLevels {
		if level == known {
			return i
		}
	}
	return -1
}

// IsClaudeModel 判断名字是否属于 Claude 系列（claude-* 前缀）。
func IsClaudeModel(name string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(name)), "claude-")
}

// IsFastVariant 判断模型名是否为 -fast 变体；此类变体对外屏蔽、不可调用。
func IsFastVariant(name string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(name)), "-fast")
}

// IsBlockedModel 判断模型名是否对外屏蔽、不可调用：所有 -fast 变体，以及
// Claude 系列的 -1m（1M 上下文）与 -thinking（思考模式）变体。
func IsBlockedModel(name string) bool {
	if IsFastVariant(name) {
		return true
	}
	lower := strings.ToLower(strings.TrimSpace(name))
	return IsClaudeModel(lower) && (strings.HasSuffix(lower, "-1m") || strings.HasSuffix(lower, "-thinking"))
}

// SplitClaudeEffort 若 uid 以合法 effort 档位结尾（且为 Claude 系列），返回去掉
// 档位后的基础名与档位；否则 ok=false。
func SplitClaudeEffort(uid string) (base, level string, ok bool) {
	if !IsClaudeModel(uid) {
		return "", "", false
	}
	i := strings.LastIndex(uid, "-")
	if i <= 0 {
		return "", "", false
	}
	level, ok = NormalizeClaudeEffort(uid[i+1:])
	if !ok {
		return "", "", false
	}
	return uid[:i], level, true
}

// claudeBaseNames 是上游基础名到对外标准名的改名表：上游 Fable 系列以
// claude-5-fable-<档位> 命名，对外统一为 claude-fable-5-<档位>。
var claudeBaseNames = map[string]string{"claude-5-fable": "claude-fable-5"}

// PublicClaudeName 把上游基础名映射为对外标准名；无需改名则原样返回。
func PublicClaudeName(upstream string) string {
	if pub, ok := claudeBaseNames[upstream]; ok {
		return pub
	}
	return upstream
}

// UpstreamClaudeName 是 PublicClaudeName 的反向映射：对外标准名 → 上游基础名。
// 旧的上游名原样保留可用（兼容已有客户端）。
func UpstreamClaudeName(public string) string {
	for upstream, pub := range claudeBaseNames {
		if pub == public {
			return upstream
		}
	}
	return public
}

// PublicModels 把上游目录投影为对外的模型列表：
//   - 屏蔽 -fast 变体，以及 Claude 的 -1m / -thinking 变体（见 IsBlockedModel）；
//   - Claude 的带档位 uid 折叠为标准名（claude-opus-5-high → claude-opus-5），
//     同名只保留第一条。
//
// 只改对外投影，内部目录（上游路由校验依据）保持不变。
func PublicModels(models []ModelInfo) []ModelInfo {
	out := make([]ModelInfo, 0, len(models))
	seen := make(map[string]bool, len(models))
	for _, m := range models {
		if IsBlockedModel(m.ID) {
			continue
		}
		if base, _, ok := SplitClaudeEffort(m.ID); ok {
			m.ID = base
		}
		m.ID = PublicClaudeName(m.ID)
		if seen[m.ID] {
			continue
		}
		seen[m.ID] = true
		out = append(out, m)
	}
	return out
}
