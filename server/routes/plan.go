package routes

import (
	"net/http"
	"strconv"
	"time"

	"myworker/db"
	"myworker/logger"
	"myworker/middleware"
)

// Plan 计划
type Plan struct {
	ID         int    `json:"id"`
	UserID     string `json:"user_id"`
	Title      string `json:"title"`
	Content    string `json:"content"` // 计划内容/描述
	Icon       string `json:"icon"`
	Color      string `json:"color"`
	FreqType   string `json:"freq_type"`   // daily=每天, weekday=工作日, custom=自定义
	TargetDays int    `json:"target_days"` // 目标天数，0=无限
	IsPublic   int    `json:"is_public"`   // 是否公开可见（排行榜展示），1=公开，0=私密
	Status     int    `json:"status_code"` // 1=active, 0=paused, 2=completed, -1=archived
	StatusText string `json:"status"`      // 前端友好的状态文本
	CreatedAt  string `json:"created_at"`
}

// PlanCheckin 计划打卡记录
type PlanCheckin struct {
	ID        int    `json:"id"`
	PlanID    int    `json:"plan_id"`
	UserID    string `json:"user_id"`
	Date      string `json:"date"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
}

// PlanAchievement 计划成就统计
type PlanAchievement struct {
	PlanID         int     `json:"plan_id"`
	Title          string  `json:"title"`
	Icon           string  `json:"icon"`
	Color          string  `json:"color"`
	TotalDays      int     `json:"total_days"`
	CurrentStreak  int     `json:"current_streak"`
	MaxStreak      int     `json:"max_streak"`
	TargetDays     int     `json:"target_days"`
	CompletionRate float64 `json:"completion_rate"`
	Status         string  `json:"status"`
}

// RegisterPlanRoutes 注册计划路由
func RegisterPlanRoutes(mux *http.ServeMux) {
	auth := middleware.AuthMiddleware

	// 计划 CRUD
	mux.Handle("/api/plan/list", methodMiddleware("GET", auth(http.HandlerFunc(handlePlanList))))
	mux.Handle("/api/plan/create", methodMiddleware("POST", auth(http.HandlerFunc(handlePlanCreate))))
	mux.Handle("/api/plan/update", methodMiddleware("PUT", auth(http.HandlerFunc(handlePlanUpdate))))
	mux.Handle("/api/plan/delete", methodMiddleware("DELETE", auth(http.HandlerFunc(handlePlanDelete))))

	// 计划打卡
	mux.Handle("/api/plan/checkin", methodMiddleware("POST", auth(http.HandlerFunc(handlePlanCheckin))))
	mux.Handle("/api/plan/uncheckin", methodMiddleware("POST", auth(http.HandlerFunc(handlePlanUncheckin))))
	mux.Handle("/api/plan/checkin/records", methodMiddleware("GET", auth(http.HandlerFunc(handlePlanCheckinRecords))))

	// 成就统计
	mux.Handle("/api/plan/achievements", methodMiddleware("GET", auth(http.HandlerFunc(handlePlanAchievements))))
}

// handlePlanList 获取用户的计划列表
func handlePlanList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	status := r.URL.Query().Get("status")
	d := db.GetDB()

	query := "SELECT id, user_id, title, COALESCE(content,''), icon, color, freq_type, target_days, COALESCE(is_public,1), status, created_at FROM plans WHERE user_id = ?"
	args := []interface{}{userID}

	if status != "" {
		statusCode := planStatusToCode(status)
		query += " AND status = ?"
		args = append(args, statusCode)
	}
	query += " ORDER BY created_at DESC"

	rows, err := d.Query(query, args...)
	if err != nil {
		logger.Error("查询计划列表失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Content, &p.Icon, &p.Color, &p.FreqType, &p.TargetDays, &p.IsPublic, &p.Status, &p.CreatedAt); err != nil {
			continue
		}
		p.StatusText = planStatusToText(p.Status)
		plans = append(plans, p)
	}

	if plans == nil {
		plans = []Plan{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

// handlePlanCreate 创建计划
func handlePlanCreate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	var body struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		Icon       string `json:"icon"`
		Color      string `json:"color"`
		Frequency  string `json:"frequency"`
		TargetDays int    `json:"target_days"`
		IsPublic   *int   `json:"is_public"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "计划名称不能为空"})
		return
	}

	if body.Icon == "" {
		body.Icon = "🎯"
	}
	if body.Color == "" {
		body.Color = "#A855F7"
	}
	if body.Frequency == "" {
		body.Frequency = "daily"
	}
	isPublic := 1
	if body.IsPublic != nil {
		isPublic = *body.IsPublic
	}

	// 限制每个用户最多20个活跃计划
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM plans WHERE user_id = ? AND status = 1", userID).Scan(&count)
	if err == nil && count >= 20 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "活跃计划数量已达上限（20个）"})
		return
	}

	result, err := d.Exec(
		"INSERT INTO plans (user_id, title, content, icon, color, freq_type, target_days, is_public, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1)",
		userID, body.Title, body.Content, body.Icon, body.Color, body.Frequency, body.TargetDays, isPublic,
	)
	if err != nil {
		logger.Error("创建计划失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}

	planID, _ := result.LastInsertId()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "计划创建成功",
		"plan_id": planID,
	})
}

