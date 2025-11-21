import type { StreamItem, SubscriberItem } from "@/services/video";
import { getStreamInfo, getSubscribers } from "@/services/video";
import { useEffect, useRef, useState } from "react";

export type MetricPoint = {
  ts: number;
  upBps?: number;
  downBps?: number;
  fps?: number;
  gop?: number;
  subscribers?: number;
};

const POLL_INTERVAL_MS = 3000;

const useDetailPolling = (visible: boolean, path: string) => {
  const [info, setInfo] = useState<StreamItem | null>(null);
  const [subscribers, setSubscribers] = useState<SubscriberItem[]>([]);
  const [metrics, setMetrics] = useState<MetricPoint[]>([]);
  const timerRef = useRef<number | null>(null);

  useEffect(() => {
    const stopTimer = () => {
      if (timerRef.current) {
        clearInterval(timerRef.current as any);
        timerRef.current = null;
      }
    };

    if (visible && path) {
      const fetchAll = async () => {
        try {
          const [infoResp, subsResp] = await Promise.all([
            getStreamInfo(path, { format: "json" }),
            getSubscribers(path, { format: "json" }),
          ]);
          if (infoResp && infoResp.code === 0) setInfo(infoResp.data);
          if (infoResp && infoResp.code === 0) {
            const vt = infoResp.data?.videoTrack;
            const sample: MetricPoint = {
              ts: Date.now(),
              upBps: typeof vt?.bpsOut === "number" ? vt?.bpsOut : undefined,
              downBps: typeof vt?.bps === "number" ? vt?.bps : undefined,
              fps: typeof vt?.fps === "number" ? vt?.fps : undefined,
              gop:
                typeof vt?.gop === "number"
                  ? vt?.gop
                  : typeof infoResp.data?.gop === "number"
                  ? infoResp.data?.gop
                  : undefined,
              subscribers:
                typeof infoResp.data?.subscribers === "number"
                  ? infoResp.data?.subscribers
                  : undefined,
            };
            setMetrics((prev) => {
              const next = [...prev, sample];
              if (next.length > 120) next.splice(0, next.length - 120);
              return next;
            });
          }
          if (subsResp && subsResp.code === 0)
            setSubscribers(subsResp.data || []);
        } catch (_e) {
          // ignore
        }
      };
      fetchAll();
      stopTimer();
      timerRef.current = window.setInterval(fetchAll, POLL_INTERVAL_MS);
      return () => {
        stopTimer();
      };
    }

    setInfo(null);
    setSubscribers([]);
    setMetrics([]);
    stopTimer();
    return () => {};
  }, [visible, path]);

  return { info, subscribers, metrics };
};

export default useDetailPolling;
