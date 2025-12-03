import type { StreamItem, SubscriberItem } from "@/services/video";
import { Line } from "@ant-design/plots";
import { ProCard, ProTable } from "@ant-design/pro-components";
import {
  Badge,
  Descriptions,
  Divider,
  Drawer,
  Space,
  Tag,
  Typography,
  theme,
} from "antd";
import dayjs from "dayjs";
import React from "react";
import styles from "./DetailDrawer.less";

export type MetricPoint = {
  ts: number;
  upBps?: number;
  downBps?: number;
  fps?: number;
  gop?: number;
  subscribers?: number;
};

interface DetailDrawerProps {
  open: boolean;
  path: string;
  info: StreamItem | null;
  subscribers: SubscriberItem[];
  metrics: MetricPoint[];
  autoRefresh: boolean;
  onClose: () => void;
}

const stateBadgeStatus = (state: number) => {
  if (state >= 2) return "success";
  if (state === 1) return "warning";
  return "default";
};

const typeTag = (type?: string) => {
  if (!type) return <Tag>unknown</Tag>;
  const color = type === "live" ? "green" : type === "vod" ? "blue" : "default";
  return <Tag color={color}>{type}</Tag>;
};

const formatTrack = (track?: StreamItem["videoTrack"] | null) => {
  if (!track) return "-";
  const codec = track.codec || "";
  const fps = track.fps ? `${track.fps}fps` : "";
  const res =
    track.width && track.height ? `${track.width}x${track.height}` : "";
  return [codec, res, fps].filter(Boolean).join(" · ");
};

