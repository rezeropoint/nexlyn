import type { UpdateDeviceRequest } from "@/services/iot";
import { listTags } from "@/services/iot";
import {
  DrawerForm,
  ProFormDatePicker,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
} from "@ant-design/pro-components";
import { message } from "antd";
import dayjs from "dayjs";
import React, { useEffect, useState } from "react";
import { DEVICE_STATUS_OPTIONS } from "../constants";
import type { DeviceBindingSummary } from "../types";

interface EditDeviceFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: UpdateDeviceRequest) => Promise<boolean>;
  currentRow?: DeviceBindingSummary;
}

/**
 * 编辑设备表单组件
 */
const EditDeviceForm: React.FC<EditDeviceFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentRow,
}) => {
  const [tags, setTags] = useState<any[]>([]);

  // 加载标签列表
  useEffect(() => {
    if (open) {
      loadTags();
    }
  }, [open]);

  const loadTags = async () => {
    try {
      const response = await listTags({ pageSize: 100 });
      if (response.code === 0 && response.data?.list) {
        setTags(response.data.list);
      }
    } catch (_error: any) {
      message.error("加载标签列表失败");
    }
  };

  return (
    <DrawerForm<UpdateDeviceRequest>
      title="编辑设备信息"
      open={open}
      onOpenChange={onOpenChange}
      width={600}
      onFinish={onFinish}
      drawerProps={{
        destroyOnHidden: true,
      }}
      initialValues={
        currentRow
          ? {
              ...currentRow,
              orgId: currentRow.orgId, // 包含 orgId（必填）
              installationDate: currentRow.installationDate
                ? dayjs(currentRow.installationDate)
                : undefined,
              tagIds: currentRow.tags?.map((tag) => tag.id) || [],
            }
          : undefined
      }
    >
      <ProFormText
        name="deviceId"
        label="设备ID"
        disabled
        tooltip="设备ID不可修改"
      />

      <ProFormText
        name="deviceName"
        label="设备名称"
        placeholder="请输入设备名称"
        rules={[
          { required: true, message: "请输入设备名称" },
          { max: 100, message: "设备名称不能超过100个字符" },
        ]}
      />

      {/* 隐藏的 orgId 字段（必填，用于后端验证） */}
      <ProFormText name="orgId" hidden />

      <ProFormText
        name="deviceAlias"
        label="设备别名"
        placeholder="请输入设备别名"
        rules={[{ max: 100, message: "设备别名不能超过100个字符" }]}
      />

      <ProFormText
        name="location"
        label="安装位置"
        placeholder="请输入安装位置"
        rules={[{ max: 200, message: "安装位置不能超过200个字符" }]}
      />

      <ProFormDatePicker
        name="installationDate"
        label="安装日期"
        placeholder="请选择安装日期"
        width="lg"
      />

      <ProFormSelect
        name="status"
        label="设备状态"
        placeholder="请选择设备状态"
        options={DEVICE_STATUS_OPTIONS}
      />

      <ProFormSelect
        name="tagIds"
        label="设备标签"
        placeholder="请选择设备标签（可多选）"
        mode="multiple"
        options={tags.map((tag) => ({
          label: tag.name,
          value: tag.id,
        }))}
        fieldProps={{
          maxTagCount: "responsive",
          showSearch: true,
          optionFilterProp: "label",
        }}
        tooltip="为设备添加标签，方便分类管理"
      />

      <ProFormTextArea
        name="description"
        label="设备描述"
        placeholder="请输入设备描述"
        fieldProps={{
          rows: 4,
          maxLength: 500,
          showCount: true,
        }}
      />
    </DrawerForm>
  );
};

export default EditDeviceForm;
