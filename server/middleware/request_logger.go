package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"myworker/logger"
)

// 需要脱敏的请求头关键词（不区分大小写匹配）
var sensitiveHeaders = []string{
	"authorization",
	"sign",
	"secret",
	"token",
	"password",
}

// isSensitiveHeader 判断请求头是否需要脱敏
func isSensitiveHeader(name string) bool {
	lower := strings.ToLower(name)
	for _, keyword := range sensitiveHeaders {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

// responseRecorder 用于捕获响应状态码和响应体
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b) // 同时写入缓冲区
	return r.ResponseWriter.Write(b)
}

// getClientIP 获取客户端真实 IP
func getClientIP(r *http.Request) string {
	// 优先从 X-Forwarded-For 获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个 IP（最原始的客户端 IP）
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	// 其次从 X-Real-IP 获取
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// 最后使用 RemoteAddr
	return r.RemoteAddr
}

// formatHeaders 格式化请求头，敏感信息脱敏
func formatHeaders(header http.Header) string {
	// 按 key 排序，保证输出稳定
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		val := strings.Join(header[k], ", ")
		if isSensitiveHeader(k) {
			val = "***"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, val))
	}
	return strings.Join(parts, "; ")
}

// statusEmoji 根据状态码返回对应的 emoji
func statusEmoji(code int) string {
	if code >= 200 && code < 300 {
		return "✅"
	} else if code >= 300 && code < 400 {
		return "↗️"
	} else if code >= 400 && code < 500 {
		return "⚠️"
	}
	return "❌"
}

// truncateBody 截断过长的响应体，避免日志过大
func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "...(已截断)"
}

// RequestLogger 请求日志中间件
// 记录每个 API 请求的详细信息，包括请求方法、路径、来源IP、请求头、响应状态码、响应体和耗时
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只记录 API 请求的日志，静态资源不记录
		if !strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		// 包装 ResponseWriter 以捕获响应
		rec := newResponseRecorder(w)

		// 执行下一个处理器
		next.ServeHTTP(rec, r)

		// 计算耗时
		elapsed := time.Since(start)

		// 获取请求信息
		clientIP := getClientIP(r)
		headers := formatHeaders(r.Header)
		statusCode := rec.statusCode
		emoji := statusEmoji(statusCode)
		respBody := truncateBody(rec.body.String(), 1024)

		// 按用户要求的样式输出日志
		logMsg := fmt.Sprintf("\n  ▶ 请求: %s %s\n  ▶ 来源IP: %s\n  ▶ 请求头: %s\n  ◀ 状态码: %d %s\n  ◀ 响应体: %s\n  ⏱ 耗时: %v",
			r.Method, r.URL.RequestURI(),
			clientIP,
			headers,
			statusCode, emoji,
			respBody,
			elapsed,
		)

		// 根据状态码选择日志级别
		if statusCode >= 500 {
			logger.Error(logMsg)
		} else if statusCode >= 400 {
			logger.Warn(logMsg)
		} else {
			logger.Info(logMsg)
		}
	})
}
