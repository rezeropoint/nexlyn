package dispatcher

import (
	"context"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/rezeropoint/go-skylark/v2/engine"
)

type Manager interface {
	DispatchInfo(ctx context.Context, configs []core.DispatchConfig, data *map[string]core.TypedValue, taskInfo core.TaskInfo) error
}

func NewManager(config Config, getPlatformConfig core.GetPlatformConfigFunc, skylarkEngine engine.SkylarkEngine, submitAsync core.SubmitAsyncFunc) Manager {
	return newManager(config, getPlatformConfig, skylarkEngine, submitAsync)
}
