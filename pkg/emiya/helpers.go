package emiya

import "encoding/json"

// analyzingJSON 递归解析嵌套的 JSON 字符串
// 该函数会遍历 map 中的所有值，如果发现字符串类型的值可以解析为 JSON 对象，
// 则将其解析并替换原始字符串值
func analyzingJSON(data map[string]any) error {
	for key, value := range data {
		if str, ok := value.(string); ok {
			// 如果字符串为空，设置为 nil
			if str == "" {
				data[key] = nil
				continue
			}

			// 尝试将字符串解析为 JSON 对象
			var nestedData map[string]any
			if err := json.Unmarshal([]byte(str), &nestedData); err == nil {
				// 递归解析嵌套数据
				if err := analyzingJSON(nestedData); err != nil {
					return err
				}
				// 用解析后的对象替换原始字符串
				data[key] = nestedData
			}
			// 如果解析失败，保持原始字符串值不变
		} else if nestedMap, ok := value.(map[string]any); ok {
			// 如果值本身就是 map，递归处理
			if err := analyzingJSON(nestedMap); err != nil {
				return err
			}
		}
	}
	return nil
}
