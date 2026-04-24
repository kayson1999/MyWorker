package routes

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"time"

	"myworker/db"
	"myworker/logger"
	"myworker/middleware"
	"myworker/utils"
)

// ClockRecord 打卡记录
type ClockRecord struct {
	ID        int     `json:"id"`
	UserID    string  `json:"user_id"`
	Date      string  `json:"date"`
	ClockIn   *string `json:"clock_in"`
	ClockOut  *string `json:"clock_out"`
	Duration  float64 `json:"duration"`
	IsManual  int     `json:"is_manual"`
	CreatedAt string  `json:"created_at"`
}

// RegisterClockinRoutes 注册打卡路由
func RegisterClockinRoutes(mux *http.ServeMux) {
	auth := middleware.AuthMiddleware

	mux.Handle("/api/clockin/in", methodMiddleware("POST", auth(http.HandlerFunc(handleClockIn))))
	mux.Handle("/api/clockin/out", methodMiddleware("POST", auth(http.HandlerFunc(handleClockOut))))
	mux.Handle("/api/clockin/manual", methodMiddleware("POST", auth(http.HandlerFunc(handleManual))))
	mux.Handle("/api/clockin/adjust", methodMiddleware("PUT", auth(http.HandlerFunc(handleAdjust))))
	mux.Handle("/api/clockin/today", methodMiddleware("GET", auth(http.HandlerFunc(handleToday))))
	mux.Handle("/api/clockin/records", methodMiddleware("GET", auth(http.HandlerFunc(handleRecords))))
	mux.Handle("/api/clockin/stats", methodMiddleware("GET", auth(http.HandlerFunc(handleStats))))

	// 新增称号相关路由
	mux.Handle("/api/clockin/titles", methodMiddleware("GET", auth(http.HandlerFunc(handleGetTitles))))
	mux.Handle("/api/clockin/today-title", methodMiddleware("GET", auth(http.HandlerFunc(handleGetTodayTitle))))
}

// nowTime 获取当前时间 HH:mm
func nowTime() string {
	return time.Now().Format("15:04")
}

// todayDate 获取今天日期 YYYY-MM-DD
func todayDate() string {
	return time.Now().Format("2006-01-02")
}

// calcDuration 计算工作时长（小时），支持跨午夜场景
func calcDuration(clockIn, clockOut string) float64 {
	if clockIn == "" || clockOut == "" {
		return 0
	}
	var inH, inM, outH, outM int
	fmt.Sscanf(clockIn, "%d:%d", &inH, &inM)
	fmt.Sscanf(clockOut, "%d:%d", &outH, &outM)
	inMinutes := inH*60 + inM
	outMinutes := outH*60 + outM
	// 支持跨午夜：如果下班时间小于上班时间，加24小时
	if outMinutes < inMinutes {
		outMinutes += 24 * 60
	}
	diff := float64(outMinutes-inMinutes) / 60.0
	return math.Round(diff*100) / 100
}

// getDateRange 根据周期获取日期范围
func getDateRange(period string) (startDate, endDate string) {
	today := time.Now()
	endDate = today.Format("2006-01-02")

	switch period {
	case "week":
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := today.AddDate(0, 0, -(weekday - 1))
		startDate = monday.Format("2006-01-02")
	case "year":
		startDate = fmt.Sprintf("%d-01-01", today.Year())
	default: // month
		startDate = fmt.Sprintf("%d-%02d-01", today.Year(), today.Month())
	}
	return
}

// scanRecord 扫描打卡记录
func scanRecord(row interface{ Scan(...interface{}) error }) (*ClockRecord, error) {
	var r ClockRecord
	err := row.Scan(&r.ID, &r.UserID, &r.Date, &r.ClockIn, &r.ClockOut, &r.Duration, &r.IsManual, &r.CreatedAt)
	return &r, err
}

// getRecordByUserDate 获取用户某天的打卡记录
func getRecordByUserDate(userID string, date string) *ClockRecord {
	d := db.GetDB()
	row := d.QueryRow("SELECT id, user_id, date, clock_in, clock_out, duration, is_manual, created_at FROM clock_records WHERE user_id = ? AND date = ?", userID, date)
	r, err := scanRecord(row)
	if err != nil {
		return nil
	}
	return r
}

