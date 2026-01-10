<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-form">
        <h2 class="title">Clock 任务调度平台</h2>
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
          <el-form-item prop="userName">
            <el-input v-model="form.userName" placeholder="用户名" prefix-icon="User" size="large" />
          </el-form-item>
          <el-form-item prop="userPwd">
            <el-input
              v-model="form.userPwd"
              type="password"
              placeholder="密码"
              prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" size="large" :loading="loading" style="width: 100%" @click="handleLogin">
              登录
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  userName: 'admin',
  userPwd: 'admin'
})

const rules: FormRules = {
  userName: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  userPwd: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const success = await userStore.handleLogin(form.userName, form.userPwd)
    if (success) {
      ElMessage.success('登录成功')
      router.push('/home')
    } else {
      ElMessage.error('登录失败')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-box {
  width: 400px;
  border-radius: 8px;
  overflow: hidden;
}

.login-form {
  padding: 40px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;

  .title {
    text-align: center;
    margin-bottom: 30px;
    color: #333;
    font-size: 24px;
  }
}
</style>
