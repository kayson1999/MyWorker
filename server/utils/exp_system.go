package utils

import (
	"database/sql"
	"fmt"

	"myworker/logger"
)

// ==================== 等级系统配置 ====================

// LevelConfig 等级配置
type LevelConfig struct {
	Level  int    `json:"level"`
	Title  string `json:"title"`
	MinExp int    `json:"min_exp"`
	Icon   string `json:"icon"`
}

// 等级配置表（15级）
var LevelTable = []LevelConfig{
	{1, "实习打工人", 0, "🌱"},
	{2, "初级打工人", 50, "🌿"},
	{3, "正式打工人", 150, "🍀"},
	{4, "资深打工人", 350, "🌳"},
	{5, "高级打工人", 650, "⭐"},
	{6, "精英打工人", 1100, "🌟"},
	{7, "骨干打工人", 1700, "💫"},
	{8, "专家打工人", 2500, "🔥"},
	{9, "大师打工人", 3500, "💎"},
	{10, "首席打工人", 5000, "👑"},
	{11, "传说打工人", 7000, "🏆"},
	{12, "神话打工人", 10000, "⚡"},
	{13, "不朽打工人", 15000, "🌈"},
	{14, "超越打工人", 22000, "🚀"},
	{15, "终极打工人", 30000, "🎆"},
}

// ==================== 经验值获取规则 ====================

// ExpRule 经验值规则
type ExpRule struct {
	SourceType string
	Reason     string
	Amount     int
}

// 经验值规则表
var ExpRules = map[string]ExpRule{
	// 打卡相关
	"clockin_in":        {SourceType: "clockin", Reason: "上班打卡", Amount: 5},
	"clockin_out":       {SourceType: "clockin", Reason: "下班打卡", Amount: 5},
	"clockin_full":      {SourceType: "clockin", Reason: "完成全天打卡", Amount: 10},
	"clockin_early":     {SourceType: "clockin", Reason: "早起打卡(8:30前)", Amount: 5},
	"clockin_overtime":  {SourceType: "clockin", Reason: "加班(工时>=10h)", Amount: 10},
	"clockin_streak_7":  {SourceType: "clockin", Reason: "连续打卡7天", Amount: 30},
	"clockin_streak_30": {SourceType: "clockin", Reason: "连续打卡30天", Amount: 100},

	// 计划相关
	"plan_create":   {SourceType: "plan", Reason: "创建计划", Amount: 5},
	"plan_checkin":  {SourceType: "plan", Reason: "计划签到", Amount: 10},
	"plan_complete": {SourceType: "plan", Reason: "完成计划目标", Amount: 20},
}

// ==================== 成就系统配置 ====================

// Achievement 成就定义
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
}

// 成就列表
var Achievements = []Achievement{
	// 打卡里程碑
	{ID: "first_clock", Name: "初来乍到", Description: "完成第一次打卡", Icon: "🎉", Category: "clockin"},
	{ID: "clock_7", Name: "一周坚持", Description: "累计打卡7天", Icon: "📅", Category: "clockin"},
	{ID: "clock_30", Name: "月度达人", Description: "累计打卡30天", Icon: "🗓️", Category: "clockin"},
	{ID: "clock_100", Name: "百日坚持", Description: "累计打卡100天", Icon: "💯", Category: "clockin"},
	{ID: "clock_365", Name: "全年无休", Description: "累计打卡365天", Icon: "🏅", Category: "clockin"},

	// 连续打卡
	{ID: "streak_7", Name: "连续七天", Description: "连续打卡7天", Icon: "🔥", Category: "streak"},
	{ID: "streak_30", Name: "铁人意志", Description: "连续打卡30天", Icon: "💪", Category: "streak"},
	{ID: "streak_100", Name: "百日不断", Description: "连续打卡100天", Icon: "⚡", Category: "streak"},

	// 工时成就
	{ID: "hours_100", Name: "百小时", Description: "累计工时达100小时", Icon: "⏰", Category: "hours"},
	{ID: "hours_500", Name: "五百小时", Description: "累计工时达500小时", Icon: "⏱️", Category: "hours"},
	{ID: "hours_1000", Name: "千小时", Description: "累计工时达1000小时", Icon: "🕐", Category: "hours"},

	// 计划成就
	{ID: "first_plan", Name: "计划先行", Description: "创建第一个计划", Icon: "🎯", Category: "plan"},
	{ID: "plan_10", Name: "计划达人", Description: "累计签到10次", Icon: "📋", Category: "plan"},
	{ID: "plan_50", Name: "计划大师", Description: "累计签到50次", Icon: "📊", Category: "plan"},

	// 等级成就
	{ID: "level_5", Name: "初露锋芒", Description: "达到5级", Icon: "⭐", Category: "level"},
	{ID: "level_10", Name: "登峰造极", Description: "达到10级", Icon: "👑", Category: "level"},
	{ID: "level_15", Name: "终极打工人", Description: "达到15级", Icon: "🎆", Category: "level"},
}

