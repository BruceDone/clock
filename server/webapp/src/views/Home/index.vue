<template>
  <div class="home-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">系统监控</h1>
      <p class="page-subtitle">实时监控任务调度状态和系统资源</p>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="24" class="stat-cards">
      <el-col :span="6" v-for="(item, index) in stats" :key="item.title">
        <div class="stat-card" :class="[`stat-${index}`]" @mouseenter="onCardHover(index)" @mouseleave="onCardLeave(index)">
          <div class="stat-glow"></div>
          <div class="stat-icon-wrapper">
            <div class="stat-icon-bg"></div>
            <el-icon :size="28" class="stat-icon"><component :is="item.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value" :ref="el => statValues[index] = el">{{ animatedCounts[index] }}</div>
            <div class="stat-label">{{ item.title }}</div>
          </div>
          <div class="stat-decoration">
            <div class="dec-line"></div>
            <div class="dec-dot"></div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="24" class="charts-row">
      <el-col :span="8">
        <el-card class="chart-card">
          <template #header>
            <div class="chart-header">
              <div class="chart-title">
                <el-icon><PieChart /></el-icon>
                <span>内存使用</span>
              </div>
              <div class="chart-badge">实时</div>
            </div>
          </template>
          <div ref="memChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="16">
        <el-card class="chart-card">
          <template #header>
            <div class="chart-header">
              <div class="chart-title">
                <el-icon><TrendCharts /></el-icon>
                <span>系统负载</span>
              </div>
              <div class="chart-badge">实时</div>
            </div>
          </template>
          <div ref="loadChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="24" class="charts-row">
      <el-col :span="24">
        <el-card class="chart-card">
          <template #header>
            <div class="chart-header">
              <div class="chart-title">
                <el-icon><Odometer /></el-icon>
                <span>CPU 使用率</span>
              </div>
              <div class="chart-badge">实时</div>
            </div>
          </template>
          <div ref="cpuChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getMessages } from '@/api/message'
import { getSystemLoad, getSystemMem, getSystemCpu } from '@/api/system'

const memChartRef = ref<HTMLElement>()
const loadChartRef = ref<HTMLElement>()
const cpuChartRef = ref<HTMLElement>()

let memChart: echarts.ECharts | null = null
let loadChart: echarts.ECharts | null = null
let cpuChart: echarts.ECharts | null = null
let refreshTimer: number | null = null

const statValues = ref<Array<any>>([])
const animatedCounts = ref([0, 0, 0, 0])
const targetCounts = ref([0, 0, 0, 0])
let animationTimer: number | null = null

interface StatItem {
  title: string
  icon: string
  count: number
  color: string
}

const stats = ref<StatItem[]>([
  { title: '等待运行', icon: 'Clock', count: 0, color: '#909399' },
  { title: '正在运行', icon: 'Loading', count: 0, color: '#00ccff' },
  { title: '运行成功', icon: 'CircleCheckFilled', count: 0, color: '#00ff88' },
  { title: '运行失败', icon: 'CircleCloseFilled', count: 0, color: '#ff4466' }
])

// 数字动画
function animateCounts() {
  if (animationTimer) cancelAnimationFrame(animationTimer)

  function update() {
    let allDone = true
    animatedCounts.value = animatedCounts.value.map((current, i) => {
      const target = targetCounts.value[i]
      if (Math.abs(target - current) > 0.5) {
        allDone = false
        return current + (target - current) * 0.1
      }
      return target
    })

    if (!allDone) {
      animationTimer = requestAnimationFrame(update)
    }
  }
  update()
}

function onCardHover(index: number) {
  const card = document.querySelector(`.stat-${index}`) as HTMLElement
  if (card) {
    card.style.transform = 'translateY(-5px)'
    card.style.boxShadow = '0 10px 40px rgba(0, 255, 136, 0.2)'
  }
}

function onCardLeave(index: number) {
  const card = document.querySelector(`.stat-${index}`) as HTMLElement
  if (card) {
    card.style.transform = 'translateY(0)'
    card.style.boxShadow = ''
  }
}

