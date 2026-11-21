// 本文件实现 Claude 系列的 effort 路由：客户端用标准名（claude-opus-5）加
// 请求里的 effort 指定模型与推理强度，这里把它映射为上游带档位的 uid
// （claude-opus-5-high）。-fast 变体对所有模型屏蔽。
package devin

import (
	"fmt"
	"strings"

	"connectrpc.com/connect"

	"github.com/WncFht/devin2api/internal/adapter"
)

func isClaudeModelName(name string) bool { return adapter.IsClaudeModel(name) }

func isBlockedModelName(name string) bool { return adapter.IsBlockedModel(name) }

// stripLongContextSuffix 去掉 Claude 模型名上的 [1m] 声明后缀（CC Switch /
// Claude Code 的写法，如 claude-opus-4-6[1m]），返回基础名与是否声明了 1M。
func stripLongContextSuffix(model string) (string, bool) {
	if strings.HasSuffix(strings.ToLower(model), "[1m]") {
		return strings.TrimSpace(model[:len(model)-len("[1m]")]), true
	}
	return model, false
}

// hasExactAlias 判断模型名是否被显式配置为别名键；显式别名优先于标准名路由。
func hasExactAlias(aliases map[string]string, model string) bool {
	_, ok := aliases[model]
	return ok
}

func notAvailableError(model string) error {
	return connect.NewError(connect.CodeNotFound, fmt.Errorf("model %q is not available", model))
}

// catalogUIDSet 返回当前目录的 uid 集合；目录尚未加载时返回 nil。
func (a *Adapter) catalogUIDSet() map[string]bool {
	a.modelsMu.RLock()
	defer a.modelsMu.RUnlock()
	if a.models == nil {
		return nil
	}
	set := make(map[string]bool, len(a.models))
	for _, m := range a.models {
		set[m.ID] = true
	}
	return set
}

// routeClaudeEffort 把 Claude 标准名与 effort 映射为上游 uid。longContext 为
// true 时（客户端声明 1M）优先选 <基础名>-1m 变体；目录里没有该变体则忽略声明。
//   - -fast 变体直接拒绝；
//   - 已带档位后缀的旧式 uid 原样透传（兼容旧客户端，此时忽略 effort）；
//   - 否则取请求 effort（缺省 high），优先选同档位 uid，缺失时向下、再向上
//     找最近的可用档位；目录中没有任何档位但存在无档位基础名（如
//     claude-opus-4-6）时使用基础名；都没有则交给上游裁决。
//
// catalog 为 nil 表示目录尚未加载，直接拼接档位返回。
func routeClaudeEffort(model, effort string, longContext bool, catalog map[string]bool) (string, error) {
	if isBlockedModelName(model) {
		return "", notAvailableError(model)
	}
	model = adapter.UpstreamClaudeName(model)
	if _, _, tiered := adapter.SplitClaudeEffort(model); tiered {
		return model, nil
	}
	if longContext && catalog != nil && catalog[model+"-1m"] {
		return model + "-1m", nil
	}
	level, ok := adapter.NormalizeClaudeEffort(effort)
	if !ok {
		level = adapter.DefaultClaudeEffort
	}
	if catalog == nil {
		return model + "-" + level, nil
	}
	levels := adapter.ClaudeEffortLevels
	idx := adapter.ClaudeEffortIndex(level)
	for i := idx; i >= 0; i-- {
		if candidate := model + "-" + levels[i]; catalog[candidate] {
			return candidate, nil
		}
	}
	for i := idx + 1; i < len(levels); i++ {
		if candidate := model + "-" + levels[i]; catalog[candidate] {
			return candidate, nil
		}
	}
	if catalog[model] {
		return model, nil
	}
	return model + "-" + level, nil
}
