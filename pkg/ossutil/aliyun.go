package ossutil

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AliyunClient 阿里云OSS客户端
type AliyunClient struct {
	client    *oss.Client
	bucket    *oss.Bucket
	config    Config
	cdnDomain string
}

// NewAliyunClient 创建阿里云OSS客户端
func NewAliyunClient(cfg Config) (*AliyunClient, error) {
	// 创建OSS客户端
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建阿里云OSS客户端失败: %w", err)
	}

	// 获取存储桶
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取存储桶失败: %w", err)
	}

	return &AliyunClient{
		client:    client,
		bucket:    bucket,
		config:    cfg,
		cdnDomain: cfg.CDNDomain,
	}, nil
}

// UploadFile 上传文件到阿里云OSS
func (c *AliyunClient) UploadFile(ctx context.Context, objectKey string, reader io.Reader, contentType string) (string, error) {
	// 设置上传选项
	options := []oss.Option{
		oss.ContentType(contentType),
		oss.ObjectACL(oss.ACLPublicRead), // 设置文件为公共读
	}

	// 上传文件
	err := c.bucket.PutObject(objectKey, reader, options...)
	if err != nil {
		return "", fmt.Errorf("上传文件到阿里云OSS失败: %w", err)
	}

	// 返回文件URL
	return c.GetFileURL(objectKey), nil
}

// DeleteFile 删除阿里云OSS中的文件
func (c *AliyunClient) DeleteFile(ctx context.Context, objectKey string) error {
	err := c.bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("删除阿里云OSS文件失败: %w", err)
	}
	return nil
}

// GetFileURL 获取文件的访问URL
func (c *AliyunClient) GetFileURL(objectKey string) string {
	// 如果配置了CDN域名，使用CDN域名
	if c.cdnDomain != "" {
		cdnDomain := strings.TrimSuffix(c.cdnDomain, "/")
		return fmt.Sprintf("%s/%s", cdnDomain, objectKey)
	}

	// 否则使用OSS原始域名
	// 格式: https://{bucket}.{endpoint}/{objectKey}
	endpoint := strings.TrimPrefix(c.config.Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	return fmt.Sprintf("https://%s.%s/%s", c.config.BucketName, endpoint, objectKey)
}

// Close 关闭客户端连接（阿里云SDK无需显式关闭）
func (c *AliyunClient) Close() error {
	return nil
}
