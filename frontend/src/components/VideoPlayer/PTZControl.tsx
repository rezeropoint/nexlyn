import { ptzControl } from "@/services/video";
import {
  DownOutlined,
  FullscreenExitOutlined,
  FullscreenOutlined,
  LeftOutlined,
  MinusOutlined,
  PauseOutlined,
  PlusOutlined,
  RightOutlined,
  UpOutlined,
} from "@ant-design/icons";
import { Button, message, Radio, Space, Tabs } from "antd";
import React, { useCallback, useRef, useState } from "react";
import "./PTZControl.less";

const { TabPane } = Tabs;
const { Group: RadioGroup } = Radio;

interface PTZControlProps {
  deviceId: string;
  channelId: string;
  visible: boolean;
  onClose?: () => void;
}

// 云台控制指令类型
type PTZDirection =
  | "up"
  | "down"
  | "left"
  | "right"
  | "leftup"
  | "rightup"
  | "leftdown"
  | "rightdown"
  | "stop"
  | "zoomin"
  | "zoomout"
  | "focusin"
  | "focusout"
  | "apertureopen"
  | "apertureclose"
  | "wiperon"
  | "wiperoff";

// 云台控制指令生成函数 - 使用固定指令，待后续根据协议修改
const generatePTZCommand = (direction: PTZDirection): string => {
  // 所有指令暂时使用固定值，待根据实际GB28181协议规范修改
  const commandMap: Record<PTZDirection, string> = {
    // PTZ方向控制指令 - 需要根据实际协议修改
    up: "A50F0108007D003A", // 上
    down: "A50F0104007D0036", // 下
    left: "A50F01027D000034", // 左
    right: "A50F01017D000033", // 右
    leftup: "A50F010A7D7D00B9", // 左上
    rightup: "A50F01097D7D00B8", // 右上
    leftdown: "A50F01067D7D00B5", // 左下
    rightdown: "A50F01057D7D00B4", // 右下
    stop: "A50F0100000000B5", // 停止

    // 镜头控制 - 需要根据实际协议修改
    zoomin: "A50F011000009055", // 放大
    zoomout: "A50F012000009065", // 缩小

    // 焦距控制 - 需要根据实际协议修改
    focusin: "A50F01417D000073", // 聚焦远
    focusout: "A50F01427D000074", // 聚焦近

    // 光圈控制 - 需要根据实际协议修改
    apertureopen: "A50F0144007D0076", // 光圈开
    apertureclose: "A50F0148007D007A", // 光圈关

    // 雨刷控制 - 需要根据实际协议修改
    wiperon: "A50F018C01000042", // 雨刷开
    wiperoff: "A50F018D01000043", // 雨刷关
  };

  const command = commandMap[direction];
  if (!command) {
    console.error(`未知的云台控制指令: ${direction}`);
    return "A50F0199999999"; // 返回默认指令
  }

  console.log(`发送PTZ指令: ${direction} -> ${command}`);

  return command;
};

