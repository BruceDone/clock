<template>
  <el-container class="layout-container">
    <el-aside :width="collapsed ? '72px' : '240px'" class="aside">
      <div class="logo">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.5"/>
            <path d="M12 6V12L16 14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
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
// Layout 组件样式 - Premium 风格

.layout-container {
  height: 100vh;
}

.aside {
  background: var(--bg-card);
  backdrop-filter: blur(20px);
  border-right: 1px solid var(--border-color);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 10;

  &::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    width: 1px;
    height: 100%;
    background: linear-gradient(180deg, transparent, var(--primary-glow), transparent);
    opacity: 0.5;
  }

  .logo {
    height: 72px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 14px;
    border-bottom: 1px solid var(--border-color);
    padding: 0 20px;

    .logo-icon {
      width: 36px;
      height: 36px;
      color: var(--primary-color);
      transition: all 0.3s;

      svg {
        width: 100%;
        height: 100%;
        filter: drop-shadow(0 0 8px var(--primary-glow));
      }
    }

    &:hover .logo-icon {
      transform: rotate(15deg);
      filter: drop-shadow(0 0 12px var(--primary-color));
    }

    .logo-text {
      color: var(--text-primary);
      font-size: 18px;
      font-weight: 700;
      letter-spacing: 4px;
      background: linear-gradient(135deg, var(--primary-color), var(--primary-dim));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
  }

  .side-menu {
    flex: 1;
    border-right: none;
    background: transparent !important;
    padding: 16px 12px;

    :deep(.el-menu-item) {
      margin: 6px 0;
      border-radius: 12px;
      color: var(--text-secondary);
      transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
      height: 48px;
      line-height: 48px;
      padding-left: 16px !important;
      position: relative;
      overflow: hidden;

      &::before {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        transform: translateY(-50%);
        width: 3px;
        height: 0;
        background: var(--primary-color);
        border-radius: 0 2px 2px 0;
        transition: height 0.2s;
      }

      &:hover {
        color: var(--primary-color);
        background: var(--primary-glow) !important;

        .el-icon {
          color: var(--primary-color);
        }
      }

      &.is-active {
        color: var(--primary-color);
        background: var(--primary-glow) !important;
        font-weight: 600;

        &::before {
          height: 24px;
        }

        .el-icon {
          color: var(--primary-color);
        }
      }

      .el-icon {
        color: var(--text-muted);
        transition: color 0.2s;
        font-size: 18px;
        margin-right: 12px;
      }

      span {
        font-size: 14px;
        font-weight: 500;
      }
    }
  }
}

.header {
  background: var(--bg-glass);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  height: 72px;
  position: relative;
  z-index: 5;

  &::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, var(--border-color), transparent);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;

    .collapse-btn {
      font-size: 20px;
      color: var(--text-secondary);
      cursor: pointer;
      transition: all 0.2s;
      padding: 10px;
      width: 40px;
      height: 40px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 10px;
      background: transparent;
      border: 1px solid transparent;

      &:hover {
        color: var(--primary-color);
        background: var(--primary-glow);
        border-color: var(--border-color);
      }

      &:active {
        transform: scale(0.92);
      }
    }

    :deep(.el-breadcrumb) {
      .el-breadcrumb__inner {
        color: var(--text-muted);
        font-weight: 500;
        font-size: 14px;

        &.is-link:hover {
          color: var(--primary-color);
        }
      }

      .el-breadcrumb__item:last-child .el-breadcrumb__inner {
        color: var(--text-primary);
        font-weight: 600;
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .system-status {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 14px;
      background: rgba(34, 197, 94, 0.08);
      border: 1px solid rgba(34, 197, 94, 0.2);
      border-radius: 24px;

      .status-dot {
        width: 8px;
        height: 8px;
        background: var(--success-color);
        border-radius: 50%;
        box-shadow: 0 0 8px var(--success-color);
        animation: status-glow 2s ease-in-out infinite;
      }

      .status-text {
        font-size: 13px;
        color: var(--success-color);
        font-weight: 600;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;
      color: var(--text-regular);
      padding: 10px 14px;
      border-radius: 12px;
      transition: all 0.2s;
      border: 1px solid transparent;

      .el-icon {
        color: var(--text-secondary);
        font-size: 18px;
      }

      .user-name {
        font-weight: 600;
        font-size: 14px;
      }

      &:hover {
        color: var(--primary-color);
        background: var(--primary-glow);
        border-color: var(--border-color);

        .el-icon {
          color: var(--primary-color);
        }
      }
    }
  }
}

@keyframes status-glow {
  0%, 100% {
    box-shadow: 0 0 4px var(--success-color);
    opacity: 1;
  }
  50% {
    box-shadow: 0 0 12px var(--success-color);
    opacity: 0.7;
  }
}

.main {
  background: var(--bg-secondary);
  padding: 28px;
  overflow-y: auto;
  height: calc(100vh - 72px);
}

:deep(.user-dropdown) {
  background: var(--bg-card) !important;
  border: 1px solid var(--border-color) !important;
  backdrop-filter: blur(20px);
  border-radius: 12px !important;
  padding: 8px !important;
  box-shadow: 0 8px 32px var(--shadow-color);

  .el-dropdown-menu__item {
    color: var(--text-regular);
    border-radius: 8px;
    padding: 10px 16px;
    font-size: 14px;

    &:hover {
      background: var(--primary-glow);
      color: var(--primary-color);
    }

    .el-icon {
      color: var(--text-muted);
      margin-right: 10px;
    }
  }
}
</style>
