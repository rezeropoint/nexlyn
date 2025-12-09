/**
 * Tab2: 可视化编辑器（完整版本）
 * @description 使用AntV X6实现完整的图形编辑功能
 */

import { getBlockSpecs } from "@/services/lynxmanager/api";
import type {
  BlockSpec,
  EdgeConfig,
  GraphConfigDetail,
  NodeConfig,
} from "@/services/lynxmanager/types";
import { ExclamationCircleOutlined } from "@ant-design/icons";
import { App, Modal, Splitter } from "antd";
import dagre from "dagre";
import React, { useCallback, useEffect, useRef, useState } from "react";
import { useGraphEditor } from "../../hooks/useGraphEditor";
import BlockPanel from "./components/BlockPanel";
import type { GraphCanvasRef } from "./components/GraphCanvas";
import GraphCanvas from "./components/GraphCanvas";
import PropertyPanel from "./components/PropertyPanel";
import Toolbar from "./components/Toolbar";
import "./index.less";

interface ConfigField {
  type?: string;
  title?: string;
  check?: string;
  properties?: Record<string, ConfigField>;
  required?: string[];
}

/**
 * 验证节点配置的必填字段
 * @returns 错误信息数组，空数组表示验证通过
 */
const validateNodeConfigs = (
  nodes: NodeConfig[],
  blockSpecs: Record<string, BlockSpec>
): string[] => {
  const errors: string[] = [];

  nodes.forEach((node) => {
    const spec = blockSpecs[node.blockType];
    if (!spec?.configSchema) return;

    let schema: { properties?: Record<string, ConfigField>; required?: string[] };
    try {
      schema = JSON.parse(spec.configSchema);
    } catch {
      return;
    }

    const properties = schema.properties || {};
    const requiredFields = schema.required || [];
    const config = typeof node.blockConfig === "string"
      ? JSON.parse(node.blockConfig || "{}")
      : node.blockConfig || {};

    // 检查每个字段
    Object.entries(properties).forEach(([fieldName, fieldSchema]) => {
      const isRequired = requiredFields.includes(fieldName) || fieldSchema.check === "must";
      if (!isRequired) return;

      const value = config[fieldName];
      const isEmpty = value === undefined || value === null || value === "" ||
        (Array.isArray(value) && value.length === 0);

      if (isEmpty) {
        const fieldTitle = fieldSchema.title || fieldName;
        errors.push(`节点 "${node.id}" 的 "${fieldTitle}" 字段为必填项`);
      }
    });
  });

  return errors;
};

interface VisualEditorTabProps {
  data: GraphConfigDetail | null;
  loading: boolean;
  tenantId?: string;
  onSave: (nodes: NodeConfig[], edges: EdgeConfig[]) => Promise<boolean>;
}

/**
 * 可视化编辑器Tab组件
 */
