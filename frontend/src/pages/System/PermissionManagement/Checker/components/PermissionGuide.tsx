import { InfoCircleOutlined } from "@ant-design/icons";
import { Alert, Card, Divider, Space, Table, Tag, Typography } from "antd";
import React from "react";

const { Title, Paragraph, Text } = Typography;

/**
 * 权限说明组件
 */
const PermissionGuide: React.FC = () => {
  // 权限操作说明
  const permissionActions = [
    {
      key: "read",
      action: "read",
      name: "读取",
      description: "允许查看和获取资源信息",
      examples: "查看用户列表、获取租户信息、查看图配置等",
    },
    {
      key: "write",
      action: "write",
      name: "写入",
      description: "允许创建和修改资源",
      examples: "创建用户、更新租户信息、修改图配置等",
    },
    {
      key: "delete",
      action: "delete",
      name: "删除",
      description: "允许删除资源",
      examples: "删除用户、删除租户、删除图配置等",
    },
    {
      key: "admin",
      action: "admin",
      name: "管理",
      description: "拥有该资源的所有权限（包含读取、写入、删除）",
      examples: "完全管理用户、完全管理租户等",
    },
  ];

  // 资源类型说明
  const resourceTypes = [
    {
      key: "user",
      resource: "user",
      name: "用户管理",
      description: "用户相关的所有操作权限",
      subResources: ["user:read", "user:write", "user:delete"],
    },
    {
      key: "tenant",
      resource: "tenant",
      name: "租户管理",
      description: "租户相关的所有操作权限",
      subResources: ["tenant:read", "tenant:write", "tenant:delete"],
    },
    {
      key: "graph",
      resource: "graph",
      name: "图配置管理",
      description: "图配置相关的所有操作权限",
      subResources: ["graph:read", "graph:write", "graph:delete"],
    },
    {
      key: "infoatom",
      resource: "infoatom",
      name: "信息原子管理",
      description: "信息原子类型相关的操作权限",
      subResources: ["infoatom:read", "infoatom:write", "infoatom:delete"],
    },
    {
      key: "tags",
      resource: "tags",
      name: "标签管理",
      description: "标签相关的操作权限",
      subResources: ["tags:read", "tags:write", "tags:delete"],
    },
    {
      key: "permission",
      resource: "permission",
      name: "权限管理",
      description: "权限和角色相关的操作权限",
      subResources: [
        "permission:read",
        "permission:write",
        "permission:delete",
      ],
    },
  ];

  const actionColumns = [
    {
      title: "操作",
      dataIndex: "action",
      key: "action",
      render: (action: string) => <Tag color="blue">{action}</Tag>,
    },
    {
      title: "名称",
      dataIndex: "name",
      key: "name",
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: "描述",
      dataIndex: "description",
      key: "description",
    },
    {
      title: "示例",
      dataIndex: "examples",
      key: "examples",
    },
  ];

  const resourceColumns = [
    {
      title: "资源",
      dataIndex: "resource",
      key: "resource",
      render: (resource: string) => <Tag color="green">{resource}</Tag>,
    },
    {
      title: "名称",
      dataIndex: "name",
      key: "name",
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: "描述",
      dataIndex: "description",
      key: "description",
    },
    {
      title: "子资源",
      dataIndex: "subResources",
      key: "subResources",
      render: (subResources: string[]) => (
        <Space wrap>
          {subResources.map((sub) => (
            <Tag key={sub} color="orange">
              {sub}
            </Tag>
          ))}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Alert
        message="CasbinX 权限系统说明 v0.7.11+"
        description="Nexlyn 使用基于 CasbinX 的安全权限控制模型。新版本引入了权限分级保护机制，修复了提权漏洞，添加了防自我提权、系统权限保护等安全特性。用户权限由角色权限和直接权限组成，所有操作都受到安全验证保护。"
        type="info"
        icon={<InfoCircleOutlined />}
        style={{ marginBottom: 24 }}
      />

      <Card title="权限分级保护机制" style={{ marginBottom: 24 }}>
        <Typography>
          <Paragraph>
            CasbinX v0.7.11+ 引入了重要的权限分级保护机制，将权限分为三个级别：
          </Paragraph>
          <ul>
            <li>
              <Text strong>系统权限 (System)：</Text>
              完全不可变更的核心系统权限，如租户管理、权限管理等
            </li>
            <li>
              <Text strong>管理权限 (Management)：</Text>
              有传递深度限制的管理类权限，防止权限无限扩散
            </li>
            <li>
              <Text strong>普通权限 (Normal)：</Text>
              可自由授予和撤销的日常操作权限
            </li>
          </ul>
        </Typography>
      </Card>

      <Card title="权限组成" style={{ marginBottom: 24 }}>
        <Typography>
          <Paragraph>权限由以下四个要素组成：</Paragraph>
          <ul>
            <li>
              <Text strong>用户 (User)：</Text>权限的主体，即谁拥有这个权限
            </li>
            <li>
              <Text strong>租户 (Tenant)：</Text>
              权限的作用域，即在哪个租户下有效
            </li>
            <li>
              <Text strong>资源 (Resource)：</Text>
              权限作用的对象，如用户、租户、图配置等
            </li>
            <li>
              <Text strong>操作 (Action)：</Text>
              对资源允许执行的操作，如读取、写入、删除等
            </li>
          </ul>
          <Paragraph>
            <Text strong>权限格式：</Text>{" "}
            <Text code>[用户] 在 [租户] 下对 [资源] 具有 [操作] 权限</Text>
          </Paragraph>
        </Typography>
      </Card>

      <Card title="操作类型说明" style={{ marginBottom: 24 }}>
        <Table
          dataSource={permissionActions}
          columns={actionColumns}
          pagination={false}
          size="middle"
        />
      </Card>

      <Card title="资源类型说明" style={{ marginBottom: 24 }}>
        <Table
          dataSource={resourceTypes}
          columns={resourceColumns}
          pagination={false}
          size="middle"
        />
      </Card>

      <Card title="核心安全特性" style={{ marginBottom: 24 }}>
        <Typography>
          <Title level={4}>防护机制</Title>
          <ul>
            <li>
              <Text strong>防止自我提权：</Text>
              默认启用，防止用户给自己分配管理权限，避免权限滥用
            </li>
            <li>
              <Text strong>系统权限保护：</Text>
              系统级权限（如租户管理、权限管理）完全不可授予也不可撤销
            </li>
            <li>
              <Text strong>角色系统权限保护：</Text>
              角色中的系统权限在角色更新时自动保留，防止意外移除关键权限
            </li>
            <li>
              <Text strong>操作者权限验证：</Text>
              所有权限管理操作都会验证操作者是否有执行权限
            </li>
            <li>
              <Text strong>跨域权限继承修复：</Text>
              正确处理角色权限在全局域、用户-角色关系在特定域的场景
            </li>
          </ul>

          <Title level={4}>新API接口要求</Title>
          <Paragraph>
            <Text type="warning">重要变更：</Text>所有权限管理操作现在都需要提供{" "}
            <Text code>operatorKey</Text> 参数，用于验证操作者权限。
          </Paragraph>
          <ul>
            <li>
              用户权限管理：
              <Text code>
                GrantPermission(operatorKey, userKey, tenantKey, permission)
              </Text>
            </li>
            <li>
              角色分配：
              <Text code>
                AssignRole(operatorKey, userKey, roleKey, tenantKey)
              </Text>
            </li>
            <li>
              安全权限查询：
              <Text code>
                GetDirectPermissionsSecure(operatorKey, userKey, tenantKey)
              </Text>
            </li>
          </ul>
        </Typography>
      </Card>

      <Card title="权限检查规则">
        <Typography>
          <Title level={4}>权限来源</Title>
          <Paragraph>用户权限可能来自以下两个来源：</Paragraph>
          <ol>
            <li>
              <Text strong>角色权限：</Text>通过分配角色继承的权限规则
            </li>
            <li>
              <Text strong>直接权限：</Text>通过"权限规则"页面直接分配的权限规则
            </li>
          </ol>
          <Paragraph>
            <Text strong>最终权限 = 角色权限 ∪ 直接权限</Text>（取并集）
          </Paragraph>

          <Divider />

          <Title level={4}>权限管理页面</Title>
          <ul>
            <li>
              <Text strong>角色权限管理：</Text>
              统一配置所有角色的权限，影响所有拥有该角色的用户
            </li>
            <li>
              <Text strong>用户角色分配：</Text>
              为用户分配或移除角色，用户继承角色权限
            </li>
            <li>
              <Text strong>权限规则管理：</Text>
              为特定用户添加补充权限（角色之外的额外权限）
            </li>
            <li>
              <Text strong>权限检查工具：</Text>查看用户的最终权限组合和来源分析
            </li>
          </ul>

          <Divider />

          <Title level={4}>权限继承</Title>
          <Paragraph>资源权限具有继承关系：</Paragraph>
          <ul>
            <li>
              <Text code>user:admin</Text> 包含 <Text code>user:read</Text>,{" "}
              <Text code>user:write</Text>, <Text code>user:delete</Text>
            </li>
            <li>
              <Text code>graph:admin</Text> 包含 <Text code>graph:read</Text>,{" "}
              <Text code>graph:write</Text>, <Text code>graph:delete</Text>
            </li>
            <li>
              <Text code>permission:admin</Text> 包含所有权限管理相关操作
            </li>
          </ul>

          <Divider />

          <Title level={4}>安全检查逻辑</Title>
          <ol>
            <li>
              <Text strong>权限分级验证：</Text>根据权限类型应用不同的安全策略
            </li>
            <li>
              <Text strong>操作者身份验证：</Text>
              所有权限操作都需要验证操作者权限
            </li>
            <li>
              <Text strong>租户隔离：</Text>权限检查时会严格按照租户进行隔离
            </li>
            <li>
              <Text strong>权限叠加：</Text>角色权限和直接权限进行叠加
            </li>
            <li>
              <Text strong>管理权限：</Text>拥有 admin
              操作的用户可执行该资源的所有操作
            </li>
            <li>
              <Text strong>超级权限：</Text>
              <Text code>*:*</Text> 权限可以访问所有资源
            </li>
          </ol>

          <Divider />

          <Title level={4}>安全操作场景</Title>
          <ul>
            <li>
              <Text strong>普通权限授予：</Text>可以自由授予撤销，如{" "}
              <Text code>document:read</Text> ✅
            </li>
            <li>
              <Text strong>系统权限保护：</Text>完全不可变更，如{" "}
              <Text code>tenant:create</Text>、<Text code>user:manage</Text> ❌
            </li>
            <li>
              <Text strong>自我提权防护：</Text>用户不能给自己授予管理权限 ❌
            </li>
            <li>
              <Text strong>角色系统权限保护：</Text>角色更新时系统权限自动保留
              🔒
            </li>
          </ul>

          <Divider />

          <Title level={4}>安全错误类型</Title>
          <ul>
            <li>
              <Text strong>ErrSelfElevationPrevented：</Text>自我提权被阻止
            </li>
            <li>
              <Text strong>ErrSystemPermissionImmutable：</Text>系统权限不可变更
            </li>
            <li>
              <Text strong>ErrInvalidPermissionType：</Text>无效的权限类型
            </li>
          </ul>

          <Divider />

          <Title level={4}>特殊权限</Title>
          <Paragraph>
            <Text strong>超级管理员：</Text>拥有 <Text code>*:*</Text>{" "}
            权限的用户，对所有租户的所有资源都有完全权限。
          </Paragraph>
        </Typography>
      </Card>
    </div>
  );
};

export default PermissionGuide;
