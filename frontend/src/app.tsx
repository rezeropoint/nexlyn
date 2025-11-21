import {
  AvatarDropdown,
  AvatarName,
  Footer,
  Question,
  SelectLang,
} from "@/components";
import { ThemeSwitcher } from "@/components/RightContent";
import { getCurrentUserWithPermissions } from "@/services/user";
import { LinkOutlined } from "@ant-design/icons";
import type { Settings as LayoutSettings } from "@ant-design/pro-components";
import { SettingDrawer } from "@ant-design/pro-components";
import type { RequestConfig, RunTimeLayoutConfig } from "@umijs/max";
import { history, Link } from "@umijs/max";
import { App, ConfigProvider, theme } from "antd";
import { appList } from "../config/appList";
import defaultSettings from "../config/defaultSettings";
import { errorConfig } from "./requestErrorConfig";

const isDev = process.env.NODE_ENV === "development";
const loginPath = "/user/login";

// 主题配置本地存储key
const SETTINGS_STORAGE_KEY = "nexlyn-pro-settings";
const THEME_MODE_STORAGE_KEY = "nexlyn-theme-mode";

// 主题模式类型定义
export type ThemeMode = "light" | "dark" | "compact" | "dark-compact";

// 获取主题算法
const getAlgorithm = (mode: ThemeMode) => {
  switch (mode) {
    case "dark":
      return theme.darkAlgorithm;
    case "compact":
      return theme.compactAlgorithm;
    case "dark-compact":
      return [theme.darkAlgorithm, theme.compactAlgorithm];
    default:
      return theme.defaultAlgorithm;
  }
};

/**
 * @see https://umijs.org/docs/api/runtime-config#getinitialstate
 * */
export async function getInitialState(): Promise<{
  settings?: Partial<LayoutSettings>;
  currentUser?: API.CurrentUser;
  loading?: boolean;
  fetchUserInfo?: () => Promise<API.CurrentUser | undefined>;
  themeMode?: ThemeMode;
}> {
  const fetchUserInfo = async () => {
    try {
      const msg = await getCurrentUserWithPermissions({
        skipErrorHandler: true,
      });
      // 检查响应状态，只有成功时才返回数据
      if (msg.code === 0 && msg.data) {
        return msg.data;
      } else {
        // API返回错误状态码，跳转到登录页
        history.push(loginPath);
        return undefined;
      }
    } catch (_error) {
      history.push(loginPath);
    }
    return undefined;
  };

  // 从localStorage读取保存的主题配置
  let mergedSettings = { ...defaultSettings } as Partial<LayoutSettings>;
  try {
    const savedSettings = localStorage.getItem(SETTINGS_STORAGE_KEY);
    if (savedSettings) {
      const parsed = JSON.parse(savedSettings);
      mergedSettings = { ...mergedSettings, ...parsed };
    }
  } catch (error) {
    console.error("Failed to load settings from localStorage:", error);
  }

  // 从localStorage读取主题模式
  let themeMode: ThemeMode = "light";
  try {
    const savedThemeMode = localStorage.getItem(THEME_MODE_STORAGE_KEY);
    if (
      savedThemeMode &&
      ["light", "dark", "compact", "dark-compact"].includes(savedThemeMode)
    ) {
      themeMode = savedThemeMode as ThemeMode;
    }
  } catch (error) {
    console.error("Failed to load theme mode from localStorage:", error);
  }

  // 如果不是登录页面，执行
  const { location } = history;
  if (
    ![loginPath, "/user/register", "/user/register-result"].includes(
      location.pathname
    )
  ) {
    const currentUser = await fetchUserInfo();
    return {
      fetchUserInfo,
      currentUser,
      settings: mergedSettings,
      themeMode,
    };
  }
  return {
    fetchUserInfo,
    settings: mergedSettings,
    themeMode,
  };
}

