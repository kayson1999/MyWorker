<template>
  <div id="worker-app">
    <router-view />

    <!-- 全局登录弹窗 -->
    <teleport to="body">
      <div class="login-modal-overlay" v-if="authStore.showLoginModal.value" @click.self="authStore.closeLogin()">
        <ClockInLogin @login-success="handleLoginSuccess" @close="authStore.closeLogin()" />
      </div>
    </teleport>
  </div>
</template>

<script>
import { onMounted } from 'vue'
import { authStore } from '@/services/auth.store.js'
import ClockInLogin from '@/components/clockin/ClockInLogin.vue'

export default {
  name: 'App',
  components: { ClockInLogin },
  setup() {
    onMounted(() => {
      authStore.init()
    })

    const handleLoginSuccess = (user) => {
      authStore.updateUser(user)
      authStore.closeLogin()
    }

    return { authStore, handleLoginSuccess }
  }
}
</script>

<style>
/* 登录弹窗遮罩 */
.login-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 10000;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
