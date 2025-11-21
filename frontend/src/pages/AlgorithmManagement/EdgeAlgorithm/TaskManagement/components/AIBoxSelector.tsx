import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  EnvironmentOutlined,
} from "@ant-design/icons";
import { Select, Space, Tag } from "antd";
import React from "react";
import type { AIBoxDeviceOption } from "../types";
import styles from "./AIBoxSelector.less";

interface AIBoxSelectorProps {
  value?: string; // 当前选中的设备ID
  options: AIBoxDeviceOption[]; // 设备选项列表
  loading?: boolean; // 加载状态
  disabled?: boolean; // 是否禁用
  onChange?: (deviceId: string) => void; // 设备选择变化回调
}

/**
 * AI Box 设备选择器
 *
 * 功能：
 * - 下拉选择当前组织下的AI Box设备
 * - 显示设备在线状态
 * - 显示设备位置信息
 * - 支持搜索和筛选
 */
const AIBoxSelector: React.FC<AIBoxSelectorProps> = ({
  value,
  options,
  loading = false,
  disabled = false,
  onChange,
}) => {
  // 渲染选项
  const renderOption = (option: AIBoxDeviceOption) => {
    return (
      <div className={styles.optionContainer}>
        <Space>
          {/* 在线状态图标 */}
          {option.isOnline ? (
            <CheckCircleOutlined className={styles.onlineIcon} />
          ) : (
            <CloseCircleOutlined className={styles.offlineIcon} />
          )}
          {/* 设备名称 */}
          <span>
            <strong>{option.deviceAlias || option.deviceName}</strong>
            {option.deviceAlias && (
              <span className={styles.deviceHint}>({option.deviceName})</span>
            )}
          </span>
        </Space>

        {/* 位置标签 */}
        {option.location && (
          <Tag icon={<EnvironmentOutlined />} color="blue">
            {option.location}
          </Tag>
        )}
      </div>
    );
  };

  return (
    <Select
      value={value}
      onChange={onChange}
      loading={loading}
      disabled={disabled}
      placeholder="请选择AI Box设备"
      style={{ width: 360 }}
      showSearch
      optionFilterProp="label"
      filterOption={(input, option) => {
        const optionData = options.find((opt) => opt.value === option?.value);
        if (!optionData) return false;

        const searchText = input.toLowerCase();
        return (
          optionData.deviceId.toLowerCase().includes(searchText) ||
          optionData.deviceName.toLowerCase().includes(searchText) ||
          (optionData.deviceAlias?.toLowerCase().includes(searchText) ??
            false) ||
          (optionData.location?.toLowerCase().includes(searchText) ?? false)
        );
      }}
      options={options.map((option) => ({
        value: option.value,
        label: option.label,
      }))}
      optionRender={(option) => {
        const optionData = options.find((opt) => opt.value === option.value);
        return optionData ? renderOption(optionData) : null;
      }}
      notFoundContent={
        loading
          ? "加载中..."
          : options.length === 0
          ? "暂无AI Box设备"
          : "未找到匹配设备"
      }
    />
  );
};

export default AIBoxSelector;
