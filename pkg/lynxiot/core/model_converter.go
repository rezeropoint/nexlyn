// Package core 定义IoT设备管理的核心领域模型
//
// 文件说明：model_converter.go
// 职责：型号转换器接口定义
package core

// ModelConverter 型号转换器接口
// 用于将统一的命令格式转换为特定型号的MQTT消息格式，以及反向转换响应
type ModelConverter interface {
	// ConvertCommand 将统一命令转换为特定型号的MQTT消息
	// 参数：
	//   - model: 设备型号（如 AI-200, AI-300）
	//   - command: 统一命令对象（如 AIBoxControlCommand）
	// 返回：
	//   - MQTT消息体（JSON字节流）
	//   - 错误信息
	ConvertCommand(model string, command interface{}) ([]byte, error)

	// ConvertResponse 将特定型号的响应转换为统一格式
	// 参数：
	//   - model: 设备型号（如 AI-200, AI-300）
	//   - response: MQTT响应消息（JSON字节流）
	// 返回：
	//   - 统一响应对象（如 AIBoxControlResponse）
	//   - 错误信息
	ConvertResponse(model string, response []byte) (interface{}, error)
}
