/**
 * 逻辑图配置卡片视图容器组件
 */

import type {
  GetGraphConfigListRequest,
  GraphConfigMetadata,
} from '@/services/lynxmanager';
import { listGraphConfigs } from '@/services/lynxmanager';
import { useModel } from '@@/exports';
import { ColumnHeightOutlined, PlusOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { type ActionType, ProCard, ProForm, ProFormSelect, ProFormText } from '@ant-design/pro-components';
import { App, Button, Col, Dropdown, List, Pagination, Row, Space, Spin } from 'antd';
import type { MenuProps } from 'antd';
import React, { useCallback, useEffect, useState } from 'react';
import { TABLE_CONFIG } from '../constants';
import GraphConfigCard from './GraphConfigCard';
import styles from './GraphConfigCardView.less';

interface GraphConfigCardViewProps {
  actionRef: React.MutableRefObject<ActionType | undefined>;
  onView: (record: GraphConfigMetadata) => void;
  onCopy: (record: GraphConfigMetadata) => void;
  onDelete: (record: GraphConfigMetadata) => void;
  onCreateConfig: () => void;
}

/**
 * 卡片视图容器
 */
const GraphConfigCardView: React.FC<GraphConfigCardViewProps> = ({
  actionRef,
  onView,
  onCopy,
  onDelete,
  onCreateConfig,
}) => {
  const { message } = App.useApp();
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;

  const [loading, setLoading] = useState(false);
  const [dataSource, setDataSource] = useState<GraphConfigMetadata[]>([]);
  const [total, setTotal] = useState(0);
  const [current, setCurrent] = useState(1);
  const [pageSize, setPageSize] = useState(TABLE_CONFIG.DEFAULT_PAGE_SIZE);
  const [searchParams, setSearchParams] = useState<any>({});
  const [density, setDensity] = useState<'loose' | 'middle' | 'compact'>('middle');

  // 密度配置
  const densityConfig = {
    loose: { gutter: 24, cardPadding: 24 },
    middle: { gutter: 16, cardPadding: 16 },
    compact: { gutter: 12, cardPadding: 12 },
  };

  // 密度菜单项
  const densityMenuItems: MenuProps['items'] = [
    {
      key: 'loose',
      label: '宽松',
      onClick: () => setDensity('loose'),
    },
    {
      key: 'middle',
      label: '中等',
      onClick: () => setDensity('middle'),
    },
    {
      key: 'compact',
      label: '紧凑',
      onClick: () => setDensity('compact'),
    },
  ];

  // 加载数据
  const loadData = async (params: GetGraphConfigListRequest) => {
    try {
      setLoading(true);
      const tenantId = currentUser?.tenantInfo?.tenantId;
      if (!tenantId) {
        setDataSource([]);
        setTotal(0);
        return;
      }

      const response = await listGraphConfigs({
        ...params,
        tenantId,
        page: params.page || current,
        pageSize: params.pageSize || pageSize,
      });

      if (response.code === 0 && response.data) {
        setDataSource(response.data.list || []);
        setTotal(response.total || 0);
      } else {
        message.error(response.msg || response.message || '获取逻辑图配置列表失败');
        setDataSource([]);
        setTotal(0);
      }
    } catch (error) {
      console.error('获取逻辑图配置列表失败:', error);
      message.error('获取逻辑图配置列表失败');
      setDataSource([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  };

  // 处理搜索
  const handleSearch = async (values: any) => {
    setSearchParams(values);
    setCurrent(1);
    await loadData({ ...values, page: 1, pageSize });
  };

  // 处理分页变化
  const handlePageChange = async (page: number, size: number) => {
    setCurrent(page);
    setPageSize(size);
    await loadData({ ...searchParams, page, pageSize: size });
  };

  // 手动刷新
  const handleRefresh = useCallback(async () => {
    const tenantId = currentUser?.tenantInfo?.tenantId;
    if (tenantId) {
      await loadData({ ...searchParams, page: current, pageSize, tenantId });
    }
  }, [currentUser?.tenantInfo?.tenantId, searchParams, current, pageSize]);

  // 暴露 reload 方法给父组件
  React.useImperativeHandle(actionRef, () => ({
    reload: async () => {
      await handleRefresh();
    },
    reloadAndRest: async () => {
      await handleRefresh();
    },
    reset: async () => {
      const tenantId = currentUser?.tenantInfo?.tenantId;
      if (tenantId) {
        setSearchParams({});
        setCurrent(1);
        await loadData({ page: 1, pageSize, tenantId });
      }
    },
  } as any), [handleRefresh, currentUser?.tenantInfo?.tenantId, pageSize]);

  // 初始加载
  useEffect(() => {
    const tenantId = currentUser?.tenantInfo?.tenantId;
    if (tenantId) {
      loadData({ page: 1, pageSize, tenantId });
    }
  }, [currentUser?.tenantInfo?.tenantId]);

  return (
    <div className={styles.cardViewContainer}>
      {/* 搜索区域 */}
      <ProCard className={styles.searchCard}>
        <div className={styles.searchFormWrapper}>
          <ProForm
            layout="inline"
            onFinish={handleSearch}
            submitter={{
              searchConfig: {
                submitText: '查询',
                resetText: '重置',
              },
              submitButtonProps: {
                icon: <SearchOutlined />,
              },
            }}
          >
            <ProFormText
              name="name"
              label="图名称"
              placeholder="请输入图名称"
              fieldProps={{
                allowClear: true,
              }}
            />
            <ProFormSelect
              name="isEnabled"
              label="启用状态"
              placeholder="请选择启用状态"
              options={[
                { label: '已启用', value: true },
                { label: '未启用', value: false },
              ]}
              fieldProps={{
                allowClear: true,
              }}
            />
          </ProForm>
          <Space className={styles.extraActions}>
            <Button
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
            />
            <Dropdown
              menu={{ items: densityMenuItems, selectable: true, selectedKeys: [density] }}
              trigger={['click']}
            >
              <Button icon={<ColumnHeightOutlined />} />
            </Dropdown>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={onCreateConfig}
            >
              新建
            </Button>
          </Space>
        </div>
      </ProCard>

      {/* 卡片列表 */}
      <Spin spinning={loading}>
        <List
          grid={{
            gutter: densityConfig[density].gutter,
            xs: 1,
            sm: 2,
            md: 2,
            lg: 3,
            xl: 4,
            xxl: 4,
          }}
          dataSource={dataSource}
          locale={{ emptyText: '暂无数据' }}
          renderItem={(item) => (
            <List.Item style={{ marginBottom: densityConfig[density].gutter }}>
              <GraphConfigCard
                data={item}
                onView={onView}
                onCopy={onCopy}
                onDelete={onDelete}
                bodyPadding={densityConfig[density].cardPadding}
              />
            </List.Item>
          )}
        />

        {/* 分页 */}
        {total > 0 && (
          <Row justify="end" className={styles.paginationRow}>
            <Col>
              <Pagination
                current={current}
                pageSize={pageSize}
                total={total}
                showQuickJumper
                showSizeChanger
                showTotal={(total) => `共 ${total} 条`}
                onChange={handlePageChange}
              />
            </Col>
          </Row>
        )}
      </Spin>
    </div>
  );
};

export default GraphConfigCardView;
