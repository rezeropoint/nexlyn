package plugin_nexlyn

import (
	"net/http"
)

// handleDeviceList 获取设备列表
func (n *NexlynPlugin) handleDeviceList(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynGetDeviceList(w, r, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}

// handleDeviceBind 设备绑定
func (n *NexlynPlugin) handleDeviceBind(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynBindDevice(w, r, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}

// handleDeviceDetail 获取设备详情
func (n *NexlynPlugin) handleDeviceDetail(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynGetDeviceDetail(w, r, ctx.DeviceID, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}

// handleDeviceAlias 更新设备别名
func (n *NexlynPlugin) handleDeviceAlias(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynUpdateDeviceAlias(w, r, ctx.DeviceID, ctx.UserInfo.TenantId)
}

// handleDeviceTags 更新设备标签
func (n *NexlynPlugin) handleDeviceTags(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynUpdateDeviceTags(w, r, ctx.DeviceID, ctx.UserInfo.TenantId)
}

// handleDeviceUnbind 设备解绑
func (n *NexlynPlugin) handleDeviceUnbind(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynUnbindDevice(w, r, ctx.DeviceID, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}

// handleDeviceDelete 彻底删除设备
func (n *NexlynPlugin) handleDeviceDelete(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynDeleteDevice(w, r, ctx.DeviceID, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}

// handleTagList 获取标签列表
func (n *NexlynPlugin) handleTagList(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynGetTagList(w, r, ctx.UserInfo.TenantId)
}

// handleTagCreate 创建标签
func (n *NexlynPlugin) handleTagCreate(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynCreateTag(w, r, ctx.UserInfo.TenantId)
}

// handleTagUpdate 更新标签
func (n *NexlynPlugin) handleTagUpdate(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynUpdateTag(w, r, ctx.TagID, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}

// handleTagDelete 删除标签
func (n *NexlynPlugin) handleTagDelete(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynDeleteTag(w, r, ctx.TagID, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}

// handleRecordQuery 查询录像记录
func (n *NexlynPlugin) handleRecordQuery(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynGetRecordList(w, r, ctx.DeviceID, ctx.ChannelID, ctx.UserInfo.TenantId)
}

// handleGetDeviceStatistics 获取设备统计信息
func (n *NexlynPlugin) handleGetDeviceStatistics(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	n.gb28181Plugin.HandleNexlynGetDeviceStatistics(w, r, ctx.UserInfo.TenantId, ctx.UserInfo.UserId)
}
