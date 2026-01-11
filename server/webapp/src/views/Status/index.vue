<template>
  <div class="status-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">实时状态</h1>
        <p class="page-subtitle">SSE 实时任务执行监控</p>
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
    <div class="log-container scanline">
      <div class="log-bg-grid"></div>
      <div class="log-content" ref="logContainerRef">
        <div v-if="logs.length === 0" class="log-empty">
          <el-icon size="48"><ChatDotRound /></el-icon>
          <p>等待任务日志...</p>
        </div>
        <div v-for="log in logs" :key="log.id" class="log-item">
          <span class="timestamp">[{{ log.timestamp }}]</span>
          <span class="message" :class="log.type">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'

interface StreamEvent {
  id: number
  ts: number
  kind: 'task_start' | 'task_end' | 'stdout' | 'stderr' | 'meta' | string
  tid?: number
  cid?: number
  taskName?: string
  status?: string
  durationMs?: number
  msg?: string
}

interface LogMessage {
  id: number
  timestamp: string
  message: string
  type: 'info' | 'success' | 'error' | 'stdout' | 'stderr'
}

const MAX_LOGS = 1000

const logContainerRef = ref<HTMLElement>()
const logs = ref<LogMessage[]>([])
const autoScroll = ref(true)
const connecting = ref(false)
const isConnected = ref(false)

let es: EventSource | null = null
let localLogID = -1
let lastErrorAt = 0

const pendingQueue: LogMessage[] = []
let flushHandle: number | null = null

function formatTs(ms: number) {
  return new Date(ms).toLocaleTimeString('zh-CN', { hour12: false })
}

function isNearBottom(el: HTMLElement, thresholdPx = 80) {
  return el.scrollHeight - el.scrollTop - el.clientHeight < thresholdPx
}

function scheduleFlush() {
  if (flushHandle != null) return

  flushHandle = window.requestAnimationFrame(async () => {
    flushHandle = null
    if (pendingQueue.length === 0) return

    const el = logContainerRef.value
    const shouldStick = !!el && autoScroll.value && isNearBottom(el)

    logs.value.push(...pendingQueue.splice(0))

    if (logs.value.length > MAX_LOGS) {
      logs.value.splice(0, logs.value.length - MAX_LOGS)
    }

    if (shouldStick) {
      await nextTick()
      if (logContainerRef.value) {
        logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
      }
    }

    if (pendingQueue.length > 0) {
      scheduleFlush()
    }
  })
}

function enqueueLog(type: LogMessage['type'], message: string, id?: number, ts?: number) {
  pendingQueue.push({
    id: id ?? localLogID--,
    timestamp: formatTs(ts ?? Date.now()),
    message,
    type
  })
  scheduleFlush()
}

function formatPrefix(ev: StreamEvent) {
  const parts: string[] = []
  if (ev.taskName) parts.push(ev.taskName)
  if (ev.tid != null) parts.push(`TID:${ev.tid}`)
  if (ev.cid != null) parts.push(`CID:${ev.cid}`)
  return parts.length ? `[${parts.join(' ')}] ` : ''
}

function onStreamEvent(ev: StreamEvent) {
  const prefix = formatPrefix(ev)

  switch (ev.kind) {
    case 'stdout':
      enqueueLog('stdout', `${prefix}${ev.msg ?? ''}`, ev.id, ev.ts)
      return
    case 'stderr':
      enqueueLog('stderr', `${prefix}${ev.msg ?? ''}`, ev.id, ev.ts)
      return
    case 'task_start':
      enqueueLog('info', `${prefix}开始执行`, ev.id, ev.ts)
      return
    case 'task_end': {
      const ok = ev.status === 'success'
      const extra = ev.durationMs != null ? ` (${ev.durationMs}ms)` : ''
      const msg = ev.msg ? `: ${ev.msg}` : ''
      enqueueLog(ok ? 'success' : 'error', `${prefix}${ev.status ?? 'done'}${extra}${msg}`, ev.id, ev.ts)
      return
    }
    case 'meta':
      enqueueLog('info', `${prefix}${ev.msg ?? ''}`, ev.id, ev.ts)
      return
    default:
      enqueueLog('info', `${prefix}${ev.kind}: ${ev.msg ?? ''}`, ev.id, ev.ts)
  }
}

function connect() {
  if (connecting.value) return

  disconnect()
  connecting.value = true

  // Note: /v1 is proxied in dev (see vite.config.ts).
  // EventSource can't set custom headers, so we use cookie auth primarily.
  // Fallback: also send token via query (supports existing localStorage token).
  const token = localStorage.getItem('token')
  const url = token ? `/v1/task/status?token=${encodeURIComponent(token)}` : '/v1/task/status'
  es = new EventSource(url, { withCredentials: true })

  es.onopen = () => {
    connecting.value = false
    isConnected.value = true
    enqueueLog('success', 'SSE 连接已建立')
  }

  es.onerror = () => {
    connecting.value = false
    isConnected.value = false

    const now = Date.now()
    if (now - lastErrorAt > 3000) {
      // EventSource will auto-reconnect; keep this message lightweight.
      enqueueLog('error', 'SSE 连接异常（自动重连中...）')
      lastErrorAt = now
    }
  }

  es.addEventListener('log', (event) => {
    try {
      const msgEvent = event as MessageEvent
      const data = JSON.parse(msgEvent.data) as StreamEvent
      onStreamEvent(data)
    } catch {
      enqueueLog('info', String((event as MessageEvent).data ?? ''))
    }
  })
}

function disconnect() {
  if (es) {
    es.close()
    es = null
  }
  connecting.value = false
  isConnected.value = false
}

function clearLogs() {
  logs.value = []
  pendingQueue.splice(0)
}

onMounted(() => {
  connect()
})

onUnmounted(() => {
  disconnect()
  if (flushHandle != null) {
    window.cancelAnimationFrame(flushHandle)
    flushHandle = null
  }
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
      scroll-behavior: auto;

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
