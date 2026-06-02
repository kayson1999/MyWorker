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
             @click="showAchievementDetail(a)">
          <span class="achievement-icon">{{ a.unlocked ? a.icon : '🔒' }}</span>
          <span class="achievement-name">{{ a.name }}</span>
        </div>
      </div>
      <div v-else class="empty-hint">完成打卡和计划来解锁成就吧！</div>
      <!-- 分页控件 -->
      <div class="achievement-pagination" v-if="achievementTotal > achievementPageSize">
        <button class="page-btn" :disabled="achievementPage <= 1" @click="changeAchievementPage(achievementPage - 1)">←</button>
        <span class="page-info font-mono">{{ achievementPage }} / {{ achievementTotalPages }}</span>
        <button class="page-btn" :disabled="achievementPage >= achievementTotalPages" @click="changeAchievementPage(achievementPage + 1)">→</button>
      </div>
    </div>

    <!-- 成就详情弹窗 -->
    <div class="achievement-modal-overlay" v-if="selectedAchievement" @click.self="selectedAchievement = null">
      <div class="achievement-modal glass-card">
        <button class="modal-close" @click="selectedAchievement = null">✕</button>
        <div class="modal-icon">{{ selectedAchievement.unlocked ? selectedAchievement.icon : '🔒' }}</div>
        <h3 class="modal-name">{{ selectedAchievement.name }}</h3>
        <p class="modal-desc">{{ selectedAchievement.description }}</p>
        <div class="modal-divider"></div>
        <div class="modal-condition-label">🎯 达成条件</div>
        <p class="modal-condition">{{ selectedAchievement.condition || selectedAchievement.description }}</p>
        <div class="modal-status" :class="selectedAchievement.unlocked ? 'status-unlocked' : 'status-locked'">
          {{ selectedAchievement.unlocked ? '✅ 已解锁' + (selectedAchievement.unlocked_at ? ' · ' + selectedAchievement.unlocked_at.slice(0, 10) : '') : '🔒 未解锁' }}
        </div>
      </div>
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
    const achievementPage = ref(1)
    const achievementPageSize = ref(12)
    const achievementTotal = ref(0)
    const achievementTotalPages = ref(1)
    const selectedAchievement = ref(null)

    // 获取总览数据
    const fetchOverview = async () => {
      try {
        const data = await userCenterAPI.getOverview()
        overview.value = data
      } catch (err) {
        console.error('获取个人中心数据失败:', err)
      }
    }

    // 获取成就列表（分页）
    const fetchAchievements = async (page = 1) => {
      try {
        const data = await userCenterAPI.getAchievements(page, achievementPageSize.value)
        achievements.value = data.achievements || []
        achievementTotal.value = data.total || 0
        achievementPage.value = data.page || 1
        achievementTotalPages.value = Math.ceil(achievementTotal.value / achievementPageSize.value) || 1
      } catch (err) {
        console.error('获取成就失败:', err)
      }
    }

    // 切换成就分页
    const changeAchievementPage = (page) => {
      if (page < 1 || page > achievementTotalPages.value) return
      fetchAchievements(page)
    }

    // 展示成就详情弹窗
    const showAchievementDetail = (achievement) => {
      selectedAchievement.value = achievement
    }

    onMounted(() => {
      fetchOverview()
      fetchAchievements()
    })

    return {
      overview, achievements,
      achievementPage, achievementPageSize, achievementTotal, achievementTotalPages,
      selectedAchievement,
      changeAchievementPage, showAchievementDetail
    }
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
  min-width: 44px;
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
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

.achievement-item {
  cursor: pointer;
}

/* ==================== 成就分页 ==================== */
.achievement-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  margin-top: var(--space-4);
  padding-top: var(--space-3);
  border-top: 1px solid var(--border-color);
}

.page-btn {
  padding: var(--space-1) var(--space-3);
  min-width: 44px;
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.3s ease;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

.page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.page-info {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

/* ==================== 成就详情弹窗 ==================== */
.achievement-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: var(--space-4);
}

.achievement-modal {
  position: relative;
  width: 100%;
  max-width: 360px;
  padding: var(--space-6);
  text-align: center;
  animation: modal-in 0.3s ease;
}

@keyframes modal-in {
  from { opacity: 0; transform: scale(0.9) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-close {
  position: absolute;
  top: var(--space-3);
  right: var(--space-3);
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: var(--text-lg);
  cursor: pointer;
  padding: var(--space-1);
  line-height: 1;
  transition: color 0.2s;
}

.modal-close:hover {
  color: var(--text-primary);
}

.modal-icon {
  font-size: 3rem;
  margin-bottom: var(--space-3);
}

.modal-name {
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: var(--space-2);
}

.modal-desc {
  font-size: var(--text-sm);
  color: var(--text-secondary);
  margin-bottom: var(--space-4);
}

.modal-divider {
  height: 1px;
  background: var(--border-color);
  margin: var(--space-3) 0;
}

.modal-condition-label {
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-primary-light);
  margin-bottom: var(--space-2);
}

.modal-condition {
  font-size: var(--text-sm);
  color: var(--text-secondary);
  line-height: 1.6;
  margin-bottom: var(--space-4);
}

.modal-status {
  display: inline-block;
  padding: var(--space-1) var(--space-3);
  border-radius: var(--radius-full);
  font-size: var(--text-xs);
  font-weight: 600;
}

.modal-status.status-unlocked {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.2);
}

.modal-status.status-locked {
  background: var(--bg-tertiary);
  color: var(--text-muted);
  border: 1px solid var(--border-color);
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
  .level-card { padding: var(--space-4); }
  .section { padding: var(--space-4); }
  .stat-card { padding: var(--space-3); }
  .stat-icon { font-size: 1.2rem; }
  .stat-value { font-size: var(--text-base); }
  .achievement-modal { max-width: 90vw; padding: var(--space-5); }
  .modal-icon { font-size: 2.5rem; }
  .edit-profile-btn { font-size: var(--text-xs); }
  .level-avatar { width: 50px; height: 50px; font-size: 2rem; }
}
</style>
