import { CopyOutlined } from "@ant-design/icons";
import {
  ProFormDependency,
  ProFormDigit,
  ProFormList,
  ProFormSelect,
  ProFormText,
} from "@ant-design/pro-components";
import { Alert, Card, Divider, Space, Typography } from "antd";
import React from "react";
import styles from "./OnlineDetectionTab.less";

const { Paragraph, Text } = Typography;

/**
 * 在线检测配置标签页组件（重构版）
 */
const OnlineDetectionTab: React.FC = () => {
  return (
    <>
      <Alert
        message="在线检测配置"
        description="根据MQTT消息判断设备在线状态。不配置则默认设备为离线状态。"
        type="info"
        showIcon
        className={styles.infoAlert}
      />

      <ProFormText
        name="onlineTopicSuffix"
        label="MQTT主题后缀"
        placeholder="例如：heartbeat、status、online"
        tooltip="用于在线检测的MQTT主题后缀，将与类别和设备ID组合成完整主题"
        rules={[
          {
            pattern: /^[a-zA-Z0-9_/-]+$/,
            message: "只能包含字母、数字、下划线、中划线和斜杠",
          },
        ]}
      />

      {/* 完整主题预览 */}
      <ProFormDependency name={["category", "model", "onlineTopicSuffix"]}>
        {({ category, model, onlineTopicSuffix }) => {
          if (category && model && onlineTopicSuffix) {
            const fullTopic = `nexlyn/iot/${category}/${model}/{device_id}/${onlineTopicSuffix}`;
            return (
              <Card size="small" className={styles.topicPreviewCard}>
                <Space
                  direction="vertical"
                  size={4}
                  className={styles.topicPreviewContainer}
                >
                  <Text type="secondary" className={styles.topicLabel}>
                    完整MQTT主题（可复制）：
                  </Text>
                  <Paragraph
                    copyable={{
                      text: fullTopic,
                      icon: [
                        <CopyOutlined key="copy-icon" />,
                        <Text key="copy-success" type="success">
                          已复制
                        </Text>,
                      ],
                    }}
                    className={styles.topicParagraph}
                  >
                    {fullTopic}
                  </Paragraph>
                  <Text type="secondary" className={styles.topicHint}>
                    提示：{"{device_id}"} 将被替换为实际的设备ID
                  </Text>
                </Space>
              </Card>
            );
          }
          return null;
        }}
      </ProFormDependency>

      <ProFormSelect
        name="onlineStrategy"
        label="检测策略"
        tooltip="设备在线状态的判断方式"
        options={[
          {
            label: "任意数据",
            value: "any_data",
            description: "接收到任意数据即判定为在线",
          },
          {
            label: "字段检查",
            value: "field_check",
            description: "检查消息中的特定字段是否符合条件",
          },
        ]}
        fieldProps={{
          optionLabelProp: "label",
        }}
      />

      <ProFormDigit
        name="timeoutSeconds"
        label="超时时间（秒）"
        placeholder="300"
        tooltip="超过此时间未收到符合条件的数据则判定为离线"
        fieldProps={{ min: 1, precision: 0 }}
      />

      <ProFormDependency name={["onlineStrategy"]}>
        {({ onlineStrategy }) => {
          if (onlineStrategy !== "field_check") {
            return null;
          }

          return (
            <>
              <Divider orientation="left" className={styles.fieldCheckDivider}>
                字段检查规则
              </Divider>

              <Card size="small" className={styles.operatorGuideCard}>
                <Space
                  direction="vertical"
                  size={8}
                  className={styles.operatorGuideContainer}
                >
                  <Text strong className={styles.operatorGuideTitle}>
                    操作符说明
                  </Text>
                  <div className={styles.operatorGuideContent}>
                    <div>
                      <Text code>exists</Text> - 字段存在即可（不检查值）
                    </div>
                    <div className={styles.operatorExample}>
                      示例：检查 heartbeat 字段是否存在
                    </div>

                    <div className={styles.operatorExampleSpacing}>
                      <Text code>equals</Text> - 字段值等于指定值
                    </div>
                    <div className={styles.operatorExample}>
                      示例：status = "online" 或 error_code = 0
                    </div>

                    <div className={styles.operatorExampleSpacing}>
                      <Text code>in</Text> - 字段值在列表中
                    </div>
                    <div className={styles.operatorExample}>
                      示例：status ∈ ["online", "idle"] 或 code ∈ [0, 100]
                    </div>
                  </div>

                  <Divider className={styles.operatorDivider} />

                  <Text type="secondary" className={styles.operatorHint}>
                    💡 多个规则之间为"与"关系，所有规则都满足才判定为在线
                  </Text>
                </Space>
              </Card>

              <ProFormList
                name="fieldChecks"
                label="检查规则列表"
                tooltip="定义需要检查的字段及其期望值"
                creatorButtonProps={{
                  position: "bottom",
                  creatorButtonText: "+ 添加检查规则",
                }}
                copyIconProps={false}
                deleteIconProps={{ tooltipText: "删除此规则" }}
                itemRender={({ listDom, action }, { index }) => (
                  <Card
                    size="small"
                    className={styles.checkRuleCard}
                    title={
                      <Space>
                        <Text strong className={styles.checkRuleTitle}>
                          规则 #{index + 1}
                        </Text>
                      </Space>
                    }
                    extra={action}
                  >
                    {listDom}
                  </Card>
                )}
              >
                <Space
                  direction="vertical"
                  className={styles.checkRuleContainer}
                  size="middle"
                >
                  <ProFormText
                    name="field"
                    label="字段路径"
                    placeholder="例如：status、heartbeat、data.alive"
                    tooltip="MQTT消息JSON中的字段路径，使用点分隔符"
                    rules={[{ required: true, message: "请输入字段路径" }]}
                  />

                  <ProFormSelect
                    name="operator"
                    label="检查操作符"
                    placeholder="请选择检查操作符"
                    tooltip="定义如何检查字段"
                    options={[
                      {
                        label: "exists - 字段存在即可",
                        value: "exists",
                      },
                      {
                        label: "equals - 字段值等于指定值",
                        value: "equals",
                      },
                      {
                        label: "in - 字段值在列表中",
                        value: "in",
                      },
                    ]}
                  />

                  <ProFormDependency name={["operator"]}>
                    {({ operator }) => {
                      if (operator === "equals") {
                        return (
                          <ProFormText
                            name="value"
                            label="期望值"
                            placeholder="例如：online、0、true"
                            tooltip="字段值必须等于此值（支持字符串、数字、布尔值）"
                            rules={[
                              { required: true, message: "请输入期望值" },
                            ]}
                          />
                        );
                      }

                      if (operator === "in") {
                        return (
                          <ProFormSelect
                            name="values"
                            label="期望值列表"
                            mode="tags"
                            placeholder="输入值后按回车。例如：online, offline, idle"
                            tooltip="字段值必须在此列表中（支持字符串、数字、布尔值）"
                            fieldProps={{
                              tokenSeparators: [",", " ", ";"],
                            }}
                            rules={[
                              {
                                required: true,
                                message: "请至少输入一个期望值",
                              },
                            ]}
                          />
                        );
                      }

                      return null;
                    }}
                  </ProFormDependency>
                </Space>
              </ProFormList>
            </>
          );
        }}
      </ProFormDependency>
    </>
  );
};

export default OnlineDetectionTab;
