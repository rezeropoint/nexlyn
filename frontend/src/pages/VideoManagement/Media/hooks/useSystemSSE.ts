import { useEffect, useRef, useState } from "react";

export interface SystemMemoryInfo {
  total: number;
  free: number;
  used: number;
  usage: number;
}

export interface SystemHardDiskInfo {
  total: number;
  free: number;
  used: number;
  usage: number;
}

export interface SystemNetworkInfo {
  name: string;
  receive: number;
  sent: number;
  receiveSpeed?: number;
  sentSpeed?: number;
}

export interface SystemSSEPayload {
  memory?: SystemMemoryInfo;
  cpuUsage?: number;
  hardDisk?: SystemHardDiskInfo;
  netWork?: SystemNetworkInfo[];
  streamCount?: number;
  pullCount?: number;
}

const useSystemSSE = (url: string = "/api/summary/sse") => {
  const [systemData, setSystemData] = useState<SystemSSEPayload | undefined>();
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const sseRef = useRef<EventSource | null>(null);
  const retryCountRef = useRef(0);
  const maxRetries = 5;

  const connectSSE = () => {
    try {
      if (sseRef.current) {
        sseRef.current.close();
        sseRef.current = null;
      }

      const es = new EventSource(url);
      sseRef.current = es;

      es.onopen = () => {
        setIsConnected(true);
        setError(null);
        retryCountRef.current = 0;
      };

      es.onmessage = (evt) => {
        try {
          const data = JSON.parse(evt.data || "{}");
          setSystemData(data);
        } catch (parseError) {
          console.warn("解析SSE数据失败:", parseError);
        }
      };

      es.onerror = () => {
        setIsConnected(false);
        es.close();
        sseRef.current = null;

        // 重试连接
        if (retryCountRef.current < maxRetries) {
          retryCountRef.current += 1;
          const retryDelay = Math.min(1000 * 2 ** retryCountRef.current, 30000);
          setTimeout(() => {
            if (retryCountRef.current <= maxRetries) {
              connectSSE();
            }
          }, retryDelay);
          setError(
            `连接断开，${retryDelay / 1000}秒后重试 (${
              retryCountRef.current
            }/${maxRetries})`
          );
        } else {
          setError("连接失败，已达到最大重试次数");
        }
      };
    } catch (_connectError) {
      setError("无法建立SSE连接");
      setIsConnected(false);
    }
  };

  useEffect(() => {
    connectSSE();

    return () => {
      if (sseRef.current) {
        sseRef.current.close();
        sseRef.current = null;
      }
    };
  }, [url]);

  // 手动重新连接
  const reconnect = () => {
    retryCountRef.current = 0;
    setError(null);
    connectSSE();
  };

  return {
    systemData,
    isConnected,
    error,
    reconnect,
  };
};

export default useSystemSSE;
