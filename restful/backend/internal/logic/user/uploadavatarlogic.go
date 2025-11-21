// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rezeropoint/nexlyn/internal/auth"
	"github.com/rezeropoint/nexlyn/pkg/imageutil"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/svc"
	"github.com/rezeropoint/nexlyn/restful/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传用户头像
func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadAvatarLogic) UploadAvatar(r *http.Request) (resp *types.UploadAvatarResponse, err error) {
	// 1. 检查OSS是否启用
	if !l.svcCtx.Config.OSSConfig.Enabled {
		logx.WithContext(l.ctx).Error("OSS功能未启用")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "头像上传功能未启用，请联系管理员配置OSS",
			},
		}, nil
	}

	// 2. 检查OSS客户端是否初始化
	if l.svcCtx.OSSClient == nil {
		logx.WithContext(l.ctx).Error("OSS客户端未初始化")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "OSS客户端未初始化",
			},
		}, nil
	}

	// 3. 从JWT获取用户信息
	jwtUser, err := auth.GetUserFromJWT(l.ctx)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("error", err.Error()),
		).Error("从JWT获取用户信息失败")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 401,
				Msg:  "未授权",
			},
		}, nil
	}

	// 4. 解析multipart表单
	err = r.ParseMultipartForm(l.svcCtx.Config.OSSConfig.MaxFileSize)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("error", err.Error()),
		).Error("解析multipart表单失败")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "请求格式错误",
			},
		}, nil
	}

	// 5. 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("error", err.Error()),
		).Error("获取上传文件失败")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  "未找到上传文件，请使用 'file' 字段上传",
			},
		}, nil
	}
	defer file.Close()

	// 6. 验证文件大小
	if header.Size > l.svcCtx.Config.OSSConfig.MaxFileSize {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("fileSize", header.Size),
			logx.Field("maxSize", l.svcCtx.Config.OSSConfig.MaxFileSize),
		).Error("文件大小超过限制")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  fmt.Sprintf("文件大小超过限制（最大 %d MB）", l.svcCtx.Config.OSSConfig.MaxFileSize/1024/1024),
			},
		}, nil
	}

	// 7. 读取文件内容
	fileData, err := io.ReadAll(file)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("error", err.Error()),
		).Error("读取文件内容失败")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "读取文件失败",
			},
		}, nil
	}

	// 8. 验证文件类型
	err = imageutil.ValidateImageFile(header.Filename, bytes.NewReader(fileData))
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("error", err.Error()),
			logx.Field("filename", header.Filename),
		).Error("文件验证失败")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 400,
				Msg:  err.Error(),
			},
		}, nil
	}

	// 9. 压缩图片
	compressedReader, newFilename, err := imageutil.CompressImage(
		bytes.NewReader(fileData),
		header.Filename,
		imageutil.DefaultCompressOptions,
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("error", err.Error()),
		).Error("压缩图片失败")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "图片处理失败",
			},
		}, nil
	}

	// 10. 生成对象存储路径: avatars/{tenant_id}/{user_id}_{timestamp}{ext}
	timestamp := time.Now().Unix()
	ext := filepath.Ext(newFilename)
	objectKey := fmt.Sprintf("avatars/%s/%s_%d%s", jwtUser.TenantId, jwtUser.UserId, timestamp, ext)

	// 11. 上传到OSS
	avatarURL, err := l.svcCtx.OSSClient.UploadFile(
		l.ctx,
		objectKey,
		compressedReader,
		imageutil.GetContentType(newFilename),
	)
	if err != nil {
		logx.WithContext(l.ctx).WithFields(
			logx.Field("error", err.Error()),
			logx.Field("objectKey", objectKey),
		).Error("上传到OSS失败")
		return &types.UploadAvatarResponse{
			BaseResponse: types.BaseResponse{
				Code: 500,
				Msg:  "文件上传失败",
			},
		}, nil
	}

	logx.WithContext(l.ctx).WithFields(
		logx.Field("userId", jwtUser.UserId),
		logx.Field("tenantId", jwtUser.TenantId),
		logx.Field("avatarUrl", avatarURL),
	).Info("头像上传成功")

	// 12. 返回成功响应
	return &types.UploadAvatarResponse{
		BaseResponse: types.BaseResponse{
			Code: 0,
			Msg:  "上传成功",
		},
		Data: types.UploadAvatarData{
			AvatarUrl: avatarURL,
		},
	}, nil
}
