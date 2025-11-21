/**
 * Tab5: 统计分析
 * 使用Ant Design Charts展示执行统计和性能指标
 */

import {
  PlayCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import { Alert, Card, Col, Row, Space, Statistic } from 'antd';
import { Line, Pie, Column } from '@ant-design/charts';
import React from 'react';
import { useGraphStats } from '../../hooks/useGraphStats';

interface StatsTabProps {
  graphId: string;
}

/**
 * 统计分析Tab组件
 */
const StatsTab: React.FC<StatsTabProps> = () => {
  const { stats, trendData, nodeStats } = useGraphStats();

  // 执行趋势数据（转换为图表格式）
  const trendChartData = trendData.flatMap((d) => [
    { time: d.time, type: '成功', count: d.success },
    { time: d.time, type: '失败', count: d.failed },
  ]);

  // 性能指标数据
  const performanceData = [
    { metric: '平均耗时', value: stats.avgDuration },
    { metric: 'P95耗时', value: stats.p95Duration },
    { metric: 'P99耗时', value: stats.p99Duration },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Alert
        message="数据说明"
        description="当前展示的是Mock数据，用于演示统计分析功能。后续版本将对接真实的执行数据和性能指标。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        closable
      />

      {/* 统计卡片 */}
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic
              title="执行总数"
              value={stats.totalExecutions}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功次数"
              value={stats.successCount}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="失败次数"
              value={stats.failureCount}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功率"
              value={stats.successRate}
              suffix="%"
              precision={1}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 执行趋势图 */}
      <Card title="执行趋势（24小时）">
        <Line
          data={trendChartData}
          xField="time"
          yField="count"
          seriesField="type"
          height={300}
          color={['#52c41a', '#ff4d4f']}
          smooth
          legend={{ position: 'top' }}
          point={{
            size: 3,
            shape: 'circle',
          }}
        />
      </Card>

      {/* 性能指标 */}
      <Card title="性能指标（毫秒）">
        <Column
          data={performanceData}
          xField="metric"
          yField="value"
          height={300}
          color="#1890ff"
          label={{
            position: 'top',
            style: { fill: '#000' },
          }}
        />
      </Card>

      {/* 节点执行统计 */}
      <Card title="节点执行统计">
        <Pie
          data={nodeStats}
          angleField="executionCount"
          colorField="nodeName"
          height={300}
          label={{
            type: 'outer',
            content: '{name} {percentage}',
          }}
          interactions={[{ type: 'element-active' }]}
        />
      </Card>
    </Space>
  );
};

export default StatsTab;
