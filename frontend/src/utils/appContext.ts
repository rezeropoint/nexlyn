import { App } from "antd";

/**
 * 使用 App 组件的 hooks
 * 用于替代静态方法调用 message.xxx、Modal.xxx、notification.xxx
 *
 * @example
 * ```tsx
 * import { useApp } from '@/utils/appContext';
 *
 * const MyComponent = () => {
 *   const { message, modal, notification } = useApp();
 *
 *   const handleClick = () => {
 *     message.success('操作成功');
 *     modal.confirm({ title: '确认操作?' });
 *   };
 * };
 * ```
 */
export const useApp = App.useApp;