const DetailDrawer: React.FC<DetailDrawerProps> = ({
  open,
  path,
  info,
  subscribers,
  metrics,
  autoRefresh,
  onClose,
}) => {
  const { token } = theme.useToken();
  const toTime = (ts: number) => dayjs(ts).format("HH:mm:ss");
  const fpsData = metrics
    .filter((p) => typeof p.fps === "number")
    .map((p) => ({ time: toTime(p.ts), value: p.fps as number }));
  const gopData = metrics
    .filter((p) => typeof p.gop === "number")
    .map((p) => ({ time: toTime(p.ts), value: p.gop as number }));
  const subsData = metrics
    .filter((p) => typeof p.subscribers === "number")
    .map((p) => ({ time: toTime(p.ts), value: p.subscribers as number }));
  const bitrateData = metrics.flatMap((p) => {
    const time = toTime(p.ts);
    const points: { time: string; type: string; value: number }[] = [];
    if (typeof p.upBps === "number")
      points.push({
        time,
        type: "上行",
        value: Math.round((p.upBps / 1024) * 100) / 100,
      });
    if (typeof p.downBps === "number")
      points.push({
        time,
        type: "下行",
        value: Math.round((p.downBps / 1024) * 100) / 100,
      });
    return points;
  });

  return (
    <Drawer
      title={
        <Space>
          <span>流详情</span>
          {path ? <Tag color="geekblue">{path}</Tag> : null}
        </Space>
      }
      placement="right"
      width="85vw"
      open={open}
      onClose={onClose}
      destroyOnHidden={false}
      extra={
        <Space>
          <Badge
            title={autoRefresh ? "列表自动刷新中" : "列表已暂停自动刷新"}
            status={autoRefresh ? "processing" : "default"}
          />
        </Space>
      }
    >
      <Descriptions
        bordered
        size="small"
        column={2}
        className={styles.descriptionsSpacing}
      >
        <Descriptions.Item label="状态" span={1}>
          {typeof info?.state === "number" ? (
            <Badge
              status={stateBadgeStatus(info.state) as any}
              text={
                info.state >= 2 ? "推流中" : info.state === 1 ? "空闲" : "停止"
              }
            />
          ) : (
            "-"
          )}
        </Descriptions.Item>
        <Descriptions.Item label="订阅数" span={1}>
          {info?.subscribers ?? "-"}
        </Descriptions.Item>
        <Descriptions.Item label="类型" span={1}>
          {typeTag(info?.type)}
        </Descriptions.Item>
        <Descriptions.Item label="插件" span={1}>
          {info?.pluginName || "-"}
        </Descriptions.Item>
        <Descriptions.Item label="开始时间" span={1}>
          {info?.startTime
            ? dayjs(info.startTime).format("YYYY-MM-DD HH:mm:ss")
            : "-"}
        </Descriptions.Item>
        <Descriptions.Item label="暂停" span={1}>
          {info?.isPaused ? "是" : "否"}
        </Descriptions.Item>
        <Descriptions.Item label="速度" span={1}>
          {info?.speed ?? "-"}
        </Descriptions.Item>
        <Descriptions.Item label="缓冲" span={1}>
          {info?.bufferTime ?? "-"}
        </Descriptions.Item>
        <Descriptions.Item label="停空闲" span={1}>
          {info?.stopOnIdle ? "是" : "否"}
        </Descriptions.Item>
        <Descriptions.Item label="录制" span={2}>
          {info?.recording && info.recording.length > 0
            ? info.recording.join(",")
            : "-"}
        </Descriptions.Item>
        <Descriptions.Item label="音频" span={2}>
          {formatTrack(info?.audioTrack)}
        </Descriptions.Item>
        <Descriptions.Item label="视频" span={2}>
          {formatTrack(info?.videoTrack)}
        </Descriptions.Item>
        <Descriptions.Item label="视频BPS" span={1}>
          {(() => {
            const up = info?.videoTrack?.bpsOut;
            const down = info?.videoTrack?.bps;
            const fmt = (v?: number) =>
              typeof v === "number" ? `${(v / 1024).toFixed(2)} KB` : "-";
            if (up == null && down == null) return "-";
            return `↑${fmt(up)} ↓${fmt(down)}`;
          })()}
        </Descriptions.Item>
        <Descriptions.Item label="视频FPS" span={1}>
          {info?.videoTrack?.fps ?? "-"}
        </Descriptions.Item>
        <Descriptions.Item label="分辨率" span={1}>
          {info?.videoTrack?.width && info?.videoTrack?.height
            ? `${info.videoTrack.width}x${info.videoTrack.height}`
            : "-"}
        </Descriptions.Item>
        <Descriptions.Item label="GOP" span={1}>
          {info?.videoTrack?.gop ?? info?.gop ?? "-"}
        </Descriptions.Item>
        <Descriptions.Item label="Meta" span={2}>
          <Typography.Paragraph copyable className={styles.metaText}>
            {info?.meta || "-"}
          </Typography.Paragraph>
        </Descriptions.Item>
      </Descriptions>

      <Divider orientation="left">实时趋势</Divider>
      <ProCard gutter={16} wrap>
        <ProCard colSpan="50%" title="视频码率（KB/s）">
          <Line
            data={bitrateData}
            xField="time"
            yField="value"
            seriesField="type"
            smooth
            color={[token.colorError, token.colorPrimary]}
            style={{ lineWidth: 2 }}
            animation={false}
            height={240}
            legend={{ position: "top" }}
            xAxis={{ type: "timeCat", tickCount: 6 }}
            yAxis={{ min: 0, title: "KB/s" }}
            meta={{ time: { alias: "时间" }, value: { alias: "码率 (KB/s)" } }}
            point={{ size: 2, style: { stroke: undefined } }}
          />
        </ProCard>
        <ProCard colSpan="50%" title="帧率(FPS)">
          <Line
            data={fpsData}
            xField="time"
            yField="value"
            shapeField="smooth"
            style={{ stroke: token.colorSuccess, lineWidth: 2 }}
            animation={false}
            height={240}
            xAxis={{ type: "timeCat", tickCount: 6 }}
            yAxis={{ min: 0, title: "FPS" }}
            meta={{
              time: { alias: "时间" },
              value: { alias: "FPS", formatter: (v: number) => v.toFixed(0) },
            }}
            point={{
              size: 3,
              style: { fill: token.colorSuccess, stroke: token.colorSuccess },
            }}
          />
        </ProCard>
        <ProCard colSpan="50%" title="GOP">
          <Line
            data={gopData}
            xField="time"
            yField="value"
            shapeField="smooth"
            style={{ stroke: token.colorWarning, lineWidth: 2 }}
            animation={false}
            height={240}
            xAxis={{ type: "timeCat", tickCount: 6 }}
            yAxis={{ min: 0, title: "GOP" }}
            meta={{
              time: { alias: "时间" },
              value: { alias: "GOP", formatter: (v: number) => v.toFixed(0) },
            }}
            point={{
              size: 3,
              style: { fill: token.colorWarning, stroke: token.colorWarning },
            }}
          />
        </ProCard>
        <ProCard colSpan="50%" title="订阅者数量">
          <Line
            data={subsData}
            xField="time"
            yField="value"
            shapeField="smooth"
            style={{ stroke: token.colorInfo, lineWidth: 2 }}
            animation={false}
            height={240}
            xAxis={{ type: "timeCat", tickCount: 6 }}
            yAxis={{ min: 0, title: "订阅数" }}
            meta={{
              time: { alias: "时间" },
              value: {
                alias: "订阅数",
                formatter: (v: number) => v.toFixed(0),
              },
            }}
            point={{
              size: 3,
              style: { fill: token.colorInfo, stroke: token.colorInfo },
            }}
          />
        </ProCard>
      </ProCard>

      <Divider orientation="left">订阅者</Divider>
      <ProTable<SubscriberItem>
        size="small"
        rowKey={(r) => String(r.id)}
        search={false}
        options={false}
        toolBarRender={false}
        dataSource={subscribers}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
        }}
        columns={[
          { title: "插件", dataIndex: "pluginName", width: 110 },
          { title: "类型", dataIndex: "type", width: 90 },
          {
            title: "开始时间",
            dataIndex: "startTime",
            width: 180,
            render: (_, r) =>
              r.startTime
                ? dayjs(r.startTime).format("YYYY-MM-DD HH:mm:ss")
                : "-",
          },
          {
            title: "远端地址",
            dataIndex: "remoteAddr",
            width: 200,
            ellipsis: true,
          },
          {
            title: "状态",
            dataIndex: "videoReader",
            width: 120,
            render: (_, r) =>
              typeof r.videoReader?.state === "number" ? (
                <Badge
                  status={stateBadgeStatus(r.videoReader.state) as any}
                  text={
                    r.videoReader.state >= 2
                      ? "推流中"
                      : r.videoReader.state === 1
                      ? "空闲"
                      : "停止"
                  }
                />
              ) : (
                "-"
              ),
          },
          {
            title: "序号",
            dataIndex: "videoReader",
            width: 100,
            render: (_, r) => r.videoReader?.sequence ?? "-",
          },
          {
            title: "时间戳",
            dataIndex: "videoReader",
            width: 120,
            render: (_, r) => r.videoReader?.timestamp ?? "-",
          },
          {
            title: "延迟",
            dataIndex: "videoReader",
            width: 90,
            render: (_, r) => r.videoReader?.delay ?? "-",
          },
          {
            title: "BPS",
            dataIndex: "videoReader",
            width: 110,
            render: (_, r) =>
              typeof r.videoReader?.bps === "number"
                ? `${(r.videoReader.bps / 1024).toFixed(2)} KB`
                : "-",
          },
          { title: "缓冲", dataIndex: "bufferTime", width: 100 },
          { title: "SubMode", dataIndex: "subMode", width: 90 },
          { title: "SyncMode", dataIndex: "syncMode", width: 90 },
          {
            title: "Meta",
            dataIndex: "meta",
            ellipsis: true,
            render: (_, r) => (
              <Typography.Paragraph
                copyable={{ text: r.meta || "" }}
                className={styles.subscriberMetaText}
              >
                {r.meta || "-"}
              </Typography.Paragraph>
            ),
          },
        ]}
        scroll={{ x: 1400 }}
      />
    </Drawer>
  );
};

export default DetailDrawer;
