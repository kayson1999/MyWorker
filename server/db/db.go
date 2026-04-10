package db

import (
	"database/sql"
	"path/filepath"
	"sync"

	"myworker/logger"

	_ "github.com/mattn/go-sqlite3"
)

var (
	instance *sql.DB
	once     sync.Once
	dbDir    string
)

// SetDBDir 设置数据库目录
func SetDBDir(dir string) {
	dbDir = dir
}

// GetDB 获取数据库实例（单例）
func GetDB() *sql.DB {
	once.Do(func() {
		if dbDir == "" {
			dbDir = "./data"
		}
		dbPath := filepath.Join(dbDir, "clockin.db")

		var err error
		instance, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000")
		if err != nil {
			logger.Error("打开数据库失败: %v", err)
			panic(err)
		}

		// 设置连接池
		// SQLite 在 WAL 模式下支持多读单写；保留少量并发连接可避免登录等请求因单连接被占用而长时间阻塞。
		instance.SetMaxOpenConns(10)
		instance.SetMaxIdleConns(5)

		initTables()
	})
	return instance
}

// initTables 初始化数据库表
func initTables() {
	schema := `
	-- 本地用户表（业务扩展字段）
	-- id 与用户中心的 userId 保持一致，不再自增
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		nickname TEXT NOT NULL DEFAULT '',
		avatar TEXT DEFAULT '😎',
		profession TEXT DEFAULT '',
		position TEXT DEFAULT '',
		city TEXT DEFAULT '',
		standard_start TEXT DEFAULT '09:00',
		standard_end TEXT DEFAULT '18:00',
		created_at TEXT DEFAULT (datetime('now', 'localtime'))
	);

	CREATE TABLE IF NOT EXISTS clock_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		clock_in TEXT,
		clock_out TEXT,
		duration REAL DEFAULT 0,
		is_manual INTEGER DEFAULT 0,
		created_at TEXT DEFAULT (datetime('now', 'localtime')),
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE(user_id, date)
	);

	CREATE INDEX IF NOT EXISTS idx_records_user_date ON clock_records(user_id, date);
	CREATE INDEX IF NOT EXISTS idx_records_date ON clock_records(date);
	`

	_, err := instance.Exec(schema)
	if err != nil {
		logger.Error("初始化数据库表失败: %v", err)
		panic(err)
	}
}

// Close 关闭数据库连接
func Close() {
	if instance != nil {
		instance.Close()
	}
}
