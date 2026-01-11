import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const router = useRouter()

  const token = ref(localStorage.getItem('token') || '')
  const userName = ref(localStorage.getItem('userName') || '')

  const isLoggedIn = computed(() => !!token.value)

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('token', t)
  }

  function setUserName(name: string) {
    userName.value = name
    localStorage.setItem('userName', name)
  }

  async function handleLogin(userNameVal: string, userPwd: string) {
    try {
      const res = await login({ user_name: userNameVal, user_pwd: userPwd })
      if (res.data && typeof res.data === 'string') {
        setToken(res.data)
        setUserName(userNameVal)
        return true
      }
      return false
    } catch {
      return false
    }
  }

  function handleLogOut() {
    token.value = ''
    userName.value = ''
    localStorage.removeItem('token')
    localStorage.removeItem('userName')
    router.push('/login')
  }

  return {
    token,
    userName,
    isLoggedIn,
    setToken,
    setUserName,
    handleLogin,
    handleLogOut
  }
})
