<template>
  <div class="status-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>实时任务状态</span>
          <div class="header-actions">
            <el-switch v-model="autoScroll" active-text="自动滚动" />
            <el-button type="primary" @click="connect" :loading="connecting">
              {{ connecting ? '连接中...' : '重新连接' }}
            </el-button>
          </div>
        </div>
      </template>

      <div class="log-container" ref="logContainerRef">
        <div v-for="(log, index) in logs" :key="index" class="log-item">
          <span class="timestamp">{{ log.timestamp }}</span>
          <span class="message" :class="log.type">{{ log.message }}</span>
        </div>
      </div>
    </el-card>
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
    addLog('info', 'WebSocket 连接已建立')
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
    addLog('error', 'WebSocket 连接已断开，3秒后自动重连...')
    reconnectTimer = setTimeout(() => connect(), 3000)
  }

  ws.onerror = () => {
    connecting.value = false
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
  const timestamp = new Date().toLocaleTimeString()
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

onMounted(() => {
  connect()
})

onUnmounted(() => {
  disconnect()
})
</script>

<style lang="scss" scoped>
.status-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .header-actions {
      display: flex;
      align-items: center;
      gap: 15px;
    }
  }

  .log-container {
    height: calc(100vh - 200px);
    overflow-y: auto;
    background: #1e1e1e;
    padding: 15px;
    border-radius: 4px;
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 13px;
  }

  .log-item {
    margin-bottom: 5px;
    line-height: 1.6;

    .timestamp {
      color: #666;
      margin-right: 10px;
    }

    .message {
      &.info { color: #409EFF; }
      &.success { color: #67C23A; }
      &.error { color: #F56C6C; }
      &.stdout { color: #fff; }
      &.stderr { color: #F56C6C; }
    }
  }
}
</style>
