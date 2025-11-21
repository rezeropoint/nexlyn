import { useApp } from "@/utils/appContext";
import {
  ApartmentOutlined,
  DashboardOutlined,
  FullscreenExitOutlined,
  FullscreenOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import { PageContainer } from "@ant-design/pro-components";
import { useModel } from "@umijs/max";
import { Button, Flex, FloatButton, TreeSelect } from "antd";
import React, { useCallback, useEffect, useMemo, useState } from "react";
import DeviceStatistics from "./components/DeviceStatistics";
import EventStatistics from "./components/EventStatistics";
import MapView from "./components/MapView";
import { useFullscreen } from "./hooks/useFullscreen";
import { useOrganization } from "./hooks/useOrganization";
import "./index.less";

const VisualizationDashboard: React.FC = () => {
  const { message } = useApp();
  const { initialState } = useModel("@@initialState");
  const [lastUpdateTime, setLastUpdateTime] = useState(new Date());

  // 使用自定义hooks
  const { selectedOrg, setSelectedOrg, orgTreeData, loadingOrg } =
    useOrganization(initialState?.currentUser, message);
  const { isFullscreen, toggleFullscreen } = useFullscreen(message);

  // 缓存组织数据，避免重复计算
  const memoizedOrgData = useMemo(() => orgTreeData, [orgTreeData]);

  // 优化组织变更回调
  const handleOrgChange = useCallback(
    (value: string) => {
      setSelectedOrg(value);
    },
    [setSelectedOrg]
  );

  // 模拟数据刷新
  const handleRefresh = () => {
    setLastUpdateTime(new Date());
    message.success("数据已刷新");
  };

  // 自动刷新
  useEffect(() => {
    const interval = setInterval(() => {
      setLastUpdateTime(new Date());
    }, 30000); // 30秒自动刷新一次

    return () => clearInterval(interval);
  }, []);

  // 格式化更新时间
  const formatUpdateTime = (time: Date) => {
    return time.toLocaleTimeString("zh-CN", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  };

  return (
    <PageContainer
      className={`visualization-dashboard ${isFullscreen ? "fullscreen" : ""}`}
      pageHeaderRender={false}
      subTitle="实时监控设备状态、事件统计和地理分布，支持全屏展示"
    >
      {/* 工具栏 */}
      {!isFullscreen && (
        <div className="dashboard-toolbar">
          <div className="toolbar-left">
            <span className="dashboard-title">
              <DashboardOutlined className="title-icon" />
              可视化大屏
            </span>
          </div>
          <div className="toolbar-center">
            <div className="organization-selector">
              <span className="selector-label">
                <ApartmentOutlined className="label-icon" />
                组织架构:
              </span>
              <TreeSelect
                className="org-tree-select"
                value={selectedOrg}
                treeData={memoizedOrgData}
                placeholder={loadingOrg ? "加载中..." : "请选择组织架构"}
                treeDefaultExpandAll
                onChange={handleOrgChange}
                size="small"
                showSearch={false}
                virtual={false}
                popupMatchSelectWidth={false}
                loading={loadingOrg}
                disabled={loadingOrg || memoizedOrgData.length === 0}
              />
            </div>
          </div>
          <div className="toolbar-right">
            <div className="update-time">
              最后更新: {formatUpdateTime(lastUpdateTime)}
            </div>
            <Button
              type="text"
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              className="toolbar-button"
            >
              刷新
            </Button>
            <Button
              type="text"
              icon={<FullscreenOutlined />}
              onClick={toggleFullscreen}
              className="toolbar-button"
            >
              全屏
            </Button>
          </div>
        </div>
      )}

      <div className="dashboard-container">
        <Flex gap={16} className="dashboard-row">
          {/* 左侧设备统计 - 固定宽度 */}
          <div
            style={{ flex: "0 0 320px" }}
            className="statistics-col left-col"
          >
            <div className="statistics-panel">
              <DeviceStatistics orgId={selectedOrg} />
            </div>
          </div>

          {/* 中间地图区域 - 自适应宽度 */}
          <div style={{ flex: 1 }} className="center-col">
            <div className="center-panel">
              <MapView />
            </div>
          </div>

          {/* 右侧事件统计 - 固定宽度 */}
          <div
            style={{ flex: "0 0 320px" }}
            className="statistics-col right-col"
          >
            <div className="statistics-panel">
              <EventStatistics orgId={selectedOrg} />
            </div>
          </div>
        </Flex>
      </div>

      {/* 全屏模式下的标题栏 */}
      {isFullscreen && (
        <div className="fullscreen-header">
          <div className="header-left">
            <DashboardOutlined className="header-icon" />
            <span className="header-title">Nexlyn 可视化大屏</span>
          </div>
          <div className="header-right">
            <span className="header-time">
              {new Date().toLocaleString("zh-CN", {
                year: "numeric",
                month: "2-digit",
                day: "2-digit",
                hour: "2-digit",
                minute: "2-digit",
                second: "2-digit",
                hour12: false,
              })}
            </span>
            <Button
              type="text"
              icon={<FullscreenExitOutlined />}
              onClick={toggleFullscreen}
              className="exit-fullscreen-btn"
            >
              退出全屏
            </Button>
          </div>
        </div>
      )}

      {/* 动态背景效果 */}
      <div className="background-effects">
        <div className="grid-bg"></div>
        <div className="particles"></div>
      </div>

      {/* 返回顶部按钮 */}
      <FloatButton.BackTop visibilityHeight={300} />
    </PageContainer>
  );
};

export default VisualizationDashboard;
