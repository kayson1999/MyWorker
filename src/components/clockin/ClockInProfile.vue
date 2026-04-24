<template>
  <div class="profile-panel">
    <div class="profile-card glass-card">
      <div class="profile-header">
        <span class="profile-avatar">{{ user.avatar || '😎' }}</span>
        <div class="profile-info">
          <h3 class="profile-name">{{ user.nickname }}</h3>
          <p class="profile-meta font-mono">@{{ user.username }}</p>
        </div>
        <button class="edit-btn" @click="editing = !editing">
          {{ editing ? '取消' : '✏️ 编辑' }}
        </button>
      </div>

      <!-- 展示模式 -->
      <div class="profile-details" v-if="!editing">
        <!-- 称号展示 -->
        <div class="title-section">
          <h4 class="title-heading">🏷️ 工作风格标签</h4>
          <div class="title-grid">
            <div class="title-item">
              <span class="title-icon">📅</span>
              <span class="title-label">本周风格</span>
              <span class="title-value">{{ titles.week_title || '暂无称号' }}</span>
            </div>
            <div class="title-item">
              <span class="title-icon">📊</span>
              <span class="title-label">本月风格</span>
              <span class="title-value">{{ titles.month_title || '暂无称号' }}</span>
            </div>
            <div class="title-item">
              <span class="title-icon">🎯</span>
              <span class="title-label">本年风格</span>
              <span class="title-value">{{ titles.year_title || '暂无称号' }}</span>
            </div>
          </div>
        </div>

        <!-- 原有信息 -->
        <div class="detail-item">
          <span class="detail-icon">💼</span>
          <span class="detail-label">职业</span>
          <span class="detail-value">{{ user.profession || '未设置' }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-icon">🏢</span>
          <span class="detail-label">岗位</span>
          <span class="detail-value">{{ user.position || '未设置' }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-icon">📍</span>
          <span class="detail-label">城市</span>
          <span class="detail-value">{{ user.city || '未设置' }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-icon">⏰</span>
          <span class="detail-label">标准工时</span>
          <span class="detail-value">{{ user.standard_start || '09:00' }} - {{ user.standard_end || '18:00' }}</span>
        </div>
      </div>

      <!-- 编辑模式 -->
      <form v-else class="profile-form" @submit.prevent="saveProfile">
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">昵称</label>
            <input v-model="form.nickname" class="form-input" placeholder="昵称" />
          </div>
          <div class="form-group">
            <label class="form-label">头像</label>
            <div class="avatar-picker">
              <span v-for="a in avatars" :key="a" class="avatar-option"
                    :class="{ active: form.avatar === a }" @click="form.avatar = a">{{ a }}</span>
            </div>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">职业</label>
            <input v-model="form.profession" class="form-input" placeholder="如：程序员" />
          </div>
          <div class="form-group">
            <label class="form-label">岗位</label>
            <input v-model="form.position" class="form-input" placeholder="如：前端开发" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">城市</label>
            <input v-model="form.city" class="form-input" placeholder="如：深圳" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">标准上班时间</label>
            <input v-model="form.standard_start" type="time" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">标准下班时间</label>
            <input v-model="form.standard_end" type="time" class="form-input" />
          </div>
        </div>
        <p class="error-msg" v-if="error">{{ error }}</p>
        <button type="submit" class="save-btn" :disabled="saving">
          {{ saving ? '保存中...' : '💾 保存' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script>
import { ref, reactive, watch, onMounted } from 'vue'
import { userAPI, titleAPI } from '@/services/clockin.api.js'

export default {
  name: 'ClockInProfile',
  props: {
    user: { type: Object, required: true }
  },
  emits: ['profile-updated'],
  setup(props, { emit }) {
    const editing = ref(false)
    const saving = ref(false)
    const error = ref('')
    const avatars = ['😎', '🚀', '🔥', '💻', '🎮', '⚡', '🌟', '🦊', '🐱', '🐶', '🦁', '🐼', '🐨', '🐯', '🦄', '🐸']
    const titles = ref({})

    const form = reactive({
      nickname: '',
      avatar: '',
      profession: '',
      position: '',
      city: '',
      standard_start: '09:00',
      standard_end: '18:00'
    })

    // 监听 user 变化，同步到表单
    watch(() => props.user, (u) => {
      if (u) {
        form.nickname = u.nickname || ''
        form.avatar = u.avatar || '😎'
        form.profession = u.profession || ''
        form.position = u.position || ''
        form.city = u.city || ''
        form.standard_start = u.standard_start || '09:00'
        form.standard_end = u.standard_end || '18:00'
      }
    }, { immediate: true })

    // 获取用户称号
    const fetchTitles = async () => {
      try {
        const data = await titleAPI.getTitles()
        titles.value = data
      } catch (err) {
        console.error('获取称号失败:', err)
        titles.value = {}
      }
    }

    // 组件挂载时获取称号
    onMounted(() => {
      fetchTitles()
    })

    const saveProfile = async () => {
      error.value = ''
      saving.value = true
      try {
        const result = await userAPI.updateProfile({ ...form })
        emit('profile-updated', result.user)
        editing.value = false
      } catch (err) {
        error.value = err.message || '保存失败'
      } finally {
        saving.value = false
      }
    }

    return { editing, saving, error, avatars, titles, form, saveProfile }
  }
}
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }

.profile-card {
  padding: var(--space-6);
  margin-bottom: var(--space-6);
}

.profile-header {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.profile-avatar {
  font-size: 2.5rem;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--radius-xl);
  border: 2px solid var(--border-color);
}

.profile-info {
  flex: 1;
}

.profile-name {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--text-primary);
}

.profile-meta {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.edit-btn {
  padding: var(--space-2) var(--space-4);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.3s ease;
}

.edit-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

.profile-details {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--space-4);
}

/* 称号样式 */
.title-section {
  grid-column: 1 / -1;
  margin-bottom: var(--space-4);
  padding: var(--space-4);
  background: var(--bg-tertiary);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
}

.title-heading {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--space-3);
}

.title-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-3);
}

