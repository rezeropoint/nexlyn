/**
 * @name umi 的路由配置
 * @description 只支持 path,component,routes,redirect,wrappers,name,icon 的配置
 * @param path  path 只支持两种占位符配置，第一种是动态参数 :id 的形式，第二种是 * 通配符，通配符只能出现路由字符串的最后。
 * @param component 配置 location 和 path 匹配后用于渲染的 React 组件路径。可以是绝对路径，也可以是相对路径，如果是相对路径，会从 src/pages 开始找起。
 * @param routes 配置子路由，通常在需要为多个路径增加 layout 组件时使用。
 * @param redirect 配置路由跳转
 * @param wrappers 配置路由组件的包装组件，通过包装组件可以为当前的路由组件组合进更多的功能。 比如，可以用于路由级别的权限校验
 * @param name 配置路由的标题，默认读取国际化文件 menu.ts 中 menu.xxxx 的值，如配置 name 为 login，则读取 menu.ts 中 menu.login 的取值作为标题
 * @param icon 配置路由的图标，取值参考 https://ant.design/components/icon-cn， 注意去除风格后缀和大小写，如想要配置图标为 <StepBackwardOutlined /> 则取值应为 stepBackward 或 StepBackward，如想要配置图标为 <UserOutlined /> 则取值应为 user 或者 User
 * @doc https://umijs.org/docs/guides/routes
 */
