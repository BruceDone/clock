<template>
  <div class="container-config">
    <el-page-header @back="router.back()" style="margin-bottom: 20px">
      <template #content>
        <div class="page-header-content">
          <span class="page-title">容器配置 - {{ container?.name }}</span>
          <el-button
            v-if="runningTasksForContainer.length > 0"
            type="danger"
            size="small"
            @click="handleCancelContainerRun"
          >
            <el-icon><CircleClose /></el-icon>
            取消运行 ({{ runningTasksForContainer.length }})
          </el-button>
        </div>
      </template>
    </el-page-header>

    <el-row :gutter="20">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>任务列表</span>
              <el-button type="primary" size="small" @click="showAddTaskDialog">新增任务</el-button>
            </div>
          </template>
          <div class="task-list">
            <div
              v-for="task in tasks"
              :key="task.tid"
              class="task-item"
              :class="{ active: selectedTask?.tid === task.tid }"
              @click="selectTask(task)"
            >
              <div class="task-name">{{ task.name }}</div>
              <div class="task-command">{{ task.command }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>DAG 关系图</span>
              <div class="graph-actions">
                <span class="graph-tip">Shift+拖拽连线 | 双击编辑 | 右键删除</span>
                <el-button-group>
                  <el-button size="small" @click="handleFitView">适应视图</el-button>
                  <el-button size="small" @click="savePositions">保存位置</el-button>
                </el-button-group>
              </div>
            </div>
          </template>
          <div class="graph-container">
            <DagCanvas
              ref="dagCanvasRef"
              :nodes="dagNodes"
              :edges="dagEdges"
              @node-click="onNodeClick"
              @node-move="onNodeMove"
              @edge-create="onEdgeCreate"
              @edge-delete="onEdgeDelete"
              @node-delete="onNodeDelete"
            />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="showTaskDialog" :title="isEditTask ? '编辑任务' : '新增任务'" width="600px">
      <el-form ref="taskFormRef" :model="taskForm" :rules="taskRules" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="taskForm.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="Bash 命令" prop="command">
          <el-input v-model="taskForm.command" type="textarea" :rows="3" placeholder="请输入要执行的 bash 命令" />
        </el-form-item>
        <el-form-item label="工作目录" prop="directory">
          <el-input v-model="taskForm.directory" placeholder="命令执行的工作目录" />
        </el-form-item>
        <el-form-item label="超时时间" prop="timeout">
          <el-input-number v-model="taskForm.timeout" :min="0" placeholder="超时时间（秒）" />
        </el-form-item>
        <el-form-item label="日志启用">
          <el-switch v-model="taskForm.log_enable" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTaskDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitTask">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getContainer } from '@/api/container'
import { getTasks, putTask, deleteTask, getRunningTasks, cancelRun, type RunningTaskInfo } from '@/api/task'
import { getRelations, addRelation, deleteRelation } from '@/api/relation'
import { putNodes } from '@/api/node'
import type { Container, Task, RelationResponse, Link } from '@/types/model'
import DagCanvas from '@/components/DagCanvas/index.vue'

const route = useRoute()
const router = useRouter()
const cid = parseInt(route.params.cid as string)

const container = ref<Container | null>(null)
const tasks = ref<Task[]>([])
const relations = ref<RelationResponse | null>(null)
const selectedTask = ref<Task | null>(null)
const runningTasks = ref<RunningTaskInfo[]>([])
let refreshTimer: number | null = null

const dagCanvasRef = ref<InstanceType<typeof DagCanvas>>()

const showTaskDialog = ref(false)
const isEditTask = ref(false)
const editingTaskId = ref<number>()

const taskFormRef = ref<FormInstance>()
const taskForm = reactive({
  name: '',
  command: '',
  directory: '',
  timeout: 30,
  log_enable: true
})

const taskRules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  command: [{ required: true, message: '请输入命令', trigger: 'blur' }]
}

// 该容器运行中的任务
const runningTasksForContainer = computed(() => {
  return runningTasks.value.filter(t => t.cid === cid)
})

// 转换为 DAG Canvas 需要的节点格式
const dagNodes = computed(() => {
  // 优先使用 relations 中的节点位置
  const nodeMap = new Map<number, { x: number; y: number }>()
  relations.value?.nodes?.forEach(n => {
    nodeMap.set(n.id, { x: n.x, y: n.y })
  })

  return tasks.value.map((task, index) => {
    const pos = nodeMap.get(task.tid)
    return {
      id: task.tid,
      name: task.name,
      status: task.status,
      x: pos?.x && pos.x > 0 ? pos.x : 100 + index * 150,
      y: pos?.y && pos.y > 0 ? pos.y : 250
    }
  })
})

// 转换为 DAG Canvas 需要的边格式
const dagEdges = computed(() => {
  return (relations.value?.links || []).map((link: Link) => ({
    id: link.id,
    source: link.tid,
    target: link.next_tid
  }))
})

function showAddTaskDialog() {
  isEditTask.value = false
  editingTaskId.value = undefined
  taskForm.name = ''
  taskForm.command = ''
  taskForm.directory = ''
  taskForm.timeout = 30
  taskForm.log_enable = true
  showTaskDialog.value = true
}

function selectTask(task: Task) {
  selectedTask.value = task
  isEditTask.value = true
  editingTaskId.value = task.tid
  taskForm.name = task.name
  taskForm.command = task.command
  taskForm.directory = task.directory
  taskForm.timeout = task.timeout
  taskForm.log_enable = task.log_enable
  showTaskDialog.value = true
}

