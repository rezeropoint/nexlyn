import type { NexlynDevice } from "@/services/video";
import {
  getDeviceStatistics,
  getNexlynDeviceList,
  removeDevice,
  syncDevice,
} from "@/services/video";
import {
  DeleteOutlined,
  EditOutlined,
  LinkOutlined,
  MoreOutlined,
  SyncOutlined,
  TagsOutlined,
} from "@ant-design/icons";
import type { ActionType, ProColumns } from "@ant-design/pro-components";
import { ProTable } from "@ant-design/pro-components";
import type { MenuProps } from "antd";
import { Button, Dropdown, message, Modal, Space, Tag, Tooltip } from "antd";
import React from "react";
import type { DeviceActionConfig, DeviceBindingFilter } from "../Device/types";
import styles from "./DeviceTable.less";
import OrganizationSelector from "./OrganizationSelector";

export interface DeviceSummary {
  deviceTotal: number;
  deviceOnline: number;
  deviceOffline: number;
}

interface DeviceTableProps {
  actionRef?: React.MutableRefObject<ActionType | undefined>;
  onViewChannels?: (device: NexlynDevice) => void; // 改为可选，用于兼容
  onRowClick?: (device: NexlynDevice) => void; // 新增行点击回调
  selectedDeviceId?: string; // 当前选中的设备ID
  onSummaryChange?: (summary: DeviceSummary) => void;
  onEdit?: (device: NexlynDevice) => void;
  onEditAlias?: (device: NexlynDevice) => void;
  onEditTags?: (device: NexlynDevice) => void;
  onUnbind?: (device: NexlynDevice) => void;
  bindingFilter: DeviceBindingFilter;
  actionConfig?: DeviceActionConfig;
  // 工具栏相关
  onBindDevice?: () => void;
  bindLoading?: boolean;
  showBindButton?: boolean;
  // 筛选器相关
  onBindingFilterChange?: (filter: DeviceBindingFilter) => void;
  bindingFilterOptions?: Array<{ label: string; value: DeviceBindingFilter }>;
  // 组织架构筛选
  organizationId?: string;
  onOrganizationChange?: (organizationId: string | undefined) => void;
  hideOrgSelector?: boolean; // 是否隐藏组织架构选择器
}

const getDeviceStatusColor = (record: NexlynDevice) => {
  // 检查在线状态，支持多种字段格式
  const isOnline = record.online || record.status === "ON";
  if (isOnline) {
    return "green";
  } else {
    return "red";
  }
};

const calculateDeviceSummary = (devices: NexlynDevice[]): DeviceSummary => {
  const deviceTotal = devices.length;
  let deviceOnline = 0;
  let deviceOffline = 0;

  devices.forEach((d) => {
    if (d.online || d.status === "ON") {
      deviceOnline++;
    } else {
      deviceOffline++;
    }
  });

  return { deviceTotal, deviceOnline, deviceOffline };
};

