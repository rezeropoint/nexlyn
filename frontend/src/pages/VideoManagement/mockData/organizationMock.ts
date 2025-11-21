// 组织架构模拟数据
// 临时数据，后续集成真实API时可删除此文件

export interface MockOrganization {
  id: string;
  name: string;
  parentId: string | null;
  level: number;
  children?: MockOrganization[];
}

// 模拟的组织架构数据
export const mockOrganizations: MockOrganization[] = [
  {
    id: "org-hq",
    name: "总部",
    parentId: null,
    level: 1,
    children: [
      {
        id: "org-guizhou",
        name: "贵州分公司",
        parentId: "org-hq",
        level: 2,
        children: [
          {
            id: "org-guiyang-support",
            name: "贵阳支撑中心",
            parentId: "org-guizhou",
            level: 3,
          },
        ],
      },
      {
        id: "org-shandong",
        name: "山东分公司",
        parentId: "org-hq",
        level: 2,
        children: [
          {
            id: "org-jinan-support",
            name: "济南支撑中心",
            parentId: "org-shandong",
            level: 3,
          },
        ],
      },
      {
        id: "org-guangdong",
        name: "广东分公司",
        parentId: "org-hq",
        level: 2,
        children: [
          {
            id: "org-guangzhou-support",
            name: "广州支撑中心",
            parentId: "org-guangdong",
            level: 3,
          },
        ],
      },
      {
        id: "org-sichuan",
        name: "四川工程",
        parentId: "org-hq",
        level: 2,
        children: [
          {
            id: "org-sichuan-eng1",
            name: "工程一中心",
            parentId: "org-sichuan",
            level: 3,
          },
          {
            id: "org-sichuan-eng2",
            name: "工程二中心",
            parentId: "org-sichuan",
            level: 3,
          },
          {
            id: "org-chengdu-net",
            name: "成都网络维护中心",
            parentId: "org-sichuan",
            level: 3,
            children: [
              {
                id: "org-wenjiang-support",
                name: "温江支撑中心",
                parentId: "org-chengdu-net",
                level: 4,
              },
            ],
          },
        ],
      },
    ],
  },
];

// 扁平化组织数据，用于快速查找
export const flatOrganizations: MockOrganization[] = [
  { id: "org-hq", name: "总部", parentId: null, level: 1 },
  { id: "org-guizhou", name: "贵州分公司", parentId: "org-hq", level: 2 },
  {
    id: "org-guiyang-support",
    name: "贵阳支撑中心",
    parentId: "org-guizhou",
    level: 3,
  },
  { id: "org-shandong", name: "山东分公司", parentId: "org-hq", level: 2 },
  {
    id: "org-jinan-support",
    name: "济南支撑中心",
    parentId: "org-shandong",
    level: 3,
  },
  { id: "org-guangdong", name: "广东分公司", parentId: "org-hq", level: 2 },
  {
    id: "org-guangzhou-support",
    name: "广州支撑中心",
    parentId: "org-guangdong",
    level: 3,
  },
  { id: "org-sichuan", name: "四川工程", parentId: "org-hq", level: 2 },
  {
    id: "org-sichuan-eng1",
    name: "工程一中心",
    parentId: "org-sichuan",
    level: 3,
  },
  {
    id: "org-sichuan-eng2",
    name: "工程二中心",
    parentId: "org-sichuan",
    level: 3,
  },
  {
    id: "org-chengdu-net",
    name: "成都网络维护中心",
    parentId: "org-sichuan",
    level: 3,
  },
  {
    id: "org-wenjiang-support",
    name: "温江支撑中心",
    parentId: "org-chengdu-net",
    level: 4,
  },
];

// 根据ID获取组织名称
export const getOrganizationName = (id: string): string => {
  const org = flatOrganizations.find((o) => o.id === id);
  return org?.name || "未知组织";
};

// 获取组织路径（用于显示完整路径）
export const getOrganizationPath = (id: string): string => {
  const org = flatOrganizations.find((o) => o.id === id);
  if (!org) return "";

  const path: string[] = [org.name];
  let current = org;

  while (current.parentId) {
    const parent = flatOrganizations.find((o) => o.id === current.parentId);
    if (parent) {
      path.unshift(parent.name);
      current = parent;
    } else {
      break;
    }
  }

  return path.join(" / ");
};
