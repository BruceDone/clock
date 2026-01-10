<template>
  <div class="log-list">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-form :inline="true" :model="filters">
            <el-form-item label="容器">
              <el-select v-model="filters.cid" placeholder="选择容器" clearable @change="fetchList">
                <el-option v-for="c in containers" :key="c.cid" :label="c.name" :value="c.cid" />
              </el-select>
            </el-form-item>
            <el-form-item label="任务">
              <el-select v-model="filters.tid" placeholder="选择任务" clearable @change="fetchList">
                <el-option v-for="t in tasks" :key="t.tid" :label="t.name" :value="t.tid" />
              </el-select>
            </el-form-item>
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="dateRange"
                type="datetimerange"
                range-separator="至"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
                @change="handleDateChange"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchList">查询</el-button>
              <el-button @click="handleClear">清空筛选</el-button>
            </el-form-item>
          </el-form>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="lid" label="ID" width="200" />
        <el-table-column prop="tid" label="任务ID" width="80" />
        <el-table-column prop="cid" label="容器ID" width="80" />
        <el-table-column prop="update_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.update_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template #default="{ row }" width="100">
            <el-button type="primary" link @click="showDetail(row)">查看</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
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
    <el-dialog v-model="showDetailDialog" title="日志详情" width="800px">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="标准输出" name="stdout">
          <pre class="log-content stdout">{{ currentLog?.std_out }}</pre>
        </el-tab-pane>
        <el-tab-pane label="错误输出" name="stderr">
          <pre class="log-content stderr">{{ currentLog?.std_err }}</pre>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getLogs, deleteLogs } from '@/api/log'
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

function handleClear() {
  filters.cid = undefined
  filters.tid = undefined
  filters.left_ts = 0
  filters.right_ts = 0
  dateRange.value = null
  fetchList()
}

async function fetchContainers() {
  const res = await getContainers({ count: 100 })
  containers.value = (res.data as any)?.items || []
}

async function fetchTasks() {
  const res = await getTasks({ count: 100 })
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
    await deleteLogs({ tid: log.tid, cid: log.cid })
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    console.error(error)
  }
}

watch(() => filters.cid, (newCid) => {
  if (newCid) {
    fetchTasks()
  }
})

onMounted(() => {
  fetchContainers()
  fetchTasks()
  fetchList()
})
</script>

<style lang="scss" scoped>
.log-list {
  .card-header {
    :deep(.el-form) {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .log-content {
    max-height: 400px;
    overflow: auto;
    padding: 15px;
    background: #1e1e1e;
    color: #fff;
    border-radius: 4px;
    font-family: 'Consolas', monospace;
    font-size: 13px;
    white-space: pre-wrap;
    word-break: break-all;

    &.stdout {
      background: #1e1e1e;
    }

    &.stderr {
      background: #2d0000;
    }
  }
}
</style>