export default [
  {
    path: '/user',
    layout: false,
    routes: [
      {
        name: 'login',
        path: '/user/login',
        component: './System/Login',
      },
    ],
  },
  {
    path: '/workbench',
    name: '工作台',
    icon: 'HomeOutlined',
    routes: [
      {
        path: '/workbench',
        redirect: '/workbench/dashboard',
      },
      {
        name: '仪表盘',
        icon: 'DashboardOutlined',
        path: '/workbench/dashboard',
        component: './Workbench/Dashboard',
      },
      {
        name: '待办任务',
        icon: 'ClockCircleOutlined',
        path: '/workbench/pending-tasks',
        component: './Workbench/PendingTasks',
      },
      {
        name: '我发起的',
        icon: 'SendOutlined',
        path: '/workbench/proposed-journeys',
        component: './Workbench/ProposedJourneys',
      },
      {
        name: '抄送消息',
        icon: 'MailOutlined',
        path: '/workbench/cc-messages',
        component: './Workbench/CCMessages',
      },
    ],
  },
  {
    name: '可视化大屏',
    icon: 'DashboardOutlined',
    path: '/visualization-dashboard',
    component: './VisualizationDashboard',
  },
  {
    name: '视频管理',
    icon: 'VideoCameraOutlined',
    path: '/video-management',
    access: 'canAccessVideoModule',
    routes: [
      {
        path: '/video-management',
        redirect: '/video-management/device',
      },
      {
        name: '设备管理',
        icon: 'VideoCameraOutlined',
        path: '/video-management/device',
        component: './VideoManagement/Device',
        access: 'canAccessGB28181Device',
      },
      {
        name: '分屏监控',
        icon: 'AppstoreOutlined',
        path: '/video-management/split-screen',
        component: './VideoManagement/SplitScreen',
        access: 'canAccessGB28181SplitScreen',
      },
      {
        name: '标签管理',
        icon: 'TagsOutlined',
        path: '/video-management/device-tag-management',
        component: './VideoManagement/DeviceTagManagement',
        access: 'canManageDeviceTags', // 只有管理员可以访问
      },
      {
        name: '媒体管理',
        icon: 'PlayCircleOutlined',
        path: '/video-management/media',
        component: './VideoManagement/Media',
        access: 'canAccessMediaManagement', // 需要gb28181_stream:read权限
      },
      {
        name: '录像管理',
        icon: 'PlaySquareOutlined',
        path: '/video-management/recording',
        component: './VideoManagement/RecordingManagement',
        access: 'canAccessGB28181Recording',
      },
    ],
  },
  {
    name: '算法管理',
    icon: 'ExperimentOutlined',
    path: '/algorithm-management',
    routes: [
      {
        path: '/algorithm-management',
        redirect: '/algorithm-management/edge-algorithm',
      },
      {
        name: '边缘算法',
        icon: 'NodeIndexOutlined',
        path: '/algorithm-management/edge-algorithm',
        routes: [
          {
            path: '/algorithm-management/edge-algorithm',
            redirect: '/algorithm-management/edge-algorithm/task-management',
          },
          {
            name: '任务管理',
            icon: 'ExperimentOutlined',
            path: '/algorithm-management/edge-algorithm/task-management',
            component: './AlgorithmManagement/EdgeAlgorithm/TaskManagement',
          },
          {
            name: '媒体通道',
            icon: 'PlayCircleOutlined',
            path: '/algorithm-management/edge-algorithm/media-channel',
            component: './AlgorithmManagement/EdgeAlgorithm/MediaChannel',
          },
          {
            name: '系统设置',
            icon: 'ControlOutlined',
            path: '/algorithm-management/edge-algorithm/system-settings',
            component: './AlgorithmManagement/EdgeAlgorithm/SystemSettings',
          },
        ],
      },
      {
        name: '云端算法',
        icon: 'CloudServerOutlined',
        path: '/algorithm-management/cloud-algorithm',
        component: './AlgorithmManagement/CloudAlgorithm',
      },
      {
        name: '端侧算法',
        icon: 'MobileOutlined',
        path: '/algorithm-management/device-algorithm',
        component: './AlgorithmManagement/DeviceAlgorithm',
      },
    ],
  },
  {
    name: '物联管理',
    icon: 'ApiOutlined',
    path: '/iot-management',
    access: 'canAccessIoTModule',
    routes: [
      {
        path: '/iot-management',
        redirect: '/iot-management/device-management',
      },
      {
        name: '设备管理',
        icon: 'ApiOutlined',
        path: '/iot-management/device-management',
        component: './IoTManagement/DeviceManagement',
        access: 'canAccessIoTDevice',
      },
      {
        name: '设备模板',
        icon: 'ProfileOutlined',
        path: '/iot-management/sensor-template',
        component: './IoTManagement/SensorTemplate',
        access: 'canAccessIoTTemplate',
      },
      {
        name: '设备标签',
        icon: 'TagsOutlined',
        path: '/iot-management/device-tag',
        component: './IoTManagement/DeviceTag',
        access: 'canAccessIoTTag',
      },
      {
        name: '平台管理',
        icon: 'CloudServerOutlined',
        path: '/iot-management/platform-management',
        component: './IoTManagement/PlatformManagement',
        access: 'canAccessIoTPlatform',
      },
      {
        name: 'HTTP 接收',
        icon: 'CloudDownloadOutlined',
        path: '/iot-management/http-receive',
        component: './IoTManagement/HttpReceive',
        access: 'canAccessIoTHttpReceive',
      },
    ],
  },
  {
    name: '逻辑引擎',
    icon: 'ApartmentOutlined',
    path: '/logic-engine',
    access: 'canAccessLynxModule',
    routes: [
      {
        path: '/logic-engine',
        redirect: '/logic-engine/overview',
      },
      {
        name: '引擎概览',
        icon: 'DashboardOutlined',
        path: '/logic-engine/overview',
        component: './LogicEngine',
        access: 'canAccessLynxOverview',
      },
      {
        name: '标签管理',
        icon: 'TagsOutlined',
        path: '/logic-engine/tag-management',
        component: './LogicEngine/TagManagement',
        access: 'canAccessLynxTag',
      },
      {
        name: '信息原子',
        icon: 'NodeIndexOutlined',
        path: '/logic-engine/infoatom-type',
        component: './LogicEngine/InfoAtomType',
        access: 'canAccessLynxInfoAtomType',
      },
      {
        name: '逻辑图配置',
        icon: 'PartitionOutlined',
        path: '/logic-engine/graph-config',
        component: './LogicEngine/GraphConfig',
        access: 'canAccessLynxGraphConfig',
      },
      {
        name: '逻辑图详情',
        path: '/logic-engine/graph-config/:id',
        component: './LogicEngine/GraphConfig/Detail',
        hideInMenu: true,
        access: 'canAccessLynxGraphConfig',
      },
    ],
  },
  {
    name: '事件管理',
    icon: 'DeploymentUnitOutlined',
    path: '/event-management',
    access: 'canAccessEventModule',
    routes: [
      {
        path: '/event-management',
        redirect: '/event-management/event-data',
      },
      {
        name: '事件查询',
        icon: 'SearchOutlined',
        path: '/event-management/event-data',
        component: './EventManagement/EventData',
        access: 'canAccessEventData',
      },
      {
        name: '事件配置',
        icon: 'ControlOutlined',
        path: '/event-management/event-config',
        component: './EventManagement/EventConfig',
        access: 'canAccessEventConfig',
      },
      {
        name: '组织映射',
        icon: 'ApartmentOutlined',
        path: '/event-management/org-mapping',
        component: './EventManagement/OrgMapping',
        access: 'canAccessOrgMapping',
      },
      {
        name: '平台配置',
        icon: 'CloudServerOutlined',
        path: '/event-management/platform-config',
        component: './EventManagement/PlatformConfig',
        access: 'canAccessSkylarkPlatform',
      },
    ],
  },
  {
    name: 'system',
    icon: 'setting',
    path: '/system',
    access: 'canAccessSystemModule',
    routes: [
      {
        path: '/system',
        redirect: '/system/user-management',
      },
      {
        name: '用户管理',
        icon: 'team',
        path: '/system/user-management',
        component: './System/UserManagement',
        access: 'canManageUsers', // 只有有用户管理权限的用户可以访问
      },
      {
        name: '权限管理',
        icon: 'key',
        path: '/system/permission-management',
        access: 'canManagePermissions', // 只有有权限管理权限的用户可以访问
        routes: [
          {
            path: '/system/permission-management',
            redirect: '/system/permission-management/roles',
          },
          {
            name: '用户角色',
            icon: 'UserOutlined',
            path: '/system/permission-management/roles',
            component: './System/PermissionManagement/Roles',
          },
          {
            name: '角色权限',
            icon: 'SettingOutlined',
            path: '/system/permission-management/role-permissions',
            component: './System/PermissionManagement/RolePermissions',
          },
          {
            name: '权限规则',
            icon: 'SecurityScanOutlined',
            path: '/system/permission-management/rules',
            component: './System/PermissionManagement/Rules',
          },
          {
            name: '权限检查',
            icon: 'CheckCircleOutlined',
            path: '/system/permission-management/checker',
            component: './System/PermissionManagement/Checker',
          },
        ],
      },
      {
        name: '租户管理',
        icon: 'BankOutlined',
        path: '/system/tenant-management',
        component: './System/TenantManagement',
        access: 'canManageTenants', // 只有有租户管理权限的用户可以访问
      },
      {
        name: '标签管理',
        icon: 'tag',
        path: '/system/tag-management',
        component: './System/TagManagement',
        access: 'canManageTags', // 只有有标签管理权限的用户可以访问
      },
      {
        name: '组织管理',
        icon: 'apartment',
        path: '/system/organization-management',
        component: './System/OrganizationManagement',
        access: 'canManageOrganizations', // 只有有组织管理权限的用户可以访问
      },
    ],
  },
  {
    path: '/account',
    name: '个人中心',
    hideInMenu: true,
    routes: [
      {
        name: '个人中心',
        path: '/account/center',
        component: './Account/Center',
      },
    ],
  },
  {
    path: '/',
    redirect: '/workbench',
  },
  {
    path: '*',
    layout: false,
    component: './404',
  },
];
