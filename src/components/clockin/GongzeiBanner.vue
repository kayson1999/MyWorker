<template>
  <transition name="banner-slide">
    <div v-if="visible && list.length > 0" class="gongzei-banner">
      <div class="banner-content">
        <div class="banner-track" :style="trackStyle">
          <!-- 重复两份实现无缝滚动 -->
          <span class="banner-text" v-for="(_, idx) in 2" :key="idx">
            <span class="banner-icon">🏴</span>
            <span class="banner-label">上周工贼榜</span>
            <span v-for="(item, i) in list" :key="i" class="banner-item">
              <span class="item-rank">{{ ['🥇','🥈','🥉'][i] }}</span>
              <span class="item-avatar">{{ item.avatar }}</span>
              <span class="item-name">{{ item.nickname }}</span>
              <span class="item-hours">{{ item.label }}</span>
            </span>
            <span class="banner-gap"></span>
          </span>
        </div>
      </div>
      <button class="banner-close" @click="closeBanner" title="关闭">✕</button>
    </div>
  </transition>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { gongzeiAPI } from '@/services/clockin.api.js'

export default {
  name: 'GongzeiBanner',
  setup() {
    const visible = ref(true)
    const list = ref([])
    const week = ref('')

    // 动画持续时间（秒），根据内容长度动态调整
    const animDuration = computed(() => {
      const base = 12
      return base + list.value.length * 3
    })

    const trackStyle = computed(() => ({
      animationDuration: `${animDuration.value}s`
    }))

    const fetchData = async () => {
      try {
        const data = await gongzeiAPI.getTop()
        list.value = data.list || []
        week.value = data.week || ''
      } catch (e) {
        // 未登录或请求失败时静默处理
        list.value = []
      }
    }

    const closeBanner = () => {
      visible.value = false
    }

    onMounted(() => {
      fetchData()
    })

    return { visible, list, week, trackStyle, closeBanner }
  }
}
</script>

<style scoped>
.gongzei-banner {
  position: relative;
  width: 100%;
  background: linear-gradient(90deg, rgba(168, 85, 247, 0.12), rgba(236, 72, 153, 0.12), rgba(168, 85, 247, 0.12));
  border-bottom: 1px solid rgba(168, 85, 247, 0.2);
  overflow: hidden;
  height: 36px;
  display: flex;
  align-items: center;
  z-index: 100;
}

.banner-content {
  flex: 1;
  overflow: hidden;
  position: relative;
  height: 100%;
  display: flex;
  align-items: center;
  mask-image: linear-gradient(to right, transparent 0%, black 5%, black 95%, transparent 100%);
  -webkit-mask-image: linear-gradient(to right, transparent 0%, black 5%, black 95%, transparent 100%);
}

.banner-track {
  display: inline-flex;
  white-space: nowrap;
  animation: scroll-left linear infinite;
  will-change: transform;
}

@keyframes scroll-left {
  0% {
    transform: translateX(0);
  }
  100% {
    transform: translateX(-50%);
  }
}

.banner-text {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.banner-icon {
  font-size: 14px;
}

.banner-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--neon-pink, #f472b6);
  margin-right: 8px;
  letter-spacing: 0.5px;
}

.banner-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 10px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  border: 1px solid rgba(168, 85, 247, 0.15);
  margin-right: 6px;
}

.item-rank {
  font-size: 14px;
}

.item-avatar {
  font-size: 14px;
}

.item-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary, #e2e8f0);
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-hours {
  font-size: 11px;
  font-weight: 700;
  color: var(--neon-cyan, #22d3ee);
  background: rgba(34, 211, 238, 0.1);
  padding: 1px 6px;
  border-radius: 8px;
}

.banner-gap {
  display: inline-block;
  width: 60px;
}

.banner-close {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  /* 扩大触摸目标到44px，满足移动端最小触摸区域要求 */
  min-width: 44px;
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  color: var(--text-muted, #94a3b8);
  font-size: 11px;
  cursor: pointer;
  margin-right: 4px;
  transition: all 0.2s ease;
  line-height: 1;
}

.banner-close:hover {
  background: rgba(244, 114, 182, 0.2);
  color: var(--neon-pink, #f472b6);
  border-color: rgba(244, 114, 182, 0.3);
}

/* 进出动画 */
.banner-slide-enter-active,
.banner-slide-leave-active {
  transition: all 0.4s ease;
}

.banner-slide-enter-from,
.banner-slide-leave-to {
  height: 0;
  opacity: 0;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .gongzei-banner {
    height: 32px;
  }
  .banner-label {
    font-size: 11px;
  }
  .item-name {
    max-width: 60px;
    font-size: 11px;
  }
  .item-hours {
    font-size: 10px;
  }
  .banner-close {
    width: 24px;
    height: 24px;
    min-width: 44px;
    min-height: 44px;
    font-size: 10px;
    margin-right: 2px;
  }
}
</style>
