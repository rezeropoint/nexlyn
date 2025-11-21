package block

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/skylark/journeys"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/anomaly"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/log"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/queryhistorydata"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/rateofchange"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/statistical"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/trend"

	"github.com/zeromicro/go-zero/core/logx"
)

// RegisterStandardBlocks 注册所有标准逻辑块到指定的注册表
// 只注册依赖的服务已就绪的逻辑块，缺少服务依赖的逻辑块会被跳过并记录日志
func RegisterStandardBlocks(reg BlockRegistry, service core.Service) error {
	if reg == nil {
		return core.ErrRegistryNil
	}

	// 注册 SkylarkJourneyCreateBlock
	skylarkJourneyCreateSpec := journeys.GetSkylarkJourneyCreateSpec()
	// 检查服务依赖
	if !checkServiceDependencies(skylarkJourneyCreateSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeSkylarkJourneyCreate,
			skylarkJourneyCreateSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeSkylarkJourneyCreate, Version: skylarkJourneyCreateSpec.Version()},
			journeys.NewSkylarkJourneyCreateBlock,
			skylarkJourneyCreateSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 LogBlock
	logSpec := log.GetLogSpec()
	// 检查服务依赖
	if !checkServiceDependencies(logSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeLog,
			logSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeLog, Version: logSpec.Version()},
			log.NewLogBlock,
			logSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 FetchSensorDataBlock（拉取传感器数据）
	queryHistoryDataSpec := queryhistorydata.GetQueryHistoryDataSpec()
	// 检查服务依赖
	if !checkServiceDependencies(queryHistoryDataSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeFetchSensorData,
			queryHistoryDataSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeFetchSensorData, Version: queryHistoryDataSpec.Version()},
			queryhistorydata.NewQueryHistoryDataBlock,
			queryHistoryDataSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 StatisticalAnalyzerBlock（统计分析）
	statisticalSpec := statistical.GetStatisticalAnalyzerSpec()
	if !checkServiceDependencies(statisticalSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeStatisticalAnalyzer,
			statisticalSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeStatisticalAnalyzer, Version: statisticalSpec.Version()},
			statistical.NewStatisticalAnalyzerBlock,
			statisticalSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 RateOfChangeBlock（变化率计算）
	rateOfChangeSpec := rateofchange.GetRateOfChangeSpec()
	if !checkServiceDependencies(rateOfChangeSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeRateOfChange,
			rateOfChangeSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeRateOfChange, Version: rateOfChangeSpec.Version()},
			rateofchange.NewRateOfChangeBlock,
			rateOfChangeSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 AnomalyDetectorBlock（异常检测）
	anomalySpec := anomaly.GetAnomalyDetectorSpec()
	if !checkServiceDependencies(anomalySpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeAnomalyDetector,
			anomalySpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeAnomalyDetector, Version: anomalySpec.Version()},
			anomaly.NewAnomalyDetectorBlock,
			anomalySpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 TrendAnalyzerBlock（趋势分析）
	trendSpec := trend.GetTrendAnalyzerSpec()
	if !checkServiceDependencies(trendSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeTrendAnalyzer,
			trendSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeTrendAnalyzer, Version: trendSpec.Version()},
			trend.NewTrendAnalyzerBlock,
			trendSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	return nil
}

// checkServiceDependencies 检查逻辑块依赖的服务是否都已注册
func checkServiceDependencies(spec core.BlockSpec, service core.Service) bool {
	if service == nil {
		// 如果没有提供 service，说明不需要检查依赖（例如在 Manager 模式下）
		return true
	}

	requiredTypes := spec.RequiredServiceTypes()
	if len(requiredTypes) == 0 {
		// 没有服务依赖，直接通过
		return true
	}

	// 检查所有依赖的服务是否已注册
	for _, serviceType := range requiredTypes {
		if !service.HasServiceType(serviceType) {
			return false
		}
	}

	return true
}
