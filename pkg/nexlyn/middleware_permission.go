package plugin_nexlyn

import (
	"fmt"
	"net/http"

	"github.com/rezeropoint/casbinx/core"
)

// WithPermission 权限检查中间件 - 使用NexlynContext中已有的用户信息
func (n *NexlynPlugin) WithPermission(resource, action string) MiddlewareChain {
	return func(handler NexlynHandlerFunc) NexlynHandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
			// 先确保用户信息存在（JWT中间件和WithAuth应该已经处理了）
			if ctx.UserInfo == nil {
				// 如果NexlynContext中没有用户信息，说明认证流程有问题
				user, err := GetUserFromJWT(r.Context())
				if err != nil {
					n.sendErrorResponse(w, http.StatusUnauthorized, "获取用户信息失败", err)
					return
				}
				ctx.UserInfo = user
			}

			// 检查用户是否活跃
			if !ctx.UserInfo.IsActive() {
				n.sendErrorResponse(w, http.StatusForbidden, "用户账户已被禁用", nil)
				return
			}

			// CheckUserPermission 检查用户权限（使用CasbinX）

			permission := core.Permission{
				Resource: core.Resource(resource),
				Action:   core.Action(action),
			}
			hasPermission, err := n.Casbinx.CheckPermission(ctx.UserInfo.GetUserKey(), ctx.UserInfo.GetTenantKey(), permission)
			if err != nil {
				n.Error("CasbinX权限检查失败", "error", err)
				return
			}

			// 检查权限
			if !hasPermission {
				n.sendErrorResponse(w, http.StatusForbidden, "权限不足", fmt.Errorf("用户 %s 没有 %s:%s 权限", ctx.UserInfo.GetUserKey(), resource, action))
				return
			}

			// 继续处理请求
			handler(w, r, ctx)
		}
	}
}
