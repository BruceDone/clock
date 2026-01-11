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

    <!-- 运行中任务面板 -->
    <el-card class="running-tasks-card" v-if="runningTasks.length > 0">
      <template #header>
        <div class="card-header">
          <span class="card-title">
            <el-icon class="running-icon"><Loading /></el-icon>
            运行中任务 ({{ runningTasks.length }})
          </span>
          <el-button v-if="groupedByRunId.size > 0" type="danger" size="small" plain @click="handleCancelAllRuns">
            取消全部
          </el-button>
        </div>
      </template>
      <div class="running-tasks-list">
        <!-- 按 runId 分组显示 -->
        <div v-for="[runId, tasks] in groupedByRunId" :key="runId || 'single'" class="run-group">
          <div class="run-header" v-if="runId">
            <span class="run-id">RunID: {{ runId }}</span>
            <el-button type="danger" size="small" link @click="handleCancelRun(runId)">
              <el-icon><CircleClose /></el-icon>
              取消整个 Run
            </el-button>
          </div>
          <div class="task-items">
            <div v-for="task in tasks" :key="task.tid" class="task-item">
              <div class="task-info">
                <span class="task-name">{{ task.taskName }}</span>
                <span class="task-meta">TID: {{ task.tid }} | CID: {{ task.cid }}</span>
                <span class="task-duration">{{ formatDuration(task.startAt) }}</span>
              </div>
              <el-button type="danger" size="small" @click="handleCancelTask(task.tid)">
                <el-icon><CircleClose /></el-icon>
                取消
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </el-card>

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
        <div v-if="logGroups.length === 0" class="log-empty">
          <el-icon size="48"><ChatDotRound /></el-icon>
          <p>等待任务日志...</p>
        </div>
        <!-- 按 runId 分组展示日志 -->
        <div v-for="group in logGroups" :key="group.runId || group.id" class="log-group">
          <!-- 系统消息（无 runId） -->
          <template v-if="!group.runId">
            <div class="log-item system-log">
              <span class="timestamp">[{{ group.timestamp }}]</span>
              <span class="message" :class="group.type">{{ group.message }}</span>
            </div>
          </template>
          <!-- 任务执行组 -->
          <template v-else>
            <div class="run-group-header" @click="toggleGroup(group.runId)">
              <span class="run-indicator" :class="{ collapsed: collapsedGroups.has(group.runId) }">
                <el-icon><ArrowRight /></el-icon>
              </span>
              <span class="run-time">[{{ group.timestamp }}]</span>
              <span class="run-badge">{{ group.runId }}</span>
              <span class="run-summary">
                <template v-if="group.tasks.length > 0">
                  {{ group.tasks.map(t => t.taskName).join(' -> ') }}
                </template>
              </span>
              <span class="run-status" :class="getGroupStatus(group)">
                {{ getGroupStatusText(group) }}
              </span>
            </div>
            <div class="run-group-content" v-show="!collapsedGroups.has(group.runId)">
              <div v-for="task in group.tasks" :key="task.tid" class="task-block">
                <div class="task-header">
                  <span class="task-icon" :class="task.status">
                    <el-icon v-if="task.status === 'running'"><Loading /></el-icon>
                    <el-icon v-else-if="task.status === 'success'"><CircleCheck /></el-icon>
                    <el-icon v-else-if="task.status === 'error'"><CircleClose /></el-icon>
                    <el-icon v-else-if="task.status === 'cancelled'"><RemoveFilled /></el-icon>
                    <el-icon v-else><Clock /></el-icon>
                  </span>
                  <span class="task-name">{{ task.taskName }}</span>
                  <span class="task-id">TID:{{ task.tid }}</span>
                  <span v-if="task.duration" class="task-duration">{{ task.duration }}ms</span>
                </div>
                <div class="task-logs" v-if="task.logs.length > 0">
                  <div v-for="log in task.logs" :key="log.id" class="task-log-line" :class="log.type">
                    {{ log.message }}
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRunningTasks, cancelTask, cancelRun, type RunningTaskInfo } from '@/api/task'

