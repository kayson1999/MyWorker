package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"myworker/localuser"
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

// ==================== Token 验证缓存 ====================
// 同一个 token 验证通过后缓存用户信息，避免每次请求都调用用户中心
// 缓存有效期 5 分钟，过期后下一次请求会重新验证

const tokenCacheTTL = 5 * time.Minute

// tokenCacheEntry 缓存条目
type tokenCacheEntry struct {
	user     *usercenter.UCUser
	expireAt time.Time
}

// tokenVerifyCache token 验证结果缓存
var tokenVerifyCache = struct {
	sync.RWMutex
	items map[string]*tokenCacheEntry
}{items: make(map[string]*tokenCacheEntry)}

// cacheGet 从缓存获取已验证的用户信息
func cacheGet(token string) *usercenter.UCUser {
	tokenVerifyCache.RLock()
	defer tokenVerifyCache.RUnlock()
	entry, ok := tokenVerifyCache.items[token]
	if !ok || time.Now().After(entry.expireAt) {
		return nil
	}
	return entry.user
}

// cacheSet 将验证结果写入缓存
func cacheSet(token string, user *usercenter.UCUser) {
	tokenVerifyCache.Lock()
	defer tokenVerifyCache.Unlock()
	tokenVerifyCache.items[token] = &tokenCacheEntry{
		user:     user,
		expireAt: time.Now().Add(tokenCacheTTL),
	}
}

// CacheRemove 移除缓存条目（token 失效/登出时调用）
func CacheRemove(token string) {
	tokenVerifyCache.Lock()
	defer tokenVerifyCache.Unlock()
	delete(tokenVerifyCache.items, token)
}

// init 启动后台清理协程，定期清除过期缓存
func init() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			tokenVerifyCache.Lock()
			for k, v := range tokenVerifyCache.items {
				if now.After(v.expireAt) {
					delete(tokenVerifyCache.items, k)
				}
			}
			tokenVerifyCache.Unlock()
		}
	}()
}

// ==================== 认证中间件 ====================

// AuthMiddleware JWT 认证中间件
// 通过调用用户中心 /auth/verify 接口验证 Token
// 验证通过后缓存结果，后续请求直接使用缓存，不再重复调用用户中心
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "未登录，请先登录"})
			return
		}

		token := authHeader[7:]

		// 优先从缓存获取已验证的用户信息
		if cachedUser := cacheGet(token); cachedUser != nil {
			// 缓存命中，直接使用，无需调用用户中心
			localuser.EnsureLocalUserExists(cachedUser)

			ctx := context.WithValue(r.Context(), CtxUserID, cachedUser.ID)
			ctx = context.WithValue(ctx, CtxUCUser, cachedUser)
			ctx = context.WithValue(ctx, CtxToken, token)

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 缓存未命中，调用用户中心验证
		result, err := usercenter.Verify(token)
		if err != nil {
			logger.Error("认证中间件 - 用户中心通信失败: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "认证服务暂时不可用，请稍后重试"})
			return
		}

		if !result.OK {
			if result.Status == 401 {
				// token 已失效，清除缓存
				CacheRemove(token)
				errMsg := "登录已过期，请重新登录"
				if e, ok := result.Data["error"].(string); ok {
					errMsg = e
				}
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": errMsg})
				return
			}
			if result.Status == 403 {
				CacheRemove(token)
				errMsg := "账号已被禁用"
				if e, ok := result.Data["error"].(string); ok {
					errMsg = e
				}
				writeJSON(w, http.StatusForbidden, map[string]string{"error": errMsg})
				return
			}
			// 用户中心返回 429 限流或其他非认证错误，不应返回 401（避免前端误清 token）
			if result.Status == 429 {
				logger.Warn("认证中间件 - 用户中心限流: status=%d", result.Status)
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "服务繁忙，请稍后重试"})
				return
			}
			logger.Warn("认证中间件 - 用户中心返回异常: status=%d", result.Status)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "认证服务异常，请稍后重试"})
			return
		}

		ucUser := usercenter.ParseUCUser(result.Data)
		if ucUser == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "认证失败，用户信息解析错误"})
			return
		}

		// 验证通过，写入缓存
		cacheSet(token, ucUser)

		// 确保本地用户记录存在（SSO 登录等场景下本地可能没有记录）
		localuser.EnsureLocalUserExists(ucUser)

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
