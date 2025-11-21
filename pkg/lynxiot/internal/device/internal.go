package device

import (
	"context"
	"fmt"

	"github.com/rezeropoint/nexlyn/pkg/lynxiot/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// replaceDeviceTags 完全替换设备的标签关联
func (m *deviceManager) replaceDeviceTags(ctx context.Context, deviceID string, tagIDs []string, tenantID string) error {
	// 1. 删除现有关联
	deleteQuery := "DELETE FROM iot_device_tag_relations WHERE device_id = $1"
	_, err := m.dbConn.ExecCtx(ctx, deleteQuery, deviceID)
	if err != nil {
		return fmt.Errorf("删除现有标签关联失败: %w", err)
	}

	// 2. 如果tagIDs为空，则只删除不添加
	if len(tagIDs) == 0 {
		return nil
	}

	// 3. 为每个标签创建新关联
	for _, tagID := range tagIDs {
		// 验证标签是否存在且属于该租户
		var count int
		checkQuery := "SELECT COUNT(*) FROM iot_tag_definitions WHERE id = $1 AND tenant_id = $2"
		err = m.dbConn.QueryRowCtx(ctx, &count, checkQuery, tagID, tenantID)
		if err != nil {
			return fmt.Errorf("查询标签失败: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("标签 %s 不存在或无权限访问", tagID)
		}

		// 创建关联
		insertQuery := "INSERT INTO iot_device_tag_relations (id, device_id, tag_id) VALUES ($1, $2, $3)"
		_, err = m.dbConn.ExecCtx(ctx, insertQuery, uuid.New().String(), deviceID, tagID)
		if err != nil {
			return fmt.Errorf("创建标签关联失败: %w", err)
		}
	}

	return nil
}

// getDeviceTags 获取设备的所有标签
func (m *deviceManager) getDeviceTags(ctx context.Context, deviceID string) ([]core.DeviceTagSummary, error) {
	query := `
		SELECT t.id, t.label, t.description, t.color, t.created_at
		FROM iot_tag_definitions t
		INNER JOIN iot_device_tag_relations r ON t.id = r.tag_id
		WHERE r.device_id = $1
		ORDER BY r.created_at DESC
	`

	var results []deviceTagDB

	err := m.dbConn.QueryRowsCtx(ctx, &results, query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("查询设备标签失败: %w", err)
	}

	tags := make([]core.DeviceTagSummary, 0, len(results))
	for _, r := range results {
		tag := core.DeviceTagSummary{
			ID:    r.ID,
			Label: r.Label,
		}
		if r.Description.Valid {
			tag.Description = r.Description.String
		}
		if r.Color.Valid {
			tag.Color = r.Color.String
		}
		if r.CreatedAt.Valid {
			tag.CreatedAt = r.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// listBoundDeviceIDs 获取指定租户在指定组织范围内所有已绑定的设备ID列表
func (m *deviceManager) listBoundDeviceIDs(ctx context.Context, tenantID string, orgIDs []string) ([]string, error) {
	query := `
		SELECT device_id
		FROM iot_device_bindings
		WHERE tenant_id = $1 AND org_id = ANY($2)
	`

	var results []boundDeviceDB

	err := m.dbConn.QueryRowsCtx(ctx, &results, query, tenantID, pq.Array(orgIDs))
	if err != nil {
		return nil, fmt.Errorf("查询已绑定设备ID失败: %w", err)
	}

	deviceIDs := make([]string, 0, len(results))
	for _, r := range results {
		deviceIDs = append(deviceIDs, r.DeviceID)
	}

	return deviceIDs, nil
}

// listBoundDeviceIDsForTenant 获取指定租户所有已绑定的设备ID列表（不限制组织）
func (m *deviceManager) listBoundDeviceIDsForTenant(ctx context.Context, tenantID string) ([]string, error) {
	query := `
		SELECT device_id
		FROM iot_device_bindings
		WHERE tenant_id = $1
	`

	var results []boundDeviceDB

	err := m.dbConn.QueryRowsCtx(ctx, &results, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("查询租户已绑定设备ID失败: %w", err)
	}

	deviceIDs := make([]string, 0, len(results))
	for _, r := range results {
		deviceIDs = append(deviceIDs, r.DeviceID)
	}

	return deviceIDs, nil
}
