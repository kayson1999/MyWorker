<template>
  <div class="panel-wrapper">
    <!-- 实时时钟 -->
    <div class="clock-display glass-card">
      <div class="clock-time font-mono">{{ currentTime }}</div>
      <div class="clock-date">{{ currentDate }}</div>
    </div>

    <!-- 打卡按钮区 -->
    <div class="clock-actions">
      <div class="action-card glass-card" :class="{ done: todayRecord?.clock_in }">
        <div class="action-icon">☀️</div>
        <div class="action-label">上班打卡</div>
        <div class="record-date" v-if="isOvernightPending">正在完成 {{ todayRecord.date }} 的打卡</div>
        <div class="action-time font-mono" v-if="todayRecord?.clock_in && adjusting !== 'clock_in'">
          {{ todayRecord.clock_in }}
        </div>
        <!-- 调整上班时间输入框 -->
        <div class="adjust-form" v-if="adjusting === 'clock_in'">
          <input v-model="adjustTime" type="time" class="adjust-input" />
          <div class="adjust-actions">
            <button class="adjust-confirm" @click="confirmAdjust" :disabled="loading">✓</button>
            <button class="adjust-cancel" @click="cancelAdjust">✕</button>
          </div>
        </div>
        <div class="action-status" v-if="todayRecord?.clock_in && adjusting !== 'clock_in'">
          ✅ 已打卡
          <button class="adjust-btn" @click="startAdjust('clock_in')" title="调整打卡时间">🕐 调整</button>
        </div>
        <button class="action-btn" v-if="!todayRecord?.clock_in" @click="handleClockIn" :disabled="loading">
          {{ loading ? '打卡中...' : '打卡' }}
        </button>
      </div>

      <div class="action-card glass-card" :class="{ done: todayRecord?.clock_out }">
        <div class="action-icon">🌙</div>
        <div class="action-label">下班打卡</div>
        <div class="action-time font-mono" v-if="todayRecord?.clock_out && adjusting !== 'clock_out'">
          {{ todayRecord.clock_out }}
        </div>
        <!-- 调整下班时间输入框 -->
        <div class="adjust-form" v-if="adjusting === 'clock_out'">
          <input v-model="adjustTime" type="time" class="adjust-input" />
          <div class="adjust-actions">
            <button class="adjust-confirm" @click="confirmAdjust" :disabled="loading">✓</button>
            <button class="adjust-cancel" @click="cancelAdjust">✕</button>
          </div>
        </div>
        <div class="action-status" v-if="todayRecord?.clock_out && adjusting !== 'clock_out'">
          ✅ 已打卡
          <button class="adjust-btn" @click="startAdjust('clock_out')" title="调整打卡时间">🕐 调整</button>
        </div>
        <button class="action-btn" v-else-if="!todayRecord?.clock_out && todayRecord?.clock_in"
                @click="handleClockOut" :disabled="loading">
          {{ loading ? '打卡中...' : '打卡' }}
        </button>
        <div class="action-status waiting" v-if="!todayRecord?.clock_out && !todayRecord?.clock_in">⏳ 请先上班打卡</div>
      </div>
    </div>

    <!-- 今日工时 -->
    <div class="today-summary glass-card" v-if="todayRecord?.clock_in">
      <div class="summary-item">
        <span class="summary-icon">📊</span>
        <span class="summary-label">{{ isOvernightPending ? '本次工时' : '今日工时' }}</span>
        <span class="summary-value font-mono">
          {{ todayRecord.clock_out ? todayRecord.duration.toFixed(1) + 'h' : liveHours }}
        </span>
      </div>
      <div class="summary-item" v-if="todayRecord.is_manual">
        <span class="summary-icon">📝</span>
        <span class="summary-label">补卡记录</span>
      </div>
      <div class="summary-item">
        <span class="summary-icon">🏷️</span>
        <span class="summary-label">今日称号</span>
        <span class="summary-value title-text">
          {{ todayTitle?.title || '—' }}
          <span v-if="todayTitle?.sub_title" class="title-sub">{{ todayTitle.sub_title }}</span>
        </span>
      </div>
    </div>

    <!-- 手动补卡 -->
    <div class="manual-section">
      <button class="manual-toggle" @click="showManual = !showManual">
        {{ showManual ? '收起' : '📝 补卡 / 修改历史记录' }}
      </button>
      <div class="manual-form glass-card" v-if="showManual">
        <p class="manual-hint">选择任意过去日期即可新增或覆盖当天打卡时间；下班时间早于上班时间时，会按凌晨跨日下班计算工时。</p>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">日期</label>
            <input v-model="manualForm.date" type="date" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">上班时间</label>
            <input v-model="manualForm.clock_in" type="time" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">下班时间</label>
            <input v-model="manualForm.clock_out" type="time" class="form-input" />
          </div>
        </div>
        <p class="error-msg" v-if="manualError">{{ manualError }}</p>
        <button class="save-btn" @click="handleManual" :disabled="loading">
          {{ loading ? '提交中...' : '保存记录' }}
        </button>
      </div>
    </div>

    <!-- 提示消息 -->
    <div class="toast" v-if="toast" :class="toast.type">{{ toast.message }}</div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { clockinAPI, titleAPI } from '@/services/clockin.api.js'

