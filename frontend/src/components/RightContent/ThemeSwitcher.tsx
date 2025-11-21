import type { ThemeMode } from "@/app";
import {
  BulbFilled,
  BulbOutlined,
  CheckOutlined,
  CompressOutlined,
} from "@ant-design/icons";
import type { MenuProps } from "antd";
import { Dropdown } from "antd";
import classNames from "classnames";
import React from "react";
import styles from "./ThemeSwitcher.less";

export interface ThemeSwitcherProps {
  mode: ThemeMode;
  onChange: (mode: ThemeMode) => void;
}

/**
 * 主题模式切换器
 * 支持4种模式：默认、暗色、紧凑、暗色紧凑
 * 使用彩色徽章+图标的特殊设计
 */
const ThemeSwitcher: React.FC<ThemeSwitcherProps> = ({ mode, onChange }) => {
  // 根据当前主题获取对应的图标
  const getCurrentIcon = () => {
    switch (mode) {
      case "light":
        return <BulbOutlined className={styles.themeIcon} />;
      case "dark":
        return <BulbFilled className={styles.themeIcon} />;
      case "compact":
        return <CompressOutlined className={styles.themeIcon} />;
      case "dark-compact":
        return <CompressOutlined className={styles.themeIcon} />;
      default:
        return <BulbOutlined className={styles.themeIcon} />;
    }
  };

  const getMenuIcon = (itemKey: string, defaultIcon: React.ReactNode) => {
    return itemKey === mode ? (
      <CheckOutlined style={{ color: "var(--ant-color-primary)" }} />
    ) : (
      defaultIcon
    );
  };

  const menuItems: MenuProps["items"] = [
    {
      key: "light",
      label: "默认主题",
      icon: getMenuIcon("light", <BulbOutlined />),
      onClick: () => onChange("light"),
    },
    {
      key: "dark",
      label: "暗色主题",
      icon: getMenuIcon("dark", <BulbFilled />),
      onClick: () => onChange("dark"),
    },
    {
      key: "compact",
      label: "紧凑模式",
      icon: getMenuIcon("compact", <CompressOutlined />),
      onClick: () => onChange("compact"),
    },
    {
      key: "dark-compact",
      label: "暗色紧凑",
      icon: getMenuIcon("dark-compact", <CompressOutlined />),
      onClick: () => onChange("dark-compact"),
    },
  ];

  return (
    <Dropdown
      menu={{ items: menuItems, selectedKeys: [mode] }}
      placement="bottomRight"
      trigger={["click"]}
    >
      <div className={styles.themeSwitcher}>
        {/* 彩色圆形徽章 - 根据主题显示不同的渐变色 */}
        <div className={classNames(styles.themeIndicator, styles[mode])} />
        {/* 动态图标 - 根据主题显示不同的图标 */}
        {getCurrentIcon()}
      </div>
    </Dropdown>
  );
};

export default ThemeSwitcher;
