-- ==================== SQLite 升级语句 v3 ====================
-- 为 exp_logs 表添加 UNIQUE(user_id, source_type, source_id) 约束
-- SQLite 不支持 ALTER TABLE ADD CONSTRAINT，需要重建表

-- 1. 先清理可能存在的重复数据（保留最早的记录）
DELETE FROM exp_logs WHERE id NOT IN (
    SELECT MIN(id) FROM exp_logs
    WHERE source_id IS NOT NULL AND source_id != ''
    GROUP BY user_id, source_type, source_id
) AND source_id IS NOT NULL AND source_id != '';

-- 2. 重建表
ALTER TABLE exp_logs RENAME TO exp_logs_old;

CREATE TABLE exp_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT '',
    source_id TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now', 'localtime')),
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE(user_id, source_type, source_id)
);

-- 3. 迁移数据
INSERT INTO exp_logs (id, user_id, amount, reason, source_type, source_id, created_at)
SELECT id, user_id, amount, reason, source_type, source_id, created_at FROM exp_logs_old;

-- 4. 删除旧表
DROP TABLE exp_logs_old;

-- 5. 重建索引
CREATE INDEX IF NOT EXISTS idx_exp_logs_user ON exp_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_exp_logs_source ON exp_logs(source_type, source_id);
