<template>
  <div class="plan-wrapper">
    <!-- 顶部操作栏 -->
    <div class="plan-header">
      <h3 class="section-title">📋 我的计划</h3>
      <button class="add-plan-btn" @click="showCreateModal = true">＋ 新建计划</button>
    </div>

    <!-- 计划列表 -->
    <div class="plan-list" v-if="plans.length > 0">
      <div class="plan-card glass-card" v-for="plan in plans" :key="plan.id"
           :class="{ active: selectedPlan?.id === plan.id, completed: plan.status === 'completed' }"
           @click="selectPlan(plan)">
        <div class="plan-card-header">
          <span class="plan-icon" :style="{ background: plan.color + '22' }">{{ plan.icon }}</span>
          <div class="plan-card-info">
            <span class="plan-card-title">{{ plan.title }}<span class="plan-visibility" v-if="plan.is_public === 0" title="仅自己可见">🔒</span></span>
            <span class="plan-card-content" v-if="plan.content">{{ plan.content }}</span>
            <span class="plan-card-meta">
              <template v-if="plan.status === 'completed'">🎉 已完成</template>
              <template v-else-if="getAchievement(plan.id)">
                {{ getAchievement(plan.id).total_days }}天 · 🔥{{ getAchievement(plan.id).current_streak }}天连续
              </template>
              <template v-else>刚创建</template>
            </span>
          </div>
          <div class="plan-card-actions">
            <button class="plan-action-btn" @click.stop="editPlan(plan)" title="编辑">✏️</button>
            <button class="plan-action-btn danger" @click.stop="deletePlan(plan)" title="归档">🗑️</button>
          </div>
        </div>
        <!-- 进度条（有目标时显示） -->
        <div class="plan-progress" v-if="plan.target_days > 0 && getAchievement(plan.id)">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: Math.min(getAchievement(plan.id).completion_rate, 100) + '%', background: plan.color }"></div>
          </div>
          <span class="progress-text">{{ getAchievement(plan.id).total_days }}/{{ plan.target_days }}天</span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="empty-state glass-card" v-if="!loading && plans.length === 0">
      <span class="empty-icon">🎯</span>
      <p>还没有计划，创建一个开始打卡吧！</p>
      <button class="create-first-btn" @click="showCreateModal = true">＋ 创建第一个计划</button>
    </div>

    <!-- 日历打卡区域 -->
    <div class="calendar-section glass-card" v-if="selectedPlan">
      <div class="calendar-header">
        <button class="cal-nav-btn" @click="prevMonth">◀</button>
        <h4 class="cal-title">{{ calendarYear }}年{{ calendarMonth }}月</h4>
        <button class="cal-nav-btn" @click="nextMonth">▶</button>
      </div>
      <div class="calendar-weekdays">
        <span v-for="d in weekdays" :key="d">{{ d }}</span>
      </div>
      <div class="calendar-grid">
        <div v-for="(day, idx) in calendarDays" :key="idx"
             class="calendar-day"
             :class="{
               empty: !day.date,
               checked: day.checked,
               today: day.isToday,
               future: day.isFuture,
               'other-month': day.otherMonth,
               'just-checked': day.date === justCheckedDate,
               'just-unchecked': day.date === justUncheckedDate
             }"
             @click="toggleCheckin(day)">
          <span class="day-num" v-if="day.date">{{ day.day }}</span>
        </div>
      </div>
      <div class="calendar-footer">
        <div class="calendar-legend">
          <span class="legend-item"><span class="legend-dot checked"></span> 已打卡</span>
          <span class="legend-item"><span class="legend-dot today"></span> 今天</span>
          <span class="legend-item"><span class="legend-dot checked-today"></span> 今天已打卡</span>
          <span class="legend-tip">点击日期打卡 · 支持补卡</span>
        </div>
        <!-- 打卡反馈 Toast -->
        <transition name="toast-fade">
          <div class="checkin-toast" v-if="toastMsg" :class="toastType">
            {{ toastMsg }}
          </div>
        </transition>
      </div>
    </div>

    <!-- 成就展示 -->
    <div class="achievements-section" v-if="achievements.length > 0">
      <h3 class="section-title">🏅 成就展示</h3>
      <div class="achievement-grid">
        <div class="achievement-card glass-card" v-for="a in achievements" :key="a.plan_id">
          <div class="achievement-icon" :style="{ background: a.color + '22' }">{{ a.icon }}</div>
          <div class="achievement-title">{{ a.title }}</div>
          <div class="achievement-stats">
            <div class="stat-item">
              <span class="stat-value">{{ a.total_days }}</span>
              <span class="stat-label">累计天数</span>
            </div>
            <div class="stat-item">
              <span class="stat-value fire">🔥{{ a.current_streak }}</span>
              <span class="stat-label">连续天数</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ a.max_streak }}</span>
              <span class="stat-label">最长连续</span>
            </div>
          </div>
          <div class="achievement-badges">
            <span class="badge" v-if="a.total_days >= 7">🌟 坚持7天</span>
            <span class="badge" v-if="a.total_days >= 30">💪 坚持30天</span>
            <span class="badge" v-if="a.total_days >= 100">🏆 百日达人</span>
            <span class="badge" v-if="a.total_days >= 365">👑 年度之星</span>
            <span class="badge" v-if="a.current_streak >= 7">🔥 连续7天</span>
            <span class="badge" v-if="a.current_streak >= 30">⚡ 连续30天</span>
            <span class="badge" v-if="a.max_streak >= 50">🎖️ 最长50天</span>
            <span class="badge" v-if="a.status === 'completed'">🎉 目标达成</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑计划弹窗 -->
    <div class="modal-overlay" v-if="showCreateModal || showEditModal" @click.self="closeModal">
      <div class="modal-content glass-card">
        <h3 class="modal-title">{{ showEditModal ? '编辑计划' : '创建新计划' }}</h3>
        <div class="form-group">
          <label>计划名称</label>
          <input v-model="formData.title" placeholder="例如：每日阅读、健身打卡..." maxlength="30" />
        </div>
        <div class="form-group">
          <label>计划内容 <span class="form-hint">（可选，描述你的计划详情）</span></label>
          <textarea v-model="formData.content" placeholder="例如：每天阅读30分钟技术书籍..." maxlength="200" rows="3" class="form-textarea"></textarea>
        </div>
        <div class="form-group">
          <label>图标</label>
          <div class="icon-picker">
            <span v-for="icon in iconOptions" :key="icon" class="icon-option"
                  :class="{ selected: formData.icon === icon }"
                  @click="formData.icon = icon">{{ icon }}</span>
          </div>
        </div>
        <div class="form-group">
          <label>主题色</label>
          <div class="color-picker">
            <span v-for="color in colorOptions" :key="color" class="color-option"
                  :class="{ selected: formData.color === color }"
                  :style="{ background: color }"
                  @click="formData.color = color"></span>
          </div>
        </div>
        <div class="form-group">
          <label>打卡频率</label>
          <div class="freq-picker">
            <button type="button" class="freq-option"
                    :class="{ selected: formData.frequency === 'daily' }"
                    @click="formData.frequency = 'daily'">📅 每天</button>
            <button type="button" class="freq-option"
                    :class="{ selected: formData.frequency === 'weekday' }"
                    @click="formData.frequency = 'weekday'">💼 工作日</button>
          </div>
        </div>
        <div class="form-group">
          <label>目标天数 <span class="form-hint">（0 = 无限期）</span></label>
          <input v-model.number="formData.target_days" type="number" min="0" max="9999" placeholder="0" />
        </div>
        <div class="form-group">
          <label>排行榜可见</label>
          <div class="visibility-switch">
            <button type="button" class="vis-option"
                    :class="{ selected: formData.is_public === 1 }"
                    @click="formData.is_public = 1">🌍 公开</button>
            <button type="button" class="vis-option"
                    :class="{ selected: formData.is_public === 0 }"
                    @click="formData.is_public = 0">🔒 仅自己</button>
          </div>
          <span class="form-hint">公开的计划会展示在排行榜中</span>
        </div>
        <div class="modal-actions">
          <button class="modal-btn cancel" @click="closeModal">取消</button>
          <button class="modal-btn confirm" @click="submitPlan" :disabled="!formData.title.trim()">
            {{ showEditModal ? '保存' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 加载中 -->
    <div class="loading-state" v-if="loading">
      <span class="loading-icon">⏳</span>
      <p>加载中...</p>
    </div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { planAPI } from '@/services/plan.api.js'

export default {
  name: 'PlanPanel',
  setup() {
    const plans = ref([])
    const achievements = ref([])
    const selectedPlan = ref(null)
    const checkinDates = ref(new Set())
    const loading = ref(false)
    const showCreateModal = ref(false)
    const showEditModal = ref(false)
    const toastMsg = ref('')
    const toastType = ref('success')
    const justCheckedDate = ref('')
    const justUncheckedDate = ref('')
    const checkinLock = ref(false)
    let toastTimer = null

    const calendarYear = ref(new Date().getFullYear())
    const calendarMonth = ref(new Date().getMonth() + 1)

    const weekdays = ['一', '二', '三', '四', '五', '六', '日']

    const iconOptions = ['🎯', '📚', '💪', '🏃', '🧘', '✍️', '🎨', '🎵', '💻', '🌱', '💧', '🍎', '😴', '🧠', '📝', '🔬']
    const colorOptions = ['#A855F7', '#6366F1', '#EC4899', '#F59E0B', '#10B981', '#3B82F6', '#EF4444', '#8B5CF6']

    const formData = reactive({
      id: 0,
      title: '',
      content: '',
      icon: '🎯',
      color: '#A855F7',
      frequency: 'daily',
      target_days: 0,
      is_public: 1,
      status: 'active'
    })

    // 日历天数计算
    const calendarDays = computed(() => {
      const year = calendarYear.value
      const month = calendarMonth.value
      const firstDay = new Date(year, month - 1, 1)
      const lastDay = new Date(year, month, 0)
      const today = new Date()
      const todayStr = formatDate(today)

      // 周一开始
      let startWeekday = firstDay.getDay()
      if (startWeekday === 0) startWeekday = 7
      startWeekday -= 1

      const days = []

      // 上月补位
      const prevMonthLastDay = new Date(year, month - 1, 0).getDate()
      for (let i = startWeekday - 1; i >= 0; i--) {
        const d = prevMonthLastDay - i
        const date = new Date(year, month - 2, d)
        days.push({
          day: d,
          date: formatDate(date),
          checked: checkinDates.value.has(formatDate(date)),
          isToday: false,
          isFuture: date > today,
          otherMonth: true
        })
      }

      // 本月
      for (let d = 1; d <= lastDay.getDate(); d++) {
        const date = new Date(year, month - 1, d)
        const dateStr = formatDate(date)
        days.push({
          day: d,
          date: dateStr,
          checked: checkinDates.value.has(dateStr),
          isToday: dateStr === todayStr,
          isFuture: date > today,
          otherMonth: false
        })
      }

      // 下月补位（补满6行 = 42格）
      const remaining = 42 - days.length
      for (let d = 1; d <= remaining; d++) {
        const date = new Date(year, month, d)
        days.push({
          day: d,
          date: formatDate(date),
          checked: checkinDates.value.has(formatDate(date)),
          isToday: false,
          isFuture: date > today,
          otherMonth: true
        })
      }

      return days
    })

    function formatDate(date) {
      const y = date.getFullYear()
      const m = String(date.getMonth() + 1).padStart(2, '0')
      const d = String(date.getDate()).padStart(2, '0')
      return `${y}-${m}-${d}`
    }

    function getAchievement(planId) {
      return achievements.value.find(a => a.plan_id === planId)
    }

    async function fetchPlans() {
      loading.value = true
      try {
        const data = await planAPI.getList()
        plans.value = (data.plans || []).filter(p => p.status !== 'archived')
        if (selectedPlan.value) {
          const found = plans.value.find(p => p.id === selectedPlan.value.id)
          if (found) {
            selectedPlan.value = found
          } else {
            selectedPlan.value = plans.value[0] || null
          }
        } else if (plans.value.length > 0) {
          selectedPlan.value = plans.value[0]
        }
      } catch (err) {
        console.error('获取计划列表失败:', err)
      } finally {
        loading.value = false
      }
    }

    async function fetchAchievements() {
      try {
        const data = await planAPI.getAchievements()
        achievements.value = data.achievements || []
      } catch (err) {
        console.error('获取成就失败:', err)
      }
    }

    async function fetchCheckinRecords() {
      if (!selectedPlan.value) return
      try {
        const year = calendarYear.value
        const month = calendarMonth.value
        // 多取前后一个月的数据用于日历补位显示
        const startMonth = month - 1 > 0 ? month - 1 : 12
        const startYear = month - 1 > 0 ? year : year - 1
        const start = `${startYear}-${String(startMonth).padStart(2, '0')}-01`
        const endMonth = month + 1 > 12 ? 1 : month + 1
        const endYear = month + 1 > 12 ? year + 1 : year
        const end = `${endYear}-${String(endMonth).padStart(2, '0')}-28`

        const data = await planAPI.getCheckinRecords(selectedPlan.value.id, start, end)
        const dates = new Set((data.records || []).map(r => r.date))
        checkinDates.value = dates
      } catch (err) {
        console.error('获取打卡记录失败:', err)
      }
    }

    function selectPlan(plan) {
      selectedPlan.value = plan
    }

    function showToast(msg, type = 'success') {
      toastMsg.value = msg
      toastType.value = type
      if (toastTimer) clearTimeout(toastTimer)
      toastTimer = setTimeout(() => { toastMsg.value = '' }, 2000)
    }

    async function toggleCheckin(day) {
      if (!day.date || day.isFuture || !selectedPlan.value) return
      if (selectedPlan.value.status !== 'active' && selectedPlan.value.status !== 'completed') return
      // 防止连续快速点击导致并发请求
      if (checkinLock.value) return
      checkinLock.value = true

      // 清除上次动画
      justCheckedDate.value = ''
      justUncheckedDate.value = ''

      try {
        if (day.checked) {
          await planAPI.uncheckin(selectedPlan.value.id, day.date)
          checkinDates.value.delete(day.date)
          checkinDates.value = new Set(checkinDates.value)
          justUncheckedDate.value = day.date
          showToast('已取消打卡 ✕', 'cancel')
        } else {
          await planAPI.checkin(selectedPlan.value.id, day.date)
          checkinDates.value.add(day.date)
          checkinDates.value = new Set(checkinDates.value)
          justCheckedDate.value = day.date
          const isToday = day.date === formatDate(new Date())
          showToast(isToday ? '打卡成功 ✅' : '补卡成功 ✅', 'success')
        }
        // 刷新成就
        fetchAchievements()
        fetchPlans()
        // 清除动画标记
        setTimeout(() => {
          justCheckedDate.value = ''
          justUncheckedDate.value = ''
        }, 600)
      } catch (err) {
        showToast(err.message || '操作失败', 'error')
      } finally {
        checkinLock.value = false
      }
    }

    function prevMonth() {
      if (calendarMonth.value === 1) {
        calendarMonth.value = 12
        calendarYear.value--
      } else {
        calendarMonth.value--
      }
    }

    function nextMonth() {
      if (calendarMonth.value === 12) {
        calendarMonth.value = 1
        calendarYear.value++
      } else {
        calendarMonth.value++
      }
    }

    function editPlan(plan) {
      formData.id = plan.id
      formData.title = plan.title
      formData.content = plan.content || ''
      formData.icon = plan.icon
      formData.color = plan.color
      formData.frequency = plan.freq_type
      formData.target_days = plan.target_days
      formData.is_public = plan.is_public ?? 1
      formData.status = plan.status
      showEditModal.value = true
    }

    async function deletePlan(plan) {
      if (!confirm(`确定要归档「${plan.title}」吗？归档后不会删除打卡记录。`)) return
      try {
        await planAPI.delete(plan.id)
        if (selectedPlan.value?.id === plan.id) {
          selectedPlan.value = null
        }
        fetchPlans()
        fetchAchievements()
      } catch (err) {
        alert(err.message || '操作失败')
      }
    }

    async function submitPlan() {
      if (!formData.title.trim()) return
      try {
        if (showEditModal.value) {
          await planAPI.update({ ...formData })
        } else {
          await planAPI.create({
            title: formData.title,
            content: formData.content,
            icon: formData.icon,
            color: formData.color,
            frequency: formData.frequency,
            target_days: formData.target_days,
            is_public: formData.is_public
          })
        }
        closeModal()
        fetchPlans()
        fetchAchievements()
      } catch (err) {
        alert(err.message || '操作失败')
      }
    }

    function closeModal() {
      showCreateModal.value = false
      showEditModal.value = false
      formData.id = 0
      formData.title = ''
      formData.content = ''
      formData.icon = '🎯'
      formData.color = '#A855F7'
      formData.frequency = 'daily'
      formData.target_days = 0
      formData.is_public = 1
      formData.status = 'active'
    }

    // 监听选中计划和月份变化，刷新打卡记录
    watch([selectedPlan, calendarYear, calendarMonth], () => {
      fetchCheckinRecords()
    })

    onMounted(() => {
      fetchPlans()
      fetchAchievements()
    })

    return {
      plans, achievements, selectedPlan, checkinDates, loading,
      showCreateModal, showEditModal, formData,
      calendarYear, calendarMonth, calendarDays, weekdays,
      iconOptions, colorOptions,
      toastMsg, toastType, justCheckedDate, justUncheckedDate,
      getAchievement, selectPlan, toggleCheckin, showToast,
      prevMonth, nextMonth, editPlan, deletePlan,
      submitPlan, closeModal
    }
  }
}
</script>

<style scoped>
.plan-wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* 顶部操作栏 */
.plan-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--text-primary);
}

