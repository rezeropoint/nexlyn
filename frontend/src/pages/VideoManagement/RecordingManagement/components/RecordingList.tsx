import { getRecordings } from "@/services/video";
import {
  BarChartOutlined,
  BarsOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import {
  Button,
  DatePicker,
  message,
  Radio,
  Space,
  Table,
  Tag,
  Tooltip,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import dayjs from "dayjs";
import React, { useCallback, useEffect, useState } from "react";
import type { ChannelInfo, RecordingItem } from "../index";
import styles from "./RecordingList.less";
import RecordingTimeline from "./RecordingTimeline";

const { RangePicker } = DatePicker;

interface RecordingListProps {
  selectedChannel: ChannelInfo | null;
  recordings: RecordingItem[];
  onRecordingsUpdate: (recordings: RecordingItem[]) => void;
  onPlayRecording: (recording: RecordingItem) => void;
  playingRecording: RecordingItem | null;
  loading: boolean;
  onLoadingChange: (loading: boolean) => void;
}

type ViewMode = "list" | "timeline";

const RecordingList: React.FC<RecordingListProps> = ({
  selectedChannel,
  recordings,
  onRecordingsUpdate,
  onPlayRecording,
  playingRecording,
  loading,
  onLoadingChange,
}) => {
  const [viewMode, setViewMode] = useState<ViewMode>("list");
  const [timeRange, setTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(
    [
      dayjs().startOf("day"), // 默认查询当天00:00
      dayjs(), // 到当前时间
    ]
  );
  const [pagination, setPagination] = useState<{
    current: number;
    pageSize: number;
  }>({ current: 1, pageSize: 10 });

  // 查询录像列表
  const fetchRecordings = useCallback(async () => {
    if (!selectedChannel) {
      message.warning("请先选择设备通道");
      return;
    }

    if (!timeRange || !timeRange[0] || !timeRange[1]) {
      message.warning("请选择查询时间范围");
      return;
    }

    onLoadingChange(true);
    try {
      const startTime = timeRange[0].valueOf();
      const endTime = timeRange[1].valueOf();

      const response = await getRecordings(
        selectedChannel.deviceId,
        selectedChannel.channelId,
        {
          startTime: startTime,
          endTime: endTime,
        }
      );

      if (response.code === 0) {
        const recordingList = response.data?.records || [];
        onRecordingsUpdate(recordingList);
        message.success(`查询到 ${recordingList.length} 条录像记录`);
      } else {
        message.error(response.message || "查询录像失败");
        onRecordingsUpdate([]);
      }
    } catch (error: any) {
      console.error("查询录像失败:", error);
      // 从错误对象中提取后端返回的错误信息
      const errorMessage =
        error?.data?.message ||
        error?.response?.data?.message ||
        error?.message ||
        "查询录像失败";
      message.error(errorMessage);
      onRecordingsUpdate([]);
    } finally {
      onLoadingChange(false);
    }
  }, [selectedChannel, timeRange, onRecordingsUpdate, onLoadingChange]);

  // 格式化时间显示
  const formatTime = (timeStr: string) => {
    return dayjs(timeStr).format("YYYY-MM-DD HH:mm:ss");
  };

  // 计算录像时长
  const getDuration = (startTime: string, endTime: string) => {
    const start = dayjs(startTime);
    const end = dayjs(endTime);
    const duration = end.diff(start, "second");

    if (duration < 60) {
      return `${duration}秒`;
    } else if (duration < 3600) {
      return `${Math.floor(duration / 60)}分${duration % 60}秒`;
    } else {
      const hours = Math.floor(duration / 3600);
      const minutes = Math.floor((duration % 3600) / 60);
      const seconds = duration % 60;
      return `${hours}时${minutes}分${seconds}秒`;
    }
  };

  // 表格列定义
  const columns: ColumnsType<RecordingItem> = [
    {
      title: "名称",
      dataIndex: "name",
      key: "name",
      width: 200,
      ellipsis: {
        showTitle: false,
      },
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          {text}
        </Tooltip>
      ),
    },
    {
      title: "开始时间",
      dataIndex: "startTime",
      key: "startTime",
      width: 180,
      render: formatTime,
      sorter: (a, b) =>
        dayjs(a.startTime).valueOf() - dayjs(b.startTime).valueOf(),
      defaultSortOrder: "descend",
    },
    {
      title: "结束时间",
      dataIndex: "endTime",
      key: "endTime",
      width: 180,
      render: formatTime,
    },
    {
      title: "时长",
      key: "duration",
      width: 120,
      render: (_, record) => getDuration(record.startTime, record.endTime),
    },
    {
      title: "类型",
      dataIndex: "type",
      key: "type",
      width: 80,
      render: (type: string) => (
        <Tag color={type === "all" ? "blue" : "green"}>
          {type === "all" ? "全部" : type}
        </Tag>
      ),
    },
    {
      title: "地址",
      dataIndex: "address",
      key: "address",
      width: 150,
      ellipsis: {
        showTitle: false,
      },
      render: (text: string) => (
        <Tooltip placement="topLeft" title={text}>
          {text}
        </Tooltip>
      ),
    },
    {
      title: "操作",
      key: "action",
      width: 120,
      fixed: "right",
      render: (_, record) => {
        const isPlaying =
          playingRecording?.startTime === record.startTime &&
          playingRecording?.endTime === record.endTime;

        return (
          <Space>
            <Button
              type={isPlaying ? "primary" : "default"}
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => onPlayRecording(record)}
              disabled={isPlaying}
            >
              {isPlaying ? "播放中" : "播放"}
            </Button>
          </Space>
        );
      },
    },
  ];

  // 当选中通道变化时，自动查询录像
  useEffect(() => {
    if (selectedChannel && timeRange) {
      fetchRecordings();
    }
  }, [selectedChannel, fetchRecordings]);

  // 当录像数据变化时，重置到第一页
  useEffect(() => {
    setPagination((prev) => ({ ...prev, current: 1 }));
  }, [recordings]);

  return (
    <div className={styles.container}>
      {/* 查询条件 */}
      <div className={styles.querySection}>
        <Space wrap>
          <RangePicker
            value={timeRange}
            onChange={(dates) =>
              setTimeRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)
            }
            showTime
            format="YYYY-MM-DD HH:mm:ss"
            placeholder={["开始时间", "结束时间"]}
            className={styles.rangePicker}
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={fetchRecordings}
            loading={loading}
            disabled={!selectedChannel}
          >
            查询录像
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={fetchRecordings}
            loading={loading}
            disabled={!selectedChannel}
          >
            刷新
          </Button>
          <Radio.Group
            value={viewMode}
            onChange={(e) => setViewMode(e.target.value)}
            size="small"
            className={styles.viewModeGroup}
          >
            <Radio.Button value="list">
              <BarsOutlined /> 列表视图
            </Radio.Button>
            <Radio.Button value="timeline">
              <BarChartOutlined /> 时间轴
            </Radio.Button>
          </Radio.Group>
        </Space>
      </div>

      {/* 录像显示区域 */}
      <div className={styles.recordingSection}>
        {viewMode === "list" ? (
          <Table
            columns={columns}
            dataSource={recordings}
            loading={loading}
            size="small"
            scroll={{ x: 1000 }}
            rowKey={(record) =>
              `${record.startTime}-${record.endTime}-${record.filePath}`
            }
            pagination={{
              total: recordings.length,
              current: pagination.current,
              pageSize: pagination.pageSize,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total, range) =>
                `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
              size: "small",
              onChange: (page, pageSize) =>
                setPagination({ current: page, pageSize }),
              onShowSizeChange: (_page, pageSize) =>
                setPagination({ current: 1, pageSize }),
            }}
            rowClassName={(record) => {
              const isPlaying =
                playingRecording?.startTime === record.startTime &&
                playingRecording?.endTime === record.endTime;
              return isPlaying ? "recording-playing-row" : "";
            }}
          />
        ) : (
          <RecordingTimeline
            recordings={recordings}
            onPlayRecording={onPlayRecording}
            playingRecording={playingRecording}
            timeRange={timeRange}
          />
        )}
      </div>
    </div>
  );
};

export default RecordingList;
