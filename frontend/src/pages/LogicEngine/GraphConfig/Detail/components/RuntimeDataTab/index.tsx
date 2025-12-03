/**
 * Tab3: 运行时数据
 * 展示执行日志、信息原子列表、图上下文
 */

import { ReloadOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { ProTable } from '@ant-design/pro-components';
import { Alert, Button, Card, Empty, Space } from 'antd';
import React from 'react';

interface RuntimeDataTabProps {
  graphId: string;
}

/**
 * 运行时数据Tab组件
 */
const RuntimeDataTab: React.FC<RuntimeDataTabProps> = () => {
  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Alert
        message="功能开发中"
        description="运行时数据功能正在开发中，后续版本将对接真实的Redis运行时数据。"
        type="info"
        icon={<InfoCircleOutlined />}
        showIcon
        closable
      />

      {/* 执行日志表格 */}
      <Card
        title="执行日志"
        extra={
          <Button icon={<ReloadOutlined />} onClick={() => null}>
            刷新
          </Button>
        }
      >
        <ProTable
          dataSource={[]}
          columns={[
            {
              title: '信息原子ID',
              dataIndex: 'infoAtomId',
              width: 150,
            },
            {
              title: '信息原子类型',
              dataIndex: 'infoAtomType',
              width: 180,
            },
            {
              title: '状态',
              dataIndex: 'status',
              width: 100,
            },
            {
              title: '耗时(ms)',
              dataIndex: 'duration',
              width: 100,
            },
            {
              title: '入口节点',
              dataIndex: 'entryNode',
              width: 120,
            },
            {
              title: '出口节点',
              dataIndex: 'exitNode',
              width: 120,
            },
            {
              title: '执行时间',
              dataIndex: 'executedAt',
              width: 180,
            },
          ]}
          search={false}
          pagination={{ pageSize: 10 }}
          options={false}
          locale={{ emptyText: <Empty description="暂无执行日志数据" /> }}
        />
      </Card>

      {/* 信息原子列表 */}
      <Card title="信息原子列表">
        <Empty description="暂无信息原子数据" />
      </Card>

      {/* 图上下文查看器 */}
      <Card title="图上下文">
        <Empty description="暂无图上下文数据" />
      </Card>
    </Space>
  );
};

export default RuntimeDataTab;
