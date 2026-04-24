package routes

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"myworker/db"
	"myworker/logger"
	"myworker/middleware"
)

// RankItem 排行榜项
type RankItem struct {
	Rank       int             `json:"rank"`
	UserID     string          `json:"userId"`
	Nickname   string          `json:"nickname"`
	Avatar     string          `json:"avatar"`
	Profession string          `json:"profession"`
	City       string          `json:"city"`
	Value      float64         `json:"value"`
	Label      string          `json:"label"`
	Plans      []RankPlanBrief `json:"plans,omitempty"` // 用户的公开计划摘要（仅计划排行榜使用）
}

// RankPlanBrief 排行榜中展示的计划摘要
type RankPlanBrief struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Icon    string `json:"icon"`
}

// TitleRankingItem 称号排行榜项
type TitleRankingItem struct {
	UserID     string `json:"userId"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Profession string `json:"profession"`
	City       string `json:"city"`
	Title      string `json:"title"`
	Rank       int    `json:"rank"`
	Label      string `json:"label"`
}

// finalizeRankList 对排行榜列表进行排序、截断和设置排名
func finalizeRankList(list []RankItem, descending bool) []RankItem {
	if descending {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Value > list[j].Value
		})
	} else {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Value < list[j].Value
		})
	}
	if len(list) > 50 {
		list = list[:50]
	}
	for i := range list {
		list[i].Rank = i + 1
	}
	if list == nil {
		list = []RankItem{}
	}
	return list
}

// getPeriodParam 从请求中获取period参数，默认为"week"
func getPeriodParam(r *http.Request) string {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
	return period
}

// userBasicInfo 用户基本信息（排行榜查询用）
type userBasicInfo struct {
	ID         string
	Nickname   string
	Avatar     string
	Profession string
	City       string
}

// getAllUsers 获取所有用户基本信息
func getAllUsers() ([]userBasicInfo, error) {
	d := db.GetDB()
	rows, err := d.Query("SELECT id, nickname, avatar, profession, city FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []userBasicInfo
	for rows.Next() {
		var u userBasicInfo
		if err := rows.Scan(&u.ID, &u.Nickname, &u.Avatar, &u.Profession, &u.City); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

// RegisterRankingRoutes 注册排行榜路由
func RegisterRankingRoutes(mux *http.ServeMux) {
	auth := middleware.AuthMiddleware

	mux.Handle("/api/ranking/workhours", methodMiddleware("GET", auth(http.HandlerFunc(handleWorkhoursRanking))))
	mux.Handle("/api/ranking/avgworkhours", methodMiddleware("GET", auth(http.HandlerFunc(handleAvgWorkhoursRanking))))
	mux.Handle("/api/ranking/early", methodMiddleware("GET", auth(http.HandlerFunc(handleEarlyRanking))))
	mux.Handle("/api/ranking/late", methodMiddleware("GET", auth(http.HandlerFunc(handleLateRanking))))
	mux.Handle("/api/ranking/ontime", methodMiddleware("GET", auth(http.HandlerFunc(handleOntimeRanking))))
	mux.Handle("/api/ranking/streak", methodMiddleware("GET", auth(http.HandlerFunc(handleStreakRanking))))
	mux.Handle("/api/ranking/titles", methodMiddleware("GET", auth(http.HandlerFunc(handleTitlesRanking))))
}

// ==================== 称号排行榜 ====================

// handleTitlesRanking 处理称号排行榜
func handleTitlesRanking(w http.ResponseWriter, r *http.Request) {
	period := getPeriodParam(r)
	d := db.GetDB()

	// 根据周期选择称号字段
	var titleField string
	switch period {
	case "week":
		titleField = "week_title"
	case "month":
		titleField = "month_title"
	case "year":
		titleField = "year_title"
	default:
		titleField = "week_title"
	}

	// 获取所有用户及其称号
	rows, err := d.Query(`
		SELECT id, nickname, avatar, profession, city, ` + titleField + ` 
		FROM users 
		WHERE ` + titleField + ` != '' AND ` + titleField + ` IS NOT NULL
		ORDER BY nickname
	`)
	if err != nil {
		logger.Error("查询称号排行榜失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取排行榜失败"})
		return
	}
	defer rows.Close()

	var items []TitleRankingItem
	for rows.Next() {
		var item TitleRankingItem
		var title sql.NullString
		err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Profession, &item.City, &title)
		if err != nil {
			continue
		}
		if title.Valid && title.String != "" {
			item.Title = title.String
			item.Label = title.String
			items = append(items, item)
		}
	}

	// 按称号权重排序
	sort.Slice(items, func(i, j int) bool {
		return getTitleWeight(items[i].Title) > getTitleWeight(items[j].Title)
	})

	// 添加排名
	for i := range items {
		items[i].Rank = i + 1
	}

	if items == nil {
		items = []TitleRankingItem{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period": period,
		"list":   items,
	})
}

// getTitleWeight 获取称号权重（用于排序，权重越高排名越前）
func getTitleWeight(title string) int {
	switch {
	case strings.Contains(title, "终极") || strings.Contains(title, "超级"):
		return 100
	case strings.Contains(title, "卷王") || strings.Contains(title, "肝帝"):
		return 90
	case strings.Contains(title, "资深"):
		return 80
	case strings.Contains(title, "战士") || strings.Contains(title, "达人"):
		return 70
	case strings.Contains(title, "标准") || strings.Contains(title, "打工人"):
		return 60
	case strings.Contains(title, "自由") || strings.Contains(title, "摸鱼"):
		return 50
	case strings.Contains(title, "萌新"):
		return 40
	default:
		return 30
	}
}

// ==================== 工时排行榜 ====================

// handleWorkhoursRanking 总工时榜
func handleWorkhoursRanking(w http.ResponseWriter, r *http.Request) {
	period := getPeriodParam(r)
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
		       SUM(cr.duration) as value
		FROM clock_records cr
		JOIN users u ON u.id = cr.user_id
		WHERE cr.date >= ? AND cr.date <= ? AND cr.duration > 0
		GROUP BY cr.user_id
		ORDER BY value DESC
		LIMIT 50`, startDate, endDate)
	if err != nil {
		logger.Error("总工时榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var list []RankItem
	rank := 1
	for rows.Next() {
		var item RankItem
		if err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Profession, &item.City, &item.Value); err != nil {
			continue
		}
		item.Rank = rank
		item.Value = math.Round(item.Value*100) / 100
		item.Label = fmt.Sprintf("%.1fh", item.Value)
		list = append(list, item)
		rank++
	}

	if list == nil {
		list = []RankItem{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period": period,
		"list":   list,
	})
}

