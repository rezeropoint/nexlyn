package dispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// isEdgeConditionSatisfied 检查边的条件是否满足
// 目前简单实现，仅支持非空条件和空条件
func isEdgeConditionSatisfied(edge core.Edge, ctx context.Context, datastore core.Store) bool {
	condition := edge.GetCondition()

	// 空条件默认为真
	if condition == "" {
		return true
	}

	// TODO: 实现真正的条件解析和求值
	// 默认为真，允许通过
	return true
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
			fmt.Printf("获取图 %s 失败: %v\n", graphKey.ID, err)
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
				fmt.Printf("节点 %s 没有关联的逻辑块\n", currentNode.GetID())
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
				fmt.Printf("执行节点 %s 的逻辑块失败: %v\n", currentNode.GetID(), err)
				continue
			}

			if !success {
				// 逻辑块执行成功但返回失败状态，不继续处理该分支
				fmt.Printf("节点 %s 的逻辑块执行返回失败\n", currentNode.GetID())
				continue
			}

			// 获取所有出边，检查条件，添加下一步节点
			edges := graph.GetOutgoingEdges(currentNode.GetID())
			for _, edge := range edges {
				// 检查边的条件是否满足
				if isEdgeConditionSatisfied(edge, ctx, r.datastore) {
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
