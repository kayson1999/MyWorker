<template>
  <div class="usercenter-wrapper">
    <!-- 等级卡片 -->
    <div class="level-card glass-card">
      <div class="level-header">
        <div class="level-avatar-area">
          <span class="level-avatar">{{ user.avatar || '😎' }}</span>
          <div class="level-badge">{{ overview?.level_info?.icon || '🌱' }}</div>
        </div>
        <div class="level-info">
          <h3 class="level-name">{{ user.nickname }}</h3>
          <div class="level-title">
            <span class="level-tag">Lv.{{ overview?.level_info?.level || 1 }}</span>
            <span class="level-title-text">{{ overview?.level_info?.title || '实习打工人' }}</span>
          </div>
        </div>
        <button class="edit-profile-btn" @click="$emit('edit-profile')">✏️</button>
      </div>

      <!-- 经验条 -->
      <div class="exp-bar-area">
        <div class="exp-bar-bg">
          <div class="exp-bar-fill" :style="{ width: (overview?.level_info?.progress || 0) + '%' }"></div>
        </div>
        <div class="exp-bar-text font-mono">
          <span>{{ overview?.level_info?.exp || 0 }} / {{ overview?.level_info?.next_min || 50 }} EXP</span>
          <span v-if="overview?.level_info?.max_level">🎆 已满级</span>
          <span v-else>{{ overview?.level_info?.progress || 0 }}%</span>
        </div>
      </div>

      <!-- 今日状态 -->
      <div class="today-status" :class="'status-' + (overview?.today_status === '已完成' ? 'done' : overview?.today_status === '工作中' ? 'working' : 'idle')">
        <span class="status-dot"></span>
        <span>{{ overview?.today_status || '未打卡' }}</span>
      </div>
    </div>

    <!-- 核心数据卡片 -->
    <div class="stats-grid">
      <div class="stat-card glass-card">
        <span class="stat-icon">📅</span>
        <span class="stat-value font-mono">{{ overview?.stats?.total_days || 0 }}</span>
        <span class="stat-label">总打卡天数</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">⏱️</span>
        <span class="stat-value font-mono">{{ overview?.stats?.total_hours || 0 }}h</span>
        <span class="stat-label">总工时</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">📊</span>
        <span class="stat-value font-mono">{{ overview?.stats?.avg_hours || 0 }}h</span>
        <span class="stat-label">日均工时</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">🔥</span>
        <span class="stat-value font-mono">{{ overview?.stats?.streak || 0 }}天</span>
        <span class="stat-label">连续打卡</span>
      </div>
    </div>

    <!-- 工作风格标签 -->
    <div class="section glass-card">
      <h4 class="section-title">🏷️ 工作风格标签</h4>
      <div class="style-tags">
        <div class="style-tag">
          <span class="tag-period">本周</span>
          <span class="tag-value">{{ overview?.style_tags?.week || '暂无' }}</span>
        </div>
        <div class="style-tag">
          <span class="tag-period">本月</span>
          <span class="tag-value">{{ overview?.style_tags?.month || '暂无' }}</span>
        </div>
        <div class="style-tag">
          <span class="tag-period">本年</span>
          <span class="tag-value">{{ overview?.style_tags?.year || '暂无' }}</span>
        </div>
      </div>
    </div>

    <!-- 成就徽章 -->
    <div class="section glass-card">
      <h4 class="section-title">
        🏆 成就徽章
        <span class="achievement-count font-mono">{{ overview?.achievements?.unlocked || 0 }}/{{ overview?.achievements?.total || 0 }}</span>
      </h4>
      <div class="achievement-grid" v-if="achievements.length > 0">
        <div v-for="a in achievements" :key="a.id" class="achievement-item"
             :class="{ unlocked: a.unlocked, locked: !a.unlocked }"
             :title="a.description">
          <span class="achievement-icon">{{ a.unlocked ? a.icon : '🔒' }}</span>
          <span class="achievement-name">{{ a.name }}</span>
        </div>
      </div>
      <div v-else class="empty-hint">完成打卡和计划来解锁成就吧！</div>
    </div>

    <!-- 打卡热力图 -->
    <div class="section glass-card">
      <h4 class="section-title">📊 打卡热力图</h4>
  <div class="heatmap-container">
        <div class="heatmap-scroll-hint">← 左右滑动查看更多</div>
        <div class="heatmap-grid">
          <div v-for="(day, idx) in heatmapData" :key="idx"
               class="heatmap-cell"
               :class="'level-' + day.level"
               :title="day.date + (day.duration > 0 ? ' · ' + day.duration + 'h' : '')">
          </div>
        </div>
        <div class="heatmap-legend">
          <span class="legend-label">少</span>
          <span class="heatmap-cell level-0 legend-cell"></span>
          <span class="heatmap-cell level-1 legend-cell"></span>
          <span class="heatmap-cell level-2 legend-cell"></span>
          <span class="heatmap-cell level-3 legend-cell"></span>
          <span class="heatmap-cell level-4 legend-cell"></span>
          <span class="legend-label">多</span>
        </div>
      </div>
    </div>

    <!-- 最近经验值记录 -->
    <div class="section glass-card">
      <h4 class="section-title">⚡ 最近经验值</h4>
      <div class="exp-log-list" v-if="expLogs.length > 0">
        <div v-for="(log, idx) in expLogs.slice(0, 10)" :key="idx" class="exp-log-item">
          <span class="exp-log-reason">{{ log.reason }}</span>
          <span class="exp-log-amount font-mono">+{{ log.amount }} EXP</span>
          <span class="exp-log-time">{{ formatTime(log.created_at) }}</span>
        </div>
      </div>
      <div v-else class="empty-hint">暂无经验值记录，快去打卡吧！</div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { userCenterAPI } from '@/services/clockin.api.js'