.add-plan-btn {
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

.add-plan-btn:hover {
  box-shadow: 0 4px 20px rgba(168, 85, 247, 0.4);
  transform: translateY(-2px);
}

/* 计划列表 */
.plan-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.plan-card {
  padding: var(--space-4);
  cursor: pointer;
  transition: all 0.3s ease;
  border: 1px solid var(--border-color);
}

.plan-card:hover {
  border-color: var(--border-glow);
  transform: translateX(4px);
}

.plan-card.active {
  border-color: var(--color-primary);
  box-shadow: 0 0 20px rgba(168, 85, 247, 0.15);
}

.plan-card.completed {
  opacity: 0.7;
}

.plan-card-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.plan-icon {
  font-size: 1.5rem;
  width: 42px;
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-lg);
  flex-shrink: 0;
}

.plan-card-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.plan-card-title {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.plan-visibility {
  font-size: 11px;
  opacity: 0.6;
}

.plan-card-content {
  font-size: var(--text-xs);
  color: var(--text-secondary);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.plan-card-meta {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.plan-card-actions {
  display: flex;
  gap: var(--space-1);
  opacity: 0;
  transition: opacity 0.2s;
}

.plan-card:hover .plan-card-actions {
  opacity: 1;
}

.plan-action-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.plan-action-btn:hover {
  border-color: var(--color-primary);
}

.plan-action-btn.danger:hover {
  border-color: #EF4444;
  background: rgba(239, 68, 68, 0.1);
}

/* 进度条 */
.plan-progress {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-top: var(--space-3);
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.5s ease;
}

.progress-text {
  font-size: var(--text-xs);
  color: var(--text-muted);
  font-family: var(--font-mono);
  white-space: nowrap;
}

/* 日历 */
.calendar-section {
  padding: var(--space-5);
}

.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.cal-title {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--text-primary);
}

.cal-nav-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: var(--text-sm);
  transition: all 0.2s;
}

.cal-nav-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary-light);
}

