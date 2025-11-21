import {
  listDevices,
  queryTimeSeries,
  type TimeSeriesDataItem,
} from "@/services/iot";
import type { MessageInstance } from "antd/es/message/interface";
import dayjs from "dayjs";
import { useEffect, useState } from "react";
import { getAlarmLevel } from "../config";
import type { AlarmRecord, TimeRange } from "../types";

/**
 * 告警数据hook
 * @param orgId 组织ID
 * @param timeRange 时间范围
 * @param message message 实例
 */
export const useAlarmData = (
  orgId: string | undefined,
  timeRange: TimeRange,
  message: MessageInstance
) => {
  const [loading, setLoading] = useState(false);
  const [alarmRecords, setAlarmRecords] = useState<AlarmRecord[]>([]);
  const [hasError, setHasError] = useState(false);

  useEffect(() => {
    const fetchAlarmData = async () => {
      if (!orgId) {
        return;
      }

      setLoading(true);
      setHasError(false);
      try {
        // 1. 获取设备列表（用于设备名称和位置映射）
        const devicesRes = await listDevices({
          page: 1,
          pageSize: 1000,
          orgId,
        });

        const deviceMap = new Map();
        if (devicesRes.code === 0 && devicesRes.data?.list) {
          devicesRes.data.list.forEach((device) => {
            deviceMap.set(device.deviceId, {
              name: device.deviceName,
              location: device.location || "未知位置",
            });
          });
        }

        // 2. 计算查询时间范围
        const now = dayjs();
        let startTime: dayjs.Dayjs;
        switch (timeRange) {
          case "today":
            startTime = now.startOf("day");
            break;
          case "week":
            startTime = now.subtract(7, "day");
            break;
          case "month":
            startTime = now.subtract(30, "day");
            break;
          default:
            startTime = now.startOf("day");
        }

        // 3. 查询时序数据（温度、湿度、烟雾）
        const res = await queryTimeSeries({
          orgId,
          fieldNames: ["temperature", "humidity", "smoke"],
          startTime: startTime.toISOString(),
          endTime: now.toISOString(),
          limit: 1000,
          orderBy: "timestamp",
          orderDir: "desc",
        });

        if (res.code === 0 && res.data?.list) {
          // 4. 分析数据并生成告警记录
          const alarms: AlarmRecord[] = [];
          res.data.list.forEach((item: TimeSeriesDataItem) => {
            const level = getAlarmLevel(item.fieldName, item.value as number);
            if (level) {
              const deviceInfo = deviceMap.get(item.deviceId) || {
                name: item.deviceId,
                location: "未知位置",
              };

              let alarmType = "环境异常";
              let msg = "";
              if (item.fieldName === "temperature") {
                alarmType = "环境异常";
                msg = `温度超过阈值：${item.value}℃`;
              } else if (item.fieldName === "humidity") {
                alarmType = "环境异常";
                msg = `湿度超过阈值：${item.value}%`;
              } else if (item.fieldName === "smoke") {
                alarmType = "烟雾告警";
                msg = `检测到烟雾浓度异常：${item.value}ppm`;
              }

              alarms.push({
                id: `${item.deviceId}-${item.fieldName}-${item.timestamp}`,
                deviceId: item.deviceId,
                deviceName: deviceInfo.name,
                alarmType,
                level,
                message: msg,
                location: deviceInfo.location,
                timestamp: item.timestamp,
              });
            }
          });

          setAlarmRecords(alarms);
        } else {
          setHasError(true);
        }
      } catch (_error) {
        setHasError(true);
        message.error("获取告警数据失败");
      } finally {
        setLoading(false);
      }
    };

    fetchAlarmData();
  }, [orgId, timeRange]);

  return {
    loading,
    alarmRecords,
    hasError,
  };
};
