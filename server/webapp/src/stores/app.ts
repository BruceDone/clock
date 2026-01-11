import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const breadcrumbs = ref<Array<{ title: string; path: string }>>([])

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setBreadcrumbs(list: Array<{ title: string; path: string }>) {
    breadcrumbs.value = list
  }

  return {
    sidebarCollapsed,
    breadcrumbs,
    toggleSidebar,
    setBreadcrumbs
  }
})
