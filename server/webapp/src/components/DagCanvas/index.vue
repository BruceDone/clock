<template>
  <div class="dag-canvas-container" ref="containerRef">
    <canvas ref="canvasRef" @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp"
      @mouseleave="onMouseUp" @contextmenu.prevent="onContextMenu" @dblclick="onDoubleClick"></canvas>
    <!-- 右键菜单 -->
    <div v-if="contextMenu.visible" class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
      <div class="menu-item" @click="handleDeleteEdge" v-if="contextMenu.type === 'edge'">删除连线</div>
      <div class="menu-item" @click="handleDeleteNode" v-if="contextMenu.type === 'node'">删除任务</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, watch, nextTick } from 'vue'

interface DagNode {
  id: number
  name: string
  status: number
  x: number
  y: number
}

interface DagEdge {
  id: number
  source: number
  target: number
}

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  radius: number
  alpha: number
  color: string
}

const props = defineProps<{
  nodes: DagNode[]
  edges: DagEdge[]
}>()

const emit = defineEmits<{
  (e: 'node-click', node: DagNode): void
  (e: 'node-move', node: DagNode): void
  (e: 'edge-create', source: number, target: number): void
  (e: 'edge-delete', edgeId: number): void
  (e: 'node-delete', nodeId: number): void
}>()

const containerRef = ref<HTMLElement>()
const canvasRef = ref<HTMLCanvasElement>()
let ctx: CanvasRenderingContext2D | null = null
let animationId: number | null = null

// 画布状态
const NODE_RADIUS = 35
const canvasState = reactive({
  width: 0,
  height: 0,
  scale: 1,
  offsetX: 0,
  offsetY: 0
})

// 交互状态
const interaction = reactive({
  dragging: false,
  dragNode: null as DagNode | null,
  dragStartX: 0,
  dragStartY: 0,
  connecting: false,
  connectSource: null as DagNode | null,
  connectTargetX: 0,
  connectTargetY: 0,
  hoveredNode: null as DagNode | null,
  hoveredEdge: null as DagEdge | null
})

// 右键菜单
const contextMenu = reactive({
  visible: false,
  x: 0,
  y: 0,
  type: '' as 'node' | 'edge' | '',
  target: null as DagNode | DagEdge | null
})

// 粒子系统
let particles: Particle[] = []
const PARTICLE_COUNT = 50

// 状态颜色映射
function getStatusColor(status: number): string {
  const colors: Record<number, string> = {
    1: '#909399', // pending - 灰色
    2: '#409EFF', // running - 蓝色
    3: '#67C23A', // success - 绿色
    4: '#F56C6C'  // failure - 红色
  }
  return colors[status] || '#909399'
}

// 初始化粒子
function initParticles() {
  particles = []
  for (let i = 0; i < PARTICLE_COUNT; i++) {
    particles.push(createParticle())
  }
}

function createParticle(): Particle {
  return {
    x: Math.random() * canvasState.width,
    y: Math.random() * canvasState.height,
    vx: (Math.random() - 0.5) * 0.5,
    vy: (Math.random() - 0.5) * 0.5,
    radius: Math.random() * 2 + 1,
    alpha: Math.random() * 0.5 + 0.1,
    color: `hsla(${210 + Math.random() * 30}, 70%, 60%, `
  }
}

// 更新粒子位置
function updateParticles() {
  particles.forEach(p => {
    p.x += p.vx
    p.y += p.vy

    // 边界反弹
    if (p.x < 0 || p.x > canvasState.width) p.vx *= -1
    if (p.y < 0 || p.y > canvasState.height) p.vy *= -1

    // 保持在边界内
    p.x = Math.max(0, Math.min(canvasState.width, p.x))
    p.y = Math.max(0, Math.min(canvasState.height, p.y))
  })
}

// 绘制粒子
function drawParticles() {
  if (!ctx) return
  particles.forEach(p => {
    ctx!.beginPath()
    ctx!.arc(p.x, p.y, p.radius, 0, Math.PI * 2)
    ctx!.fillStyle = p.color + p.alpha + ')'
    ctx!.fill()
  })

  // 绘制粒子之间的连线
  particles.forEach((p1, i) => {
    particles.slice(i + 1).forEach(p2 => {
      const dx = p1.x - p2.x
      const dy = p1.y - p2.y
      const dist = Math.sqrt(dx * dx + dy * dy)
      if (dist < 100) {
        ctx!.beginPath()
        ctx!.moveTo(p1.x, p1.y)
        ctx!.lineTo(p2.x, p2.y)
        ctx!.strokeStyle = `rgba(64, 158, 255, ${0.1 * (1 - dist / 100)})`
        ctx!.lineWidth = 0.5
        ctx!.stroke()
      }
    })
  })
}