interface StreamEvent {
  id: number
  ts: number
  kind: 'task_start' | 'task_end' | 'stdout' | 'stderr' | 'meta' | string
  runId?: string
  tid?: number
  cid?: number
  taskName?: string
  status?: string
  durationMs?: number
  msg?: string
}

interface TaskLog {
  id: number
  type: 'stdout' | 'stderr' | 'info'
  message: string
}

interface TaskBlock {
  tid: number
  taskName: string
  status: 'pending' | 'running' | 'success' | 'error' | 'cancelled'
  duration?: number
  logs: TaskLog[]
}

interface LogGroup {
  id: number
  runId?: string
  timestamp: string
  message?: string
  type?: 'info' | 'success' | 'error' | 'warning'
  tasks: TaskBlock[]
}

const MAX_GROUPS = 100

const logContainerRef = ref<HTMLElement>()
const logGroups = ref<LogGroup[]>([])
const collapsedGroups = reactive(new Set<string>())
const autoScroll = ref(true)
const connecting = ref(false)
const isConnected = ref(false)
const runningTasks = ref<RunningTaskInfo[]>([])

let es: EventSource | null = null
let localLogID = -1
let lastErrorAt = 0
let refreshTimer: number | null = null

// runId -> LogGroup 映射，用于快速查找
const runIdToGroup = new Map<string, LogGroup>()

// 按 runId 分组
const groupedByRunId = computed(() => {
  const map = new Map<string, RunningTaskInfo[]>()
  for (const task of runningTasks.value) {
    const key = task.runId || ''
    if (!map.has(key)) {
      map.set(key, [])
    }
    map.get(key)!.push(task)
  }
  return map
})

function formatTs(ms: number) {
  return new Date(ms).toLocaleTimeString('zh-CN', { hour12: false })
}

function formatDuration(startAt: number) {
  const now = Date.now()
  const duration = Math.floor((now - startAt) / 1000)
  if (duration < 60) return `${duration}s`
  const minutes = Math.floor(duration / 60)
  const seconds = duration % 60
  return `${minutes}m ${seconds}s`
}

function toggleGroup(runId: string) {
  if (collapsedGroups.has(runId)) {
    collapsedGroups.delete(runId)
  } else {
    collapsedGroups.add(runId)
  }
}

function getGroupStatus(group: LogGroup): string {
  if (group.tasks.length === 0) return 'pending'
  const hasError = group.tasks.some(t => t.status === 'error')
  const hasCancelled = group.tasks.some(t => t.status === 'cancelled')
  const allSuccess = group.tasks.every(t => t.status === 'success')
  const hasRunning = group.tasks.some(t => t.status === 'running')
  
  if (hasRunning) return 'running'
  if (hasError) return 'error'
  if (hasCancelled) return 'cancelled'
  if (allSuccess) return 'success'
  return 'pending'
}

function getGroupStatusText(group: LogGroup): string {
  const status = getGroupStatus(group)
  const total = group.tasks.length
  const completed = group.tasks.filter(t => t.status === 'success').length
  const failed = group.tasks.filter(t => t.status === 'error').length
  
  switch (status) {
    case 'running': return `运行中 (${completed}/${total})`
    case 'success': return `完成 (${total})`
    case 'error': return `失败 (${failed}/${total})`
    case 'cancelled': return '已取消'
    default: return '等待中'
  }
}

async function scrollToBottom() {
  await nextTick()
  if (logContainerRef.value && autoScroll.value) {
    logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
  }
}

function addSystemLog(type: LogGroup['type'], message: string, ts?: number) {
  const group: LogGroup = {
    id: localLogID--,
    timestamp: formatTs(ts ?? Date.now()),
    message,
    type,
    tasks: []
  }
  logGroups.value.push(group)
  
  // 限制数量
  if (logGroups.value.length > MAX_GROUPS) {
    const removed = logGroups.value.shift()
    if (removed?.runId) {
      runIdToGroup.delete(removed.runId)
    }
  }
  
  scrollToBottom()
}