export default {
  name: 'UserCenter',
  props: {
    user: { type: Object, required: true }
  },
  emits: ['edit-profile'],
  setup() {
    const overview = ref(null)
    const achievements = ref([])
    const expLogs = ref([])
    const heatmapData = ref([])

    // 获取总览数据
    const fetchOverview = async () => {
      try {
        const data = await userCenterAPI.getOverview()
        overview.value = data
      } catch (err) {
        console.error('获取个人中心数据失败:', err)
      }
    }

    // 获取成就列表
    const fetchAchievements = async () => {
      try {
        const data = await userCenterAPI.getAchievements()
        achievements.value = data.achievements || []
      } catch (err) {
        console.error('获取成就失败:', err)
      }
    }

    // 获取经验值日志
    const fetchExpLogs = async () => {
      try {
        const data = await userCenterAPI.getExpLogs()
        expLogs.value = data.logs || []
      } catch (err) {
        console.error('获取经验值日志失败:', err)
      }
    }

    // 获取热力图数据
    const fetchHeatmap = async () => {
      try {
        const data = await userCenterAPI.getHeatmap()
        // 补全过去一年的每一天
        const heatmapMap = {}
        for (const item of (data.heatmap || [])) {
          heatmapMap[item.date] = item
        }

        const days = []
        const today = new Date()
        for (let i = 364; i >= 0; i--) {
          const d = new Date(today)
          d.setDate(d.getDate() - i)
          const dateStr = d.toISOString().slice(0, 10)
          days.push(heatmapMap[dateStr] || { date: dateStr, duration: 0, level: 0 })
        }
        heatmapData.value = days
      } catch (err) {
        console.error('获取热力图失败:', err)
      }
    }

    // 格式化时间
    const formatTime = (timeStr) => {
      if (!timeStr) return ''
      const d = new Date(timeStr.replace(' ', 'T'))
      const now = new Date()
      const diff = now - d
      if (diff < 60000) return '刚刚'
      if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
      if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前'
      return timeStr.slice(5, 16)
    }

    onMounted(() => {
      fetchOverview()
      fetchAchievements()
      fetchExpLogs()
      fetchHeatmap()
    })

    return { overview, achievements, expLogs, heatmapData, formatTime }
  }
}
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }

.usercenter-wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

/* ==================== 等级卡片 ==================== */
.level-card {
  padding: var(--space-6);
}

.level-header {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-bottom: var(--space-5);
}

.level-avatar-area {
  position: relative;
}

.level-avatar {
  font-size: 2.5rem;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  border: 2px solid rgba(168, 85, 247, 0.3);
}

.level-badge {
  position: absolute;
  bottom: -4px;
  right: -4px;
  font-size: 1rem;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  border-radius: 50%;
  border: 2px solid var(--color-primary);
}

.level-info {
  flex: 1;
}

.level-name {
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--text-primary);
}

.level-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-1);
}

.level-tag {
  padding: 1px 8px;
  background: var(--gradient-primary);
  color: white;
  border-radius: var(--radius-full);
  font-size: var(--text-xs);
  font-weight: 700;
  font-family: var(--font-mono);
}

.level-title-text {
  font-size: var(--text-sm);
  color: var(--color-primary-light);
  font-weight: 500;
}

.edit-profile-btn {
  padding: var(--space-2) var(--space-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.3s ease;
}

.edit-profile-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

/* 经验条 */
.exp-bar-area {
  margin-bottom: var(--space-4);
}

.exp-bar-bg {
  height: 8px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-full);
  overflow: hidden;
  border: 1px solid var(--border-color);
}

