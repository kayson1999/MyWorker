package routes

import (
	"net/http"

	"myworker/db"
	"myworker/localuser"
	"myworker/logger"
	"myworker/middleware"
	"myworker/usercenter"
)

// RegisterUserRoutes 注册用户信息路由
func RegisterUserRoutes(mux *http.ServeMux) {
	mux.Handle("/api/user/profile", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			middleware.AuthMiddleware(http.HandlerFunc(handleGetProfile)).ServeHTTP(w, r)
		case "PUT":
			middleware.AuthMiddleware(http.HandlerFunc(handleUpdateProfile)).ServeHTTP(w, r)
		default:
			w.Header().Set("Allow", "GET, PUT")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "方法不允许"})
		}
	}))
}

// handleGetProfile 获取个人信息
func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	ucUser := middleware.GetUCUser(r)
	userID := middleware.GetUserID(r)

	// 确保本地用户记录存在
	localuser.EnsureLocalUserExists(ucUser)

	// 从本地获取业务字段
	lu := localuser.GetLocalUser(userID)

	profession := localuser.LocalDefaults["profession"]
	position := localuser.LocalDefaults["position"]
	city := localuser.LocalDefaults["city"]
	standardStart := localuser.LocalDefaults["standard_start"]
	standardEnd := localuser.LocalDefaults["standard_end"]
	createdAt := ucUser.CreatedAt

	if lu != nil {
		if lu.Profession != "" {
			profession = lu.Profession
		}
		if lu.Position != "" {
			position = lu.Position
		}
		if lu.City != "" {
			city = lu.City
		}
		if lu.StandardStart != "" {
			standardStart = lu.StandardStart
		}
		if lu.StandardEnd != "" {
			standardEnd = lu.StandardEnd
		}
		if lu.CreatedAt != "" {
			createdAt = lu.CreatedAt
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":             ucUser.ID,
			"username":       ucUser.Username,
			"nickname":       ucUser.Nickname,
			"avatar":         ucUser.Avatar,
			"profession":     profession,
			"position":       position,
			"city":           city,
			"standard_start": standardStart,
			"standard_end":   standardEnd,
			"created_at":     createdAt,
		},
	})
}

// handleUpdateProfile 更新个人信息
func handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	ucUser := middleware.GetUCUser(r)
	userID := middleware.GetUserID(r)
	token := middleware.GetToken(r)

	var body struct {
		Nickname      *string `json:"nickname"`
		Avatar        *string `json:"avatar"`
		Profession    *string `json:"profession"`
		Position      *string `json:"position"`
		City          *string `json:"city"`
		StandardStart *string `json:"standard_start"`
		StandardEnd   *string `json:"standard_end"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	// 1. 基础字段更新到用户中心
	ucUpdates := make(map[string]interface{})
	if body.Nickname != nil {
		ucUpdates["nickname"] = *body.Nickname
	}
	if body.Avatar != nil {
		ucUpdates["avatar"] = *body.Avatar
	}

	currentUCUser := ucUser
	if len(ucUpdates) > 0 {
		result, err := usercenter.UpdateProfile(token, ucUpdates)
		if err != nil {
			logger.Error("更新用户中心信息失败: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "更新基础信息失败"})
			return
		}
		if !result.OK {
			errMsg := "更新基础信息失败"
			if e, ok := result.Data["error"].(string); ok {
				errMsg = e
			}
			writeJSON(w, result.Status, map[string]string{"error": errMsg})
			return
		}
		parsed := usercenter.ParseUCUser(result.Data)
		if parsed != nil {
			currentUCUser = parsed
		}
	}

	// 2. 确保本地用户存在
	localuser.EnsureLocalUserExists(currentUCUser)

	// 3. 业务字段更新到本地
	d := db.GetDB()
	_, err := d.Exec(`
		UPDATE users SET
			nickname = ?,
			avatar = ?,
			profession = COALESCE(?, profession),
			position = COALESCE(?, position),
			city = COALESCE(?, city),
			standard_start = COALESCE(?, standard_start),
			standard_end = COALESCE(?, standard_end)
		WHERE id = ?`,
		currentUCUser.Nickname, currentUCUser.Avatar,
		nilStr(body.Profession), nilStr(body.Position), nilStr(body.City),
		nilStr(body.StandardStart), nilStr(body.StandardEnd),
		userID,
	)
	if err != nil {
		logger.Error("更新本地用户信息失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新用户信息失败"})
		return
	}

	// 获取更新后的本地用户
	lu := localuser.GetLocalUser(userID)

	profession := localuser.LocalDefaults["profession"]
	position := localuser.LocalDefaults["position"]
	city := localuser.LocalDefaults["city"]
	standardStart := localuser.LocalDefaults["standard_start"]
	standardEnd := localuser.LocalDefaults["standard_end"]
	createdAt := currentUCUser.CreatedAt

	if lu != nil {
		if lu.Profession != "" {
			profession = lu.Profession
		}
		if lu.Position != "" {
			position = lu.Position
		}
		if lu.City != "" {
			city = lu.City
		}
		if lu.StandardStart != "" {
			standardStart = lu.StandardStart
		}
		if lu.StandardEnd != "" {
			standardEnd = lu.StandardEnd
		}
		if lu.CreatedAt != "" {
			createdAt = lu.CreatedAt
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":             currentUCUser.ID,
			"username":       currentUCUser.Username,
			"nickname":       currentUCUser.Nickname,
			"avatar":         currentUCUser.Avatar,
			"profession":     profession,
			"position":       position,
			"city":           city,
			"standard_start": standardStart,
			"standard_end":   standardEnd,
			"created_at":     createdAt,
		},
	})
}

// nilStr 将 *string 转为 interface{}（nil 或 string）
func nilStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}
