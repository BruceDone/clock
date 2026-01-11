<template>
  <div class="task-list">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">任务管理</h1>
        <p class="page-subtitle">管理系统任务和执行计划</p>
      </div>
    </div>

    <!-- 操作栏 -->
    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-left">
          <el-select
            v-model="filterCid"
            placeholder="选择容器"
            clearable
            class="filter-select"
            @change="fetchList"
          >
            <template #prefix>
              <el-icon><Box /></el-icon>
            </template>
            <el-option v-for="c in containers" :key="c.cid" :label="c.name" :value="c.cid" />
          </el-select>
        </div>
        <div class="filter-right">
          <el-button type="primary" @click="showDialog = true">
            <el-icon><Plus /></el-icon>
            新增任务
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 任务表格 -->
    <el-card class="table-card">
      <el-table :data="list" v-loading="loading" stripe class="task-table">
        <el-table-column prop="tid" label="ID" width="80">
          <template #default="{ row }">
            <span class="cell-id">#{{ row.tid }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="任务名称" min-width="150">
          <template #default="{ row }">
            <div class="task-name">
              <el-icon><List /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="command" label="执行命令" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <code class="command-text">{{ row.command }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" class="status-tag" effect="dark">
              <span class="status-dot" :class="`status-${row.status}`"></span>
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="timeout" label="超时" width="100">
          <template #default="{ row }">
            <span class="cell-time">{{ row.timeout }}s</span>
          </template>
        </el-table-column>
        <el-table-column prop="log_enable" label="日志" width="80">
          <template #default="{ row }">
            <el-icon v-if="row.log_enable" class="cell-icon success"><CircleCheckFilled /></el-icon>
            <el-icon v-else class="cell-icon disabled"><CircleCloseFilled /></el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="disable" label="启用" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.disable"
              :active-value="false"
              :inactive-value="true"
              active-color="#00ff88"
              inactive-color="#5c6b7f"
              @change="handleToggle(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button type="primary" link @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button v-if="row.status === 2" type="warning" link @click="handleCancel(row)">
                <el-icon><CircleClose /></el-icon>
                取消
              </el-button>
              <el-button v-else type="success" link @click="handleRun(row)">
                <el-icon><VideoPlay /></el-icon>
                运行
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
          background
          @size-change="fetchList"
          @current-change="fetchList"
        />
      </div>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="showDialog" :title="isEdit ? '编辑任务' : '新增任务'" width="600px" class="task-dialog">
      <template #header>
        <div class="dialog-header">
          <el-icon><component :is="isEdit ? 'Edit' : 'Plus'" /></el-icon>
          <span>{{ isEdit ? '编辑任务' : '新增任务' }}</span>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" class="task-form">
        <el-form-item label="所属容器" prop="cid">
          <el-select v-model="form.cid" placeholder="选择容器" style="width: 100%">
            <template #prefix>
              <el-icon><Box /></el-icon>
            </template>
            <el-option v-for="c in containers" :key="c.cid" :label="c.name" :value="c.cid" />
          </el-select>
        </el-form-item>
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称">
            <template #prefix>
              <el-icon><List /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="Bash 命令" prop="command">
          <el-input v-model="form.command" type="textarea" :rows="3" placeholder="请输入要执行的 bash 命令" />
        </el-form-item>
        <el-form-item label="工作目录" prop="directory">
          <el-input v-model="form.directory" placeholder="命令执行的工作目录，留空则使用默认目录">
            <template #prefix>
              <el-icon><FolderOpened /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="超时时间(秒)">
              <el-input-number v-model="form.timeout" :min="0" :step="10" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="启用日志">
              <el-switch v-model="form.log_enable" active-text="是" inactive-text="否" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showDialog = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">
            <el-icon><Check /></el-icon>
            确定
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getContainers } from '@/api/container'
import { getTasks, putTask, deleteTask, runTask, cancelTask } from '@/api/task'
import type { Container, Task } from '@/types/model'

const loading = ref(false)
const list = ref<Task[]>([])
const containers = ref<Container[]>([])
const filterCid = ref<number | undefined>()
const showDialog = ref(false)
const isEdit = ref(false)
const editingId = ref<number>()

const page = reactive({ count: 10, index: 1, total: 0 })

