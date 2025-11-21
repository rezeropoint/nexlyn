/**
 * 用户选择组件
 * 支持单选和多选模式，集成搜索功能
 * 复用系统的 getUserOptions API
 */
import { getUserOptions } from "@/services/user";
import { ProFormSelect } from "@ant-design/pro-components";
import type { ProFormSelectProps } from "@ant-design/pro-components";
import React from "react";

interface UserSelectProps extends Omit<ProFormSelectProps, "request"> {
  mode?: "multiple" | undefined; // 支持单选和多选
  showEmail?: boolean; // 是否在label中显示邮箱
}

/**
 * 用户选择组件
 */
const UserSelect: React.FC<UserSelectProps> = ({
  mode,
  showEmail = true,
  ...restProps
}) => {
  return (
    <ProFormSelect
      mode={mode}
      showSearch
      request={async (params) => {
        try {
          const response = await getUserOptions({
            keyword: params.keyWords,
            limit: 50,
          });

          if (response.code === 0 && response.data?.list) {
            return response.data.list.map((user) => ({
              label: showEmail
                ? `${user.name} (${user.userName}) - ${user.email}`
                : `${user.name} (${user.userName})`,
              value: user.userKey,
            }));
          }
          return [];
        } catch (error) {
          console.error("获取用户选项失败:", error);
          return [];
        }
      }}
      debounceTime={300}
      fieldProps={{
        filterOption: false,
      }}
      {...restProps}
    />
  );
};

export default UserSelect;
