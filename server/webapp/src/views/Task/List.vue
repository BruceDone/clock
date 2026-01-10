<template>
  <div class="task-list">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <el-select v-model="filterCid" placeholder="选择容器" clearable style="width: 200px" @change="fetchList">
            <el-option v-for="c in containers" :key="c.cid" :label="c.name" :value="c.cid" />
          </el-select>
          <el-button type="primary" @click="showDialog = true">
            <el-icon><Plus /></el-icon>
            新增任务
          </el-button>
        </div>
      </template>

      <el-table :data="list" v-loading="loading" stripe>
        <el-table-column prop="tid" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="command" label="命令" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="timeout" label="超时(秒)" width="100" />
        <el-table-column prop="log_enable" label="日志" width="80">
          <template #default="{ row }">
            <el-tag :type="row.log_enable ? 'success' : 'info'" size="small">
              {{ row.log_enable ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="disable" label="启用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.disable" :active-value="false" :inactive-value="true" @change="handleToggle(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
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
    <el-dialog v-model="showDialog" :title="isEdit ? '编辑任务' : '新增任务'" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="所属容器">
          <el-select v-model="form.cid" placeholder="选择容器" style="width: 100%">
            <el-option v-for="c in containers" :key="c.cid" :label="c.name" :value="c.cid" />
          </el-select>
        </el-form-item>
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="Bash 命令" prop="command">
          <el-input v-model="form.command" type="textarea" :rows="3" placeholder="请输入要执行的 bash 命令" />
        </el-form-item>
        <el-form-item label="工作目录" prop="directory">
          <el-input v-model="form.directory" placeholder="命令执行的工作目录，留空则使用默认目录" />
        </el-form-item>
        <el-form-item label="超时时间(秒)">
          <el-input-number v-model="form.timeout" :min="0" :step="10" />
        </el-form-item>
        <el-form-item label="启用日志">
          <el-switch v-model="form.log_enable" />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getContainers } from '@/api/container'
import { getTasks, putTask, deleteTask, runTask } from '@/api/task'
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
  const map: Record<number, string> = { 1: 'info', 2: 'primary', 3: 'success', 4: 'danger' }
  return map[status] || 'info'
}

function getStatusText(status: number) {
  const map: Record<number, string> = { 1: '等待', 2: '运行中', 3: '成功', 4: '失败' }
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
  } catch (error) {
    console.error(error)
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
.task-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
