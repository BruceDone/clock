<template>
  <el-container class="layout-container">
    <el-aside :width="collapsed ? '64px' : '200px'" class="aside">
      <div class="logo">
        <img src="@/assets/logo.svg" alt="logo" />
        <span v-show="!collapsed">Clock</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="collapsed"
        router
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
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
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><User /></el-icon>
              {{ userStore.userName }}
            </span>
            <template #dropdown>
              <el-dropdown-menu>
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
.layout-container {
  height: 100vh;
}

.aside {
  background-color: #304156;
  transition: width 0.3s;

  .logo {
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    background-color: #263445;

    img {
      width: 32px;
      height: 32px;
    }

    span {
      color: #fff;
      font-size: 18px;
      font-weight: bold;
    }
  }

  .el-menu {
    border-right: none;
  }
}

.header {
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;

  .header-left {
    display: flex;
    align-items: center;
    gap: 15px;

    .collapse-btn {
      font-size: 20px;
      cursor: pointer;
    }
  }

  .header-right {
    .user-info {
      display: flex;
      align-items: center;
      gap: 5px;
      cursor: pointer;
      color: #5a5e66;
    }
  }
}

.main {
  background-color: #f5f7fa;
  padding: 20px;
}
</style>
