package imageutil

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// AllowedImageTypes 允许的图片MIME类型
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
}

// AllowedImageExtensions 允许的图片扩展名
var AllowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// ValidateImageFile 验证文件是否为有效的图片文件
// 检查文件扩展名和MIME类型
func ValidateImageFile(filename string, reader io.Reader) error {
	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	if !AllowedImageExtensions[ext] {
		return fmt.Errorf("不支持的文件类型: %s, 仅支持 jpg/jpeg/png/webp", ext)
	}

	// 检测文件的MIME类型（读取前512字节）
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 检测MIME类型
	contentType := http.DetectContentType(buffer[:n])
	if !AllowedImageTypes[contentType] {
		return fmt.Errorf("文件内容类型不匹配: %s, 仅支持图片文件", contentType)
	}

	return nil
}

// ValidateFileSize 验证文件大小
func ValidateFileSize(size int64, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("文件大小超过限制: %d bytes (最大: %d bytes)", size, maxSize)
	}
	return nil
}

// GetContentType 根据文件扩展名获取Content-Type
func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
