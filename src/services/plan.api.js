/**
 * 计划打卡 — 前端 API 封装层
 */
import { request } from './auth.api.js'

// ==================== 计划 CRUD ====================
export const planAPI = {
  // 获取计划列表
  getList(status = '') {
    const query = status ? `?status=${status}` : ''
    return request(`/plan/list${query}`)
  },

  // 创建计划
  create(data) {
    return request('/plan/create', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  // 更新计划
  update(data) {
    return request('/plan/update', {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  },

  // 删除（归档）计划
  delete(id) {
    return request('/plan/delete', {
      method: 'DELETE',
      body: JSON.stringify({ id })
    })
  },

  // 计划打卡
  checkin(planId, date = '', note = '') {
    return request('/plan/checkin', {
      method: 'POST',
      body: JSON.stringify({ plan_id: planId, date, note })
    })
  },

  // 取消打卡
  uncheckin(planId, date) {
    return request('/plan/uncheckin', {
      method: 'POST',
      body: JSON.stringify({ plan_id: planId, date })
    })
  },

  // 获取打卡记录
  getCheckinRecords(planId, start = '', end = '') {
    let query = `?plan_id=${planId}`
    if (start && end) query += `&start=${start}&end=${end}`
    return request(`/plan/checkin/records${query}`)
  },

  // 获取成就统计
  getAchievements() {
    return request('/plan/achievements')
  }
}

// ==================== 计划排行榜 ====================
export const planRankingAPI = {
  getTotal(period = 'all') {
    return request(`/plan-ranking/total?period=${period}`)
  },

  getStreak() {
    return request('/plan-ranking/streak')
  },

  getPlans() {
    return request('/plan-ranking/plans')
  },

  getCompletion() {
    return request('/plan-ranking/completion')
  }
}