// ProLayout 支持的api https://procomponents.ant.design/components/layout
export const layout: RunTimeLayoutConfig = ({
  initialState,
  setInitialState,
}) => {
  // 统一的设置更新处理函数（同时更新状态和localStorage）
  const updateSettings = (newSettings: Partial<LayoutSettings>) => {
    setInitialState((s) => ({
      ...s,
      settings: { ...s?.settings, ...newSettings },
    }));
    // 保存到localStorage
    try {
      localStorage.setItem(
        SETTINGS_STORAGE_KEY,
        JSON.stringify({ ...initialState?.settings, ...newSettings })
      );
    } catch (error) {
      console.error("Failed to save settings to localStorage:", error);
    }
  };

  // 主题模式切换处理函数
  const handleThemeModeChange = (mode: ThemeMode) => {
    setInitialState((s) => ({
      ...s,
      themeMode: mode,
    }));
    // 保存到localStorage
    try {
      localStorage.setItem(THEME_MODE_STORAGE_KEY, mode);
    } catch (error) {
      console.error("Failed to save theme mode to localStorage:", error);
    }

    // 同步更新navTheme以保持ProLayout的暗色模式
    const isDark = mode === "dark" || mode === "dark-compact";
    updateSettings({ navTheme: isDark ? "realDark" : "light" });
  };

  // 获取当前主题模式
  const currentThemeMode = initialState?.themeMode || "light";

  return {
    layout: "mix",
    splitMenus: true,
    appList,
    actionsRender: () => [
      <Question key="doc" />,
      <SelectLang key="SelectLang" />,
      <ThemeSwitcher
        key="ThemeSwitcher"
        mode={currentThemeMode}
        onChange={handleThemeModeChange}
      />,
    ],
    avatarProps: {
      src: initialState?.currentUser?.avatar || undefined,
      title: <AvatarName />,
      // 当没有头像时，显示用户名首字母
      children: !initialState?.currentUser?.avatar
        ? (initialState?.currentUser?.userName?.charAt(0).toUpperCase() || 'U')
        : undefined,
      render: (_, avatarChildren) => {
        return <AvatarDropdown menu>{avatarChildren}</AvatarDropdown>;
      },
    },
    waterMarkProps: {
      content: initialState?.currentUser?.name,
    },
    footerRender: () => {
      const { pathname } = history.location;
      if (pathname.startsWith("/visualization-dashboard")) {
        return null;
      }
      return <Footer />;
    },
    onPageChange: () => {
      const { location } = history;
      // 如果没有登录，重定向到 login
      if (!initialState?.currentUser && location.pathname !== loginPath) {
        history.push(loginPath);
      }
    },
    bgLayoutImgList: [
      {
        src: "https://mdn.alipayobjects.com/yuyan_qk0oxh/afts/img/D2LWSqNny4sAAAAAAAAAAAAAFl94AQBr",
        left: 85,
        bottom: 100,
        height: "303px",
      },
      {
        src: "https://mdn.alipayobjects.com/yuyan_qk0oxh/afts/img/C2TWRpJpiC0AAAAAAAAAAAAAFl94AQBr",
        bottom: -68,
        right: -45,
        height: "303px",
      },
      {
        src: "https://mdn.alipayobjects.com/yuyan_qk0oxh/afts/img/F6vSTbj8KpYAAAAAAAAAAAAAFl94AQBr",
        bottom: 0,
        left: 0,
        width: "331px",
      },
    ],
    links: isDev
      ? [
          <Link key="openapi" to="/umi/plugin/openapi" target="_blank">
            <LinkOutlined />
            <span>OpenAPI 文档</span>
          </Link>,
        ]
      : [],
    menuHeaderRender: undefined,
    // 自定义 403 页面
    // unAccessible: <div>unAccessible</div>,
    // 增加一个 loading 的状态
    childrenRender: (children) => {
      // if (initialState?.loading) return <PageLoading />;
      return (
        <ConfigProvider
          theme={{
            algorithm: getAlgorithm(currentThemeMode),
          }}
        >
          <App message={{ top: 80 }}>
            {children}
            {isDev && (
              <SettingDrawer
                disableUrlParams
                enableDarkTheme
                settings={initialState?.settings}
                onSettingChange={(newSettings) => {
                  // 使用统一的更新函数，确保同步到localStorage
                  updateSettings(newSettings);
                }}
              />
            )}
          </App>
        </ConfigProvider>
      );
    },
    ...initialState?.settings,
    // 强制设置导航栏断点，确保在 < 1200px 时自动收起
    breakpoint: 'xl',
  };
};

/**
 * @name request 配置，可以配置错误处理
 * 它基于 axios 和 ahooks 的 useRequest 提供了一套统一的网络请求和错误处理方案。
 * @doc https://umijs.org/docs/max/request#配置
 */
export const request: RequestConfig = {
  ...errorConfig,
};
