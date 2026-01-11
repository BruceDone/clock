<template>
  <div class="login-container">
    <!-- 背景装饰 -->
    <div class="bg-grid"></div>
    <div class="bg-particles" ref="particlesRef"></div>

    <!-- 装饰性边框 -->
    <div class="corner-decor tl"></div>
    <div class="corner-decor tr"></div>
    <div class="corner-decor bl"></div>
    <div class="corner-decor br"></div>

    <!-- 登录框 -->
    <div class="login-box">
      <div class="login-header">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
            <path d="M12 6V12L16 14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </div>
        <h1 class="title">CLOCK</h1>
        <p class="subtitle">任务调度平台</p>
      </div>

      <div class="login-form">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
          <el-form-item prop="userName">
            <div class="input-wrapper">
              <el-icon class="input-icon"><User /></el-icon>
              <el-input
                v-model="form.userName"
                placeholder="用户名"
                size="large"
                @focus="onInputFocus"
                @blur="onInputBlur"
              />
            </div>
          </el-form-item>
          <el-form-item prop="userPwd">
            <div class="input-wrapper">
              <el-icon class="input-icon"><Lock /></el-icon>
              <el-input
                v-model="form.userPwd"
                type="password"
                placeholder="密码"
                size="large"
                show-password
                @focus="onInputFocus"
                @blur="onInputBlur"
              />
            </div>
          </el-form-item>
          <el-form-item>
            <button
              type="button"
              class="login-btn"
              :class="{ loading: loading }"
              :disabled="loading"
              @click="handleLogin"
            >
              <span class="btn-text">{{ loading ? '登录中...' : '登 录' }}</span>
              <span class="btn-glow"></span>
            </button>
          </el-form-item>
        </el-form>
      </div>

      <div class="login-footer">
        <div class="system-info">
          <span class="info-dot"></span>
          <span>系统运行正常</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const particlesRef = ref<HTMLElement>()

let particleInterval: number | null = null

const form = reactive({
  userName: 'admin',
  userPwd: 'admin'
})

const rules: FormRules = {
  userName: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  userPwd: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

// 创建粒子效果
function createParticles() {
  if (!particlesRef.value) return

  const container = particlesRef.value
  const particleCount = 80 // 增加粒子数量

  for (let i = 0; i < particleCount; i++) {
    const particle = document.createElement('div')
    // 随机分配不同类型的粒子
    const particleType = Math.random()
    if (particleType < 0.3) {
      particle.className = 'particle particle-large'
    } else if (particleType < 0.6) {
      particle.className = 'particle particle-glow'
    } else {
      particle.className = 'particle'
    }
    particle.style.left = Math.random() * 100 + '%'
    particle.style.top = Math.random() * 100 + '%'
    particle.style.animationDelay = Math.random() * 8 + 's'
    particle.style.animationDuration = (Math.random() * 4 + 4) + 's'
    container.appendChild(particle)
  }
}

function onInputFocus(e: FocusEvent) {
  const wrapper = (e.target as HTMLElement).closest('.input-wrapper')
  if (wrapper) {
    wrapper.classList.add('focused')
  }
}

function onInputBlur(e: FocusEvent) {
  const wrapper = (e.target as HTMLElement).closest('.input-wrapper')
  if (wrapper) {
    wrapper.classList.remove('focused')
  }
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

onMounted(() => {
  createParticles()
})

onUnmounted(() => {
  if (particleInterval) {
    clearInterval(particleInterval)
  }
})
</script>

<style lang="scss" scoped>
// Login 页面样式 - 使用 CSS 变量

.login-container {
  width: 100%;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  background: var(--bg-primary);
  overflow: hidden;
}

.bg-grid {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image:
    linear-gradient(rgba(var(--primary-color), 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(var(--primary-color), 0.03) 1px, transparent 1px);
  background-size: 50px 50px;
  animation: grid-move 20s linear infinite;
}

@keyframes grid-move {
  0% { transform: translate(0, 0); }
  100% { transform: translate(50px, 50px); }
}

.bg-particles {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;

  :deep(.particle) {
    position: absolute;
    width: 4px;
    height: 4px;
    background: var(--primary-color);
    border-radius: 50%;
    opacity: 0;
    animation: float-particle 6s ease-in-out infinite;
    box-shadow: 0 0 6px var(--primary-color);
  }

  :deep(.particle-large) {
    width: 8px;
    height: 8px;
    box-shadow: 0 0 12px var(--primary-color), 0 0 24px var(--primary-color);
    animation: float-particle-large 8s ease-in-out infinite;
  }

  :deep(.particle-glow) {
    width: 6px;
    height: 6px;
    background: linear-gradient(135deg, var(--primary-color), var(--info-color));
    box-shadow: 0 0 15px var(--primary-color), 0 0 30px rgba(var(--primary-color), 0.5);
    animation: float-particle-glow 7s ease-in-out infinite;
  }
}

@keyframes float-particle {
  0%, 100% {
    opacity: 0;
    transform: translateY(0) scale(0);
  }
  10% {
    opacity: 0.8;
    transform: translateY(-10px) scale(1);
  }
  90% {
    opacity: 0.8;
    transform: translateY(-150px) scale(1);
  }
  100% {
    opacity: 0;
    transform: translateY(-180px) scale(0);
  }
}

@keyframes float-particle-large {
  0%, 100% {
    opacity: 0;
    transform: translateY(0) scale(0) rotate(0deg);
  }
  10% {
    opacity: 0.9;
    transform: translateY(-20px) scale(1) rotate(45deg);
  }
  50% {
    opacity: 1;
    transform: translateY(-100px) scale(1.2) rotate(180deg);
  }
  90% {
    opacity: 0.9;
    transform: translateY(-200px) scale(1) rotate(315deg);
  }
  100% {
    opacity: 0;
    transform: translateY(-250px) scale(0) rotate(360deg);
  }
}

@keyframes float-particle-glow {
  0%, 100% {
    opacity: 0;
    transform: translateY(0) scale(0);
    filter: hue-rotate(0deg);
  }
  10% {
    opacity: 1;
    transform: translateY(-15px) scale(1);
  }
  50% {
    opacity: 1;
    transform: translateY(-80px) scale(1.3);
    filter: hue-rotate(30deg);
  }
  90% {
    opacity: 1;
    transform: translateY(-160px) scale(1);
  }
  100% {
    opacity: 0;
    transform: translateY(-200px) scale(0);
    filter: hue-rotate(60deg);
  }
}

.corner-decor {
  position: absolute;
  width: 60px;
  height: 60px;
  border-color: rgba(var(--primary-color), 0.4);
  border-style: solid;

  &.tl {
    top: 30px;
    left: 30px;
    border-width: 2px 0 0 2px;
  }

  &.tr {
    top: 30px;
    right: 30px;
    border-width: 2px 2px 0 0;
  }

  &.bl {
    bottom: 30px;
    left: 30px;
    border-width: 0 0 2px 2px;
  }

  &.br {
    bottom: 30px;
    right: 30px;
    border-width: 0 2px 2px 0;
  }
}

.login-box {
  width: 420px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-xl);
  padding: 40px;
  backdrop-filter: blur(10px);
  position: relative;
  z-index: 1;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--primary-color), transparent);
  }
}

