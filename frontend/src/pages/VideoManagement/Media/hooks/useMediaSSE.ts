import type { SummarySSEPayload } from "@/services/video";
import { useEffect, useRef, useState } from "react";

const useMediaSSE = () => {
  const [summary, setSummary] = useState<SummarySSEPayload | undefined>();
  const sseRef = useRef<EventSource | null>(null);

  useEffect(() => {
    try {
      if (sseRef.current) {
        sseRef.current.close();
        sseRef.current = null;
      }

      const es = new EventSource("/api/summary/sse");
      sseRef.current = es;

      es.onmessage = (evt) => {
        try {
          const data = JSON.parse(evt.data || "{}");
          setSummary(data);
        } catch (_e) {}
      };

      es.onerror = () => {
        es.close();
        sseRef.current = null;
      };
    } catch (_e) {}
    return () => {
      if (sseRef.current) {
        sseRef.current.close();
        sseRef.current = null;
      }
    };
  }, []);

  return summary;
};

export default useMediaSSE;
