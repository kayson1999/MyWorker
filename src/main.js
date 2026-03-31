/**
 * 打工人打卡 — 应用入口
 */
import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'
import './styles/global.css'

const app = createApp(App)
app.use(router)
app.mount('#app')
