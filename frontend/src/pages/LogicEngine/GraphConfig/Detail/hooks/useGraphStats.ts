/**
 * 逻辑图统计数据管理Hook（Mock数据）
 */

import { useState } from 'react';

export interface GraphStats {
  totalExecutions: number;
  successCount: number;
  failureCount: number;
  successRate: number;
  avgDuration: number;
  p95Duration: number;
  p99Duration: number;
}

export interface TrendDataItem {
  time: string;
  success: number;
  failed: number;
}

export interface NodeStat {
  nodeId: string;
  nodeName: string;
  executionCount: number;
}

export const useGraphStats = () => {
  // Mock统计数据
  const [stats] = useState<GraphStats>({
    totalExecutions: 1234,
    successCount: 1154,
    failureCount: 80,
    successRate: 93.5,
    avgDuration: 125,
    p95Duration: 280,
    p99Duration: 450,
  });

  // Mock趋势数据（24小时）
  const [trendData] = useState<TrendDataItem[]>(
    Array.from({ length: 24 }, (_, i) => ({
      time: `${String(i).padStart(2, '0')}:00`,
      success: Math.floor(Math.random() * 50) + 10,
      failed: Math.floor(Math.random() * 10),
    }))
  );

  // Mock节点统计
  const [nodeStats] = useState<NodeStat[]>([
    { nodeId: 'node-001', nodeName: 'Filter节点', executionCount: 450 },
    { nodeId: 'node-002', nodeName: 'Action节点', executionCount: 380 },
    { nodeId: 'node-003', nodeName: 'State节点', executionCount: 324 },
  ]);

  return {
    stats,
    trendData,
    nodeStats,
  };
};
