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
      <button class="period-btn" :class="{ active: period === 'year' }" @click="switchPeriod('year')">
        📆 全年
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

    <!-- 工时图表 -->
    <div class="chart-section glass-card" v-if="stats && chartData.length > 0">
      <div class="chart-header">
        <h3 class="chart-title">📊 {{ period === 'year' ? '月度工时' : '每日工时' }}</h3>
        <div class="chart-legend">
          <span class="legend-item"><span class="legend-dot legend-overtime"></span>加班</span>
          <span class="legend-item"><span class="legend-dot legend-normal"></span>达标</span>
          <span class="legend-item"><span class="legend-dot legend-short"></span>不足</span>
        </div>
      </div>
      <div class="chart-body">
        <!-- Y轴 -->
        <div class="chart-y-axis">
          <span class="y-label font-mono" v-for="tick in yTicks" :key="tick" :style="{ bottom: (tick / maxY * 100) + '%' }">{{ tick }}h</span>
        </div>
        <!-- 图表主体 -->
        <div class="chart-main">
          <!-- 水平参考线 -->
          <div class="grid-lines">
            <div class="grid-line" v-for="tick in yTicks" :key="'g'+tick" :style="{ bottom: (tick / maxY * 100) + '%' }"></div>
            <!-- 8h标准线 -->
            <div class="standard-line" :style="{ bottom: (8 / maxY * 100) + '%' }">
              <span class="standard-label font-mono">8h</span>
            </div>
          </div>
          <!-- 柱体区域 -->
          <div class="bars-area">
            <div class="bar-group" v-for="(item, idx) in chartData" :key="item.label"
              @mouseenter="activeBar = idx" @mouseleave="activeBar = -1"
                 @touchstart.passive="activeBar = idx" @touchend="activeBar = -1">
              <div class="bar-col">
                <div class="bar-fill" :class="[barClass(item.value), { 'bar-active': activeBar === idx }]"
                     :style="{ height: (item.value / maxY * 100) + '%', transitionDelay: (idx * 30) + 'ms' }">
                </div>
                <!-- 悬浮提示 -->
                <transition name="tooltip-fade">
                  <div class="bar-tooltip glass-card" v-if="activeBar === idx">
                    <span class="tooltip-label">{{ item.fullLabel }}</span>
                    <span class="tooltip-value font-mono">{{ item.value.toFixed(1) }}h</span>
                  </div>
                </transition>
              </div>
              <span class="bar-label font-mono" :class="{ 'label-active': activeBar === idx }">{{ item.label }}</span>
            </div>
          </div>
        </div>
      </div>
      <!-- 图表底部摘要 -->
      <div class="chart-summary">
        <span class="summary-item">平均 <strong class="font-mono">{{ stats.avgHours }}h</strong></span>
        <span class="summary-divider">·</span>
        <span class="summary-item">最高 <strong class="font-mono">{{ maxDuration.toFixed(1) }}h</strong></span>
        <span class="summary-divider">·</span>
        <span class="summary-item">共 <strong class="font-mono">{{ chartData.length }}</strong> {{ period === 'year' ? '个月' : '天' }}</span>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state glass-card" v-if="stats && chartData.length === 0">
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
import { ref, computed, onMounted } from 'vue'
import { clockinAPI } from '@/services/clockin.api.js'

export default {
  name: 'ClockInStats',
  setup() {
    const period = ref('week')
    const stats = ref(null)
    const loading = ref(false)
    const activeBar = ref(-1)

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
      activeBar.value = -1
      fetchStats()
    }

    // 图表数据：全年模式按月聚合，其他模式按天
    const chartData = computed(() => {
      if (!stats.value || !stats.value.records) return []
      const records = stats.value.records

      if (period.value === 'year') {
        // 按月聚合
        const monthMap = {}
        records.forEach(r => {
          const month = r.date.slice(0, 7) // YYYY-MM
          if (!monthMap[month]) monthMap[month] = { total: 0, count: 0 }
          monthMap[month].total += r.duration
          monthMap[month].count++
        })
        const months = ['01','02','03','04','05','06','07','08','09','10','11','12']
        const year = new Date().getFullYear()
        return months
          .filter(m => monthMap[year + '-' + m])
          .map(m => {
            const key = year + '-' + m
            const d = monthMap[key]
            return {
              label: parseInt(m) + '月',
              fullLabel: key,
              value: Math.round(d.total / d.count * 100) / 100 // 月均日工时
            }
          })
      }

      return records.map(r => ({
        label: r.date.slice(5), // MM-DD
        fullLabel: r.date,
        value: r.duration
      }))
    })

    // 最大工时（用于Y轴缩放）
    const maxDuration = computed(() => {
      if (!chartData.value.length) return 0
      return Math.max(...chartData.value.map(d => d.value))
    })

    // Y轴最大值（向上取整到偶数）
    const maxY = computed(() => {
      const m = maxDuration.value
      if (m <= 10) return 12
      return Math.ceil(m / 2) * 2 + 2
    })

    // Y轴刻度
    const yTicks = computed(() => {
      const ticks = []
      const step = maxY.value <= 12 ? 2 : 4
      for (let i = step; i <= maxY.value; i += step) {
        ticks.push(i)
      }
      return ticks
    })

    const barClass = (duration) => {
      if (duration >= 10) return 'bar-overtime'
      if (duration >= 8) return 'bar-normal'
      return 'bar-short'
    }

    onMounted(fetchStats)

    return {
      period, stats, loading, activeBar,
      switchPeriod, barClass,
      chartData, maxDuration, maxY, yTicks
    }
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

/* ===== 图表区域 ===== */
.chart-section {
  padding: var(--space-6);
  overflow: hidden;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-5);
}

