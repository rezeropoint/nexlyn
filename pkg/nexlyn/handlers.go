package plugin_nexlyn

import (
	"fmt"
	"net/http"
	"strings"
)

// handleTest 测试接口
func (n *NexlynPlugin) handleTest(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("test"))
}

// handlePing 简单测试接口（无认证）
func (n *NexlynPlugin) handlePing(w http.ResponseWriter, r *http.Request, ctx *NexlynContext) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"Nexlyn plugin is working","method":"` + r.Method + `"}`))
}

// sendErrorResponse 发送错误响应的辅助函数
func (n *NexlynPlugin) sendErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"code":    statusCode,
		"message": message,
	}

	if err != nil {
		response["error"] = err.Error()
	}

	fmt.Fprintf(w, `{"code": %d, "message": "%s"`, statusCode, message)
	if err != nil {
		fmt.Fprintf(w, `, "error": "%s"`, strings.ReplaceAll(err.Error(), `"`, `\"`))
	}
	fmt.Fprint(w, `}`)
}
