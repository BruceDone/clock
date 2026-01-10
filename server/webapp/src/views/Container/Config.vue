<template>
  <div class="container-config">
    <el-page-header @back="router.back()" style="margin-bottom: 20px">
      <template #content>
        <span class="page-title">容器配置 - {{ container?.name }}</span>
      </template>
    </el-page-header>

    <el-row :gutter="20">
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <div class="card-header">
              <span>任务列表</span>
              <el-button type="primary" size="small" @click="showTaskDialog = true">新增任务</el-button>
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
              <el-button-group>
                <el-button size="small" @click="fitView">适应视图</el-button>
                <el-button size="small" @click="savePositions">保存位置</el-button>
              </el-button-group>
            </div>
          </template>
          <div ref="graphRef" class="graph-container"></div>
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
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Graph } from '@antv/g6'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getContainer } from '@/api/container'
import { getTasks, putTask } from '@/api/task'
import { getRelations } from '@/api/relation'
import { putNodes } from '@/api/node'
import type { Container, Task, RelationResponse, Node } from '@/types/model'

const route = useRoute()
const router = useRouter()
const cid = parseInt(route.params.cid as string)

const container = ref<Container | null>(null)
const tasks = ref<Task[]>([])
const relations = ref<RelationResponse | null>(null)
const selectedTask = ref<Task | null>(null)

const graphRef = ref<HTMLElement>()
let graph: Graph | null = null

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

function getNodeColor(status: number): string {
  const map: Record<number, string> = { 1: '#909399', 2: '#409EFF', 3: '#67C23A', 4: '#F56C6C' }
  return map[status] || '#909399'
}

function initGraph() {
  if (!graphRef.value) return

  const width = graphRef.value.clientWidth
  const height = graphRef.value.clientHeight || 500

  graph = new Graph({
    container: graphRef.value,
    width,
    height,
    data: { nodes: [], edges: [] },
    node: {
      style: {
        size: 60,
        fill: '#fff',
        stroke: '#5B8FF9',
        lineWidth: 2
      }
    },
    edge: {
      style: {
        stroke: '#A3B1BF',
        lineWidth: 2,
        endArrow: true
      }
    },
    behaviors: ['drag-canvas', 'zoom-canvas', 'drag-node']
  })

  updateGraph()
  window.addEventListener('resize', handleResize)
}

function updateGraph() {
  if (!graph || !relations.value) return

  const nodes = relations.value.nodes.map((node: Node) => ({
    id: node.id.toString(),
    label: node.name,
    x: node.x,
    y: node.y,
    style: { fill: getNodeColor(node.status) }
  }))

  const edges = relations.value.links.map((link: any) => ({
    source: link.tid.toString(),
    target: link.next_tid.toString()
  }))

  graph.setData({ nodes, edges })
  graph.render()
}

function handleResize() {
  if (!graph || !graphRef.value) return
  graph.setSize(graphRef.value.clientWidth, graphRef.value.clientHeight || 500)
}

function fitView() {
  graph?.fitView()
}

async function fetchData() {
  try {
    const [containerRes, tasksRes, relationsRes] = await Promise.all([
      getContainer(cid),
      getTasks({ cid }),
      getRelations({ cid })
    ])

    container.value = containerRes.data
    tasks.value = (tasksRes.data as any)?.items || []
    relations.value = relationsRes.data

    if (graph) {
      updateGraph()
    } else {
      initGraph()
    }
  } catch (error) {
    console.error(error)
  }
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
  if (!graph || !relations.value) return

  const nodeData = relations.value.nodes.map((node: any) => ({
    id: node.id,
    name: node.name,
    status: node.status,
    x: node.x,
    y: node.y
  }))

  try {
    await putNodes(nodeData)
    ElMessage.success('位置保存成功')
  } catch (error) {
    console.error(error)
  }
}

onMounted(() => {
  fetchData()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  graph?.destroy()
})
</script>

<style lang="scss" scoped>
.container-config {
  .page-title { font-size: 18px; font-weight: bold; }
  .card-header { display: flex; justify-content: space-between; align-items: center; }
  .task-list { max-height: 500px; overflow-y: auto; }
  .task-item {
    padding: 10px;
    border: 1px solid #eee;
    border-radius: 4px;
    margin-bottom: 10px;
    cursor: pointer;
    transition: all 0.3s;
    &:hover { border-color: #409EFF; }
    &.active { border-color: #409EFF; background: #ecf5ff; }
    .task-name { font-weight: bold; margin-bottom: 5px; }
    .task-command {
      font-size: 12px;
      color: #666;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }
  .graph-container { height: 500px; border: 1px solid #eee; }
}
</style>
