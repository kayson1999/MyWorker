package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志级别定义
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var levelNames = map[int]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

var levelMap = map[string]int{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// Logger 日志实例
type Logger struct {
	mu          sync.Mutex
	level       int
	logDir      string
	maxFiles    int
	toConsole   bool
	currentDate string
	fileWriter  *lumberjack.Logger
	multiWriter io.Writer
}

// 全局日志实例
var defaultLogger *Logger

// LogConfig 日志配置
type LogConfig struct {
	LogDir    string
	LogLevel  string
	MaxFiles  int
	ToConsole bool
}

// Init 初始化全局日志
func Init(cfg LogConfig) {
	level, ok := levelMap[strings.ToLower(cfg.LogLevel)]
	if !ok {
		level = LevelInfo
	}

	defaultLogger = &Logger{
		level:     level,
		logDir:    cfg.LogDir,
		maxFiles:  cfg.MaxFiles,
		toConsole: cfg.ToConsole,
	}

	// 确保日志目录存在
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		log.Printf("[Logger] 创建日志目录失败: %v", err)
	}

	// 初始化文件写入器
	defaultLogger.rotateIfNeeded()
}

// rotateIfNeeded 检查是否需要切换日志文件（按天）
func (l *Logger) rotateIfNeeded() {
	today := time.Now().Format("2006-01-02")
	if l.currentDate == today && l.fileWriter != nil {
		return
	}

	l.currentDate = today
	logFile := filepath.Join(l.logDir, fmt.Sprintf("app-%s.log", today))

	// 关闭旧的写入器
	if l.fileWriter != nil {
		l.fileWriter.Close()
	}

	l.fileWriter = &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // MB，单个文件最大 100MB
		MaxBackups: 0,   // 不限制备份数量，由我们自己管理
		LocalTime:  true,
		Compress:   false,
	}

	// 构建多路写入器
	writers := []io.Writer{l.fileWriter}
	if l.toConsole {
		writers = append(writers, os.Stdout)
	}
	l.multiWriter = io.MultiWriter(writers...)

	// 清理过期日志
	l.cleanOldLogs()
}

// cleanOldLogs 清理过期日志文件
func (l *Logger) cleanOldLogs() {
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return
	}

	var logFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "app-") && strings.HasSuffix(name, ".log") {
			logFiles = append(logFiles, name)
		}
	}

	// 按文件名排序（日期排序），最新的在后
	sort.Strings(logFiles)

	// 保留最新的 maxFiles 个文件，删除其余的
	if len(logFiles) > l.maxFiles {
		toDelete := logFiles[:len(logFiles)-l.maxFiles]
		for _, f := range toDelete {
			os.Remove(filepath.Join(l.logDir, f))
		}
	}
}

// getCallerInfo 获取调用者信息（函数名、文件名、行号）
func getCallerInfo(skip int) (funcName, fileName string, line int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", "unknown", 0
	}

	// 获取函数名（去掉包路径前缀，只保留函数名）
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		fullName := fn.Name()
		// 取最后一个 . 后面的部分作为函数名
		if idx := strings.LastIndex(fullName, "."); idx >= 0 {
			funcName = fullName[idx+1:]
		} else {
			funcName = fullName
		}
	} else {
		funcName = "unknown"
	}

	// 只保留文件名，不要完整路径
	fileName = filepath.Base(file)
	return
}

// log 核心日志写入
func (l *Logger) log(level int, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 检查是否需要切换日志文件
	l.rotateIfNeeded()

	// 获取调用者信息（skip=3: log -> Debug/Info/Warn/Error -> 实际调用者）
	funcName, fileName, line := getCallerInfo(3)

	// 格式化时间
	now := time.Now()
	ts := now.Format("2006-01-02 15:04:05.000")

	// 格式化消息
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	// 组装日志行
	levelTag := levelNames[level]
	callerTag := fmt.Sprintf("%s:%d@%s", fileName, line, funcName)
	logLine := fmt.Sprintf("[%s] [%-5s] [%s] %s\n", ts, levelTag, callerTag, msg)

	// 写入
	if l.multiWriter != nil {
		l.multiWriter.Write([]byte(logLine))
	}
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelDebug, format, args...)
	}
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelInfo, format, args...)
	}
}

// Warn 警告日志
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelWarn, format, args...)
	}
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelError, format, args...)
	}
}