.title-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
}

.title-icon {
  font-size: 1.1rem;
}

.title-label {
  font-size: var(--text-xs);
  color: var(--text-muted);
  min-width: 60px;
}

.title-value {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary-light);
  flex: 1;
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
}

.detail-icon {
  font-size: 1.2rem;
}

.detail-label {
  font-size: var(--text-xs);
  color: var(--text-muted);
  min-width: 40px;
}

.detail-value {
  font-size: var(--text-sm);
  color: var(--text-primary);
  font-weight: 500;
}

/* 编辑表单 */
.profile-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-size: var(--text-xs);
  font-weight: 500;
  color: var(--text-secondary);
}

.form-input {
  padding: var(--space-2) var(--space-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: var(--text-sm);
  outline: none;
  transition: all 0.3s ease;
}

.form-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(168, 85, 247, 0.15);
}

.avatar-picker {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.avatar-option {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s ease;
  font-size: 1.1rem;
}

.avatar-option:hover {
  background: var(--bg-tertiary);
}

.avatar-option.active {
  border-color: var(--color-primary);
  background: rgba(168, 85, 247, 0.15);
}

.error-msg {
  color: var(--neon-pink);
  font-size: var(--text-sm);
  text-align: center;
}

.save-btn {
  padding: var(--space-3) var(--space-6);
  background: var(--gradient-primary);
  color: white;
  border: none;
  border-radius: var(--radius-lg);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  align-self: flex-start;
}

.save-btn:hover:not(:disabled) {
  box-shadow: 0 4px 20px rgba(168, 85, 247, 0.4);
  transform: translateY(-2px);
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .profile-header { flex-wrap: wrap; }
  .form-row { grid-template-columns: 1fr; }
  .profile-details { grid-template-columns: 1fr; }
  .title-grid { grid-template-columns: 1fr; }
  .title-value {
    text-align: left;
    white-space: normal;
    word-break: break-word;
  }
}
</style>
