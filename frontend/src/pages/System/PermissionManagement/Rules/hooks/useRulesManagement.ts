import {
  addPermission,
  checkPermission,
  removePermission,
} from "@/services/permission";
import { message } from "antd";
import { useCallback, useState } from "react";

export const useRulesManagement = () => {
  const [loading, setLoading] = useState(false);

  // 添加权限规则
  const handleAddRule = useCallback(
    async (values: API.AddPermissionRequest): Promise<boolean> => {
      try {
        setLoading(true);
        const res = await addPermission(values);

        if (res.code === 0) {
          message.success(res.msg || "添加权限规则成功");
          return true;
        } else {
          message.error(res.msg || "添加失败");
          return false;
        }
      } catch (error) {
        console.error("添加权限规则失败:", error);
        message.error("添加权限规则失败");
        return false;
      } finally {
        setLoading(false);
      }
    },
    []
  );

  // 删除权限规则
  const handleDeleteRule = useCallback(
    async (record: API.PermissionRule): Promise<void> => {
      try {
        setLoading(true);
        const res = await removePermission({
          userKey: record.userKey,
          tenantKey: record.tenantKey,
          resource: record.resource,
          action: record.action,
        });

        if (res.code === 0) {
          message.success(res.msg || "删除权限规则成功");
        } else {
          message.error(res.msg || "删除失败");
        }
      } catch (error) {
        console.error("删除权限规则失败:", error);
        message.error("删除权限规则失败");
      } finally {
        setLoading(false);
      }
    },
    []
  );

  // 检查权限
  const handleCheckPermission = useCallback(
    async (record: API.PermissionRule): Promise<void> => {
      try {
        setLoading(true);
        const res = await checkPermission({
          userKey: record.userKey,
          tenantKey: record.tenantKey,
          resource: record.resource,
          action: record.action,
        });

        if (res.code === 0) {
          message.success(res.hasPermission ? "✅ 有权限" : "❌ 无权限");
        } else {
          message.error(res.msg || "检查失败");
        }
      } catch (error) {
        console.error("检查权限失败:", error);
        message.error("检查权限失败");
      } finally {
        setLoading(false);
      }
    },
    []
  );

  return {
    loading,
    handleAddRule,
    handleDeleteRule,
    handleCheckPermission,
  };
};
