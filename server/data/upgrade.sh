#!/bin/bash
# ============================================================
# MyWorker 数据库升级脚本 (v2 + v3)
# 使用方式: bash upgrade.sh [数据库路径]
# 默认数据库路径: ./clockin.db
# ============================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 数据库路径（支持传参，默认当前目录下的 clockin.db）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DB_PATH="${1:-${SCRIPT_DIR}/clockin.db}"

echo -e "${BLUE}============================================${NC}"
echo -e "${BLUE}  MyWorker 数据库升级脚本 (v2 + v3)${NC}"
echo -e "${BLUE}============================================${NC}"
echo ""

# 检查 sqlite3 是否可用
if ! command -v sqlite3 &> /dev/null; then
    echo -e "${RED}[错误] 未找到 sqlite3 命令，请先安装 sqlite3${NC}"
    exit 1
fi

# 检查数据库文件是否存在
if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}[错误] 数据库文件不存在: ${DB_PATH}${NC}"
    echo -e "${YELLOW}用法: bash upgrade.sh [数据库路径]${NC}"
    exit 1
fi

echo -e "${GREEN}[信息] 数据库路径: ${DB_PATH}${NC}"

# ==================== 备份 ====================
BACKUP_PATH="${DB_PATH}.backup_$(date +%Y%m%d_%H%M%S)"
echo -e "${YELLOW}[步骤 1/4] 备份数据库...${NC}"
cp "$DB_PATH" "$BACKUP_PATH"
echo -e "${GREEN}  ✅ 备份完成: ${BACKUP_PATH}${NC}"
echo ""

# ==================== 升级 v2 ====================
echo -e "${YELLOW}[步骤 2/4] 检查并执行 v2 升级...${NC}"

# 检查 users 表是否已有 level 字段
HAS_LEVEL=$(sqlite3 "$DB_PATH" "PRAGMA table_info(users);" | grep -c "level" || true)

if [ "$HAS_LEVEL" -gt 0 ]; then
    echo -e "${GREEN}  ⏭️  v2 已执行过（users.level 字段已存在），跳过${NC}"
else
    echo -e "${BLUE}  ▶ 执行 v2 升级...${NC}"

    # 添加 users 表字段
    sqlite3 "$DB_PATH" "ALTER TABLE users ADD COLUMN level INTEGER DEFAULT 1;"
    echo -e "${GREEN}    ✅ 添加 users.level 字段${NC}"

    sqlite3 "$DB_PATH" "ALTER TABLE users ADD COLUMN exp INTEGER DEFAULT 0;"
    echo -e "${GREEN}    ✅ 添加 users.exp 字段${NC}"

    sqlite3 "$DB_PATH" "ALTER TABLE users ADD COLUMN total_exp INTEGER DEFAULT 0;"
    echo -e "${GREEN}    ✅ 添加 users.total_exp 字段${NC}"

    # 创建 exp_logs 表
    sqlite3 "$DB_PATH" "
    CREATE TABLE IF NOT EXISTS exp_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        amount INTEGER NOT NULL,
        reason TEXT NOT NULL DEFAULT '',
        source_type TEXT NOT NULL DEFAULT '',
        source_id TEXT DEFAULT '',
        created_at TEXT DEFAULT (datetime('now', 'localtime')),
        FOREIGN KEY (user_id) REFERENCES users(id)
    );"
    echo -e "${GREEN}    ✅ 创建 exp_logs 表${NC}"

    # 创建 user_achievements 表
    sqlite3 "$DB_PATH" "
    CREATE TABLE IF NOT EXISTS user_achievements (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        achievement_id TEXT NOT NULL,
        unlocked_at TEXT DEFAULT (datetime('now', 'localtime')),
        FOREIGN KEY (user_id) REFERENCES users(id),
        UNIQUE(user_id, achievement_id)
    );"
    echo -e "${GREEN}    ✅ 创建 user_achievements 表${NC}"

    # 创建索引
    sqlite3 "$DB_PATH" "CREATE INDEX IF NOT EXISTS idx_exp_logs_user ON exp_logs(user_id);"
    sqlite3 "$DB_PATH" "CREATE INDEX IF NOT EXISTS idx_exp_logs_source ON exp_logs(source_type, source_id);"
    sqlite3 "$DB_PATH" "CREATE INDEX IF NOT EXISTS idx_achievements_user ON user_achievements(user_id);"
    echo -e "${GREEN}    ✅ 创建索引${NC}"

    echo -e "${GREEN}  ✅ v2 升级完成${NC}"
