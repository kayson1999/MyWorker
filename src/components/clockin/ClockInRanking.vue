<template>
  <div class="ranking-wrapper">
    <!-- 榜单切换 -->
    <div class="ranking-tabs">
      <button v-for="tab in tabs" :key="tab.id" class="tab-btn"
              :class="{ active: activeTab === tab.id }" @click="switchTab(tab.id)">
        {{ tab.icon }} {{ tab.label }}
      </button>
    </div>

    <!-- 周期切换 -->
    <div class="period-switch" v-if="activeTab !== 'streak'">
      <button class="period-btn" :class="{ active: period === 'week' }" @click="switchPeriod('week')">
        本周
      </button>
      <button class="period-btn" :class="{ active: period === 'month' }" @click="switchPeriod('month')">
        本月
      </button>
    </div>

    <!-- 排行列表 -->
    <div class="ranking-list" v-if="!loading && list.length > 0">
      <div class="ranking-item glass-card" v-for="item in list" :key="item.userId"
           :class="{ 'top-3': item.rank <= 3 }">
        <div class="rank-badge" :class="'rank-' + item.rank">
          {{ item.rank <= 3 ? medals[item.rank - 1] : '#' + item.rank }}
        </div>
        <span class="rank-avatar">{{ item.avatar }}</span>
        <div class="rank-info">
          <span class="rank-name">{{ item.nickname }}</span>
          <span class="rank-meta">
            {{ item.profession || '' }}{{ item.profession && item.city ? ' · ' : '' }}{{ item.city || '' }}
          </span>
        </div>
        <div class="rank-value font-mono">{{ item.label }}</div>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state glass-card" v-if="!loading && list.length === 0">
      <span class="empty-icon">🏆</span>
      <p>暂无排行数据，快去打卡上榜吧！</p>
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
import { rankingAPI } from '@/services/clockin.api.js'

export default {
  name: 'ClockInRanking',
  setup() {
    const activeTab = ref('workhours')
    const period = ref('week')
    const list = ref([])
    const loading = ref(false)
    const medals = ['🥇', '🥈', '🥉']

    const tabs = [
      { id: 'workhours', icon: '🏆', label: '总工时榜' },
      { id: 'avgworkhours', icon: '📊', label: '日均工时' },
      { id: 'early', icon: '🌅', label: '早起榜' },
      { id: 'late', icon: '🌙', label: '夜猫榜' },
      { id: 'ontime', icon: '🎯', label: '准时榜' },
      { id: 'streak', icon: '🔥', label: '连续打卡' }
    ]

    const fetchRanking = async () => {
      loading.value = true
      try {
        let data
        switch (activeTab.value) {
          case 'workhours':
            data = await rankingAPI.getWorkhours(period.value)
            break
          case 'avgworkhours':
            data = await rankingAPI.getAvgWorkhours(period.value)
            break
          case 'early':
            data = await rankingAPI.getEarly(period.value)
            break
          case 'late':
            data = await rankingAPI.getLate(period.value)
            break
          case 'ontime':
            data = await rankingAPI.getOntime(period.value)
            break
          case 'streak':
            data = await rankingAPI.getStreak()
            break
        }
        list.value = data.list || []
      } catch (err) {
        console.error('获取排行榜失败:', err)
        list.value = []
      } finally {
        loading.value = false
      }
    }

    const switchTab = (tab) => {
      activeTab.value = tab
      fetchRanking()
    }

    const switchPeriod = (p) => {
      period.value = p
      fetchRanking()
    }

    onMounted(fetchRanking)

    return { activeTab, period, list, loading, medals, tabs, switchTab, switchPeriod }
  }
}
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }

.ranking-wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

/* 榜单切换 */
.ranking-tabs {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.tab-btn {
  padding: var(--space-2) var(--space-4);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  color: var(--text-secondary);
  font-size: var(--text-xs);
  cursor: pointer;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.tab-btn.active {
  background: var(--gradient-primary);
  color: white;
  border-color: transparent;
}

.tab-btn:hover:not(.active) {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

/* 周期切换 */
.period-switch {
  display: flex;
  gap: var(--space-2);
}

.period-btn {
  padding: var(--space-1) var(--space-4);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: var(--text-xs);
  cursor: pointer;
  transition: all 0.3s ease;
}

.period-btn.active {
  background: rgba(168, 85, 247, 0.15);
  color: var(--color-primary-light);
  border-color: var(--color-primary);
}

/* 排行列表 */
.ranking-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  transition: all 0.3s ease;
}

.ranking-item:hover {
  transform: translateX(4px);
  border-color: var(--border-glow);
}

.ranking-item.top-3 {
  border-color: rgba(168, 85, 247, 0.2);
}

.rank-badge {
  min-width: 40px;
  text-align: center;
  font-size: var(--text-base);
  font-weight: 700;
  color: var(--text-muted);
}

.rank-badge.rank-1 { font-size: 1.5rem; }
.rank-badge.rank-2 { font-size: 1.3rem; }
.rank-badge.rank-3 { font-size: 1.2rem; }

.rank-avatar {
  font-size: 1.8rem;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
}

.rank-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.rank-name {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--text-primary);
}

.rank-meta {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.rank-value {
  font-size: var(--text-base);
  font-weight: 700;
  color: var(--color-primary-light);
  white-space: nowrap;
}

/* 空状态 & 加载 */
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
  .ranking-tabs { gap: var(--space-1); }
  .tab-btn { padding: var(--space-1) var(--space-3); font-size: 11px; }
  .ranking-item { gap: var(--space-3); padding: var(--space-3) var(--space-4); }
}
</style>