// 绘制箭头
function drawArrow(fromX: number, fromY: number, toX: number, toY: number, color: string) {
  if (!ctx) return

  const angle = Math.atan2(toY - fromY, toX - fromX)
  const arrowLength = 12
  const arrowAngle = Math.PI / 6

  // 计算箭头终点（在节点边缘）
  const endX = toX - NODE_RADIUS * Math.cos(angle)
  const endY = toY - NODE_RADIUS * Math.sin(angle)

  // 计算箭头起点（从源节点边缘）
  const startX = fromX + NODE_RADIUS * Math.cos(angle)
  const startY = fromY + NODE_RADIUS * Math.sin(angle)

  // 绘制线条
  ctx.beginPath()
  ctx.moveTo(startX, startY)
  ctx.lineTo(endX, endY)
  ctx.strokeStyle = color
  ctx.lineWidth = 2
  ctx.stroke()

  // 绘制箭头
  ctx.beginPath()
  ctx.moveTo(endX, endY)
  ctx.lineTo(
    endX - arrowLength * Math.cos(angle - arrowAngle),
    endY - arrowLength * Math.sin(angle - arrowAngle)
  )
  ctx.lineTo(
    endX - arrowLength * Math.cos(angle + arrowAngle),
    endY - arrowLength * Math.sin(angle + arrowAngle)
  )
  ctx.closePath()
  ctx.fillStyle = color
  ctx.fill()
}

// 绘制节点
function drawNode(node: DagNode, isHovered: boolean = false) {
  if (!ctx) return

  const color = getStatusColor(node.status)

  // 绘制光晕效果
  if (isHovered) {
    const gradient = ctx.createRadialGradient(node.x, node.y, NODE_RADIUS, node.x, node.y, NODE_RADIUS + 15)
    gradient.addColorStop(0, color + '40')
    gradient.addColorStop(1, 'transparent')
    ctx.beginPath()
    ctx.arc(node.x, node.y, NODE_RADIUS + 15, 0, Math.PI * 2)
    ctx.fillStyle = gradient
    ctx.fill()
  }

  // 绘制节点圆形
  ctx.beginPath()
  ctx.arc(node.x, node.y, NODE_RADIUS, 0, Math.PI * 2)

  // 渐变填充
  const gradient = ctx.createRadialGradient(
    node.x - 10, node.y - 10, 0,
    node.x, node.y, NODE_RADIUS
  )
  gradient.addColorStop(0, color)
  gradient.addColorStop(1, adjustColor(color, -30))
  ctx.fillStyle = gradient
  ctx.fill()

  // 边框
  ctx.strokeStyle = isHovered ? '#fff' : adjustColor(color, -20)
  ctx.lineWidth = isHovered ? 3 : 2
  ctx.stroke()

  // 绘制文字
  ctx.fillStyle = '#fff'
  ctx.font = 'bold 12px Arial'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'

  // 文字截断
  const maxWidth = NODE_RADIUS * 1.6
  let text = node.name
  if (ctx.measureText(text).width > maxWidth) {
    while (ctx.measureText(text + '...').width > maxWidth && text.length > 0) {
      text = text.slice(0, -1)
    }
    text += '...'
  }
  ctx.fillText(text, node.x, node.y)
}

// 调整颜色明暗
function adjustColor(color: string, amount: number): string {
  const hex = color.replace('#', '')
  const r = Math.max(0, Math.min(255, parseInt(hex.slice(0, 2), 16) + amount))
  const g = Math.max(0, Math.min(255, parseInt(hex.slice(2, 4), 16) + amount))
  const b = Math.max(0, Math.min(255, parseInt(hex.slice(4, 6), 16) + amount))
  return `rgb(${r}, ${g}, ${b})`
}

// 绘制连线
function drawEdges() {
  if (!ctx) return

  props.edges.forEach(edge => {
    const sourceNode = props.nodes.find(n => n.id === edge.source)
    const targetNode = props.nodes.find(n => n.id === edge.target)
    if (!sourceNode || !targetNode) return

    const isHovered = interaction.hoveredEdge?.id === edge.id
    const color = isHovered ? '#409EFF' : '#A3B1BF'
    drawArrow(sourceNode.x, sourceNode.y, targetNode.x, targetNode.y, color)
  })
}

// 绘制正在创建的连线
function drawConnectingLine() {
  if (!ctx || !interaction.connecting || !interaction.connectSource) return

  ctx.beginPath()
  ctx.moveTo(interaction.connectSource.x, interaction.connectSource.y)
  ctx.lineTo(interaction.connectTargetX, interaction.connectTargetY)
  ctx.strokeStyle = '#409EFF'
  ctx.lineWidth = 2
  ctx.setLineDash([5, 5])
  ctx.stroke()
  ctx.setLineDash([])
}

