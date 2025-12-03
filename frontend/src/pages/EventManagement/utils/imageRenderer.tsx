/**
 * 图片渲染工具函数
 * 用于检测和渲染 Base64 编码或 URL 格式的图片
 */
import { Image } from "antd";
import React from "react";

/**
 * 检测值是否为 Base64 图片
 * 支持两种格式：
 * 1. data:image/xxx;base64,xxx 格式
 * 2. 纯 Base64 编码字符串（至少100字符）
 */
export const isBase64Image = (value: any): boolean => {
  if (typeof value !== "string") return false;
  return (
    value.startsWith("data:image/") || /^[A-Za-z0-9+/=]{100,}$/.test(value)
  );
};

/**
 * 检测值是否为图片 URL
 * 支持以下情况：
 * 1. 路径以图片扩展名结尾（如 /image.png?token=xxx）
 * 2. 查询参数中包含图片文件名（如 ?attname=image.png）
 */
export const isImageUrl = (value: any): boolean => {
  if (typeof value !== "string") return false;
  if (!value.startsWith("http://") && !value.startsWith("https://")) return false;

  const imageExtensions = /\.(jpg|jpeg|png|gif|webp|bmp|svg)/i;

  // 检测路径是否以图片扩展名结尾
  try {
    const url = new URL(value);
    if (imageExtensions.test(url.pathname)) return true;

    // 检测查询参数中是否包含图片文件名（如 attname=xxx.png）
    for (const paramValue of url.searchParams.values()) {
      if (imageExtensions.test(paramValue)) return true;
    }
  } catch {
    // URL 解析失败，使用简单正则匹配
    return imageExtensions.test(value);
  }

  return false;
};

/**
 * 检测值是否为图片（Base64 或 URL）
 */
export const isImage = (value: any): boolean => {
  return isBase64Image(value) || isImageUrl(value);
};

/**
 * 格式化 Base64 图片源
 * 如果不包含 data:image 前缀，自动添加 PNG 格式前缀
 */
export const formatBase64Src = (value: string): string => {
  return value.startsWith("data:") ? value : `data:image/png;base64,${value}`;
};

/**
 * 获取图片源
 * 根据值类型返回对应的图片源
 */
export const getImageSrc = (value: string): string => {
  if (isImageUrl(value)) {
    return value; // URL 直接使用
  }
  return formatBase64Src(value); // Base64 需要格式化
};

/**
 * 渲染 Base64 图片组件
 * 使用 Ant Design Image 组件，支持点击预览
 */
export const renderBase64Image = (value: string, size = 60): React.ReactNode => (
  <Image
    src={formatBase64Src(value)}
    width={size}
    height={size}
    style={{ objectFit: "cover", borderRadius: 4 }}
    preview={{
      mask: "预览",
    }}
  />
);

/**
 * 渲染图片组件（支持 Base64 和 URL）
 * 使用 Ant Design Image 组件，支持点击预览
 */
export const renderImage = (value: string, size = 60): React.ReactNode => (
  <Image
    src={getImageSrc(value)}
    width={size}
    height={size}
    style={{ objectFit: "cover", borderRadius: 4 }}
    preview={{
      mask: "预览",
    }}
  />
);
