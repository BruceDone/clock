// 任务状态
export const TaskStatus = {
  PENDING: 1,
  START: 2,
  SUCCESS: 3,
  FAILURE: 4
} as const

export type TaskStatusType = typeof TaskStatus[keyof typeof TaskStatus]

// 容器
export interface Container {
  cid: number
  entry_id: number
  name: string
  expression: string
  status: number
  disable: boolean
  blocking: boolean  // 阻塞模式（上次未完成则跳过）
  update_at: number
}

// 任务
export interface Task {
  tid: number
  cid: number
  command: string
  name: string
  directory: string
  disable: boolean
  status: TaskStatusType
  timeout: number
  update_at: number
  log_enable: boolean
  point_x: number
  point_y: number
}

// 关系
export interface Relation {
  rid: number
  cid: number
  tid: number
  next_tid: number
  update_at: number
}

// 关系图响应
export interface RelationResponse {
  nodes: Node[]
  links: Link[]
}

// 节点 (用于关系图)
export interface Node {
  id: number
  name: string
  status: TaskStatusType
  x: number
  y: number
}

// 连线
export interface Link {
  id: number
  name: string
  cid: number
  tid: number
  next_tid: number
}

// 任务日志
export interface TaskLog {
  lid: string
  tid: number
  cid: number
  std_out: string
  std_err: string
  update_at: number
}

// 统计卡片
export interface TaskCounter {
  title: string
  icon: string
  count: number
  color: string
}

// 分页参数
export interface PageParams {
  count?: number
  index?: number
  total?: number
  order?: string
}

// 分页响应
export interface ListResponse<T> {
  items: T[]
  page: PageParams
}

// API 响应
export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

// 登录请求
export interface LoginRequest {
  user: string
  pwd: string
}

// 登录响应
export interface LoginResponse {
  token: string
}

// 系统负载
export interface SystemLoad {
  load1: number
  load5: number
  load15: number
}

// 内存使用
export interface MemoryUsage {
  name: string
  value: number
}

// CPU 使用
export interface CpuUsage {
  [key: string]: number
}
