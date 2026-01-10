<template>
  <div class="home-page">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6" v-for="item in stats" :key="item.title">
        <el-card shadow="hover" :body-style="{ padding: '20px' }">
          <div class="stat-card">
            <div class="stat-icon" :style="{ background: item.color }">
              <el-icon :size="24"><component :is="item.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ item.count }}</div>
              <div class="stat-label">{{ item.title }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <span>内存使用</span>
          </template>
          <div ref="memChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <span>系统负载</span>
          </template>
          <div ref="loadChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="charts-row">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <span>CPU 使用率</span>
          </template>
          <div ref="cpuChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { getMessages } from '@/api/message'
import { getSystemLoad, getSystemMem, getSystemCpu } from '@/api/system'

const memChartRef = ref<HTMLElement>()
const loadChartRef = ref<HTMLElement>()
const cpuChartRef = ref<HTMLElement>()

let memChart: echarts.ECharts | null = null
let loadChart: echarts.ECharts | null = null
let cpuChart: echarts.ECharts | null = null

interface StatItem {
  title: string
  icon: string
  count: number
  color: string
}

const stats = ref<StatItem[]>([
  { title: '等待运行', icon: 'Clock', count: 0, color: '#909399' },
  { title: '正在运行', icon: 'Loading', count: 0, color: '#409EFF' },
  { title: '运行成功', icon: 'SuccessFilled', count: 0, color: '#67C23A' },
  { title: '运行失败', icon: 'CloseFilled', count: 0, color: '#F56C6C' }
])

async function fetchStats() {
  try {
    const res = await getMessages()
    if (res.data) {
      stats.value = res.data.map((item: any) => ({
        title: item.title,
        icon: item.icon || 'InfoFilled',
        count: item.count,
        color: item.color || '#909399'
      }))
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

async function fetchMemChart() {
  try {
    const res = await getSystemMem()
    if (res.data) {
      const option = {
        tooltip: { trigger: 'item' as const },
        legend: { bottom: 0 },
        series: [{
          type: 'pie',
          radius: ['40%', '70%'],
          avoidLabelOverlap: false,
          label: { show: false },
          data: res.data
        }]
      }
      memChart?.setOption(option)
    }
  } catch (error) {
    console.error('获取内存数据失败:', error)
  }
}

async function fetchLoadChart() {
  try {
    const res = await getSystemLoad()
    if (res.data) {
      const option = {
        tooltip: { trigger: 'axis' as const },
        legend: { data: ['1分钟', '5分钟', '15分钟'] },
        xAxis: { type: 'category', data: ['当前'] },
        yAxis: { type: 'value' },
        series: [
          { name: '1分钟', type: 'bar', data: [res.data.load1] },
          { name: '5分钟', type: 'bar', data: [res.data.load5] },
          { name: '15分钟', type: 'bar', data: [res.data.load15] }
        ]
      }
      loadChart?.setOption(option)
    }
  } catch (error) {
    console.error('获取负载数据失败:', error)
  }
}

async function fetchCpuChart() {
  try {
    const res = await getSystemCpu()
    if (res.data) {
      const names = Object.keys(res.data)
      const values = Object.values(res.data)
      const option = {
        tooltip: { trigger: 'axis' as const },
        legend: { data: names },
        xAxis: { type: 'category', data: ['CPU 使用率'] },
        yAxis: { type: 'value', max: 100 },
        series: names.map((name, index) => ({
          name,
          type: 'bar',
          data: [values[index]]
        }))
      }
      cpuChart?.setOption(option)
    }
  } catch (error) {
    console.error('获取 CPU 数据失败:', error)
  }
}

function initCharts() {
  if (memChartRef.value) {
    memChart = echarts.init(memChartRef.value)
  }
  if (loadChartRef.value) {
    loadChart = echarts.init(loadChartRef.value)
  }
  if (cpuChartRef.value) {
    cpuChart = echarts.init(cpuChartRef.value)
  }
}

function handleResize() {
  memChart?.resize()
  loadChart?.resize()
  cpuChart?.resize()
}

onMounted(() => {
  initCharts()
  fetchStats()
  fetchMemChart()
  fetchLoadChart()
  fetchCpuChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  memChart?.dispose()
  loadChart?.dispose()
  cpuChart?.dispose()
})
</script>

<style lang="scss" scoped>
.home-page {
  .stat-cards {
    margin-bottom: 20px;
  }

  .stat-card {
    display: flex;
    align-items: center;
    gap: 15px;

    .stat-icon {
      width: 50px;
      height: 50px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
    }

    .stat-info {
      .stat-value {
        font-size: 24px;
        font-weight: bold;
        color: #333;
      }

      .stat-label {
        font-size: 14px;
        color: #999;
      }
    }
  }

  .charts-row {
    margin-bottom: 20px;
  }

  .chart-container {
    height: 250px;
  }
}
</style>
