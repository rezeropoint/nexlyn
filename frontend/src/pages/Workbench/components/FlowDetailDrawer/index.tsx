/**
 * 统一的流程详情 Drawer
 * 用于工作台三个列表页
 */
import {
  CheckCircleOutlined,
  CheckOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  CloseOutlined,
  FileOutlined,
  StopOutlined,
  UserOutlined,
  UserSwitchOutlined,
} from "@ant-design/icons";
import {
  DrawerForm,
  ProForm,
  ProFormCheckbox,
  ProFormDatePicker,
  ProFormDigit,
  ProFormRadio,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  type ProFormInstance,
} from "@ant-design/pro-components";
import {
  Avatar,
  Badge,
  Button,
  Descriptions,
  Empty,
  Flex,
  Popconfirm,
  Space,
  Spin,
  Tag,
  Timeline,
  Typography,
} from "antd";
import classNames from "classnames";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  abortJourney,
  getJourneyFullDetail,
  updateFlowJourneyStatus,
} from "@/services/workbench";
import { getUsersByIds } from "@/services/user";
import type {
  JourneyDetail,
  Moment,
  PendingNode,
  ProcessingUser,
  TypedValue,
  VertexField,
} from "@/services/workbench/types";
import { useApp } from "@/utils/appContext";
import UserSelect from "@/components/UserSelect";
import ApprovalModal from "../ApprovalModal";
import styles from "./index.less";

const { Text } = Typography;

export interface FlowDetailDrawerProps {
  visible: boolean;
  flowId?: number;
  journeyId?: number;
  assignmentId?: number;
  onClose: () => void;
  onActionSuccess?: () => void;
}

const statusConfig: Record<
  string,
  { text: string; badge: "processing" | "success" | "warning" | "error"; tagClass: string }
> = {
  processing: { text: "进行中", badge: "processing", tagClass: styles.statusTagProcessing },
  completed: { text: "已完成", badge: "success", tagClass: styles.statusTagCompleted },
  aborted: { text: "已终止", badge: "error", tagClass: styles.statusTagAborted },
  stashed: { text: "已暂存", badge: "warning", tagClass: styles.statusTagDraft },
};

const actionIcons: Record<string, React.ReactNode> = {
  approved: <CheckCircleOutlined />,
  refused: <CloseCircleOutlined />,
  transferred: <UserOutlined />,
  cancelled: <ClockCircleOutlined />,
};

dayjs.extend(relativeTime);

/** 精准格式化处理时长：天/小时/分钟/秒 */
const formatDuration = (totalSeconds: number): string => {
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = Math.floor(totalSeconds % 60);

  const parts: string[] = [];
  if (days > 0) parts.push(`${days} 天`);
  if (hours > 0) parts.push(`${hours} 小时`);
  if (minutes > 0) parts.push(`${minutes} 分钟`);
  if (seconds > 0 || parts.length === 0) parts.push(`${seconds} 秒`);

  return parts.join(" ");
};

