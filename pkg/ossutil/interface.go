package ossutil

import (
	"context"
	"io"
)

// Client OSS客户端接口
type Client interface {
	// UploadFile 上传文件到OSS
	// objectKey: 对象存储路径，如 "avatars/tenant1/user123.jpg"
	// reader: 文件内容
	// contentType: 文件MIME类型
	// 返回文件的公网访问URL
	UploadFile(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error)

	// DeleteFile 删除OSS中的文件
	// objectKey: 对象存储路径
	DeleteFile(ctx context.Context, objectKey string) error

	// GetFileURL 获取文件的访问URL
	// objectKey: 对象存储路径
	GetFileURL(objectKey string) string

	// Close 关闭客户端连接
	Close() error
}

// Config OSS客户端配置（仅支持阿里云）
type Config struct {
	Endpoint        string // OSS端点（如: oss-cn-hangzhou.aliyuncs.com）
	AccessKeyID     string // 访问密钥ID
	AccessKeySecret string // 访问密钥Secret
	BucketName      string // 存储桶名称
	CDNDomain       string // CDN域名（可选，用于加速访问）
}
