/**
 * 打卡业务 — 前端 API 封装层
 */

export { authAPI, userAPI } from './auth.api.js'
import { request } from './auth.api.js'

// ==================== 打卡 ====================
export const clockinAPI = {
  clockIn() {
    return request('/clockin/in', { method: 'POST' })
  },

  clockOut() {
    return request('/clockin/out', { method: 'POST' })
  },

  manual(data) {
    return request('/clockin/manual', {
      method: 'POST',
      body: JSON.stringify(data)
    })
  },

  adjust(data) {
    return request('/clockin/adjust', {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  },

  getToday() {
    return request('/clockin/today')
  },

  getRecords(start, end) {
    return request(`/clockin/records?start=${start}&end=${end}`)
  },

  getStats(period = 'week') {
    return request(`/clockin/stats?period=${period}`)
  }
}

// ==================== 工作风格标签 ====================
export const titleAPI = {
  getTitles() {
    return request('/clockin/titles')
  },

  getTodayTitle() {
    return request('/clockin/today-title')
  }
}

// ==================== 排行榜 ====================
export const rankingAPI = {
  getWorkhours(period = 'week') {
    return request(`/ranking/workhours?period=${period}`)
  },

  getAvgWorkhours(period = 'week') {
    return request(`/ranking/avgworkhours?period=${period}`)
  },

  getEarly(period = 'week') {
    return request(`/ranking/early?period=${period}`)
  },

  getLate(period = 'week') {
    return request(`/ranking/late?period=${period}`)
  },

  getStreak() {
    return request('/ranking/streak')
  },

  getOntime(period = 'week') {
    return request(`/ranking/ontime?period=${period}`)
  },

  getTitles(period = 'week') {
    return request(`/ranking/titles?period=${period}`)
  }
}

// ==================== 工贼榜 ====================
export const gongzeiAPI = {
  getTop() {
    return request('/gongzei/top')
  }
}

// ==================== 个人中心 ====================
export const userCenterAPI = {
  // 数据总览（等级、风格标签、核心统计、成就概览）
  getOverview() {
    return request('/usercenter/overview')
  },

  // 成就列表
  getAchievements() {
    return request('/usercenter/achievements')
  },

  // 经验值日志
  getExpLogs() {
    return request('/usercenter/exp-logs')
  },

  // 打卡热力图
  getHeatmap() {
    return request('/usercenter/heatmap')
  }
}
