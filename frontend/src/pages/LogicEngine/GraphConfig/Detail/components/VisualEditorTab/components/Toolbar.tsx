/**
 * 工具栏组件
 * @description 提供保存、撤销、重做、自动布局等操作
 */

import {
  SaveOutlined,
  UndoOutlined,
  RedoOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  AimOutlined,
  CompressOutlined,
  ClearOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import { Button, Space, Divider, Tooltip, Badge, Popover } from 'antd';
import React from 'react';
import './Toolbar.less';

interface ToolbarProps {
  hasChanges: boolean;
  canUndo: boolean;
  canRedo: boolean;
  onSave: () => void;
  onUndo: () => void;
  onRedo: () => void;
  onAutoLayout: () => void;
  onZoomIn: () => void;
  onZoomOut: () => void;
  onFitView: () => void;
  onCenter: () => void;
  onClear: () => void;
  saving?: boolean;
}

/**
 * 工具栏组件
 */
const Toolbar: React.FC<ToolbarProps> = ({
  hasChanges,
  canUndo,
  canRedo,
  onSave,
  onUndo,
  onRedo,
  onAutoLayout,
  onZoomIn,
  onZoomOut,
  onFitView,
  onCenter,
  onClear,
  saving,
}) => {
  // 操作提示内容
  const helpContent = (
    <div className="help-content">
      <p className="help-title">画布操作说明：</p>
      <ul className="help-list">
        <li><strong>平移画布：</strong>鼠标拖拽空白区域</li>
        <li><strong>缩放画布：</strong>按住 Ctrl + 鼠标滚轮</li>
        <li><strong>拖拽节点：</strong>直接鼠标拖拽节点</li>
        <li><strong>连接节点：</strong>从节点端口拖拽到另一个节点</li>
        <li><strong>选择节点：</strong>单击节点</li>
        <li><strong>删除节点：</strong>选中后按 Delete 键</li>
        <li><strong>复制粘贴：</strong>Ctrl+C / Ctrl+V</li>
      </ul>
    </div>
  );

  return (
    <div className="editor-toolbar">
      <Space split={<Divider type="vertical" />}>
        {/* 保存操作 */}
        <Space size="small">
          <Badge dot={hasChanges} offset={[-4, 4]}>
            <Tooltip title="保存逻辑图 (Ctrl+S)">
              <Button
                type="primary"
                icon={<SaveOutlined />}
                onClick={onSave}
                loading={saving}
                disabled={!hasChanges}
              >
                保存
              </Button>
            </Tooltip>
          </Badge>
        </Space>

        {/* 撤销/重做 */}
        <Space size="small">
          <Tooltip title="撤销 (Ctrl+Z)">
            <Button icon={<UndoOutlined />} onClick={onUndo} disabled={!canUndo} />
          </Tooltip>
          <Tooltip title="重做 (Ctrl+Shift+Z)">
            <Button icon={<RedoOutlined />} onClick={onRedo} disabled={!canRedo} />
          </Tooltip>
        </Space>

        {/* 视图操作 */}
        <Space size="small">
          <Tooltip title="放大">
            <Button icon={<ZoomInOutlined />} onClick={onZoomIn} />
          </Tooltip>
          <Tooltip title="缩小">
            <Button icon={<ZoomOutOutlined />} onClick={onZoomOut} />
          </Tooltip>
          <Tooltip title="适应画布">
            <Button icon={<CompressOutlined />} onClick={onFitView} />
          </Tooltip>
          <Tooltip title="居中显示">
            <Button icon={<AimOutlined />} onClick={onCenter} />
          </Tooltip>
        </Space>

        {/* 布局操作 */}
        <Space size="small">
          <Tooltip title="自动布局">
            <Button onClick={onAutoLayout}>自动布局</Button>
          </Tooltip>
          <Tooltip title="清空画布">
            <Button icon={<ClearOutlined />} onClick={onClear} danger>
              清空
            </Button>
          </Tooltip>
        </Space>
      </Space>

      {/* 操作提示 */}
      <Popover content={helpContent} title="操作指南" trigger="hover" placement="bottomRight">
        <Button type="text" icon={<QuestionCircleOutlined />} />
      </Popover>
    </div>
  );
};

export default Toolbar;
