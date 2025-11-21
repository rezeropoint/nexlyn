import { BulbOutlined } from "@ant-design/icons";
import { Tooltip, theme } from "antd";
import React from "react";

export interface ThemeSwitchProps {
  isDark: boolean;
  onChange: (isDark: boolean) => void;
}

export const ThemeSwitch: React.FC<ThemeSwitchProps> = ({
  isDark,
  onChange,
}) => {
  const { token } = theme.useToken();

  return (
    <Tooltip title={isDark ? "切换到亮色主题" : "切换到暗色主题"}>
      <div
        onClick={() => onChange(!isDark)}
        style={{
          display: "inline-flex",
          padding: "4px",
          fontSize: "18px",
          cursor: "pointer",
          color: "inherit",
        }}
      >
        <BulbOutlined
          style={{ color: isDark ? token.colorWarning : "inherit" }}
        />
      </div>
    </Tooltip>
  );
};