export default {
  name: 'ClockInPanel',
  emits: ['record-updated'],
  setup(props, { emit }) {
    const todayRecord = ref(null)
    const todayTitle = ref(null)
    const loading = ref(false)
    const showManual = ref(false)
    const manualError = ref('')
    const toast = ref(null)
    const currentTime = ref('')
    const currentDate = ref('')
    let timer = null

    const adjusting = ref(null)   // 当前正在调整的类型: 'clock_in' | 'clock_out' | null
    const adjustTime = ref('')     // 调整的目标时间

    const manualForm = reactive({
      date: '',
      clock_in: '09:00',
      clock_out: '18:00'
    })

    const getTodayISO = () => {
      const now = new Date()
      const y = now.getFullYear()
      const m = String(now.getMonth() + 1).padStart(2, '0')
      const d = String(now.getDate()).padStart(2, '0')
      return `${y}-${m}-${d}`
    }

    const isOvernightPending = computed(() => {
      return !!todayRecord.value?.date && todayRecord.value.date !== getTodayISO() && !todayRecord.value.clock_out
    })

    // 实时时钟
    const updateClock = () => {
      const now = new Date()
      currentTime.value = now.toTimeString().slice(0, 8)
      const weekDays = ['日', '一', '二', '三', '四', '五', '六']
      currentDate.value = `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日 星期${weekDays[now.getDay()]}`
    }

    // 实时工时计算
    const liveHours = computed(() => {
      if (!todayRecord.value?.clock_in || todayRecord.value?.clock_out) return '0.0h'
      const [h, m] = todayRecord.value.clock_in.split(':').map(Number)
      const now = new Date()
      let diff = (now.getHours() * 60 + now.getMinutes()) - (h * 60 + m)
      if (isOvernightPending.value && diff < 0) diff += 24 * 60
      return (Math.max(0, diff) / 60).toFixed(1) + 'h (进行中...)'
    })

    const showToast = (message, type = 'success') => {
      toast.value = { message, type }
      setTimeout(() => { toast.value = null }, 3000)
    }

    // 获取今日记录
    const fetchToday = async () => {
      try {
        const data = await clockinAPI.getToday()
        todayRecord.value = data.record
        fetchTodayTitle() // 同时获取今日称号
      } catch (err) {
        console.error('获取今日记录失败:', err)
      }
    }

    // 获取今日称号
    const fetchTodayTitle = async () => {
      try {
        const data = await titleAPI.getTodayTitle()
        todayTitle.value = data
      } catch (err) {
        console.error('获取今日称号失败:', err)
        todayTitle.value = null
      }
    }

    // 上班打卡
    const handleClockIn = async () => {
      loading.value = true
      try {
        const data = await clockinAPI.clockIn()
        todayRecord.value = data.record
        let msg = '☀️ 上班打卡成功！'
        if (data.exp_gained > 0) msg += ` +${data.exp_gained} EXP`
        if (data.leveled_up) msg += ` 🎉 升级到 Lv.${data.new_level}！`
        showToast(msg, 'success')
        emit('record-updated')
        fetchTodayTitle()
      } catch (err) {
        showToast(err.message || '打卡失败', 'error')
      } finally {
        loading.value = false
      }
    }

    // 下班打卡
    const handleClockOut = async () => {
      loading.value = true
      try {
        const data = await clockinAPI.clockOut()
        if (data.record?.date && data.record.date !== getTodayISO()) {
          await fetchToday()
        } else {
          todayRecord.value = data.record
        }
        let msg = '🌙 下班打卡成功！辛苦了~'
        if (data.exp_gained > 0) msg += ` +${data.exp_gained} EXP`
        if (data.leveled_up) msg += ` 🎉 升级到 Lv.${data.new_level}！`
        showToast(msg, 'success')
        emit('record-updated')
        fetchTodayTitle()
      } catch (err) {
        showToast(err.message || '打卡失败', 'error')
      } finally {
        loading.value = false
      }
    }

    // 手动补卡
    const handleManual = async () => {
      manualError.value = ''
      if (!manualForm.date || !manualForm.clock_in || !manualForm.clock_out) {
        manualError.value = '请填写完整信息'
        return
      }
      loading.value = true
      try {
        await clockinAPI.manual({ ...manualForm })
        showToast('📝 打卡记录已保存！')
        showManual.value = false
        fetchToday()
        emit('record-updated')
      } catch (err) {
        manualError.value = err.message || '保存失败'
      } finally {
        loading.value = false
      }
    }

    // 开始调整打卡时间
    const startAdjust = (type) => {
      adjusting.value = type
      // 预填当前打卡时间
      adjustTime.value = type === 'clock_in' ? todayRecord.value.clock_in : todayRecord.value.clock_out
    }

    // 取消调整
    const cancelAdjust = () => {
      adjusting.value = null
      adjustTime.value = ''
    }

    // 确认调整
    const confirmAdjust = async () => {
      if (!adjustTime.value) {
        showToast('请选择时间', 'error')
        return
      }
      loading.value = true
      try {
        const data = await clockinAPI.adjust({
          date: todayRecord.value?.date,
          type: adjusting.value,
          time: adjustTime.value
        })
        if (data.record?.date && data.record.date !== getTodayISO()) {
          await fetchToday()
        } else {
          todayRecord.value = data.record
        }
        showToast('🕐 打卡时间已调整！')
        adjusting.value = null
        adjustTime.value = ''
        emit('record-updated')
      } catch (err) {
        showToast(err.message || '调整失败', 'error')
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      updateClock()
      timer = setInterval(updateClock, 1000)
      fetchToday()
    })

    onUnmounted(() => {
      if (timer) clearInterval(timer)
    })

    return {
      todayRecord, todayTitle, loading, showManual, manualError, toast,
      currentTime, currentDate, liveHours, isOvernightPending,
      manualForm, handleClockIn, handleClockOut, handleManual,
      adjusting, adjustTime, startAdjust, cancelAdjust, confirmAdjust
    }
  }
}
</script>

