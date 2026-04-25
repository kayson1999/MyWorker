package routes

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"

	"myworker/db"
	"myworker/logger"
)

// RegisterPlanRankingRoutes 注册计划打卡排行榜路由（公开接口）
func RegisterPlanRankingRoutes(mux *http.ServeMux) {
	mux.Handle("/api/plan-ranking/total", methodMiddleware("GET", http.HandlerFunc(handlePlanTotalRanking)))
	mux.Handle("/api/plan-ranking/streak", methodMiddleware("GET", http.HandlerFunc(handlePlanStreakRanking)))
	mux.Handle("/api/plan-ranking/plans", methodMiddleware("GET", http.HandlerFunc(handlePlanCountRanking)))
	mux.Handle("/api/plan-ranking/completion", methodMiddleware("GET", http.HandlerFunc(handlePlanCompletionRanking)))
}

// handlePlanTotalRanking 计划总打卡天数榜
func handlePlanTotalRanking(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	d := db.GetDB()

	var query string
	var args []interface{}

	if period != "" && period != "all" {
		startDate, endDate := getDateRange(period)
		query = `
			SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
			       COUNT(DISTINCT pc.date) as value
			FROM plan_checkins pc
			JOIN users u ON u.id = pc.user_id
			JOIN plans p ON p.id = pc.plan_id AND p.is_public = 1
			WHERE pc.date >= ? AND pc.date <= ?
			GROUP BY pc.user_id
			ORDER BY value DESC
			LIMIT 50`
		args = []interface{}{startDate, endDate}
	} else {
		query = `
			SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
			       COUNT(DISTINCT pc.date) as value
			FROM plan_checkins pc
			JOIN users u ON u.id = pc.user_id
			JOIN plans p ON p.id = pc.plan_id AND p.is_public = 1
			GROUP BY pc.user_id
			ORDER BY value DESC
			LIMIT 50`
	}

	rows, err := d.Query(query, args...)
	if err != nil {
		logger.Error("计划总打卡榜查询失败: %v", err)
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
		item.Label = fmt.Sprintf("%d天", int(item.Value))
		item.Plans = getUserPublicPlans(d, item.UserID)
		list = append(list, item)
		rank++
	}

	if list == nil {
		list = []RankItem{}
	}

	page, pageSize := getPaginationParams(r)
	pagedList, total := paginateRankList(list, page, pageSize)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period":    period,
		"list":      pagedList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// handlePlanStreakRanking 计划连续打卡榜（取每个用户所有计划中最长的连续打卡）
func handlePlanStreakRanking(w http.ResponseWriter, r *http.Request) {
	d := db.GetDB()

	// 获取所有有公开活跃计划的用户
	userRows, err := d.Query(`
		SELECT DISTINCT u.id, u.nickname, u.avatar, u.profession, u.city
		FROM plans p
		JOIN users u ON u.id = p.user_id
		WHERE p.status IN (1, 2) AND p.is_public = 1`)
	if err != nil {
		logger.Error("计划连续打卡榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer userRows.Close()

	type userInfo struct {
		ID, Nickname, Avatar, Profession, City string
	}
	var users []userInfo
	for userRows.Next() {
		var u userInfo
		if err := userRows.Scan(&u.ID, &u.Nickname, &u.Avatar, &u.Profession, &u.City); err != nil {
			continue
		}
		users = append(users, u)
	}

	var list []RankItem

	for _, u := range users {
		// 获取该用户所有计划（先收集 ID，关闭 rows 后再计算，避免 SQLite 单连接死锁）
		planRows, err := d.Query("SELECT id FROM plans WHERE user_id = ? AND status IN (1, 2) AND is_public = 1", u.ID)
		if err != nil {
			continue
		}

		var planIDs []int
		for planRows.Next() {
			var planID int
			if err := planRows.Scan(&planID); err != nil {
				continue
			}
			planIDs = append(planIDs, planID)
		}
		planRows.Close()

		maxStreak := 0
		for _, planID := range planIDs {
			streak := calcPlanStreak(planID)
			if streak > maxStreak {
				maxStreak = streak
			}
		}

		if maxStreak > 0 {
			list = append(list, RankItem{
				UserID:     u.ID,
				Nickname:   u.Nickname,
				Avatar:     u.Avatar,
				Profession: u.Profession,
				City:       u.City,
				Value:      float64(maxStreak),
				Label:      fmt.Sprintf("%d天", maxStreak),
				Plans:      getUserPublicPlans(d, u.ID),
			})
		}
	}

	list = finalizeRankList(list, true)

	page, pageSize := getPaginationParams(r)
	pagedList, total := paginateRankList(list, page, pageSize)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"list":      pagedList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// handlePlanCountRanking 活跃计划数榜
func handlePlanCountRanking(w http.ResponseWriter, r *http.Request) {
	d := db.GetDB()

	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
		       COUNT(*) as value
		FROM plans p
		JOIN users u ON u.id = p.user_id
		WHERE p.status IN (1, 2) AND p.is_public = 1
		GROUP BY p.user_id
		ORDER BY value DESC
		LIMIT 50`)
	if err != nil {
		logger.Error("活跃计划数榜查询失败: %v", err)
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
		item.Label = fmt.Sprintf("%d个", int(item.Value))
		item.Plans = getUserPublicPlans(d, item.UserID)
		list = append(list, item)
		rank++
	}

	if list == nil {
		list = []RankItem{}
	}

	page, pageSize := getPaginationParams(r)
	pagedList, total := paginateRankList(list, page, pageSize)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"list":      pagedList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// handlePlanCompletionRanking 计划完成率榜
func handlePlanCompletionRanking(w http.ResponseWriter, r *http.Request) {
	d := db.GetDB()

	// 查询有目标天数的计划，计算完成率
	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, u.profession, u.city,
		p.id as plan_id, p.title, p.target_days
		FROM plans p
		JOIN users u ON u.id = p.user_id
		WHERE p.target_days > 0 AND p.status IN (1, 2) AND p.is_public = 1`)
	if err != nil {
		logger.Error("计划完成率榜查询失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	// 按用户聚合：计算每个用户所有有目标计划的平均完成率
	type userAgg struct {
		Nickname, Avatar, Profession, City string
		TotalRate                          float64
		PlanCount                          int
	}

	// 先收集所有计划数据，关闭 rows 后再查询打卡数，避免 SQLite 单连接死锁
	type planInfo struct {
		UserID, Nickname, Avatar, Profession, City string
		PlanID, TargetDays                         int
	}
	var planInfos []planInfo
	for rows.Next() {
		var pi planInfo
		var title string
		if err := rows.Scan(&pi.UserID, &pi.Nickname, &pi.Avatar, &pi.Profession, &pi.City, &pi.PlanID, &title, &pi.TargetDays); err != nil {
			continue
		}
		planInfos = append(planInfos, pi)
	}
	rows.Close()

	userMap := make(map[string]*userAgg)
	for _, pi := range planInfos {
		var checkinCount int
		d.QueryRow("SELECT COUNT(*) FROM plan_checkins WHERE plan_id = ?", pi.PlanID).Scan(&checkinCount)

		rate := float64(checkinCount) / float64(pi.TargetDays) * 100
		if rate > 100 {
			rate = 100
		}

		if _, ok := userMap[pi.UserID]; !ok {
			userMap[pi.UserID] = &userAgg{
				Nickname:   pi.Nickname,
				Avatar:     pi.Avatar,
				Profession: pi.Profession,
				City:       pi.City,
			}
		}
		userMap[pi.UserID].TotalRate += rate
		userMap[pi.UserID].PlanCount++
	}

	var list []RankItem
	for userID, agg := range userMap {
		avgRate := math.Round(agg.TotalRate/float64(agg.PlanCount)*10) / 10
		list = append(list, RankItem{
			UserID:     userID,
			Nickname:   agg.Nickname,
			Avatar:     agg.Avatar,
			Profession: agg.Profession,
			City:       agg.City,
			Value:      avgRate,
			Label:      fmt.Sprintf("%.1f%%", avgRate),
		})
	}

	list = finalizeRankList(list, true)

	for i := range list {
		list[i].Plans = getUserPublicPlans(d, list[i].UserID)
	}

	page, pageSize := getPaginationParams(r)
	pagedList, total := paginateRankList(list, page, pageSize)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"list":      pagedList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// getUserPublicPlans 获取用户的公开计划摘要列表
func getUserPublicPlans(d *sql.DB, userID string) []RankPlanBrief {
	rows, err := d.Query(
		"SELECT title, COALESCE(content,''), icon FROM plans WHERE user_id = ? AND is_public = 1 AND status IN (1, 2) ORDER BY created_at DESC LIMIT 5",
		userID,
	)
	if err != nil {
		return []RankPlanBrief{}
	}
	defer rows.Close()

	var plans []RankPlanBrief
	for rows.Next() {
		var p RankPlanBrief
		if err := rows.Scan(&p.Title, &p.Content, &p.Icon); err != nil {
			continue
		}
		plans = append(plans, p)
	}
	if plans == nil {
		plans = []RankPlanBrief{}
	}
	return plans
}