// handleClockIn 上班打卡
func handleClockIn(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	date := todayDate()
	t := nowTime()
	d := db.GetDB()

	// 使用事务保护，避免并发问题
	tx, err := d.Begin()
	if err != nil {
		logger.Error("开启事务失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
		return
	}
	defer tx.Rollback()

	// 在事务内查询
	var existing *ClockRecord
	row := tx.QueryRow("SELECT id, user_id, date, clock_in, clock_out, duration, is_manual, created_at FROM clock_records WHERE user_id = ? AND date = ?", userID, date)
	r2, scanErr := scanRecord(row)
	if scanErr == nil {
		existing = r2
	}

	if existing != nil && existing.ClockIn != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "今天已经打过上班卡了",
			"record": existing,
		})
		return
	}

	if existing != nil {
		_, err = tx.Exec("UPDATE clock_records SET clock_in = ? WHERE id = ?", t, existing.ID)
		if err != nil {
			logger.Error("上班打卡失败: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
			return
		}
	} else {
		_, err = tx.Exec("INSERT INTO clock_records (user_id, date, clock_in) VALUES (?, ?, ?)", userID, date, t)
		if err != nil {
			logger.Error("上班打卡失败: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
			return
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Error("提交事务失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
		return
	}

	// 发放上班打卡经验值
	expSourceID := fmt.Sprintf("%s_%s_in", userID, date)
	leveledUp, newLevel, expAmount := utils.AddExp(d, userID, "clockin_in", expSourceID)

	// 早起打卡奖励(8:30前)
	var earlyH, earlyM int
	fmt.Sscanf(t, "%d:%d", &earlyH, &earlyM)
	if earlyH*60+earlyM < 8*60+30 {
		earlySourceID := fmt.Sprintf("%s_%s_early", userID, date)
		utils.AddExp(d, userID, "clockin_early", earlySourceID)
	}

	record := getRecordByUserDate(userID, date)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "上班打卡成功",
		"record":     record,
		"exp_gained": expAmount,
		"leveled_up": leveledUp,
		"new_level":  newLevel,
	})
}

// handleClockOut 下班打卡
func handleClockOut(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	date := todayDate()
	t := nowTime()
	d := db.GetDB()

	// 使用事务保护，避免并发问题
	tx, err := d.Begin()
	if err != nil {
		logger.Error("开启事务失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
		return
	}
	defer tx.Rollback()

	// 在事务内查询
	var existing *ClockRecord
	row := tx.QueryRow("SELECT id, user_id, date, clock_in, clock_out, duration, is_manual, created_at FROM clock_records WHERE user_id = ? AND date = ?", userID, date)
	r2, scanErr := scanRecord(row)
	if scanErr == nil {
		existing = r2
	}

	if existing == nil || existing.ClockIn == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先打上班卡"})
		return
	}

	// Bug修复：防止重复下班打卡
	if existing.ClockOut != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "今天已经打过下班卡了，如需修改请使用调整功能",
			"record": existing,
		})
		return
	}

	duration := calcDuration(*existing.ClockIn, t)

	_, err = tx.Exec("UPDATE clock_records SET clock_out = ?, duration = ? WHERE id = ?", t, duration, existing.ID)
	if err != nil {
		logger.Error("下班打卡失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
		return
	}

	if err = tx.Commit(); err != nil {
		logger.Error("提交事务失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
		return
	}

	// 更新用户工作风格标签
	updateUserTitles(userID)

	// 发放下班打卡经验值
	expSourceID := fmt.Sprintf("%s_%s_out", userID, date)
	leveledUp, newLevel, expAmount := utils.AddExp(d, userID, "clockin_out", expSourceID)

	// 完成全天打卡奖励
	fullSourceID := fmt.Sprintf("%s_%s_full", userID, date)
	utils.AddExp(d, userID, "clockin_full", fullSourceID)

	// 加班奖励(工时>=10h)
	if duration >= 10 {
		otSourceID := fmt.Sprintf("%s_%s_overtime", userID, date)
		utils.AddExp(d, userID, "clockin_overtime", otSourceID)
	}

	// 检查连续打卡经验奖励
	checkStreakExpReward(d, userID)

	// 检查打卡成就
	checkClockInAchievementsForUser(d, userID)

	record := getRecordByUserDate(userID, date)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "下班打卡成功",
		"record":     record,
		"exp_gained": expAmount,
		"leveled_up": leveledUp,
		"new_level":  newLevel,
	})
}

