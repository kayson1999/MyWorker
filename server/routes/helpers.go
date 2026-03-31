package routes

import (
	"encoding/json"
	"net/http"
)

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// readJSON 读取 JSON 请求体
func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// methodHandler 包装 handler，限制只允许指定的 HTTP 方法
// Go 1.21 的 http.ServeMux 不支持 "METHOD /path" 语法，需要手动检查方法
func methodHandler(method string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "方法不允许"})
			return
		}
		h(w, r)
	}
}

// methodMiddleware 包装 http.Handler，限制只允许指定的 HTTP 方法
func methodMiddleware(method string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "方法不允许"})
			return
		}
		h.ServeHTTP(w, r)
	})
}
