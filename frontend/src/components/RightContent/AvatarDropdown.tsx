import {
  LogoutOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { history, useModel } from "@umijs/max";
import { Spin } from "antd";
import { createStyles } from "antd-style";
import type { MenuInfo } from "rc-menu/lib/interface";
import React, { useCallback } from "react";
import { flushSync } from "react-dom";
import HeaderDropdown from "../HeaderDropdown";

export type GlobalHeaderRightProps = {
  menu?: boolean;
  children?: React.ReactNode;
};

export const AvatarName = () => {
  const { initialState } = useModel("@@initialState");
  const { currentUser } = initialState || {};
  return <span className="anticon">{currentUser?.name}</span>;
};

const useStyles = createStyles(({ token }) => {
  return {
    action: {
      display: "flex",
      height: "48px",
      marginLeft: "auto",
      overflow: "hidden",
      alignItems: "center",
      padding: "0 8px",
      cursor: "pointer",
      borderRadius: token.borderRadius,
      "&:hover": {
        backgroundColor: token.colorBgTextHover,
      },
    },
  };
});

export const AvatarDropdown: React.FC<GlobalHeaderRightProps> = ({
  menu,
  children,
}) => {
  const { styles } = useStyles();

  const { initialState, setInitialState } = useModel("@@initialState");

  const onMenuClick = useCallback(
    (event: MenuInfo) => {
      const { key } = event;
      if (key === "logout") {
        // 直接清除token并跳转，不先清除用户状态以避免loading状态
        localStorage.removeItem("token");
        const { search, pathname } = window.location;

        // 清除用户状态
        flushSync(() => {
          setInitialState((s) => ({ ...s, currentUser: undefined }));
        });

        // 立即跳转到登录页
        history.replace({
          pathname: "/user/login",
          search: `?${new URLSearchParams({
            redirect: pathname + search,
          })}`,
        });
        return;
      }
      history.push(`/account/${key}`);
    },
    [setInitialState]
  );

  const loading = (
    <span className={styles.action}>
      <Spin
        size="small"
        style={{
          marginLeft: 8,
          marginRight: 8,
        }}
      />
    </span>
  );

  if (!initialState) {
    return loading;
  }

  const { currentUser } = initialState;

  if (!currentUser || !currentUser.name) {
    return loading;
  }

  const menuItems = [
    ...(menu
      ? [
          {
            key: "center",
            icon: <UserOutlined />,
            label: "个人中心",
          },
          {
            type: "divider" as const,
          },
        ]
      : []),
    {
      key: "logout",
      icon: <LogoutOutlined />,
      label: "退出登录",
    },
  ];

  return (
    <HeaderDropdown
      menu={{
        selectedKeys: [],
        onClick: onMenuClick,
        items: menuItems,
      }}
    >
      {children}
    </HeaderDropdown>
  );
};
