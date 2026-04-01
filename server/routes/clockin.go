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
}

// nowTime 获取当前时间 HH:mm
func nowTime() string {
	return time.Now().Format("15:04")
}

// todayDate 获取今天日期 YYYY-MM-DD
func todayDate() string {
	return time.Now().Format("2006-01-02")
}

// calcDuration 计算工作时长（小时）
func calcDuration(clockIn, clockOut string) float64 {
	if clockIn == "" || clockOut == "" {
		return 0
	}
	var inH, inM, outH, outM int
	fmt.Sscanf(clockIn, "%d:%d", &inH, &inM)
	fmt.Sscanf(clockOut, "%d:%d", &outH, &outM)
	diff := float64((outH*60+outM)-(inH*60+inM)) / 60.0
	if diff < 0 {
		diff = 0
	}
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

	existing := getRecordByUserDate(userID, date)

	if existing != nil && existing.ClockIn != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "今天已经打过上班卡了",
			"record": existing,
		})
		return
	}

	if existing != nil {
		_, err := d.Exec("UPDATE clock_records SET clock_in = ? WHERE id = ?", t, existing.ID)
		if err != nil {
			logger.Error("上班打卡失败: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
			return
		}
	} else {
		_, err := d.Exec("INSERT INTO clock_records (user_id, date, clock_in) VALUES (?, ?, ?)", userID, date, t)
		if err != nil {
			logger.Error("上班打卡失败: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
			return
		}
	}

	record := getRecordByUserDate(userID, date)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "上班打卡成功",
		"record":  record,
	})
}

// handleClockOut 下班打卡
func handleClockOut(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	date := todayDate()
	t := nowTime()
	d := db.GetDB()

	existing := getRecordByUserDate(userID, date)

	if existing == nil || existing.ClockIn == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先打上班卡"})
		return
	}

	duration := calcDuration(*existing.ClockIn, t)

	_, err := d.Exec("UPDATE clock_records SET clock_out = ?, duration = ? WHERE id = ?", t, duration, existing.ID)
	if err != nil {
		logger.Error("下班打卡失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "打卡失败，请稍后重试"})
		return
	}

	record := getRecordByUserDate(userID, date)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "下班打卡成功",
		"record":  record,
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

	d := db.GetDB()
	rows, err := d.Query(
		"SELECT id, user_id, date, clock_in, clock_out, duration, is_manual, created_at FROM clock_records WHERE user_id = ? AND date >= ? AND date <= ? ORDER BY date ASC",
		userID, start, end,
	)
	if err != nil {
		logger.Error("查询记录失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
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

	rows, err := d.Query(
		"SELECT id, user_id, date, clock_in, clock_out, duration, is_manual, created_at FROM clock_records WHERE user_id = ? AND date >= ? AND date <= ? ORDER BY date ASC",
		userID, startDate, endDate,
	)
	if err != nil {
		logger.Error("统计失败: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "统计失败"})
		return
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

	// 连续打卡天数
	streak := 0
	dateSet := make(map[string]bool)
	for _, rec := range records {
		dateSet[rec.Date] = true
	}
	checkDate := time.Now()
	for {
		ds := checkDate.Format("2006-01-02")
		if dateSet[ds] {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	// 获取用户标准时间
	var standardStart string
	err = d.QueryRow("SELECT standard_start FROM users WHERE id = ?", userID).Scan(&standardStart)
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
