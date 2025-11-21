import type { PermissionItem } from "@/services/permission/role";
import { InfoCircleOutlined } from "@ant-design/icons";
import { Card, Space, Tag, Tooltip, Transfer, Typography } from "antd";
import React, { useEffect, useState } from "react";
import styles from "./PermissionSelector.less";

const { Text } = Typography;

interface PermissionSelectorProps {
  value?: string[];
  onChange?: (value: string[]) => void;
  availablePermissions: PermissionItem[];
  disabled?: boolean;
}

interface TransferItem {
  key: string;
  title: string;
  description: string;
  category: string;
  disabled?: boolean;
}

const PermissionSelector: React.FC<PermissionSelectorProps> = ({
  value = [],
  onChange,
  availablePermissions,
  disabled = false,
}) => {
  const [targetKeys, setTargetKeys] = useState<React.Key[]>(value);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);

  useEffect(() => {
    setTargetKeys(value);
  }, [value]);

  // 将权限项转换为Transfer组件需要的格式
  const dataSource: TransferItem[] = availablePermissions.map((item) => ({
    key: `${item.resource}:${item.action}`,
    title: item.description,
    description: `${item.resource}:${item.action}`,
    category: item.category,
  }));

  // 按分类分组权限
  const groupedPermissions = availablePermissions.reduce((groups, item) => {
    if (!groups[item.category]) {
      groups[item.category] = [];
    }
    groups[item.category].push(item);
    return groups;
  }, {} as Record<string, PermissionItem[]>);

  const handleChange = (
    newTargetKeys: React.Key[],
    _direction: string,
    _moveKeys: React.Key[]
  ) => {
    setTargetKeys(newTargetKeys);
    onChange?.(newTargetKeys as string[]);
  };

  const handleSelectChange = (
    sourceSelectedKeys: React.Key[],
    targetSelectedKeys: React.Key[]
  ) => {
    setSelectedKeys([...sourceSelectedKeys, ...targetSelectedKeys]);
  };

  // 自定义渲染项目
  const renderItem = (item: TransferItem) => {
    const customLabel = (
      <Space direction="vertical" size="small" style={{ width: "100%" }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <Text strong>{item.title}</Text>
          <Tooltip title={`资源权限: ${item.description}`}>
            <InfoCircleOutlined className={styles.infoIcon} />
          </Tooltip>
        </div>
        <div>
          <Tag color="blue">{item.category}</Tag>
          <Text type="secondary" style={{ fontSize: "12px" }}>
            {item.description}
          </Text>
        </div>
      </Space>
    );

    return {
      label: customLabel,
      value: item.key,
    };
  };

  // 渲染权限分类统计
  const renderFooter = (props: any) => {
    const { direction } = props;
    if (direction === "right") {
      const selectedCategories = targetKeys.reduce((categories, key) => {
        const permission = availablePermissions.find(
          (item) => `${item.resource}:${item.action}` === String(key)
        );
        if (permission) {
          categories[permission.category] =
            (categories[permission.category] || 0) + 1;
        }
        return categories;
      }, {} as Record<string, number>);

      return (
        <div className={styles.selectedStats}>
          <Text type="secondary" className={styles.selectedStatsText}>
            已选择 {targetKeys.length} 个权限
          </Text>
          <div style={{ marginTop: "4px" }}>
            {Object.entries(selectedCategories).map(([category, count]) => (
              <Tag key={category} style={{ margin: "2px" }}>
                {category}: {count}
              </Tag>
            ))}
          </div>
        </div>
      );
    }
    return null;
  };

  return (
    <div>
      {/* 权限分类概览 */}
      <Card size="small" style={{ marginBottom: "16px" }}>
        <Text strong>权限分类概览:</Text>
        <div style={{ marginTop: "8px" }}>
          {Object.entries(groupedPermissions).map(([category, permissions]) => (
            <Tag key={category} color="default" style={{ margin: "2px" }}>
              {category} ({permissions.length}个)
            </Tag>
          ))}
        </div>
      </Card>

      {/* 权限选择器 */}
      <Transfer
        dataSource={dataSource}
        targetKeys={targetKeys}
        selectedKeys={selectedKeys}
        onChange={handleChange}
        onSelectChange={handleSelectChange}
        render={renderItem}
        footer={renderFooter}
        titles={["可用权限", "已选权限"]}
        showSearch={!disabled}
        disabled={disabled}
        filterOption={(inputValue, item) =>
          item.title?.toLowerCase().includes(inputValue.toLowerCase()) ||
          item.description?.toLowerCase().includes(inputValue.toLowerCase()) ||
          item.category?.toLowerCase().includes(inputValue.toLowerCase())
        }
        oneWay={false}
        style={{ width: "100%" }}
        listStyle={{
          width: "45%",
          height: "400px",
        }}
      />

      {/* 选择统计 */}
      <div style={{ marginTop: "16px", textAlign: "center" }}>
        <Text type="secondary">
          共 {availablePermissions.length} 个可用权限，已选择{" "}
          {targetKeys.length} 个
        </Text>
      </div>
    </div>
  );
};

export default PermissionSelector;
