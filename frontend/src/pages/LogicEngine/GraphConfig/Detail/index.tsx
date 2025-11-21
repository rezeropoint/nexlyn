/**
 * 逻辑图详情页主入口
 * 包含5个Tab：基本信息、可视化编辑器、运行时数据、Webhooks配置、统计分析
 */

import { useNavigate, useParams } from "@@/exports";
import { ArrowLeftOutlined, ReloadOutlined } from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { App, Button, Space, Tabs } from "antd";
import React, { useState } from "react";
import BasicInfoTab from "./components/BasicInfoTab";
import RuntimeDataTab from "./components/RuntimeDataTab";
import StatsTab from "./components/StatsTab";
import VisualEditorTab from "./components/VisualEditorTab";
import WebhooksTab from "./components/WebhooksTab";
import { useGraphDetail } from "./hooks/useGraphDetail";

const GraphConfigDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [activeTab, setActiveTab] = useState("basic");

  // 获取详情数据
  const { data, loading, refresh, updateBasicInfo, updateGraphStructure } =
    useGraphDetail(id || "");

  // Tab配置
  const tabItems = [
    {
      key: "basic",
      label: "基本信息",
      children: (
        <BasicInfoTab
          data={data}
          loading={loading}
          onUpdate={updateBasicInfo}
        />
      ),
    },
    {
      key: "visual",
      label: "可视化编辑器",
      children: (
        <VisualEditorTab
          data={data}
          loading={loading}
          tenantId={data?.tenantId}
          onSave={updateGraphStructure}
        />
      ),
    },
    {
      key: "runtime",
      label: "运行时数据",
      children: <RuntimeDataTab graphId={id || ""} />,
    },
    {
      key: "webhooks",
      label: "Webhooks配置",
      children: <WebhooksTab graphId={id || ""} />,
    },
    {
      key: "stats",
      label: "统计分析",
      children: <StatsTab graphId={id || ""} />,
    },
  ];

  return (
    <PageContainer
      title={data?.name || "逻辑图详情"}
      subTitle={data?.version && `版本：${data.version}`}
      loading={loading}
      breadcrumb={{
        items: [
          { title: "逻辑引擎" },
          {
            title: "逻辑图配置",
            href: "/logic-engine/graph-config",
          },
          { title: "详情" },
        ],
      }}
      extra={
        <Space>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              refresh();
              message.success("刷新成功");
            }}
          >
            刷新
          </Button>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate("/logic-engine/graph-config")}
          >
            返回列表
          </Button>
        </Space>
      }
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        style={{ padding: "0 24px" }}
      />
    </PageContainer>
  );
};

export default GraphConfigDetail;