.login-header {
  text-align: center;
  margin-bottom: 40px;

  .logo-icon {
    width: 48px;
    height: 48px;
    margin: 0 auto 16px;
    color: var(--primary-color);
    animation: clock-pulse 2s ease-in-out infinite;

    svg {
      width: 100%;
      height: 100%;
    }
  }

  .title {
    font-size: 32px;
    font-weight: 700;
    letter-spacing: 8px;
    background: linear-gradient(135deg, var(--primary-color), var(--info-color));
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    margin-bottom: 8px;
  }

  .subtitle {
    color: var(--text-muted);
    font-size: 14px;
    letter-spacing: 2px;
  }
}

@keyframes clock-pulse {
  0%, 100% {
    opacity: 1;
    filter: drop-shadow(0 0 5px var(--primary-glow));
  }
  50% {
    opacity: 0.8;
    filter: drop-shadow(0 0 20px var(--primary-color));
  }
}

.login-form {
  width: 100%;
  
  :deep(.el-form) {
    width: 100%;
  }
  
  :deep(.el-form-item) {
    width: 100%;
    margin-bottom: 24px;
  }
  
  :deep(.el-form-item__content) {
    width: 100%;
  }
}

.input-wrapper {
  position: relative;
  width: 100%;
  transition: all 0.3s ease;
  
  :deep(.el-input) {
    width: 100%;
  }

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
    transition: height 0.3s ease;
  }

  &.focused::before {
    height: 24px;
  }

  .input-icon {
    position: absolute;
    left: 16px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-muted);
    z-index: 1;
    transition: color 0.3s ease;
  }

  &.focused .input-icon {
    color: var(--primary-color);
  }

  :deep(.el-input__wrapper) {
    padding-left: 48px !important;
    background: var(--input-bg) !important;
    border: 1px solid var(--input-border) !important;
    border-radius: var(--border-radius-base) !important;
    box-shadow: none !important;
    transition: all 0.3s ease;

    &:hover {
      border-color: rgba(var(--primary-color), 0.5) !important;
    }

    &.is-focus {
      border-color: var(--primary-color) !important;
      box-shadow: 0 0 0 2px rgba(var(--primary-color), 0.1) !important;
    }
  }

  :deep(.el-input__inner) {
    color: var(--text-primary) !important;
    height: 48px;
    font-size: 15px;

    &::placeholder {
      color: var(--input-placeholder) !important;
    }
  }
}

.login-btn {
  width: 100%;
  height: 48px;
  position: relative;
  background: transparent;
  border: 1px solid var(--primary-color);
  border-radius: var(--border-radius-base);
  color: var(--primary-color);
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 4px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.3s ease;

  .btn-text {
    position: relative;
    z-index: 1;
  }

  .btn-glow {
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(var(--primary-color), 0.2),
      transparent
    );
    transition: left 0.5s ease;
  }

  &:hover {
    background: rgba(var(--primary-color), 0.1);
    box-shadow: 0 0 30px rgba(var(--primary-color), 0.2);
    border-color: var(--primary-color);

    .btn-glow {
      left: 100%;
    }
  }

  &:active {
    transform: scale(0.98);
  }

  &.loading {
    opacity: 0.7;
    cursor: not-allowed;
  }
}

.login-footer {
  margin-top: 32px;
  text-align: center;

  .system-info {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: rgba(var(--success-color), 0.1);
    border: 1px solid rgba(var(--success-color), 0.2);
    border-radius: 20px;
    font-size: 12px;
    color: var(--success-color);

    .info-dot {
      width: 6px;
      height: 6px;
      background: var(--success-color);
      border-radius: 50%;
      animation: status-pulse 2s ease-in-out infinite;
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
</style>
