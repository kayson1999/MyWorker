package localuser

import (
	"database/sql"

	"myworker/db"
	"myworker/usercenter"
)

// LocalDefaults 本地业务字段默认值
var LocalDefaults = map[string]string{
	"profession":     "",
	"position":       "",
	"city":           "",
	"standard_start": "09:00",
	"standard_end":   "18:00",
}

// LocalUser 本地业务用户
type LocalUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Profession    string `json:"profession"`
	Position      string `json:"position"`
	City          string `json:"city"`
	StandardStart string `json:"standard_start"`
	StandardEnd   string `json:"standard_end"`
	CreatedAt     string `json:"created_at"`
}

// MergedUser 合并后的用户信息
type MergedUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Profession    string `json:"profession"`
	Position      string `json:"position"`
	City          string `json:"city"`
	StandardStart string `json:"standard_start"`
	StandardEnd   string `json:"standard_end"`
	CreatedAt     string `json:"created_at"`
}

// EnsureLocalUser 确保本地业务用户存在，返回合并后的用户信息
func EnsureLocalUser(ucUser *usercenter.UCUser) *MergedUser {
	d := db.GetDB()

	var existing LocalUser
	err := d.QueryRow("SELECT id, username, nickname, avatar, profession, position, city, standard_start, standard_end, created_at FROM users WHERE id = ?", ucUser.ID).
		Scan(&existing.ID, &existing.Username, &existing.Nickname, &existing.Avatar,
			&existing.Profession, &existing.Position, &existing.City,
			&existing.StandardStart, &existing.StandardEnd, &existing.CreatedAt)

	if err == sql.ErrNoRows {
		// 首次登录本服务，创建本地业务记录
		_, _ = d.Exec("INSERT INTO users (id, username, nickname, avatar) VALUES (?, ?, ?, ?)",
			ucUser.ID, ucUser.Username, ucUser.Nickname, ucUser.Avatar)

		return &MergedUser{
			ID:            ucUser.ID,
			Username:      ucUser.Username,
			Nickname:      ucUser.Nickname,
			Avatar:        ucUser.Avatar,
			Profession:    LocalDefaults["profession"],
			Position:      LocalDefaults["position"],
			City:          LocalDefaults["city"],
			StandardStart: LocalDefaults["standard_start"],
			StandardEnd:   LocalDefaults["standard_end"],
			CreatedAt:     ucUser.CreatedAt,
		}
	}

	// 同步用户中心的基础信息到本地
	_, _ = d.Exec("UPDATE users SET username = ?, nickname = ?, avatar = ? WHERE id = ?",
		ucUser.Username, ucUser.Nickname, ucUser.Avatar, ucUser.ID)

	return &MergedUser{
		ID:            ucUser.ID,
		Username:      ucUser.Username,
		Nickname:      ucUser.Nickname,
		Avatar:        ucUser.Avatar,
		Profession:    orDefault(existing.Profession, LocalDefaults["profession"]),
		Position:      orDefault(existing.Position, LocalDefaults["position"]),
		City:          orDefault(existing.City, LocalDefaults["city"]),
		StandardStart: orDefault(existing.StandardStart, LocalDefaults["standard_start"]),
		StandardEnd:   orDefault(existing.StandardEnd, LocalDefaults["standard_end"]),
		CreatedAt:     existing.CreatedAt,
	}
}

// GetLocalUser 获取本地业务用户
func GetLocalUser(userID string) *LocalUser {
	d := db.GetDB()
	var u LocalUser
	err := d.QueryRow("SELECT id, profession, position, city, standard_start, standard_end, created_at FROM users WHERE id = ?", userID).
		Scan(&u.ID, &u.Profession, &u.Position, &u.City, &u.StandardStart, &u.StandardEnd, &u.CreatedAt)
	if err != nil {
		return nil
	}
	return &u
}

// EnsureLocalUserExists 确保本地用户记录存在（仅创建，不更新）
func EnsureLocalUserExists(ucUser *usercenter.UCUser) {
	d := db.GetDB()
	_, _ = d.Exec("INSERT OR IGNORE INTO users (id, username, nickname, avatar) VALUES (?, ?, ?, ?)",
		ucUser.ID, ucUser.Username, ucUser.Nickname, ucUser.Avatar)
}

func orDefault(val, defaultVal string) string {
	if val != "" {
		return val
	}
	return defaultVal
}
