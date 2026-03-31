<template>
  <div class="stats-wrapper">
    <!-- 周期切换 -->
    <div class="period-switch">
      <button class="period-btn" :class="{ active: period === 'week' }" @click="switchPeriod('week')">
        📅 本周
      </button>
      <button class="period-btn" :class="{ active: period === 'month' }" @click="switchPeriod('month')">
        🗓️ 本月
      </button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid" v-if="stats">
      <div class="stat-card glass-card">
        <span class="stat-icon">⏱️</span>
        <span class="stat-value font-mono">{{ stats.totalHours }}h</span>
        <span class="stat-label">总工时</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">📊</span>
        <span class="stat-value font-mono">{{ stats.avgHours }}h</span>
        <span class="stat-label">日均工时</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">📆</span>
        <span class="stat-value font-mono">{{ stats.totalDays }}天</span>
        <span class="stat-label">打卡天数</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">🔥</span>
        <span class="stat-value font-mono">{{ stats.streak }}天</span>
        <span class="stat-label">连续打卡</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">🌅</span>
        <span class="stat-value font-mono">{{ stats.earliestIn || '--:--' }}</span>
        <span class="stat-label">最早上班</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">🌙</span>
        <span class="stat-value font-mono">{{ stats.latestOut || '--:--' }}</span>
        <span class="stat-label">最晚下班</span>
      </div>
      <div class="stat-card glass-card">
        <span class="stat-icon">🎯</span>
        <span class="stat-value font-mono">{{ stats.onTimeRate }}%</span>
        <span class="stat-label">准时率</span>
      </div>
    </div>

    <!-- 工时柱状图 -->
    <div class="chart-section glass-card" v-if="stats && stats.records.length > 0">
      <h3 class="chart-title">📊 每日工时</h3>
      <div class="bar-chart">
        <div class="bar-item" v-for="record in stats.records" :key="record.date">
          <div class="bar-container">
            <div class="bar" :style="{ height: barHeight(record.duration) + '%' }"
                 :class="barClass(record.duration)">
              <span class="bar-value font-mono" v-if="record.duration > 0">
                {{ record.duration.toFixed(1) }}
              </span>
            </div>
          </div>
          <span class="bar-label font-mono">{{ formatDate(record.date) }}</span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state glass-card" v-if="stats && stats.records.length === 0">
      <span class="empty-icon">📭</span>
      <p>暂无打卡记录，快去打卡吧！</p>
    </div>

    <!-- 加载中 -->
    <div class="loading-state" v-if="loading">
      <span class="loading-icon">⏳</span>
      <p>加载中...</p>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { clockinAPI } from '@/services/clockin.api.js'

export default {
  name: 'ClockInStats',
  setup() {
    const period = ref('week')
    const stats = ref(null)
    const loading = ref(false)

    const fetchStats = async () => {
      loading.value = true
      try {
        const data = await clockinAPI.getStats(period.value)
        stats.value = data
      } catch (err) {
        console.error('获取统计失败:', err)
      } finally {
        loading.value = false
      }
    }

    const switchPeriod = (p) => {
      period.value = p
      fetchStats()
    }

    const barHeight = (duration) => {
      if (!duration) return 0
      // 最大12小时为100%
      return Math.min(100, (duration / 12) * 100)
    }

    const barClass = (duration) => {
      if (duration >= 10) return 'bar-overtime'
      if (duration >= 8) return 'bar-normal'
      return 'bar-short'
    }

    const formatDate = (dateStr) => {
      return dateStr.slice(5) // MM-DD
    }

    onMounted(fetchStats)

    return { period, stats, loading, switchPeriod, barHeight, barClass, formatDate }
  }
}
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }

.stats-wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* 周期切换 */
.period-switch {
  display: flex;
  gap: var(--space-3);
}

.period-btn {
  padding: var(--space-2) var(--space-5);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.3s ease;
}

.period-btn.active {
  background: var(--gradient-primary);
  color: white;
  border-color: transparent;
}

.period-btn:hover:not(.active) {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: var(--space-4);
}

.stat-card {
  padding: var(--space-5);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  border-color: var(--border-glow);
  box-shadow: var(--shadow-glow);
}

.stat-icon {
  font-size: 1.5rem;
}

.stat-value {
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--color-primary-light);
}

.stat-label {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

/* 柱状图 */
.chart-section {
  padding: var(--space-6);
}

.chart-title {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--space-5);
}

.bar-chart {
  display: flex;
  align-items: flex-end;
  gap: var(--space-3);
  height: 200px;
  padding-top: var(--space-4);
}

.bar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  height: 100%;
}

.bar-container {
  flex: 1;
  width: 100%;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.bar {
  width: 70%;
  max-width: 40px;
  border-radius: var(--radius-sm) var(--radius-sm) 0 0;
  transition: height 0.5s ease;
  position: relative;
  min-height: 4px;
}

.bar-value {
  position: absolute;
  top: -20px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 10px;
  color: var(--text-secondary);
  white-space: nowrap;
}

.bar-normal {
  background: var(--gradient-primary);
}

.bar-overtime {
  background: linear-gradient(180deg, #F472B6, #A855F7);
}

.bar-short {
  background: rgba(168, 85, 247, 0.3);
}

.bar-label {
  font-size: 10px;
  color: var(--text-muted);
}

/* 空状态 */
.empty-state {
  padding: var(--space-10);
  text-align: center;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: var(--space-3);
}

.loading-state {
  text-align: center;
  padding: var(--space-8);
  color: var(--text-muted);
}

.loading-icon {
  font-size: 2rem;
  display: block;
  margin-bottom: var(--space-2);
  animation: glow-pulse 1.5s ease-in-out infinite;
}

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .bar-chart { gap: var(--space-1); }
}
</style>