const FlowDetailDrawer: React.FC<FlowDetailDrawerProps> = ({
  visible,
  flowId,
  journeyId,
  assignmentId,
  onClose,
  onActionSuccess,
}) => {
  const { message } = useApp();
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [detail, setDetail] = useState<JourneyDetail | null>(null);
  const [moments, setMoments] = useState<Moment[]>([]);
  const [processingUsers, setProcessingUsers] = useState<ProcessingUser[]>([]);
  const [pendingNodes, setPendingNodes] = useState<PendingNode[]>([]);
  const [transferModalVisible, setTransferModalVisible] = useState(false);
  const [showRefuseVertex, setShowRefuseVertex] = useState(false);

  /**
   * 判断当前用户是否有权限操作
   * 简化判断：只要有assignmentId就允许操作（说明这是当前用户的待办任务）
   */
  const canOperate = useMemo(() => {
    return assignmentId !== undefined;
  }, [assignmentId]);

  // 使用 formRef 替代 Form.useForm
  const formRef = useRef<ProFormInstance>();
  const shouldShowForm = canOperate && detail?.status === "processing";

  const loadDetail = async () => {
    if (!flowId || !journeyId) return;
    setLoading(true);
    try {
      const res = await getJourneyFullDetail({ flowId, journeyId });

      if (res.code === 0) {
        const { basicInfo, history, pendingNodes } = res.data;

        // 设置基础信息和审批历史
        setDetail(basicInfo);
        setMoments(history || []);

        // 保存 pendingNodes（用于渲染动态字段）
        setPendingNodes(pendingNodes || []);

        // 从 pendingNodes 提取处理人 ID 并获取用户详情
        const allAssigneeIds: string[] = [];
        pendingNodes?.forEach((node) => {
          node.assigneeIds?.forEach((id) => {
            if (!allAssigneeIds.includes(id)) {
              allAssigneeIds.push(id);
            }
          });
        });

        // 调用用户接口获取详细信息
        if (allAssigneeIds.length > 0) {
          const userRes = await getUsersByIds(allAssigneeIds);
          if (userRes.code !== 0 || !userRes.data?.list) {
            throw new Error(userRes.msg || "获取处理人信息失败");
          }
          const userMap = new Map(userRes.data.list.map((u) => [u.id, u]));
          const users: ProcessingUser[] = allAssigneeIds.map((id) => {
            const user = userMap.get(id);
            if (!user) {
              throw new Error(`未找到用户信息: ${id}`);
            }
            return {
              id,
              name: user.name,
              nickname: "",
              phone: "",
              identifier: "",
              headimgurl: user.avatar || "",
              tags: [],
            };
          });
          setProcessingUsers(users);
        } else {
          setProcessingUsers([]);
        }
      } else {
        message.error(res.msg || "加载流程详情失败");
      }
    } catch (error) {
      message.error("加载流程详情失败");
      console.error("Failed to load flow detail:", error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible && flowId && journeyId) {
      loadDetail();
      if (shouldShowForm) {
        formRef.current?.resetFields();
      }
      setShowRefuseVertex(false);
    } else if (!visible) {
      setDetail(null);
      setMoments([]);
      setProcessingUsers([]);
      setPendingNodes([]);
      if (shouldShowForm) {
        formRef.current?.resetFields();
      }
      setShowRefuseVertex(false);
    }
  }, [visible, flowId, journeyId, shouldShowForm]);

  /**
   * 获取所有待处理节点的字段列表
   * 注意：显示所有字段，但根据 editable 状态控制是否可编辑
   */
  const pendingFields = useMemo(() => {
    const fields: VertexField[] = [];
    pendingNodes.forEach(node => {
      node.fields?.forEach(field => {
        fields.push(field);
      });
    });
    return fields;
  }, [pendingNodes]);

  /** 支持填写的字段类型 */
  const EDITABLE_FIELD_TYPES = ['Field::RadioButton', 'Field::TextField'];

  /**
   * 根据字段类型转换表单值为 TypedValue
   * 仅支持 Field::RadioButton 和 Field::TextField
   */
  const convertToTypedValue = (field: VertexField, value: any): TypedValue | null => {
    const fieldType = field.type;

    // 仅支持指定的字段类型
    if (!EDITABLE_FIELD_TYPES.includes(fieldType)) {
      return null;
    }

    if (value === undefined || value === null || value === '') {
      return { type: 'string', value: '' };
    }

    switch (fieldType) {
      case 'Field::RadioButton': {
        // 单选：根据选项 ID 查找选项 value
        const option = field.options?.find((opt) => opt.id === Number(value));
        return { type: 'string', value: option?.value ?? '' };
      }
      case 'Field::TextField':
        return { type: 'string', value: String(value) };
      default:
        return null;
    }
  };

  /**
   * 处理审批操作（通过、回退、撤销）
   */
  const handleApprovalAction = async (operation: "approve" | "refuse" | "cancel") => {
    if (!flowId || !journeyId || !assignmentId) return;

    // 回退操作需要先显示节点选择
    if (operation === "refuse" && !showRefuseVertex) {
      setShowRefuseVertex(true);
      return;
    }

    try {
      const values = await formRef.current?.validateFields();
      if (!values) return;
      setSubmitting(true);

      const params: any = {
        flowId,
        journeyId,
        assignmentId,
        operation,
        comment: values.comment,
      };

      // 抄送人列表
      if (values.carbonCopyUserIds && values.carbonCopyUserIds.length > 0) {
        params.carbonCopyUserIds = values.carbonCopyUserIds;
      }

      // 收集动态字段数据（只收集支持的可编辑字段）
      if (pendingFields.length > 0) {
        const data: Record<string, TypedValue> = {};
        pendingFields.forEach(field => {
          // 只有可编辑且支持的字段类型才提交
          if (field.editable && EDITABLE_FIELD_TYPES.includes(field.type)) {
            const fieldValue = values[`field_${field.identityKey}`];
            if (fieldValue !== undefined && fieldValue !== null && fieldValue !== '') {
              const typedValue = convertToTypedValue(field, fieldValue);
              if (typedValue) {
                data[field.identityKey] = typedValue;
              }
            }
          }
        });
        if (Object.keys(data).length > 0) {
          params.data = data;
        }
      }

      const res = await updateFlowJourneyStatus(params);

      if (res.code === 0) {
        const operationNames: Record<string, string> = {
          approve: "审批通过",
          refuse: "审批回退",
          cancel: "撤销审批",
        };
        message.success(`${operationNames[operation]}成功`);
        onActionSuccess?.();
        onClose();
      } else {
        message.error(res.msg || "操作失败");
      }
    } catch (error: any) {
      // 表单验证失败时不显示错误消息
      if (error.errorFields) {
        return;
      }
      message.error("操作失败");
      console.error("Approval operation failed:", error);
    } finally {
      setSubmitting(false);
    }
  };

  /**
   * 打开转交Modal
   */
  const handleOpenTransfer = () => {
    setTransferModalVisible(true);
  };

  /**
   * 终止流程
   */
  const handleAbort = async () => {
    if (!flowId || !journeyId) return;
    try {
      const res = await abortJourney({ flowId, journeyId });
      if (res.code === 0) {
        message.success("终止流程成功");
        onActionSuccess?.();
        onClose();
      } else {
        message.error(res.msg || "终止流程失败");
      }
    } catch (error) {
      message.error("终止流程失败");
      console.error("Failed to abort journey:", error);
    }
  };

  /**
   * 转交操作成功回调
   */
  const handleTransferSuccess = () => {
    setTransferModalVisible(false);
    onActionSuccess?.();
    onClose();
  };

  /**
   * 渲染动态字段表单项
   * 仅支持 Field::RadioButton 和 Field::TextField 可编辑，其他类型禁用
   */
  const renderDynamicField = (field: VertexField) => {
    const fieldName = `field_${field.identityKey}`;
    // 仅支持的字段类型才可编辑
    const isSupported = EDITABLE_FIELD_TYPES.includes(field.type);
    const isDisabled = !field.editable || !isSupported;
    const rules = field.editable && isSupported && field.required
      ? [{ required: true, message: `请填写${field.title}` }]
      : undefined;

    switch (field.type) {
      case 'Field::TextField':
        return (
          <ProFormText
            key={field.id}
            name={fieldName}
            label={field.title}
            rules={rules}
            placeholder={`请输入${field.title}`}
            disabled={isDisabled}
          />
        );

      case 'Field::RadioButton':
        return (
          <ProFormRadio.Group
            key={field.id}
            name={fieldName}
            label={field.title}
            rules={rules}
            options={field.options?.map(option => ({
              label: option.value,
              value: option.id,
            }))}
            disabled={isDisabled}
          />
        );

      // 以下类型不支持编辑，仅显示提示
      default:
        return (
          <ProForm.Item key={field.id} label={field.title}>
            <Text type="secondary">暂不支持此字段类型</Text>
          </ProForm.Item>
        );
    }
  };

  const detailStatus = detail ? statusConfig[detail.status] : undefined;

  const infoCards = useMemo(() => {
    if (!detail) return [];
    return [
      { label: "流程ID", value: detail.flowId },
      { label: "当前节点ID", value: detail.currentVertexId ?? "-" },
      { label: "创建时间", value: dayjs(detail.createdAt).format("YYYY-MM-DD HH:mm") },
      { label: "更新时间", value: dayjs(detail.updatedAt).format("YYYY-MM-DD HH:mm") },
    ];
  }, [detail]);

  return (
    <>
      <DrawerForm
        title={null}
        width={840}
        open={visible}
        onOpenChange={(open) => !open && onClose()}
        formRef={formRef}
        submitter={false}
        drawerProps={{
          destroyOnHidden: true,
          className: styles.flowDrawer,
        }}
      >
        <Spin spinning={loading}>
        {detail ? (
          <div className={styles.drawerContent}>
            <div className={styles.hero}>
              <div>
                <div className={styles.snRow}>
                  <Text className={styles.sn}>{detail.sn}</Text>
                  {detailStatus && (
                    <Tag className={detailStatus.tagClass} bordered={false}>
                      {detailStatus.text}
                    </Tag>
                  )}
                </div>
                <div className={styles.metaRow}>
                  <span>流程ID {detail.flowId}</span>
                  <span>节点ID {detail.currentVertexId ?? "-"}</span>
                  {detail.initiator && <span>发起人 {detail.initiator.name}</span>}
                </div>
              </div>
              <div className={styles.heroTimes}>
                <div>
                  <span>创建</span>
                  <strong>{dayjs(detail.createdAt).format("MM-DD HH:mm")}</strong>
                </div>
                <div>
                  <span>最近更新</span>
                  <strong>{dayjs(detail.updatedAt).fromNow()}</strong>
                </div>
              </div>
            </div>

            <div className={styles.infoGrid}>
              {infoCards.map((card) => (
                <div key={card.label} className={styles.infoCard}>
                  <span>{card.label}</span>
                  <strong>{card.value}</strong>
                </div>
              ))}
            </div>

            {processingUsers.length > 0 && (
              <section className={styles.block}>
                <div className={styles.blockHeader}>
                  <Text strong>当前处理人</Text>
                  <Badge status="processing" text={`${processingUsers.length} 人`} />
                </div>
                <div className={styles.userGrid}>
                  {processingUsers.map((user) => (
                    <div key={user.id} className={styles.userCard}>
                      <Avatar
                        src={user.headimgurl}
                        icon={<UserOutlined />}
                        className={styles.userAvatar}
                      />
                      <div>
                        <div className={styles.userName}>{user.name}</div>
                        {user.phone && (
                          <div className={styles.userMeta}>
                            <span>{user.phone}</span>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </section>
            )}

            {detail.businessData && Object.keys(detail.businessData).length > 0 && (
              <section className={styles.block}>
                <div className={styles.blockHeader}>
                  <Text strong>业务数据</Text>
                </div>
                <Descriptions
                  column={2}
                  size="small"
                  className={styles.businessDescriptions}
                >
                  {Object.entries(detail.businessData).map(([key, value]) => (
                    <Descriptions.Item label={key} key={key}>
                      {typeof value === "object" ? JSON.stringify(value) : String(value)}
                    </Descriptions.Item>
                  ))}
                </Descriptions>
              </section>
            )}

            <section className={styles.block}>
              <div className={styles.blockHeader}>
                <Text strong>审批历史</Text>
                <Badge count={moments.length} showZero />
              </div>
              {moments.length > 0 ? (
                <Timeline
                  className={styles.timeline}
                  items={moments.map((moment) => ({
                    key: moment.id,
                    dot: actionIcons[moment.status] || <FileOutlined />,
                    children: (
                      <>
                        <Flex justify="space-between" align="center">
                          <div className={styles.timelineTitle}>
                            <span className={styles.timelineOperator}>
                              {moment.operatorName || `用户${moment.operatorId}`}
                            </span>
                            <Tag className={styles.timelineAction} bordered={false}>
                              {moment.status}
                            </Tag>
                            {moment.vertexName && (
                              <Text type="secondary">{moment.vertexName}</Text>
                            )}
                          </div>
                          <span className={styles.timelineTime}>
                            {dayjs(moment.updatedAt).format("YYYY-MM-DD HH:mm")}
                          </span>
                        </Flex>
                        {moment.comment && (
                          <div className={styles.timelineComment}>{moment.comment}</div>
                        )}
                        {moment.createdAt && moment.updatedAt && (
                          <div className={styles.timelineDuration}>
                            处理时长：
                            {formatDuration(
                              dayjs(moment.updatedAt).diff(dayjs(moment.createdAt), "second")
                            )}
                          </div>
                        )}
                      </>
                    ),
                  }))}
                />
              ) : (
                <Empty description="暂无审批历史" />
              )}
            </section>

            {detail.attachments && detail.attachments.length > 0 && (
              <section className={styles.block}>
                <div className={styles.blockHeader}>
                  <Text strong>附件</Text>
                  <Badge count={detail.attachments.length} />
                </div>
                  <div className={styles.attachmentList}>
                    {detail.attachments.map((attachment) => (
                      <a
                        key={attachment.id}
                        href={attachment.url}
                        target="_blank"
                        rel="noreferrer"
                        className={styles.attachmentItem}
                      >
                        <FileOutlined className={styles.attachmentIcon} />
                        <div>
                          <div className={styles.attachmentName}>{attachment.name}</div>
                          <div className={styles.attachmentMeta}>
                            {(attachment.size / 1024).toFixed(1)} KB ·{" "}
                            {dayjs(attachment.createdAt).format("MM-DD HH:mm")}
                          </div>
                        </div>
                      </a>
                    ))}
                  </div>
              </section>
            )}

            {/* 审批操作表单 */}
            <div className={classNames(styles.approvalFormWrapper, { [styles.hidden]: !shouldShowForm })}>
              <section className={styles.block}>
                <div className={styles.blockHeader}>
                  <Text strong>审批操作</Text>
                </div>
                <div className={styles.approvalForm}>
                  {/* 动态字段 - 来自 pendingNodes */}
                  {pendingFields.length > 0 &&
                    pendingFields.map(field => renderDynamicField(field))}

                  {/* 处理意见 */}
                  <ProFormTextArea
                    name="comment"
                    label="处理意见"
                    placeholder="请输入处理意见（可选）"
                    fieldProps={{
                      rows: 3,
                      showCount: true,
                      maxLength: 500,
                    }}
                    rules={[{ max: 500, message: "处理意见不能超过500字" }]}
                  />

                  {/* 抄送人 */}
                  <ProForm.Item name="carbonCopyUserIds" label="抄送给">
                    <UserSelect
                      placeholder="搜索并选择抄送人（可选）"
                      mode="multiple"
                      showEmail={false}
                    />
                  </ProForm.Item>

                  {/* 操作按钮 */}
                  <Space style={{ marginTop: 16 }}>
                    <Button
                      type="primary"
                      icon={<CheckOutlined />}
                      loading={submitting}
                      onClick={() => handleApprovalAction("approve")}
                    >
                      通过
                    </Button>
                    <Button
                      icon={<CloseOutlined />}
                      loading={submitting}
                      onClick={() => handleApprovalAction("refuse")}
                    >
                      回退
                    </Button>
                    <Button
                      icon={<UserSwitchOutlined />}
                      loading={submitting}
                      onClick={handleOpenTransfer}
                    >
                      转交
                    </Button>
                    <Popconfirm
                      title="确定要终止此流程吗？"
                      description="终止后流程将无法继续进行"
                      onConfirm={handleAbort}
                      okText="确定"
                      cancelText="取消"
                    >
                      <Button danger icon={<StopOutlined />}>
                        终止流程
                      </Button>
                    </Popconfirm>
                  </Space>
                </div>
              </section>
            </div>
          </div>
        ) : (
          !loading && <Empty description="暂无数据" />
        )}
      </Spin>
    </DrawerForm>

    {/* 转交操作Modal */}
    {transferModalVisible && flowId && journeyId && assignmentId && (
      <ApprovalModal
        visible={transferModalVisible}
        operation="transfer"
        flowId={flowId}
        journeyId={journeyId}
        assignmentId={assignmentId}
        onSuccess={handleTransferSuccess}
        onCancel={() => setTransferModalVisible(false)}
      />
    )}
  </>
  );
};

export default FlowDetailDrawer;

