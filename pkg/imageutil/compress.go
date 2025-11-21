package imageutil

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

// CompressOptions 图片压缩选项
type CompressOptions struct {
	MaxWidth  int // 最大宽度，0表示不限制
	MaxHeight int // 最大高度，0表示不限制
	Quality   int // JPEG质量，1-100
}

// DefaultCompressOptions 默认压缩选项
var DefaultCompressOptions = CompressOptions{
	MaxWidth:  800,
	MaxHeight: 800,
	Quality:   85,
}

// CompressImage 压缩图片
// 如果图片尺寸超过限制，将按比例缩放
// 返回压缩后的图片数据和新的文件名（扩展名可能改变）
func CompressImage(reader io.Reader, filename string, opts CompressOptions) (io.Reader, string, error) {
	// 解码图片
	img, _, err := decodeImage(reader)
	if err != nil {
		return nil, "", fmt.Errorf("解码图片失败: %w", err)
	}

	// 检查是否需要缩放
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	needResize := false
	newWidth := width
	newHeight := height

	// 按比例缩放
	if opts.MaxWidth > 0 && width > opts.MaxWidth {
		ratio := float64(opts.MaxWidth) / float64(width)
		newWidth = opts.MaxWidth
		newHeight = int(float64(height) * ratio)
		needResize = true
	}
	if opts.MaxHeight > 0 && newHeight > opts.MaxHeight {
		ratio := float64(opts.MaxHeight) / float64(newHeight)
		newHeight = opts.MaxHeight
		newWidth = int(float64(newWidth) * ratio)
		needResize = true
	}

	// 执行缩放
	var resizedImg image.Image
	if needResize {
		resizedImg = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	} else {
		resizedImg = img
	}

	// 编码为JPEG（统一转换为JPEG以减小文件大小）
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: opts.Quality})
	if err != nil {
		return nil, "", fmt.Errorf("编码JPEG失败: %w", err)
	}

	// 修改文件扩展名为.jpg
	newFilename := changeExtension(filename, ".jpg")

	return &buf, newFilename, nil
}

// decodeImage 解码图片，支持多种格式
func decodeImage(reader io.Reader) (image.Image, string, error) {
	// 读取所有数据
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}

	// 尝试解码
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// 如果标准库解码失败，尝试webp
		img, err = webp.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, "", err
		}
		return img, "webp", nil
	}

	return img, format, nil
}

// changeExtension 更改文件扩展名
func changeExtension(filename, newExt string) string {
	// 移除原有扩展名
	base := filename
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			base = filename[:i]
			break
		}
	}
	return base + newExt
}

// EncodeToFormat 根据格式编码图片
func EncodeToFormat(img image.Image, format string, quality int) (io.Reader, error) {
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
	case "png":
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的图片格式: %s", format)
	}
	return &buf, nil
}
