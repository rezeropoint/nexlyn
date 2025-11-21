import { TeamOutlined } from "@ant-design/icons";
import { Alert, Card, Descriptions, Space, Table, Tag, Typography } from "antd";
import React from "react";

const { Title, Paragraph, Text } = Typography;

/**
 * 角色与权限分级说明组件
 * 基于 CasbinX v0.7.11+ 安全特性
 */
const RoleGuide: React.FC = () => {
  // 权限分级示例 - 基于 README_SECURITY.md
  const permissionTypes = [
    {
      key: "system",
      type: "系统权限",
      typeCode: "PermissionTypeSystem",
      description: "完全不可变更的核心系统权限，任何人都无法授予或撤销",
      examples: [
        "tenant:write",
        "tenant:delete",
        "system:write",
        "user:write",
        "permission:write",
      ],
      features: ["完全不可变更", "系统核心功能", "安全保护", "不可授予撤销"],
      color: "red",
    },
    {
      key: "normal",
      type: "普通权限",
      typeCode: "PermissionTypeNormal",
      description: "可自由授予和撤销的日常操作权限",
      examples: [
        "document:read",
        "document:write",
        "graph:read",
        "graph:write",
        "infoatom:read",
      ],
      features: ["自由授予撤销", "日常操作", "无特殊限制"],
      color: "green",
    },
  ];

  // 业务角色示例 - 实际系统中可创建的角色
  const roleExamples = [
    {
      key: "admin",
      roleKey: "admin",
      roleName: "管理员",
      description: "租户内的管理员角色，拥有业务数据的完全管理权限",
      permissions: [
        "graph:read",
        "graph:write",
        "graph:delete",
        "infoatom:read",
        "infoatom:write",
        "infoatom:delete",
        "tags:read",
        "tags:write",
        "tags:delete",
      ],
      features: ["图配置管理", "信息原子管理", "标签管理", "业务数据完全控制"],
      permissionTypes: ["普通权限"],
    },
    {
      key: "editor",
      roleKey: "editor",
      roleName: "编辑者",
      description: "可以编辑业务数据但不能删除的角色",
      permissions: [
        "graph:read",
        "graph:write",
        "infoatom:read",
        "infoatom:write",
        "tags:read",
        "tags:write",
      ],
      features: ["图配置编辑", "信息原子编辑", "标签编辑"],
      permissionTypes: ["普通权限"],
    },
    {
      key: "viewer",
      roleKey: "viewer",
      roleName: "查看者",
      description: "只读角色，可以查看所有业务数据但不能修改",
      permissions: ["graph:read", "infoatom:read", "tags:read"],
      features: ["查看图配置", "查看信息原子", "查看标签"],
      permissionTypes: ["普通权限"],
    },
  ];

  // 业务角色等级 - 实际系统中的角色层次
  const roleHierarchy = [
    {
      key: 1,
      level: 1,
      role: "admin",
      name: "管理员",
      scope: "租户内",
      description: "最高业务权限级别，拥有业务数据的完全管理权限",
    },
    {
      key: 2,
      level: 2,
      role: "editor",
      name: "编辑者",
      scope: "租户内",
      description: "中等权限级别，可以编辑业务数据但不能删除",
    },
    {
      key: 3,
      level: 3,
      role: "viewer",
      name: "查看者",
      scope: "租户内",
      description: "基础权限级别，只能查看业务数据",
    },
  ];

  // 权限分级表格列
  const permissionTypeColumns = [
    {
      title: "权限类型",
      dataIndex: "type",
      key: "type",
      render: (type: string, record: any) => (
        <Tag color={record.color}>
          <Text strong>{type}</Text>
        </Tag>
      ),
    },
    {
      title: "类型常量",
      dataIndex: "typeCode",
      key: "typeCode",
      render: (code: string) => <Text code>{code}</Text>,
    },
    {
      title: "描述",
      dataIndex: "description",
      key: "description",
    },
    {
      title: "权限示例",
      dataIndex: "examples",
      key: "examples",
      render: (examples: string[]) => (
        <Space wrap>
          {examples.map((example) => (
            <Tag key={example} color="blue">
              {example}
            </Tag>
          ))}
        </Space>
      ),
    },
    {
      title: "安全特性",
      dataIndex: "features",
      key: "features",
      render: (features: string[]) => (
        <Space wrap>
          {features.map((feature) => (
            <Tag key={feature} color="cyan">
              {feature}
            </Tag>
          ))}
        </Space>
      ),
    },
  ];

  // 角色示例表格列
  const roleColumns = [
    {
      title: "角色键",
      dataIndex: "roleKey",
      key: "roleKey",
      render: (roleKey: string) => <Tag color="blue">{roleKey}</Tag>,
    },
    {
      title: "角色名称",
      dataIndex: "roleName",
      key: "roleName",
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: "描述",
      dataIndex: "description",
      key: "description",
    },
    {
      title: "权限示例",
      dataIndex: "permissions",
      key: "permissions",
      render: (permissions: string[]) => (
        <Space wrap>
          {permissions.slice(0, 3).map((permission) => (
            <Tag key={permission} color="green">
              {permission}
            </Tag>
          ))}
          {permissions.length > 3 && (
            <Text type="secondary">等{permissions.length}个权限</Text>
          )}
        </Space>
      ),
    },
    {
      title: "包含权限类型",
      dataIndex: "permissionTypes",
      key: "permissionTypes",
      render: (types: string[]) => (
        <Space wrap>
          {types.map((type) => {
            const color = type === "系统权限" ? "red" : "green";
            return (
              <Tag key={type} color={color}>
                {type}
              </Tag>
            );
          })}
        </Space>
      ),
    },
  ];

  const hierarchyColumns = [
    {
      title: "等级",
      dataIndex: "level",
      key: "level",
      render: (level: number) => (
        <Tag
          color={
            level === 1
              ? "red"
              : level === 2
              ? "orange"
              : level === 3
              ? "blue"
              : "default"
          }
        >
          Level {level}
        </Tag>
      ),
    },
    {
      title: "角色",
      dataIndex: "role",
      key: "role",
      render: (role: string) => <Tag color="blue">{role}</Tag>,
    },
    {
      title: "角色名称",
      dataIndex: "name",
      key: "name",
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: "作用范围",
      dataIndex: "scope",
      key: "scope",
    },
    {
      title: "说明",
      dataIndex: "description",
      key: "description",
    },
  ];

  return (
    <div>
      <Alert
        message="CasbinX 权限与角色系统说明"
        description="基于 CasbinX v0.7.11+ 的权限分级保护机制，实现了防自我提权、系统权限保护、管理权限传递控制等安全特性。角色是权限的集合，通过权限分级实现精细化的安全控制。"
        type="info"
        icon={<TeamOutlined />}
        style={{ marginBottom: 24 }}
      />

      <Card title="权限分级保护机制" style={{ marginBottom: 24 }}>
        <Typography>
          <Paragraph>
            CasbinX
            将权限分为三个级别来实现精细化的安全控制，每个级别都有不同的安全约束：
          </Paragraph>
        </Typography>
        <Table
          dataSource={permissionTypes}
          columns={permissionTypeColumns}
          pagination={false}
          size="middle"
          scroll={{ x: "max-content" }}
        />
      </Card>

      <Card title="核心安全特性" style={{ marginBottom: 24 }}>
        <Descriptions column={1} bordered>
          <Descriptions.Item label="防止自我提权">
            默认启用，防止用户给自己分配管理权限，避免权限滥用。配置项：PreventSelfElevation
          </Descriptions.Item>
          <Descriptions.Item label="系统权限保护">
            系统级权限（如租户管理、用户管理、权限管理）完全不可授予也不可撤销。配置项：SystemPermissions
          </Descriptions.Item>
          <Descriptions.Item label="操作者权限验证">
            所有权限管理操作都需要提供operatorKey，验证操作者是否有执行权限
          </Descriptions.Item>
          <Descriptions.Item label="角色系统权限保护">
            角色中的系统权限在角色更新时自动保留，防止意外移除关键权限
          </Descriptions.Item>
          <Descriptions.Item label="跨域权限继承修复">
            修复了Casbin原生GetImplicitPermissionsForUser在跨域场景下的问题，正确处理角色权限继承
          </Descriptions.Item>
          <Descriptions.Item label="安全权限查询">
            提供安全版本的权限查询方法，实施访问控制：用户可查看自己的权限，有用户查看权限的用户可查看他人权限
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="角色权限等级" style={{ marginBottom: 24 }}>
        <Typography>
          <Paragraph>
            系统角色按照权限等级从高到低排序，高等级角色通常包含低等级角色的所有权限：
          </Paragraph>
        </Typography>
        <Table
          dataSource={roleHierarchy}
          columns={hierarchyColumns}
          pagination={false}
          size="middle"
        />
      </Card>

      <Card title="角色权限组合示例" style={{ marginBottom: 24 }}>
        <Typography>
          <Paragraph>
            角色是权限的集合，可以包含不同级别的权限。以下是常见的角色权限组合示例：
          </Paragraph>
        </Typography>
        <Table
          dataSource={roleExamples}
          columns={roleColumns}
          pagination={false}
          size="middle"
          scroll={{ x: "max-content" }}
        />
      </Card>

      <Card title="安全使用规则与最佳实践">
        <Typography>
          <Title level={4}>权限分配规则</Title>
          <ol>
            <li>
              <Text strong>最小权限原则：</Text>
              只授予完成工作所必需的最小权限集合
            </li>
            <li>
              <Text strong>权限分级遵循：</Text>严格按照系统权限 &gt;
              普通权限的两级层次进行管理
            </li>
            <li>
              <Text strong>租户隔离：</Text>
              权限严格按租户隔离，跨租户操作需要系统级权限
            </li>
            <li>
              <Text strong>操作者验证：</Text>
              所有权限操作都需要明确的操作者身份和权限验证
            </li>
          </ol>

          <Title level={4}>权限组合计算</Title>
          <Paragraph>
            用户的最终权限由以下来源组成，权限检查时满足任一来源即可通过：
          </Paragraph>
          <ul>
            <li>
              <Text strong>角色权限：</Text>用户被分配的所有角色的权限并集
            </li>
            <li>
              <Text strong>直接权限：</Text>通过权限规则直接分配给用户的权限
            </li>
            <li>
              <Text strong>安全约束：</Text>
              所有权限操作都受到权限分级的安全约束保护
            </li>
          </ul>

          <Title level={4}>安全操作场景</Title>
          <ul>
            <li>
              <Text strong>普通权限授予：</Text>可以自由授予撤销，如{" "}
              <Text code>document:read</Text>、<Text code>graph:write</Text> ✅
            </li>
            <li>
              <Text strong>系统权限保护：</Text>完全不可变更，如{" "}
              <Text code>tenant:write</Text>、<Text code>user:write</Text>、
              <Text code>permission:write</Text> ❌
            </li>
            <li>
              <Text strong>自我提权防护：</Text>
              用户不能给自己授予管理权限，防止权限滥用 🛡️
            </li>
            <li>
              <Text strong>角色系统权限保护：</Text>
              角色更新时系统权限自动保留，防止意外移除关键权限 🔒
            </li>
          </ul>

          <Title level={4}>权限管理工具</Title>
          <ul>
            <li>
              <Text strong>角色权限管理：</Text>
              创建和管理角色，配置角色的权限组合
            </li>
            <li>
              <Text strong>用户角色分配：</Text>
              为用户分配或移除角色，受权限验证保护
            </li>
            <li>
              <Text strong>权限规则管理：</Text>
              直接权限分配，用于补充特殊权限需求
            </li>
            <li>
              <Text strong>权限检查工具：</Text>
              查看用户的最终权限和来源，支持安全查询
            </li>
          </ul>

          <Title level={4}>权限管理约束</Title>
          <ul>
            <li>
              <Text strong>系统权限限制：</Text>
              系统权限（如用户管理、权限管理）无法通过界面授予或撤销
            </li>
            <li>
              <Text strong>操作者验证：</Text>
              所有权限操作都需要验证操作者身份和权限
            </li>
            <li>
              <Text strong>自我提权防护：</Text>用户无法给自己分配管理类权限
            </li>
            <li>
              <Text strong>权限审核：</Text>
              定期检查权限分配，确保符合最小权限原则
            </li>
            <li>
              <Text strong>安全监控：</Text>系统会记录并监控被阻止的权限操作
            </li>
            <li>
              <Text strong>错误提示：</Text>违反安全规则时会显示明确的错误信息
            </li>
          </ul>
        </Typography>
      </Card>
    </div>
  );
};

export default RoleGuide;