.calendar-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  text-align: center;
  margin-bottom: var(--space-2);
}

.calendar-weekdays span {
  font-size: var(--text-xs);
  color: var(--text-muted);
  font-weight: 600;
  padding: var(--space-1) 0;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
}

.calendar-day {
  aspect-ratio: 1;
  max-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.calendar-day:hover:not(.empty):not(.future):not(.other-month) {
  background: var(--bg-tertiary);
  transform: scale(1.08);
}

.calendar-day.empty {
  cursor: default;
}

.calendar-day.other-month {
  opacity: 0.2;
}

.calendar-day.future {
  opacity: 0.25;
  cursor: not-allowed;
}

.calendar-day.today {
  background: rgba(168, 85, 247, 0.1);
  box-shadow: inset 0 0 0 2px var(--color-primary);
}

.calendar-day.checked {
  background: transparent;
  box-shadow: inset 0 0 0 2.5px #EF4444;
}

.calendar-day.checked .day-num {
  color: #EF4444;
  font-weight: 700;
}

/* 今天 + 已打卡 同时存在时：红圈 + 紫色内发光 */
.calendar-day.today.checked {
  box-shadow: inset 0 0 0 2.5px #EF4444, 0 0 8px rgba(239, 68, 68, 0.3);
  background: rgba(168, 85, 247, 0.08);
}

.calendar-day.today.checked .day-num {
  color: #EF4444;
  font-weight: 800;
}

.day-num {
  font-size: var(--text-sm);
  color: var(--text-secondary);
  font-weight: 500;
  z-index: 1;
  line-height: 1;
}

.calendar-footer {
  margin-top: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.calendar-legend {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.legend-dot.checked {
  background: transparent;
  box-shadow: inset 0 0 0 2px #EF4444;
}

.legend-dot.today {
  box-shadow: inset 0 0 0 2px var(--color-primary);
  background: rgba(168, 85, 247, 0.1);
}

.legend-dot.checked-today {
  box-shadow: inset 0 0 0 2px #EF4444;
  background: rgba(168, 85, 247, 0.15);
}

.legend-tip {
  margin-left: auto;
  font-size: var(--text-xs);
  color: var(--text-muted);
  font-style: italic;
}

/* 打卡反馈 Toast */
.checkin-toast {
  text-align: center;
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-lg);
  font-size: var(--text-sm);
  font-weight: 600;
  animation: toast-pop 0.3s ease;
}

.checkin-toast.success {
  background: rgba(16, 185, 129, 0.15);
  color: #10B981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.checkin-toast.cancel {
  background: rgba(245, 158, 11, 0.15);
  color: #F59E0B;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.checkin-toast.error {
  background: rgba(239, 68, 68, 0.15);
  color: #EF4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.toast-fade-enter-active { animation: toast-pop 0.3s ease; }
.toast-fade-leave-active { animation: toast-pop 0.3s ease reverse; }

@keyframes toast-pop {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 打卡/取消动画 */
.calendar-day.just-checked {
  animation: check-ring-pop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.calendar-day.just-unchecked {
  animation: uncheck-fade 0.35s ease;
}

@keyframes check-ring-pop {
  0% {
    transform: scale(0.6);
    box-shadow: inset 0 0 0 0px #EF4444;
    opacity: 0.5;
  }
  50% {
    transform: scale(1.2);
    box-shadow: inset 0 0 0 3px #EF4444, 0 0 16px rgba(239, 68, 68, 0.4);
  }
  100% {
    transform: scale(1);
    box-shadow: inset 0 0 0 2.5px #EF4444;
    opacity: 1;
  }
}

@keyframes uncheck-fade {
  0% {
    transform: scale(1);
    box-shadow: inset 0 0 0 2.5px #EF4444;
    opacity: 1;
  }
  50% {
    transform: scale(0.85);
    opacity: 0.4;
  }
  100% {
    transform: scale(1);
    box-shadow: none;
    opacity: 1;
  }
}

/* 频率选择器 */
.freq-picker {
  display: flex;
  gap: var(--space-2);
}

.freq-option {
  flex: 1;
  padding: var(--space-2) var(--space-4);
  background: var(--bg-tertiary);
  border: 2px solid var(--border-color);
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.freq-option:hover {
  border-color: var(--border-glow);
}

.freq-option.selected {
  border-color: var(--color-primary);
  background: rgba(168, 85, 247, 0.15);
  color: var(--color-primary-light);
  font-weight: 600;
}

/* 成就展示 */
.achievements-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.achievement-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: var(--space-4);
}

.achievement-card {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  text-align: center;
}

.achievement-icon {
  font-size: 2rem;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-xl);
}

.achievement-title {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--text-primary);
}

.achievement-stats {
  display: flex;
  gap: var(--space-5);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.stat-value {
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--color-primary-light);
  font-family: var(--font-mono);
}

.stat-value.fire {
  color: #F59E0B;
}

.stat-label {
  font-size: var(--text-xs);
  color: var(--text-muted);
}

.achievement-badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  justify-content: center;
}

.badge {
  padding: 2px var(--space-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  font-size: 11px;
  color: var(--text-secondary);
  white-space: nowrap;
}

/* 弹窗 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-content {
  width: 90%;
  max-width: 420px;
  padding: var(--space-8);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.modal-title {
  font-size: var(--text-lg);
  font-weight: 700;
  color: var(--text-primary);
  text-align: center;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-group label {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--text-secondary);
}

.form-hint {
  font-weight: 400;
  color: var(--text-muted);
  font-size: var(--text-xs);
}

.form-group input {
  padding: var(--space-3) var(--space-4);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  color: var(--text-primary);
  font-size: var(--text-sm);
  outline: none;
  transition: border-color 0.2s;
}

.form-group input:focus {
  border-color: var(--color-primary);
}

.form-textarea {
  padding: var(--space-3) var(--space-4);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  color: var(--text-primary);
  font-size: var(--text-sm);
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s;
  resize: vertical;
  min-height: 60px;
}

.form-textarea:focus {
  border-color: var(--color-primary);
}

.form-textarea::placeholder {
  color: var(--text-muted);
}

/* 可见性切换 */
.visibility-switch {
  display: flex;
  gap: var(--space-2);
}

.vis-option {
  flex: 1;
  padding: var(--space-2) var(--space-4);
  background: var(--bg-tertiary);
  border: 2px solid var(--border-color);
  border-radius: var(--radius-lg);
  color: var(--text-secondary);
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.vis-option:hover {
  border-color: var(--border-glow);
}

.vis-option.selected {
  border-color: var(--color-primary);
  background: rgba(168, 85, 247, 0.15);
  color: var(--color-primary-light);
  font-weight: 600;
}

.icon-picker {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.icon-option {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border: 2px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 1.2rem;
  transition: all 0.2s;
}

.icon-option:hover {
  border-color: var(--border-glow);
}

.icon-option.selected {
  border-color: var(--color-primary);
  background: rgba(168, 85, 247, 0.15);
}

.color-picker {
  display: flex;
  gap: var(--space-2);
}

.color-option {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-full);
  cursor: pointer;
  border: 3px solid transparent;
  transition: all 0.2s;
}

.color-option:hover {
  transform: scale(1.2);
}

.color-option.selected {
  border-color: white;
  box-shadow: 0 0 10px rgba(168, 85, 247, 0.5);
}

.modal-actions {
  display: flex;
  gap: var(--space-3);
  justify-content: flex-end;
}

.modal-btn {
  padding: var(--space-2) var(--space-6);
  border: none;
  border-radius: var(--radius-lg);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.modal-btn.cancel {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

.modal-btn.cancel:hover {
  border-color: var(--text-muted);
}

.modal-btn.confirm {
  background: var(--gradient-primary);
  color: white;
}

.modal-btn.confirm:hover:not(:disabled) {
  box-shadow: 0 4px 15px rgba(168, 85, 247, 0.4);
}

.modal-btn.confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 空状态 */
.empty-state {
  padding: var(--space-10);
  text-align: center;
  color: var(--text-muted);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.empty-icon {
  font-size: 3rem;
}

.create-first-btn {
  padding: var(--space-3) var(--space-6);
  background: var(--gradient-primary);
  color: white;
  border: none;
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.create-first-btn:hover {
  box-shadow: 0 4px 20px rgba(168, 85, 247, 0.4);
  transform: translateY(-2px);
}

/* 加载 */
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

/* 响应式 */
@media (max-width: 768px) {
  .plan-card-actions { opacity: 1; }
  .achievement-grid { grid-template-columns: 1fr; }
  .calendar-day { max-height: 36px; }
  .modal-content { padding: var(--space-5); }
  .freq-picker { flex-direction: column; }
}
</style>
