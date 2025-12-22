package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

// ConditionItem 单个条件配置（前端表单生成的结构化条件）
type ConditionItem struct {
	Source   string `json:"source"`   // 变量来源: context / atom
	Path     string `json:"path"`     // 变量路径
	Operator string `json:"operator"` // 操作符: ==, !=, >, <, >=, <=, contains
	Value    string `json:"value"`    // 比较值
}

// ConditionConfig 条件组配置
type ConditionConfig struct {
	Logic      string          `json:"logic"`      // 逻辑: and / or
	Conditions []ConditionItem `json:"conditions"` // 条件列表
}

// isEdgeConditionSatisfied 检查边的条件是否满足
// 支持两种格式：
// 1. 结构化JSON: {"logic":"and","conditions":[...]}
// 2. 表达式字符串: context.xxx.yyy == "value"（向后兼容）
func isEdgeConditionSatisfied(edge core.Edge, ctx context.Context, execCtx core.ExecutionContext, datastore core.Store) bool {
	condition := edge.GetCondition()

	// 空条件默认为真
	if condition == "" {
		return true
	}

	// 尝试解析为结构化条件
	var config ConditionConfig
	if err := json.Unmarshal([]byte(condition), &config); err == nil && len(config.Conditions) > 0 {
		result := evaluateStructuredCondition(ctx, &config, execCtx, datastore)
		logx.WithContext(ctx).Debugf("[EdgeCondition] 结构化条件 = %v", result)
		return result
	}

	// 回退到表达式解析（向后兼容）
	result, err := evaluateEdgeCondition(ctx, condition, execCtx, datastore)
	if err != nil {
		logx.WithContext(ctx).Errorf("[EdgeCondition] 求值失败: %s, 错误: %v", condition, err)
		return false
	}

	logx.WithContext(ctx).Debugf("[EdgeCondition] %s = %v", condition, result)
	return result
}

// evaluateStructuredCondition 求值结构化条件
func evaluateStructuredCondition(ctx context.Context, config *ConditionConfig, execCtx core.ExecutionContext, datastore core.Store) bool {
	if len(config.Conditions) == 0 {
		return true
	}

	results := make([]bool, len(config.Conditions))
	for i, cond := range config.Conditions {
		results[i] = evaluateSingleCondition(ctx, &cond, execCtx, datastore)
	}

	// 根据逻辑组合结果
	if config.Logic == "or" {
		for _, r := range results {
			if r {
				return true
			}
		}
		return false
	}

	// 默认 AND 逻辑
	for _, r := range results {
		if !r {
			return false
		}
	}
	return true
}

// evaluateSingleCondition 求值单个条件
func evaluateSingleCondition(ctx context.Context, cond *ConditionItem, execCtx core.ExecutionContext, datastore core.Store) bool {
	// 获取变量值
	var actualValue any
	if cond.Source == "atom" {
		// atom 来源：支持多级路径访问
		actualValue = getNestedValue(execCtx.GetInfoAtom().GetPayload(), cond.Path)
	} else {
		// context 来源
		parts := strings.SplitN(cond.Path, ".", 2)
		contextKey := parts[0]
		graphContext, err := datastore.GetGraphContext(ctx, execCtx.GetTenantId(), execCtx.GetGraphKey(), contextKey)
		if err != nil || graphContext == nil {
			logx.WithContext(ctx).Debugf("[EdgeCondition] 获取上下文 %s 失败", contextKey)
			return false
		}
		payload := graphContext.GetPayload()
		if len(parts) > 1 {
			// 支持多级路径访问，如 data.on_field_work 或 data.0.field
			actualValue = getNestedValue(payload, parts[1])
		} else {
			actualValue = payload
		}
	}

	logx.WithContext(ctx).Debugf("[EdgeCondition] 路径 %s.%s 的值: %v (类型: %T)",
		cond.Source, cond.Path, actualValue, actualValue)

	// 比较
	return compareValues(actualValue, cond.Operator, cond.Value)
}