.chart-title {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--text-primary);
}

.chart-legend {
  display: flex;
  gap: var(--space-4);
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 2px;
}

.legend-overtime { background: linear-gradient(135deg, #F472B6, #A855F7); }
.legend-normal { background: linear-gradient(135deg, #818CF8, #6366F1); }
.legend-short { background: rgba(168, 85, 247, 0.25); }

/* 图表主体布局 */
.chart-body {
  display: flex;
  gap: 0;
  height: 220px;
  position: relative;
}

/* Y轴 */
.chart-y-axis {
  width: 36px;
  position: relative;
  flex-shrink: 0;
}

.y-label {
  position: absolute;
  right: 6px;
  transform: translateY(50%);
  font-size: 10px;
  color: var(--text-muted);
  line-height: 1;
}

/* 图表绘制区 */
.chart-main {
  flex: 1;
  position: relative;
  border-left: 1px solid rgba(168, 85, 247, 0.15);
  border-bottom: 1px solid rgba(168, 85, 247, 0.15);
}

/* 网格线 */
.grid-lines {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.grid-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 1px;
  background: rgba(168, 85, 247, 0.06);
}

/* 8h标准线 */
.standard-line {
  position: absolute;
  left: 0;
  right: 0;
  height: 1px;
  background: repeating-linear-gradient(
    90deg,
    rgba(99, 102, 241, 0.4) 0px,
    rgba(99, 102, 241, 0.4) 4px,
    transparent 4px,
    transparent 8px
  );
  z-index: 2;
}

.standard-label {
  position: absolute;
  right: 4px;
  top: -14px;
  font-size: 9px;
  color: rgba(99, 102, 241, 0.7);
  background: var(--bg-secondary);
  padding: 0 3px;
  border-radius: 2px;
}

/* 柱体区域 */
.bars-area {
  display: flex;
  align-items: flex-end;
  height: 100%;
  padding: 0 8px;
  gap: 2px;
  position: relative;
  z-index: 3;
}

.bar-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 100%;
  min-width: 0;
  cursor: pointer;
}

.bar-col {
  flex: 1;
  width: 100%;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  position: relative;
}

.bar-fill {
  width: 60%;
  max-width: 36px;
  min-width: 6px;
  min-height: 3px;
  border-radius: 4px 4px 1px 1px;
  transition: height 0.6s cubic-bezier(0.34, 1.56, 0.64, 1), filter 0.2s ease, transform 0.2s ease;
  position: relative;
}

.bar-fill.bar-active {
  filter: brightness(1.3);
  transform: scaleX(1.1);
}

.bar-fill.bar-normal {
  background: linear-gradient(180deg, #818CF8 0%, #6366F1 100%);
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
}

.bar-fill.bar-overtime {
  background: linear-gradient(180deg, #F472B6 0%, #A855F7 100%);
  box-shadow: 0 2px 8px rgba(168, 85, 247, 0.3);
}

.bar-fill.bar-short {
  background: linear-gradient(180deg, rgba(168, 85, 247, 0.35) 0%, rgba(168, 85, 247, 0.15) 100%);
}

/* 悬浮提示 */
.bar-tooltip {
  position: absolute;
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
  padding: 6px 10px;
  border-radius: 8px;
  white-space: nowrap;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  z-index: 10;
  pointer-events: none;
  border: 1px solid rgba(168, 85, 247, 0.2) !important;
  background: var(--bg-secondary) !important;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.bar-tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 5px solid transparent;
  border-top-color: rgba(168, 85, 247, 0.2);
}

.tooltip-label {
  font-size: 10px;
  color: var(--text-muted);
}

.tooltip-value {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-primary-light);
}

.tooltip-fade-enter-active,
.tooltip-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.tooltip-fade-enter-from,
.tooltip-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(4px);
}

/* 柱体下方标签 */
.bar-label {
  font-size: 9px;
  color: var(--text-muted);
  padding-top: 6px;
  transition: color 0.2s ease;
  text-align: center;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bar-label.label-active {
  color: var(--color-primary-light);
}

/* 图表底部摘要 */
.chart-summary {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-3);
  margin-top: var(--space-5);
  padding-top: var(--space-4);
  border-top: 1px solid rgba(168, 85, 247, 0.08);
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.summary-item strong {
  color: var(--color-primary-light);
}

.summary-divider {
  color: rgba(168, 85, 247, 0.2);
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
  .chart-section { padding: var(--space-4); }
  .chart-body { height: 180px; }
  .chart-legend { display: none; }
  .chart-header { margin-bottom: var(--space-3); }
  .bar-fill { max-width: 24px; }
  .bars-area { padding: 0 2px; gap: 1px; }
  /* 移动端图表标签：天数多时隔一个显示，避免重叠 */
  .bar-group:nth-child(odd) .bar-label {
    visibility: visible;
  }
  .bar-group:nth-child(even) .bar-label {
    visibility: hidden;
  }
  /* 移动端 tooltip 触摸支持 */
  .bar-tooltip {
    pointer-events: auto;
  }
  /* 统计卡片间距缩小 */
  .stats-grid { gap: var(--space-3); }
  .stat-card { padding: var(--space-3); }
  .stat-card:hover { transform: none; }
  .stat-icon { font-size: 1.2rem; }
  .stat-value { font-size: var(--text-lg); }
}
</style>