function getOrCreateRunGroup(runId: string, ts: number): LogGroup {
  let group = runIdToGroup.get(runId)
  if (!group) {
    group = {
      id: localLogID--,
      runId,
      timestamp: formatTs(ts),
      tasks: []
    }
    runIdToGroup.set(runId, group)
    logGroups.value.push(group)
    
    // 限制数量
    if (logGroups.value.length > MAX_GROUPS) {
      const removed = logGroups.value.shift()
      if (removed?.runId) {
        runIdToGroup.delete(removed.runId)
      }
    }
  }
  return group
}

function getOrCreateTask(group: LogGroup, tid: number, taskName: string): TaskBlock {
  let task = group.tasks.find(t => t.tid === tid)
  if (!task) {
    task = {
      tid,
      taskName,
      status: 'pending',
      logs: []
    }
    group.tasks.push(task)
  }
  return task
}

function onStreamEvent(ev: StreamEvent) {
  // 没有 runId 的事件作为系统消息处理
  if (!ev.runId) {
    const msg = ev.msg || ev.kind
    addSystemLog('info', msg, ev.ts)
    return
  }

  const group = getOrCreateRunGroup(ev.runId, ev.ts)
  
  switch (ev.kind) {
    case 'task_start': {
      const task = getOrCreateTask(group, ev.tid!, ev.taskName || `Task ${ev.tid}`)
      task.status = 'running'
      fetchRunningTasks()
      break
    }
    case 'task_end': {
      const task = getOrCreateTask(group, ev.tid!, ev.taskName || `Task ${ev.tid}`)
      task.duration = ev.durationMs
      if (ev.status === 'cancelled') {
        task.status = 'cancelled'
      } else if (ev.status === 'success') {
        task.status = 'success'
      } else {
        task.status = 'error'
        if (ev.msg) {
          task.logs.push({
            id: localLogID--,
            type: 'stderr',
            message: ev.msg
          })
        }
      }
      fetchRunningTasks()
      break
    }
    case 'stdout': {
      const task = getOrCreateTask(group, ev.tid!, ev.taskName || `Task ${ev.tid}`)
      if (ev.msg) {
        task.logs.push({
          id: ev.id,
          type: 'stdout',
          message: ev.msg
        })
      }
      break
    }
    case 'stderr': {
      const task = getOrCreateTask(group, ev.tid!, ev.taskName || `Task ${ev.tid}`)
      if (ev.msg) {
        task.logs.push({
          id: ev.id,
          type: 'stderr',
          message: ev.msg
        })
      }
      break
    }
    case 'meta': {
      const task = getOrCreateTask(group, ev.tid!, ev.taskName || `Task ${ev.tid}`)
      if (ev.msg) {
        task.logs.push({
          id: ev.id,
          type: 'info',
          message: ev.msg
        })
      }
      break
    }
  }
  
  scrollToBottom()
}

async function fetchRunningTasks() {
  try {
    const res = await getRunningTasks()
    runningTasks.value = res.data || []
  } catch (error) {
    console.error('Failed to fetch running tasks:', error)
  }
}

async function handleCancelTask(tid: number) {
  try {
    await ElMessageBox.confirm('确定要取消该任务吗？', '确认取消', { type: 'warning' })
    await cancelTask(tid)
    ElMessage.success('任务取消请求已发送')
    fetchRunningTasks()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to cancel task:', error)
    }
  }
}

async function handleCancelRun(runId: string) {
  try {
    await ElMessageBox.confirm('确定要取消该 Run 的所有任务吗？', '确认取消', { type: 'warning' })
    await cancelRun(runId)
    ElMessage.success('Run 取消请求已发送')
    fetchRunningTasks()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to cancel run:', error)
    }
  }
}