fi
echo ""

# ==================== 升级 v3 ====================
echo -e "${YELLOW}[步骤 3/4] 检查并执行 v3 升级...${NC}"

# 检查 exp_logs 表是否已有 UNIQUE 约束
# 通过检查建表语句中是否包含 UNIQUE 来判断
HAS_UNIQUE=$(sqlite3 "$DB_PATH" "SELECT sql FROM sqlite_master WHERE type='table' AND name='exp_logs';" | grep -ci "UNIQUE" || true)

if [ "$HAS_UNIQUE" -gt 0 ]; then
    echo -e "${GREEN}  ⏭️  v3 已执行过（exp_logs UNIQUE 约束已存在），跳过${NC}"
else
    echo -e "${BLUE}  ▶ 执行 v3 升级...${NC}"

    # 清理重复数据
    echo -e "${BLUE}    ▶ 清理重复数据...${NC}"
    sqlite3 "$DB_PATH" "
    DELETE FROM exp_logs WHERE id NOT IN (
        SELECT MIN(id) FROM exp_logs
        WHERE source_id IS NOT NULL AND source_id != ''
        GROUP BY user_id, source_type, source_id
    ) AND source_id IS NOT NULL AND source_id != '';
    "
    echo -e "${GREEN}    ✅ 重复数据已清理${NC}"

    # 重建表（添加 UNIQUE 约束）
    echo -e "${BLUE}    ▶ 重建 exp_logs 表...${NC}"
    sqlite3 "$DB_PATH" "
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

    INSERT INTO exp_logs (id, user_id, amount, reason, source_type, source_id, created_at)
    SELECT id, user_id, amount, reason, source_type, source_id, created_at FROM exp_logs_old;

    DROP TABLE exp_logs_old;
    "
    echo -e "${GREEN}    ✅ exp_logs 表重建完成（已添加 UNIQUE 约束）${NC}"

    # 重建索引
    sqlite3 "$DB_PATH" "CREATE INDEX IF NOT EXISTS idx_exp_logs_user ON exp_logs(user_id);"
    sqlite3 "$DB_PATH" "CREATE INDEX IF NOT EXISTS idx_exp_logs_source ON exp_logs(source_type, source_id);"
    echo -e "${GREEN}    ✅ 索引重建完成${NC}"

    echo -e "${GREEN}  ✅ v3 升级完成${NC}"
fi
echo ""

# ==================== 验证 ====================
echo -e "${YELLOW}[步骤 4/4] 验证升级结果...${NC}"

echo -e "${BLUE}  📋 users 表结构:${NC}"
sqlite3 "$DB_PATH" "PRAGMA table_info(users);" | while IFS='|' read -r cid name type notnull dflt pk; do
    echo -e "     ${name} (${type}, default: ${dflt})"
done

echo ""
echo -e "${BLUE}  📋 exp_logs 表结构:${NC}"
sqlite3 "$DB_PATH" "SELECT sql FROM sqlite_master WHERE type='table' AND name='exp_logs';" | sed 's/^/     /'

echo ""
echo -e "${BLUE}  📋 user_achievements 表结构:${NC}"
sqlite3 "$DB_PATH" "SELECT sql FROM sqlite_master WHERE type='table' AND name='user_achievements';" | sed 's/^/     /'

echo ""
echo -e "${BLUE}  📊 数据统计:${NC}"
USER_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users;")
EXP_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM exp_logs;")
ACH_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM user_achievements;")
echo -e "     用户数: ${USER_COUNT}"
echo -e "     经验日志数: ${EXP_COUNT}"
echo -e "     成就记录数: ${ACH_COUNT}"

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  ✅ 所有升级已完成！${NC}"
echo -e "${GREEN}  📦 备份文件: ${BACKUP_PATH}${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "${YELLOW}提示: 如需回滚，执行以下命令:${NC}"
echo -e "  cp ${BACKUP_PATH} ${DB_PATH}"
