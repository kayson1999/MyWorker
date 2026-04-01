/**
 * 统一认证 API 层
 *
 * 认证由统一用户中心（MyUserCenter）管理，
 * 前端通过本服务后端代理与用户中心通信。
 * 支持从 URL 参数接收主站传递的 SSO Token。
 */

// 根据 Vite 构建时注入的 base 路径，自动拼接 API 前缀
// 部署在 /Worker/ 下时 → BASE_URL = '/Worker/api'
// 本地开发（base='/'）时 → BASE_URL = '/api'
const _base = (import.meta.env.BASE_URL || '/').replace(/\/+$/, '')
const BASE_URL = `${_base}/api`
const TOKEN_KEY = 'worker_auth_token'

// ==================== SSO Token 接收 ====================

/**
 * 处理 SSO Token：检测 URL 中的 sso_token 参数，
 * 如果存在则存入 localStorage 并清除 URL 参数。
 * 应在应用初始化时、auth.store.init() 之前调用。
 */
export function handleSsoToken() {
  try {
    const params = new URLSearchParams(window.location.search)
    const ssoToken = params.get('sso_token')
    if (!ssoToken) return false

    // 将主站传递的 Token 存入本地（覆盖已有 Token）
    setToken(ssoToken)

    // 清除 URL 中的 sso_token 参数，避免泄露和重复处理
    params.delete('sso_token')
    const cleanSearch = params.toString()
    const cleanUrl = window.location.pathname + (cleanSearch ? '?' + cleanSearch : '') + window.location.hash
    window.history.replaceState(null, '', cleanUrl)

    console.log('[SSO] 已从主站同步登录态')
    return true
  } catch (e) {
    console.warn('[SSO] 处理 SSO Token 失败:', e)
    return false
  }
}

// ==================== Token 管理 ====================

/** 获取 token */
export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function isLoggedIn() {
  return !!getToken()
}

// ==================== 通用请求封装 ====================

export async function request(path, options = {}) {
  const token = getToken()
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    ...options
  }

  const res = await fetch(`${BASE_URL}${path}`, config)
  const data = await res.json()

  if (!res.ok) {
    if (res.status === 401) {
      clearToken()
      throw { code: 'AUTH_EXPIRED', message: data.error || '登录已过期' }
    }
    throw { code: 'API_ERROR', message: data.error || '请求失败', status: res.status }
  }

  return data
}

// ==================== 认证 API ====================

export const authAPI = {
  async register(data) {
    const result = await request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data)
    })
    setToken(result.token)
    return result
  },

  async login(data) {
    const result = await request('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data)
    })
    setToken(result.token)
    return result
  },

  async logout() {
    try {
      await request('/auth/logout', { method: 'POST' })
    } catch (err) {
      console.warn('登出请求失败:', err.message)
    } finally {
      clearToken()
    }
  }
}

// ==================== 用户 API ====================

export const userAPI = {
  getProfile() {
    return request('/user/profile')
  },

  updateProfile(data) {
    return request('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(data)
    })
  }
}