const PTZControl: React.FC<PTZControlProps> = ({
  deviceId,
  channelId,
  visible,
  onClose,
}) => {
  const [controlMode, setControlMode] = useState<"continue" | "click">(
    "continue"
  );
  const [activeDirection, setActiveDirection] = useState<string | null>(null);
  const [isControlling, setIsControlling] = useState(false);

  const lastCommandRef = useRef<string>("");

  // 发送云台控制指令
  const sendPTZCommand = useCallback(
    async (direction: PTZDirection) => {
      try {
        const command = generatePTZCommand(direction);

        // 避免重复发送相同指令
        if (command === lastCommandRef.current) {
          return;
        }

        lastCommandRef.current = command;

        console.log(`发送云台控制指令: ${direction}, 指令: ${command}`);

        await ptzControl(deviceId, channelId, command);
      } catch (error) {
        console.error("云台控制失败:", error);
        message.error("云台控制失败");
      }
    },
    [deviceId, channelId]
  );

  // 处理方向控制开始
  const handleDirectionStart = useCallback(
    async (direction: PTZDirection) => {
      if (isControlling) return;

      setActiveDirection(direction);
      setIsControlling(true);

      if (controlMode === "continue") {
        // 连动模式：只发送移动指令，不自动停止
        await sendPTZCommand(direction);
        // 连动模式下保持控制状态，等待用户手动停止
      } else {
        // 点动模式：发送移动指令，固定0.25秒后发送停止指令
        try {
          // 立即发送移动指令，不等待响应
          sendPTZCommand(direction);

          // 固定等待0.25秒后发送停止指令
          setTimeout(async () => {
            try {
              await sendPTZCommand("stop");
            } catch (stopError) {
              console.error("发送停止指令失败:", stopError);
            } finally {
              // 重置状态
              setActiveDirection(null);
              setIsControlling(false);
              lastCommandRef.current = "";
            }
          }, 250);
        } catch (error) {
          console.error("点动模式控制失败:", error);
          // 出错时也要重置状态
          setActiveDirection(null);
          setIsControlling(false);
          lastCommandRef.current = "";
        }
      }
    },
    [controlMode, sendPTZCommand, isControlling]
  );

  // 处理方向控制停止（仅用于手动点击停止按钮）
  const _handleDirectionStop = useCallback(() => {
    // 手动停止，只在连动模式且正在控制时有效
    if (controlMode === "continue" && isControlling) {
      sendPTZCommand("stop");
      setActiveDirection(null);
      setIsControlling(false);
      lastCommandRef.current = "";
    }
  }, [sendPTZCommand, isControlling, controlMode]);

  // 处理镜头控制
  const handleZoomControl = useCallback(
    (direction: "in" | "out") => {
      const command = direction === "in" ? "zoomin" : "zoomout";
      // 镜头控制没有停止指令，连动和点动模式都是发送一次指令
      sendPTZCommand(command);
    },
    [sendPTZCommand]
  );

  // 处理焦距控制
  const handleFocusControl = useCallback(
    (direction: "in" | "out") => {
      const command = direction === "in" ? "focusin" : "focusout";
      // 焦距控制没有停止指令，连动和点动模式都是发送一次指令
      sendPTZCommand(command);
    },
    [sendPTZCommand]
  );

  // 处理光圈控制
  const handleApertureControl = useCallback(
    (direction: "open" | "close") => {
      const command = direction === "open" ? "apertureopen" : "apertureclose";
      // 光圈控制没有停止指令，连动和点动模式都是发送一次指令
      sendPTZCommand(command);
    },
    [sendPTZCommand]
  );

  // 处理雨刷控制
  const handleWiperControl = useCallback(
    (action: "on" | "off") => {
      const command = action === "on" ? "wiperon" : "wiperoff";
      sendPTZCommand(command);
    },
    [sendPTZCommand]
  );

  // 方向控制按钮通用属性
  const getDirectionButtonProps = (direction: PTZDirection) => ({
    onClick: () => {
      handleDirectionStart(direction).catch(console.error);
    },
    className: `direction-btn ${activeDirection === direction ? "active" : ""}`,
    type: "primary" as const,
    disabled: isControlling && activeDirection !== direction,
  });

  if (!visible) {
    return null;
  }

  return (
    <div className="ptz-control-container">
      <div className="ptz-control-header">
        <span>云台控制</span>
        {onClose && (
          <Button type="text" size="small" onClick={onClose}>
            ×
          </Button>
        )}
      </div>

      <div className="ptz-control-content">
        {/* 方向控制 */}
        <div className="direction-control">
          <div className="direction-grid">
            <div className="direction-row">
              <Button
                icon={<span className="diagonal-icon">↖</span>}
                {...getDirectionButtonProps("leftup")}
              />
              <Button
                icon={<UpOutlined />}
                {...getDirectionButtonProps("up")}
              />
              <Button
                icon={<span className="diagonal-icon">↗</span>}
                {...getDirectionButtonProps("rightup")}
              />
            </div>
            <div className="direction-row">
              <Button
                icon={<LeftOutlined />}
                {...getDirectionButtonProps("left")}
              />
              <Button
                icon={<PauseOutlined />}
                onClick={() => {
                  // 停止按钮：连动模式可用，点动模式禁用
                  if (controlMode === "continue") {
                    sendPTZCommand("stop");
                    setActiveDirection(null);
                    setIsControlling(false);
                    lastCommandRef.current = "";
                  }
                }}
                className="stop-btn"
                disabled={controlMode === "click"}
              />
              <Button
                icon={<RightOutlined />}
                {...getDirectionButtonProps("right")}
              />
            </div>
            <div className="direction-row">
              <Button
                icon={<span className="diagonal-icon">↙</span>}
                {...getDirectionButtonProps("leftdown")}
              />
              <Button
                icon={<DownOutlined />}
                {...getDirectionButtonProps("down")}
              />
              <Button
                icon={<span className="diagonal-icon">↘</span>}
                {...getDirectionButtonProps("rightdown")}
              />
            </div>
          </div>
        </div>

        {/* 镜头、焦距、光圈控制 */}
        <div className="lens-control">
          <div className="control-group">
            <Button
              icon={<PlusOutlined />}
              onClick={() => handleZoomControl("in")}
              size="small"
            />
            <span>镜头</span>
            <Button
              icon={<MinusOutlined />}
              onClick={() => handleZoomControl("out")}
              size="small"
            />
          </div>

          <div className="control-group">
            <Button
              icon={<FullscreenOutlined />}
              onClick={() => handleFocusControl("in")}
              size="small"
            />
            <span>焦距</span>
            <Button
              icon={<FullscreenExitOutlined />}
              onClick={() => handleFocusControl("out")}
              size="small"
            />
          </div>

          <div className="control-group">
            <Button
              icon={<PlusOutlined />}
              onClick={() => handleApertureControl("open")}
              size="small"
            />
            <span>光圈</span>
            <Button
              icon={<MinusOutlined />}
              onClick={() => handleApertureControl("close")}
              size="small"
            />
          </div>
        </div>

        {/* 控制模式和雨刷选项 */}
        <Tabs defaultActiveKey="mode" size="small">
          <TabPane tab="控制模式" key="mode">
            <RadioGroup
              value={controlMode}
              onChange={(e) => setControlMode(e.target.value)}
              size="small"
            >
              <Radio value="continue">连动</Radio>
              <Radio value="click">点动</Radio>
            </RadioGroup>
          </TabPane>
          <TabPane tab="雨刷" key="wiper">
            <div className="wiper-control">
              <Space
                direction="vertical"
                size="small"
                style={{ width: "100%" }}
              >
                <Button
                  size="small"
                  block
                  onClick={() => handleWiperControl("on")}
                >
                  开启雨刷
                </Button>
                <Button
                  size="small"
                  block
                  onClick={() => handleWiperControl("off")}
                >
                  关闭雨刷
                </Button>
              </Space>
            </div>
          </TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default PTZControl;
