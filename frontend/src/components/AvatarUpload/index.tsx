import { PlusOutlined } from '@ant-design/icons';
import { App, Avatar, Upload } from 'antd';
import type { RcFile, UploadChangeParam, UploadFile, UploadProps } from 'antd/es/upload';
import ImgCrop from 'antd-img-crop';
import React, { useState } from 'react';
import './index.less';

export interface AvatarUploadProps {
  value?: string; // 头像URL
  onChange?: (url: string) => void; // 头像变更回调
  userName?: string; // 用户名（用于默认头像）
  size?: number; // 头像大小
  disabled?: boolean; // 是否禁用
}

const AvatarUpload: React.FC<AvatarUploadProps> = ({
  value,
  onChange,
  userName = 'User',
  size = 100,
  disabled = false,
}) => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [imageUrl, setImageUrl] = useState<string | undefined>(value);

  // 获取用户名首字母
  const getInitials = (name: string): string => {
    return name?.charAt(0).toUpperCase() || 'U';
  };

  // 上传前验证
  const beforeUpload = (file: RcFile) => {
    const isImage = file.type.startsWith('image/');
    if (!isImage) {
      message.error('只能上传图片文件！');
      return false;
    }

    const isLt5M = file.size / 1024 / 1024 < 5;
    if (!isLt5M) {
      message.error('图片大小不能超过5MB！');
      return false;
    }

    return true;
  };

  // 上传状态改变
  const handleChange: UploadProps['onChange'] = (info: UploadChangeParam<UploadFile>) => {
    if (info.file.status === 'uploading') {
      setLoading(true);
      return;
    }

    if (info.file.status === 'done') {
      // 处理上传成功响应
      const response = info.file.response;
      if (response && response.code === 0) {
        const avatarUrl = response.data?.avatarUrl;
        setImageUrl(avatarUrl);
        setLoading(false);
        message.success('头像上传成功！');

        // 触发onChange回调
        if (onChange && avatarUrl) {
          onChange(avatarUrl);
        }
      } else {
        setLoading(false);
        message.error(response?.msg || '上传失败');
      }
    }

    if (info.file.status === 'error') {
      setLoading(false);
      message.error('上传失败，请稍后重试');
    }
  };

  // 根据头像大小动态计算字体大小
  const fontSize = Math.floor(size * 0.35);

  return (
    <div className="avatar-upload-container" style={{ width: size, height: size }}>
      <ImgCrop
        rotationSlider
        aspectSlider
        showGrid
        quality={0.8}
        modalTitle="裁剪头像"
        modalOk="确定"
        modalCancel="取消"
      >
        <Upload
          name="file"
          listType="text"
          className="avatar-uploader"
          showUploadList={false}
          action="/api/v1/user/avatar/upload"
          beforeUpload={beforeUpload}
          onChange={handleChange}
          disabled={disabled || loading}
          headers={{
            Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
          }}
        >
          <div className="avatar-wrapper">
            {imageUrl ? (
              <Avatar src={imageUrl} size={size} className="avatar-image">
                {!imageUrl && getInitials(userName)}
              </Avatar>
            ) : (
              <Avatar size={size} className="avatar-placeholder" style={{ fontSize }}>
                {getInitials(userName)}
              </Avatar>
            )}
            {!disabled && !loading && (
              <div className="avatar-upload-mask">
                <PlusOutlined />
                <div>更换头像</div>
              </div>
            )}
          </div>
        </Upload>
      </ImgCrop>
    </div>
  );
};

export default AvatarUpload;