// verifyPlanOwner 验证计划归属，返回是否通过验证
// 如果验证失败会直接写入错误响应
func verifyPlanOwner(w http.ResponseWriter, planID int, userID string) bool {
	d := db.GetDB()
	var ownerID string
	err := d.QueryRow("SELECT user_id FROM plans WHERE id = ?", planID).Scan(&ownerID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "计划不存在"})
		return false
	}
	if ownerID != userID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权操作此计划"})
		return false
	}
	return true
}

// handlePlanUpdate 更新计划
func handlePlanUpdate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	var body struct {
		ID         int    `json:"id"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Icon       string `json:"icon"`
		Color      string `json:"color"`
		Frequency  string `json:"frequency"`
		TargetDays int    `json:"target_days"`
		IsPublic   *int   `json:"is_public"`
		Status     string `json:"status"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.ID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "计划ID不能为空"})
		return
	}

	if !verifyPlanOwner(w, body.ID, userID) {
		return
	}

	statusCode := planStatusToCode(body.Status)
	isPublicVal := 1
	if body.IsPublic != nil {
		isPublicVal = *body.IsPublic
	}
	_, err := d.Exec(
		"UPDATE plans SET title = ?, content = ?, icon = ?, color = ?, freq_type = ?, target_days = ?, is_public = ?, status = ? WHERE id = ?",
		body.Title, body.Content, body.Icon, body.Color, body.Frequency, body.TargetDays, isPublicVal, statusCode, body.ID,
	)
	if err != nil {
		logger.Error("更新计划失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "计划更新成功",
	})
}

// handlePlanDelete 删除计划（软删除，改为archived）
func handlePlanDelete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	var body struct {
		ID int `json:"id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.ID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "计划ID不能为空"})
		return
	}

	if !verifyPlanOwner(w, body.ID, userID) {
		return
	}

	_, err := d.Exec("UPDATE plans SET status = -1 WHERE id = ?", body.ID)
	if err != nil {
		logger.Error("删除计划失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "计划已归档",
	})
}

// handlePlanCheckin 计划打卡
func handlePlanCheckin(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	var body struct {
		PlanID int    `json:"plan_id"`
		Date   string `json:"date"`
		Note   string `json:"note"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.PlanID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "计划ID不能为空"})
		return
	}

	if body.Date == "" {
		body.Date = time.Now().Format("2006-01-02")
	}

	// 不能打未来的卡
	today := time.Now().Format("2006-01-02")
	if body.Date > today {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "不能为未来日期打卡"})
		return
	}

	if !verifyPlanOwner(w, body.PlanID, userID) {
		return
	}

	// 检查计划状态
	var planStatus int
	d.QueryRow("SELECT status FROM plans WHERE id = ?", body.PlanID).Scan(&planStatus)
	if planStatus != 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "该计划不在进行中"})
		return
	}

	// 检查是否已打卡
	var existCount int
	d.QueryRow("SELECT COUNT(*) FROM plan_checkins WHERE plan_id = ? AND date = ?", body.PlanID, body.Date).Scan(&existCount)
	if existCount > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "该日期已打卡"})
		return
	}

	_, err := d.Exec(
		"INSERT INTO plan_checkins (plan_id, user_id, date, note) VALUES (?, ?, ?, ?)",
		body.PlanID, userID, body.Date, body.Note,
	)
	if err != nil {
		logger.Error("计划打卡失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败"})
		return
	}

	// 检查是否达成目标
	var targetDays int
	d.QueryRow("SELECT target_days FROM plans WHERE id = ?", body.PlanID).Scan(&targetDays)
	if targetDays > 0 {
		var totalCheckins int
		d.QueryRow("SELECT COUNT(*) FROM plan_checkins WHERE plan_id = ?", body.PlanID).Scan(&totalCheckins)
		if totalCheckins >= targetDays {
			d.Exec("UPDATE plans SET status = 2 WHERE id = ?", body.PlanID)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "打卡成功 ✅",
	})
}

