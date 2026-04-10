<template>
  <div class="login-panel">
    <div class="login-card glass-card">
      <!-- 关闭按钮 -->
      <button class="close-btn" @click="$emit('close')" title="关闭">✕</button>

      <!-- 顶部装饰 -->
      <div class="login-decor">
        <div class="decor-ring"></div>
        <span class="login-logo">⏰</span>
      </div>

      <div class="login-header">
        <h2 class="login-title">{{ isRegister ? '创建账号' : '欢迎回来' }}</h2>
        <p class="login-subtitle">{{ isRegister ? '加入打工人打卡，记录每一天' : '登录后继续记录吧！' }}</p>
      </div>

      <form @submit.prevent="handleSubmit" class="login-form">
        <div class="form-group">
          <div class="input-wrapper">
            <span class="input-icon">👤</span>
            <input v-model="form.username" type="text" class="form-input"
                   placeholder="用户名（3-20位）" autocomplete="username" />
          </div>
        </div>

        <div class="form-group" v-if="isRegister">
          <div class="input-wrapper">
            <span class="input-icon">✨</span>
            <input v-model="form.nickname" type="text" class="form-input"
                   placeholder="给自己取个昵称" />
          </div>
        </div>

        <div class="form-group">
          <div class="input-wrapper">
            <span class="input-icon">🔒</span>
            <input v-model="form.password" type="password" class="form-input"
                   placeholder="密码（至少6位）" autocomplete="current-password" />
          </div>
        </div>

        <p class="error-msg" v-if="error">⚠️ {{ error }}</p>

        <button type="submit" class="submit-btn" :disabled="loading">
          <span class="btn-content">
            <span v-if="loading" class="loading-spinner"></span>
            <span v-else class="btn-icon">{{ isRegister ? '🚀' : '👋' }}</span>
            {{ loading ? '请稍候...' : (isRegister ? '注册并开始' : '登录') }}
          </span>
        </button>
      </form>

      <div class="login-divider">
        <span class="divider-text">or</span>
      </div>

      <div class="login-footer">
        <button class="switch-btn" @click="toggleMode">
          {{ isRegister ? '已有账号？直接登录 →' : '没有账号？立即注册 →' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { authAPI } from '@/services/clockin.api.js'

export default {
  name: 'ClockInLogin',
  emits: ['login-success', 'close'],
  setup(props, { emit }) {
    const isRegister = ref(false)
    const loading = ref(false)
    const error = ref('')
    const form = reactive({
      username: '',
      password: '',
      nickname: ''
    })

    const toggleMode = () => {
      isRegister.value = !isRegister.value
      error.value = ''
    }

    const handleSubmit = async () => {
      error.value = ''
      loading.value = true

      try {
        let result
        if (isRegister.value) {
          if (!form.nickname) {
            error.value = '请输入昵称'
            loading.value = false
            return
          }
          result = await authAPI.register({
            username: form.username,
            password: form.password,
            nickname: form.nickname
          })
        } else {
          result = await authAPI.login({
            username: form.username,
            password: form.password
          })
        }
        emit('login-success', result.user)
      } catch (err) {
        error.value = err.message || '操作失败，请重试'
      } finally {
        loading.value = false
      }
    }

    return { isRegister, loading, error, form, toggleMode, handleSubmit }
  }
}
</script>

<style scoped>
.login-panel {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--space-8);
}

.login-card {
  position: relative;
  width: 100%;
  max-width: 400px;
  padding: var(--space-10) var(--space-8);
  border: 1px solid rgba(168, 85, 247, 0.15);
  box-shadow:
    0 0 40px rgba(168, 85, 247, 0.08),
    0 20px 60px rgba(0, 0, 0, 0.5);
  animation: card-appear 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes card-appear {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 关闭按钮 */
.close-btn {
  position: absolute;
  top: var(--space-4);
  right: var(--space-4);
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-full);
  color: var(--text-muted);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(244, 114, 182, 0.15);
  border-color: rgba(244, 114, 182, 0.3);
  color: var(--neon-pink);
}

/* 顶部装饰 */
.login-decor {
  display: flex;
  justify-content: center;
  margin-bottom: var(--space-5);
  position: relative;
}

.decor-ring {
  position: absolute;
  width: 72px;
  height: 72px;
  border-radius: var(--radius-full);
  border: 2px solid rgba(168, 85, 247, 0.2);
  animation: ring-pulse 3s ease-in-out infinite;
}

@keyframes ring-pulse {
  0%, 100% {
    transform: scale(1);
    border-color: rgba(168, 85, 247, 0.2);
  }
  50% {
    transform: scale(1.1);
    border-color: rgba(129, 140, 248, 0.3);
  }
}

.login-logo {
  font-size: 2.5rem;
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(168, 85, 247, 0.15), rgba(99, 102, 241, 0.1));
  border-radius: var(--radius-full);
  border: 1px solid rgba(168, 85, 247, 0.2);
  position: relative;
  z-index: 1;
}

.login-header {
  text-align: center;
  margin-bottom: var(--space-8);
}

.login-title {
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: var(--space-2);
}

.login-subtitle {
  font-size: var(--text-sm);
  color: var(--text-muted);
  line-height: 1.5;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
}

/* 输入框带图标 */
.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: var(--space-4);
  font-size: var(--text-base);
  pointer-events: none;
  z-index: 1;
}

