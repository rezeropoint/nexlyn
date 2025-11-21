/**
 * 逻辑积木面板组件
 * @description 展示可用的逻辑积木类型，支持拖拽到画布
 */

import { SearchOutlined, AppstoreOutlined } from '@ant-design/icons';
import { getBlockSpecs } from '@/services/lynxmanager/api';
import type { BlockSpec } from '@/services/lynxmanager/types';
import { Card, Collapse, Input, Empty, Spin, message, Tag } from 'antd';
import React, { useEffect, useState } from 'react';
import './BlockPanel.less';

type BlockPanelProps = Record<string, never>;

/**
 * 逻辑积木面板组件
 */
const BlockPanel: React.FC<BlockPanelProps> = () => {
  const [blockSpecs, setBlockSpecs] = useState<BlockSpec[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');

  // 加载逻辑积木规格列表
  useEffect(() => {
    const fetchBlockSpecs = async () => {
      setLoading(true);
      try {
        const response = await getBlockSpecs();
        if (response.code === 0 && response.data?.list) {
          setBlockSpecs(response.data.list);
        } else {
          message.error('加载逻辑积木列表失败');
        }
      } catch (error) {
        message.error(`加载逻辑积木列表失败：${(error as Error).message}`);
      } finally {
        setLoading(false);
      }
    };

    fetchBlockSpecs();
  }, []);

  // 按category分组
  const groupedBlocks = React.useMemo(() => {
    const filtered = blockSpecs.filter(
      (block) =>
        !searchText ||
        block.name.toLowerCase().includes(searchText.toLowerCase()) ||
        block.blockType.toLowerCase().includes(searchText.toLowerCase()) ||
        block.category.toLowerCase().includes(searchText.toLowerCase()),
    );

    const groups: Record<string, BlockSpec[]> = {};
    filtered.forEach((block) => {
      if (!groups[block.category]) {
        groups[block.category] = [];
      }
      groups[block.category].push(block);
    });

    return groups;
  }, [blockSpecs, searchText]);

  // 拖拽开始事件
  const handleDragStart = (e: React.DragEvent, block: BlockSpec) => {
    const dragData = {
      blockType: block.blockType,
      blockVersion: block.version,
      name: block.name,
      category: block.category,
    };
    e.dataTransfer.setData('application/json', JSON.stringify(dragData));
    e.dataTransfer.effectAllowed = 'copy';
  };

  return (
    <div className="block-panel">
      <Card
        title={
          <span>
            <AppstoreOutlined /> 逻辑积木库
          </span>
        }
        size="small"
        className="block-panel-card"
      >
        <div className="search-box">
          <Input
            placeholder="搜索逻辑积木..."
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
          />
        </div>

        <Spin spinning={loading}>
          {Object.keys(groupedBlocks).length === 0 ? (
            <Empty description="暂无逻辑积木" style={{ marginTop: 24 }} />
          ) : (
            <Collapse
              defaultActiveKey={Object.keys(groupedBlocks)}
              size="small"
              className="block-category-collapse"
              items={Object.entries(groupedBlocks).map(([category, blocks]) => ({
                key: category,
                label: (
                  <span>
                    <Tag color="blue">{blocks.length}</Tag>
                    {category}
                  </span>
                ),
                children: (
                  <div className="block-list">
                    {blocks.map((block) => (
                      <div
                        key={`${block.blockType}-${block.version}`}
                        className="block-item"
                        draggable
                        onDragStart={(e) => handleDragStart(e, block)}
                      >
                        <div className="block-name">{block.name}</div>
                        <div className="block-type">{block.blockType}</div>
                        <div className="block-desc">{block.description}</div>
                      </div>
                    ))}
                  </div>
                ),
              }))}
            />
          )}
        </Spin>
      </Card>
    </div>
  );
};

export default BlockPanel;
