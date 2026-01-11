import { get, put, del, post, ApiResponse } from '.'
import type { Task, ListResponse } from '@/types/model'

// 运行中任务信息
export interface RunningTaskInfo {
  tid: number
  cid: number
  runId: string
  taskName: string
  startAt: number
}

export function getTasks(params?: { count?: number; index?: number; cid?: number }): Promise<ApiResponse<ListResponse<Task>>> {
  return get('/task', params)
}

export function getTask(tid: number): Promise<ApiResponse<Task>> {
  return get(`/task/${tid}`)
}

export function putTask(data: Partial<Task> & { tid?: number }): Promise<ApiResponse<{ tid: number }>> {
  return put('/task', data)
}

export function deleteTask(tid: number): Promise<ApiResponse> {
  return del(`/task/${tid}`)
}

export function runTask(tid: number): Promise<ApiResponse> {
  return get('/task/run', { tid })
}

// 取消单个任务
export function cancelTask(tid: number): Promise<ApiResponse> {
  return post('/task/cancel', { tid })
}

// 取消整个 run
export function cancelRun(runId: string): Promise<ApiResponse> {
  return post('/run/cancel', { runId })
}

// 获取运行中的任务列表
export function getRunningTasks(): Promise<ApiResponse<RunningTaskInfo[]>> {
  return get('/task/running')
}
