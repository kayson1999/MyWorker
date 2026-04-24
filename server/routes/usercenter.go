package routes

import (
	"database/sql"
	"math"
	"net/http"
	"time"

	"myworker/db"
	"myworker/middleware"
	"myworker/utils"
)

// RegisterUserCenterRoutes 注册个人中心路由
func RegisterUserCenterRoutes(mux *http.ServeMux) {
	auth := middleware.AuthMiddleware

	mux.Handle("/api/usercenter/overview", methodMiddleware("GET", auth(http.HandlerFunc(handleUserCenterOverview))))
	mux.Handle("/api/usercenter/achievements", methodMiddleware("GET", auth(http.HandlerFunc(handleUserAchievements))))
	mux.Handle("/api/usercenter/exp-logs", methodMiddleware("GET", auth(http.HandlerFunc(handleExpLogs))))
	mux.Handle("/api/usercenter/heatmap", methodMiddleware("GET", auth(http.HandlerFunc(handleHeatmap))))
}

// handleUserCenterOverview 个人中心数据总览
func handleUserCenterOverview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	// 1. 等级信息
	levelInfo := utils.GetUserLevelInfo(d, userID)

	// 2. 工作风格标签（周/月/年）
	var weekTitle, monthTitle, yearTitle string
	d.QueryRow("SELECT COALESCE(week_title,''), COALESCE(month_title,''), COALESCE(year_title,'') FROM users WHERE id = ?", userID).
		Scan(&weekTitle, &monthTitle, &yearTitle)

	// 3. 核心数据统计
	// 总打卡天数
	var totalDays int
	d.QueryRow("SELECT COUNT(*) FROM clock_records WHERE user_id = ?", userID).Scan(&totalDays)

	// 总工时
	var totalHours float64
	d.QueryRow("SELECT COALESCE(SUM(duration), 0) FROM clock_records WHERE user_id = ? AND duration > 0", userID).Scan(&totalHours)
	totalHours = math.Round(totalHours*100) / 100

	// 日均工时
	var workedDays int
	d.QueryRow("SELECT COUNT(*) FROM clock_records WHERE user_id = ? AND duration > 0", userID).Scan(&workedDays)
	avgHours := 0.0
	if workedDays > 0 {
		avgHours = math.Round(totalHours/float64(workedDays)*100) / 100
	}

	// 连续打卡天数
	streak := CalcClockinStreak(d, userID)

	// 成就统计
	var unlockedCount int
	d.QueryRow("SELECT COUNT(*) FROM user_achievements WHERE user_id = ?", userID).Scan(&unlockedCount)
	totalAchievements := len(utils.Achievements)

	// 4. 今日状态
	todayDate := time.Now().Format("2006-01-02")
	todayRecord := getRecordByUserDate(userID, todayDate)
	todayStatus := "未打卡"
	if todayRecord != nil {
		if todayRecord.ClockOut != nil {
			todayStatus = "已完成"
		} else if todayRecord.ClockIn != nil {
			todayStatus = "工作中"
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"level_info": levelInfo,
		"style_tags": map[string]string{
			"week":  weekTitle,
			"month": monthTitle,
			"year":  yearTitle,
		},
		"stats": map[string]interface{}{
			"total_days":  totalDays,
			"total_hours": totalHours,
			"avg_hours":   avgHours,
			"streak":      streak,
		},
		"achievements": map[string]interface{}{
			"unlocked": unlockedCount,
			"total":    totalAchievements,
		},
		"today_status": todayStatus,
	})
}

// calcStreakFromDates 通用连续天数计算（公共函数）
// 从日期集合中计算连续天数：今日已有记录从今天开始向前追溯，否则从昨天开始
func calcStreakFromDates(dateSet map[string]bool) int {
	now := time.Now()
	todayStr := now.Format("2006-01-02")
	streak := 0
	checkDate := now
	if !dateSet[todayStr] {
		checkDate = now.AddDate(0, 0, -1)
	}
	for {
		ds := checkDate.Format("2006-01-02")
		if dateSet[ds] {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}
	return streak
}

// CalcClockinStreak 计算用户工时打卡连续天数（公共函数）
// 逻辑：今日已打卡从今天开始向前追溯，今日未打卡从昨天开始向前追溯
func CalcClockinStreak(d *sql.DB, userID string) int {
	rows, err := d.Query("SELECT date FROM clock_records WHERE user_id = ? ORDER BY date DESC", userID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	dateSet := make(map[string]bool)
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err == nil {
			dateSet[date] = true
		}
	}

	return calcStreakFromDates(dateSet)
}

// handleUserAchievements 获取用户成就列表
func handleUserAchievements(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	achievements := utils.GetUserAchievements(d, userID)
	if achievements == nil {
		achievements = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"achievements": achievements,
	})
}

// handleExpLogs 获取经验值日志
func handleExpLogs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	logs := utils.GetExpLogs(d, userID, 30)
	if logs == nil {
		logs = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs": logs,
	})
}

// handleHeatmap 获取打卡热力图数据（过去一年）
func handleHeatmap(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	// 获取过去一年的打卡记录
	oneYearAgo := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	rows, err := d.Query(
		"SELECT date, duration FROM clock_records WHERE user_id = ? AND date >= ? AND date <= ? ORDER BY date ASC",
		userID, oneYearAgo, today,
	)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"heatmap": []interface{}{}})
		return
	}
	defer rows.Close()

	var heatmap []map[string]interface{}
	for rows.Next() {
		var date string
		var duration float64
		if err := rows.Scan(&date, &duration); err == nil {
			// 热力等级：0=无, 1=<4h, 2=4-8h, 3=8-10h, 4=>10h
			level := 0
			if duration > 0 && duration < 4 {
				level = 1
			} else if duration >= 4 && duration < 8 {
				level = 2
			} else if duration >= 8 && duration < 10 {
				level = 3
			} else if duration >= 10 {
				level = 4
			}
			heatmap = append(heatmap, map[string]interface{}{
				"date":     date,
				"duration": math.Round(duration*100) / 100,
				"level":    level,
			})
		}
	}

	if heatmap == nil {
		heatmap = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"heatmap": heatmap,
	})
}
