import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'

export type ThemeName = 'light' | 'dark' | 'hacker'

interface ThemeInfo {
  name: ThemeName
  label: string
  icon: string
  description: string
}

export const themes: ThemeInfo[] = [
  {
    name: 'light',
    label: '浅色',
    icon: 'Sunny',
    description: '清新明亮，适合白天使用'
  },
  {
    name: 'dark',
    label: '深色',
    icon: 'Moon',
    description: '柔和暗色，减少眼睛疲劳'
  },
  {
    name: 'hacker',
    label: '黑客终端',
    icon: 'Monitor',
    description: '绿色荧光，科技感十足'
  }
]

export const useThemeStore = defineStore('theme', () => {
  // 从 localStorage 读取主题，默认为 hacker
  const savedTheme = localStorage.getItem('theme') as ThemeName | null
  const currentTheme = ref<ThemeName>(savedTheme || 'hacker')

  // 获取当前主题信息 - 使用 computed 使其响应式
  const currentThemeInfo = computed(() => 
    themes.find(t => t.name === currentTheme.value) || themes[2]
  )

  // 应用主题到 DOM
  function applyTheme(theme: ThemeName) {
    document.documentElement.setAttribute('data-theme', theme)
    document.body.setAttribute('data-theme', theme)
  }

  // 切换主题
  function setTheme(theme: ThemeName) {
    currentTheme.value = theme
    localStorage.setItem('theme', theme)
    applyTheme(theme)
  }

  // 初始化主题
  function initTheme() {
    applyTheme(currentTheme.value)
  }

  // 监听主题变化
  watch(currentTheme, (theme) => {
    applyTheme(theme)
  })

  return {
    currentTheme,
    currentThemeInfo,
    themes,
    setTheme,
    initTheme
  }
})