async function fetchStats() {
  try {
    const res = await getMessages()
    if (res.data && Array.isArray(res.data)) {
      const iconMap: Record<string, string> = {
        'container': 'Box',
        'task': 'List',
        'running': 'Loading',
        'success': 'CircleCheckFilled',
        'failure': 'CircleCloseFilled',
        'pending': 'Clock'
      }

      const newStats = res.data.map((item: any) => ({
        title: item.title,
        icon: iconMap[item.icon] || 'InfoFilled',
        count: item.count,
        color: item.color || '#909399'
      }))

      stats.value = newStats
      targetCounts.value = newStats.map(s => s.count)
      animateCounts()
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

// ECharts 绿色主题配置
function getChartOption(type: 'pie' | 'bar', data: any, options: any = {}) {
  const colors = ['#00ff88', '#00ccff', '#ffaa00']

  if (type === 'pie') {
    return {
      tooltip: {
        trigger: 'item',
        backgroundColor: 'rgba(15, 23, 35, 0.9)',
        borderColor: '#00ff88',
        textStyle: { color: '#e0e6ed' }
      },
      legend: {
        bottom: 0,
        textStyle: { color: '#8b9bb4' }
      },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['50%', '45%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 8,
          borderColor: 'rgba(15, 23, 35, 0.9)',
          borderWidth: 2
        },
        label: { show: false },
        emphasis: {
          label: { show: false },
          itemStyle: {
            shadowBlur: 20,
            shadowColor: 'rgba(0, 255, 136, 0.5)'
          }
        },
        data: data
      }]
    }
  }

  if (type === 'bar') {
    return {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        backgroundColor: 'rgba(15, 23, 35, 0.9)',
        borderColor: '#00ff88',
        textStyle: { color: '#e0e6ed' }
      },
      legend: {
        data: options.legendData || [],
        textStyle: { color: '#8b9bb4' }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '10%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: options.xData || [],
        axisLine: { lineStyle: { color: 'rgba(0, 255, 136, 0.2)' } },
        axisLabel: { color: '#8b9bb4' }
      },
      yAxis: {
        type: 'value',
        max: options.yMax || 'dataMax',
        axisLine: { lineStyle: { color: 'rgba(0, 255, 136, 0.2)' } },
        axisLabel: { color: '#8b9bb4' },
        splitLine: { lineStyle: { color: 'rgba(0, 255, 136, 0.08)' } }
      },
      series: (options.seriesData || []).map((s: any, i: number) => ({
        ...s,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: colors[i % colors.length] },
            { offset: 1, color: colors[i % colors.length] + '80' }
          ]),
          borderRadius: [4, 4, 0, 0]
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowColor: 'rgba(0, 255, 136, 0.3)'
          }
        }
      }))
    }
  }

  return {}
}

async function fetchMemChart() {
  try {
    const res = await getSystemMem()
    if (res.data !== undefined && res.data !== null) {
      const usedPercent = res.data
      const freePercent = 100 - usedPercent
      const option = getChartOption('pie', [
        { name: '已使用', value: usedPercent, itemStyle: { color: '#00ff88' } },
        { name: '空闲', value: freePercent, itemStyle: { color: '#0f171f' } }
      ])
      memChart?.setOption(option)
    }
  } catch (error) {
    console.error('获取内存数据失败:', error)
  }
}

async function fetchLoadChart() {
  try {
    const res = await getSystemLoad()
    if (res.data && Array.isArray(res.data) && res.data.length >= 3) {
      const option = getChartOption('bar', null, {
        legendData: ['1分钟', '5分钟', '15分钟'],
        xData: ['当前'],
        seriesData: [
          { name: '1分钟', type: 'bar', data: [res.data[0]] },
          { name: '5分钟', type: 'bar', data: [res.data[1]] },
          { name: '15分钟', type: 'bar', data: [res.data[2]] }
        ]
      })
      loadChart?.setOption(option)
    }
  } catch (error) {
    console.error('获取负载数据失败:', error)
  }
}