// handleAvgWorkhoursRanking 日均工时榜
func handleAvgWorkhoursRanking(w http.ResponseWriter, r *http.Request) {
	period := getPeriodParam(r)
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
		       SUM(cr.duration) as total_hours,
		       COUNT(CASE WHEN cr.duration > 0 THEN 1 END) as worked_days
		FROM clock_records cr
		JOIN users u ON u.id = cr.user_id
		WHERE cr.date >= ? AND cr.date <= ? AND cr.duration > 0
		GROUP BY cr.user_id
		HAVING worked_days > 0
		ORDER BY (total_hours / worked_days) DESC
		LIMIT 50`, startDate, endDate)
	if err != nil {
		logger.Error("日均工时榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var list []RankItem
	rank := 1
	for rows.Next() {
		var item RankItem
		var totalHours float64
		var workedDays int
		if err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Profession, &item.City, &totalHours, &workedDays); err != nil {
			continue
		}
		item.Rank = rank
		item.Value = math.Round(totalHours/float64(workedDays)*100) / 100
		item.Label = fmt.Sprintf("%.1fh/天", item.Value)
		list = append(list, item)
		rank++
	}

	if list == nil {
		list = []RankItem{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period": period,
		"list":   list,
	})
}

// ==================== 时间排行榜 ====================

// timeRanking 通用时间排行榜（早起榜/夜猫榜共用逻辑）
// field: "clock_in" 或 "clock_out"
// aggFunc: "MIN" 或 "MAX"
// ascending: true=升序(早起榜), false=降序(夜猫榜)
// labelPrefix: label 前缀文本
func timeRanking(w http.ResponseWriter, r *http.Request, field, aggFunc string, ascending bool, labelPrefix string) {
	period := getPeriodParam(r)
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	orderDir := "ASC"
	if !ascending {
		orderDir = "DESC"
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
		       %s(cr.%s) as extreme,
		       AVG(
		         CAST(SUBSTR(cr.%s, 1, 2) AS REAL) * 60 +
		         CAST(SUBSTR(cr.%s, 4, 2) AS REAL)
		       ) as avgMinutes
		FROM clock_records cr
		JOIN users u ON u.id = cr.user_id
		WHERE cr.date >= ? AND cr.date <= ? AND cr.%s IS NOT NULL
		GROUP BY cr.user_id
		ORDER BY avgMinutes %s
		LIMIT 50`, aggFunc, field, field, field, field, orderDir)

	rows, err := d.Query(query, startDate, endDate)
	if err != nil {
		logger.Error("时间排行榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var list []RankItem
	rank := 1
	for rows.Next() {
		var item RankItem
		var extreme string
		if err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Profession, &item.City, &extreme, &item.Value); err != nil {
			continue
		}
		item.Rank = rank
		h := int(item.Value) / 60
		m := int(math.Round(item.Value)) % 60
		item.Label = fmt.Sprintf("%s %02d:%02d", labelPrefix, h, m)
		list = append(list, item)
		rank++
	}

	if list == nil {
		list = []RankItem{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period": period,
		"list":   list,
	})
}