// ==================== 核心函数 ====================

// AddExp 增加经验值（统一入口）
// 返回：是否升级、新等级、获得的经验值
func AddExp(d *sql.DB, userID string, ruleKey string, sourceID string) (leveledUp bool, newLevel int, amount int) {
	rule, ok := ExpRules[ruleKey]
	if !ok {
		logger.Error("未知的经验值规则: %s", ruleKey)
		return false, 0, 0
	}

	// 防重复：使用 INSERT ... ON CONFLICT 原子操作，避免 TOCTOU 竞态
	if sourceID != "" {
		result, insertErr := d.Exec(
			"INSERT INTO exp_logs (user_id, amount, reason, source_type, source_id) VALUES (?, ?, ?, ?, ?) ON CONFLICT(user_id, source_type, source_id) DO NOTHING",
			userID, rule.Amount, rule.Reason, rule.SourceType, sourceID)
		if insertErr != nil {
			logger.Error("记录经验值日志失败: %v", insertErr)
			return false, 0, 0
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return false, 0, 0 // 已发放过，跳过
		}
	} else {
		// 无 sourceID 时直接插入（不做防重复）
		_, insertErr := d.Exec("INSERT INTO exp_logs (user_id, amount, reason, source_type, source_id) VALUES (?, ?, ?, ?, ?)",
			userID, rule.Amount, rule.Reason, rule.SourceType, sourceID)
		if insertErr != nil {
			logger.Error("记录经验值日志失败: %v", insertErr)
			return false, 0, 0
		}
	}

	// 更新用户经验值
	_, err := d.Exec("UPDATE users SET exp = exp + ?, total_exp = total_exp + ? WHERE id = ?",
		rule.Amount, rule.Amount, userID)
	if err != nil {
		logger.Error("更新用户经验值失败: %v", err)
		return false, 0, 0
	}

	// 检查是否升级
	var currentExp, currentLevel int
	err = d.QueryRow("SELECT exp, level FROM users WHERE id = ?", userID).Scan(&currentExp, &currentLevel)
	if err != nil {
		return false, 0, rule.Amount
	}

	newLevel = CalculateLevel(currentExp)
	if newLevel > currentLevel {
		_, err = d.Exec("UPDATE users SET level = ? WHERE id = ?", newLevel, userID)
		if err != nil {
			logger.Error("更新用户等级失败: %v", err)
		}
		// 检查等级成就
		CheckLevelAchievements(d, userID, newLevel)
		return true, newLevel, rule.Amount
	}

	return false, currentLevel, rule.Amount
}

// CalculateLevel 根据经验值计算等级
func CalculateLevel(exp int) int {
	level := 1
	for _, lc := range LevelTable {
		if exp >= lc.MinExp {
			level = lc.Level
		} else {
			break
		}
	}
	return level
}

// GetLevelConfig 获取等级配置
func GetLevelConfig(level int) LevelConfig {
	if level < 1 {
		level = 1
	}
	if level > len(LevelTable) {
		level = len(LevelTable)
	}
	return LevelTable[level-1]
}

// GetNextLevelExp 获取下一级所需经验值，返回(当前级所需, 下一级所需)
func GetNextLevelExp(level int) (int, int) {
	if level < 1 {
		level = 1
	}
	currentMin := LevelTable[level-1].MinExp
	if level >= len(LevelTable) {
		return currentMin, currentMin // 已满级
	}
	return currentMin, LevelTable[level].MinExp
}

// ==================== 成就检查 ====================

// UnlockAchievement 解锁成就
func UnlockAchievement(d *sql.DB, userID string, achievementID string) bool {
	_, err := d.Exec("INSERT OR IGNORE INTO user_achievements (user_id, achievement_id) VALUES (?, ?)",
		userID, achievementID)
	if err != nil {
		logger.Error("解锁成就失败: %v", err)
		return false
	}
	return true
}

// HasAchievement 检查是否已解锁成就
func HasAchievement(d *sql.DB, userID string, achievementID string) bool {
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM user_achievements WHERE user_id = ? AND achievement_id = ?",
		userID, achievementID).Scan(&count)
	return err == nil && count > 0
}