function onNodeClick(node: { id: number }) {
  const task = tasks.value.find(t => t.tid === node.id)
  if (task) {
    selectTask(task)
  }
}

function onNodeMove(node: { id: number; x: number; y: number }) {
  // 更新本地节点位置
  const dagNode = dagNodes.value.find(n => n.id === node.id)
  if (dagNode) {
    dagNode.x = node.x
    dagNode.y = node.y
  }
}

async function onEdgeCreate(sourceId: number, targetId: number) {
  if (sourceId === targetId) {
    ElMessage.warning('不能连接自己')
    return
  }

  // 检查是否已存在连线
  const exists = dagEdges.value.some(e => e.source === sourceId && e.target === targetId)
  if (exists) {
    ElMessage.warning('连线已存在')
    return
  }

  try {
    await addRelation({
      cid,
      tid: sourceId,
      next_tid: targetId
    })
    ElMessage.success('连线创建成功')
    fetchData()
  } catch (error) {
    console.error(error)
  }
}

async function onEdgeDelete(edgeId: number) {
  try {
    await ElMessageBox.confirm('确定删除这条连线吗？', '确认删除', {
      type: 'warning'
    })
    await deleteRelation(edgeId)
    ElMessage.success('连线删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

async function onNodeDelete(nodeId: number) {
  try {
    await ElMessageBox.confirm('确定删除这个任务吗？相关的连线也会被删除。', '确认删除', {
      type: 'warning'
    })
    await deleteTask(nodeId)
    ElMessage.success('任务删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

function handleFitView() {
  dagCanvasRef.value?.fitView()
}

async function fetchData() {
  try {
    const [containerRes, tasksRes, relationsRes] = await Promise.all([
      getContainer(cid),
      getTasks({ cid }),
      getRelations({ cid })
    ])

    container.value = (containerRes as any).data
    tasks.value = (tasksRes as any).data?.items || []
    relations.value = (relationsRes as any).data
  } catch (error) {
    console.error(error)
  }
}

async function handleSubmitTask() {
  const valid = await taskFormRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    await putTask({ tid: editingTaskId.value, cid, ...taskForm })
    ElMessage.success('保存成功')
    showTaskDialog.value = false
    fetchData()
  } catch (error) {
    console.error(error)
  }
}

async function savePositions() {
  const nodeData = dagNodes.value.map(node => ({
    id: node.id,
    name: node.name,
    status: node.status,
    x: Math.round(node.x),
    y: Math.round(node.y)
  }))

  try {
    await putNodes(nodeData)
    ElMessage.success('位置保存成功')
  } catch (error) {
    console.error(error)
  }
}

async function fetchRunningTasks() {
  try {
    const res = await getRunningTasks()
    runningTasks.value = res.data || []
  } catch (error) {
    console.error('Failed to fetch running tasks:', error)
  }
}

async function handleCancelContainerRun() {
  try {
    await ElMessageBox.confirm('确定要取消该容器的所有运行中任务吗？', '确认取消', { type: 'warning' })
    
    // 收集该容器的所有 runId
    const runIds = new Set<string>()
    const singleTasks: number[] = []
    
    for (const task of runningTasksForContainer.value) {
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
    
    ElMessage.success('取消请求已发送')
    fetchRunningTasks()
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

onMounted(() => {
  fetchData()
  fetchRunningTasks()
  // 定时刷新运行中任务列表
  refreshTimer = window.setInterval(fetchRunningTasks, 5000)
})

import { onUnmounted } from 'vue'

onUnmounted(() => {
  if (refreshTimer != null) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<style lang="scss" scoped>
.container-config {
  .page-title {
    font-size: 18px;
    font-weight: bold;
    color: var(--text-primary);
  }
  .page-header-content {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    color: var(--text-primary);
  }
  .graph-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    .graph-tip {
      font-size: 12px;
      color: var(--text-muted);
    }
  }
  .task-list {
    max-height: 500px;
    overflow-y: auto;
  }
  .task-item {
    padding: 10px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    margin-bottom: 10px;
    cursor: pointer;
    transition: all 0.3s;
    background: var(--bg-card);

    &:hover {
      border-color: var(--primary-color);
      background: rgba(var(--primary-color), 0.05);
    }
    &.active {
      border-color: var(--primary-color);
      background: rgba(var(--primary-color), 0.1);
    }
    .task-name {
      font-weight: bold;
      margin-bottom: 5px;
      color: var(--text-primary);
    }
    .task-command {
      font-size: 12px;
      color: var(--text-secondary);
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
  .graph-container {
    height: 500px;
    border: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }
}

// 对话框样式适配
:deep(.el-dialog) {
  background: var(--bg-card) !important;

  .el-dialog__header {
    color: var(--text-primary);
  }

  .el-dialog__body {
    color: var(--text-primary);
  }

  .el-form-item__label {
    color: var(--text-secondary) !important;
  }

  .el-input__wrapper {
    background: var(--bg-secondary) !important;
  }

  .el-input__inner {
    color: var(--text-primary) !important;
  }

  .el-textarea__inner {
    background: var(--bg-secondary) !important;
    color: var(--text-primary) !important;
  }

  .el-input-number {
    .el-input__wrapper {
      background: var(--bg-secondary) !important;
    }
  }
}

// 卡片样式适配
:deep(.el-card) {
  background: var(--bg-card) !important;
  border-color: var(--border-color) !important;

  .el-card__header {
    border-bottom-color: var(--border-color) !important;
    color: var(--text-primary);
  }

  .el-card__body {
    color: var(--text-primary);
  }
}
</style>