const formRef = ref<FormInstance>()
const form = reactive({
  cid: 0,
  name: '',
  command: '',
  directory: '',
  timeout: 30,
  log_enable: true
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  command: [{ required: true, message: '请输入命令', trigger: 'blur' }]
}

function getStatusType(status: number) {
  const map: Record<number, string> = { 1: 'info', 2: 'primary', 3: 'success', 4: 'danger', 5: 'warning' }
  return map[status] || 'info'
}

function getStatusText(status: number) {
  const map: Record<number, string> = { 1: '等待', 2: '运行中', 3: '成功', 4: '失败', 5: '已取消' }
  return map[status] || '未知'
}

async function fetchContainers() {
  const res = await getContainers({ count: 100 })
  containers.value = (res.data as any)?.items || []
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getTasks({ count: page.count, index: page.index, cid: filterCid.value })
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

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  try {
    await putTask({ ...form, tid: editingId.value })
    ElMessage.success('操作成功')
    showDialog.value = false
    fetchList()
  } catch (error) {
    console.error(error)
  }
}

async function handleDelete(row: Task) {
  try {
    await ElMessageBox.confirm('确定要删除该任务吗？', '提示', { type: 'warning' })
    await deleteTask(row.tid)
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    console.error(error)
  }
}

async function handleRun(row: Task) {
  try {
    await runTask(row.tid)
    ElMessage.success('触发运行成功')
    fetchList()
  } catch (error) {
    console.error(error)
  }
}

async function handleCancel(row: Task) {
  try {
    await ElMessageBox.confirm('确定要取消该任务吗？', '确认取消', { type: 'warning' })
    await cancelTask(row.tid)
    ElMessage.success('任务取消请求已发送')
    fetchList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(error)
    }
  }
}

async function handleToggle(row: Task) {
  try {
    await putTask(row)
    ElMessage.success(row.disable ? '已禁用' : '已启用')
  } catch (error) {
    console.error(error)
    row.disable = !row.disable
  }
}

function handleEdit(row: Task) {
  isEdit.value = true
  editingId.value = row.tid
  form.cid = row.cid
  form.name = row.name
  form.command = row.command
  form.directory = row.directory
  form.timeout = row.timeout
  form.log_enable = row.log_enable
  showDialog.value = true
}

onMounted(() => {
  fetchContainers()
  fetchList()
})
</script>

<style lang="scss" scoped>
// Task List 页面样式 - 使用 CSS 变量

.task-list {
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

      .filter-select {
        width: 240px;
      }

      :deep(.el-input__wrapper) {
        background: var(--input-bg) !important;
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

  .task-table {
    :deep(.el-table__row) {
      transition: all 0.2s ease;
      background: var(--table-bg) !important;

      &:hover {
        background: var(--table-row-hover) !important;
      }
    }

    .cell-id {
      font-family: var(--font-family-mono);
      color: var(--primary-color);
      font-weight: 600;
    }

    .task-name {
      display: flex;
      align-items: center;
      gap: 8px;
      color: var(--text-primary);

      .el-icon {
        color: var(--primary-color);
      }
    }

    .command-text {
      font-family: var(--font-family-mono);
      font-size: 12px;
      color: var(--text-secondary);
      background: rgba(var(--primary-color), 0.1);
      padding: 4px 8px;
      border-radius: 4px;
    }

    .status-tag {
      border: none !important;
      display: flex;
      align-items: center;
      gap: 6px;

      .status-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: currentColor;

        &.status-1 { background: var(--text-muted); }
        &.status-2 { background: var(--info-color); animation: status-pulse 1s infinite; }
        &.status-3 { background: var(--success-color); }
        &.status-4 { background: var(--danger-color); }
        &.status-5 { background: var(--warning-color); }
      }
    }

    @keyframes status-pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.5; }
    }

    .cell-time {
      font-family: var(--font-family-mono);
      color: var(--text-secondary);
    }

    .cell-icon {
      font-size: 18px;

      &.success { color: var(--success-color); }
      &.disabled { color: var(--text-muted); }
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

    :deep(.el-pagination) {
      --el-pagination-bg-color: var(--bg-primary);
      --el-pagination-text-color: var(--text-primary);
      --el-pagination-button-bg-color: var(--bg-secondary);
      --el-pagination-hover-color: var(--primary-color);
    }
  }
}

.task-dialog {
  .dialog-header {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--primary-color);
    font-size: 18px;
    font-weight: 600;
  }

  .task-form {
    :deep(.el-form-item__label) {
      color: var(--text-secondary);
    }

    :deep(.el-input__wrapper) {
      background: var(--input-bg) !important;
    }

    :deep(.el-textarea__inner) {
      background: var(--input-bg) !important;
      color: var(--text-primary) !important;
    }

    :deep(.el-input-number) {
      .el-input__wrapper {
        background: var(--input-bg) !important;
      }
    }

    :deep(.el-switch) {
      --el-switch-off-color: var(--text-muted);
    }
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
}
</style>
