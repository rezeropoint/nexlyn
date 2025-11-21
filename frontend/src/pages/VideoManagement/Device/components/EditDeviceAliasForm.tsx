import type { NexlynDevice } from "@/services/video";
import {
  DrawerForm,
  type ProFormInstance,
  ProFormText,
} from "@ant-design/pro-components";
import React, { useEffect, useRef } from "react";
import type { DeviceAliasFormValues } from "../types";

interface EditDeviceAliasFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: DeviceAliasFormValues) => Promise<boolean>;
  currentDevice?: NexlynDevice;
}

/**
 * 编辑设备别名表单组件
 */
const EditDeviceAliasForm: React.FC<EditDeviceAliasFormProps> = ({
  open,
  onOpenChange,
  onFinish,
  currentDevice,
}) => {
  const formRef = useRef<ProFormInstance>();

  // 同步表单数据
  useEffect(() => {
    if (open && currentDevice && formRef.current) {
      formRef.current.setFieldsValue({
        deviceAlias: currentDevice.deviceAlias || "",
      });
    }
  }, [open, currentDevice]);

  // 处理抽屉关闭
  const handleOpenChange = (visible: boolean) => {
    if (!visible) {
      formRef.current?.resetFields();
    }
    onOpenChange(visible);
  };

  return (
    <DrawerForm<DeviceAliasFormValues>
      title={`编辑设备别名 - ${currentDevice?.name || currentDevice?.deviceId}`}
      open={open}
      onOpenChange={handleOpenChange}
      onFinish={onFinish}
      formRef={formRef}
      width={600}
      drawerProps={{ destroyOnHidden: true }}
    >
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
    </DrawerForm>
  );
};

export default EditDeviceAliasForm;
