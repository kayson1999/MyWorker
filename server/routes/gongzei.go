package routes

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"myworker/db"
	"myworker/logger"
	"myworker/middleware"
)

// GongzeiItem 工贼榜/光荣榜项
type GongzeiItem struct {
	UserID   string  `json:"userId"`
	Nickname string  `json:"nickname"`
	Avatar   string  `json:"avatar"`
	Hours    float64 `json:"hours"`
	Label    string  `json:"label"`
}

// gongzeiCache 工贼榜内存缓存
var (
	gongzeiMu   sync.RWMutex
	gongzeiList []GongzeiItem
	gongzeiWeek string // 缓存对应的周标识，如 "2026-W16"
)

// guangrongCache 光荣榜内存缓存（上周工时最短TOP3）
var (
	guangrongMu   sync.RWMutex
	guangrongList []GongzeiItem
	guangrongWeek string
)

// RegisterGongzeiRoutes 注册工贼榜 & 光荣榜路由
func RegisterGongzeiRoutes(mux *http.ServeMux) {
	auth := middleware.AuthMiddleware
	mux.Handle("/api/gongzei/top", methodMiddleware("GET", auth(http.HandlerFunc(handleGongzeiTop))))
	mux.Handle("/api/gongzei/glory", methodMiddleware("GET", auth(http.HandlerFunc(handleGuangrongTop))))
	mux.Handle("/api/gongzei/all", methodMiddleware("GET", auth(http.HandlerFunc(handleBoardAll))))
}

// InitGongzeiCache 服务启动时加载工贼榜 & 光荣榜缓存，并启动定时刷新
func InitGongzeiCache() {
	// 启动时立即加载一次
	refreshGongzeiCache()
	refreshGuangrongCache()
	logger.Info("🏴 工贼榜 & 🏅 光荣榜缓存已初始化")

	// 启动后台定时器：每周一早晨8点刷新
	go gongzeiScheduler()
}

// gongzeiScheduler 定时调度器：计算到下一个周一8:00的等待时间，然后每周执行一次
func gongzeiScheduler() {
	for {
		now := time.Now()
		next := nextMondayEight(now)
		waitDuration := next.Sub(now)
		logger.Info("🏴 榜单下次刷新时间: %s (等待 %s)", next.Format("2006-01-02 15:04:05"), waitDuration)

		timer := time.NewTimer(waitDuration)
		<-timer.C

		logger.Info("🏴 榜单定时刷新触发")
		refreshGongzeiCache()
		refreshGuangrongCache()
	}
}

// nextMondayEight 计算从 now 开始的下一个周一 08:00:00
func nextMondayEight(now time.Time) time.Time {
	// 找到本周一
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	mondayEight := time.Date(monday.Year(), monday.Month(), monday.Day(), 8, 0, 0, 0, now.Location())

	// 如果当前时间已经过了本周一8点，则目标是下周一8点
	if now.After(mondayEight) || now.Equal(mondayEight) {
		mondayEight = mondayEight.AddDate(0, 0, 7)
	}
	return mondayEight
}

// getLastWeekRange 获取上周的日期范围和周标识（复用逻辑）
func getLastWeekRange() (startDate, endDate, weekLabel string) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := now.AddDate(0, 0, -(weekday - 1))
	lastMonday := thisMonday.AddDate(0, 0, -7)
	lastSunday := thisMonday.AddDate(0, 0, -1)

	startDate = lastMonday.Format("2006-01-02")
	endDate = lastSunday.Format("2006-01-02")

	year, week := lastMonday.ISOWeek()
	weekLabel = fmt.Sprintf("%d-W%02d", year, week)
	return
}

// refreshGongzeiCache 从数据库查询上周总工时TOP3（最高）并更新内存缓存
func refreshGongzeiCache() {
	d := db.GetDB()
	startDate, endDate, weekLabel := getLastWeekRange()

	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, SUM(cr.duration) as total_hours
		FROM clock_records cr
		JOIN users u ON u.id = cr.user_id
		WHERE cr.date >= ? AND cr.date <= ? AND cr.duration > 0
		GROUP BY cr.user_id
		ORDER BY total_hours DESC
		LIMIT 3`, startDate, endDate)
	if err != nil {
		logger.Error("工贼榜查询失败: %v", err)
		return
	}
	defer rows.Close()

	var list []GongzeiItem
	for rows.Next() {
		var item GongzeiItem
		if err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Hours); err != nil {
			continue
		}
		item.Hours = float64(int(item.Hours*100)) / 100
		item.Label = fmt.Sprintf("%.1fh", item.Hours)
		list = append(list, item)
	}

	if list == nil {
		list = []GongzeiItem{}
	}

	gongzeiMu.Lock()
	gongzeiList = list
	gongzeiWeek = weekLabel
	gongzeiMu.Unlock()

	logger.Info("🏴 工贼榜已刷新: %s, 共 %d 人", weekLabel, len(list))
}

// refreshGuangrongCache 从数据库查询上周总工时最短TOP3并更新内存缓存
func refreshGuangrongCache() {
	d := db.GetDB()
	startDate, endDate, weekLabel := getLastWeekRange()

	rows, err := d.Query(`
		SELECT u.id, u.nickname, u.avatar, SUM(cr.duration) as total_hours
		FROM clock_records cr
		JOIN users u ON u.id = cr.user_id
		WHERE cr.date >= ? AND cr.date <= ? AND cr.duration > 0
		GROUP BY cr.user_id
		ORDER BY total_hours ASC
		LIMIT 3`, startDate, endDate)
	if err != nil {
		logger.Error("光荣榜查询失败: %v", err)
		return
	}
	defer rows.Close()

	var list []GongzeiItem
	for rows.Next() {
		var item GongzeiItem
		if err := rows.Scan(&item.UserID, &item.Nickname, &item.Avatar, &item.Hours); err != nil {
			continue
		}
		item.Hours = float64(int(item.Hours*100)) / 100
		item.Label = fmt.Sprintf("%.1fh", item.Hours)
		list = append(list, item)
	}

	if list == nil {
		list = []GongzeiItem{}
	}

	guangrongMu.Lock()
	guangrongList = list
	guangrongWeek = weekLabel
	guangrongMu.Unlock()

	logger.Info("🏅 光荣榜已刷新: %s, 共 %d 人", weekLabel, len(list))
}

// handleGongzeiTop 获取工贼榜TOP3（从内存缓存读取）
func handleGongzeiTop(w http.ResponseWriter, r *http.Request) {
	gongzeiMu.RLock()
	list := gongzeiList
	week := gongzeiWeek
	gongzeiMu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"week": week,
		"list": list,
	})
}

// handleGuangrongTop 获取光荣榜TOP3（从内存缓存读取）
func handleGuangrongTop(w http.ResponseWriter, r *http.Request) {
	guangrongMu.RLock()
	list := guangrongList
	week := guangrongWeek
	guangrongMu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"week": week,
		"list": list,
	})
}

// handleBoardAll 同时获取工贼榜和光荣榜（一次请求返回两个榜单）
func handleBoardAll(w http.ResponseWriter, r *http.Request) {
	gongzeiMu.RLock()
	gzList := gongzeiList
	gzWeek := gongzeiWeek
	gongzeiMu.RUnlock()

	guangrongMu.RLock()
	grList := guangrongList
	grWeek := guangrongWeek
	guangrongMu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"gongzei": map[string]interface{}{
			"week": gzWeek,
			"list": gzList,
		},
		"guangrong": map[string]interface{}{
			"week": grWeek,
			"list": grList,
		},
	})
}
