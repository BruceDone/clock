<template>
  <el-container class="layout-container">
    <el-aside :width="collapsed ? '64px' : '200px'" class="aside">
      <div class="logo">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
            <path d="M12 6V12L16 14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <span v-show="!collapsed" class="logo-text">CLOCK</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        router
        class="side-menu"
      >
        <el-menu-item index="/home">
          <el-icon><Monitor /></el-icon>
          <span>首页</span>
        </el-menu-item>
        <el-menu-item index="/container/list">
          <el-icon><Box /></el-icon>
          <span>容器管理</span>
        </el-menu-item>
        <el-menu-item index="/task/list">
          <el-icon><List /></el-icon>
          <span>任务管理</span>
        </el-menu-item>
        <el-menu-item index="/status">
          <el-icon><VideoPlay /></el-icon>
          <span>实时状态</span>
        </el-menu-item>
        <el-menu-item index="/log/list">
          <el-icon><Document /></el-icon>
          <span>日志中心</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="toggleSidebar">
            <Fold v-if="!collapsed" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
              {{ item.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <ThemeSwitcher />
          <div class="system-status">
            <span class="status-dot"></span>
            <span class="status-text">系统运行中</span>
          </div>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><User /></el-icon>
              <span class="user-name">{{ userStore.userName }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu class="user-dropdown">
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import ThemeSwitcher from '@/components/ThemeSwitcher/index.vue'

const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()

const collapsed = computed(() => appStore.sidebarCollapsed)
const breadcrumbs = computed(() => appStore.breadcrumbs)
const activeMenu = computed(() => route.path)

function toggleSidebar() {
  appStore.toggleSidebar()
}

function handleCommand(command: string) {
  if (command === 'logout') {
    userStore.handleLogOut()
  }
}
</script>

<style lang="scss" scoped>
// Layout 组件样式 - 使用 CSS 变量

.layout-container {
  height: 100vh;
}

.aside {
  background: var(--bg-card);
  backdrop-filter: blur(10px);
  border-right: 1px solid var(--border-color);
  transition: width var(--transition-base);
  display: flex;
  flex-direction: column;

  .logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    border-bottom: 1px solid var(--border-color-light);
    padding: 0 16px;

    .logo-icon {
      width: 32px;
      height: 32px;
      color: var(--primary-color);
      animation: clock-pulse 2s ease-in-out infinite;

      svg {
        width: 100%;
        height: 100%;
      }
    }

    .logo-text {
      color: var(--text-primary);
      font-size: 20px;
      font-weight: 700;
      letter-spacing: 3px;
      background: linear-gradient(135deg, var(--primary-color), var(--info-color));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
  }

  .side-menu {
    flex: 1;
    border-right: none;
    background: transparent !important;
    padding: 12px 8px;

    :deep(.el-menu-item) {
      margin: 4px 0;
      border-radius: var(--border-radius-base);
      color: var(--text-secondary);
      transition: all var(--transition-base);
      height: 48px;
      line-height: 48px;

      &:hover {
        color: var(--primary-color);
        background: rgba(var(--primary-color), 0.1) !important;

        .el-icon {
          color: var(--primary-color);
        }
      }

      &.is-active {
        color: var(--primary-color);
        background: rgba(var(--primary-color), 0.15) !important;
        border-left: 3px solid var(--primary-color);

        .el-icon {
          color: var(--primary-color);
        }
      }

      .el-icon {
        color: var(--text-muted);
        transition: color var(--transition-base);
      }
    }
  }
}

@keyframes clock-pulse {
  0%, 100% {
    opacity: 1;
    filter: drop-shadow(0 0 5px var(--primary-glow));
  }
  50% {
    opacity: 0.8;
    filter: drop-shadow(0 0 15px var(--primary-color));
  }
}

.header {
  background: var(--bg-glass);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;

  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;

    .collapse-btn {
      font-size: 18px;
      color: var(--primary-color);
      cursor: pointer;
      transition: all var(--transition-base);
      padding: 8px;
      width: 36px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: var(--border-radius-base);
      background: transparent;
      border: 1px solid var(--border-color-light);

      &:hover {
        background: var(--bg-secondary);
        border-color: var(--primary-color);
        box-shadow: 0 0 12px var(--primary-glow);
      }

      &:active {
        transform: scale(0.92);
      }
    }

    :deep(.el-breadcrumb) {
      .el-breadcrumb__inner {
        color: var(--text-muted);
        font-weight: 500;

        &.is-link:hover {
          color: var(--primary-color);
        }
      }

      .el-breadcrumb__item:last-child .el-breadcrumb__inner {
        color: var(--text-primary);
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 20px;

    .system-status {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 6px 12px;
      background: rgba(var(--success-color), 0.1);
      border: 1px solid rgba(var(--success-color), 0.3);
      border-radius: 20px;

      .status-dot {
        width: 8px;
        height: 8px;
        background: var(--success-color);
        border-radius: 50%;
        animation: status-pulse 2s ease-in-out infinite;
      }

      .status-text {
        font-size: 12px;
        color: var(--success-color);
        font-weight: 500;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
      color: var(--text-regular);
      padding: 8px 12px;
      border-radius: var(--border-radius-base);
      transition: all var(--transition-base);

      .el-icon {
        color: var(--text-secondary);
      }

      .user-name {
        font-weight: 500;
      }

      &:hover {
        color: var(--primary-color);
        background: rgba(var(--primary-color), 0.1);

        .el-icon {
          color: var(--primary-color);
        }
      }
    }
  }
}

@keyframes status-pulse {
  0%, 100% {
    opacity: 1;
    box-shadow: 0 0 0 0 rgba(var(--success-color), 0.4);
  }
  50% {
    opacity: 0.7;
    box-shadow: 0 0 0 6px rgba(var(--success-color), 0);
  }
}

.main {
  background: transparent;
  padding: 24px;
  overflow-y: auto;
  height: calc(100vh - 64px);
}

:deep(.user-dropdown) {
  background: var(--bg-card) !important;
  border: 1px solid var(--border-color) !important;
  backdrop-filter: blur(10px);

  .el-dropdown-menu__item {
    color: var(--text-regular);

    &:hover {
      background: rgba(var(--primary-color), 0.1);
      color: var(--primary-color);
    }

    .el-icon {
      color: var(--text-muted);
    }
  }
}
</style>
