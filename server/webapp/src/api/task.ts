import { get, put, del, ApiResponse } from '.'
import type { Task, ListResponse } from '@/types/model'

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