// handleManual 手动补卡
func handleManual(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	var body struct {
		Date     string `json:"date"`
		ClockIn  string `json:"clock_in"`
		ClockOut string `json:"clock_out"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.Date == "" || body.ClockIn == "" || body.ClockOut == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "日期、上班时间和下班时间不能为空"})
		return
	}

	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(body.Date) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "日期格式不正确，应为 YYYY-MM-DD"})
		return
	}

	if body.Date > todayDate() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "不能补未来日期的卡"})
		return
	}

	duration := calcDuration(body.ClockIn, body.ClockOut)

	_, err := d.Exec(`
		INSERT INTO clock_records (user_id, date, clock_in, clock_out, duration, is_manual)
		VALUES (?, ?, ?, ?, ?, 1)
		ON CONFLICT(user_id, date) DO UPDATE SET
			clock_in = excluded.clock_in,
			clock_out = excluded.clock_out,
			duration = excluded.duration,
			is_manual = 1`,
		userID, body.Date, body.ClockIn, body.ClockOut, duration,
	)
	if err != nil {
		logger.Error("补卡失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "补卡失败，请稍后重试"})
		return
	}

	record := getRecordByUserDate(userID, body.Date)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "补卡成功",
		"record":  record,
	})
}

// handleAdjust 调整打卡时间
func handleAdjust(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()
	date := todayDate()

	var body struct {
		Type string `json:"type"`
		Time string `json:"time"`
	}
	if err := readJSON(r, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数格式错误"})
		return
	}

	if body.Type == "" || body.Time == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供调整类型(type)和时间(time)"})
		return
	}

	if body.Type != "clock_in" && body.Type != "clock_out" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type 只能是 clock_in 或 clock_out"})
		return
	}

	timeRegex := regexp.MustCompile(`^\d{2}:\d{2}$`)
	if !timeRegex.MatchString(body.Time) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "时间格式不正确，应为 HH:mm"})
		return
	}

	existing := getRecordByUserDate(userID, date)
	if existing == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "今天还没有打卡记录，无法调整"})
		return
	}

	if body.Type == "clock_in" {
		if existing.ClockIn == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "还没有上班打卡记录，无法调整"})
			return
		}
		if existing.ClockOut != nil && body.Time >= *existing.ClockOut {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "上班时间不能晚于或等于下班时间"})
			return
		}
		duration := 0.0
		if existing.ClockOut != nil {
			duration = calcDuration(body.Time, *existing.ClockOut)
		}
		_, err := d.Exec("UPDATE clock_records SET clock_in = ?, duration = ? WHERE id = ?", body.Time, duration, existing.ID)
		if err != nil {
			logger.Error("调整打卡时间失败: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "调整失败，请稍后重试"})
			return
		}
	} else {
		if existing.ClockOut == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "还没有下班打卡记录，无法调整"})
			return
		}
		if existing.ClockIn != nil && body.Time <= *existing.ClockIn {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "下班时间不能早于或等于上班时间"})
			return
		}
		duration := 0.0
		if existing.ClockIn != nil {
			duration = calcDuration(*existing.ClockIn, body.Time)
		}
		_, err := d.Exec("UPDATE clock_records SET clock_out = ?, duration = ? WHERE id = ?", body.Time, duration, existing.ID)
		if err != nil {
			logger.Error("调整打卡时间失败: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "调整失败，请稍后重试"})
			return
		}
	}

	record := getRecordByUserDate(userID, date)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "打卡时间调整成功",
		"record":  record,
	})
}

// handleToday 获取今日打卡记录
func handleToday(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	record := getRecordByUserDate(userID, todayDate())
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"record": record,
	})
}

// handleRecords 查询日期范围记录
func handleRecords(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	if start == "" || end == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供 start 和 end 日期参数"})
		return
	}

	records := queryRecordsByUserAndDateRange(db.GetDB(), userID, start, end)
	if records == nil {
		records = []ClockRecord{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
	})
}

// handleStats 统计数据
func handleStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	startDate, endDate := getDateRange(period)
	d := db.GetDB()

	records := queryRecordsByUserAndDateRange(d, userID, startDate, endDate)

	totalDays := len(records)
	totalHours := 0.0
	workedDays := 0 // 有实际工时的打卡天数
	var clockInTimes []string
	var clockOutTimes []string

	for _, rec := range records {
		totalHours += rec.Duration
		if rec.Duration > 0 {
			workedDays++
		}
		if rec.ClockIn != nil {
			clockInTimes = append(clockInTimes, *rec.ClockIn)
		}
		if rec.ClockOut != nil {
			clockOutTimes = append(clockOutTimes, *rec.ClockOut)
		}
	}

	// 日均工时 = 总工时 / 有实际工时的打卡天数
	avgHours := 0.0
	if workedDays > 0 {
		avgHours = math.Round(totalHours/float64(workedDays)*100) / 100
	}

	// 最早上班
	var earliestIn *string
	if len(clockInTimes) > 0 {
		sort.Strings(clockInTimes)
		earliestIn = &clockInTimes[0]
	}

	// 最晚下班
	var latestOut *string
	if len(clockOutTimes) > 0 {
		sort.Sort(sort.Reverse(sort.StringSlice(clockOutTimes)))
		latestOut = &clockOutTimes[0]
	}

	// 连续打卡天数：调用公共函数计算，不受 period 范围限制
	streak := CalcClockinStreak(d, userID)

	// 获取用户标准时间
	var standardStart string
	err := d.QueryRow("SELECT standard_start FROM users WHERE id = ?", userID).Scan(&standardStart)
	if err != nil {
		standardStart = "09:00"
	}

	// 准时率
	onTimeCount := 0
	for _, rec := range records {
		if rec.ClockIn != nil && *rec.ClockIn <= standardStart {
			onTimeCount++
		}
	}
	onTimeRate := 0.0
	if totalDays > 0 {
		onTimeRate = math.Round(float64(onTimeCount)/float64(totalDays)*1000) / 10
	}

	if records == nil {
		records = []ClockRecord{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"period":     period,
		"startDate":  startDate,
		"endDate":    endDate,
		"totalDays":  totalDays,
		"totalHours": math.Round(totalHours*100) / 100,
		"avgHours":   avgHours,
		"earliestIn": earliestIn,
		"latestOut":  latestOut,
		"streak":     streak,
		"onTimeRate": onTimeRate,
		"records":    records,
	})
}

// queryRecordsByUserAndDateRange 查询用户日期范围内的记录
func queryRecordsByUserAndDateRange(d *sql.DB, userID string, startDate, endDate string) []ClockRecord {
	rows, err := d.Query(
		"SELECT id, user_id, date, clock_in, clock_out, duration, is_manual, created_at FROM clock_records WHERE user_id = ? AND date >= ? AND date <= ? ORDER BY date ASC",
		userID, startDate, endDate,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var records []ClockRecord
	for rows.Next() {
		var rec ClockRecord
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.Date, &rec.ClockIn, &rec.ClockOut, &rec.Duration, &rec.IsManual, &rec.CreatedAt); err != nil {
			continue
		}
		records = append(records, rec)
	}
	return records
}

// calcPeriodTitleFromRecords 从打卡记录中计算周期称号（公共函数，消除重复代码）
func calcPeriodTitleFromRecords(records []ClockRecord) string {
	var dailyTitles []utils.DailyTitle
	for _, rec := range records {
		if rec.ClockIn != nil && rec.ClockOut != nil {
			title := utils.CalculateDailyTitle(*rec.ClockIn, *rec.ClockOut, rec.Duration)
			dailyTitles = append(dailyTitles, title)
		}
	}
	return utils.CalculatePeriodTitle(dailyTitles)
}

// updateUserTitles 更新用户称号（在下班打卡成功后调用）
func updateUserTitles(userID string) {
	d := db.GetDB()

	// 验证用户存在
	var exists int
	err := d.QueryRow("SELECT 1 FROM users WHERE id = ?", userID).Scan(&exists)
	if err != nil {
		logger.Error("用户不存在，跳过称号更新: userID=%s, err=%v", userID, err)
		return
	}

	// 获取各周期记录并计算称号
	weekStart, weekEnd := getDateRange("week")
	weekTitle := calcPeriodTitleFromRecords(queryRecordsByUserAndDateRange(d, userID, weekStart, weekEnd))

	monthStart, monthEnd := getDateRange("month")
	monthTitle := calcPeriodTitleFromRecords(queryRecordsByUserAndDateRange(d, userID, monthStart, monthEnd))

	yearStart, yearEnd := getDateRange("year")
	yearTitle := calcPeriodTitleFromRecords(queryRecordsByUserAndDateRange(d, userID, yearStart, yearEnd))

	// 更新用户称号
	_, err = d.Exec("UPDATE users SET week_title = ?, month_title = ?, year_title = ? WHERE id = ?",
		weekTitle, monthTitle, yearTitle, userID)
	if err != nil {
		logger.Error("更新用户称号失败: %v", err)
	}
}

// handleGetTitles 获取用户称号
func handleGetTitles(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	d := db.GetDB()

	var weekTitle, monthTitle, yearTitle sql.NullString
	err := d.QueryRow("SELECT week_title, month_title, year_title FROM users WHERE id = ?", userID).
		Scan(&weekTitle, &monthTitle, &yearTitle)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取称号失败"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"week_title":  weekTitle.String,
		"month_title": monthTitle.String,
		"year_title":  yearTitle.String,
	})
}

// checkStreakExpReward 检查连续打卡经验奖励
func checkStreakExpReward(d *sql.DB, userID string) {
	streak := CalcClockinStreak(d, userID)

	// Bug修复：只在恰好达到里程碑时发放奖励，避免每天重复发放
	// sourceID 不含日期，确保每个里程碑只发放一次
	if streak == 7 {
		sourceID := fmt.Sprintf("%s_streak7", userID)
		utils.AddExp(d, userID, "clockin_streak_7", sourceID)
	}
	if streak == 30 {
		sourceID := fmt.Sprintf("%s_streak30", userID)
		utils.AddExp(d, userID, "clockin_streak_30", sourceID)
	}
}

// checkClockInAchievementsForUser 检查用户打卡成就
func checkClockInAchievementsForUser(d *sql.DB, userID string) {
	// 总打卡天数
	var totalDays int
	d.QueryRow("SELECT COUNT(*) FROM clock_records WHERE user_id = ?", userID).Scan(&totalDays)

	// 总工时
	var totalHours float64
	d.QueryRow("SELECT COALESCE(SUM(duration), 0) FROM clock_records WHERE user_id = ? AND duration > 0", userID).Scan(&totalHours)

	// 连续打卡天数
	streak := CalcClockinStreak(d, userID)

	utils.CheckClockInAchievements(d, userID, totalDays, streak, totalHours)
}

// handleGetTodayTitle 获取今日工作风格标签
func handleGetTodayTitle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	date := todayDate()

	record := getRecordByUserDate(userID, date)
	if record == nil || record.ClockIn == nil || record.ClockOut == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"title":     "未完成打卡",
			"sub_title": "请完成上下班打卡",
			"score":     0,
		})
		return
	}

	calculator_title := utils.CalculateDailyTitle(*record.ClockIn, *record.ClockOut, record.Duration)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"title":     calculator_title.Title,
		"sub_title": calculator_title.SubTitle,
		"clock_in":  calculator_title.ClockIn,
		"clock_out": calculator_title.ClockOut,
		"duration":  calculator_title.Duration,
		"score":     calculator_title.Score,
	})
}
