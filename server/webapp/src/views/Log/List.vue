<template>
  <div class="log-list">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">日志中心</h1>
        <p class="page-subtitle">查看任务执行日志和历史记录</p>
      </div>
    </div>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-left">
          <el-select 
            v-model="filters.cid" 
            placeholder="选择容器" 
            clearable 
            class="filter-select"
            @change="handleContainerChange"
          >
            <template #prefix>
              <el-icon><Box /></el-icon>
            </template>
            <el-option v-for="c in containers" :key="c.cid" :label="c.name" :value="c.cid" />
          </el-select>
          <el-select 
            v-model="filters.tid" 
            placeholder="选择任务" 
            clearable 
            class="filter-select"
            @change="fetchList"
          >
            <template #prefix>
              <el-icon><List /></el-icon>
            </template>
            <el-option v-for="t in filteredTasks" :key="t.tid" :label="t.name" :value="t.tid" />
          </el-select>
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            class="date-picker"
            @change="handleDateChange"
          />
        </div>
        <div class="filter-right">
          <el-button type="primary" @click="fetchList">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="handleClearFilter">
            <el-icon><RefreshRight /></el-icon>
            重置
          </el-button>
          <el-button type="danger" @click="handleClearAllLogs">
            <el-icon><Delete /></el-icon>
            清除所有
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 日志表格 -->
    <el-card class="table-card">
      <el-table :data="list" v-loading="loading" stripe class="log-table">
        <el-table-column prop="lid" label="日志ID" min-width="180">
          <template #default="{ row }">
            <span class="cell-id">{{ row.lid }}</span>
          </template>
        </el-table-column>
        <el-table-column label="任务" min-width="150">
          <template #default="{ row }">
            <div class="task-name">
              <el-icon><List /></el-icon>
              <span>{{ getTaskName(row.tid) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="容器" min-width="150">
          <template #default="{ row }">
            <div class="container-name">
              <el-icon><Box /></el-icon>
              <span>{{ getContainerName(row.cid) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="update_at" label="执行时间" min-width="180">
          <template #default="{ row }">
            <span class="cell-time">{{ formatTime(row.update_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" link @click="showDetail(row)">
                <el-icon><View /></el-icon>
                查看
              </el-button>
              <el-button type="danger" link @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page.index"
          v-model:page-size="page.count"
          :total="page.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 日志详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="日志详情" width="800px" class="log-dialog">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="标准输出" name="stdout">
          <pre class="log-content stdout">{{ currentLog?.std_out || '(无输出)' }}</pre>
        </el-tab-pane>
        <el-tab-pane label="错误输出" name="stderr">
          <pre class="log-content stderr">{{ currentLog?.std_err || '(无输出)' }}</pre>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Box, List, Search, RefreshRight, Delete, View } from '@element-plus/icons-vue'
import { getLogs, deleteLogByID, deleteAllLogs } from '@/api/log'
import { getContainers } from '@/api/container'
import { getTasks } from '@/api/task'
import type { TaskLog, Container, Task } from '@/types/model'

const loading = ref(false)
const list = ref<TaskLog[]>([])
const containers = ref<Container[]>([])
const tasks = ref<Task[]>([])

const showDetailDialog = ref(false)
const currentLog = ref<TaskLog | null>(null)
const activeTab = ref('stdout')
const dateRange = ref<[Date, Date] | null>(null)

const page = reactive({ count: 10, index: 1, total: 0 })
const filters = reactive({
  cid: undefined as number | undefined,
  tid: undefined as number | undefined,
  left_ts: 0,
  right_ts: 0
})

// 根据选中的容器筛选任务列表
const filteredTasks = computed(() => {
  if (!filters.cid) {
    return tasks.value
  }
  return tasks.value.filter(t => t.cid === filters.cid)
})

// 获取任务名称
function getTaskName(tid: number): string {
  const task = tasks.value.find(t => t.tid === tid)
  return task?.name || `任务${tid}`
}

// 获取容器名称
function getContainerName(cid: number): string {
  const container = containers.value.find(c => c.cid === cid)
  return container?.name || `容器${cid}`
}

function formatTime(timestamp: number) {
  return new Date(timestamp * 1000).toLocaleString()
}

function handleDateChange() {
  if (dateRange.value) {
    filters.left_ts = Math.floor(dateRange.value[0].getTime() / 1000)
    filters.right_ts = Math.floor(dateRange.value[1].getTime() / 1000)
  } else {
    filters.left_ts = 0
    filters.right_ts = 0
  }
}

function handleContainerChange() {
  // 当容器改变时，清空任务选择
  filters.tid = undefined
  fetchList()
}

function handleClearFilter() {
  filters.cid = undefined
  filters.tid = undefined
  filters.left_ts = 0
  filters.right_ts = 0
  dateRange.value = null
  fetchList()
}

async function handleClearAllLogs() {
  try {
    await ElMessageBox.confirm(
      '确定要清除所有日志吗？此操作不可恢复！', 
      '警告', 
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
    await deleteAllLogs()
    ElMessage.success('已清除所有日志')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

async function fetchContainers() {
  const res = await getContainers({ count: 100 })
  containers.value = (res.data as any)?.items || []
}

async function fetchTasks() {
  const res = await getTasks({ count: 1000 })
  tasks.value = (res.data as any)?.items || []
}

async function fetchList() {
  loading.value = true
  try {
    const params = {
      count: page.count,
      index: page.index,
      cid: filters.cid,
      tid: filters.tid,
      left_ts: filters.left_ts || undefined,
      right_ts: filters.right_ts || undefined
    }
    const res = await getLogs(params)
    if (res.data) {
      list.value = res.data.items || []
      page.total = res.data.page?.total || 0
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

function showDetail(log: TaskLog) {
  currentLog.value = log
  activeTab.value = 'stdout'
  showDetailDialog.value = true
}

async function handleDelete(log: TaskLog) {
  try {
    await ElMessageBox.confirm('确定要删除该日志吗？', '提示', { type: 'warning' })
    await deleteLogByID(log.lid)
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

onMounted(() => {
  fetchContainers()
  fetchTasks()
  fetchList()
})
</script>

<style lang="scss" scoped>
.log-list {
  .page-header {
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
  }

  .filter-card {
    background: var(--bg-card) !important;
    border: 1px solid var(--border-color) !important;
    border-radius: var(--border-radius-lg) !important;
    margin-bottom: 24px;
    backdrop-filter: blur(10px);

    :deep(.el-card__body) {
      padding: 16px 20px;
    }

    .filter-bar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-wrap: wrap;
      gap: 12px;

      .filter-left {
        display: flex;
        align-items: center;
        gap: 12px;
        flex-wrap: wrap;
      }

      .filter-right {
        display: flex;
        align-items: center;
        gap: 8px;
      }

      .filter-select {
        width: 200px;
      }

      .date-picker {
        width: 360px;
      }

      :deep(.el-input__wrapper) {
        background: var(--input-bg) !important;
      }

      :deep(.el-input__inner) {
        color: var(--text-primary) !important;
      }

      // 修复下拉框选中后的文字颜色
      :deep(.el-select) {
        .el-input__inner {
          color: var(--text-primary) !important;
        }
      }
    }
  }

  .table-card {
    background: var(--bg-card) !important;
    border: 1px solid var(--border-color) !important;
    border-radius: var(--border-radius-lg) !important;
    backdrop-filter: blur(10px);

    :deep(.el-card__body) {
      padding: 0;
    }
  }

  .log-table {
    width: 100%;

    :deep(.el-table__row) {
      transition: all 0.2s ease;
      background: var(--table-bg) !important;

      &:hover {
        background: var(--table-row-hover) !important;
      }
    }

    .cell-id {
      font-family: var(--font-family-mono);
      color: var(--text-secondary);
      font-size: 12px;
    }

    .task-name,
    .container-name {
      display: flex;
      align-items: center;
      gap: 8px;
      color: var(--text-primary);

      .el-icon {
        color: var(--primary-color);
      }
    }

    .cell-time {
      font-family: var(--font-family-mono);
      color: var(--text-secondary);
    }

    .action-buttons {
      display: flex;
      gap: 8px;

      .el-button {
        padding: 4px 8px;
        font-size: 13px;

        .el-icon {
          margin-right: 4px;
        }
      }
    }
  }

  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    padding: 20px;
    border-top: 1px solid var(--border-color-light);
  }

  .log-content {
    max-height: 400px;
    overflow: auto;
    padding: 15px;
    background: var(--bg-primary);
    color: var(--text-primary);
    border-radius: var(--border-radius-base);
    font-family: 'Consolas', 'Monaco', monospace;
    font-size: 13px;
    white-space: pre-wrap;
    word-break: break-all;
    border: 1px solid var(--border-color);

    &.stdout {
      background: var(--bg-primary);
    }

    &.stderr {
      background: rgba(var(--danger-color), 0.1);
      border-color: rgba(var(--danger-color), 0.3);
    }
  }
}
</style>
