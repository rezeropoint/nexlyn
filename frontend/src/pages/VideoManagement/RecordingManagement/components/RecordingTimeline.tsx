import {
  MinusOutlined,
  PlusOutlined,
  RedoOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
} from "@ant-design/icons";
import { Button, Slider, Space, Tooltip, theme } from "antd";
import classNames from "classnames";
import dayjs from "dayjs";
import React, { useCallback, useMemo, useRef, useState } from "react";
import type { RecordingItem } from "../index";
import styles from "./RecordingTimeline.less";

interface RecordingTimelineProps {
  recordings: RecordingItem[];
  onPlayRecording: (recording: RecordingItem, customStartTime?: number) => void;
  playingRecording: RecordingItem | null;
  timeRange: [dayjs.Dayjs, dayjs.Dayjs] | null;
}

// 时间段接口 - 合并后的连续录像时间段
interface TimeSegment {
  startMs: number;
  endMs: number;
  type: "normal" | "alarm" | "mixed";
  files: RecordingItem[]; // 该时间段包含的原始文件
  leftPx: number;
  widthPx: number;
  color: string;
}

// 选择的时间范围
interface TimeSelection {
  startMs: number;
  endMs: number;
  leftPx: number;
  widthPx: number;
}

const RecordingTimeline: React.FC<RecordingTimelineProps> = ({
  recordings,
  onPlayRecording,
  playingRecording,
  timeRange,
}) => {
  // 使用 Ant Design 5 的 theme API
  const { token } = theme.useToken();

  const containerRef = useRef<HTMLDivElement>(null);
  const MIN_SCALE = 0.5;
  const MAX_SCALE = 16;
  const BASE_PX_PER_HOUR = 240; // 基准每小时像素，增加默认显示密度
  const [scale, setScale] = useState(2); // 增加默认缩放值
  const [selectedTimeRange, setSelectedTimeRange] =
    useState<TimeSelection | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<{
    x: number;
    timeMs: number;
  } | null>(null);
  const [hoverTime, setHoverTime] = useState<{
    timeMs: number;
    x: number;
  } | null>(null);

  const clamp = (v: number, min: number, max: number) =>
    Math.min(max, Math.max(min, v));

  // 合并录像文件为连续时间段
  const mergeRecordings = useCallback(
    (recordings: RecordingItem[]): TimeSegment[] => {
      if (recordings.length === 0) return [];

      // 按开始时间排序
      const sorted = [...recordings].sort(
        (a, b) => dayjs(a.startTime).valueOf() - dayjs(b.startTime).valueOf()
      );

      const segments: TimeSegment[] = [];
      let currentSegment: TimeSegment | null = null;

      for (const recording of sorted) {
        const startMs = dayjs(recording.startTime).valueOf();
        const endMs = dayjs(recording.endTime).valueOf();
        const isAlarm = recording.type === "alarm";

        if (!currentSegment) {
          // 创建第一个段
          currentSegment = {
            startMs,
            endMs,
            type: isAlarm ? "alarm" : "normal",
            files: [recording],
            leftPx: 0,
            widthPx: 0,
            color: isAlarm ? token.colorError : token.colorPrimary,
          };
        } else {
          // 检查是否可以合并（间隔小于5秒认为是连续的）
          const gap = startMs - currentSegment.endMs;
          if (gap <= 5000) {
            // 合并到当前段
            currentSegment.endMs = Math.max(currentSegment.endMs, endMs);
            currentSegment.files.push(recording);

            // 更新类型
            if (isAlarm && currentSegment.type !== "alarm") {
              currentSegment.type =
                currentSegment.type === "normal" ? "mixed" : "mixed";
            } else if (!isAlarm && currentSegment.type === "alarm") {
              currentSegment.type = "mixed";
            }

            // 更新颜色
            if (currentSegment.type === "mixed") {
              currentSegment.color = token.colorWarning;
            } else if (currentSegment.type === "alarm") {
              currentSegment.color = token.colorError;
            } else {
              currentSegment.color = token.colorPrimary;
            }
          } else {
            // 保存当前段，创建新段
            segments.push(currentSegment);
            currentSegment = {
              startMs,
              endMs,
              type: isAlarm ? "alarm" : "normal",
              files: [recording],
              leftPx: 0,
              widthPx: 0,
              color: isAlarm ? token.colorError : token.colorPrimary,
            };
          }
        }
      }

      // 添加最后一个段
      if (currentSegment) {
        segments.push(currentSegment);
      }

      return segments;
    },
    [token]
  );

  const computed = useMemo(() => {
    if (!timeRange || !timeRange[0] || !timeRange[1]) {
      return {
        startMs: 0,
        endMs: 0,
        durationMs: 0,
        pxPerMs: 0,
        contentWidth: 0,
        segments: [] as TimeSegment[],
        markers: [] as Array<{ time: number; leftPx: number; label: string }>,
      };
    }

    const startMs = timeRange[0].valueOf();
    const endMs = timeRange[1].valueOf();
    const durationMs = Math.max(1, endMs - startMs);
    const pxPerMs = (BASE_PX_PER_HOUR * scale) / (60 * 60 * 1000);
    const contentWidth = Math.max(800, Math.ceil(durationMs * pxPerMs));

    // 合并录像为时间段
    const mergedSegments = mergeRecordings(recordings);

    // 裁剪时间段到显示范围并计算像素位置
    const segments: TimeSegment[] = mergedSegments
      .map((segment) => {
        const clippedStartMs = Math.max(segment.startMs, startMs);
        const clippedEndMs = Math.min(segment.endMs, endMs);

        if (clippedEndMs <= clippedStartMs) {
          return null; // 段不在显示范围内
        }

        const leftPx = Math.max(0, (clippedStartMs - startMs) * pxPerMs);
        const widthPx = Math.max(2, (clippedEndMs - clippedStartMs) * pxPerMs);

        return {
          ...segment,
          startMs: clippedStartMs,
          endMs: clippedEndMs,
          leftPx,
          widthPx,
        };
      })
      .filter((seg): seg is TimeSegment => seg !== null);

    // 刻度
    const pxPerHour = BASE_PX_PER_HOUR * scale;
    let stepHours = 24;
    if (pxPerHour >= 280) stepHours = 1;
    else if (pxPerHour >= 140) stepHours = 2;
    else if (pxPerHour >= 70) stepHours = 4;
    else if (pxPerHour >= 35) stepHours = 6;
    else if (pxPerHour >= 20) stepHours = 12;
    else stepHours = 24;

    const firstMarker = dayjs(startMs).startOf("hour");
    const markers: Array<{ time: number; leftPx: number; label: string }> = [];
    for (
      let t = firstMarker.valueOf();
      t <= endMs;
      t += stepHours * 60 * 60 * 1000
    ) {
      const leftPx = Math.max(
        0,
        Math.min(contentWidth, (t - startMs) * pxPerMs)
      );
      const label = dayjs(t).format(
        stepHours < 24 ? "MM-DD HH:mm" : "YYYY-MM-DD"
      );
      markers.push({ time: t, leftPx, label });
    }

    return {
      startMs,
      endMs,
      durationMs,
      pxPerMs,
      contentWidth,
      segments,
      markers,
    };
  }, [recordings, timeRange, scale, mergeRecordings]);

  // 防抖处理
  const debounceRef = useRef<NodeJS.Timeout | null>(null);

  const zoomTo = useCallback(
    (factor: number, anchorClientX?: number) => {
      const nextScale = clamp(scale * factor, MIN_SCALE, MAX_SCALE);
      if (!timeRange || !timeRange[0]) {
        setScale(nextScale);
        return;
      }

      const startMs = timeRange[0].valueOf();
      const container = containerRef.current;
      if (!container) {
        setScale(nextScale);
        return;
      }

      const rect = container.getBoundingClientRect();
      const anchorX =
        typeof anchorClientX === "number"
          ? anchorClientX - rect.left + container.scrollLeft
          : container.scrollLeft + rect.width / 2;
      const anchorTime =
        startMs + anchorX / ((BASE_PX_PER_HOUR * scale) / (60 * 60 * 1000));

      // 添加过渡动画
      if (container) {
        container.style.transition = "transform 0.2s ease-out";
      }

      setScale(nextScale);

      // 在下一帧根据新scale调整scroll，保持锚点不动
      requestAnimationFrame(() => {
        const newPxPerMs = (BASE_PX_PER_HOUR * nextScale) / (60 * 60 * 1000);
        const newScrollLeft =
          (anchorTime - startMs) * newPxPerMs -
          (typeof anchorClientX === "number"
            ? anchorClientX - rect.left
            : rect.width / 2);
        container.scrollLeft = Math.max(0, newScrollLeft);

        // 移除过渡动画
        setTimeout(() => {
          if (container) {
            container.style.transition = "";
          }
        }, 200);
      });
    },
    [scale, timeRange]
  );

  // 直接设置缩放值
  const setScaleDirectly = useCallback((newScale: number) => {
    const nextScale = clamp(newScale, MIN_SCALE, MAX_SCALE);
    setScale(nextScale);
  }, []);

  const handleWheel = useCallback(
    (e: React.WheelEvent<HTMLDivElement>) => {
      if (!e.ctrlKey) return; // 仅在 Ctrl+滚轮 时缩放，避免影响普通滚动
      e.preventDefault();

      // 防抖处理
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }

      debounceRef.current = setTimeout(() => {
        const factor = e.deltaY < 0 ? 1.15 : 0.85;
        zoomTo(factor, e.clientX);
      }, 16); // 约60fps
    },
    [zoomTo]
  );

  // 处理鼠标按下开始拖拽
  const handleMouseDown = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (e.button !== 0) return; // 只处理左键

      const container = containerRef.current;
      if (!container || !computed.pxPerMs) return;

      const rect = container.getBoundingClientRect();
      const clickX = e.clientX - rect.left + container.scrollLeft;
      const timeMs = computed.startMs + clickX / computed.pxPerMs;

      setDragStart({ x: clickX, timeMs });
      setIsDragging(true);
      setSelectedTimeRange(null);

      e.preventDefault();
    },
    [computed]
  );

  // 处理鼠标移动拖拽
  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (!isDragging || !dragStart || !computed.pxPerMs) return;

      const container = containerRef.current;
      if (!container) return;

      const rect = container.getBoundingClientRect();
      const currentX = e.clientX - rect.left + container.scrollLeft;
      const currentTimeMs = computed.startMs + currentX / computed.pxPerMs;

      const startMs = Math.min(dragStart.timeMs, currentTimeMs);
      const endMs = Math.max(dragStart.timeMs, currentTimeMs);
      const startX = Math.min(dragStart.x, currentX);
      const width = Math.abs(currentX - dragStart.x);

      setSelectedTimeRange({
        startMs,
        endMs,
        leftPx: startX,
        widthPx: width,
      });
    },
    [isDragging, dragStart, computed]
  );

  // 处理鼠标抬起结束拖拽
  const handleMouseUp = useCallback(() => {
    if (!isDragging) return;

    setIsDragging(false);
    setDragStart(null);

    // 如果选择了有效的时间范围，播放该范围的录像
    if (selectedTimeRange && selectedTimeRange.widthPx > 20) {
      // 找到选择范围内的第一个录像文件
      const selectedSegment = computed.segments.find(
        (seg) =>
          seg.startMs <= selectedTimeRange.endMs &&
          seg.endMs >= selectedTimeRange.startMs
      );

      if (selectedSegment && selectedSegment.files.length > 0) {
        // 从选择的开始时间播放
        const playStartTime = Math.max(
          selectedTimeRange.startMs,
          selectedSegment.startMs
        );

        // 找到包含开始时间的录像文件
        const targetFile =
          selectedSegment.files.find((file) => {
            const fileStart = dayjs(file.startTime).valueOf();
            const fileEnd = dayjs(file.endTime).valueOf();
            return playStartTime >= fileStart && playStartTime <= fileEnd;
          }) || selectedSegment.files[0];

        onPlayRecording(targetFile, playStartTime);
      }
    }
  }, [isDragging, selectedTimeRange, computed.segments, onPlayRecording]);

  // 添加全局鼠标事件监听
  React.useEffect(() => {
    if (isDragging) {
      const handleGlobalMouseMove = (e: MouseEvent) => {
        const container = containerRef.current;
        if (!container || !dragStart || !computed.pxPerMs) return;

        const rect = container.getBoundingClientRect();
        const currentX = e.clientX - rect.left + container.scrollLeft;
        const currentTimeMs = computed.startMs + currentX / computed.pxPerMs;

        const startMs = Math.min(dragStart.timeMs, currentTimeMs);
        const endMs = Math.max(dragStart.timeMs, currentTimeMs);
        const startX = Math.min(dragStart.x, currentX);
        const width = Math.abs(currentX - dragStart.x);

        setSelectedTimeRange({
          startMs,
          endMs,
          leftPx: startX,
          widthPx: width,
        });
      };

      const handleGlobalMouseUp = () => {
        handleMouseUp();
      };

      document.addEventListener("mousemove", handleGlobalMouseMove);
      document.addEventListener("mouseup", handleGlobalMouseUp);

      return () => {
        document.removeEventListener("mousemove", handleGlobalMouseMove);
        document.removeEventListener("mouseup", handleGlobalMouseUp);
      };
    }
    return undefined;
  }, [isDragging, dragStart, computed, handleMouseUp]);

  // 处理时间轴点击播放
  const handleTimelineClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>, segment: TimeSegment) => {
      // 阻止事件冒泡到容器的拖拽事件
      e.stopPropagation();

      const container = containerRef.current;
      if (!container || !computed.pxPerMs) return;

      // 计算点击的精确时间点
      const rect = container.getBoundingClientRect();
      const clickX = e.clientX - rect.left + container.scrollLeft;
      const clickTimeMs = computed.startMs + clickX / computed.pxPerMs;

      // 确保点击时间在时间段范围内
      const playStartTime = Math.max(
        segment.startMs,
        Math.min(clickTimeMs, segment.endMs)
      );

      // 找到最合适的录像文件（包含该时间点的文件）
      const targetFile =
        segment.files.find((file) => {
          const fileStart = dayjs(file.startTime).valueOf();
          const fileEnd = dayjs(file.endTime).valueOf();
          return playStartTime >= fileStart && playStartTime <= fileEnd;
        }) || segment.files[0]; // 如果没找到，使用第一个文件

      // 播放指定时间点
      onPlayRecording(targetFile, playStartTime);
    },
    [computed, onPlayRecording]
  );

  // 处理鼠标悬停显示时间
  const handleMouseHover = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      const container = containerRef.current;
      if (!container || !computed.pxPerMs || isDragging) return;

      const rect = container.getBoundingClientRect();
      const hoverX = e.clientX - rect.left + container.scrollLeft;
      const hoverTimeMs = computed.startMs + hoverX / computed.pxPerMs;

      // 只在有录像的时间范围内显示时间提示
      const isInRecordingRange = computed.segments.some(
        (seg) => hoverTimeMs >= seg.startMs && hoverTimeMs <= seg.endMs
      );

      if (isInRecordingRange) {
        setHoverTime({ timeMs: hoverTimeMs, x: hoverX });
      } else {
        setHoverTime(null);
      }
    },
    [computed, isDragging]
  );

  // 鼠标离开时隐藏时间提示
  const handleMouseLeave = useCallback(() => {
    setHoverTime(null);
  }, []);

  // 自动定位到首个或正在播放的录像段
  const scrollToSegment = useCallback((seg?: TimeSegment) => {
    const container = containerRef.current;
    if (!container || !seg) return;
    const viewportWidth = container.clientWidth;
    const targetLeft = Math.max(
      0,
      seg.leftPx - Math.max(0, (viewportWidth - seg.widthPx) / 2)
    );
    container.scrollLeft = targetLeft;
  }, []);

  // 数据变化时定位到首个录像段
  React.useEffect(() => {
    if (computed.segments.length > 0) {
      scrollToSegment(computed.segments[0]);
    }
    // 仅在 recordings 或 timeRange 改变时初次定位
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [recordings, timeRange]);

  // 播放项变化时定位到对应段
  React.useEffect(() => {
    if (!playingRecording || computed.segments.length === 0) return;
    const seg = computed.segments.find((s) =>
      s.files.some(
        (file) =>
          file.startTime === playingRecording.startTime &&
          file.endTime === playingRecording.endTime
      )
    );
    if (seg) scrollToSegment(seg);
  }, [playingRecording, computed.segments, scrollToSegment]);

  // 格式化时长（支持字符串时间和毫秒时间戳）
  const formatDuration = (
    startTime: string | number,
    endTime: string | number
  ) => {
    const start =
      typeof startTime === "number" ? dayjs(startTime) : dayjs(startTime);
    const end = typeof endTime === "number" ? dayjs(endTime) : dayjs(endTime);
    const duration = end.diff(start, "second");

    if (duration < 60) {
      return `${duration}秒`;
    } else if (duration < 3600) {
      return `${Math.floor(duration / 60)}分${duration % 60}秒`;
    } else {
      const hours = Math.floor(duration / 3600);
      const minutes = Math.floor((duration % 3600) / 60);
      return `${hours}时${minutes}分`;
    }
  };

  // 检查是否正在播放
  const _isPlaying = (rec: RecordingItem) =>
    playingRecording?.startTime === rec.startTime &&
    playingRecording?.endTime === rec.endTime;

  if (!timeRange || recordings.length === 0) {
    return <div className={styles.emptyState}>暂无录像数据</div>;
  }

  return (
    <div className={styles.timelineContainer}>
      <div className={styles.timelineHeader}>
        <div className={styles.timelineTitle}>录像时间轴</div>

        {/* 缩放滑块 */}
        <div className={styles.scaleControls}>
          <MinusOutlined className={styles.scaleIcon} />
          <Slider
            min={MIN_SCALE}
            max={MAX_SCALE}
            step={0.1}
            value={scale}
            onChange={setScaleDirectly}
            className={styles.scaleSlider}
            tooltip={{
              formatter: (value) => `${value?.toFixed(1)}x`,
            }}
          />
          <PlusOutlined className={styles.scaleIcon} />
        </div>

        {/* 缩放比例显示 */}
        <div className={styles.scaleDisplay}>{scale.toFixed(1)}x</div>

        <div className={styles.spacer} />

        {/* 缩放按钮组 */}
        <Space>
          <Button
            size="small"
            icon={<ZoomOutOutlined />}
            onClick={() => zoomTo(0.8)}
            disabled={scale <= MIN_SCALE}
          >
            缩小
          </Button>
          <Button
            size="small"
            icon={<RedoOutlined />}
            onClick={() => setScaleDirectly(2)}
          >
            重置
          </Button>
          <Button
            size="small"
            icon={<ZoomInOutlined />}
            onClick={() => zoomTo(1.25)}
            disabled={scale >= MAX_SCALE}
          >
            放大
          </Button>
        </Space>
      </div>

      {/* 视口容器（可水平滚动） */}
      <div
        ref={containerRef}
        onWheel={handleWheel}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseLeave}
        className={classNames(styles.viewportContainer, {
          [styles.dragging]: isDragging,
          [styles.normal]: !isDragging,
        })}
      >
        {/* 内容宽度由缩放决定 */}
        <div
          className={styles.contentWrapper}
          style={{ width: `${computed.contentWidth}px` }}
        >
          {/* 顶部刻度 */}
          <div className={styles.markersContainer}>
            {computed.markers.map((m) => (
              <div
                key={`marker-${m.time}-${m.leftPx}`}
                className={styles.markerLabel}
                style={{ left: `${m.leftPx}px` }}
              >
                {m.label}
              </div>
            ))}
          </div>

          {/* 网格线 */}
          <div className={styles.timelineArea} onMouseMove={handleMouseHover}>
            {computed.markers.map((m) => (
              <div
                key={`grid-${m.time}-${m.leftPx}`}
                className={styles.gridLine}
                style={{ left: `${m.leftPx}px` }}
              />
            ))}

            {/* 悬停时间游标线 */}
            {hoverTime && (
              <div
                className={styles.hoverCursor}
                style={{ left: `${hoverTime.x}px` }}
              />
            )}

            {/* 悬停时间提示 */}
            {hoverTime && (
              <div
                className={styles.hoverTimeTooltip}
                style={{ left: `${hoverTime.x}px` }}
              >
                {dayjs(hoverTime.timeMs).format("HH:mm:ss")}
              </div>
            )}

            {/* 时间段（单轴显示） */}
            {computed.segments.length === 0 ? (
              <div className={styles.noRecordingHint}>该时间范围内无录像</div>
            ) : (
              computed.segments.map((seg) => {
                const isPlayingSegment =
                  playingRecording &&
                  seg.files.some(
                    (file) =>
                      file.startTime === playingRecording.startTime &&
                      file.endTime === playingRecording.endTime
                  );

                return (
                  <Tooltip
                    key={`seg-${seg.startMs}-${seg.endMs}`}
                    title={
                      <div>
                        <div>
                          <strong>录像时间段</strong>
                        </div>
                        <div>
                          开始:{" "}
                          {dayjs(seg.startMs).format("YYYY-MM-DD HH:mm:ss")}
                        </div>
                        <div>
                          结束: {dayjs(seg.endMs).format("YYYY-MM-DD HH:mm:ss")}
                        </div>
                        <div>
                          时长: {formatDuration(seg.startMs, seg.endMs)}
                        </div>
                        <div>
                          类型:{" "}
                          {seg.type === "normal"
                            ? "普通录像"
                            : seg.type === "alarm"
                            ? "报警录像"
                            : "混合录像"}
                        </div>
                        <div>文件数: {seg.files.length}个</div>
                        <div
                          style={{
                            marginTop: "6px",
                            color: token.colorPrimary,
                          }}
                        >
                          点击播放时间段
                        </div>
                      </div>
                    }
                    placement="top"
                  >
                    <div
                      className={classNames(styles.recordingSegment, {
                        [styles.normal]: seg.type === "normal",
                        [styles.alarm]: seg.type === "alarm",
                        [styles.mixed]: seg.type === "mixed",
                        [styles.playing]: isPlayingSegment,
                      })}
                      style={{
                        left: `${seg.leftPx}px`,
                        width: `${seg.widthPx}px`,
                      }}
                      onClick={(e) => handleTimelineClick(e, seg)}
                    >
                      {seg.widthPx > 40
                        ? formatDuration(seg.startMs, seg.endMs)
                        : ""}
                    </div>
                  </Tooltip>
                );
              })
            )}

            {/* 时间范围选择框 */}
            {selectedTimeRange && (
              <div
                className={styles.timeSelection}
                style={{
                  left: `${selectedTimeRange.leftPx}px`,
                  width: `${selectedTimeRange.widthPx}px`,
                }}
              >
                {/* 选择时间提示 */}
                {selectedTimeRange.widthPx > 100 && (
                  <div className={styles.selectionDuration}>
                    {formatDuration(
                      selectedTimeRange.startMs,
                      selectedTimeRange.endMs
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 图例 */}
      <div className={styles.legendContainer}>
        <div className={styles.legendItem}>
          <div className={classNames(styles.legendColor, styles.normal)} />
          <span>普通录像</span>
        </div>
        <div className={styles.legendItem}>
          <div className={classNames(styles.legendColor, styles.alarm)} />
          <span>报警录像</span>
        </div>
        <div className={styles.legendItem}>
          <div className={classNames(styles.legendColor, styles.mixed)} />
          <span>混合录像</span>
        </div>
        <div className={styles.hintText}>
          提示：点击时间轴任意位置播放该时间点的录像，拖拽选择时间范围播放，悬停显示精确时间
        </div>
      </div>
    </div>
  );
};

export default RecordingTimeline;
