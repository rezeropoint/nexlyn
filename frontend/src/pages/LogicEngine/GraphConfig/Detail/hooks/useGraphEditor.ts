/**
 * 逻辑图编辑器状态管理Hook
 * @description 管理节点、边的状态，提供增删改查等操作方法
 */

import type { NodeConfig, EdgeConfig } from '@/services/lynxmanager/types';
import { useState, useCallback, useMemo, useRef } from 'react';
import { nanoid } from 'nanoid';

export interface SelectedItem {
  type: 'node' | 'edge';
  id: string;
}

export interface GraphEditorState {
  nodes: NodeConfig[];
  edges: EdgeConfig[];
  selectedItem: SelectedItem | null;
  hasChanges: boolean; // 是否有未保存的修改
}

export interface GraphEditorActions {
  // 节点操作
  addNode: (node: Omit<NodeConfig, 'id'>) => string; // 返回新节点ID
  updateNode: (id: string, updates: Partial<NodeConfig>) => void;
  deleteNode: (id: string) => void;
  getNode: (id: string) => NodeConfig | undefined;

  // 边操作
  addEdge: (edge: Omit<EdgeConfig, 'id'>) => string; // 返回新边ID
  updateEdge: (id: string, updates: Partial<EdgeConfig>) => void;
  deleteEdge: (id: string) => void;
  getEdge: (id: string) => EdgeConfig | undefined;

  // 选中操作
  selectNode: (id: string) => void;
  selectEdge: (id: string) => void;
  clearSelection: () => void;

  // 批量操作
  setNodesAndEdges: (nodes: NodeConfig[], edges: EdgeConfig[]) => void;
  clearGraph: () => void;

  // 状态管理
  markAsChanged: () => void;
  markAsSaved: () => void;
}

/**
 * 逻辑图编辑器Hook
 */
export const useGraphEditor = (
  initialNodes: NodeConfig[] = [],
  initialEdges: EdgeConfig[] = [],
): [GraphEditorState, GraphEditorActions] => {
  const [nodes, setNodes] = useState<NodeConfig[]>(initialNodes);
  const [edges, setEdges] = useState<EdgeConfig[]>(initialEdges);
  const [selectedItem, setSelectedItem] = useState<SelectedItem | null>(null);
  const [hasChanges, setHasChanges] = useState(false);

  // 使用 ref 追踪 selectedItem 以打破 useCallback 的依赖循环
  const selectedItemRef = useRef(selectedItem);
  selectedItemRef.current = selectedItem;

  // ===== 节点操作 =====

  const addNode = useCallback((node: Omit<NodeConfig, 'id'>): string => {
    const newId = `node-${nanoid(8)}`;
    const newNode: NodeConfig = {
      ...node,
      id: newId,
    };
    setNodes((prev) => [...prev, newNode]);
    setHasChanges(true);
    return newId;
  }, []);

  const updateNode = useCallback((id: string, updates: Partial<NodeConfig>) => {
    setNodes((prev) =>
      prev.map((node) => (node.id === id ? { ...node, ...updates } : node)),
    );
    setHasChanges(true);
  }, []);

  const deleteNode = useCallback((id: string) => {
    // 删除节点及其相关的边
    setNodes((prev) => prev.filter((node) => node.id !== id));
    setEdges((prev) => prev.filter((edge) => edge.sourceID !== id && edge.targetID !== id));
    if (selectedItemRef.current?.type === 'node' && selectedItemRef.current.id === id) {
      setSelectedItem(null);
    }
    setHasChanges(true);
  }, []);

  const getNode = useCallback(
    (id: string): NodeConfig | undefined => {
      return nodes.find((node) => node.id === id);
    },
    [nodes],
  );

  // ===== 边操作 =====

  const addEdge = useCallback((edge: Omit<EdgeConfig, 'id'>): string => {
    const newId = `edge-${nanoid(8)}`;
    const newEdge: EdgeConfig = {
      ...edge,
      id: newId,
    };
    setEdges((prev) => [...prev, newEdge]);
    setHasChanges(true);
    return newId;
  }, []);

  const updateEdge = useCallback((id: string, updates: Partial<EdgeConfig>) => {
    setEdges((prev) =>
      prev.map((edge) => (edge.id === id ? { ...edge, ...updates } : edge)),
    );
    setHasChanges(true);
  }, []);

  const deleteEdge = useCallback((id: string) => {
    setEdges((prev) => prev.filter((edge) => edge.id !== id));
    if (selectedItemRef.current?.type === 'edge' && selectedItemRef.current.id === id) {
      setSelectedItem(null);
    }
    setHasChanges(true);
  }, []);

  const getEdge = useCallback(
    (id: string): EdgeConfig | undefined => {
      return edges.find((edge) => edge.id === id);
    },
    [edges],
  );

  // ===== 选中操作 =====

  const selectNode = useCallback((id: string) => {
    setSelectedItem({ type: 'node', id });
  }, []);

  const selectEdge = useCallback((id: string) => {
    setSelectedItem({ type: 'edge', id });
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedItem(null);
  }, []);

  // ===== 批量操作 =====

  const setNodesAndEdges = useCallback((newNodes: NodeConfig[], newEdges: EdgeConfig[]) => {
    setNodes(newNodes);
    setEdges(newEdges);
    setSelectedItem(null);
    setHasChanges(false);
  }, []);

  const clearGraph = useCallback(() => {
    setNodes([]);
    setEdges([]);
    setSelectedItem(null);
    setHasChanges(true);
  }, []);

  // ===== 状态管理 =====

  const markAsChanged = useCallback(() => {
    setHasChanges(true);
  }, []);

  const markAsSaved = useCallback(() => {
    setHasChanges(false);
  }, []);

  // ===== 返回值 =====

  const state: GraphEditorState = useMemo(
    () => ({
      nodes,
      edges,
      selectedItem,
      hasChanges,
    }),
    [nodes, edges, selectedItem, hasChanges],
  );

  const actions: GraphEditorActions = useMemo(
    () => ({
      addNode,
      updateNode,
      deleteNode,
      getNode,
      addEdge,
      updateEdge,
      deleteEdge,
      getEdge,
      selectNode,
      selectEdge,
      clearSelection,
      setNodesAndEdges,
      clearGraph,
      markAsChanged,
      markAsSaved,
    }),
    [
      addNode,
      updateNode,
      deleteNode,
      getNode,
      addEdge,
      updateEdge,
      deleteEdge,
      getEdge,
      selectNode,
      selectEdge,
      clearSelection,
      setNodesAndEdges,
      clearGraph,
      markAsChanged,
      markAsSaved,
    ],
  );

  return [state, actions];
};
