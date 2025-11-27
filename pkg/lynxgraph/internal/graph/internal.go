package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/zeromicro/go-zero/core/logx"
)

// watchConfigChanges 监听etcd变更
// 该函数订阅etcd配置变更，当图的配置发生变化时自动更新内存中的图实例
func (r *graphRegistry) watchConfigChanges() error {
	if r.etcdClient == nil {
		return fmt.Errorf("etcd客户端未初始化")
	}

	// 订阅etcd变更，初始化时会触发所有当前键的回调
	if err := r.etcdClient.Subscribe(r.config.EtcdConfig.Key, func(key string, value []byte) error {
		return r.handleConfigChange(key, value)
	}); err != nil {
		return fmt.Errorf("订阅etcd配置变更失败: %v", err)
	}

	return nil
}

// handleConfigChange 处理配置变更
func (r *graphRegistry) handleConfigChange(key string, value []byte) error {
	// 解析变更类型
	// 键格式为: /nexlyn/lynxgraph/action/id
	// 例如: /nexlyn/lynxgraph/update/c55c716a-0d10-40d3-aac8-1c94418933a6
	// 移除 prefix 前缀后再解析
	prefix := r.config.EtcdConfig.Key // /nexlyn/lynxgraph
	if !strings.HasPrefix(key, prefix) {
		return fmt.Errorf("键前缀不匹配: %s", key)
	}

	// 移除前缀和开头的斜杠，得到 action/id
	remaining := strings.TrimPrefix(key, prefix)
	remaining = strings.TrimPrefix(remaining, "/")

	// 分割得到 action 和 id
	parts := strings.Split(remaining, "/")
	if len(parts) < 2 {
		return fmt.Errorf("无效的etcd键格式: %s (期望格式: prefix/action/id)", key)
	}

	action := parts[0]

	logx.Debugf("处理etcd配置变更: %s, action: %s", key, action)

	switch action {
	case "update", "create":
		// 解析通知消息
		var notifyMsg NotifyMessage
		if err := json.Unmarshal(value, &notifyMsg); err != nil {
			return fmt.Errorf("解析通知消息失败: %v, 键: %s", err, key)
		}

		// 使用ID创建GraphKey
		graphKey := core.GraphKey{ID: notifyMsg.ID}

		// 从数据库加载最新配置
		var dbConfig core.GraphConfig
		dbConfig, err := r.GetGraphConfig(context.Background(), graphKey)
		if err == nil {
			// 数据库中存在配置，检查是否启用
			if !dbConfig.Enable {
				// 图未启用，从内存中移除（如果存在）
				_ = r.UnregisterGraph(graphKey)
				return nil
			}

			// 加载图到内存
			if err := r.LoadGraph(dbConfig); err != nil {
				return fmt.Errorf("加载图失败: %v, 键: %s", err, key)
			}
		} else {
			// 数据库中不存在配置，可能是以下情况：
			// 1. REST API 先写Etcd后写数据库，事务还未提交（时序问题）
			// 2. 配置已被删除
			// 3. 权限问题或配置错误
			// 这种情况不应该报错，只记录info日志
			// logx.Infof("Etcd通知的图配置在数据库中不存在，跳过加载: id=%s, key=%s", notifyMsg.ID, key)
			return nil
		}

	case "delete":
		// 解析通知消息
		var notifyMsg NotifyMessage
		if err := json.Unmarshal(value, &notifyMsg); err != nil {
			return fmt.Errorf("解析通知消息失败: %v, 键: %s", err, key)
		}

		// 使用ID创建GraphKey
		graphKey := core.GraphKey{ID: notifyMsg.ID}

		// 从内存中移除图
		if err := r.UnregisterGraph(graphKey); err != nil {
			return fmt.Errorf("注销图失败: %v, 键: %s", err, key)
		}

	default:
		return fmt.Errorf("未知的操作类型: %s, 键: %s", action, key)
	}

	return nil
}

// updateInfoAtomTypeIndex 更新信息原子类型索引
// 该函数为给定的逻辑图中的每个入口节点建立信息原子类型索引
// 索引用于快速查找订阅了特定信息原子类型的图和节点
func (r *graphRegistry) updateInfoAtomTypeIndex(gKey core.GraphKey, graph core.LogicGraph) {
	// 遍历图中的所有节点
	for _, node := range getAllNodes(graph) {
		// 只为入口节点创建索引
		if !node.IsEntryPoint() {
			continue
		}

		// 获取节点订阅的信息原子类型
		atomTypeIDs := node.GetSubscribedInfoAtomTypeIDs()
		if len(atomTypeIDs) == 0 {
			continue // 节点没有订阅任何信息原子类型
		}

		// 为每个节点订阅的信息原子类型添加索引
		for _, atomTypeID := range atomTypeIDs {

			atomKey := core.InfoAtomTypeKey{ID: atomTypeID}

			// 初始化索引结构
			if _, exists := r.infoAtomTypeIndex[atomKey]; !exists {
				r.infoAtomTypeIndex[atomKey] = make(map[core.GraphKey][]string)
			}

			// 添加节点ID到图的节点列表中
			r.infoAtomTypeIndex[atomKey][gKey] = append(r.infoAtomTypeIndex[atomKey][gKey], node.GetID())
		}
	}
}

// findGraphsByInfoAtomType 通过信息原子类型查找订阅该类型的所有逻辑图和入口节点
func (r *graphRegistry) findGraphsByInfoAtomType(infoAtomTypeKey core.InfoAtomTypeKey) (map[core.GraphKey][]core.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 检查参数
	if infoAtomTypeKey.ID == "" {
		return nil, fmt.Errorf("信息原子类型ID不能为空")
	}

	result := make(map[core.GraphKey][]core.Node)

	// 查找索引中的图和节点
	graphToNodesMap, exists := r.infoAtomTypeIndex[infoAtomTypeKey]
	if !exists {
		// 没有找到订阅该信息原子类型的图
		return result, nil
	}

	// 预先分配足够的容量
	if len(graphToNodesMap) > 0 {
		result = make(map[core.GraphKey][]core.Node, len(graphToNodesMap))
	}

	// 构建结果集
	for gKey, nodeIDs := range graphToNodesMap {
		graph, exists := r.graphs[gKey]
		if !exists {
			// 图已不存在（可能是索引未及时清理）
			continue
		}

		// 只检索已启用的图
		if !graph.GetEnable() {
			continue
		}

		// 获取对应的节点
		nodes := make([]core.Node, 0, len(nodeIDs))
		for _, nodeID := range nodeIDs {
			node, err := graph.GetNode(nodeID)
			if err != nil {
				// 记录错误但继续处理
				continue
			}
			if node != nil && node.IsEntryPoint() {
				nodes = append(nodes, node)
			}
		}

		if len(nodes) > 0 {
			result[gKey] = nodes
		}
	}

	return result, nil
}
