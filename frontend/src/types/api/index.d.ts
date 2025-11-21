// API类型定义统一入口文件
// 这个文件用于兼容性过渡，确保所有模块化的类型定义都能被正确引入

/// <reference path="./common.d.ts" />
/// <reference path="./user.d.ts" />
/// <reference path="./tenant.d.ts" />
/// <reference path="./permission.d.ts" />
/// <reference path="./tag.d.ts" />
/// <reference path="./infoatom.d.ts" />
/// <reference path="./video.d.ts" />
/// <reference path="./organization.d.ts" />

// 保留一些兼容性类型，确保现有代码继续工作
declare namespace API {
  // ===== 保留的旧类型 (模板相关，逐步清理) =====

  type RuleListItem = {
    key?: number;
    disabled?: boolean;
    href?: string;
    avatar?: string;
    name?: string;
    owner?: string;
    desc?: string;
    callNo?: number;
    status?: number;
    updatedAt?: string;
    createdAt?: string;
    progress?: number;
  };

  type RuleList = {
    data?: RuleListItem[];
    /** 列表的内容总数 */
    total?: number;
    success?: boolean;
  };

  type FakeCaptcha = {
    code?: number;
    status?: string;
  };

  type NoticeIconList = {
    data?: NoticeIconItem[];
    /** 列表的内容总数 */
    total?: number;
    success?: boolean;
  };

  type NoticeIconItemType = "notification" | "message" | "event";

  type NoticeIconItem = {
    id?: string;
    extra?: string;
    key?: string;
    read?: boolean;
    avatar?: string;
    title?: string;
    status?: string;
    datetime?: string;
    description?: string;
    type?: NoticeIconItemType;
  };
}
