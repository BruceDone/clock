import { get, del, ApiResponse } from '.'
import type { TaskLog, ListResponse } from '@/types/model'

export function getLogs(params?: { count?: number; index?: number; left_ts?: number; right_ts?: number; tid?: number; cid?: number }): Promise<ApiResponse<ListResponse<TaskLog>>> {
  return get('/log', params)
}

export function deleteLogs(params: { left_ts?: number; right_ts?: number; tid?: number; cid?: number }): Promise<ApiResponse> {
  return del('/log', params)
}

export function deleteLogByID(lid: string): Promise<ApiResponse> {
  return del(`/log/${lid}`)
}

export function deleteAllLogs(): Promise<ApiResponse> {
  return del('/log/all')
}
