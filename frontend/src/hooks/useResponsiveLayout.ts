import { Grid } from "antd";
import { useEffect, useState } from "react";

const { useBreakpoint } = Grid;

export type LayoutDirection = "horizontal" | "vertical";

export interface UseResponsiveLayoutOptions {
  /**
   * 断点，当屏幕宽度小于等于此断点时切换为垂直布局
   * @default 'md' (中等屏幕，通常是 768px)
   */
  breakpoint?: "xs" | "sm" | "md" | "lg" | "xl" | "xxl";

  /**
   * 默认布局方向
   * @default 'horizontal'
   */
  defaultLayout?: LayoutDirection;

  /**
   * 是否启用响应式布局切换
   * @default true
   */
  responsive?: boolean;
}

/**
 * 响应式布局 Hook
 *
 * 根据屏幕尺寸自动切换布局方向（水平/垂直）
 * 适用于 Splitter 等需要响应式布局的组件
 *
 * @example
 * ```tsx
 * const layout = useResponsiveLayout({ breakpoint: 'md' });
 *
 * <Splitter layout={layout}>
 *   <Splitter.Panel>左侧</Splitter.Panel>
 *   <Splitter.Panel>右侧</Splitter.Panel>
 * </Splitter>
 * ```
 */
export const useResponsiveLayout = (
  options: UseResponsiveLayoutOptions = {}
): LayoutDirection => {
  const {
    breakpoint = "md",
    defaultLayout = "horizontal",
    responsive = true,
  } = options;

  const screens = useBreakpoint();
  const [layout, setLayout] = useState<LayoutDirection>(defaultLayout);

  useEffect(() => {
    if (!responsive) {
      setLayout(defaultLayout);
      return;
    }

    // 断点映射：xs < sm < md < lg < xl < xxl
    // 如果当前屏幕小于等于断点，则使用垂直布局
    const breakpointOrder = ["xs", "sm", "md", "lg", "xl", "xxl"];
    const breakpointIndex = breakpointOrder.indexOf(breakpoint);

    // 检查当前屏幕是否小于等于断点
    let isSmallScreen = false;

    for (let i = 0; i <= breakpointIndex; i++) {
      const bp = breakpointOrder[i] as keyof typeof screens;
      if (screens[bp]) {
        isSmallScreen = true;
        // 检查是否有更大的断点激活
        for (let j = i + 1; j < breakpointOrder.length; j++) {
          const largerBp = breakpointOrder[j] as keyof typeof screens;
          if (screens[largerBp]) {
            isSmallScreen = false;
            break;
          }
        }
        break;
      }
    }

    // 小屏幕使用垂直布局，大屏幕使用水平布局
    setLayout(isSmallScreen ? "vertical" : "horizontal");
  }, [screens, breakpoint, defaultLayout, responsive]);

  return layout;
};

export default useResponsiveLayout;
