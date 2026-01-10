<template>
  <div class="container-list">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-button type="primary" @click="showDialog = true">
            <el-icon><Plus /></el-icon>
            新增容器
          </el-button>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="cid" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="expression" label="Cron 表达式" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'info' : 'success'" size="small">
              {{ row.status === 1 ? '等待' : '运行中' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="disable" label="启用" width="100">
          <template #default="{ row }">
            <el-switch v-model="row.disable" :active-value="false" :inactive-value="true" @change="handleToggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleConfig(row)">配置</el-button>
            <el-button type="success" link @click="handleRun(row)">运行</el-button>
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

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="showDialog" :title="isEdit ? '编辑容器' : '新增容器'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入容器名称" />
        </el-form-item>
        <el-form-item label="Cron 表达式" prop="expression">
          <el-input v-model="form.expression" placeholder="如: 0 0 * * * 或 @every 1h" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
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
.container-list {
  .card-header {
    display: flex;
    justify-content: flex-end;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
