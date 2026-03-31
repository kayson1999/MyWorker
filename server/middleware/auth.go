package middleware

import (
	"context"
	"net/http"
	"strings"

	"myworker/logger"
	"myworker/usercenter"
)

// 上下文 key 类型
type contextKey string

const (
	// CtxUserID 用户 ID
	CtxUserID contextKey = "userId"
	// CtxUCUser 用户中心用户信息
	CtxUCUser contextKey = "ucUser"
	// CtxToken 原始 Token
	CtxToken contextKey = "token"
)

// AuthMiddleware JWT 认证中间件
// 通过调用用户中心 /auth/verify 接口验证 Token
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "未登录，请先登录"})
			return
		}

		token := authHeader[7:]
		result, err := usercenter.Verify(token)
		if err != nil {
			logger.Error("认证中间件 - 用户中心通信失败: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "认证服务暂时不可用，请稍后重试"})
			return
		}

		if !result.OK {
			if result.Status == 401 {
				errMsg := "登录已过期，请重新登录"
				if e, ok := result.Data["error"].(string); ok {
					errMsg = e
				}
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": errMsg})
				return
			}
			if result.Status == 403 {
				errMsg := "账号已被禁用"
				if e, ok := result.Data["error"].(string); ok {
					errMsg = e
				}
				writeJSON(w, http.StatusForbidden, map[string]string{"error": errMsg})
				return
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "认证失败，请重新登录"})
			return
		}

		ucUser := usercenter.ParseUCUser(result.Data)
		if ucUser == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "认证失败，用户信息解析错误"})
			return
		}

		// 将用户信息存入 context
		ctx := context.WithValue(r.Context(), CtxUserID, ucUser.ID)
		ctx = context.WithValue(ctx, CtxUCUser, ucUser)
		ctx = context.WithValue(ctx, CtxToken, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID 从 context 获取用户 ID
func GetUserID(r *http.Request) string {
	if id, ok := r.Context().Value(CtxUserID).(string); ok {
		return id
	}
	return ""
}

// GetUCUser 从 context 获取用户中心用户信息
func GetUCUser(r *http.Request) *usercenter.UCUser {
	if user, ok := r.Context().Value(CtxUCUser).(*usercenter.UCUser); ok {
		return user
	}
	return nil
}

// GetToken 从 context 获取 Token
func GetToken(r *http.Request) string {
	if token, ok := r.Context().Value(CtxToken).(string); ok {
		return token
	}
	return ""
}
