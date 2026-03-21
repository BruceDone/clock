<template>
  <div class="home-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-text">
          <h1 class="page-title">仪表盘</h1>
          <p class="page-subtitle">实时监控任务调度状态和系统资源</p>
        </div>
        <div class="header-actions">
          <div class="time-display">
            <span class="time-label">当前时间</span>
            <span class="time-value">{{ currentTime }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-section">
      <div class="stats-grid">
        <div 
          v-for="(item, index) in stats" 
          :key="item.title"
          class="stat-card"
          :class="[`stat-${index}`]"
          @mouseenter="onCardHover(index)" 
          @mouseleave="onCardLeave(index)"
        >
          <div class="stat-bg-gradient"></div>
          <div class="stat-content">
            <div class="stat-icon-wrapper">
              <div class="stat-icon-bg"></div>
              <el-icon :size="24" class="stat-icon"><component :is="item.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedCounts[index] }}</div>
              <div class="stat-label">{{ item.title }}</div>
            </div>
          </div>
          <div class="stat-trend" v-if="index < 2">
            <el-icon><Top /></el-icon>
            <span>+12%</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-section">
      <div class="charts-grid">
        <!-- 内存使用 -->
        <div class="chart-card">
          <div class="chart-header">
            <div class="chart-title">
              <div class="chart-icon">
                <el-icon><PieChart /></el-icon>
              </div>
              <div class="chart-info">
                <span class="chart-name">内存使用</span>
                <span class="chart-desc">实时内存占用</span>
              </div>
            </div>
            <div class="chart-badge">
              <span class="pulse-dot"></span>
              实时
            </div>
          </div>
          <div ref="memChartRef" class="chart-container"></div>
        </div>

        <!-- 系统负载 -->
        <div class="chart-card chart-wide">
          <div class="chart-header">
            <div class="chart-title">
              <div class="chart-icon">
                <el-icon><TrendCharts /></el-icon>
              </div>
              <div class="chart-info">
                <span class="chart-name">系统负载</span>
                <span class="chart-desc">CPU 负载趋势</span>
              </div>
            </div>
            <div class="chart-badge">
              <span class="pulse-dot"></span>
              实时
            </div>
          </div>
          <div ref="loadChartRef" class="chart-container"></div>
        </div>
      </div>

      <!-- CPU 使用率 -->
      <div class="chart-card chart-full">
        <div class="chart-header">
          <div class="chart-title">
            <div class="chart-icon">
              <el-icon><Odometer /></el-icon>
            </div>
            <div class="chart-info">
              <span class="chart-name">CPU 使用率</span>
              <span class="chart-desc">处理器占用情况</span>
            </div>
          </div>
          <div class="chart-badge">
            <span class="pulse-dot"></span>
            实时
          </div>
        </div>
        <div ref="cpuChartRef" class="chart-container"></div>
      </div>
    </div>
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
let clockTimer: number | null = null

const currentTime = ref('')
const animatedCounts = ref([0, 0, 0, 0])
const targetCounts = ref([0, 0, 0, 0])
let animationTimer: number | null = null

function updateTime() {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('zh-CN', { 
    hour: '2-digit', 
    minute: '2-digit', 
    second: '2-digit',
    hour12: false 
  })
}

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
        'pending': 'Clock',
        'running': 'Loading',
        'success': 'CircleCheckFilled',
        'failure': 'CircleCloseFilled'
      }

      const newStats = res.data.map((item: any) => ({
        title: item.title,
        icon: iconMap[item.icon] || 'InfoFilled',
        count: item.count,
        color: item.color || '#909399'
      }))

      stats.value = newStats
      // 确保 animatedCounts 和 targetCounts 数组长度与 newStats 一致
      if (animatedCounts.value.length !== newStats.length) {
        animatedCounts.value = new Array(newStats.length).fill(0)
      }
      targetCounts.value = newStats.map(s => s.count)
      animateCounts()
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

