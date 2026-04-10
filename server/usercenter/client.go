package usercenter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"myworker/config"
	"myworker/logger"
)

// UCResponse 用户中心响应
type UCResponse struct {
	OK     bool
	Status int
	Data   map[string]interface{}
}

// UCUser 用户中心用户信息
type UCUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"created_at"`
}

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          20,
		MaxConnsPerHost:       20,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       30 * time.Second,
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     false,
		ResponseHeaderTimeout: 5 * time.Second,
	},
}

// ucRequest 向用户中心发起请求
func ucRequest(path string, method string, body interface{}, bearerToken string) (*UCResponse, error) {
	cfg := config.C
	url := fmt.Sprintf("%s%s", cfg.UserCenterURL, path)

	// 为每个请求设置独立的超时上下文，防止连接池耗尽时请求无限等待
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Id", cfg.AppID)
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求用户中心失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &UCResponse{
		OK:     resp.StatusCode >= 200 && resp.StatusCode < 300,
		Status: resp.StatusCode,
		Data:   data,
	}, nil
}

// Register 注册
func Register(body map[string]string) (*UCResponse, error) {
	return ucRequest("/auth/register", "POST", body, "")
}

// Login 登录
func Login(body map[string]string) (*UCResponse, error) {
	return ucRequest("/auth/login", "POST", body, "")
}

// Verify 验证 Token
func Verify(token string) (*UCResponse, error) {
	return ucRequest("/auth/verify", "GET", nil, token)
}

// Logout 登出
func Logout(token string) (*UCResponse, error) {
	return ucRequest("/auth/logout", "POST", nil, token)
}

// UpdateProfile 更新用户 Profile
func UpdateProfile(token string, body map[string]interface{}) (*UCResponse, error) {
	return ucRequest("/user/profile", "PUT", body, token)
}

// ParseUCUser 从响应数据中解析用户信息
func ParseUCUser(data map[string]interface{}) *UCUser {
	userMap, ok := data["user"].(map[string]interface{})
	if !ok {
		logger.Warn("解析用户中心用户信息失败")
		return nil
	}

	user := &UCUser{}

	// id 可能是字符串或 float64
	if id, ok := userMap["id"].(string); ok {
		user.ID = id
	} else if id, ok := userMap["id"].(float64); ok {
		user.ID = fmt.Sprintf("%.0f", id)
	}
	if username, ok := userMap["username"].(string); ok {
		user.Username = username
	}
	if nickname, ok := userMap["nickname"].(string); ok {
		user.Nickname = nickname
	}
	if avatar, ok := userMap["avatar"].(string); ok {
		user.Avatar = avatar
	}
	if createdAt, ok := userMap["created_at"].(string); ok {
		user.CreatedAt = createdAt
	}

	return user
}
