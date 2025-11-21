import type { BindDeviceRequest } from "@/services/iot";
import {
  getSensorTemplateList,
  listTags,
  listUnboundDevices,
} from "@/services/iot";
import { getUserOrganizations } from "@/services/organization";
import {
  DrawerForm,
  ProFormDatePicker,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  type ProFormInstance,
} from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import { message } from "antd";
import React, { useEffect, useRef, useState } from "react";
import { DEVICE_STATUS_OPTIONS } from "../constants";

interface BindDeviceFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: BindDeviceRequest) => Promise<boolean>;
}

/**
 * 绑定设备表单组件
 */
const BindDeviceForm: React.FC<BindDeviceFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { initialState } = useModel("@@initialState");
  const [unboundDevices, setUnboundDevices] = useState<any[]>([]);
  const [templates, setTemplates] = useState<any[]>([]);
  const [filteredTemplates, setFilteredTemplates] = useState<any[]>([]);
  const [selectedDeviceCategory, setSelectedDeviceCategory] =
    useState<string>("");
  const [primaryOrg, setPrimaryOrg] = useState<API.UserOrgRelation>();
  const [loadingOrg, setLoadingOrg] = useState(false);
  const [tags, setTags] = useState<any[]>([]);

  // 加载未绑定设备列表
  useEffect(() => {
    if (open) {
      loadUnboundDevices();
      loadTemplates();
      loadPrimaryOrganization();
      loadTags();
    } else {
      // 关闭时重置
      setPrimaryOrg(undefined);
      setSelectedDeviceCategory("");
      setFilteredTemplates([]);
      formRef.current?.resetFields();
    }
  }, [open]);

  const loadUnboundDevices = async () => {
    try {
      const response = await listUnboundDevices();
      if (response.code === 0 && response.data?.list) {
        setUnboundDevices(response.data.list);
      }
    } catch (_error: any) {
      message.error("加载未绑定设备失败");
    }
  };

  const loadTemplates = async () => {
    try {
      const response = await getSensorTemplateList({ enabled: "true" });
      if (response.code === 0 && response.data?.list) {
        setTemplates(response.data.list);
      }
    } catch (_error: any) {
      message.error("加载设备模板失败");
    }
  };

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

  // 获取当前用户的主组织
  const loadPrimaryOrganization = async () => {
    setLoadingOrg(true);
    try {
      // 从 initialState 获取当前用户信息
      const currentUser = initialState?.currentUser;
      if (!currentUser) {
        message.error("未找到用户信息，请重新登录");
        return;
      }

      const userId = currentUser.id;
      if (!userId) {
        message.error("用户ID不存在，请重新登录");
        return;
      }

      // 获取用户的主组织关系
      const response = await getUserOrganizations({
        id: userId,
        isPrimary: true,
      });

      if (
        response.code === 0 &&
        response.data?.list &&
        response.data.list.length > 0
      ) {
        const primaryOrganization = response.data.list[0];
        setPrimaryOrg(primaryOrganization);
        formRef.current?.setFieldValue("orgId", primaryOrganization.orgId);
      } else {
        message.warning("未找到您的主组织，请联系管理员配置");
      }
    } catch (error) {
      console.error("获取主组织失败:", error);
      message.error("获取主组织失败");
    } finally {
      setLoadingOrg(false);
    }
  };

  // 处理设备选择变化
  const handleDeviceChange = (deviceId: string) => {
    // 查找选中的设备
    const selectedDevice = unboundDevices.find(
      (device) => device.deviceId === deviceId
    );

    if (selectedDevice?.category) {
      // 设置选中设备的类别
      setSelectedDeviceCategory(selectedDevice.category);

      // 根据类别过滤模板
      const matchedTemplates = templates.filter(
        (template) => template.category === selectedDevice.category
      );
      setFilteredTemplates(matchedTemplates);

      // 如果只有一个匹配的模板，自动选择
      if (matchedTemplates.length === 1) {
        formRef.current?.setFieldValue(
          "deviceModel",
          matchedTemplates[0].model
        );
        message.success(`已自动选择设备型号：${matchedTemplates[0].name}`);
      } else if (matchedTemplates.length === 0) {
        formRef.current?.setFieldValue("deviceModel", undefined);
        message.warning(
          `未找到适用于 ${selectedDevice.category} 类别的设备型号`
        );
      } else {
        formRef.current?.setFieldValue("deviceModel", undefined);
      }
    } else {
      // 未找到设备类别，清空过滤
      setSelectedDeviceCategory("");
      setFilteredTemplates([]);
      formRef.current?.setFieldValue("deviceModel", undefined);
    }
  };

  return (
    <DrawerForm<BindDeviceRequest>
      title="绑定设备"
      open={open}
      onOpenChange={onOpenChange}
      width={600}
      onFinish={onFinish}
      formRef={formRef}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormSelect
        name="deviceId"
        label="选择设备"
        placeholder="请选择未绑定的设备"
        rules={[{ required: true, message: "请选择设备" }]}
        options={unboundDevices.map((device) => ({
          label: `${device.deviceId} ${device.isOnline ? "(在线)" : "(离线)"}`,
          value: device.deviceId,
        }))}
        showSearch
        fieldProps={{
          notFoundContent:
            unboundDevices.length === 0 ? "暂无未绑定设备" : undefined,
          onChange: handleDeviceChange,
        }}
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

      <ProFormText
        name="deviceAlias"
        label="设备别名"
        placeholder="请输入设备别名"
        rules={[{ max: 100, message: "设备别名不能超过100个字符" }]}
      />

      <ProFormSelect
        name="deviceModel"
        label="设备型号"
        placeholder={
          selectedDeviceCategory
            ? `请选择设备型号（已根据类别 ${selectedDeviceCategory} 过滤）`
            : "请先选择设备，系统会自动过滤匹配的型号"
        }
        rules={[{ required: true, message: "请选择设备型号" }]}
        options={(selectedDeviceCategory ? filteredTemplates : templates).map(
          (template) => ({
            label: `${template.name} (${template.model})`,
            value: template.model,
          })
        )}
        showSearch
        tooltip={
          selectedDeviceCategory
            ? `已根据设备类别 ${selectedDeviceCategory} 过滤可用型号`
            : "选择设备后将自动过滤匹配的设备型号"
        }
      />

      <ProFormSelect
        name="orgId"
        label="所属组织"
        rules={[{ required: true, message: "未找到主组织" }]}
        extra="设备只能绑定到您的主组织"
        placeholder={loadingOrg ? "加载中..." : "您的主组织"}
        disabled
        fieldProps={{
          loading: loadingOrg,
        }}
        options={
          primaryOrg
            ? [
                {
                  label: primaryOrg.orgName,
                  value: primaryOrg.orgId,
                },
              ]
            : []
        }
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
        initialValue="active"
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

export default BindDeviceForm;
