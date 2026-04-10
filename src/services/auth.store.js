/**
 * 全局认证状态管理
 * 使用 Vue 3 响应式 API
 * 支持 SSO Token 自动同步
 */
import { ref, computed, readonly } from 'vue'
import { authAPI, userAPI, isLoggedIn as checkToken, clearToken, handleSsoToken } from './auth.api.js'

// ==================== 响应式状态 ====================

const currentUser = ref(null)
const loading = ref(false)
const showLoginModal = ref(false)

// ==================== 计算属性 ====================

const isLoggedIn = computed(() => !!currentUser.value)
const userAvatar = computed(() => currentUser.value?.avatar || '😎')
const userNickname = computed(() => currentUser.value?.nickname || '')

// ==================== 方法 ====================

/**
 * 初始化：处理 SSO Token + 检查 token 有效性
 */
async function init() {
  // 优先处理 SSO Token（从主站跳转过来时 URL 中携带）
  handleSsoToken()

  if (!checkToken()) return
  loading.value = true
  try {
    const data = await userAPI.getProfile()
    currentUser.value = data.user
  } catch (err) {
    // 只有认证真正过期才清除 token 和用户状态
    // 网络错误、服务端临时错误（503等）不清除，避免误退出登录
    if (err.code === 'AUTH_EXPIRED') {
      clearToken()
      currentUser.value = null
    } else {
      console.warn('获取用户信息失败（非认证过期）:', err.message)
    }
  } finally {
    loading.value = false
  }
}

/**
 * 登录
 */
async function login(credentials) {
  const result = await authAPI.login(credentials)
  currentUser.value = result.user
  showLoginModal.value = false
  return result
}

/**
 * 注册
 */
async function register(data) {
  const result = await authAPI.register(data)
  currentUser.value = result.user
  showLoginModal.value = false
  return result
}

/**
 * 退出登录
 */
async function logout() {
  await authAPI.logout()
  currentUser.value = null
}

/**
 * 打开登录弹窗
 */
function openLogin() {
  showLoginModal.value = true
}

/**
 * 关闭登录弹窗
 */
function closeLogin() {
  showLoginModal.value = false
}

/**
 * 更新用户信息
 */
function updateUser(user) {
  currentUser.value = user
}

// ==================== 导出 ====================

export const authStore = {
  currentUser: readonly(currentUser),
  loading: readonly(loading),
  showLoginModal,
  isLoggedIn,
  userAvatar,
  userNickname,

  init,
  login,
  register,
  logout,
  openLogin,
  closeLogin,
  updateUser
}
