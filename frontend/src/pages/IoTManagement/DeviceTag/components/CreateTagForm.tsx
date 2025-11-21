import type { CreateTagRequest } from "@/services/iot";
import {
  DrawerForm,
  ProFormText,
  ProFormTextArea,
  type ProFormInstance,
} from "@ant-design/pro-components";
import { ColorPicker } from "antd";
import React, { useRef, useState } from "react";
import { DEFAULT_COLOR, PRESET_COLORS } from "../constants";
import styles from "./CreateTagForm.less";

interface CreateTagFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onFinish: (values: CreateTagRequest) => Promise<boolean>;
}

/**
 * 创建标签表单组件
 */
const CreateTagForm: React.FC<CreateTagFormProps> = ({
  open,
  onOpenChange,
  onFinish,
}) => {
  const formRef = useRef<ProFormInstance>();
  const [selectedColor, setSelectedColor] = useState<string>(DEFAULT_COLOR);

  return (
    <DrawerForm<CreateTagRequest>
      formRef={formRef}
      title="创建标签"
      open={open}
      onOpenChange={(visible) => {
        if (!visible) {
          // 关闭时重置（顺序：formRef → state → callback）
          formRef.current?.resetFields();
          setSelectedColor(DEFAULT_COLOR);
        }
        onOpenChange(visible);
      }}
      width={500}
      onFinish={async (values) => {
        const result = await onFinish({
          ...values,
          color: selectedColor,
        });
        return result;
      }}
      drawerProps={{
        destroyOnHidden: true,
      }}
    >
      <ProFormText
        name="name"
        label="标签名称"
        placeholder="请输入标签名称"
        rules={[
          { required: true, message: "请输入标签名称" },
          { max: 50, message: "标签名称不能超过50个字符" },
        ]}
      />

      <ProFormText
        name="color"
        label="标签颜色"
        tooltip="选择标签颜色,用于在列表中区分不同标签"
      >
        <div className={styles.colorPickerWrapper}>
          <ColorPicker
            value={selectedColor}
            onChange={(color) => {
              setSelectedColor(color.toHexString());
            }}
            presets={[
              {
                label: "预设颜色",
                colors: PRESET_COLORS,
              },
            ]}
            showText
          />
          <span className={styles.colorHint}>当前颜色: {selectedColor}</span>
        </div>
      </ProFormText>

      <ProFormTextArea
        name="description"
        label="标签描述"
        placeholder="请输入标签描述"
        fieldProps={{
          rows: 4,
          maxLength: 200,
          showCount: true,
        }}
      />
    </DrawerForm>
  );
};

export default CreateTagForm;