// 主绘制函数
function draw() {
  if (!ctx || !canvasRef.value) return

  // 清空画布
  ctx.clearRect(0, 0, canvasState.width, canvasState.height)

  // 绘制背景
  ctx.fillStyle = '#fafafa'
  ctx.fillRect(0, 0, canvasState.width, canvasState.height)

  // 绘制网格
  drawGrid()

  // 绘制粒子
  updateParticles()
  drawParticles()

  // 绘制连线
  drawEdges()

  // 绘制正在创建的连线
  drawConnectingLine()

  // 绘制节点
  props.nodes.forEach(node => {
    const isHovered = interaction.hoveredNode?.id === node.id
    drawNode(node, isHovered)
  })

  animationId = requestAnimationFrame(draw)
}

// 绘制网格
function drawGrid() {
  if (!ctx) return

  const gridSize = 30
  ctx.strokeStyle = '#eee'
  ctx.lineWidth = 0.5

  for (let x = 0; x <= canvasState.width; x += gridSize) {
    ctx.beginPath()
    ctx.moveTo(x, 0)
    ctx.lineTo(x, canvasState.height)
    ctx.stroke()
  }

  for (let y = 0; y <= canvasState.height; y += gridSize) {
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(canvasState.width, y)
    ctx.stroke()
  }
}

// 获取鼠标位置的节点
function getNodeAtPosition(x: number, y: number): DagNode | null {
  for (const node of props.nodes) {
    const dx = x - node.x
    const dy = y - node.y
    if (dx * dx + dy * dy <= NODE_RADIUS * NODE_RADIUS) {
      return node
    }
  }
  return null
}

// 获取鼠标位置的边
function getEdgeAtPosition(x: number, y: number): DagEdge | null {
  const threshold = 8

  for (const edge of props.edges) {
    const sourceNode = props.nodes.find(n => n.id === edge.source)
    const targetNode = props.nodes.find(n => n.id === edge.target)
    if (!sourceNode || !targetNode) continue

    // 计算点到线段的距离
    const dist = pointToLineDistance(
      x, y,
      sourceNode.x, sourceNode.y,
      targetNode.x, targetNode.y
    )
    if (dist < threshold) return edge
  }
  return null
}

// 点到线段的距离
function pointToLineDistance(px: number, py: number, x1: number, y1: number, x2: number, y2: number): number {
  const A = px - x1
  const B = py - y1
  const C = x2 - x1
  const D = y2 - y1

  const dot = A * C + B * D
  const lenSq = C * C + D * D
  let param = -1

  if (lenSq !== 0) param = dot / lenSq

  let xx, yy

  if (param < 0) {
    xx = x1
    yy = y1
  } else if (param > 1) {
    xx = x2
    yy = y2
  } else {
    xx = x1 + param * C
    yy = y1 + param * D
  }

  const dx = px - xx
  const dy = py - yy
  return Math.sqrt(dx * dx + dy * dy)
}

// 获取画布坐标
function getCanvasPosition(e: MouseEvent): { x: number; y: number } {
  const rect = canvasRef.value!.getBoundingClientRect()
  return {
    x: e.clientX - rect.left,
    y: e.clientY - rect.top
  }
}

// 鼠标事件处理
function onMouseDown(e: MouseEvent) {
  if (e.button !== 0) return // 只处理左键

  contextMenu.visible = false
  const pos = getCanvasPosition(e)
  const node = getNodeAtPosition(pos.x, pos.y)

  if (node) {
    if (e.shiftKey) {
      // Shift + 拖拽开始连线
      interaction.connecting = true
      interaction.connectSource = node
      interaction.connectTargetX = pos.x
      interaction.connectTargetY = pos.y
    } else {
      // 普通拖拽节点
      interaction.dragging = true
      interaction.dragNode = node
      interaction.dragStartX = pos.x - node.x
      interaction.dragStartY = pos.y - node.y
    }
  }
}

function onMouseMove(e: MouseEvent) {
  const pos = getCanvasPosition(e)

  if (interaction.dragging && interaction.dragNode) {
    // 拖拽节点
    interaction.dragNode.x = pos.x - interaction.dragStartX
    interaction.dragNode.y = pos.y - interaction.dragStartY
  } else if (interaction.connecting) {
    // 更新连线终点
    interaction.connectTargetX = pos.x
    interaction.connectTargetY = pos.y
  } else {
    // 更新悬停状态
    interaction.hoveredNode = getNodeAtPosition(pos.x, pos.y)
    interaction.hoveredEdge = interaction.hoveredNode ? null : getEdgeAtPosition(pos.x, pos.y)

    // 更新鼠标样式
    if (canvasRef.value) {
      if (interaction.hoveredNode) {
        canvasRef.value.style.cursor = e.shiftKey ? 'crosshair' : 'move'
      } else if (interaction.hoveredEdge) {
        canvasRef.value.style.cursor = 'pointer'
      } else {
        canvasRef.value.style.cursor = 'default'
      }
    }
  }
}

