package ossutil

import (
	"fmt"
)

// NewClient 创建阿里云OSS客户端
func NewClient(cfg Config) (Client, error) {
	return NewAliyunClient(cfg)
}

// ValidateConfig 验证OSS配置是否完整
func ValidateConfig(cfg Config) error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("OSS端点不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID不能为空")
	}
	if cfg.AccessKeySecret == "" {
		return fmt.Errorf("AccessKeySecret不能为空")
	}
	if cfg.BucketName == "" {
		return fmt.Errorf("存储桶名称不能为空")
	}
	return nil
}