// handleEarlyRanking 早起榜
func handleEarlyRanking(w http.ResponseWriter, r *http.Request) {
	timeRanking(w, r, "clock_in", "MIN", true, "平均")
}

// handleLateRanking 夜猫榜
func handleLateRanking(w http.ResponseWriter, r *http.Request) {
	timeRanking(w, r, "clock_out", "MAX", false, "平均")
}

// ==================== 连续打卡 & 准时榜 ====================

// handleStreakRanking 连续打卡榜
func handleStreakRanking(w http.ResponseWriter, r *http.Request) {
	d := db.GetDB()

	users, err := getAllUsers()
	if err != nil {
		logger.Error("连续打卡榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}

	var list []RankItem

	for _, u := range users {
		streak := CalcClockinStreak(d, u.ID)

		if streak > 0 {
			list = append(list, RankItem{
				UserID:     u.ID,
				Nickname:   u.Nickname,
				Avatar:     u.Avatar,
				Profession: u.Profession,
				City:       u.City,
				Value:      float64(streak),
				Label:      fmt.Sprintf("%d天", streak),
			})
		}
	}

	list = finalizeRankList(list, true)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"list": list,
	})
}

// handleOntimeRanking 准时榜
func handleOntimeRanking(w http.ResponseWriter, r *http.Request) {
	period := getPeriodParam(r)
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	// 获取所有用户及其标准上班时间
	userRows, err := d.Query("SELECT id, nickname, avatar, profession, city, standard_start FROM users")
	if err != nil {
		logger.Error("准时榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer userRows.Close()

	type userWithStandard struct {
		userBasicInfo
		StandardStart string
	}

	var users []userWithStandard
	for userRows.Next() {
		var u userWithStandard
		if err := userRows.Scan(&u.ID, &u.Nickname, &u.Avatar, &u.Profession, &u.City, &u.StandardStart); err != nil {
			continue
		}
		users = append(users, u)
	}

	var list []RankItem

	for _, u := range users {
		recRows, err := d.Query(
			"SELECT clock_in FROM clock_records WHERE user_id = ? AND date >= ? AND date <= ? AND clock_in IS NOT NULL",
			u.ID, startDate, endDate,
		)
		if err != nil {
			continue
		}

		totalDays := 0
		onTimeCount := 0
		for recRows.Next() {
			var clockIn string
			if err := recRows.Scan(&clockIn); err != nil {
				continue
			}
			totalDays++
			if clockIn <= u.StandardStart {
				onTimeCount++
			}
		}
		recRows.Close()

		if totalDays == 0 {
			continue
		}

		rate := math.Round(float64(onTimeCount)/float64(totalDays)*1000) / 10

		list = append(list, RankItem{
			UserID:     u.ID,
			Nickname:   u.Nickname,
			Avatar:     u.Avatar,
			Profession: u.Profession,
			City:       u.City,
			Value:      rate,
			Label:      fmt.Sprintf("%.1f%%", rate),
		})
	}

	list = finalizeRankList(list, true)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period": period,
		"list":   list,
	})
}