async function handleCancelAllRuns() {
  try {
    await ElMessageBox.confirm('确定要取消所有运行中的任务吗？', '确认取消', { type: 'warning' })
    const runIds = new Set<string>()
    const singleTasks: number[] = []
    for (const task of runningTasks.value) {
      if (task.runId) {
        runIds.add(task.runId)
      } else {
        singleTasks.push(task.tid)
      }
    }
    // 取消所有 run
    for (const runId of runIds) {
      await cancelRun(runId)
    }
    // 取消单独的任务
    for (const tid of singleTasks) {
      await cancelTask(tid)
    }
    ElMessage.success('所有取消请求已发送')
    fetchRunningTasks()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to cancel all:', error)
    }
  }
}

function connect() {
  if (connecting.value) return

  disconnect()
  connecting.value = true

  const token = localStorage.getItem('token')
  const url = token ? `/v1/task/status?token=${encodeURIComponent(token)}` : '/v1/task/status'
  es = new EventSource(url, { withCredentials: true })

  es.onopen = () => {
    connecting.value = false
    isConnected.value = true
    addSystemLog('success', 'SSE 连接已建立')
  }

  es.onerror = () => {
    connecting.value = false
    isConnected.value = false

    const now = Date.now()
    if (now - lastErrorAt > 3000) {
      addSystemLog('error', 'SSE 连接异常（自动重连中...）')
      lastErrorAt = now
    }
  }

  es.addEventListener('log', (event) => {
    try {
      const msgEvent = event as MessageEvent
      const data = JSON.parse(msgEvent.data) as StreamEvent
      onStreamEvent(data)
    } catch {
      addSystemLog('info', String((event as MessageEvent).data ?? ''))
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
  logGroups.value = []
  runIdToGroup.clear()
  collapsedGroups.clear()
}

onMounted(() => {
  connect()
  fetchRunningTasks()
  refreshTimer = window.setInterval(fetchRunningTasks, 5000)
})

onUnmounted(() => {
  disconnect()
  if (refreshTimer != null) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<style lang="scss" scoped>
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

  .running-tasks-card {
    background: var(--bg-card) !important;
    border: 1px solid var(--border-color) !important;
    border-radius: var(--border-radius-lg) !important;
    margin-bottom: 24px;
    backdrop-filter: blur(10px);

    :deep(.el-card__header) {
      padding: 12px 20px;
      border-bottom: 1px solid var(--border-color);
    }

    :deep(.el-card__body) {
      padding: 16px 20px;
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .card-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-weight: 600;
        color: var(--text-primary);

        .running-icon {
          color: var(--warning-color);
          animation: spin 1s linear infinite;
        }
      }
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    .running-tasks-list {
      .run-group {
        &:not(:last-child) {
          margin-bottom: 16px;
          padding-bottom: 16px;
          border-bottom: 1px dashed var(--border-color);
        }

        .run-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 12px;

          .run-id {
            font-family: var(--font-family-mono);
            font-size: 13px;
            color: var(--primary-color);
            background: rgba(var(--primary-color), 0.1);
            padding: 4px 8px;
            border-radius: 4px;
          }
        }

        .task-items {
          display: flex;
          flex-direction: column;
          gap: 8px;

          .task-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 16px;
            background: var(--bg-secondary);
            border: 1px solid var(--border-color-light);
            border-radius: var(--border-radius);
            transition: all 0.2s ease;

            &:hover {
              border-color: var(--primary-color);
              background: rgba(var(--primary-color), 0.05);
            }

            .task-info {
              display: flex;
              flex-direction: column;
              gap: 4px;

              .task-name {
                font-weight: 600;
                color: var(--text-primary);
              }

              .task-meta {
                font-family: var(--font-family-mono);
                font-size: 12px;
                color: var(--text-muted);
              }

              .task-duration {
                font-family: var(--font-family-mono);
                font-size: 12px;
                color: var(--warning-color);
              }
            }
          }
        }
      }
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

      .log-group {
        margin-bottom: 8px;

        // 系统日志样式
        .system-log {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 6px 12px;
          border-radius: 4px;
          background: rgba(var(--info-color), 0.05);

          .timestamp {
            color: var(--text-muted);
            font-size: 12px;
          }

          .message {
            &.info { color: var(--info-color); }
            &.success { color: var(--success-color); }
            &.error { color: var(--danger-color); }
            &.warning { color: var(--warning-color); }
          }
        }

        // Run 组头部
        .run-group-header {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 10px 12px;
          background: var(--bg-secondary);
          border: 1px solid var(--border-color);
          border-radius: 6px;
          cursor: pointer;
          transition: all 0.2s ease;

          &:hover {
            background: rgba(var(--primary-color), 0.05);
            border-color: var(--primary-color);
          }

          .run-indicator {
            color: var(--text-muted);
            transition: transform 0.2s ease;
            display: flex;
            align-items: center;

            &.collapsed {
              transform: rotate(0deg);
            }

            &:not(.collapsed) {
              transform: rotate(90deg);
            }
          }

          .run-time {
            color: var(--text-muted);
            font-size: 12px;
          }

          .run-badge {
            background: var(--primary-color);
            color: #fff;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 11px;
            font-weight: 600;
          }

          .run-summary {
            flex: 1;
            color: var(--text-secondary);
            font-size: 12px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          .run-status {
            padding: 2px 10px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 500;

            &.running {
              background: rgba(var(--warning-color), 0.15);
              color: var(--warning-color);
            }
            &.success {
              background: rgba(var(--success-color), 0.15);
              color: var(--success-color);
            }
            &.error {
              background: rgba(var(--danger-color), 0.15);
              color: var(--danger-color);
            }
            &.cancelled {
              background: rgba(var(--text-muted), 0.15);
              color: var(--text-muted);
            }
            &.pending {
              background: rgba(var(--info-color), 0.15);
              color: var(--info-color);
            }
          }
        }

        // Run 组内容
        .run-group-content {
          margin-left: 20px;
          padding-left: 16px;
          border-left: 2px solid var(--border-color);
          margin-top: 8px;

          .task-block {
            margin-bottom: 12px;
            padding: 10px 12px;
            background: var(--bg-primary);
            border-radius: 6px;
            border: 1px solid var(--border-color-light);

            &:last-child {
              margin-bottom: 0;
            }

            .task-header {
              display: flex;
              align-items: center;
              gap: 10px;
              margin-bottom: 6px;

              .task-icon {
                display: flex;
                align-items: center;
                justify-content: center;
                width: 20px;
                height: 20px;
                border-radius: 50%;

                &.running {
                  color: var(--warning-color);
                  animation: spin 1s linear infinite;
                }
                &.success {
                  color: var(--success-color);
                }
                &.error {
                  color: var(--danger-color);
                }
                &.cancelled {
                  color: var(--text-muted);
                }
                &.pending {
                  color: var(--info-color);
                }
              }

              .task-name {
                font-weight: 600;
                color: var(--text-primary);
              }

              .task-id {
                font-size: 11px;
                color: var(--text-muted);
                background: var(--bg-secondary);
                padding: 2px 6px;
                border-radius: 4px;
              }

              .task-duration {
                margin-left: auto;
                font-size: 11px;
                color: var(--text-secondary);
                background: var(--bg-secondary);
                padding: 2px 8px;
                border-radius: 4px;
              }
            }

            .task-logs {
              padding-left: 30px;

              .task-log-line {
                padding: 2px 0;
                font-size: 12px;
                line-height: 1.5;
                word-break: break-all;

                &.stdout {
                  color: var(--text-primary);
                }
                &.stderr {
                  color: var(--warning-color);
                }
                &.info {
                  color: var(--text-secondary);
                }
              }
            }
          }
        }
      }
    }
  }
}
</style>
