import { Scene } from "@antv/l7";
import { GaodeMap } from "@antv/l7-maps";
import React, { useEffect, useRef } from "react";
import "./index.less";

const MapView: React.FC = () => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const sceneRef = useRef<Scene | null>(null);

  useEffect(() => {
    if (!mapContainerRef.current) return;

    // 创建地图场景
    const scene = new Scene({
      id: mapContainerRef.current,
      map: new GaodeMap({
        token: "ab152f9bd5f6122e8e2339c65f6c94b3",
        style: "light", // 使用浅色主题
        center: [105, 35], // 中国中心区域
        zoom: 4, // 初始缩放级别
        pitch: 0, // 倾斜角度
      }),
    });

    sceneRef.current = scene;

    // 清理函数
    return () => {
      if (sceneRef.current) {
        sceneRef.current.destroy();
        sceneRef.current = null;
      }
    };
  }, []);

  return <div ref={mapContainerRef} className="map-view-container" />;
};

export default MapView;
