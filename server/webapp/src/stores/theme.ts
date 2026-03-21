import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'

export type ThemeName = 'light' | 'dark' | 'hacker' | 'premium'

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
  },
  {
    name: 'premium',
    label: 'Premium',
    icon: 'Star',
    description: 'Linear 风格，高端优雅'
  }
]

export const useThemeStore = defineStore('theme', () => {
  const savedTheme = localStorage.getItem('theme') as ThemeName | null
  const currentTheme = ref<ThemeName>(savedTheme || 'premium')

  const currentThemeInfo = computed(() => 
    themes.find(t => t.name === currentTheme.value) || themes[3]
  )

  function applyTheme(theme: ThemeName) {
    document.documentElement.setAttribute('data-theme', theme)
    document.body.setAttribute('data-theme', theme)
  }

  function setTheme(theme: ThemeName) {
    currentTheme.value = theme
    localStorage.setItem('theme', theme)
    applyTheme(theme)
  }

  function initTheme() {
    applyTheme(currentTheme.value)
  }

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
