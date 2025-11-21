package logic

import (
	"context"

	"github.com/rezeropoint/nexlyn/service/lynxengine/internal/svc"
	"github.com/rezeropoint/nexlyn/service/lynxengine/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type HealthCheckLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHealthCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HealthCheckLogic {
	return &HealthCheckLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 健康检查
// 返回服务健康状态
func (l *HealthCheckLogic) HealthCheck(in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	// 检查核心组件健康状态
	healthy := l.checkServiceHealth()

	// 记录健康检查日志
	l.Logger.Infof("健康检查完成 - 健康状态: %v", healthy)

	return &pb.HealthCheckResponse{
		Healthy: healthy,
	}, nil
}

// checkServiceHealth 检查服务核心组件的健康状态
func (l *HealthCheckLogic) checkServiceHealth() bool {
	// 检查数据库连接
	if !l.checkDatabaseHealth() {
		l.Logger.Error("数据库健康检查失败")
		return false
	}

	// 检查LynxGraph引擎状态
	if l.svcCtx.LynxEngine == nil {
		l.Logger.Error("LynxGraph引擎未初始化")
		return false
	}

	// 所有检查通过
	return true
}

// checkDatabaseHealth 检查数据库连接健康状态
func (l *HealthCheckLogic) checkDatabaseHealth() bool {
	// 检查PostgreSQL连接
	if rawDB, err := l.svcCtx.DBConn.RawDB(); err != nil {
		l.Logger.Errorf("获取PostgreSQL原生连接失败: %v", err)
		return false
	} else if err := rawDB.Ping(); err != nil {
		l.Logger.Errorf("PostgreSQL连接Ping失败: %v", err)
		return false
	}

	return true
}
