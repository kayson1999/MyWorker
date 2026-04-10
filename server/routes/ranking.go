package routes

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"myworker/db"
	"myworker/logger"
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

// RegisterRankingRoutes 注册排行榜路由（公开接口，无需登录）
func RegisterRankingRoutes(mux *http.ServeMux) {
	mux.Handle("/api/ranking/workhours", methodMiddleware("GET", http.HandlerFunc(handleWorkhoursRanking)))
	mux.Handle("/api/ranking/avgworkhours", methodMiddleware("GET", http.HandlerFunc(handleAvgWorkhoursRanking)))
	mux.Handle("/api/ranking/early", methodMiddleware("GET", http.HandlerFunc(handleEarlyRanking)))
	mux.Handle("/api/ranking/late", methodMiddleware("GET", http.HandlerFunc(handleLateRanking)))
	mux.Handle("/api/ranking/streak", methodMiddleware("GET", http.HandlerFunc(handleStreakRanking)))
	mux.Handle("/api/ranking/ontime", methodMiddleware("GET", http.HandlerFunc(handleOntimeRanking)))
}

// handleWorkhoursRanking 总工时榜
func handleWorkhoursRanking(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
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
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	// 查询每个用户的总工时和有效打卡天数（duration > 0 的天数）
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

// handleEarlyRanking 早起榜
func handleEarlyRanking(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
		       MIN(cr.clock_in) as earliest,
		       AVG(
		         CAST(SUBSTR(cr.clock_in, 1, 2) AS REAL) * 60 +
		         CAST(SUBSTR(cr.clock_in, 4, 2) AS REAL)
		       ) as avgMinutes
		FROM clock_records cr
		JOIN users u ON u.id = cr.user_id
		WHERE cr.date >= ? AND cr.date <= ? AND cr.clock_in IS NOT NULL
		GROUP BY cr.user_id
		ORDER BY avgMinutes ASC
		LIMIT 50`, startDate, endDate)
	if err != nil {
		logger.Error("早起榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var list []RankItem
	rank := 1
	for rows.Next() {
		var item RankItem
		var earliest string
		if err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Profession, &item.City, &earliest, &item.Value); err != nil {
			continue
		}
		item.Rank = rank
		h := int(item.Value) / 60
		m := int(math.Round(item.Value)) % 60
		item.Label = fmt.Sprintf("平均 %02d:%02d", h, m)
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

// handleLateRanking 夜猫榜
func handleLateRanking(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
		       MAX(cr.clock_out) as latest,
		       AVG(
		         CAST(SUBSTR(cr.clock_out, 1, 2) AS REAL) * 60 +
		         CAST(SUBSTR(cr.clock_out, 4, 2) AS REAL)
		       ) as avgMinutes
		FROM clock_records cr
		JOIN users u ON u.id = cr.user_id
		WHERE cr.date >= ? AND cr.date <= ? AND cr.clock_out IS NOT NULL
		GROUP BY cr.user_id
		ORDER BY avgMinutes DESC
		LIMIT 50`, startDate, endDate)
	if err != nil {
		logger.Error("夜猫榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var list []RankItem
	rank := 1
	for rows.Next() {
		var item RankItem
		var latest string
		if err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Profession, &item.City, &latest, &item.Value); err != nil {
			continue
		}
		item.Rank = rank
		h := int(item.Value) / 60
		m := int(math.Round(item.Value)) % 60
		item.Label = fmt.Sprintf("平均 %02d:%02d", h, m)
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

// handleStreakRanking 连续打卡榜
func handleStreakRanking(w http.ResponseWriter, r *http.Request) {
	d := db.GetDB()

	// 获取所有用户
	userRows, err := d.Query("SELECT id, nickname, avatar, profession, city FROM users")
	if err != nil {
		logger.Error("连续打卡榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer userRows.Close()

	type userInfo struct {
		ID         string
		Nickname   string
		Avatar     string
		Profession string
		City       string
	}

	var users []userInfo
	for userRows.Next() {
		var u userInfo
		if err := userRows.Scan(&u.ID, &u.Nickname, &u.Avatar, &u.Profession, &u.City); err != nil {
			continue
		}
		users = append(users, u)
	}

	today := time.Now()
	var list []RankItem

	for _, u := range users {
		dateRows, err := d.Query("SELECT date FROM clock_records WHERE user_id = ? ORDER BY date DESC", u.ID)
		if err != nil {
			continue
		}

		dateSet := make(map[string]bool)
		for dateRows.Next() {
			var date string
			if err := dateRows.Scan(&date); err != nil {
				continue
			}
			dateSet[date] = true
		}
		dateRows.Close()

		streak := 0
		// 今天还没结束，连续打卡截止到昨天开始计算
		checkDate := today.AddDate(0, 0, -1)
		for {
			ds := checkDate.Format("2006-01-02")
			if dateSet[ds] {
				streak++
				checkDate = checkDate.AddDate(0, 0, -1)
			} else {
				break
			}
		}

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
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	userRows, err := d.Query("SELECT id, nickname, avatar, profession, city, standard_start FROM users")
	if err != nil {
		logger.Error("准时榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer userRows.Close()

	type userWithStandard struct {
		ID            string
		Nickname      string
		Avatar        string
		Profession    string
		City          string
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