const DeviceTable: React.FC<DeviceTableProps> = ({
  actionRef,
  onRowClick,
  selectedDeviceId,
  onSummaryChange,
  onEdit,
  onEditAlias,
  onEditTags,
  onUnbind,
  bindingFilter,
  actionConfig = {
    showBind: false,
    showUnbind: true,
    showEdit: true,
    showEditAlias: false, // 默认使用统一编辑
    showEditTags: false, // 默认使用统一编辑
    showSync: true,
    showDelete: true,
    showChannels: true,
  },
  // 工具栏相关
  onBindDevice,
  bindLoading = false,
  showBindButton = false,
  // 筛选器相关
  onBindingFilterChange,
  bindingFilterOptions = [],
  // 组织架构筛选
  organizationId,
  onOrganizationChange,
  hideOrgSelector = false,
}) => {
  const handleSyncDevice = async (device: NexlynDevice) => {
    try {
      const deviceId = device.deviceId || device.device_id;
      if (!deviceId) {
        message.error("设备ID不能为空");
        return;
      }
      const res = await syncDevice(deviceId);
      message.success(res.message || "设备同步成功");
      actionRef?.current?.reload();
    } catch (_error) {
      message.error("设备同步失败");
    }
  };

  const handleRemoveDevice = async (device: NexlynDevice) => {
    if (device.online || device.status === "ON") {
      message.error("只能删除离线设备");
      return;
    }

    Modal.confirm({
      title: "确定要删除此设备吗？",
      content: "删除后无法恢复，请谨慎操作",
      okText: "确定删除",
      cancelText: "取消",
      okType: "danger",
      onOk: async () => {
        try {
          const deviceId = device.deviceId || device.device_id;
          if (!deviceId) {
            message.error("设备ID不能为空");
            return;
          }
          const res = await removeDevice(deviceId);
          message.success(res.message || "设备删除成功");
          actionRef?.current?.reload();
        } catch (_error) {
          message.error("设备删除失败");
        }
      },
    });
  };

  const handleUnbindDevice = async (device: NexlynDevice) => {
    Modal.confirm({
      title: "确定要解绑此设备吗？",
      content: "解绑后设备将变为未绑定状态，别名和标签将被清除",
      okText: "确定解绑",
      cancelText: "取消",
      okType: "primary",
      onOk: async () => {
        if (onUnbind) {
          await onUnbind(device);
          actionRef?.current?.reload();
        }
      },
    });
  };

  const columns: ProColumns<NexlynDevice>[] = [
    {
      title: "设备ID",
      dataIndex: "deviceId",
      key: "deviceId",
      width: 160,
      copyable: true,
      search: false,
      ellipsis: true,
      render: (_, record) => record.deviceId || record.device_id,
    },
    {
      title: "设备名称",
      dataIndex: "name",
      key: "name",
      ellipsis: true,
      search: false,
    },
    {
      title: "设备别名",
      dataIndex: "deviceAlias",
      key: "deviceAlias",
      ellipsis: true,
      search: false,
      render: (_, record) => {
        const alias = record.deviceAlias || record.device_alias;
        return (
          <span>
            {alias || "-"}
            {!alias && (
              <Tooltip title="设备未设置别名">
                <Tag color="orange" className={styles.unboundTag}>
                  未绑定
                </Tag>
              </Tooltip>
            )}
          </span>
        );
      },
    },
    {
      title: "制造商",
      dataIndex: "manufacturer",
      key: "manufacturer",
      width: 100,
      ellipsis: true,
      search: false,
    },
    {
      title: "设备状态",
      dataIndex: "status",
      key: "deviceStatus",
      width: 80,
      valueType: "select",
      valueEnum: {
        "": { text: "全部" },
        ON: { text: "在线设备", status: "Success" },
        OFF: { text: "离线设备", status: "Error" },
      },
      render: (_, record) => (
        <Tag color={getDeviceStatusColor(record)}>
          {record.online || record.status === "ON" ? "在线" : "离线"}
        </Tag>
      ),
    },
    {
      title: "设备标签",
      dataIndex: "tags",
      key: "tags",
      ellipsis: true,
      search: false,
      render: (_, record) => (
        <Space size={[4, 4]} wrap>
          {record.tags && record.tags.length > 0 ? (
            record.tags.map((tag) => (
              <Tag key={tag.id} color="blue">
                {tag.name}
              </Tag>
            ))
          ) : (
            <span className={styles.noTagsText}>无标签</span>
          )}
        </Space>
      ),
    },
    // 组织架构筛选项（可通过hideOrgSelector隐藏）
    ...(!hideOrgSelector
      ? [
          {
            title: "组织架构",
            dataIndex: "organizationId",
            key: "organizationId",
            hideInTable: true,
            order: 100, // 设置为最高优先级，显示在第一位
            renderFormItem: () => (
              <OrganizationSelector
                value={organizationId}
                onChange={onOrganizationChange}
                placeholder="选择组织架构"
                allowClear
              />
            ),
          },
        ]
      : []),
    {
      title: "搜索关键词",
      dataIndex: "keyword",
      key: "keyword",
      hideInTable: true,
      order: 90,
      tooltip: "可搜索设备ID、设备名称或设备别名",
    },
    {
      title: "设备状态",
      dataIndex: "deviceStatus",
      key: "deviceStatusFilter",
      hideInTable: true,
      order: 80,
      valueType: "select",
      valueEnum: {
        "": { text: "全部" },
        ON: { text: "在线设备", status: "Success" },
        OFF: { text: "离线设备", status: "Error" },
      },
    },
    {
      title: "绑定状态",
      dataIndex: "bindingStatus",
      key: "bindingStatus",
      hideInTable: true,
      order: 70,
      valueType: "select",
      valueEnum:
        bindingFilterOptions.length > 0
          ? bindingFilterOptions.reduce((acc, option) => {
              acc[option.value] = { text: option.label };
              return acc;
            }, {} as Record<string, { text: string }>)
          : {
              bound: { text: "已绑定设备" },
              unbound: { text: "未绑定设备" },
              all: { text: "全部设备" },
            },
      fieldProps: {
        value: bindingFilter,
        onChange: (value: DeviceBindingFilter) => {
          onBindingFilterChange?.(value);
        },
      },
    },
    {
      title: "操作",
      valueType: "option",
      key: "option",
      width: 140,
      fixed: "right",
      search: false,
      render: (_, record) => {
        const actions = [];

        // 同步按钮 - 常用操作，放在外面
        if (actionConfig.showSync) {
          actions.push(
            <Button
              key="sync"
              type="link"
              size="small"
              icon={<SyncOutlined />}
              onClick={() => handleSyncDevice(record)}
            >
              同步
            </Button>
          );
        }

        // 如果操作按钮较多，将部分按钮放入下拉菜单
        const moreItems: MenuProps["items"] = [];

        // 编辑按钮 - 放入更多菜单
        if (actionConfig.showEdit && onEdit) {
          moreItems.push({
            key: "edit",
            icon: <EditOutlined />,
            label: "编辑设备",
            onClick: () => onEdit(record),
          });
        } else {
          // 如果不使用统一编辑，则显示分别的编辑选项（兼容性）
          if (actionConfig.showEditAlias && onEditAlias) {
            moreItems.push({
              key: "editAlias",
              icon: <EditOutlined />,
              label: "编辑别名",
              onClick: () => onEditAlias(record),
            });
          }

          if (actionConfig.showEditTags && onEditTags) {
            moreItems.push({
              key: "editTags",
              icon: <TagsOutlined />,
              label: "编辑标签",
              onClick: () => onEditTags(record),
            });
          }
        }

        // 解绑按钮 - 放入更多菜单
        if (
          actionConfig.showUnbind &&
          (record.deviceAlias || record.device_alias) &&
          onUnbind
        ) {
          moreItems.push({
            key: "unbind",
            icon: <LinkOutlined />,
            label: "解绑设备",
            onClick: () => handleUnbindDevice(record),
          });
        }

        // 删除按钮 - 也需要特殊处理确认对话框
        if (actionConfig.showDelete) {
          moreItems.push({
            key: "delete",
            icon: <DeleteOutlined />,
            label: "删除设备",
            disabled: record.online || record.status === "ON",
            onClick: () => {
              if (record.online || record.status === "ON") return;
              handleRemoveDevice(record); // 直接调用，让函数自己处理确认逻辑
            },
            danger: true,
          });
        }

        // 如果有更多操作，添加下拉菜单
        if (moreItems.length > 0) {
          actions.push(
            <Dropdown
              key="more"
              menu={{ items: moreItems }}
              trigger={["click"]}
            >
              <Button type="link" size="small" icon={<MoreOutlined />}>
                更多
              </Button>
            </Dropdown>
          );
        }

        return <Space size={4}>{actions}</Space>;
      },
    },
  ];

  return (
    <ProTable<NexlynDevice>
      headerTitle="设备列表"
      actionRef={actionRef as any}
      rowKey={(record) =>
        record.deviceId || record.device_id || record.id || ""
      }
      search={{
        labelWidth: "auto",
        defaultCollapsed: true, // 默认收起搜索表单
      }}
      form={{
        initialValues: {
          organizationId: organizationId,
          bindingStatus: bindingFilter,
        },
      }}
      params={{
        organizationIds: organizationId ? [organizationId] : undefined,
        bindingFilter: bindingFilter,
      }}
      cardBordered={false}
      scroll={{ x: "max-content" }}
      toolBarRender={() => [
        showBindButton && (
          <Button
            key="bind"
            type="primary"
            onClick={onBindDevice}
            loading={bindLoading}
          >
            绑定设备
          </Button>
        ),
      ]}
      request={async (params) => {
        try {
          // 如果组织ID未加载完成，返回空结果（避免首次无效请求）
          if (organizationId === undefined) {
            if (onSummaryChange)
              onSummaryChange({
                deviceTotal: 0,
                deviceOnline: 0,
                deviceOffline: 0,
              });
            return { data: [], success: true, total: 0 };
          }

          const apiParams: any = {
            page: params.current || 1,
            pageSize: params.pageSize || 20,
          };
          if (params.keyword) apiParams.keyword = params.keyword;
          if ((params as any).deviceStatus)
            apiParams.status = (params as any).deviceStatus;

          // 根据绑定状态筛选设置bindStatus参数
          if (bindingFilter === "bound") {
            apiParams.bindStatus = "bound";
          } else if (bindingFilter === "unbound") {
            apiParams.bindStatus = "unbound";
          }
          // bindingFilter === 'all' 时不设置bindStatus参数，获取所有设备

          // 组织架构筛选 - 使用单个organizationId（此时已确保不为undefined）
          apiParams.organizationId = organizationId;

          const response = await getNexlynDeviceList(apiParams, {
            skipErrorHandler: true,
          });
          if (response.code === 0) {
            const devices = response.data.devices || [];

            // 调用统计接口获取准确的统计数据（全量数据，非分页）
            if (onSummaryChange && organizationId) {
              try {
                const statsResponse = await getDeviceStatistics({
                  organizationId,
                });
                if (statsResponse.code === 0 && statsResponse.data) {
                  onSummaryChange({
                    deviceTotal: statsResponse.data.deviceTotal,
                    deviceOnline: statsResponse.data.deviceOnline,
                    deviceOffline: statsResponse.data.deviceOffline,
                  });
                } else {
                  // 如果统计接口失败，降级使用列表数据计算
                  onSummaryChange(calculateDeviceSummary(devices));
                }
              } catch (error) {
                // 统计接口异常，降级使用列表数据计算
                console.error("获取设备统计失败:", error);
                onSummaryChange(calculateDeviceSummary(devices));
              }
            }

            return {
              data: devices,
              success: true,
              total: response.data.total || devices.length,
            };
          }
          // 业务错误
          const errorMsg = response.message || "获取设备列表失败";
          message.error(errorMsg);
          if (onSummaryChange)
            onSummaryChange({
              deviceTotal: 0,
              deviceOnline: 0,
              deviceOffline: 0,
            });
          return { data: [], success: false, total: 0 };
        } catch (error: any) {
          // 网络错误或其他异常
          const errorMsg = error?.message || "网络请求失败，请检查网络连接";
          message.error(errorMsg);
          if (onSummaryChange)
            onSummaryChange({
              deviceTotal: 0,
              deviceOnline: 0,
              deviceOffline: 0,
            });
          return { data: [], success: false, total: 0 };
        }
      }}
      columns={columns}
      pagination={{
        pageSize: 20,
        showSizeChanger: true,
        showQuickJumper: true,
      }}
      onRow={(record) => ({
        onClick: () => {
          if (onRowClick) {
            onRowClick(record);
          }
        },
        className: onRowClick ? styles.clickableRow : undefined,
      })}
      rowClassName={(record) => {
        const currentDeviceId = record.deviceId || record.device_id;
        return selectedDeviceId === currentDeviceId
          ? "ant-table-row-selected"
          : "";
      }}
    />
  );
};

export default DeviceTable;
