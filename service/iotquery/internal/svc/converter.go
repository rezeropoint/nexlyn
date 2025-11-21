package svc

import (
	"encoding/json"
	"time"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"
	"github.com/rezeropoint/nexlyn/service/iotquery/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

// ConvertFieldNamesToCore 转换字符串数组为FieldName数组
func ConvertFieldNamesToCore(fieldNames []string) []core.FieldName {
	result := make([]core.FieldName, 0, len(fieldNames))
	for _, name := range fieldNames {
		result = append(result, core.FieldName(name))
	}
	return result
}

// ConvertTimeSeriesReqToCore 转换protobuf请求为core.TimeSeriesQuery
func ConvertTimeSeriesReqToCore(req *pb.QueryTimeSeriesReq) core.TimeSeriesQuery {
	return core.TimeSeriesQuery{
		TenantID:    req.TenantId,
		OrgIDs:      req.OrgIds,
		DeviceIDs:   req.DeviceIds,
		FieldNames:  ConvertFieldNamesToCore(req.FieldNames),
		StartTime:   time.Unix(req.StartTime, 0),
		EndTime:     time.Unix(req.EndTime, 0),
		Aggregation: core.AggregationType(req.Aggregation),
		Interval:    time.Duration(req.IntervalSeconds) * time.Second,
		Limit:       int(req.Limit),
		Offset:      int(req.Offset),
		OrderBy:     req.OrderBy,
		OrderDir:    req.OrderDir,
	}
}

// ConvertLatestValuesReqToCore 转换protobuf请求为core.LatestValuesQuery
func ConvertLatestValuesReqToCore(req *pb.GetLatestValuesReq) core.LatestValuesQuery {
	return core.LatestValuesQuery{
		TenantID:   req.TenantId,
		OrgIDs:     req.OrgIds,
		DeviceIDs:  req.DeviceIds,
		FieldNames: ConvertFieldNamesToCore(req.FieldNames),
	}
}

// ConvertDeviceStatisticsReqToCore 转换protobuf请求为core.DeviceStatisticsQuery
func ConvertDeviceStatisticsReqToCore(req *pb.GetDeviceStatisticsReq) core.DeviceStatisticsQuery {
	return core.DeviceStatisticsQuery{
		TenantID:   req.TenantId,
		OrgIDs:     req.OrgIds,
		DeviceIDs:  req.DeviceIds,
		FieldNames: ConvertFieldNamesToCore(req.FieldNames),
		StartTime:  time.Unix(req.StartTime, 0),
		EndTime:    time.Unix(req.EndTime, 0),
	}
}

// ConvertTimeSeriesResultToProto 转换core.TimeSeriesResult为protobuf响应
func ConvertTimeSeriesResultToProto(result *core.TimeSeriesResult) *pb.QueryTimeSeriesResp {
	records := make([]*pb.TimeSeriesRecord, len(result.Data))
	for i, data := range result.Data {
		// 将Value序列化为JSON字符串
		valueJSON, err := json.Marshal(data.Value)
		if err != nil {
			logx.Errorf("Failed to marshal value: %v", err)
			valueJSON = []byte("null")
		}

		records[i] = &pb.TimeSeriesRecord{
			Timestamp:      data.Timestamp.Unix(),
			DeviceId:       data.DeviceID,
			DeviceModel:    data.DeviceModel,
			DeviceCategory: string(data.DeviceCategory),
			FieldName:      string(data.FieldName),
			ValueJson:      string(valueJSON),
		}
	}

	return &pb.QueryTimeSeriesResp{
		Records:  records,
		Total:    result.Total,
		Page:     int32(result.Page),
		PageSize: int32(result.PageSize),
	}
}

// ConvertLatestValuesToProto 转换core.DeviceLatestValues为protobuf响应
func ConvertLatestValuesToProto(devices []core.DeviceLatestValues) *pb.GetLatestValuesResp {
	pbDevices := make([]*pb.DeviceLatestValues, len(devices))
	for i, device := range devices {
		fields := make(map[string]*pb.FieldValue)
		for fieldName, value := range device.Values {
			// 序列化值为JSON
			valueJSON, err := json.Marshal(value)
			if err != nil {
				logx.Errorf("Failed to marshal value: %v", err)
				valueJSON = []byte("null")
			}

			fields[string(fieldName)] = &pb.FieldValue{
				ValueJson: string(valueJSON),
				Timestamp: device.Timestamps[fieldName].Unix(),
			}
		}

		pbDevices[i] = &pb.DeviceLatestValues{
			DeviceId: device.DeviceID,
			Fields:   fields,
		}
	}

	return &pb.GetLatestValuesResp{
		Devices: pbDevices,
	}
}

// ConvertDeviceStatisticsToProto 转换core.DeviceStatistics为protobuf响应
func ConvertDeviceStatisticsToProto(stats []core.DeviceStatistics) *pb.GetDeviceStatisticsResp {
	pbDevices := make([]*pb.DeviceStatistics, len(stats))
	for i, stat := range stats {
		fields := make(map[string]*pb.FieldStatistics)
		for fieldName, fieldStat := range stat.FieldStats {
			// 序列化最新值为JSON
			lastValueJSON, err := json.Marshal(fieldStat.LastValue)
			if err != nil {
				logx.Errorf("Failed to marshal last value: %v", err)
				lastValueJSON = []byte("null")
			}

			fields[string(fieldName)] = &pb.FieldStatistics{
				Min:       fieldStat.Min,
				Max:       fieldStat.Max,
				Avg:       fieldStat.Avg,
				Sum:       fieldStat.Sum,
				Count:     fieldStat.Count,
				LastValue: string(lastValueJSON),
				LastTime:  fieldStat.LastTime.Unix(),
			}
		}

		pbDevices[i] = &pb.DeviceStatistics{
			DeviceId:    stat.DeviceID,
			DeviceModel: stat.DeviceModel,
			Fields:      fields,
		}
	}

	return &pb.GetDeviceStatisticsResp{
		Devices: pbDevices,
	}
}
