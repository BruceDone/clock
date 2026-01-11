<template>
  <div class="status-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">实时状态</h1>
        <p class="page-subtitle">WebSocket 实时任务执行监控</p>
      </div>
      <div class="header-right">
        <div class="connection-status" :class="{ connected: isConnected }">
          <span class="status-dot"></span>
          <span class="status-text">{{ isConnected ? '已连接' : '连接中...' }}</span>
        </div>
      </div>
    </div>

    <!-- 操作栏 -->
    <el-card class="action-card">
      <div class="action-bar">
        <div class="action-left">
          <el-switch v-model="autoScroll" active-text="自动滚动" />
        </div>
        <div class="action-right">
          <el-button type="primary" @click="connect" :loading="connecting">
            <el-icon><Connection /></el-icon>
            {{ connecting ? '连接中...' : '重新连接' }}
          </el-button>
          <el-button @click="clearLogs">
            <el-icon><Delete /></el-icon>
            清空日志
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 日志容器 -->
    <div class="log-container scanline" ref="logContainerRef">
      <div class="log-bg-grid"></div>
      <div class="log-content">
        <div v-if="logs.length === 0" class="log-empty">
          <el-icon size="48"><ChatDotRound /></el-icon>
          <p>等待任务日志...</p>
        </div>
        <div v-for="(log, index) in logs" :key="index" class="log-item" :class="log.type">
          <span class="timestamp">[{{ log.timestamp }}]</span>
          <span class="message">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

interface LogMessage {
  timestamp: string
  message: string
  type: 'info' | 'success' | 'error' | 'stdout' | 'stderr'
}

const logContainerRef = ref<HTMLElement>()
const logs = ref<LogMessage[]>([])
const autoScroll = ref(true)
const connecting = ref(false)
const isConnected = ref(false)

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

function getWsUrl() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.hostname === 'localhost' ? '127.0.0.1:9528' : window.location.host
  return `${protocol}//${host}/v1/task/status`
}

function connect() {
  if (connecting.value) return

  disconnect()
  connecting.value = true

  const url = getWsUrl()
  ws = new WebSocket(url)

  ws.onopen = () => {
    connecting.value = false
    isConnected.value = true
    addLog('success', 'WebSocket 连接已建立')
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.tid) {
        addLog('stdout', `[TID:${data.tid}] ${data.msg || data.stdout || data.stderr}`)
      } else {
        addLog('info', data.msg || JSON.stringify(data))
      }
    } catch {
      addLog('info', event.data)
    }
  }

  ws.onclose = () => {
    connecting.value = false
    isConnected.value = false
    addLog('error', 'WebSocket 连接已断开，3秒后自动重连...')
    reconnectTimer = setTimeout(() => connect(), 3000)
  }

  ws.onerror = () => {
    connecting.value = false
    isConnected.value = false
    addLog('error', 'WebSocket 连接错误')
  }
}

function disconnect() {
  if (ws) {
    ws.close()
    ws = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

function addLog(type: LogMessage['type'], message: string) {
  const timestamp = new Date().toLocaleTimeString('zh-CN', { hour12: false })
  logs.value.push({ timestamp, message, type })

  // 限制日志数量
  if (logs.value.length > 1000) {
    logs.value.shift()
  }

  // 自动滚动
  if (autoScroll.value) {
    nextTick(() => {
      if (logContainerRef.value) {
        logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
      }
    })
  }
}

function clearLogs() {
  logs.value = []
}

onMounted(() => {
  connect()
})

onUnmounted(() => {
  disconnect()
})
</script>

<style lang="scss" scoped>
// Status 页面样式 - 使用 CSS 变量

.status-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;

    .page-title {
      font-size: 28px;
      font-weight: 700;
      color: var(--text-primary);
      margin-bottom: 8px;
      background: linear-gradient(135deg, var(--text-primary), var(--primary-color));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }

    .page-subtitle {
      color: var(--text-muted);
      font-size: 14px;
    }

    .connection-status {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 16px;
      background: rgba(var(--danger-color), 0.1);
      border: 1px solid rgba(var(--danger-color), 0.3);
      border-radius: 20px;

      .status-dot {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: var(--danger-color);
      }

      .status-text {
        font-size: 13px;
        color: var(--danger-color);
      }

      &.connected {
        background: rgba(var(--success-color), 0.1);
        border-color: rgba(var(--success-color), 0.3);

        .status-dot {
          background: var(--success-color);
          animation: status-pulse 2s ease-in-out infinite;
        }

        .status-text {
          color: var(--success-color);
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

  .action-card {
    background: var(--bg-card) !important;
    border: 1px solid var(--border-color) !important;
    border-radius: var(--border-radius-lg) !important;
    margin-bottom: 24px;
    backdrop-filter: blur(10px);

    :deep(.el-card__body) {
      padding: 16px 20px;
    }

    .action-bar {
      display: flex;
      justify-content: space-between;
      align-items: center;

      :deep(.el-switch) {
        --el-switch-off-color: var(--text-muted);
      }

      .action-right {
        display: flex;
        gap: 12px;
      }
    }
  }

  .log-container {
    position: relative;
    height: calc(100vh - 280px);
    min-height: 400px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius-lg);
    overflow: hidden;

    &.scanline::after {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: repeating-linear-gradient(
        0deg,
        transparent,
        transparent 2px,
        rgba(0, 0, 0, 0.1) 2px,
        rgba(0, 0, 0, 0.1) 4px
      );
      pointer-events: none;
      z-index: 10;
    }

    .log-bg-grid {
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-image:
        linear-gradient(rgba(var(--primary-color), 0.02) 1px, transparent 1px),
        linear-gradient(90deg, rgba(var(--primary-color), 0.02) 1px, transparent 1px);
      background-size: 20px 20px;
      pointer-events: none;
    }

    .log-content {
      position: relative;
      z-index: 1;
      height: 100%;
      overflow-y: auto;
      padding: 16px;
      font-family: var(--font-family-mono);
      font-size: 13px;
      scroll-behavior: smooth;

      &::-webkit-scrollbar {
        width: 8px;
      }

      &::-webkit-scrollbar-thumb {
        background: rgba(var(--primary-color), 0.3);
        border-radius: 4px;

        &:hover {
          background: rgba(var(--primary-color), 0.5);
        }
      }

      .log-empty {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        color: var(--text-muted);

        .el-icon {
          margin-bottom: 16px;
          opacity: 0.5;
        }
      }

      .log-item {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        margin-bottom: 4px;
        line-height: 1.6;
        padding: 4px 8px;
        border-radius: 4px;
        transition: background 0.2s ease;

        &:hover {
          background: rgba(var(--primary-color), 0.05);
        }

        .timestamp {
          color: var(--primary-color);
          font-weight: 500;
          flex-shrink: 0;
        }

        .message {
          flex: 1;
          word-break: break-all;

          &.info { color: var(--info-color); }
          &.success { color: var(--success-color); }
          &.error { color: var(--danger-color); }
          &.stdout { color: var(--text-primary); }
          &.stderr { color: var(--warning-color); }
        }
      }
    }
  }
}
</style>