<style scoped>
.font-mono { font-family: var(--font-mono); }

.panel-wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* 时钟 */
.clock-display {
  text-align: center;
  padding: var(--space-8);
}

.clock-time {
  font-size: var(--text-5xl);
  font-weight: 700;
  background: var(--gradient-text);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: 0.05em;
}

.clock-date {
  font-size: var(--text-sm);
  color: var(--text-muted);
  margin-top: var(--space-2);
}

/* 打卡按钮 */
.clock-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-6);
}

.action-card {
  padding: var(--space-6);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  transition: all 0.3s ease;
}

.action-card.done {
  border-color: rgba(168, 85, 247, 0.2);
}

.action-icon {
  font-size: 2.5rem;
}

.action-label {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--text-primary);
}

.record-date {
  font-size: var(--text-xs);
  color: var(--neon-pink);
  background: rgba(244, 114, 182, 0.12);
  border: 1px solid rgba(244, 114, 182, 0.25);
  border-radius: var(--radius-full);
  padding: 2px 10px;
}

.action-time {
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-primary-light);
}

.action-status {
  font-size: var(--text-sm);
  color: var(--color-primary-light);
}

.action-status.waiting {
  color: var(--text-muted);
}

.action-status {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-wrap: wrap;
  justify-content: center;
}