// ECharts 主题自适应配置
function getChartOption(type: 'pie' | 'bar', data: any, options: any = {}) {
  const isDark = document.documentElement.getAttribute('data-theme') !== 'light'
  const primaryColor = isDark ? '#8b5cf6' : '#409eff'
  const textColor = isDark ? '#a1a1aa' : '#606266'
  const bgColor = isDark ? '#111113' : '#ffffff'
  const borderColor = isDark ? 'rgba(139, 92, 246, 0.3)' : 'rgba(64, 158, 255, 0.3)'
  const colors = isDark 
    ? ['#8b5cf6', '#3b82f6', '#22c55e'] 
    : ['#409eff', '#67c23a', '#e6a23c']

  if (type === 'pie') {
    return {
      tooltip: {
        trigger: 'item',
        backgroundColor: bgColor,
        borderColor: borderColor,
        textStyle: { color: textColor }
      },
      legend: {
        bottom: 0,
        textStyle: { color: textColor }
      },
      series: [{
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '45%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 8,
          borderColor: bgColor,
          borderWidth: 2
        },
        label: { show: false },
        emphasis: {
          label: { show: false },
          itemStyle: {
            shadowBlur: 20,
            shadowColor: primaryColor + '60'
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
        backgroundColor: bgColor,
        borderColor: borderColor,
        textStyle: { color: textColor }
      },
      legend: {
        data: options.legendData || [],
        textStyle: { color: textColor }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '12%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: options.xData || [],
        axisLine: { lineStyle: { color: borderColor } },
        axisLabel: { color: textColor }
      },
      yAxis: {
        type: 'value',
        max: options.yMax || 'dataMax',
        axisLine: { lineStyle: { color: borderColor } },
        axisLabel: { color: textColor },
        splitLine: { lineStyle: { color: borderColor + '40' } }
      },
      series: (options.seriesData || []).map((s: any, i: number) => ({
        ...s,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: colors[i % colors.length] },
            { offset: 1, color: colors[i % colors.length] + '60' }
          ]),
          borderRadius: [6, 6, 0, 0]
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowColor: primaryColor + '40'
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
      const isDark = document.documentElement.getAttribute('data-theme') !== 'light'
      const usedColor = isDark ? '#8b5cf6' : '#409eff'
      const freeColor = isDark ? '#1f1f23' : '#f0f2f5'
      const option = getChartOption('pie', [
        { name: '已使用', value: usedPercent, itemStyle: { color: usedColor } },
        { name: '空闲', value: freePercent, itemStyle: { color: freeColor } }
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
  updateTime()
  initCharts()
  await loadData()
  refreshTimer = window.setInterval(() => {
    loadData()
  }, 5000)
  clockTimer = window.setInterval(updateTime, 1000)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer)
  }
  if (clockTimer !== null) {
    clearInterval(clockTimer)
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
// Home 页面样式 - Premium 风格

.home-page {
  .page-header {
    margin-bottom: 32px;

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
    }

    .header-text {
      .page-title {
        font-size: 28px;
        font-weight: 700;
        color: var(--text-primary);
        margin-bottom: 8px;
        letter-spacing: -0.5px;
      }

      .page-subtitle {
        font-size: 14px;
        color: var(--text-secondary);
      }
    }

    .header-actions {
      .time-display {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        padding: 12px 20px;
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: var(--border-radius-lg);

        .time-label {
          font-size: 11px;
          color: var(--text-muted);
          text-transform: uppercase;
          letter-spacing: 1px;
          margin-bottom: 4px;
        }

        .time-value {
          font-size: 20px;
          font-weight: 600;
          font-family: var(--font-family-mono);
          color: var(--primary-color);
          letter-spacing: 2px;
        }
      }
    }
  }
}

.stats-section {
  margin-bottom: 32px;

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 20px;

    @media (max-width: 1200px) {
      grid-template-columns: repeat(2, 1fr);
    }

    @media (max-width: 768px) {
      grid-template-columns: 1fr;
    }
  }

  .stat-card {
    position: relative;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius-xl);
    padding: 24px;
    cursor: pointer;
    overflow: hidden;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

    &:hover {
      transform: translateY(-4px);
      border-color: var(--primary-glow-strong);
      box-shadow: 0 20px 40px var(--shadow-color), 0 0 0 1px var(--border-color-light);
    }

    .stat-bg-gradient {
      position: absolute;
      top: 0;
      right: 0;
      width: 120px;
      height: 120px;
      background: radial-gradient(circle at center, var(--primary-glow) 0%, transparent 70%);
      opacity: 0.5;
      transition: opacity 0.3s;
    }

    &:hover .stat-bg-gradient {
      opacity: 1;
    }

    .stat-content {
      position: relative;
      display: flex;
      align-items: flex-start;
      gap: 16px;

      .stat-icon-wrapper {
        position: relative;
        width: 48px;
        height: 48px;
        flex-shrink: 0;

        .stat-icon-bg {
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          border-radius: 14px;
          background: var(--primary-glow);
          border: 1px solid var(--border-color-light);
          transition: all 0.3s;
        }

        .stat-icon {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          color: var(--primary-color);
        }
      }

      &:hover .stat-icon-bg {
        transform: scale(1.05);
      }
    }

    .stat-info {
      flex: 1;

      .stat-value {
        font-size: 32px;
        font-weight: 700;
        color: var(--text-primary);
        font-family: var(--font-family-mono);
        line-height: 1.2;
        letter-spacing: -1px;
      }

      .stat-label {
        font-size: 13px;
        color: var(--text-secondary);
        margin-top: 6px;
        font-weight: 500;
      }
    }

    .stat-trend {
      position: absolute;
      top: 24px;
      right: 24px;
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 4px 8px;
      background: rgba(34, 197, 94, 0.1);
      border-radius: 8px;
      font-size: 12px;
      color: var(--success-color);
      font-weight: 600;

      .el-icon {
        font-size: 12px;
      }
    }

    &.stat-0 {
      .stat-icon-bg { background: rgba(144, 147, 153, 0.15); }
      .stat-icon { color: #909399; }
      .stat-trend { display: none; }
    }

    &.stat-1 {
      .stat-icon-bg { background: rgba(59, 130, 246, 0.15); }
      .stat-icon { color: var(--info-color); }
      .stat-trend { background: rgba(59, 130, 246, 0.1); color: var(--info-color); }
    }

    &.stat-2 {
      .stat-icon-bg { background: rgba(34, 197, 94, 0.15); }
      .stat-icon { color: var(--success-color); }
    }

    &.stat-3 {
      .stat-icon-bg { background: rgba(239, 68, 68, 0.15); }
      .stat-icon { color: var(--danger-color); }
      .stat-trend { background: rgba(239, 68, 68, 0.1); color: var(--danger-color); }
    }
  }
}

.charts-section {
  .charts-grid {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: 20px;
    margin-bottom: 20px;

    @media (max-width: 1024px) {
      grid-template-columns: 1fr;
    }
  }

  .chart-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: var(--border-radius-xl);
    padding: 24px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);

    &:hover {
      border-color: var(--border-color-light);
      box-shadow: 0 8px 32px var(--shadow-color);
    }

    .chart-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 20px;

      .chart-title {
        display: flex;
        align-items: center;
        gap: 12px;

        .chart-icon {
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--primary-glow);
          border-radius: 10px;

          .el-icon {
            font-size: 20px;
            color: var(--primary-color);
          }
        }

        .chart-info {
          display: flex;
          flex-direction: column;

          .chart-name {
            font-size: 16px;
            font-weight: 600;
            color: var(--text-primary);
          }

          .chart-desc {
            font-size: 12px;
            color: var(--text-muted);
            margin-top: 2px;
          }
        }
      }

      .chart-badge {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 6px 12px;
        background: rgba(34, 197, 94, 0.1);
        border: 1px solid rgba(34, 197, 94, 0.2);
        border-radius: 20px;
        font-size: 12px;
        color: var(--success-color);
        font-weight: 500;

        .pulse-dot {
          width: 6px;
          height: 6px;
          background: var(--success-color);
          border-radius: 50%;
          animation: pulse 2s ease-in-out infinite;
        }
      }
    }

    .chart-container {
      height: 260px;
    }

    &.chart-wide {
      .chart-container {
        height: 220px;
      }
    }

    &.chart-full {
      .chart-container {
        height: 200px;
      }
    }
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(1.2);
  }
}
</style>
