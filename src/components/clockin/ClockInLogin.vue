<template>
  <div class="login-panel">
    <div class="login-card glass-card">
      <div class="login-header">
        <span class="login-icon">🕐</span>
        <h2 class="login-title">打工人打卡</h2>
        <p class="login-subtitle">{{ isRegister ? '注册新账号' : '登录你的账号' }}</p>
      </div>

      <form @submit.prevent="handleSubmit" class="login-form">
        <div class="form-group">
          <label class="form-label">用户名</label>
          <input v-model="form.username" type="text" class="form-input"
                 placeholder="请输入用户名（3-20位）" autocomplete="username" />
        </div>

        <div class="form-group" v-if="isRegister">
          <label class="form-label">昵称</label>
          <input v-model="form.nickname" type="text" class="form-input"
                 placeholder="给自己取个昵称吧" />
        </div>

        <div class="form-group">
          <label class="form-label">密码</label>
          <input v-model="form.password" type="password" class="form-input"
                 placeholder="请输入密码（至少6位）" autocomplete="current-password" />
        </div>

        <p class="error-msg" v-if="error">{{ error }}</p>

        <button type="submit" class="submit-btn" :disabled="loading">
          <span v-if="loading" class="loading-dot">⏳</span>
          {{ loading ? '请稍候...' : (isRegister ? '注册' : '登录') }}
        </button>
      </form>

      <div class="login-footer">
        <button class="switch-btn" @click="toggleMode">
          {{ isRegister ? '已有账号？去登录' : '没有账号？去注册' }}
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
  emits: ['login-success'],
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
  min-height: 60vh;
  padding: var(--space-8);
}

.login-card {
  width: 100%;
  max-width: 420px;
  padding: var(--space-10);
}

.login-header {
  text-align: center;
  margin-bottom: var(--space-8);
}

.login-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: var(--space-3);
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
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--text-secondary);
}

.form-input {
  padding: var(--space-3) var(--space-4);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  color: var(--text-primary);
  font-size: var(--text-base);
  outline: none;
  transition: all 0.3s ease;
}

.form-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(168, 85, 247, 0.15);
}

.form-input::placeholder {
  color: var(--text-muted);
}

.error-msg {
  color: var(--neon-pink);
  font-size: var(--text-sm);
  text-align: center;
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
}

.submit-btn:hover:not(:disabled) {
  box-shadow: 0 4px 20px rgba(168, 85, 247, 0.4);
  transform: translateY(-2px);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-footer {
  text-align: center;
  margin-top: var(--space-6);
}

.switch-btn {
  background: none;
  border: none;
  color: var(--color-primary-light);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: color 0.3s ease;
}

.switch-btn:hover {
  color: var(--text-primary);
}

.loading-dot {
  margin-right: var(--space-2);
}
</style>