// handlePlanUncheckin 取消计划打卡
func handlePlanUncheckin(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	var body struct {
		PlanID int    `json:"plan_id"`
		Date   string `json:"date"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.PlanID == 0 || body.Date == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参数不完整"})
		return
	}

	if !verifyPlanOwner(w, body.PlanID, userID) {
		return
	}

	_, err := d.Exec("DELETE FROM plan_checkins WHERE plan_id = ? AND date = ? AND user_id = ?", body.PlanID, body.Date, userID)
	if err != nil {
		logger.Error("取消打卡失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "取消失败"})
		return
	}

	// 如果计划已完成，重新激活
	d.Exec("UPDATE plans SET status = 1 WHERE id = ? AND status = 2", body.PlanID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "已取消打卡",
	})
}

// handlePlanCheckinRecords 获取计划打卡记录
func handlePlanCheckinRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	planIDStr := r.URL.Query().Get("plan_id")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if planIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供 plan_id"})
		return
	}

	planID, _ := strconv.Atoi(planIDStr)

	if !verifyPlanOwner(w, planID, userID) {
		return
	}

	query := "SELECT id, plan_id, user_id, date, note, created_at FROM plan_checkins WHERE plan_id = ? AND user_id = ?"
	args := []interface{}{planID, userID}

	if start != "" && end != "" {
		query += " AND date >= ? AND date <= ?"
		args = append(args, start, end)
	}
	query += " ORDER BY date DESC"

	rows, err := d.Query(query, args...)
	if err != nil {
		logger.Error("查询打卡记录失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var records []PlanCheckin
	for rows.Next() {
		var rec PlanCheckin
		if err := rows.Scan(&rec.ID, &rec.PlanID, &rec.UserID, &rec.Date, &rec.Note, &rec.CreatedAt); err != nil {
			continue
		}
		records = append(records, rec)
	}

	if records == nil {
		records = []PlanCheckin{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
	})
}

// handlePlanAchievements 获取用户所有计划的成就统计
func handlePlanAchievements(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	rows, err := d.Query(
		"SELECT id, title, icon, color, target_days, status FROM plans WHERE user_id = ? AND status != -1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		logger.Error("查询成就失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}

	// 先收集所有计划基础数据，然后关闭 rows，避免 SQLite 单连接死锁
	type planBase struct {
		PlanID     int
		Title      string
		Icon       string
		Color      string
		TargetDays int
		StatusCode int
	}
	var planBases []planBase
	for rows.Next() {
		var pb planBase
		if err := rows.Scan(&pb.PlanID, &pb.Title, &pb.Icon, &pb.Color, &pb.TargetDays, &pb.StatusCode); err != nil {
			continue
		}
		planBases = append(planBases, pb)
	}
	rows.Close()

	var achievements []PlanAchievement
	today := time.Now()

	for _, pb := range planBases {
		a := PlanAchievement{
			PlanID:     pb.PlanID,
			Title:      pb.Title,
			Icon:       pb.Icon,
			Color:      pb.Color,
			TargetDays: pb.TargetDays,
			Status:     planStatusToText(pb.StatusCode),
		}

		// 查询总打卡天数
		d.QueryRow("SELECT COUNT(*) FROM plan_checkins WHERE plan_id = ?", a.PlanID).Scan(&a.TotalDays)

		// 计算当前连续天数
		a.CurrentStreak = calcPlanStreak(a.PlanID, today)

		// 计算最长连续天数
		a.MaxStreak = calcPlanMaxStreak(a.PlanID)

		// 计算完成率
		if a.TargetDays > 0 {
			a.CompletionRate = float64(a.TotalDays) / float64(a.TargetDays) * 100
			if a.CompletionRate > 100 {
				a.CompletionRate = 100
			}
		}

		achievements = append(achievements, a)
	}

	if achievements == nil {
		achievements = []PlanAchievement{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"achievements": achievements,
	})
}

// calcPlanStreak 计算计划当前连续打卡天数（从昨天开始往前数，因为今天还没结束）
func calcPlanStreak(planID int, today time.Time) int {
	d := db.GetDB()
	dateRows, err := d.Query("SELECT date FROM plan_checkins WHERE plan_id = ? ORDER BY date DESC", planID)
	if err != nil {
		return 0
	}
	defer dateRows.Close()

	dateSet := make(map[string]bool)
	for dateRows.Next() {
		var date string
		if err := dateRows.Scan(&date); err != nil {
			continue
		}
		dateSet[date] = true
	}

	streak := 0
	// 如果今天已打卡，则从今天开始计入连续天数；
	// 否则（今天未打卡）从昨天开始向前计算，避免"今天还没到打卡时间"导致连续天数清零。
	var checkDate time.Time
	todayStr := today.Format("2006-01-02")
	if dateSet[todayStr] {
		checkDate = today
	} else {
		checkDate = today.AddDate(0, 0, -1)
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

// calcPlanMaxStreak 计算计划最长连续打卡天数
func calcPlanMaxStreak(planID int) int {
	d := db.GetDB()
	dateRows, err := d.Query("SELECT date FROM plan_checkins WHERE plan_id = ? ORDER BY date ASC", planID)
	if err != nil {
		return 0
	}
	defer dateRows.Close()

	var dates []string
	for dateRows.Next() {
		var date string
		if err := dateRows.Scan(&date); err != nil {
			continue
		}
		dates = append(dates, date)
	}

	if len(dates) == 0 {
		return 0
	}

	maxStreak := 1
	currentStreak := 1

	for i := 1; i < len(dates); i++ {
		prev, _ := time.Parse("2006-01-02", dates[i-1])
		curr, _ := time.Parse("2006-01-02", dates[i])
		diff := curr.Sub(prev).Hours() / 24

		if diff == 1 {
			currentStreak++
			if currentStreak > maxStreak {
				maxStreak = currentStreak
			}
		} else {
			currentStreak = 1
		}
	}

	return maxStreak
}

// planStatusToCode 将状态文本转为数据库状态码
func planStatusToCode(status string) int {
	switch status {
	case "active":
		return 1
	case "paused":
		return 0
	case "completed":
		return 2
	case "archived":
		return -1
	default:
		return 1
	}
}

// planStatusToText 将数据库状态码转为前端友好文本
func planStatusToText(code int) string {
	switch code {
	case 1:
		return "active"
	case 0:
		return "paused"
	case 2:
		return "completed"
	case -1:
		return "archived"
	default:
		return "active"
	}
}

// InitPlanTables 初始化计划相关表
func InitPlanTables() {
	d := db.GetDB()

	schema := `
	CREATE TABLE IF NOT EXISTS plans (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT DEFAULT '',
		icon TEXT DEFAULT '🎯',
		color TEXT DEFAULT '#A855F7',
		freq_type TEXT DEFAULT 'daily',
		target_days INTEGER DEFAULT 0,
		is_public INTEGER DEFAULT 1,
		status INTEGER DEFAULT 1,
		created_at TEXT DEFAULT (datetime('now', 'localtime')),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS plan_checkins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plan_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		note TEXT DEFAULT '',
		created_at TEXT DEFAULT (datetime('now', 'localtime')),
		FOREIGN KEY (plan_id) REFERENCES plans(id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(plan_id, user_id, date)
	);

	CREATE INDEX IF NOT EXISTS idx_plans_user ON plans(user_id, status);
	CREATE INDEX IF NOT EXISTS idx_plan_checkins_plan ON plan_checkins(plan_id, user_id, date);
	CREATE INDEX IF NOT EXISTS idx_plan_checkins_user ON plan_checkins(user_id, date);
	`

	_, err := d.Exec(schema)
	if err != nil {
		logger.Error("初始化计划表失败: %v", err)
		panic(err)
	}

	// 兼容旧表：为已有的 plans 表添加新字段（如果不存在则忽略错误）
	d.Exec("ALTER TABLE plans ADD COLUMN content TEXT DEFAULT ''")
	d.Exec("ALTER TABLE plans ADD COLUMN is_public INTEGER DEFAULT 1")

	logger.Info("✅ 计划表初始化完成")
}
