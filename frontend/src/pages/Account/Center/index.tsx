import {
  EditOutlined,
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
  SaveOutlined,
  UserOutlined,
  CloseOutlined,
} from '@ant-design/icons';
import { ModalForm, ProCard, ProDescriptions, ProForm, type ProFormInstance, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { PageContainer } from '@ant-design/pro-layout';
import { useModel } from '@umijs/max';
import { Avatar, Button, Space, Tag, message, theme } from 'antd';
import type { FormInstance } from 'antd';
import React, { useRef, useState } from 'react';
import AvatarUpload from '@/components/AvatarUpload';
import { updateCurrentUser, changePassword } from '@/services/user';
import styles from './index.less';

/**
 * 个人中心页面
 * 整合了个人信息查看和编辑功能
 */
const AccountCenter: React.FC = () => {
  const { initialState, setInitialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;
  const { token } = theme.useToken();
  const [isEditing, setIsEditing] = useState(false); // 编辑模式状态
  const [submitting, setSubmitting] = useState(false);
  const [passwordModalVisible, setPasswordModalVisible] = useState(false); // 修改密码弹窗
  const passwordFormRef = useRef<ProFormInstance>();
  const formRef = useRef<ProFormInstance>();

  // 获取用户名首字母
  const getInitials = (name: string): string => {
    return name?.charAt(0).toUpperCase() || 'U';
  };

  // 进入编辑模式
  const handleEdit = () => {
    setIsEditing(true);
  };

  // 取消编辑
  const handleCancel = () => {
    setIsEditing(false);
    formRef.current?.resetFields();
  };

  // 保存基本信息
  const handleSaveProfile = async (values: any) => {
    if (!currentUser?.id) {
      message.error('用户信息不完整');
      return false;
    }

    setSubmitting(true);
    try {
      // 使用 updateCurrentUser 接口（用户修改自己的资料，无需管理权限）
      const response = await updateCurrentUser({
        name: values.name,
        email: values.email,
        phone: values.phone,
        title: values.title,
        signature: values.signature,
        avatar: values.avatar || currentUser.avatar,
      });

      if (response.code === 0) {
        message.success('个人信息更新成功');

        // 更新全局状态
        if (initialState) {
          setInitialState({
            ...initialState,
            currentUser: {
              ...currentUser,
              ...values,
            },
          });
        }

        setIsEditing(false);
        return true;
      } else {
        message.error(response.msg || '更新失败');
        return false;
      }
    } catch (_error) {
      message.error('更新失败，请稍后重试');
      return false;
    } finally {
      setSubmitting(false);
    }
  };

  // 修改密码
  const handleChangePassword = async (values: { oldPassword: string; newPassword: string }) => {
    try {
      const response = await changePassword({
        oldPassword: values.oldPassword,
        newPassword: values.newPassword,
      });

      if (response.code === 0) {
        message.success('密码修改成功，请重新登录');
        // 可选：强制用户重新登录
        // setTimeout(() => {
        //   history.push('/user/login');
        // }, 1500);
        return true;
      }
      message.error(response.msg || '密码修改失败');
      return false;
    } catch (error) {
      console.error('修改密码失败:', error);
      return false;
    }
  };

  return (
    <PageContainer className={styles['account-center']}>
      {/* 封面区域 */}
      <div className={styles['cover-section']}>
        <div className={styles['cover-actions']}>
          {!isEditing ? (
            <Button
              type="primary"
              size="large"
              icon={<EditOutlined />}
              onClick={handleEdit}
            >
              编辑资料
            </Button>
          ) : (
            <Space>
              <Button
                size="large"
                icon={<CloseOutlined />}
                onClick={handleCancel}
              >
                取消
              </Button>
              <Button
                type="primary"
                size="large"
                icon={<SaveOutlined />}
                loading={submitting}
                onClick={() => formRef.current?.submit()}
              >
                保存
              </Button>
            </Space>
          )}
        </div>
      </div>

      {/* 个人信息头部 */}
      <div className={styles['profile-header']}>
        <div className={styles['avatar-container']}>
          <div className={styles['avatar-wrapper']}>
            {isEditing ? (
              <ProForm.Item noStyle shouldUpdate>
                {(form: FormInstance) => {
                  const avatarUrl = form.getFieldValue('avatar') || currentUser?.avatar;
                  return (
                    <AvatarUpload
                      value={avatarUrl}
                      userName={currentUser?.userName}
                      size={138}
                      onChange={(url) => {
                        form.setFieldValue('avatar', url);
                      }}
                    />
                  );
                }}
              </ProForm.Item>
            ) : (
              currentUser?.avatar ? (
                <Avatar src={currentUser.avatar} size={138} />
              ) : (
                <Avatar size={138} style={{ backgroundColor: token.colorPrimary, fontSize: 48 }}>
                  {getInitials(currentUser?.userName || 'User')}
                </Avatar>
              )
            )}
          </div>
          {!isEditing && (
            <div className={styles['avatar-tip']}>点击编辑资料以更换头像</div>
          )}
        </div>

        <div className={styles['profile-info']}>
          <h1 className={styles['user-name']}>{currentUser?.name || '未设置姓名'}</h1>
          <div className={styles['user-meta']}>
            <span className={styles['meta-item']}>
              <UserOutlined />
              {currentUser?.userName}
            </span>
            {currentUser?.email && (
              <span className={styles['meta-item']}>
                <MailOutlined />
                {currentUser?.email}
              </span>
            )}
            {currentUser?.phone && (
              <span className={styles['meta-item']}>
                <PhoneOutlined />
                {currentUser?.phone}
              </span>
            )}
          </div>
          <p className={styles['user-signature']}>
            {currentUser?.signature || '这个人很懒，什么都没留下...'}
          </p>
        </div>
      </div>

      {/* 信息卡片区域 */}
      <div className={styles['info-cards']}>
        {/* 基本信息卡片 */}
        <ProCard title="基本信息" className={isEditing ? styles['edit-mode'] : styles['view-mode']}>
          {isEditing ? (
            <ProForm
              formRef={formRef}
              layout="vertical"
              initialValues={{
                name: currentUser?.name,
                email: currentUser?.email,
                phone: currentUser?.phone,
                title: currentUser?.title,
                signature: currentUser?.signature,
                avatar: currentUser?.avatar,
              }}
              onFinish={handleSaveProfile}
              submitter={false}
            >
              <ProForm.Item name="avatar" hidden>
                <input type="hidden" />
              </ProForm.Item>

              <ProFormText
                name="name"
                label="姓名"
                placeholder="请输入姓名"
                rules={[{ required: true, message: '请输入姓名' }]}
              />

              <ProFormText
                name="email"
                label="邮箱"
                placeholder="请输入邮箱"
                rules={[
                  { required: true, message: '请输入邮箱' },
                  { type: 'email', message: '请输入有效的邮箱地址' },
                ]}
              />

              <ProFormText
                name="phone"
                label="电话"
                placeholder="请输入电话号码"
              />

              <ProFormText
                name="title"
                label="职位"
                placeholder="请输入职位"
              />

              <ProFormTextArea
                name="signature"
                label="个性签名"
                placeholder="请输入个性签名"
                fieldProps={{
                  rows: 4,
                  maxLength: 200,
                  showCount: true,
                }}
              />
            </ProForm>
          ) : (
            <ProDescriptions column={2}>
              <ProDescriptions.Item label="姓名">
                {currentUser?.name || '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="用户名">
                {currentUser?.userName}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="邮箱">
                {currentUser?.email || '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="电话">
                {currentUser?.phone || '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="职位">
                {currentUser?.title || '-'}
              </ProDescriptions.Item>
              <ProDescriptions.Item label="个性签名" span={2}>
                {currentUser?.signature || '-'}
              </ProDescriptions.Item>
            </ProDescriptions>
          )}
        </ProCard>

        {/* 账号安全卡片 */}
        <ProCard title="账号安全">
          <div className={styles['security-item']}>
            <div className={styles['item-info']}>
              <div className={styles['item-title']}>登录密码</div>
              <div className={styles['item-desc']}>
                定期修改密码可以提高账号安全性
              </div>
            </div>
            <Button
              icon={<LockOutlined />}
              onClick={() => setPasswordModalVisible(true)}
            >
              修改密码
            </Button>
          </div>
        </ProCard>

        {/* 账号信息卡片 */}
        <ProCard title="账号信息">
          <ProDescriptions column={2}>
            <ProDescriptions.Item label="租户">
              {currentUser?.tenantInfo?.tenantName || '-'}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="角色">
              <Tag color="blue">{currentUser?.role}</Tag>
            </ProDescriptions.Item>
            <ProDescriptions.Item label="账号状态">
              {currentUser?.status === 'active' ? (
                <Tag color="success">活跃</Tag>
              ) : (
                <Tag color="default">未激活</Tag>
              )}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="国家">
              {currentUser?.country || '-'}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="地址" span={2}>
              {currentUser?.address || '-'}
            </ProDescriptions.Item>
          </ProDescriptions>
        </ProCard>

        {/* 用户标签卡片 */}
        {currentUser?.tags && currentUser.tags.length > 0 && (
          <ProCard title="用户标签" className={styles['tags-card']}>
            <Space wrap>
              {currentUser.tags.map((tag) => (
                <Tag key={tag.id}>
                  {tag.label}
                </Tag>
              ))}
            </Space>
          </ProCard>
        )}
      </div>

      {/* 修改密码弹窗 */}
      <ModalForm
        title="修改密码"
        open={passwordModalVisible}
        onOpenChange={setPasswordModalVisible}
        formRef={passwordFormRef}
        onFinish={handleChangePassword}
        modalProps={{
          destroyOnHidden: true,
          width: 500,
          okText: '确认修改',
          cancelText: '取消',
        }}
      >
        <div className={styles['password-tips']}>
          <strong>密码要求：</strong>
          <ul>
            <li>长度至少8个字符</li>
            <li>包含大小写字母、数字和特殊字符</li>
            <li>不能与用户名相同</li>
          </ul>
        </div>

        <ProFormText.Password
          label="当前密码"
          name="oldPassword"
          fieldProps={{ size: 'large' }}
          rules={[{ required: true, message: '请输入当前密码' }]}
          placeholder="请输入当前密码"
        />

        <ProFormText.Password
          label="新密码"
          name="newPassword"
          fieldProps={{ size: 'large' }}
          rules={[
            { required: true, message: '请输入新密码' },
            { min: 8, message: '密码长度至少8个字符' },
          ]}
          placeholder="请输入新密码"
        />

        <ProFormText.Password
          label="确认新密码"
          name="confirmPassword"
          fieldProps={{ size: 'large' }}
          dependencies={['newPassword']}
          rules={[
            { required: true, message: '请确认新密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('newPassword') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('两次输入的密码不一致'));
              },
            }),
          ]}
          placeholder="请再次输入新密码"
        />
      </ModalForm>
    </PageContainer>
  );
};

export default AccountCenter;
