import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login/index.vue'),
    meta: { title: '登录' }
  },
  {
    path: '/',
    component: () => import('@/components/Layout/index.vue'),
    redirect: '/home',
    children: [
      {
        path: 'home',
        name: 'Home',
        component: () => import('@/views/Home/index.vue'),
        meta: { title: '首页' }
      },
      {
        path: 'container/list',
        name: 'ContainerList',
        component: () => import('@/views/Container/List.vue'),
        meta: { title: '容器列表' }
      },
      {
        path: 'container/config/:cid',
        name: 'ContainerConfig',
        component: () => import('@/views/Container/Config.vue'),
        meta: { title: '容器配置' }
      },
      {
        path: 'task/list',
        name: 'TaskList',
        component: () => import('@/views/Task/List.vue'),
        meta: { title: '任务列表' }
      },
      {
        path: 'status',
        name: 'Status',
        component: () => import('@/views/Status/index.vue'),
        meta: { title: '实时状态' }
      },
      {
        path: 'log/list',
        name: 'LogList',
        component: () => import('@/views/Log/List.vue'),
        meta: { title: '日志中心' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/Error/404.vue'),
    meta: { title: '404' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  document.title = (to.meta.title as string) || 'Clock'

  const userStore = useUserStore()

  if (to.path !== '/login') {
    if (!userStore.isLoggedIn) {
      next({ name: 'Login' })
      return
    }
  }

  next()
})

export default router