async function fetchCpuChart() {
  try {
    const res = await getSystemCpu()
    if (res.data !== undefined && res.data !== null) {
      const option = getChartOption('bar', null, {
        legendData: ['CPU 使用率'],
        xData: ['当前'],
        yMax: 100,
        seriesData: [{
          name: 'CPU 使用率',
          type: 'bar',
          data: [res.data]
        }]
      })
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

async function loadData() {
  await fetchStats()
  await fetchMemChart()
  await fetchLoadChart()
  await fetchCpuChart()
}

onMounted(async () => {
  await nextTick()
  initCharts()
  await loadData()
  refreshTimer = window.setInterval(() => {
    loadData()
  }, 5000)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer)
  }
  if (animationTimer !== null) {
    cancelAnimationFrame(animationTimer)
  }
  window.removeEventListener('resize', handleResize)
  memChart?.dispose()
  loadChart?.dispose()
  cpuChart?.dispose()
})
</script>

<style lang="scss" scoped>
// Home 页面样式 - 使用 CSS 变量

.home-page {
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

  .stat-cards {
    margin-bottom: 24px;
  }

  .stat-card {
    position: relative;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius-lg);
    padding: 24px;
    cursor: pointer;
    transition: all var(--transition-base);
    overflow: hidden;
    backdrop-filter: blur(10px);

    &::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 2px;
      background: linear-gradient(90deg, transparent, var(--primary-color), transparent);
      opacity: 0;
      transition: opacity var(--transition-base);
    }

    &:hover::before {
      opacity: 1;
    }

    .stat-glow {
      position: absolute;
      top: -50%;
      right: -50%;
      width: 100%;
      height: 100%;
      background: radial-gradient(circle, rgba(var(--primary-color), 0.1) 0%, transparent 70%);
      opacity: 0;
      transition: opacity var(--transition-base);
    }

    &:hover .stat-glow {
      opacity: 1;
    }

    .stat-icon-wrapper {
      position: relative;
      width: 60px;
      height: 60px;
      margin-bottom: 16px;

      .stat-icon-bg {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        border-radius: 12px;
        background: rgba(var(--primary-color), 0.1);
        border: 1px solid rgba(var(--primary-color), 0.2);
      }

      .stat-icon {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        color: var(--primary-color);
      }
    }

    .stat-info {
      .stat-value {
        font-size: 36px;
        font-weight: 700;
        color: var(--text-primary);
        font-family: var(--font-family-mono);
        line-height: 1.2;
      }

      .stat-label {
        font-size: 14px;
        color: var(--text-secondary);
        margin-top: 4px;
      }
    }

    .stat-decoration {
      position: absolute;
      bottom: 16px;
      right: 16px;
      display: flex;
      align-items: center;
      gap: 8px;

      .dec-line {
        width: 30px;
        height: 2px;
        background: rgba(var(--primary-color), 0.3);
      }

      .dec-dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;
        background: var(--primary-color);
      }
    }

    // 不同状态的颜色变体
    &.stat-1 {
      .stat-icon-bg { background: rgba(var(--info-color), 0.1); border-color: rgba(var(--info-color), 0.2); }
      .stat-icon { color: var(--info-color); }
      .dec-dot { background: var(--info-color); }
    }

    &.stat-2 {
      .stat-icon-bg { background: rgba(var(--success-color), 0.1); border-color: rgba(var(--success-color), 0.2); }
      .stat-icon { color: var(--success-color); }
      .dec-dot { background: var(--success-color); }
    }

    &.stat-3 {
      .stat-icon-bg { background: rgba(var(--danger-color), 0.1); border-color: rgba(var(--danger-color), 0.2); }
      .stat-icon { color: var(--danger-color); }
      .dec-dot { background: var(--danger-color); }
    }
  }

  .charts-row {
    margin-bottom: 24px;
  }

  .chart-card {
    background: var(--bg-card) !important;
    border: 1px solid var(--border-color) !important;
    border-radius: var(--border-radius-lg) !important;
    backdrop-filter: blur(10px);
    transition: all var(--transition-base);

    &:hover {
      border-color: var(--primary-glow) !important;
      box-shadow: var(--shadow-glow);
    }

    :deep(.el-card__header) {
      background: transparent !important;
      border-bottom: 1px solid var(--border-color-light);
      padding: 16px 20px;
    }

    .chart-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .chart-title {
        display: flex;
        align-items: center;
        gap: 8px;
        color: var(--text-primary);
        font-size: 16px;
        font-weight: 600;

        .el-icon {
          color: var(--primary-color);
        }
      }

      .chart-badge {
        padding: 4px 10px;
        background: rgba(var(--success-color), 0.1);
        border: 1px solid rgba(var(--success-color), 0.3);
        border-radius: 12px;
        font-size: 12px;
        color: var(--success-color);
      }
    }

    .chart-container {
      height: 280px;
    }
  }
}
</style>
