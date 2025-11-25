/**
 * 审批操作 Modal 组件
 * 支持通过、回退、转交、撤销四种操作
 */
import UserSelect from "@/components/UserSelect";
import { updateFlowJourneyStatus } from "@/services/workbench";
import { useApp } from "@/utils/appContext";
import {
  ModalForm,
  ProFormTextArea,
} from "@ant-design/pro-components";
import React from "react";

interface ApprovalModalProps {
  visible: boolean;
  operation: "approve" | "refuse" | "transfer" | "cancel";
  flowId: number;
  journeyId: number;
  assignmentId: number;
  onSuccess: () => void;
  onCancel: () => void;
}

/**
 * 操作标题映射
 */
const OPERATION_TITLES: Record<string, string> = {
  approve: "审批通过",
  refuse: "审批回退",
  transfer: "转交处理",
  cancel: "撤销审批",
};

/**
 * 审批操作Modal
 */
const ApprovalModal: React.FC<ApprovalModalProps> = ({
  visible,
  operation,
  flowId,
  journeyId,
  assignmentId,
  onSuccess,
  onCancel,
}) => {
  const { message } = useApp();

  /**
   * 提交审批操作
   */
  const handleSubmit = async (values: any) => {
    try {
      const params: any = {
        flowId,
        journeyId,
        assignmentId,
        operation,
        comment: values.comment,
      };

      // CarbonCopyUserIds 字段处理
      // - transfer操作：传递转交目标用户（单选，存储为数组）
      // - 其他操作：传递抄送人列表（多选）
      if (values.carbonCopyUserIds) {
        // 如果是单选值（string），转换为数组
        if (typeof values.carbonCopyUserIds === 'string') {
          params.carbonCopyUserIds = [values.carbonCopyUserIds];
        } else if (Array.isArray(values.carbonCopyUserIds) && values.carbonCopyUserIds.length > 0) {
          params.carbonCopyUserIds = values.carbonCopyUserIds;
        }
      }

      const res = await updateFlowJourneyStatus(params);

      if (res.code === 0) {
        message.success(`${OPERATION_TITLES[operation]}成功`);
        onSuccess();
        return true;
      } else {
        message.error(res.msg || `${OPERATION_TITLES[operation]}失败`);
        return false;
      }
    } catch (error) {
      message.error(`${OPERATION_TITLES[operation]}失败`);
      console.error("Approval operation failed:", error);
      return false;
    }
  };

  return (
    <ModalForm
      title={OPERATION_TITLES[operation]}
      open={visible}
      onFinish={handleSubmit}
      onOpenChange={(open) => {
        if (!open) {
          onCancel();
        }
      }}
      modalProps={{
        destroyOnHidden: true,
        okText: "提交",
        cancelText: "取消",
      }}
      width={560}
    >
      {/* 处理意见（所有操作可用） */}
      <ProFormTextArea
        name="comment"
        label="处理意见"
        placeholder={`请输入${OPERATION_TITLES[operation]}的意见`}
        rules={[
          {
            max: 500,
            message: "处理意见不能超过500字",
          },
        ]}
        fieldProps={{
          rows: 4,
          showCount: true,
          maxLength: 500,
        }}
      />

      {/* CarbonCopyUserIds 字段：根据操作类型显示不同的标签和模式 */}
      {operation === "transfer" ? (
        /* 转交操作：单选目标用户（必填） */
        <UserSelect
          name="carbonCopyUserIds"
          label="转交给"
          placeholder="搜索并选择转交的处理人"
          rules={[
            {
              required: true,
              message: "请选择转交的处理人",
            },
          ]}
          showEmail={false}
        />
      ) : (
        /* 其他操作：多选抄送人（可选） */
        <UserSelect
          name="carbonCopyUserIds"
          label="抄送给"
          placeholder="搜索并选择抄送人（可选）"
          mode="multiple"
          showEmail={false}
        />
      )}
    </ModalForm>
  );
};

export default ApprovalModal;
