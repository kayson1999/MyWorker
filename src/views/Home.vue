<template>
  <div class="clockin-page">
    <!-- 工贼榜顶部流动横幅 -->
    <GongzeiBanner v-if="authStore.isLoggedIn.value" />

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

    <!-- 主 Tab 切换 -->
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

      <!-- ========== 工时打卡（含打卡 + 统计子Tab） ========== -->
      <div v-show="activeTab === 'clockin'">
        <template v-if="authStore.isLoggedIn.value">
          <!-- 子Tab切换 -->
          <div class="sub-tabs">
            <button class="sub-tab" :class="{ active: clockinSubTab === 'punch' }" @click="clockinSubTab = 'punch'">🕐 打卡</button>
            <button class="sub-tab" :class="{ active: clockinSubTab === 'stats' }" @click="clockinSubTab = 'stats'">📊 统计</button>
          </div>
          <ClockInPanel v-show="clockinSubTab === 'punch'"
            ref="panelRef"
            @record-updated="handleRecordUpdated" />
          <ClockInStats v-show="clockinSubTab === 'stats'" ref="statsRef" />
        </template>
        <div v-else class="need-login-card glass-card">
          <span class="need-login-icon">🔒</span>
          <p class="need-login-text">登录后即可使用打卡功能，记录每一天的努力</p>
          <button class="need-login-btn" @click="authStore.openLogin()">🔑 立即登录</button>
        </div>
      </div>

      <!-- ========== 计划签到 ========== -->
      <div v-show="activeTab === 'plan'">
        <template v-if="authStore.isLoggedIn.value">
          <PlanPanel />
        </template>
        <div v-else class="need-login-card glass-card">
          <span class="need-login-icon">🎯</span>
          <p class="need-login-text">登录后即可创建计划，在日历上打卡签到</p>
          <button class="need-login-btn" @click="authStore.openLogin()">🔑 立即登录</button>
        </div>
      </div>

      <!-- ========== 排行榜（公开可见） ========== -->
      <div v-show="activeTab === 'ranking'">
        <div class="sub-tabs">
          <button class="sub-tab" :class="{ active: rankingType === 'clockin' }" @click="rankingType = 'clockin'">🕐 工时排行</button>
          <button class="sub-tab" :class="{ active: rankingType === 'plan' }" @click="rankingType = 'plan'">🎯 计划排行</button>
        </div>
        <ClockInRanking v-show="rankingType === 'clockin'" />
        <PlanRanking v-show="rankingType === 'plan'" />
      </div>

      <!-- ========== 个人中心（需要登录） ========== -->
      <div v-show="activeTab === 'profile'">
        <template v-if="authStore.isLoggedIn.value">
          <!-- 个人资料编辑弹层 -->
          <ClockInProfile v-if="showProfileEdit"
            :user="authStore.currentUser.value"
            @profile-updated="(u) => { handleProfileUpdated(u); showProfileEdit = false; }" />
          <div v-if="showProfileEdit" class="back-to-center">
            <button class="back-btn" @click="showProfileEdit = false">← 返回个人中心</button>
          </div>
          <!-- 个人中心主页 -->
          <UserCenter v-if="!showProfileEdit"
            :user="authStore.currentUser.value"
            @edit-profile="showProfileEdit = true" />
        </template>
        <div v-else class="need-login-card glass-card">
          <span class="need-login-icon">👤</span>
          <p class="need-login-text">登录后查看你的个人中心</p>
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
import UserCenter from '@/components/clockin/UserCenter.vue'
import PlanPanel from '@/components/plan/PlanPanel.vue'
import PlanRanking from '@/components/plan/PlanRanking.vue'
import GongzeiBanner from '@/components/clockin/GongzeiBanner.vue'

export default {
  name: 'Home',
  components: { ClockInProfile, ClockInPanel, ClockInStats, ClockInRanking, UserCenter, PlanPanel, PlanRanking, GongzeiBanner },
  setup() {
    const activeTab = ref('clockin')
    const clockinSubTab = ref('punch')
    const panelRef = ref(null)
    const statsRef = ref(null)

    const showProfileEdit = ref(false)

    const mainTabs = [
      { id: 'clockin', icon: '⏰', label: '工时打卡' },
      { id: 'plan', icon: '🎯', label: '计划签到' },
      { id: 'ranking', icon: '🏆', label: '排行榜' },
      { id: 'profile', icon: '👤', label: '个人中心' }
    ]

    const handleProfileUpdated = (user) => {
      authStore.updateUser(user)
    }

    const handleRecordUpdated = () => {}

    const handleLogout = async () => {
      await authStore.logout()
    }

    const rankingType = ref('clockin')

    return {
      authStore, activeTab, clockinSubTab, mainTabs, panelRef, statsRef, rankingType,
      showProfileEdit, handleProfileUpdated, handleRecordUpdated, handleLogout
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

/* 主 Tab 切换 */
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

/* 子 Tab 切换（工时打卡内的打卡/统计、排行榜内的工时/计划） */
.sub-tabs {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-5);
  padding: var(--space-1);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
}

.sub-tab {
  flex: 1;
  padding: var(--space-2) var(--space-4);
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: var(--text-sm);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
}

.sub-tab.active {
  background: rgba(168, 85, 247, 0.15);
  color: var(--color-primary-light);
}

.sub-tab:hover:not(.active) {
  color: var(--text-primary);
}

/* 内容区域 */
.content-area {
  padding-bottom: var(--space-16);
}

/* 返回个人中心按钮 */
.back-to-center {
  margin-bottom: var(--space-4);
}

.back-btn {
  background: none;
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-4);
  color: var(--text-muted);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.3s ease;
}

.back-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

/* 响应式 */
@media (max-width: 768px) {
  .hero-title { font-size: var(--text-3xl); }
  .main-tabs { gap: var(--space-1); }
  .main-tab { padding: var(--space-2) var(--space-2); font-size: var(--text-xs); }
  .container { padding: 0 var(--space-4); }
  .hero-top { padding: 0 var(--space-4); }
  .user-nickname { display: none; }
  .page-hero { padding: var(--space-10) var(--space-4) var(--space-6); }
  .content-area { padding-bottom: calc(var(--space-16) + env(safe-area-inset-bottom, 0px)); }

  /* 主 Tab 栏移动端优化：固定底部导航样式 */
  .main-tabs {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 999;
    border-radius: var(--radius-xl) var(--radius-xl) 0 0;
    margin-bottom: 0;
    padding: var(--space-2);
    padding-bottom: calc(var(--space-2) + env(safe-area-inset-bottom, 0px));
    box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.4);
  }
  .main-tab {
    padding: var(--space-2) var(--space-1);
    font-size: 11px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>
