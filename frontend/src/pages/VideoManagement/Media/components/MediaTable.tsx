import type { StreamItem, StreamTrackInfo } from "@/services/video";
import { getStreamList } from "@/services/video";
import { formatDateTime } from "@/utils/date";
import {
  InfoCircleOutlined,
  LinkOutlined,
  PlayCircleOutlined,
} from "@ant-design/icons";
import type { ActionType, ProColumns } from "@ant-design/pro-components";
import { ProTable } from "@ant-design/pro-components";
import { Badge, Button, message, Tag } from "antd";
import React, { useMemo, useRef } from "react";

export interface ListStats {
  streams?: number;
  subscribers?: number;
}

interface MediaTableProps {
  actionRef?: React.MutableRefObject<ActionType | undefined>;
  onPreview: (
    deviceId: string,
    channelId: string,
    channelName?: string
  ) => void;
  onShowUrl: (path: string) => void;
  onShowDetail: (record: StreamItem) => void;
  onStatsChange?: (stats: ListStats) => void;
}

const formatTrack = (track?: StreamTrackInfo | null) => {
  if (!track) return "-";
  const codec = track.codec || "";
  const fps = track.fps ? `${track.fps}fps` : "";
  const res =
    track.width && track.height ? `${track.width}x${track.height}` : "";
  return [codec, res, fps].filter(Boolean).join(" · ");
};

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

const MediaTable: React.FC<MediaTableProps> = ({
  actionRef,
  onPreview,
  onShowUrl,
  onShowDetail,
  onStatsChange,
}) => {
  const innerActionRef = useRef<ActionType>();

  const columns = useMemo<ProColumns<StreamItem>[]>(
    () => [
      {
        title: "所属插件",
        dataIndex: "pluginName",
        width: 110,
        ellipsis: true,
        hideInSearch: true,
        render: (_, r) => r.pluginName || "-",
      },
      {
        title: "StreamPath",
        dataIndex: "path",
        width: 280,
        copyable: true,
        ellipsis: true,
      },
      {
        title: "音频",
        dataIndex: "audioTrack",
        width: 160,
        hideInSearch: true,
        render: (_, r) => formatTrack(r.audioTrack),
      },
      {
        title: "视频",
        dataIndex: "videoTrack",
        width: 200,
        hideInSearch: true,
        render: (_, r) => formatTrack(r.videoTrack),
      },
      {
        title: "状态",
        dataIndex: "state",
        width: 100,
        valueType: "select",
        valueEnum: {
          "": { text: "全部" },
          "2": { text: "推流中" },
          "1": { text: "空闲" },
          "0": { text: "停止" },
        },
        render: (_, r) => (
          <Badge
            status={stateBadgeStatus(r.state) as any}
            text={r.state >= 2 ? "推流中" : r.state === 1 ? "空闲" : "停止"}
          />
        ),
      },
      {
        title: "类型",
        dataIndex: "type",
        width: 90,
        hideInSearch: true,
        render: (_, r) => typeTag(r.type),
      },
      {
        title: "订阅",
        dataIndex: "subscribers",
        width: 80,
        hideInSearch: true,
      },
      {
        title: "创建时间",
        dataIndex: "startTime",
        width: 160,
        valueType: "dateTime",
        hideInSearch: true,
        render: (_, r) => formatDateTime(r.startTime),
      },
      {
        title: "BPS",
        dataIndex: "videoTrack",
        width: 100,
        hideInSearch: true,
        render: (_, r) => {
          const fmt = (v?: number) => {
            if (typeof v !== "number") return "-";
            const kb = v / 1024;
            return `${kb.toFixed(2)} KB`;
          };
          const up = fmt(r.videoTrack?.bpsOut);
          const down = fmt(r.videoTrack?.bps);
          if (up === "-" && down === "-") return "-";
          return `↑${up} ↓${down}`;
        },
      },
      {
        title: "录制",
        dataIndex: "recording",
        width: 120,
        hideInSearch: true,
        render: (_, r) =>
          r.recording && r.recording.length > 0 ? r.recording.join(",") : "-",
      },
      {
        title: "操作",
        valueType: "option",
        width: 240,
        fixed: "right",
        render: (_, record) => [
          <Button
            key="detail"
            type="link"
            size="small"
            icon={<InfoCircleOutlined />}
            onClick={() => onShowDetail(record)}
          >
            详情
          </Button>,
          <Button
            key="url"
            type="link"
            size="small"
            icon={<LinkOutlined />}
            onClick={() => onShowUrl(record.path)}
          >
            播放地址
          </Button>,
          <Button
            key="preview"
            type="link"
            size="small"
            icon={<PlayCircleOutlined />}
            disabled={
              !(record && typeof record.state === "number" && record.state >= 2)
            }
            onClick={() => {
              try {
                const [deviceId, channelId] = (record.path || "").split("/");
                if (!deviceId || !channelId) {
                  message.warning("无法解析 StreamPath");
                  return;
                }
                onPreview(deviceId, channelId, record.path);
              } catch (_e) {
                message.error("预览操作失败");
              }
            }}
          >
            预览
          </Button>,
        ],
      },
    ],
    [onPreview, onShowUrl, onShowDetail]
  );

  return (
    <ProTable<StreamItem>
      headerTitle="流列表"
      actionRef={(ref) => {
        (innerActionRef as any).current = ref as any;
        if (actionRef) (actionRef as any).current = ref as any;
      }}
      rowKey={(r) => r.path}
      search={{ labelWidth: "auto", defaultCollapsed: true }}
      scroll={{ x: "max-content" }}
      tableLayout="fixed"
      request={async (params) => {
        try {
          const apiParams: any = { format: "json" };
          if (params.current) apiParams.page = params.current;
          if (params.pageSize) apiParams.count = params.pageSize;
          const response = await getStreamList(apiParams);
          if (response.code === 0) {
            const list = (response.data || []).map((it: any) => ({
              path: it.path,
              state: it.state,
              subscribers: it.subscribers,
              audioTrack: it.audioTrack ?? null,
              videoTrack: it.videoTrack ?? null,
              startTime: it.startTime,
              pluginName: it.pluginName,
              type: it.type,
              meta: it.meta,
              isPaused: it.isPaused,
              gop: it.gop,
              speed: it.speed,
              bufferTime: it.bufferTime,
              stopOnIdle: it.stopOnIdle,
              recording: Array.isArray(it.recording) ? it.recording : [],
            })) as StreamItem[];
            const filtered = params.state
              ? list.filter((i) => String(i.state) === String(params.state))
              : list;
            try {
              const streamCount = filtered.length;
              const subs = filtered.reduce(
                (acc, it) =>
                  acc +
                  (typeof it.subscribers === "number" ? it.subscribers : 0),
                0
              );
              onStatsChange?.({ streams: streamCount, subscribers: subs });
            } catch (_e) {}
            return {
              data: filtered,
              success: true,
              total: response.total ?? filtered.length,
            };
          }
          message.error((response as any).message || "获取流列表失败");
          onStatsChange?.({ streams: 0, subscribers: 0 });
          return { data: [], success: false, total: 0 };
        } catch (_e) {
          message.error("获取流列表失败");
          onStatsChange?.({ streams: 0, subscribers: 0 });
          return { data: [], success: false, total: 0 };
        }
      }}
      columns={columns}
      pagination={{
        pageSize: 10,
        showSizeChanger: true,
        showQuickJumper: true,
      }}
    />
  );
};

export default MediaTable;
