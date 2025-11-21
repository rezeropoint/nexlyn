import { getUserOrganizations } from "@/services/organization";
import type { DeviceTag } from "@/services/video";
import { useApp } from "@/utils/appContext";
import {
  DrawerForm,
  type ProFormInstance,
  ProFormSelect,
  ProFormText,
} from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import React, { useEffect, useRef, useState } from "react";
import type { DeviceBindFormValues } from "../types";

interface DeviceBindFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: DeviceBindFormValues) => Promise<boolean>;
  availableTags: DeviceTag[];
}

/**
 * 设备绑定表单组件
 */
const DeviceBindForm: React.FC<DeviceBindFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  availableTags,
}) => {
  const formRef = useRef<ProFormInstance>();
  const { message } = useApp();
  const { initialState } = useModel("@@initialState");
  const [primaryOrg, setPrimaryOrg] = useState<API.UserOrgRelation>();
  const [loadingOrg, setLoadingOrg] = useState(false);

  // 将可用标签转换为选项格式
  const tagOptions = availableTags.map((tag) => ({
    label: tag.tagName,
    value: tag.tagName,
  }));

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
        formRef.current?.setFieldValue(
          "organizationId",
          primaryOrganization.orgId
        );
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

  // 表单打开时加载主组织
  useEffect(() => {
    if (open) {
      loadPrimaryOrganization();
    } else {
      // 关闭时重置
      setPrimaryOrg(undefined);
    }
  }, [open]);

  // 处理抽屉关闭
  const handleOpenChange = (visible: boolean) => {
    if (!visible) {
      formRef.current?.resetFields();
    }
    onOpenChange(visible);
  };

  return (
    <DrawerForm<DeviceBindFormValues>
      title="绑定设备"
      open={open}
      onOpenChange={handleOpenChange}
      onFinish={onFinish}
      width={600}
      formRef={formRef}
      drawerProps={{ destroyOnHidden: true }}
    >
      <ProFormText
        name="deviceId"
        label="设备ID"
        placeholder="请输入要绑定的设备ID"
        rules={[
          { required: true, message: "请输入设备ID" },
          {
            pattern: /^[0-9A-Fa-f]{20}$/,
            message: "设备ID格式不正确，应为20位数字或字母",
          },
        ]}
        extra="请输入准确的GB28181设备ID（20位数字或字母组合）"
      />

      <ProFormText
        name="deviceAlias"
        label="设备别名"
        placeholder="请输入设备别名"
        rules={[
          { required: true, message: "请输入设备别名" },
          { max: 50, message: "设备别名不能超过50个字符" },
        ]}
        extra="设备别名用于更好地识别设备，建议使用有意义的名称"
      />

      <ProFormSelect
        name="organizationId"
        label="所属组织"
        rules={[{ required: true, message: "未找到主组织" }]}
        extra="设备只能绑定到您的主组织，不支持次要组织或下属组织"
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

      <ProFormSelect
        name="tags"
        label="设备标签"
        placeholder="请选择设备标签（可选）"
        mode="tags"
        options={tagOptions}
        fieldProps={{
          maxTagCount: 10,
          maxTagTextLength: 20,
          showSearch: true,
          filterOption: (input, option) =>
            (option?.label ?? "").toLowerCase().includes(input.toLowerCase()),
        }}
        rules={[
          {
            validator: async (_, value) => {
              if (value && value.length > 10) {
                throw new Error("最多只能选择10个标签");
              }
              if (value?.some((tag: string) => tag.length > 20)) {
                throw new Error("标签名称不能超过20个字符");
              }
              return Promise.resolve();
            },
          },
        ]}
        extra="可以选择现有标签，也可以输入新标签名称。最多10个标签，每个标签最长20字符"
      />
    </DrawerForm>
  );
};

export default DeviceBindForm;