// GetUserAchievements 获取用户已解锁的成就列表
func GetUserAchievements(d *sql.DB, userID string) []map[string]interface{} {
	rows, err := d.Query("SELECT achievement_id, unlocked_at FROM user_achievements WHERE user_id = ? ORDER BY unlocked_at DESC", userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	// 构建已解锁成就的map
	unlockedMap := make(map[string]string)
	for rows.Next() {
		var aid, unlockedAt string
		if err := rows.Scan(&aid, &unlockedAt); err == nil {
			unlockedMap[aid] = unlockedAt
		}
	}

	// 构建完整成就列表（含未解锁的）
	var result []map[string]interface{}
	for _, a := range Achievements {
		item := map[string]interface{}{
			"id":          a.ID,
			"name":        a.Name,
			"description": a.Description,
			"icon":        a.Icon,
			"category":    a.Category,
			"unlocked":    false,
		}
		if unlockedAt, ok := unlockedMap[a.ID]; ok {
			item["unlocked"] = true
			item["unlocked_at"] = unlockedAt
		}
		result = append(result, item)
	}
	return result
}

// CheckClockInAchievements 检查打卡相关成就
func CheckClockInAchievements(d *sql.DB, userID string, totalDays int, streak int, totalHours float64) {
	// 打卡天数成就
	if totalDays >= 1 {
		UnlockAchievement(d, userID, "first_clock")
	}
	if totalDays >= 7 {
		UnlockAchievement(d, userID, "clock_7")
	}
	if totalDays >= 30 {
		UnlockAchievement(d, userID, "clock_30")
	}
	if totalDays >= 100 {
		UnlockAchievement(d, userID, "clock_100")
	}
	if totalDays >= 365 {
		UnlockAchievement(d, userID, "clock_365")
	}

	// 连续打卡成就
	if streak >= 7 {
		UnlockAchievement(d, userID, "streak_7")
	}
	if streak >= 30 {
		UnlockAchievement(d, userID, "streak_30")
	}
	if streak >= 100 {
		UnlockAchievement(d, userID, "streak_100")
	}

	// 工时成就
	if totalHours >= 100 {
		UnlockAchievement(d, userID, "hours_100")
	}
	if totalHours >= 500 {
		UnlockAchievement(d, userID, "hours_500")
	}
	if totalHours >= 1000 {
		UnlockAchievement(d, userID, "hours_1000")
	}
}

// CheckPlanAchievements 检查计划相关成就
func CheckPlanAchievements(d *sql.DB, userID string, totalPlans int, totalCheckins int) {
	if totalPlans >= 1 {
		UnlockAchievement(d, userID, "first_plan")
	}
	if totalCheckins >= 10 {
		UnlockAchievement(d, userID, "plan_10")
	}
	if totalCheckins >= 50 {
		UnlockAchievement(d, userID, "plan_50")
	}
}

// CheckLevelAchievements 检查等级相关成就
func CheckLevelAchievements(d *sql.DB, userID string, level int) {
	if level >= 5 {
		UnlockAchievement(d, userID, "level_5")
	}
	if level >= 10 {
		UnlockAchievement(d, userID, "level_10")
	}
	if level >= 15 {
		UnlockAchievement(d, userID, "level_15")
	}
}

// GetExpLogs 获取用户最近的经验值日志
func GetExpLogs(d *sql.DB, userID string, limit int) []map[string]interface{} {
	if limit <= 0 {
		limit = 20
	}
	rows, err := d.Query("SELECT amount, reason, source_type, created_at FROM exp_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ?",
		userID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var amount int
		var reason, sourceType, createdAt string
		if err := rows.Scan(&amount, &reason, &sourceType, &createdAt); err == nil {
			logs = append(logs, map[string]interface{}{
				"amount":      amount,
				"reason":      reason,
				"source_type": sourceType,
				"created_at":  createdAt,
			})
		}
	}
	return logs
}

// GetUserLevelInfo 获取用户等级信息
func GetUserLevelInfo(d *sql.DB, userID string) map[string]interface{} {
	var level, exp, totalExp int
	err := d.QueryRow("SELECT level, exp, total_exp FROM users WHERE id = ?", userID).Scan(&level, &exp, &totalExp)
	if err != nil {
		return map[string]interface{}{
			"level": 1, "exp": 0, "total_exp": 0,
			"title": "实习打工人", "icon": "🌱",
			"current_min": 0, "next_min": 50, "progress": 0,
		}
	}

	config := GetLevelConfig(level)
	currentMin, nextMin := GetNextLevelExp(level)

	// 计算进度百分比
	progress := 0.0
	if nextMin > currentMin {
		progress = float64(exp-currentMin) / float64(nextMin-currentMin) * 100
		if progress > 100 {
			progress = 100
		}
		if progress < 0 {
			progress = 0
		}
	} else {
		progress = 100 // 满级
	}

	return map[string]interface{}{
		"level":       level,
		"exp":         exp,
		"total_exp":   totalExp,
		"title":       config.Title,
		"icon":        config.Icon,
		"current_min": currentMin,
		"next_min":    nextMin,
		"progress":    fmt.Sprintf("%.1f", progress),
		"max_level":   level >= len(LevelTable),
	}
}
