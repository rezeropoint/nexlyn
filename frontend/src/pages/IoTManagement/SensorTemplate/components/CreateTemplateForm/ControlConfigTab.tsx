import { CopyOutlined, ThunderboltOutlined } from "@ant-design/icons";
import { ProFormDependency, ProFormText } from "@ant-design/pro-components";
import { Alert, Card, Divider, Space, Typography } from "antd";
import React from "react";
import styles from "./ControlConfigTab.less";

const { Paragraph, Text } = Typography;

/**
 * 设备控制配置标签页组件
 */
const ControlConfigTab: React.FC = () => {
  return (
    <>
      <Alert
        message="设备控制配置"
        description="配置设备MQTT控制主题，用于远程控制设备（如AI Box算法任务管理）。不配置则设备不支持远程控制。"
        type="info"
        showIcon
        className={styles.topAlert}
      />

      <ProFormText
        name="commandSuffix"
        label="命令主题后缀"
        placeholder="例如：control、command、cmd"
        tooltip="用于发送控制命令的MQTT主题后缀，将与类别和设备ID组合成完整主题"
        rules={[
          {
            pattern: /^[a-zA-Z0-9_/-]+$/,
            message: "只能包含字母、数字、下划线、中划线和斜杠",
          },
        ]}
      />

      <ProFormText
        name="responseSuffix"
        label="响应主题后缀"
        placeholder="例如：control_response、response、ack"
        tooltip="用于接收控制响应的MQTT主题后缀，将与类别和设备ID组合成完整主题"
        rules={[
          {
            pattern: /^[a-zA-Z0-9_/-]+$/,
            message: "只能包含字母、数字、下划线、中划线和斜杠",
          },
        ]}
      />

      {/* 完整主题预览 */}
      <ProFormDependency
        name={["category", "model", "commandSuffix", "responseSuffix"]}
      >
        {({ category, model, commandSuffix, responseSuffix }) => {
          if (category && model && (commandSuffix || responseSuffix)) {
            const commandTopic = commandSuffix
              ? `nexlyn/iot/${category}/${model}/{device_id}/${commandSuffix}`
              : "";
            const responseTopic = responseSuffix
              ? `nexlyn/iot/${category}/${model}/{device_id}/${responseSuffix}`
              : "";

            return (
              <Card size="small" className={styles.topicPreviewCard}>
                <Space
                  direction="vertical"
                  size={8}
                  className={styles.previewSpace}
                >
                  <Text strong className={styles.previewTitle}>
                    <ThunderboltOutlined /> 完整MQTT主题预览
                  </Text>

                  {commandTopic && (
                    <>
                      <Text type="secondary" className={styles.previewLabel}>
                        命令主题（可复制）：
                      </Text>
                      <Paragraph
                        copyable={{
                          text: commandTopic,
                          icon: [
                            <CopyOutlined key="copy-icon" />,
                            <Text key="copy-success" type="success">
                              已复制
                            </Text>,
                          ],
                        }}
                        className={styles.topicValue}
                      >
                        {commandTopic}
                      </Paragraph>
                    </>
                  )}

                  {responseTopic && (
                    <>
                      <Text type="secondary" className={styles.previewLabel}>
                        响应主题（可复制）：
                      </Text>
                      <Paragraph
                        copyable={{
                          text: responseTopic,
                          icon: [
                            <CopyOutlined key="copy-icon" />,
                            <Text key="copy-success" type="success">
                              已复制
                            </Text>,
                          ],
                        }}
                        className={styles.topicValueLast}
                      >
                        {responseTopic}
                      </Paragraph>
                    </>
                  )}

                  <Divider className={styles.previewDivider} />

                  <Text type="secondary" className={styles.previewHint}>
                    💡 提示：{"{device_id}"} 将被替换为实际的设备ID
                  </Text>
                </Space>
              </Card>
            );
          }
          return null;
        }}
      </ProFormDependency>

      {/* 详细说明卡片 */}
      <Card size="small" className={styles.instructionCard}>
        <Space
          direction="vertical"
          size={12}
          className={styles.instructionSpace}
        >
          <Text strong className={styles.instructionTitle}>
            控制流程说明
          </Text>

          <div className={styles.flowStep}>
            <div className={styles.stepTitle}>
              <Text strong>1. 发送命令</Text>
            </div>
            <div className={styles.stepContent}>
              平台向<Text code>命令主题</Text>发送控制指令（如创建AI算法任务）
            </div>

            <div className={styles.stepTitle}>
              <Text strong>2. 设备执行</Text>
            </div>
            <div className={styles.stepContent}>
              设备订阅命令主题，接收并执行控制指令
            </div>

            <div className={styles.stepTitle}>
              <Text strong>3. 返回结果</Text>
            </div>
            <div className={styles.stepContentLast}>
              设备执行完成后，向<Text code>响应主题</Text>发送执行结果
            </div>
          </div>

          <Divider className={styles.instructionDivider} />

          <div className={styles.exampleSection}>
            <Text strong>AI Box控制示例</Text>
            <div className={styles.exampleList}>
              <div>• 创建算法任务（CreateTask）</div>
              <div>• 更新任务配置（UpdateTask）</div>
              <div>• 删除算法任务（DeleteTask）</div>
              <div>• 查询任务列表（ListTasks）</div>
              <div>• 获取算法能力（GetCapabilities）</div>
            </div>
          </div>

          <Divider className={styles.instructionDivider} />

          <Text type="secondary" className={styles.warningHint}>
            ⚠️
            注意：控制配置为可选项，两个主题后缀必须同时填写或同时为空。清空主题后缀将删除控制配置。
          </Text>
        </Space>
      </Card>
    </>
  );
};

export default ControlConfigTab;
