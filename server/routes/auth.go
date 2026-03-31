package routes

import (
	"net/http"

	"myworker/localuser"
	"myworker/logger"
	"myworker/middleware"
	"myworker/usercenter"
)

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/register", methodHandler("POST", handleRegister))
	mux.HandleFunc("/api/auth/login", methodHandler("POST", handleLogin))
	mux.Handle("/api/auth/logout", methodMiddleware("POST", middleware.AuthMiddleware(http.HandlerFunc(handleLogout))))
}

// handleRegister 注册
func handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.Username == "" || body.Password == "" || body.Nickname == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "用户名、密码和昵称不能为空"})
		return
	}

	result, err := usercenter.Register(map[string]string{
		"username": body.Username,
		"password": body.Password,
		"nickname": body.Nickname,
	})
	if err != nil {
		logger.Error("注册失败: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "认证服务暂时不可用，请稍后重试"})
		return
	}

	if !result.OK {
		errMsg := "注册失败"
		if e, ok := result.Data["error"].(string); ok {
			errMsg = e
		}
		writeJSON(w, result.Status, map[string]string{"error": errMsg})
		return
	}

	ucUser := usercenter.ParseUCUser(result.Data)
	if ucUser == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "注册成功但用户信息解析失败"})
		return
	}

	mergedUser := localuser.EnsureLocalUser(ucUser)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": result.Data["token"],
		"user":  mergedUser,
	})
}

// handleLogin 登录
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.Username == "" || body.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "用户名和密码不能为空"})
		return
	}

	result, err := usercenter.Login(map[string]string{
		"username": body.Username,
		"password": body.Password,
	})
	if err != nil {
		logger.Error("登录失败: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "认证服务暂时不可用，请稍后重试"})
		return
	}

	if !result.OK {
		errMsg := "登录失败"
		if e, ok := result.Data["error"].(string); ok {
			errMsg = e
		}
		writeJSON(w, result.Status, map[string]string{"error": errMsg})
		return
	}

	ucUser := usercenter.ParseUCUser(result.Data)
	if ucUser == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "登录成功但用户信息解析失败"})
		return
	}

	mergedUser := localuser.EnsureLocalUser(ucUser)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": result.Data["token"],
		"user":  mergedUser,
	})
}

// handleLogout 登出
func handleLogout(w http.ResponseWriter, r *http.Request) {
	token := middleware.GetToken(r)

	result, err := usercenter.Logout(token)
	if err != nil {
		logger.Error("登出失败: %v", err)
	} else if !result.OK {
		errMsg := ""
		if e, ok := result.Data["error"].(string); ok {
			errMsg = e
		}
		logger.Warn("用户中心登出失败: %s", errMsg)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "已成功登出",
	})
}