/* 调整按钮 */
.adjust-btn {
  background: none;
  border: 1px dashed rgba(168, 85, 247, 0.4);
  border-radius: var(--radius-full);
  padding: 2px 10px;
  min-height: 36px;
  font-size: var(--text-xs);
  color: var(--color-primary-light);
  cursor: pointer;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.adjust-btn:hover {
  border-color: var(--color-primary);
  background: rgba(168, 85, 247, 0.1);
  box-shadow: 0 0 10px rgba(168, 85, 247, 0.2);
}

/* 调整表单 */
.adjust-form {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  animation: adjust-in 0.3s ease;
}

@keyframes adjust-in {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.adjust-input {
  padding: var(--space-2) var(--space-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--color-primary);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: var(--text-lg);
  font-family: var(--font-mono);
  text-align: center;
  outline: none;
  width: 130px;
  box-shadow: 0 0 10px rgba(168, 85, 247, 0.2);
}

.adjust-input:focus {
  box-shadow: 0 0 15px rgba(168, 85, 247, 0.4);
}

.adjust-actions {
  display: flex;
  gap: var(--space-2);
}

.adjust-confirm,
.adjust-cancel {
  width: 36px;
  height: 36px;
  min-width: 44px;
  min-height: 44px;
  border-radius: 50%;
  border: none;
  font-size: var(--text-base);
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.adjust-confirm {
  background: var(--gradient-primary);
  color: white;
}

.adjust-confirm:hover:not(:disabled) {
  box-shadow: 0 2px 15px rgba(168, 85, 247, 0.5);
  transform: scale(1.1);
}

.adjust-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.adjust-cancel {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  border: 1px solid var(--border-color);
}

.adjust-cancel:hover {
  background: rgba(244, 114, 182, 0.15);
  border-color: var(--neon-pink);
  color: var(--neon-pink);
}

.action-btn {
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

.action-btn:hover:not(:disabled) {
  box-shadow: 0 4px 25px rgba(168, 85, 247, 0.5);
  transform: translateY(-3px) scale(1.05);
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 今日汇总 */
.today-summary {
  padding: var(--space-5);
  display: flex;
  gap: var(--space-6);
  flex-wrap: wrap;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.summary-icon {
  font-size: 1.2rem;
}

.summary-label {
  font-size: var(--text-sm);
  color: var(--text-muted);
}

.summary-value {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--text-primary);
}

.title-text {
  font-family: var(--font-sans);
}

.title-sub {
  font-size: var(--text-xs);
  color: var(--text-muted);
  font-weight: 400;
  margin-left: var(--space-1);
}

/* 手动补卡 */
.manual-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.manual-toggle {
  background: none;
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--space-3);
  min-height: 44px;
  color: var(--text-muted);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.3s ease;
}

.manual-toggle:hover {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

.manual-form {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.manual-hint {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--text-xs);
  line-height: 1.6;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: var(--space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-size: var(--text-xs);
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

.error-msg {
  color: var(--neon-pink);
  font-size: var(--text-sm);
}

.save-btn {
  padding: var(--space-2) var(--space-6);
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
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Toast */
.toast {
  position: fixed;
  bottom: var(--space-8);
  left: 50%;
  transform: translateX(-50%);
  padding: var(--space-3) var(--space-6);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  font-weight: 500;
  z-index: 10000;
  animation: toast-in 0.3s ease;
}

.toast.success {
  background: rgba(168, 85, 247, 0.9);
  color: white;
  box-shadow: 0 4px 20px rgba(168, 85, 247, 0.4);
}

.toast.error {
  background: rgba(244, 114, 182, 0.9);
  color: white;
  box-shadow: 0 4px 20px rgba(244, 114, 182, 0.4);
}

@keyframes toast-in {
  from { opacity: 0; transform: translateX(-50%) translateY(20px); }
  to { opacity: 1; transform: translateX(-50%) translateY(0); }
}

@media (max-width: 768px) {
  .clock-actions { grid-template-columns: 1fr; }
  .clock-time { font-size: var(--text-4xl); }
  .today-summary { flex-direction: column; gap: var(--space-3); }
  .today-summary .summary-item { gap: var(--space-2); }
  /* 补卡表单移动端优化 */
  .form-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
  .manual-form { padding: var(--space-4); }
  .action-card { padding: var(--space-4); }
  .clock-display { padding: var(--space-5); }
  .action-icon { font-size: 2rem; }
  .action-time { font-size: var(--text-xl); }
  .adjust-btn { min-height: 44px; padding: var(--space-2) var(--space-3); }
  .toast { bottom: calc(var(--space-16) + env(safe-area-inset-bottom, 0px)); max-width: 90vw; text-align: center; }
}
</style>
