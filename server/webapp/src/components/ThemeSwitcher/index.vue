<template>
  <div class="theme-switcher">
    <div class="theme-tabs">
      <div 
        class="slider" 
        :style="{ transform: `translateX(${currentIndex * 100}%)` }"
      ></div>
      <div
        v-for="(theme, index) in themes"
        :key="theme.name"
        class="theme-tab"
        :class="{ active: theme.name === currentTheme }"
        :title="theme.description"
        @click="handleSwitch(theme.name)"
      >
        <el-icon :size="16">
          <component :is="theme.icon" />
        </el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useThemeStore, ThemeName, themes } from '@/stores/theme'
import { storeToRefs } from 'pinia'

const themeStore = useThemeStore()
const { currentTheme } = storeToRefs(themeStore)

const currentIndex = computed(() => 
  themes.findIndex(t => t.name === currentTheme.value)
)

function handleSwitch(theme: ThemeName) {
  themeStore.setTheme(theme)
}

onMounted(() => {
  themeStore.initTheme()
})
</script>

<style lang="scss" scoped>
.theme-switcher {
  display: flex;
  align-items: center;
}

.theme-tabs {
  position: relative;
  display: flex;
  align-items: center;
  background: var(--bg-secondary);
  border-radius: 20px;
  padding: 4px;
  gap: 2px;
}

.slider {
  position: absolute;
  left: 4px;
  width: calc(33.333% - 2px);
  height: calc(100% - 8px);
  background: var(--primary-color);
  border-radius: 16px;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.theme-tab {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 28px;
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-muted);

  &:hover:not(.active) {
    color: var(--text-primary);
  }

  &.active {
    color: #fff;
  }

  .el-icon {
    transition: transform 0.2s ease;
  }

  &:hover .el-icon {
    transform: scale(1.1);
  }

  &:active .el-icon {
    transform: scale(0.95);
  }
}
</style>
