/**
 * 打卡业务 — 前端 API 封装层
 */

export { authAPI, userAPI } from './auth.api.js'
import { request } from './auth.api.js'

const withQuery = (path, params = {}) => {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      query.set(key, String(value))
    }
  })
  const queryString = query.toString()
  return queryString ? `${path}?${queryString}` : path
}

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
    return request(withQuery('/clockin/records', { start, end }))
  },

  getStats(period = 'week', offset = 0) {
    return request(withQuery('/clockin/stats', { period, offset }))
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
  getWorkhours(period = 'week', page = 1, pageSize = 10) {
    return request(withQuery('/ranking/workhours', { period, page, page_size: pageSize }))
  },

  getAvgWorkhours(period = 'week', page = 1, pageSize = 10) {
    return request(withQuery('/ranking/avgworkhours', { period, page, page_size: pageSize }))
  },

  getEarly(period = 'week', page = 1, pageSize = 10) {
    return request(withQuery('/ranking/early', { period, page, page_size: pageSize }))
  },

  getLate(period = 'week', page = 1, pageSize = 10) {
    return request(withQuery('/ranking/late', { period, page, page_size: pageSize }))
  },

  getStreak(page = 1, pageSize = 10) {
    return request(withQuery('/ranking/streak', { page, page_size: pageSize }))
  },

  getOntime(period = 'week', page = 1, pageSize = 10) {
    return request(withQuery('/ranking/ontime', { period, page, page_size: pageSize }))
  },

  getTitles(period = 'week', page = 1, pageSize = 10) {
    return request(withQuery('/ranking/titles', { period, page, page_size: pageSize }))
  }
}

// ==================== 工贼榜 & 光荣榜 ====================
export const gongzeiAPI = {
  getTop() {
    return request('/gongzei/top')
  },
  getGlory() {
    return request('/gongzei/glory')
  },
  // 一次请求同时获取工贼榜和光荣榜
  getAll() {
    return request('/gongzei/all')
  }
}

// ==================== 个人中心 ====================
export const userCenterAPI = {
  // 数据总览（等级、风格标签、核心统计、成就概览）
  getOverview() {
    return request('/usercenter/overview')
  },

  // 成就列表（分页）
  getAchievements(page = 1, pageSize = 12) {
    return request(withQuery('/usercenter/achievements', { page, page_size: pageSize }))
  },

  // 经验值日志
  getExpLogs() {
    return request('/usercenter/exp-logs')
  }
}