const VisualEditorTab: React.FC<VisualEditorTabProps> = ({
  data,
  tenantId,
  onSave,
}) => {
  const { message } = App.useApp();
  const canvasRef = useRef<GraphCanvasRef>(null);
  const [saving, setSaving] = useState(false);
  const [blockSpecs, setBlockSpecs] = useState<Record<string, BlockSpec>>({});

  // 使用编辑器Hook管理状态
  const [editorState, editorActions] = useGraphEditor(
    data?.nodes || [],
    data?.edges || []
  );

  // 获取所有 BlockSpecs 用于验证
  useEffect(() => {
    const fetchBlockSpecs = async () => {
      try {
        const response = await getBlockSpecs();
        if (response.code === 0 && response.data?.list) {
          const specsMap = response.data.list.reduce((acc, spec) => {
            acc[spec.blockType] = spec;
            return acc;
          }, {} as Record<string, BlockSpec>);
          setBlockSpecs(specsMap);
        }
      } catch (error) {
        console.error("Failed to fetch block specs:", error);
      }
    };
    fetchBlockSpecs();
  }, []);

  // 追踪当前逻辑图ID，只在ID变化时重新加载数据
  const currentGraphIdRef = useRef<string | null>(null);
  const dataRef = useRef(data);
  dataRef.current = data;

  // 当数据加载时，初始化编辑器状态（仅在逻辑图ID变化时）
  useEffect(() => {
    const currentData = dataRef.current;
    const graphId = currentData?.id || null;
    if (currentData && graphId && graphId !== currentGraphIdRef.current) {
      currentGraphIdRef.current = graphId;
      editorActions.setNodesAndEdges(
        currentData.nodes || [],
        currentData.edges || []
      );
    }
  }, [data?.id, editorActions]); // 只依赖ID和actions，通过ref访问data避免重复初始化

  // 使用 useCallback 稳定事件处理函数的引用
  const handleNodeClick = useCallback(
    (node: NodeConfig) => {
      editorActions.selectNode(node.id);
    },
    [editorActions]
  );

  const handleEdgeClick = useCallback(
    (edge: EdgeConfig) => {
      editorActions.selectEdge(edge.id);
    },
    [editorActions]
  );

  const handleNodeMoved = useCallback(
    (nodeId: string, x: number, y: number) => {
      editorActions.updateNode(nodeId, { x, y });
    },
    [editorActions]
  );

  const handleNodeAdded = useCallback(
    (node: NodeConfig) => {
      editorActions.addNode(node);
    },
    [editorActions]
  );

  const handleEdgeAdded = useCallback(
    (edge: EdgeConfig) => {
      editorActions.addEdge(edge);
    },
    [editorActions]
  );

  const handleNodeDeleted = useCallback(
    (nodeId: string) => {
      editorActions.deleteNode(nodeId);
    },
    [editorActions]
  );

  const handleEdgeDeleted = useCallback(
    (edgeId: string) => {
      editorActions.deleteEdge(edgeId);
    },
    [editorActions]
  );

  const handleCanvasClick = useCallback(() => {
    editorActions.clearSelection();
  }, [editorActions]);

  // 自动布局算法
  const handleAutoLayout = () => {
    const g = new dagre.graphlib.Graph();
    g.setGraph({
      rankdir: "TB",
      nodesep: 50,
      ranksep: 80,
      marginx: 20,
      marginy: 20,
    });
    g.setDefaultEdgeLabel(() => ({}));

    // 添加节点
    editorState.nodes.forEach((node) => {
      g.setNode(node.id, { width: 160, height: 80 });
    });

    // 添加边
    editorState.edges.forEach((edge) => {
      g.setEdge(edge.sourceID, edge.targetID);
    });

    // 执行布局
    dagre.layout(g);

    // 更新节点位置
    editorState.nodes.forEach((node) => {
      const nodeWithPosition = g.node(node.id);
      if (nodeWithPosition) {
        editorActions.updateNode(node.id, {
          x: nodeWithPosition.x - 80,
          y: nodeWithPosition.y - 40,
        });
      }
    });

    canvasRef.current?.fitView();
    message.success("自动布局完成");
  };

  // 保存逻辑图
  const handleSave = async () => {
    // 客户端验证：检查所有节点的必填字段
    const validationErrors = validateNodeConfigs(editorState.nodes, blockSpecs);
    if (validationErrors.length > 0) {
      // 显示第一个错误，更多错误在控制台输出
      message.error(validationErrors[0]);
      if (validationErrors.length > 1) {
        console.warn("节点配置验证错误:", validationErrors);
      }
      return;
    }

    setSaving(true);
    try {
      const success = await onSave(editorState.nodes, editorState.edges);
      if (success) {
        editorActions.markAsSaved();
        // 保存成功后，强制重置 graphId 引用，以便下次数据刷新时能重新加载
        currentGraphIdRef.current = null;
        message.success("保存成功");
      }
    } catch (error) {
      message.error(`保存失败：${(error as Error).message}`);
    } finally {
      setSaving(false);
    }
  };

  // 清空画布
  const handleClear = () => {
    Modal.confirm({
      title: "确认清空",
      icon: <ExclamationCircleOutlined />,
      content: "确定要清空画布吗？此操作不可撤销。",
      okText: "确定",
      cancelText: "取消",
      onOk: () => {
        editorActions.clearGraph();
        message.success("画布已清空");
      },
    });
  };

  // 获取选中的节点/边
  const selectedNode =
    editorState.selectedItem?.type === "node"
      ? editorActions.getNode(editorState.selectedItem.id)
      : undefined;

  const selectedEdge =
    editorState.selectedItem?.type === "edge"
      ? editorActions.getEdge(editorState.selectedItem.id)
      : undefined;

  return (
    <div className="visual-editor-tab">
      {/* 顶部工具栏 */}
      <Toolbar
        hasChanges={editorState.hasChanges}
        canUndo={canvasRef.current?.canUndo() ?? false}
        canRedo={canvasRef.current?.canRedo() ?? false}
        onSave={handleSave}
        onUndo={() => canvasRef.current?.undo()}
        onRedo={() => canvasRef.current?.redo()}
        onAutoLayout={handleAutoLayout}
        onZoomIn={() => canvasRef.current?.zoomIn()}
        onZoomOut={() => canvasRef.current?.zoomOut()}
        onFitView={() => canvasRef.current?.fitView()}
        onCenter={() => canvasRef.current?.centerContent()}
        onClear={handleClear}
        saving={saving}
      />

      {/* 三栏布局：左侧逻辑积木面板、中间画布、右侧属性面板 */}
      <div className="editor-content">
        <Splitter>
          {/* 左侧：逻辑积木面板 */}
          <Splitter.Panel defaultSize={260} min={220} max={450}>
            <BlockPanel />
          </Splitter.Panel>

          {/* 中间：画布（自动填充剩余空间）*/}
          <Splitter.Panel min={480}>
            <GraphCanvas
              ref={canvasRef}
              nodes={editorState.nodes}
              edges={editorState.edges}
              selectedNodeId={selectedNode?.id}
              selectedEdgeId={selectedEdge?.id}
              onNodeClick={handleNodeClick}
              onEdgeClick={handleEdgeClick}
              onNodeMoved={handleNodeMoved}
              onNodeAdded={handleNodeAdded}
              onEdgeAdded={handleEdgeAdded}
              onNodeDeleted={handleNodeDeleted}
              onEdgeDeleted={handleEdgeDeleted}
              onCanvasClick={handleCanvasClick}
            />
          </Splitter.Panel>

          {/* 右侧：属性面板 */}
          <Splitter.Panel defaultSize={320} min={280} max={450} collapsible>
            <PropertyPanel
              selectedNode={selectedNode}
              selectedEdge={selectedEdge}
              tenantId={tenantId || data?.tenantId}
              nodes={editorState.nodes}
              onNodeUpdate={(nodeId, updates) =>
                editorActions.updateNode(nodeId, updates)
              }
              onEdgeUpdate={(edgeId, updates) =>
                editorActions.updateEdge(edgeId, updates)
              }
            />
          </Splitter.Panel>
        </Splitter>
      </div>
    </div>
  );
};

export default VisualEditorTab;
