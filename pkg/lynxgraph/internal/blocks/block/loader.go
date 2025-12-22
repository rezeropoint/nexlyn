package block

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/skylark/journeys"

	// "github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/anomaly"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/cleardedupcontext"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/dedupcheck"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/filterbyset"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/foreach"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/getdedupset"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/holidaycheck"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/httprequest"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/log"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/querydatabase"

	// "github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/queryhistorydata"
	// "github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/rateofchange"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/schedule"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/setcontainscheck"
	// "github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/statistical"
	switchblock "github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/switch"
	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/timewindow"

	// "github.com/rezeropoint/nexlyn/pkg/lynxgraph/internal/blocks/standard/trend"d

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

	// // 注册 FetchSensorDataBlock（拉取传感器数据）
	// queryHistoryDataSpec := queryhistorydata.GetQueryHistoryDataSpec()
	// // 检查服务依赖
	// if !checkServiceDependencies(queryHistoryDataSpec, service) {
	// 	logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
	// 		core.BlockTypeFetchSensorData,
	// 		queryHistoryDataSpec.RequiredServiceTypes())
	// } else {
	// 	err := reg.Register(
	// 		core.BlockKey{BlockType: core.BlockTypeFetchSensorData, Version: queryHistoryDataSpec.Version()},
	// 		queryhistorydata.NewQueryHistoryDataBlock,
	// 		queryHistoryDataSpec,
	// 	)
	// 	if err != nil {
	// 		logx.Must(core.ErrRegisterStandardBlock)
	// 	}
	// }

	// // 注册 StatisticalAnalyzerBlock（统计分析）
	// statisticalSpec := statistical.GetStatisticalAnalyzerSpec()
	// if !checkServiceDependencies(statisticalSpec, service) {
	// 	logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
	// 		core.BlockTypeStatisticalAnalyzer,
	// 		statisticalSpec.RequiredServiceTypes())
	// } else {
	// 	err := reg.Register(
	// 		core.BlockKey{BlockType: core.BlockTypeStatisticalAnalyzer, Version: statisticalSpec.Version()},
	// 		statistical.NewStatisticalAnalyzerBlock,
	// 		statisticalSpec,
	// 	)
	// 	if err != nil {
	// 		logx.Must(core.ErrRegisterStandardBlock)
	// 	}
	// }

	// // 注册 RateOfChangeBlock（变化率计算）
	// rateOfChangeSpec := rateofchange.GetRateOfChangeSpec()
	// if !checkServiceDependencies(rateOfChangeSpec, service) {
	// 	logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
	// 		core.BlockTypeRateOfChange,
	// 		rateOfChangeSpec.RequiredServiceTypes())
	// } else {
	// 	err := reg.Register(
	// 		core.BlockKey{BlockType: core.BlockTypeRateOfChange, Version: rateOfChangeSpec.Version()},
	// 		rateofchange.NewRateOfChangeBlock,
	// 		rateOfChangeSpec,
	// 	)
	// 	if err != nil {
	// 		logx.Must(core.ErrRegisterStandardBlock)
	// 	}
	// }

	// // 注册 AnomalyDetectorBlock（异常检测）
	// anomalySpec := anomaly.GetAnomalyDetectorSpec()
	// if !checkServiceDependencies(anomalySpec, service) {
	// 	logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
	// 		core.BlockTypeAnomalyDetector,
	// 		anomalySpec.RequiredServiceTypes())
	// } else {
	// 	err := reg.Register(
	// 		core.BlockKey{BlockType: core.BlockTypeAnomalyDetector, Version: anomalySpec.Version()},
	// 		anomaly.NewAnomalyDetectorBlock,
	// 		anomalySpec,
	// 	)
	// 	if err != nil {
	// 		logx.Must(core.ErrRegisterStandardBlock)
	// 	}
	// }

	// // 注册 TrendAnalyzerBlock（趋势分析）
	// trendSpec := trend.GetTrendAnalyzerSpec()
	// if !checkServiceDependencies(trendSpec, service) {
	// 	logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
	// 		core.BlockTypeTrendAnalyzer,
	// 		trendSpec.RequiredServiceTypes())
	// } else {
	// 	err := reg.Register(
	// 		core.BlockKey{BlockType: core.BlockTypeTrendAnalyzer, Version: trendSpec.Version()},
	// 		trend.NewTrendAnalyzerBlock,
	// 		trendSpec,
	// 	)
	// 	if err != nil {
	// 		logx.Must(core.ErrRegisterStandardBlock)
	// 	}
	// }

	// 注册 HttpRequestBlock（HTTP 请求）
	httpRequestSpec := httprequest.GetHttpRequestSpec()
	if !checkServiceDependencies(httpRequestSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeHttpRequest,
			httpRequestSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeHttpRequest, Version: httpRequestSpec.Version()},
			httprequest.NewHttpRequestBlock,
			httpRequestSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 TimeWindowCheckBlock（时间窗口检查）
	timeWindowCheckSpec := timewindow.GetTimeWindowCheckSpec()
	if !checkServiceDependencies(timeWindowCheckSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeTimeWindowCheck,
			timeWindowCheckSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeTimeWindowCheck, Version: timeWindowCheckSpec.Version()},
			timewindow.NewTimeWindowCheckBlock,
			timeWindowCheckSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 SwitchBlock（条件路由）
	switchSpec := switchblock.GetSwitchSpec()
	if !checkServiceDependencies(switchSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeSwitch,
			switchSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeSwitch, Version: switchSpec.Version()},
			switchblock.NewSwitchBlock,
			switchSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 HolidayCheckBlock（假期检查）
	holidayCheckSpec := holidaycheck.GetHolidayCheckSpec()
	if !checkServiceDependencies(holidayCheckSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeHolidayCheck,
			holidayCheckSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeHolidayCheck, Version: holidayCheckSpec.Version()},
			holidaycheck.NewHolidayCheckBlock,
			holidayCheckSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 QueryDatabaseBlock（外部数据库查询）
	queryDatabaseSpec := querydatabase.GetQueryDatabaseSpec()
	if !checkServiceDependencies(queryDatabaseSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeQueryDatabase,
			queryDatabaseSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeQueryDatabase, Version: queryDatabaseSpec.Version()},
			querydatabase.NewQueryDatabaseBlock,
			queryDatabaseSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 DedupCheckBlock（去重检查）
	dedupCheckSpec := dedupcheck.GetDedupCheckSpec()
	if !checkServiceDependencies(dedupCheckSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeDedupCheck,
			dedupCheckSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeDedupCheck, Version: dedupCheckSpec.Version()},
			dedupcheck.NewDedupCheckBlock,
			dedupCheckSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 ScheduleBlock（定时触发器）
	scheduleSpec := schedule.GetScheduleSpec()
	if !checkServiceDependencies(scheduleSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeSchedule,
			scheduleSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeSchedule, Version: scheduleSpec.Version()},
			schedule.NewScheduleBlock,
			scheduleSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 ForEachBlock（遍历数组，信息原子驱动子图）
	forEachSpec := foreach.GetForEachSpec()
	if !checkServiceDependencies(forEachSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeForEach,
			forEachSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeForEach, Version: forEachSpec.Version()},
			foreach.NewForEachBlock,
			forEachSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 GetDedupSetBlock（获取去重集合）
	getDedupSetSpec := getdedupset.GetGetDedupSetSpec()
	if !checkServiceDependencies(getDedupSetSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeGetDedupSet,
			getDedupSetSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeGetDedupSet, Version: getDedupSetSpec.Version()},
			getdedupset.NewGetDedupSetBlock,
			getDedupSetSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 SetContainsCheckBlock（集合包含检查）
	setContainsCheckSpec := setcontainscheck.GetSetContainsCheckSpec()
	if !checkServiceDependencies(setContainsCheckSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeSetContainsCheck,
			setContainsCheckSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeSetContainsCheck, Version: setContainsCheckSpec.Version()},
			setcontainscheck.NewSetContainsCheckBlock,
			setContainsCheckSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 ClearDedupContextBlock（清空去重记录）
	clearDedupContextSpec := cleardedupcontext.GetClearDedupContextSpec()
	if !checkServiceDependencies(clearDedupContextSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeClearDedupContext,
			clearDedupContextSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeClearDedupContext, Version: clearDedupContextSpec.Version()},
			cleardedupcontext.NewClearDedupContextBlock,
			clearDedupContextSpec,
		)
		if err != nil {
			logx.Must(core.ErrRegisterStandardBlock)
		}
	}

	// 注册 FilterBySetBlock（按集合过滤数组）
	filterBySetSpec := filterbyset.GetFilterBySetSpec()
	if !checkServiceDependencies(filterBySetSpec, service) {
		logx.Infof("跳过注册 %s: 缺少服务依赖 %v",
			core.BlockTypeFilterBySet,
			filterBySetSpec.RequiredServiceTypes())
	} else {
		err := reg.Register(
			core.BlockKey{BlockType: core.BlockTypeFilterBySet, Version: filterBySetSpec.Version()},
			filterbyset.NewFilterBySetBlock,
			filterBySetSpec,
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
