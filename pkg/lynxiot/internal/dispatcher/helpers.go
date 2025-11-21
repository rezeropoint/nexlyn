package dispatcher

import (
	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	skylarkCore "github.com/rezeropoint/go-skylark/v2/core"
)

// convertToSkylarkTypedValue 将 core.TypedValue 映射转换为 skylark core.TypedValue 映射
func convertToSkylarkTypedValue(data *map[string]core.TypedValue) map[string]skylarkCore.TypedValue {
	if data == nil {
		return nil
	}
	out := make(map[string]skylarkCore.TypedValue, len(*data))
	for k, v := range *data {
		out[k] = skylarkCore.TypedValue{Type: v.Type, Value: v.Value}
	}
	return out
}
