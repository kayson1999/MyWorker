<template>
  <div class="ranking-wrapper">
    <!-- 榜单切换 -->
    <div class="ranking-tabs">
      <button v-for="tab in tabs" :key="tab.id" class="tab-btn"
              :class="{ active: activeTab === tab.id }" @click="switchTab(tab.id)">
        {{ tab.icon }} {{ tab.label }}
      </button>
    </div>

    <!-- 周期切换（仅总打卡天数榜支持） -->
    <div class="period-switch" v-if="activeTab === 'total'">
      <button class="period-btn" :class="{ active: period === 'week' }" @click="switchPeriod('week')">本周</button>
      <button class="period-btn" :class="{ active: period === 'month' }" @click="switchPeriod('month')">本月</button>
      <button class="period-btn" :class="{ active: period === 'all' }" @click="switchPeriod('all')">全部</button>
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
          <!-- 展示用户的公开计划 -->
          <div class="rank-plans" v-if="item.plans && item.plans.length > 0">
            <span class="rank-plan-tag" v-for="(plan, idx) in item.plans" :key="idx" :title="plan.content || plan.title">
              {{ plan.icon }} {{ plan.title }}
            </span>
          </div>
        </div>
        <div class="rank-value font-mono">{{ item.label }}</div>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state glass-card" v-if="!loading && list.length === 0">
      <span class="empty-icon">🎯</span>
      <p>暂无排行数据，快去创建计划打卡上榜吧！</p>
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
import { planRankingAPI } from '@/services/plan.api.js'

export default {
  name: 'PlanRanking',
  setup() {
    const activeTab = ref('total')
    const period = ref('all')
    const list = ref([])
    const loading = ref(false)
    const medals = ['🥇', '🥈', '🥉']

    const tabs = [
      { id: 'total', icon: '📅', label: '总打卡天数' },
      { id: 'streak', icon: '🔥', label: '连续打卡' },
      { id: 'plans', icon: '📋', label: '活跃计划数' },
      { id: 'completion', icon: '🎯', label: '完成率' }
    ]

    const fetchRanking = async () => {
      loading.value = true
      try {
        let data
        switch (activeTab.value) {
          case 'total':
            data = await planRankingAPI.getTotal(period.value)
            break
          case 'streak':
            data = await planRankingAPI.getStreak()
            break
          case 'plans':
            data = await planRankingAPI.getPlans()
            break
          case 'completion':
            data = await planRankingAPI.getCompletion()
            break
        }
        list.value = data.list || []
      } catch (err) {
        console.error('获取计划排行榜失败:', err)
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

.rank-plans {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
}

.rank-plan-tag {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 1px var(--space-2);
  background: rgba(168, 85, 247, 0.08);
  border: 1px solid rgba(168, 85, 247, 0.12);
  border-radius: var(--radius-full);
  font-size: 11px;
  color: var(--text-secondary);
  white-space: nowrap;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rank-value {
  font-size: var(--text-base);
  font-weight: 700;
  color: var(--color-primary-light);
  white-space: nowrap;
}

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
