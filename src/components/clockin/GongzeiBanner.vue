<template>
  <transition name="banner-slide">
    <div v-if="visible && hasData" class="gongzei-banner">
      <div class="banner-rows">
        <!-- 第一行：工贼榜（工时最高） -->
        <div v-if="gongzeiList.length > 0" class="banner-row">
          <div class="banner-content">
            <div class="banner-track" :style="trackStyle(gongzeiList.length)">
              <span class="banner-text" v-for="(_, idx) in 2" :key="'gz-copy-' + idx">
                <span class="banner-icon">🏴</span>
                <span class="banner-label label-gongzei">上周工贼榜</span>
                <span v-for="(item, i) in gongzeiList" :key="'gz-' + i" class="banner-item item-gongzei">
                  <span class="item-rank">{{ ['🥇','🥈','🥉'][i] }}</span>
                  <span class="item-avatar">{{ item.avatar }}</span>
                  <span class="item-name">{{ item.nickname }}</span>
                  <span class="item-hours hours-gongzei">{{ item.label }}</span>
                </span>
                <span class="banner-gap"></span>
              </span>
            </div>
          </div>
        </div>
        <!-- 第二行：光荣榜（工时最短） -->
        <div v-if="guangrongList.length > 0" class="banner-row">
          <div class="banner-content">
            <div class="banner-track track-reverse" :style="trackStyle(guangrongList.length)">
              <span class="banner-text" v-for="(_, idx) in 2" :key="'gr-copy-' + idx">
                <span class="banner-icon">🏅</span>
                <span class="banner-label label-guangrong">上周光荣榜</span>
                <span v-for="(item, i) in guangrongList" :key="'gr-' + i" class="banner-item item-guangrong">
                  <span class="item-rank">{{ ['🥇','🥈','🥉'][i] }}</span>
                  <span class="item-avatar">{{ item.avatar }}</span>
                  <span class="item-name">{{ item.nickname }}</span>
                  <span class="item-hours hours-guangrong">{{ item.label }}</span>
                </span>
                <span class="banner-gap"></span>
              </span>
            </div>
          </div>
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
    const gongzeiList = ref([])
    const guangrongList = ref([])
    const week = ref('')

    // 是否有数据可展示
    const hasData = computed(() => gongzeiList.value.length > 0 || guangrongList.value.length > 0)

    // 根据每行的项数动态计算动画时长
    const trackStyle = (count) => ({
      animationDuration: `${12 + count * 3}s`
    })

    const fetchData = async () => {
      try {
        const data = await gongzeiAPI.getAll()
        gongzeiList.value = data.gongzei?.list || []
        guangrongList.value = data.guangrong?.list || []
        week.value = data.gongzei?.week || data.guangrong?.week || ''
      } catch (e) {
        gongzeiList.value = []
        guangrongList.value = []
      }
    }

    const closeBanner = () => {
      visible.value = false
    }

    onMounted(() => {
      fetchData()
    })

    return { visible, gongzeiList, guangrongList, week, hasData, trackStyle, closeBanner }
  }
}
</script>

<style scoped>
.gongzei-banner {
  position: relative;
  width: 100%;
  background: linear-gradient(180deg, rgba(168, 85, 247, 0.10), rgba(34, 197, 94, 0.08));
  border-bottom: 1px solid rgba(168, 85, 247, 0.2);
  overflow: hidden;
  display: flex;
  align-items: center;
  z-index: 100;
}

.banner-rows {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.banner-row {
  height: 30px;
  display: flex;
  align-items: center;
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

/* 光荣榜反向滚动，增加视觉层次感 */
.banner-track.track-reverse {
  animation-name: scroll-right;
}

@keyframes scroll-left {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

@keyframes scroll-right {
  0% { transform: translateX(-50%); }
  100% { transform: translateX(0); }
}

.banner-text {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.banner-icon {
  font-size: 13px;
}

.banner-label {
  font-size: 11px;
  font-weight: 700;
  margin-right: 8px;
  letter-spacing: 0.5px;
}

.label-gongzei {
  color: var(--neon-pink, #f472b6);
}

.label-guangrong {
  color: #4ade80;
}

.banner-item {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 1px 8px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 20px;
  margin-right: 6px;
}

.item-gongzei {
  border: 1px solid rgba(168, 85, 247, 0.15);
}

.item-guangrong {
  border: 1px solid rgba(74, 222, 128, 0.15);
}

.item-rank {
  font-size: 12px;
}

.item-avatar {
  font-size: 12px;
}

.item-name {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-primary, #e2e8f0);
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-hours {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 8px;
}

.hours-gongzei {
  color: var(--neon-cyan, #22d3ee);
  background: rgba(34, 211, 238, 0.1);
}

.hours-guangrong {
  color: #4ade80;
  background: rgba(74, 222, 128, 0.1);
}

.banner-gap {
  display: inline-block;
  width: 60px;
}

.banner-close {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
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
  max-height: 0;
  opacity: 0;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .banner-row {
    height: 26px;
  }
  .banner-label {
    font-size: 10px;
  }
  .item-name {
    max-width: 60px;
    font-size: 10px;
  }
  .item-hours {
    font-size: 9px;
  }
  .item-rank,
  .item-avatar {
    font-size: 11px;
  }
  .banner-item {
    padding: 1px 6px;
    gap: 2px;
    margin-right: 4px;
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