function onMouseUp(e: MouseEvent) {
  if (interaction.dragging && interaction.dragNode) {
    emit('node-move', { ...interaction.dragNode })
  }

  if (interaction.connecting && interaction.connectSource) {
    const pos = getCanvasPosition(e)
    const targetNode = getNodeAtPosition(pos.x, pos.y)

    if (targetNode && targetNode.id !== interaction.connectSource.id) {
      emit('edge-create', interaction.connectSource.id, targetNode.id)
    }
  }

  interaction.dragging = false
  interaction.dragNode = null
  interaction.connecting = false
  interaction.connectSource = null
}

function onContextMenu(e: MouseEvent) {
  const pos = getCanvasPosition(e)
  const node = getNodeAtPosition(pos.x, pos.y)
  const edge = getEdgeAtPosition(pos.x, pos.y)

  if (node) {
    contextMenu.visible = true
    contextMenu.x = e.offsetX
    contextMenu.y = e.offsetY
    contextMenu.type = 'node'
    contextMenu.target = node
  } else if (edge) {
    contextMenu.visible = true
    contextMenu.x = e.offsetX
    contextMenu.y = e.offsetY
    contextMenu.type = 'edge'
    contextMenu.target = edge
  } else {
    contextMenu.visible = false
  }
}

function onDoubleClick(e: MouseEvent) {
  const pos = getCanvasPosition(e)
  const node = getNodeAtPosition(pos.x, pos.y)
  if (node) {
    emit('node-click', node)
  }
}

function handleDeleteEdge() {
  if (contextMenu.target && contextMenu.type === 'edge') {
    emit('edge-delete', (contextMenu.target as DagEdge).id)
  }
  contextMenu.visible = false
}

function handleDeleteNode() {
  if (contextMenu.target && contextMenu.type === 'node') {
    emit('node-delete', (contextMenu.target as DagNode).id)
  }
  contextMenu.visible = false
}

// 初始化画布
function initCanvas() {
  if (!containerRef.value || !canvasRef.value) return

  const rect = containerRef.value.getBoundingClientRect()
  canvasState.width = rect.width
  canvasState.height = rect.height || 500

  canvasRef.value.width = canvasState.width
  canvasRef.value.height = canvasState.height

  ctx = canvasRef.value.getContext('2d')

  initParticles()
  draw()
}

// 适应视图
function fitView() {
  if (props.nodes.length === 0) return

  const padding = 50
  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity

  props.nodes.forEach(node => {
    minX = Math.min(minX, node.x)
    minY = Math.min(minY, node.y)
    maxX = Math.max(maxX, node.x)
    maxY = Math.max(maxY, node.y)
  })

  // 如果节点超出画布，调整位置
  if (minX < padding || minY < padding) {
    const offsetX = Math.max(0, padding - minX + NODE_RADIUS)
    const offsetY = Math.max(0, padding - minY + NODE_RADIUS)
    props.nodes.forEach(node => {
      node.x += offsetX
      node.y += offsetY
    })
  }
}

// 暴露方法
defineExpose({
  fitView
})

// 监听窗口大小变化
function handleResize() {
  nextTick(initCanvas)
}

onMounted(() => {
  initCanvas()
  window.addEventListener('resize', handleResize)

  // 点击其他地方关闭菜单
  document.addEventListener('click', () => {
    contextMenu.visible = false
  })
})

onUnmounted(() => {
  if (animationId) cancelAnimationFrame(animationId)
  window.removeEventListener('resize', handleResize)
})

watch(() => [props.nodes, props.edges], () => {
  // 数据变化时重新绘制
}, { deep: true })
</script>

<style lang="scss" scoped>
.dag-canvas-container {
  position: relative;
  width: 100%;
  height: 100%;

  canvas {
    display: block;
    width: 100%;
    height: 100%;
  }

  .context-menu {
    position: absolute;
    background: #fff;
    border: 1px solid #e4e7ed;
    border-radius: 4px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
    padding: 5px 0;
    z-index: 100;

    .menu-item {
      padding: 8px 16px;
      cursor: pointer;
      font-size: 14px;
      color: #606266;

      &:hover {
        background: #f5f7fa;
        color: #409EFF;
      }
    }
  }
}
</style>
