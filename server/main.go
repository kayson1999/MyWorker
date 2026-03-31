package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"myworker/config"
	"myworker/db"
	"myworker/logger"
	"myworker/middleware"
	"myworker/routes"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志
	logger.Init(logger.LogConfig{
		LogDir:    cfg.LogDir,
		LogLevel:  cfg.LogLevel,
		MaxFiles:  cfg.LogMaxFiles,
		ToConsole: cfg.LogToConsole,
	})

	// 初始化数据库
	// 数据库目录：工作目录/db（避免 go run 时存到临时目录导致数据丢失）
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	dbDir := filepath.Join(cwd, "data")
	os.MkdirAll(dbDir, 0755)
	db.SetDBDir(dbDir)
	db.GetDB()
	defer db.Close()

	logger.Info("✅ 数据库初始化完成")
	logger.Info("🔗 用户中心地址: %s", cfg.UserCenterURL)
	logger.Info("🏷️  租户标识: %s", cfg.AppID)

	// 创建路由
	mux := http.NewServeMux()

	// 注册 API 路由
	routes.RegisterAuthRoutes(mux)
	routes.RegisterUserRoutes(mux)
	routes.RegisterClockinRoutes(mux)
	routes.RegisterRankingRoutes(mux)

	// 健康检查
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// 托管前端静态文件（生产环境）
	exe, _ := os.Executable()
	distDir := filepath.Join(filepath.Dir(exe), "dist")
	if _, err := os.Stat(distDir); err == nil {
		fileServer := http.FileServer(http.Dir(distDir))

		// SPA 回退：非 API 且非静态文件的请求返回 index.html
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// 检查文件是否存在
			path := filepath.Join(distDir, r.URL.Path)
			if _, err := os.Stat(path); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
			// SPA 回退
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		})
	}

	// CORS 中间件
	handler := corsMiddleware(mux)

	// 请求日志中间件
	handler = middleware.RequestLogger(handler)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("🚀 打工人打卡服务已启动: http://localhost%s/worker", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("服务启动失败: %v", err)
		os.Exit(1)
	}
}

// corsMiddleware CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
