package switchblock

import (
	"context"
	"fmt"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// ExpressionEvaluator 表达式求值器
type ExpressionEvaluator struct {
	execCtx   core.ExecutionContext
	datastore core.Store
	ctx       context.Context
}

// NewExpressionEvaluator 创建表达式求值器
func NewExpressionEvaluator(ctx context.Context, execCtx core.ExecutionContext, datastore core.Store) *ExpressionEvaluator {
	return &ExpressionEvaluator{
		execCtx:   execCtx,
		datastore: datastore,
		ctx:       ctx,
	}
}

// Evaluate 求值表达式，返回布尔结果
func (e *ExpressionEvaluator) Evaluate(expression string) (bool, error) {
	// 预处理表达式：将 context.xxx 和 atom.xxx 转换为 govaluate 兼容的变量名
	// govaluate 不支持点号作为变量名的一部分，需要用下划线替换
	processedExpr := preprocessExpression(expression)

	// 创建 govaluate 表达式
	expr, err := govaluate.NewEvaluableExpression(processedExpr)
	if err != nil {
		return false, fmt.Errorf("解析表达式失败: %w", err)
	}

	// 构建参数
	params, err := e.buildParameters(expression)
	if err != nil {
		return false, fmt.Errorf("构建参数失败: %w", err)
	}

	// 求值
	result, err := expr.Evaluate(params)
	if err != nil {
		return false, fmt.Errorf("求值失败: %w", err)
	}

	// 转换为布尔值
	switch v := result.(type) {
	case bool:
		return v, nil
	case float64:
		return v != 0, nil
	case string:
		return v != "", nil
	default:
		return false, fmt.Errorf("表达式结果不是布尔类型: %T", result)
	}
}

// preprocessExpression 预处理表达式，将 context.xxx 和 atom.xxx 转换为 context_xxx 和 atom_xxx
func preprocessExpression(expr string) string {
	// 替换 context.xxx 为 context_xxx
	result := strings.ReplaceAll(expr, "context.", "context_")
	// 替换 atom.xxx 为 atom_xxx
	result = strings.ReplaceAll(result, "atom.", "atom_")
	return result
}

// buildParameters 构建表达式求值所需的参数
func (e *ExpressionEvaluator) buildParameters(expression string) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// 提取 atom.xxx 变量
	atomPayload := e.execCtx.GetInfoAtom().GetPayload()
	for key, value := range atomPayload {
		params["atom_"+key] = value
	}

	// 提取 context.xxx 变量（从 GraphContext 读取）
	// 需要从表达式中提取所有 context.xxx 的 key
	contextKeys := extractContextKeys(expression)
	for _, key := range contextKeys {
		graphContext, err := e.datastore.GetGraphContext(
			e.ctx,
			e.execCtx.GetTenantId(),
			e.execCtx.GetGraphKey(),
			key,
		)
		if err == nil && graphContext != nil {
			payload := graphContext.GetPayload()
			// 将 GraphContext 的 payload 展开到参数中
			for k, v := range payload {
				params["context_"+key+"_"+k] = v
			}
			// 同时保存整个 payload 的第一个值（简化访问）
			if len(payload) == 1 {
				for _, v := range payload {
					params["context_"+key] = v
				}
			} else {
				params["context_"+key] = payload
			}
		}
	}

	return params, nil
}

// extractContextKeys 从表达式中提取所有 context.xxx 的 key
func extractContextKeys(expression string) []string {
	var keys []string
	seen := make(map[string]bool)

	// 简单的字符串解析，查找 context.xxx 模式
	parts := strings.Split(expression, "context.")
	for i := 1; i < len(parts); i++ {
		// 提取变量名（到第一个非字母数字下划线字符为止）
		var key strings.Builder
		for _, ch := range parts[i] {
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '_' {
				key.WriteRune(ch)
			} else {
				break
			}
		}
		keyStr := key.String()
		if keyStr != "" && !seen[keyStr] {
			keys = append(keys, keyStr)
			seen[keyStr] = true
		}
	}

	return keys
}
