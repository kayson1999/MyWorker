<template>
  <div class="clockin-page">
    <!-- 页面头部 -->
    <section class="page-hero">
      <div class="container">
        <div class="hero-top">
          <span class="hero-tag font-mono">// 打工人打卡</span>
          <!-- 右上角登录状态区域 -->
          <div class="user-status">
            <template v-if="authStore.isLoggedIn.value">
              <span class="user-avatar">{{ authStore.userAvatar.value }}</span>
              <span class="user-nickname">{{ authStore.userNickname.value }}</span>
              <button class="logout-btn" @click="handleLogout">退出</button>
            </template>
            <button v-else class="login-btn" @click="authStore.openLogin()">🔑 登录</button>
          </div>
        </div>
        <h1 class="hero-title">
          打工人
          <span class="gradient-text">打卡</span>
        </h1>
        <p class="hero-desc">记录工时 少当工贼💪</p>
      </div>
    </section>

    <!-- Tab 切换 -->
    <div class="container">
      <div class="main-tabs">
        <button v-for="tab in mainTabs" :key="tab.id" class="main-tab"
                :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id">
          {{ tab.icon }} {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- 内容区域 -->
    <div class="container content-area">
      <!-- 打卡面板：需要登录 -->
      <div v-show="activeTab === 'clockin'">
        <ClockInPanel v-if="authStore.isLoggedIn.value"
          ref="panelRef"
          @record-updated="handleRecordUpdated" />
        <div v-else class="need-login-card glass-card">
          <span class="need-login-icon">🔒</span>
          <p class="need-login-text">登录后即可使用打卡功能，记录每一天的努力</p>
          <button class="need-login-btn" @click="authStore.openLogin()">🔑 立即登录</button>
        </div>
      </div>

      <!-- 统计：需要登录 -->
      <div v-show="activeTab === 'stats'">
        <ClockInStats v-if="authStore.isLoggedIn.value" ref="statsRef" />
        <div v-else class="need-login-card glass-card">
          <span class="need-login-icon">📊</span>
          <p class="need-login-text">登录后查看你的打卡统计数据</p>
          <button class="need-login-btn" @click="authStore.openLogin()">🔑 立即登录</button>
        </div>
      </div>

      <!-- 排行榜：公开可见 -->
      <ClockInRanking v-show="activeTab === 'ranking'" />

      <!-- 个人资料：需要登录 -->
      <div v-show="activeTab === 'profile'">
        <ClockInProfile v-if="authStore.isLoggedIn.value"
          :user="authStore.currentUser.value"
          @profile-updated="handleProfileUpdated" />
        <div v-else class="need-login-card glass-card">
          <span class="need-login-icon">👤</span>
          <p class="need-login-text">登录后管理你的个人资料</p>
          <button class="need-login-btn" @click="authStore.openLogin()">🔑 立即登录</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { authStore } from '@/services/auth.store.js'
import ClockInProfile from '@/components/clockin/ClockInProfile.vue'
import ClockInPanel from '@/components/clockin/ClockInPanel.vue'
import ClockInStats from '@/components/clockin/ClockInStats.vue'
import ClockInRanking from '@/components/clockin/ClockInRanking.vue'

export default {
  name: 'Home',
  components: { ClockInProfile, ClockInPanel, ClockInStats, ClockInRanking },
  setup() {
    const activeTab = ref('ranking')
    const panelRef = ref(null)
    const statsRef = ref(null)

    const mainTabs = [
      { id: 'clockin', icon: '🕐', label: '打卡' },
      { id: 'stats', icon: '📊', label: '统计' },
      { id: 'ranking', icon: '🏆', label: '排行榜' },
      { id: 'profile', icon: '👤', label: '我的' }
    ]

    const handleProfileUpdated = (user) => {
      authStore.updateUser(user)
    }

    const handleRecordUpdated = () => {}

    const handleLogout = async () => {
      await authStore.logout()
    }

    return {
      authStore, activeTab, mainTabs, panelRef, statsRef,
      handleProfileUpdated, handleRecordUpdated, handleLogout
    }
  }
}
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }
.container { max-width: 900px; margin: 0 auto; padding: 0 var(--space-6); }

/* 页面头部 */
.page-hero {
  padding: var(--space-16) var(--space-6) var(--space-10);
  text-align: center;
}

.hero-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 900px;
  margin: 0 auto var(--space-3);
}

.hero-tag {
  font-size: var(--text-sm);
  color: var(--neon-cyan);
}

/* 右上角用户状态区域 */
.user-status {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.user-avatar {
  font-size: 1.5rem;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--radius-full);
  border: 2px solid rgba(168, 85, 247, 0.3);
}

.user-nickname {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-primary);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.login-btn {
  padding: var(--space-2) var(--space-5);
  background: var(--gradient-primary);
  color: white;
  border: none;
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.login-btn:hover {
  box-shadow: 0 4px 20px rgba(168, 85, 247, 0.4);
  transform: translateY(-2px);
}

.logout-btn {
  padding: var(--space-1) var(--space-4);
  background: rgba(244, 114, 182, 0.1);
  border: 1px solid rgba(244, 114, 182, 0.2);
  border-radius: var(--radius-full);
  color: var(--neon-pink);
  font-size: var(--text-xs);
  cursor: pointer;
  transition: all 0.3s ease;
}

.logout-btn:hover {
  background: rgba(244, 114, 182, 0.2);
  border-color: var(--neon-pink);
}

.hero-title {
  font-size: var(--text-4xl);
  font-weight: 800;
  color: var(--text-primary);
  margin-bottom: var(--space-3);
}

.gradient-text {
  background: var(--gradient-text);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-desc {
  font-size: var(--text-base);
  color: var(--text-secondary);
  margin-bottom: var(--space-5);
}

/* 需要登录提示卡片 */
.need-login-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-16) var(--space-8);
  text-align: center;
  gap: var(--space-4);
}

.need-login-icon {
  font-size: 3rem;
}

.need-login-text {
  font-size: var(--text-base);
  color: var(--text-muted);
  max-width: 300px;
  line-height: 1.6;
}

.need-login-btn {
  padding: var(--space-3) var(--space-8);
  background: var(--gradient-primary);
  color: white;
  border: none;
  border-radius: var(--radius-full);
  font-size: var(--text-base);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.need-login-btn:hover {
  box-shadow: 0 4px 20px rgba(168, 85, 247, 0.4);
  transform: translateY(-2px);
}

/* Tab 切换 */
.main-tabs {
  display: flex;
  gap: var(--space-2);
  padding: var(--space-2);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  margin-bottom: var(--space-8);
}

.main-tab {
  flex: 1;
  padding: var(--space-3) var(--space-4);
  background: transparent;
  border: none;
  border-radius: var(--radius-lg);
  color: var(--text-muted);
  font-size: var(--text-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
}

.main-tab.active {
  background: var(--gradient-primary);
  color: white;
  box-shadow: 0 4px 15px rgba(168, 85, 247, 0.3);
}

.main-tab:hover:not(.active) {
  color: var(--text-primary);
  background: var(--bg-tertiary);
}

/* 内容区域 */
.content-area {
  padding-bottom: var(--space-16);
}

/* 响应式 */
@media (max-width: 768px) {
  .hero-title { font-size: var(--text-3xl); }
  .main-tabs { gap: var(--space-1); }
  .main-tab { padding: var(--space-2) var(--space-2); font-size: var(--text-xs); }
  .container { padding: 0 var(--space-4); }
  .hero-top { padding: 0 var(--space-4); }
  .user-nickname { display: none; }
}
</style>
