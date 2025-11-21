/**
 * 图形画布组件
 * @description AntV X6 画布，支持拖放、连线、编辑等功能
 */

import type { NodeConfig, EdgeConfig } from '@/services/lynxmanager/types';
import { Graph } from '@antv/x6';
import type { Edge } from '@antv/x6';
import { Snapline } from '@antv/x6-plugin-snapline';
import { History } from '@antv/x6-plugin-history';
import { Keyboard } from '@antv/x6-plugin-keyboard';
import { Selection } from '@antv/x6-plugin-selection';
import { Clipboard } from '@antv/x6-plugin-clipboard';
import { theme } from 'antd';
import React, { useEffect, useRef, useImperativeHandle, forwardRef } from 'react';
import './GraphCanvas.less';

export interface GraphCanvasRef {
  graph: Graph | null;
  undo: () => void;
  redo: () => void;
  canUndo: () => boolean;
  canRedo: () => boolean;
  fitView: () => void;
  zoomIn: () => void;
  zoomOut: () => void;
  zoomToFit: () => void;
  centerContent: () => void;
}

interface GraphCanvasProps {
  nodes: NodeConfig[];
  edges: EdgeConfig[];
  selectedNodeId?: string;
  selectedEdgeId?: string;
  onNodeClick?: (node: NodeConfig) => void;
  onEdgeClick?: (edge: EdgeConfig) => void;
  onNodeMoved?: (nodeId: string, x: number, y: number) => void;
  onNodeAdded?: (node: NodeConfig) => void;
  onEdgeAdded?: (edge: EdgeConfig) => void;
  onNodeDeleted?: (nodeId: string) => void;
  onEdgeDeleted?: (edgeId: string) => void;
  onCanvasClick?: () => void;
}

/**
 * 图形画布组件
 */
