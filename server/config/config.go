package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	// 服务端口
	Port int
	// 用户中心服务地址
	UserCenterURL string
	// 本服务在用户中心注册的租户标识
	AppID string

	// ── 日志配置 ──
	// 日志目录（默认为启动根路径/logs）
	LogDir string
	// 日志级别：debug | info | warn | error（默认 info）
	LogLevel string
	// 最多保留的日志文件数量，按天滚动（默认 1）
	LogMaxFiles int
	// 是否同时输出到控制台（默认 true）
	LogToConsole bool
}

// C 全局配置实例
var C *Config

// Load 加载配置
func Load() *Config {
	// 先尝试加载 .env 文件
	loadEnvFile()

	C = &Config{
		Port:          getEnvInt("PORT", 8008),
		UserCenterURL: getEnv("USERCENTER_BASE_URL", "http://localhost:4000"),
		AppID:         getEnv("USERCENTER_APP_ID", "worker"),
		LogDir:        getEnv("LOG_DIR", ""),
		LogLevel:      strings.ToLower(getEnv("LOG_LEVEL", "info")),
		LogMaxFiles:   getEnvInt("LOG_MAX_FILES", 1),
		LogToConsole:  getEnvBool("LOG_TO_CONSOLE", true),
	}

	// 日志目录默认值：启动根路径/logs
	if C.LogDir == "" {
		exe, err := os.Executable()
		if err == nil {
			C.LogDir = filepath.Join(filepath.Dir(exe), "logs")
		} else {
			C.LogDir = "./logs"
		}
	}

	return C
}

// loadEnvFile 简易 .env 文件加载
func loadEnvFile() {
	// 尝试从可执行文件所在目录加载
	paths := []string{".env"}
	exe, err := os.Executable()
	if err == nil {
		paths = append([]string{filepath.Join(filepath.Dir(exe), ".env")}, paths...)
	}

	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			idx := strings.Index(line, "=")
			if idx == -1 {
				continue
			}
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			// 不覆盖已有的环境变量
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
		break // 只加载第一个找到的 .env
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return strings.ToLower(val) == "true"
	}
	return defaultVal
}
