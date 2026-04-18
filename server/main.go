package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	// 优先使用环境变量 DB_DIR，否则使用工作目录/data
	dbDir := os.Getenv("DB_DIR")
	if dbDir == "" {
		cwd, _ := os.Getwd()
		dbDir = filepath.Join(cwd, "data")
	}
	os.MkdirAll(dbDir, 0755)
	db.SetDBDir(dbDir)
	db.GetDB()
	routes.InitPlanTables()
	defer db.Close()

	logger.Info("✅ 数据库初始化完成")
	logger.Info("🔗 用户中心地址: %s", cfg.UserCenterURL)
	logger.Info("🏷️  租户标识: %s", cfg.AppID)

	// 前端构建时的 base path（与 Vite base、APP_BASE_PATH 对应，如 "/worker/"）
	// 子路径部署时，所有请求（含 API 与静态资源）都会带 /worker/ 前缀，
	// 需在进入 mux 之前统一剥离，从而复用不带前缀的路由（/api/*）。
	appBasePath := os.Getenv("APP_BASE_PATH")
	if appBasePath == "" {
		appBasePath = "/"
	}
	if !strings.HasPrefix(appBasePath, "/") {
		appBasePath = "/" + appBasePath
	}
	if !strings.HasSuffix(appBasePath, "/") {
		appBasePath = appBasePath + "/"
	}
	// 去掉末尾 / 的前缀，用于 TrimPrefix（如 "/worker"）；根路径部署时为 ""
	prefixTrim := strings.TrimSuffix(appBasePath, "/")

	// 创建路由
	mux := http.NewServeMux()

	// 注册 API 路由
	routes.RegisterAuthRoutes(mux)
	routes.RegisterUserRoutes(mux)
	routes.RegisterClockinRoutes(mux)
	routes.RegisterRankingRoutes(mux)
	routes.RegisterPlanRoutes(mux)
	routes.RegisterPlanRankingRoutes(mux)

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
	// 注意：此时请求路径已由 basePathStripper 中间件剥离了 /worker 前缀，
	// 因此这里只需要按原始 dist 目录处理 "/" 开头的路径即可。
	exe, _ := os.Executable()
	distDir := filepath.Join(filepath.Dir(exe), "dist")

	logger.Info("📁 静态资源目录: %s", distDir)
	logger.Info("🛣️  前端 base path: %s", appBasePath)

	if _, err := os.Stat(distDir); err == nil {
		fileServer := http.FileServer(http.Dir(distDir))

		// SPA 回退：非 API 且非静态文件的请求返回 index.html
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// 检查文件是否存在（仅对真实文件提供服务，目录交给 SPA 回退处理）
			fsPath := filepath.Join(distDir, r.URL.Path)
			if info, err := os.Stat(fsPath); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
			// SPA 回退
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		})
	}

	// 中间件链：basePathStripper -> cors -> requestLogger -> mux
	// basePathStripper 必须放在最外层，确保进入 mux 前路径已被规范化
	handler := basePathStripper(prefixTrim, mux)
	handler = corsMiddleware(handler)
	handler = middleware.RequestLogger(handler)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("🚀 打工人打卡服务已启动: http://localhost%s%s", addr, appBasePath)

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

// basePathStripper 在请求进入业务路由之前统一剥离子路径前缀。
// 例如 APP_BASE_PATH=/worker/ 时，"/worker/api/xxx" 会被改写为 "/api/xxx"，
// "/worker/assets/x.js" 会被改写为 "/assets/x.js"，"/worker" 或 "/worker/" 都会变成 "/"。
// 若 prefix 为空（根路径部署），则原样透传。
func basePathStripper(prefix string, next http.Handler) http.Handler {
	if prefix == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			newPath := strings.TrimPrefix(p, prefix)
			if newPath == "" {
				newPath = "/"
			}
			r2 := r.Clone(r.Context())
			r2.URL.Path = newPath
			// RawPath 保持一致以避免某些中间件读取到不一致的编码路径
			if r.URL.RawPath != "" {
				r2.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, prefix)
				if r2.URL.RawPath == "" {
					r2.URL.RawPath = "/"
				}
			}
			next.ServeHTTP(w, r2)
			return
		}
		next.ServeHTTP(w, r)
	})
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