.form-input {
  width: 100%;
  padding: var(--space-3) var(--space-4) var(--space-3) calc(var(--space-4) + 1.8rem);
  background: rgba(22, 20, 42, 0.6);
  border: 1px solid rgba(168, 85, 247, 0.1);
  border-radius: var(--radius-lg);
  color: var(--text-primary);
  font-size: var(--text-base);
  outline: none;
  transition: all 0.3s ease;
}

.form-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(168, 85, 247, 0.1), 0 0 20px rgba(168, 85, 247, 0.05);
  background: rgba(22, 20, 42, 0.8);
}

.form-input::placeholder {
  color: var(--text-muted);
  font-size: var(--text-sm);
}

.error-msg {
  color: var(--neon-pink);
  font-size: var(--text-sm);
  text-align: center;
  padding: var(--space-2) var(--space-3);
  background: rgba(244, 114, 182, 0.08);
  border-radius: var(--radius-md);
  border: 1px solid rgba(244, 114, 182, 0.15);
}

.submit-btn {
  padding: var(--space-3) var(--space-6);
  background: var(--gradient-primary);
  color: white;
  border: none;
  border-radius: var(--radius-lg);
  font-size: var(--text-base);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-top: var(--space-2);
  position: relative;
  overflow: hidden;
}

.submit-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(255,255,255,0.1) 0%, transparent 50%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.submit-btn:hover:not(:disabled)::before {
  opacity: 1;
}

.submit-btn:hover:not(:disabled) {
  box-shadow: 0 4px 25px rgba(168, 85, 247, 0.4), 0 0 40px rgba(99, 102, 241, 0.15);
  transform: translateY(-2px);
}

.submit-btn:active:not(:disabled) {
  transform: translateY(0);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
}

.btn-icon {
  font-size: 1.1rem;
}

/* 加载旋转动画 */
.loading-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 分割线 */
.login-divider {
  display: flex;
  align-items: center;
  margin: var(--space-6) 0;
}

.login-divider::before,
.login-divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(168, 85, 247, 0.15), transparent);
}

.divider-text {
  padding: 0 var(--space-4);
  font-size: var(--text-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.login-footer {
  text-align: center;
}

.switch-btn {
  background: none;
  border: none;
  color: var(--color-primary-light);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.3s ease;
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
}

.switch-btn:hover {
  color: var(--text-primary);
  background: rgba(168, 85, 247, 0.08);
}
</style>