.exp-bar-fill {
  height: 100%;
  background: var(--gradient-primary);
  border-radius: var(--radius-full);
  transition: width 0.8s ease;
  min-width: 2px;
}

.exp-bar-text {
  display: flex;
  justify-content: space-between;
  font-size: var(--text-xs);
  color: var(--text-muted);
  margin-top: var(--space-1);
}

/* 今日状态 */
.today-status {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--text-muted);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  width: fit-content;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);
}

.status-done .status-dot { background: #22c55e; }
.status-done { color: #22c55e; }
.status-working .status-dot { background: var(--color-primary-light); animation: glow-pulse 1.5s ease-in-out infinite; }
.status-working { color: var(--color-primary-light); }

/* ==================== 核心数据 ==================== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-3);
}

.stat-card {
  padding: var(--space-4);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
}

.stat-icon { font-size: 1.5rem; }
.stat-value {
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-primary-light);
}
.stat-label {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

/* ==================== 通用 Section ==================== */
.section {
  padding: var(--space-5);
}

.section-title {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--space-4);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.achievement-count {
  margin-left: auto;
  font-size: var(--text-xs);
  color: var(--text-muted);
}

/* ==================== 工作风格标签 ==================== */
.style-tags {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-3);
}

.style-tag {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.tag-period {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.tag-value {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary-light);
  text-align: center;
  word-break: break-all;
}

/* ==================== 成就徽章 ==================== */
.achievement-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
  gap: var(--space-3);
}

.achievement-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-1);
  padding: var(--space-3);
  border-radius: var(--radius-md);
  transition: all 0.3s ease;
  cursor: default;
}

.achievement-item.unlocked {
  background: rgba(168, 85, 247, 0.08);
  border: 1px solid rgba(168, 85, 247, 0.15);
}

.achievement-item.locked {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  opacity: 0.5;
}

.achievement-item.unlocked:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
}

.achievement-icon { font-size: 1.5rem; }
.achievement-name {
  font-size: 10px;
  color: var(--text-secondary);
  text-align: center;
  line-height: 1.3;
}

/* ==================== 热力图 ==================== */
.heatmap-container {
  overflow-x: auto;
  position: relative;
  /* 渐隐边缘提示可滑动 */
  mask-image: linear-gradient(to right, black 90%, transparent 100%);
  -webkit-mask-image: linear-gradient(to right, black 90%, transparent 100%);
}

.heatmap-scroll-hint {
  display: none;
  font-size: 10px;
  color: var(--text-muted);
  text-align: center;
  margin-bottom: var(--space-2);
}

.heatmap-grid {
  display: grid;
  grid-template-rows: repeat(7, 1fr);
  grid-auto-flow: column;
  gap: 3px;
  min-width: 700px;
}

.heatmap-cell {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  background: var(--bg-tertiary);
  border: 1px solid rgba(168, 85, 247, 0.05);
}

.heatmap-cell.level-1 { background: rgba(168, 85, 247, 0.15); }
.heatmap-cell.level-2 { background: rgba(168, 85, 247, 0.35); }
.heatmap-cell.level-3 { background: rgba(168, 85, 247, 0.6); }
.heatmap-cell.level-4 { background: rgba(168, 85, 247, 0.9); }

.heatmap-legend {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  margin-top: var(--space-3);
  justify-content: flex-end;
}

.legend-label {
  font-size: 10px;
  color: var(--text-muted);
}

.legend-cell {
  width: 12px;
  height: 12px;
}

/* ==================== 经验值日志 ==================== */
.exp-log-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.exp-log-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
}

.exp-log-reason {
  flex: 1;
  color: var(--text-secondary);
}

.exp-log-amount {
  color: #22c55e;
  font-weight: 600;
  font-size: var(--text-xs);
}

.exp-log-time {
  font-size: var(--text-xs);
  color: var(--text-muted);
  min-width: 70px;
  text-align: right;
}

.empty-hint {
  text-align: center;
  color: var(--text-muted);
  font-size: var(--text-sm);
  padding: var(--space-6);
}

/* ==================== 响应式 ==================== */
@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .style-tags { grid-template-columns: 1fr; }
  .achievement-grid { grid-template-columns: repeat(auto-fill, minmax(75px, 1fr)); }
  .level-header { flex-wrap: wrap; }
  .heatmap-grid { min-width: 500px; }
  .heatmap-cell { width: 10px; height: 10px; }
  .heatmap-scroll-hint { display: block; }
  .heatmap-container {
    mask-image: linear-gradient(to right, black 85%, transparent 100%);
    -webkit-mask-image: linear-gradient(to right, black 85%, transparent 100%);
  }
  .exp-log-item { flex-wrap: wrap; gap: var(--space-2); }
  .exp-log-time { min-width: auto; }
}
</style>