// getNestedValue 从嵌套结构中获取值
// 支持多级路径如 "data.on_field_work" 或数组索引如 "data.0.field"
func getNestedValue(data any, path string) any {
	if path == "" {
		return data
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if current == nil {
			return nil
		}

		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		case []any:
			// 尝试解析数组索引
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil {
				if idx >= 0 && idx < len(v) {
					current = v[idx]
				} else {
					return nil
				}
			} else {
				return nil
			}
		case []map[string]any:
			// 尝试解析数组索引
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil {
				if idx >= 0 && idx < len(v) {
					current = v[idx]
				} else {
					return nil
				}
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

// compareValues 比较两个值
func compareValues(actual any, operator, expected string) bool {
	// 转换为字符串进行比较
	actualStr := fmt.Sprintf("%v", actual)

	switch operator {
	case "==":
		return actualStr == expected
	case "!=":
		return actualStr != expected
	case "contains":
		return strings.Contains(actualStr, expected)
	case ">", "<", ">=", "<=":
		// 尝试数值比较
		var actualNum, expectedNum float64
		if _, err := fmt.Sscanf(actualStr, "%f", &actualNum); err != nil {
			return false
		}
		if _, err := fmt.Sscanf(expected, "%f", &expectedNum); err != nil {
			return false
		}
		switch operator {
		case ">":
			return actualNum > expectedNum
		case "<":
			return actualNum < expectedNum
		case ">=":
			return actualNum >= expectedNum
		case "<=":
			return actualNum <= expectedNum
		}
	}
	return false
}

// evaluateEdgeCondition 求值边条件表达式
func evaluateEdgeCondition(ctx context.Context, expression string, execCtx core.ExecutionContext, datastore core.Store) (bool, error) {
	// 预处理表达式：将 context.xxx.yyy 和 atom.xxx 转换为 govaluate 兼容的变量名
	processedExpr := preprocessEdgeExpression(expression)

	// 创建 govaluate 表达式
	expr, err := govaluate.NewEvaluableExpression(processedExpr)
	if err != nil {
		return false, fmt.Errorf("解析表达式失败: %w", err)
	}

	// 构建参数
	params := buildEdgeParameters(ctx, expression, execCtx, datastore)

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

// preprocessEdgeExpression 预处理表达式
// 将 context.xxx.yyy 转换为 context_xxx_yyy，atom.xxx 转换为 atom_xxx
func preprocessEdgeExpression(expr string) string {
	result := expr
	// 替换 context.xxx.yyy 为 context_xxx_yyy（支持多级访问）
	for strings.Contains(result, "context.") {
		idx := strings.Index(result, "context.")
		if idx == -1 {
			break
		}
		// 找到 context. 后面的变量路径
		start := idx + len("context.")
		end := start
		for end < len(result) {
			ch := result[end]
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '_' || ch == '.' {
				end++
			} else {
				break
			}
		}
		varPath := result[start:end]
		// 将路径中的点号替换为下划线
		varName := "context_" + strings.ReplaceAll(varPath, ".", "_")
		result = result[:idx] + varName + result[end:]
	}
	// 替换 atom.xxx 为 atom_xxx
	result = strings.ReplaceAll(result, "atom.", "atom_")
	return result
}

// buildEdgeParameters 构建表达式求值所需的参数
func buildEdgeParameters(ctx context.Context, expression string, execCtx core.ExecutionContext, datastore core.Store) map[string]interface{} {
	params := make(map[string]interface{})

	// 提取 atom.xxx 变量
	atomPayload := execCtx.GetInfoAtom().GetPayload()
	for key, value := range atomPayload {
		params["atom_"+key] = value
	}

	// 提取 context.xxx.yyy 变量
	contextPaths := extractEdgeContextPaths(expression)
	for _, path := range contextPaths {
		parts := strings.SplitN(path, ".", 2)
		contextKey := parts[0]

		graphContext, err := datastore.GetGraphContext(
			ctx,
			execCtx.GetTenantId(),
			execCtx.GetGraphKey(),
			contextKey,
		)
		if err != nil || graphContext == nil {
			continue
		}

		payload := graphContext.GetPayload()
		// 展开 payload 中的所有字段
		for k, v := range payload {
			params["context_"+contextKey+"_"+k] = v
		}
		// 整体保存
		params["context_"+contextKey] = payload
	}

	return params
}

// extractEdgeContextPaths 从表达式中提取所有 context.xxx.yyy 路径
func extractEdgeContextPaths(expression string) []string {
	var paths []string
	seen := make(map[string]bool)

	parts := strings.Split(expression, "context.")
	for i := 1; i < len(parts); i++ {
		var path strings.Builder
		for _, ch := range parts[i] {
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '_' || ch == '.' {
				path.WriteRune(ch)
			} else {
				break
			}
		}
		pathStr := strings.TrimSuffix(path.String(), ".")
		if pathStr != "" && !seen[pathStr] {
			paths = append(paths, pathStr)
			seen[pathStr] = true
		}
	}

	return paths
}

// processInfoAtom 处理单个信息原子
// 这个方法包含原 Dispatch 方法的核心逻辑
func (r *dispatcherRegistry) processInfoAtom(infoAtom core.InfoAtom) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 查询依赖该信息原子的图和入口节点
	graphNodesMap, err := r.queryFunc(infoAtom)
	if err != nil {
		return fmt.Errorf("查询依赖图失败: %w", err)
	}

	// 如果没有找到依赖图，直接返回
	if len(graphNodesMap) == 0 {
		return nil
	}

	return r.processInfoAtomWithGraphs(infoAtom, graphNodesMap)
}

// processInfoAtomWithGraphs 处理信息原子并分发到指定的图和节点
// 这个方法是核心执行逻辑，可被 processInfoAtom 和 DispatchScheduled 复用
func (r *dispatcherRegistry) processInfoAtomWithGraphs(infoAtom core.InfoAtom, graphNodesMap map[core.GraphKey][]core.Node) error {
	return r.processInfoAtomWithGraphsCtx(r.ctx, infoAtom, graphNodesMap)
}

// processInfoAtomWithGraphsCtx 处理信息原子并分发到指定的图和节点（支持自定义 context）
// 用于 DispatchSubGraph 等需要传入自定义 context 的场景
func (r *dispatcherRegistry) processInfoAtomWithGraphsCtx(ctx context.Context, infoAtom core.InfoAtom, graphNodesMap map[core.GraphKey][]core.Node) error {
	// 记录已处理的节点，避免重复处理
	type processingKey struct {
		graphID string
		nodeID  string
	}
	processedNodes := make(map[processingKey]bool)

	// 执行图中的节点
	for graphKey, entryNodes := range graphNodesMap {
		// 基于根上下文创建每个图的处理上下文
		ctx := r.ctx

		// 如果配置了超时，添加超时控制
		if r.config.TimeoutMs > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(r.ctx, time.Duration(r.config.TimeoutMs)*time.Millisecond)
			defer cancel()
		}

		// 获取完整的图实例，用于高效访问图结构
		graph, err := r.getFunc(graphKey)
		if err != nil {
			logx.WithContext(ctx).Errorf("[Dispatcher] 获取图 %s 失败: %v", graphKey.ID, err)
			continue
		}

		// 将入口节点加入队列
		var nodesToProcess []core.Node
		nodesToProcess = append(nodesToProcess, entryNodes...)

		// 处理队列中的所有节点
		for len(nodesToProcess) > 0 {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				return fmt.Errorf("处理超时: %w", ctx.Err())
			default:
				// 继续执行
			}

			// 取出队列头部节点
			currentNode := nodesToProcess[0]
			nodesToProcess = nodesToProcess[1:]

			// 检查节点是否已处理过，避免循环
			key := processingKey{
				graphID: graphKey.ID,
				nodeID:  currentNode.GetID(),
			}
			if processedNodes[key] {
				continue
			}
			processedNodes[key] = true

			// 获取并执行节点的逻辑块
			block := currentNode.GetBlock()
			if block == nil {
				logx.WithContext(ctx).Errorf("[Dispatcher] 节点 %s 没有关联的逻辑块", currentNode.GetID())
				continue
			}

			// 构造执行上下文
			execCtx := core.NewExecutionContext(
				infoAtom.GetTenantId(), // 租户ID
				infoAtom,               // 信息原子
				graph,                  // 逻辑图
				currentNode,            // 当前节点
			)

			// 执行逻辑块
			success, err := block.Execute(ctx, execCtx, r.datastore, r.service)
			if err != nil {
				// 根据配置决定是否继续处理
				if !r.config.ContinueOnNodeFailure {
					return fmt.Errorf("执行节点 %s 的逻辑块失败: %w", currentNode.GetID(), err)
				}
				// 记录错误但继续处理
				logx.WithContext(ctx).Errorf("[Dispatcher] 执行节点 %s 的逻辑块失败: %v", currentNode.GetID(), err)
				continue
			}

			if !success {
				// 逻辑块返回 false 表示终止后续处理（如去重检测到重复），这是正常的流程控制
				logx.WithContext(ctx).Debugf("[Dispatcher] 节点 %s 终止后续处理", currentNode.GetID())
				continue
			}

			// 获取所有出边，检查条件，添加下一步节点
			edges := graph.GetOutgoingEdges(currentNode.GetID())
			for _, edge := range edges {
				// 检查边的条件是否满足
				if isEdgeConditionSatisfied(edge, ctx, execCtx, r.datastore) {
					// 获取目标节点并加入处理队列
					targetID := edge.GetTargetID()
					targetNode, found := graph.FindNodeByID(targetID)
					if found {
						nodesToProcess = append(nodesToProcess, targetNode)
					}
				}
			}
		}
	}

	return nil
}