const GraphCanvas = forwardRef<GraphCanvasRef, GraphCanvasProps>(
  (
    {
      nodes,
      edges,
      selectedNodeId,
      selectedEdgeId,
      onNodeClick,
      onEdgeClick,
      onNodeMoved,
      onNodeAdded,
      onEdgeAdded,
      onNodeDeleted,
      onEdgeDeleted,
      onCanvasClick,
    },
    ref,
  ) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const graphRef = useRef<Graph | null>(null);
    const { token } = theme.useToken();

    // 初始化画布
    useEffect(() => {
      if (!containerRef.current) return;

      // 获取容器尺寸（clientWidth/clientHeight 已经是内容区域，不包含边框）
      const containerWidth = containerRef.current.clientWidth;
      const containerHeight = containerRef.current.clientHeight;

      const graph: Graph = new Graph({
        container: containerRef.current,
        width: containerWidth,
        height: containerHeight,
        autoResize: true,
        grid: {
          size: 10,
          visible: true,
          type: 'doubleMesh',
          args: [
            {
              color: token.colorBorderSecondary,
              thickness: 1,
            },
            {
              color: token.colorBorder,
              thickness: 1,
              factor: 4,
            },
          ],
        },
        // 画布平移配置（按住 Shift 键拖拽空白区域平移）
        panning: {
          enabled: true,
          modifiers: 'shift', // Shift + 拖拽平移，避免与框选冲突
        },
        // 鼠标滚轮缩放
        mousewheel: {
          enabled: true,
          modifiers: 'ctrl', // Ctrl + 滚轮缩放
          minScale: 0.2,
          maxScale: 3,
        },
        connecting: {
          snap: true,
          allowBlank: false,
          allowLoop: false,
          allowNode: true,
          allowEdge: false,
          highlight: true,
          connector: 'rounded',
          connectionPoint: 'boundary',
          router: {
            name: 'manhattan',
            args: {
              padding: 10,
            },
          },
          createEdge(): Edge {
            return graph.createEdge({
              attrs: {
                line: {
                  stroke: token.colorPrimary,
                  strokeWidth: 2,
                  targetMarker: {
                    name: 'block',
                    width: 8,
                    height: 8,
                  },
                },
              },
            });
          },
          validateConnection({ sourceView, targetView }) {
            return sourceView !== targetView;
          },
        },
        highlighting: {
          magnetAvailable: {
            name: 'stroke',
            args: {
              attrs: {
                fill: token.colorPrimary,
                stroke: token.colorPrimary,
              },
            },
          },
        },
      });

      // 注册插件
      graph
        .use(new Snapline({ enabled: true }))
        .use(new History({ enabled: true }))
        .use(new Keyboard({ enabled: true }))
        .use(
          new Selection({
            enabled: true,
            rubberband: true, // 框选功能（直接拖拽空白区域）
            showNodeSelectionBox: true,
          }),
        )
        .use(new Clipboard({ enabled: true }));

      // 快捷键绑定
      graph.bindKey('delete', () => {
        const cells = graph.getSelectedCells();
        cells.forEach((cell: any) => {
          if (cell.isNode()) {
            const nodeData = cell.getData() as NodeConfig;
            onNodeDeleted?.(nodeData.id);
          } else if (cell.isEdge()) {
            const edgeData = cell.getData() as EdgeConfig;
            onEdgeDeleted?.(edgeData.id);
          }
          cell.remove();
        });
      });

      graph.bindKey(['ctrl+c', 'meta+c'], () => {
        const cells = graph.getSelectedCells();
        if (cells.length) {
          graph.copy(cells);
        }
      });

      graph.bindKey(['ctrl+v', 'meta+v'], () => {
        if (!graph.isClipboardEmpty()) {
          const cells = graph.paste({ offset: 32 });
          graph.cleanSelection();
          graph.select(cells);
        }
      });

      graph.bindKey(['ctrl+z', 'meta+z'], () => {
        graph.undo();
      });

      graph.bindKey(['ctrl+shift+z', 'meta+shift+z'], () => {
        graph.redo();
      });

      graphRef.current = graph;

      // 监听事件
      graph.on('node:click', ({ node }: any) => {
        const nodeData = node.getData() as NodeConfig;
        onNodeClick?.(nodeData);
      });

      graph.on('edge:click', ({ edge }: any) => {
        const edgeData = edge.getData() as EdgeConfig;
        onEdgeClick?.(edgeData);
      });

      graph.on('node:moved', ({ node }: any) => {
        const nodeData = node.getData() as NodeConfig;
        const position = node.getPosition();
        onNodeMoved?.(nodeData.id, position.x, position.y);
      });

      graph.on('edge:connected', ({ edge }: any) => {
        const source = edge.getSourceCellId();
        const target = edge.getTargetCellId();
        if (source && target) {
          const edgeData: EdgeConfig = {
            id: '',
            sourceID: source,
            targetID: target,
          };
          onEdgeAdded?.(edgeData);
        }
      });

      graph.on('blank:click', () => {
        onCanvasClick?.();
      });

      // 支持拖放（使用原生 HTML5 拖放 API）
      const handleDrop = (e: DragEvent) => {
        e.preventDefault();

        // 获取鼠标位置（相对于画布）
        const point = graph.clientToLocal(e.clientX, e.clientY);

        try {
          const jsonData = e.dataTransfer?.getData('application/json');

          if (!jsonData) {
            return;
          }

          const data = JSON.parse(jsonData);

          if (data.blockType) {
            const newNode: NodeConfig = {
              id: '',
              type: 'action',
              blockType: data.blockType,
              blockVersion: data.blockVersion || 'v1',
              isEntryPoint: false,
            };

            onNodeAdded?.({
              ...newNode,
              x: point.x,
              y: point.y,
            } as NodeConfig & { x: number; y: number });
          }
        } catch (error) {
          console.error('拖拽节点失败:', error);
        }
      };

      // 监听原生 drop 事件
      const container = containerRef.current;
      container.addEventListener('drop', handleDrop);

      return () => {
        container.removeEventListener('drop', handleDrop);
        graph.dispose();
      };
    }, [
      token,
      onNodeClick,
      onEdgeClick,
      onNodeMoved,
      onNodeAdded,
      onEdgeAdded,
      onNodeDeleted,
      onEdgeDeleted,
      onCanvasClick,
    ]);

    // 渲染节点和边（仅在 nodes/edges 变化时执行）
    useEffect(() => {
      const graph = graphRef.current;
      if (!graph) return;

      // 临时禁用历史记录，避免渲染操作被记录
      const historyEnabled = graph.isHistoryEnabled();
      if (historyEnabled) {
        graph.disableHistory();
      }

      try {
        // 获取当前画布上的所有节点和边ID
        const existingNodeIds = new Set(
          graph.getNodes().map((node) => node.id),
        );
        const existingEdgeIds = new Set(
          graph.getEdges().map((edge) => edge.id),
        );

        // 新数据的ID集合
        const newNodeIds = new Set(nodes.map((node) => node.id));
        const newEdgeIds = new Set(edges.map((edge) => edge.id));

        // === 1. 删除不再存在的节点和边 ===
        existingNodeIds.forEach((id) => {
          if (!newNodeIds.has(id)) {
            const cell = graph.getCellById(id);
            if (cell) {
              cell.remove({ silent: true });
            }
          }
        });

        existingEdgeIds.forEach((id) => {
          if (!newEdgeIds.has(id)) {
            const cell = graph.getCellById(id);
            if (cell) {
              cell.remove({ silent: true });
            }
          }
        });

        // === 2. 更新或添加节点 ===
        nodes.forEach((node) => {
          const existingCell = graph.getCellById(node.id);
          const x = node.x ?? 100;
          const y = node.y ?? 100;

          if (existingCell?.isNode()) {
            // 节点已存在，更新属性
            existingCell.position(x, y, { silent: true });
            existingCell.setData(node, { silent: true });
            existingCell.setAttrs({
              body: {
                fill: node.isEntryPoint ? token.colorPrimaryBg : token.colorBgContainer,
                stroke: node.isEntryPoint ? token.colorPrimary : token.colorBorder,
                strokeWidth: 2,
                rx: 6,
                ry: 6,
              },
              label: {
                text: `${node.blockType}\\n${node.id.slice(0, 8)}`,
                fill: token.colorText,
                fontSize: 12,
                textWrap: {
                  width: 140,
                  height: 60,
                  ellipsis: true,
                },
              },
            }, { silent: true });
          } else {
            // 节点不存在，添加新节点
            graph.addNode({
              id: node.id,
              shape: 'rect',
              x,
              y,
              width: 160,
              height: 80,
              data: node,
              label: `${node.blockType}\\n${node.id.slice(0, 8)}`,
              attrs: {
                body: {
                  fill: node.isEntryPoint ? token.colorPrimaryBg : token.colorBgContainer,
                  stroke: node.isEntryPoint ? token.colorPrimary : token.colorBorder,
                  strokeWidth: 2,
                  rx: 6,
                  ry: 6,
                },
                label: {
                  text: `${node.blockType}\\n${node.id.slice(0, 8)}`,
                  fill: token.colorText,
                  fontSize: 12,
                  textWrap: {
                    width: 140,
                    height: 60,
                    ellipsis: true,
                  },
                },
              },
              ports: {
                groups: {
                  in: {
                    position: 'top',
                    attrs: {
                      circle: {
                        r: 4,
                        magnet: true,
                        stroke: token.colorPrimary,
                        strokeWidth: 2,
                        fill: token.colorBgContainer,
                      },
                    },
                  },
                  out: {
                    position: 'bottom',
                    attrs: {
                      circle: {
                        r: 4,
                        magnet: true,
                        stroke: token.colorPrimary,
                        strokeWidth: 2,
                        fill: token.colorBgContainer,
                      },
                    },
                  },
                },
                items: [
                  { group: 'in', id: 'in' },
                  { group: 'out', id: 'out' },
                ],
              },
            });
          }
        });

        // === 3. 更新或添加边 ===
        edges.forEach((edge) => {
          const existingCell = graph.getCellById(edge.id);

          if (existingCell?.isEdge()) {
            // 边已存在，更新属性
            existingCell.setSource({ cell: edge.sourceID }, { silent: true });
            existingCell.setTarget({ cell: edge.targetID }, { silent: true });
            existingCell.setData(edge, { silent: true });
            existingCell.setLabels([edge.condition || ''], { silent: true });
          } else {
            // 边不存在，添加新边
            graph.addEdge({
              id: edge.id,
              source: edge.sourceID,
              target: edge.targetID,
              data: edge,
              label: edge.condition || '',
              attrs: {
                line: {
                  stroke: token.colorTextTertiary,
                  strokeWidth: 2,
                  targetMarker: {
                    name: 'block',
                    width: 8,
                    height: 8,
                  },
                },
                text: {
                  text: edge.condition || '',
                  fill: token.colorText,
                  fontSize: 11,
                },
              },
              router: {
                name: 'orth',
              },
            });
          }
        });
      } finally {
        // 恢复历史记录
        if (historyEnabled) {
          graph.enableHistory();
        }
      }
    }, [nodes, edges, token]);

    // 处理选中状态（仅在 selectedId 变化时执行）
    useEffect(() => {
      const graph = graphRef.current;
      if (!graph) return;

      const targetId = selectedNodeId || selectedEdgeId;
      const currentSelection = graph.getSelectedCells();
      const currentSelectedId = currentSelection.length === 1 ? currentSelection[0].id : null;

      // 如果选中状态已经正确，则无需任何操作
      if (targetId === currentSelectedId) {
        return;
      }

      // 清空现有选中
      graph.cleanSelection();

      // 选中目标
      if (targetId) {
        const cell = graph.getCellById(targetId);
        if (cell) {
          graph.select(cell);
        }
      }
    }, [selectedNodeId, selectedEdgeId]);

    // 暴露给父组件的方法
    useImperativeHandle(ref, () => ({
      graph: graphRef.current,
      undo: () => graphRef.current?.undo(),
      redo: () => graphRef.current?.redo(),
      canUndo: () => graphRef.current?.canUndo() ?? false,
      canRedo: () => graphRef.current?.canRedo() ?? false,
      fitView: () => graphRef.current?.zoomToFit({ padding: 20 }),
      zoomIn: () => graphRef.current?.zoom(0.1),
      zoomOut: () => graphRef.current?.zoom(-0.1),
      zoomToFit: () => graphRef.current?.zoomToFit({ padding: 20 }),
      centerContent: () => graphRef.current?.centerContent(),
    }));

    // 处理拖拽事件（允许 drop）
    const handleDragOver = (e: React.DragEvent) => {
      e.preventDefault(); // 必须阻止默认行为才能允许 drop
      e.dataTransfer.dropEffect = 'copy';
    };

    return <div ref={containerRef} className="graph-canvas" onDragOver={handleDragOver} />;
  },
);

GraphCanvas.displayName = 'GraphCanvas';

export default GraphCanvas;
