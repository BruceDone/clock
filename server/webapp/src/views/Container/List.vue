<template>
  <div class="container-list">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">容器管理</h1>
        <p class="page-subtitle">管理任务容器和定时调度</p>
      </div>
    </div>

    <!-- 操作栏 -->
    <el-card class="filter-card">
      <div class="filter-bar">
        <div class="filter-left">
          <div class="container-stats">
            <div class="stat-item">
              <span class="stat-value">{{ list.length }}</span>
              <span class="stat-label">总容器数</span>
            </div>
          </div>
        </div>
        <div class="filter-right">
          <el-button type="primary" @click="showDialog = true">
            <el-icon><Plus /></el-icon>
            新增容器
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 容器表格 -->
    <el-card class="table-card">
      <el-table :data="list" v-loading="loading" stripe class="container-table">
        <el-table-column prop="cid" label="ID" width="80">
          <template #default="{ row }">
            <span class="cell-id">#{{ row.cid }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="容器名称" min-width="150">
          <template #default="{ row }">
            <div class="container-name">
              <el-icon><Box /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="expression" label="Cron 表达式" min-width="150">
          <template #default="{ row }">
            <code class="cron-text">{{ row.expression }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'info' : 'success'" class="status-tag" effect="dark">
              <span class="status-dot" :class="row.status === 1 ? 'status-pending' : 'status-running'"></span>
              {{ row.status === 1 ? '等待' : '运行中' }}
            </el-tag>
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
              <el-button type="primary" link @click="handleConfig(row)">
                <el-icon><Setting /></el-icon>
                配置
              </el-button>
              <el-button type="success" link @click="handleRun(row)">
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
    <el-dialog v-model="showDialog" :title="isEdit ? '编辑容器' : '新增容器'" width="500px" class="container-dialog">
      <template #header>
        <div class="dialog-header">
          <el-icon><component :is="isEdit ? 'Edit' : 'Plus'" /></el-icon>
          <span>{{ isEdit ? '编辑容器' : '新增容器' }}</span>
        </div>
      </template>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" class="container-form">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入容器名称">
            <template #prefix>
              <el-icon><Box /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="Cron 表达式" prop="expression">
          <el-input v-model="form.expression" placeholder="如: 0 0 * * * 或 @every 1h">
            <template #prefix>
              <el-icon><Clock /></el-icon>
            </template>
          </el-input>
          <div class="cron-hint">
            <el-icon><InfoFilled /></el-icon>
            <span>使用标准 cron 格式或 @every 语法</span>
          </div>
        </el-form-item>
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
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getContainers, putContainer, deleteContainer, runContainer } from '@/api/container'
import type { Container } from '@/types/model'

const router = useRouter()

const loading = ref(false)
const list = ref<Container[]>([])
const showDialog = ref(false)
const isEdit = ref(false)
const editingId = ref<number>()

const page = reactive({
  count: 10,
  index: 1,
  total: 0
})

const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  expression: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  expression: [{ required: true, message: '请输入 Cron 表达式', trigger: 'blur' }]
}

async function fetchList() {
  loading.value = true
  try {
    const res = await getContainers({ count: page.count, index: page.index })
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
    await putContainer({ name: form.name, expression: form.expression, cid: editingId.value })
    ElMessage.success('操作成功')
    showDialog.value = false
    fetchList()
  } catch (error) {
    console.error(error)
  }
}

async function handleDelete(row: Container) {
  try {
    await ElMessageBox.confirm('确定要删除该容器吗？', '提示', { type: 'warning' })
    await deleteContainer(row.cid)
    ElMessage.success('删除成功')
    fetchList()
  } catch (error) {
    console.error(error)
  }
}

async function handleRun(row: Container) {
  try {
    await runContainer(row.cid)
    ElMessage.success('触发运行成功')
  } catch (error) {
    console.error(error)
  }
}

function handleConfig(row: Container) {
  router.push(`/container/config/${row.cid}`)
}

async function handleToggle(row: Container) {
  try {
    await putContainer(row)
    ElMessage.success(row.disable ? '已禁用' : '已启用')
  } catch (error) {
    console.error(error)
    row.disable = !row.disable
  }
}

onMounted(() => {
  fetchList()
})
</script>

<style lang="scss" scoped>
// Container List 页面样式 - 使用 CSS 变量

.container-list {
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

      .container-stats {
        display: flex;
        gap: 24px;

        .stat-item {
          display: flex;
          align-items: center;
          gap: 8px;

          .stat-value {
            font-size: 24px;
            font-weight: 700;
            color: var(--primary-color);
            font-family: var(--font-family-mono);
          }

          .stat-label {
            font-size: 14px;
            color: var(--text-secondary);
          }
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

  .container-table {
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

    .container-name {
      display: flex;
      align-items: center;
      gap: 8px;
      color: var(--text-primary);

      .el-icon {
        color: var(--primary-color);
      }
    }

    .cron-text {
      font-family: var(--font-family-mono);
      font-size: 12px;
      color: var(--info-color);
      background: rgba(var(--info-color), 0.1);
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

        &.status-pending {
          background: var(--text-muted);
        }

        &.status-running {
          background: var(--success-color);
          animation: status-pulse 1s infinite;
        }
      }
    }

    @keyframes status-pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.5; }
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

.container-dialog {
  .dialog-header {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--primary-color);
    font-size: 18px;
    font-weight: 600;
  }

  .container-form {
    :deep(.el-form-item__label) {
      color: var(--text-secondary);
    }

    :deep(.el-input__wrapper) {
      background: var(--input-bg) !important;
    }

    .cron-hint {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-top: 8px;
      font-size: 12px;
      color: var(--text-muted);

      .el-icon {
        font-size: 14px;
      }
    }
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
}
</style>
