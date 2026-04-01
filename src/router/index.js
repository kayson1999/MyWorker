/**
 * 打工人打卡 — 路由配置
 */
import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  }
]

const router = createRouter({
  // 使用 Vite 注入的 BASE_URL 作为路由 base，支持子路径部署
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
