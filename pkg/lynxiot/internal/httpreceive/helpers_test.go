package httpreceive

import (
	"testing"
)

func TestGetFieldValue(t *testing.T) {
	// 测试数据
	data := map[string]any{
		"name": "test",
		"data": map[string]any{
			"temperature": 25.5,
			"humidity":    60,
		},
		"Result": map[string]any{
			"Tags": []any{"tag1", "tag2", "tag3"},
			"Items": []any{
				map[string]any{"id": 1, "name": "item1"},
				map[string]any{"id": 2, "name": "item2"},
			},
		},
		"matrix": []any{
			[]any{1, 2, 3},
			[]any{4, 5, 6},
		},
	}

	tests := []struct {
		name      string
		fieldPath string
		want      any
		wantExist bool
	}{
		// 简单字段
		{
			name:      "简单字段",
			fieldPath: "name",
			want:      "test",
			wantExist: true,
		},
		// 嵌套对象
		{
			name:      "嵌套对象-温度",
			fieldPath: "data.temperature",
			want:      25.5,
			wantExist: true,
		},
		{
			name:      "嵌套对象-湿度",
			fieldPath: "data.humidity",
			want:      60,
			wantExist: true,
		},
		// 数组索引
		{
			name:      "数组索引-第一个元素",
			fieldPath: "Result.Tags[0]",
			want:      "tag1",
			wantExist: true,
		},
		{
			name:      "数组索引-第二个元素",
			fieldPath: "Result.Tags[1]",
			want:      "tag2",
			wantExist: true,
		},
		{
			name:      "数组索引-第三个元素",
			fieldPath: "Result.Tags[2]",
			want:      "tag3",
			wantExist: true,
		},
		// 数组中的对象
		{
			name:      "数组对象-第一个元素的name",
			fieldPath: "Result.Items[0].name",
			want:      "item1",
			wantExist: true,
		},
		{
			name:      "数组对象-第二个元素的id",
			fieldPath: "Result.Items[1].id",
			want:      2,
			wantExist: true,
		},
		// 二维数组（暂不支持嵌套数组索引，但可以获取数组本身）
		{
			name:      "二维数组-获取第一行",
			fieldPath: "matrix[0]",
			want:      []any{1, 2, 3},
			wantExist: true,
		},
		// 错误情况
		{
			name:      "不存在的字段",
			fieldPath: "notexist",
			want:      nil,
			wantExist: false,
		},
		{
			name:      "数组索引越界",
			fieldPath: "Result.Tags[10]",
			want:      nil,
			wantExist: false,
		},
		{
			name:      "负数索引",
			fieldPath: "Result.Tags[-1]",
			want:      nil,
			wantExist: false,
		},
		{
			name:      "非法索引",
			fieldPath: "Result.Tags[abc]",
			want:      nil,
			wantExist: false,
		},
		{
			name:      "空路径",
			fieldPath: "",
			want:      nil,
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := getFieldValue(data, tt.fieldPath)
			if exists != tt.wantExist {
				t.Errorf("getFieldValue() exists = %v, want %v", exists, tt.wantExist)
				return
			}
			if exists {
				// 对于数组类型，需要特殊比较
				switch want := tt.want.(type) {
				case []any:
					gotSlice, ok := got.([]any)
					if !ok {
						t.Errorf("getFieldValue() got type %T, want []any", got)
						return
					}
					if len(gotSlice) != len(want) {
						t.Errorf("getFieldValue() slice length = %v, want %v", len(gotSlice), len(want))
						return
					}
					for i := range want {
						if gotSlice[i] != want[i] {
							t.Errorf("getFieldValue() slice[%d] = %v, want %v", i, gotSlice[i], want[i])
						}
					}
				default:
					if got != tt.want {
						t.Errorf("getFieldValue() got = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}

func TestParseFieldPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []pathPart
	}{
		{
			name: "简单字段",
			path: "name",
			want: []pathPart{
				{Type: pathTypeField, Value: "name"},
			},
		},
		{
			name: "嵌套对象",
			path: "data.temperature",
			want: []pathPart{
				{Type: pathTypeField, Value: "data"},
				{Type: pathTypeField, Value: "temperature"},
			},
		},
		{
			name: "数组索引",
			path: "Result.Tags[0]",
			want: []pathPart{
				{Type: pathTypeField, Value: "Result"},
				{Type: pathTypeField, Value: "Tags"},
				{Type: pathTypeArrayIndex, Value: "0"},
			},
		},
		{
			name: "混合路径",
			path: "data.items[2].name",
			want: []pathPart{
				{Type: pathTypeField, Value: "data"},
				{Type: pathTypeField, Value: "items"},
				{Type: pathTypeArrayIndex, Value: "2"},
				{Type: pathTypeField, Value: "name"},
			},
		},
		{
			name: "多级数组",
			path: "matrix[0][1]",
			want: []pathPart{
				{Type: pathTypeField, Value: "matrix"},
				{Type: pathTypeArrayIndex, Value: "0"},
				{Type: pathTypeArrayIndex, Value: "1"},
			},
		},
		{
			name: "未闭合的括号",
			path: "data[0",
			want: nil,
		},
		{
			name: "未匹配的右括号",
			path: "data]",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFieldPath(tt.path)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseFieldPath() should return nil for invalid path, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("parseFieldPath() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range tt.want {
				if got[i].Type != tt.want[i].Type || got[i].Value != tt.want[i].Value {
					t.Errorf("parseFieldPath()[%d] = {%v, %v}, want {%v, %v}",
						i, got[i].Type, got[i].Value, tt.want[i].Type, tt.want[i].Value)
				}
			}
		})
	}
}
