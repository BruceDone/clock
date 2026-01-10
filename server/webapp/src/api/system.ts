import { get, ApiResponse } from '.'

export function getSystemLoad(): Promise<ApiResponse<{ load1: number; load5: number; load15: number }>> {
  return get('/system/load')
}

export function getSystemMem(): Promise<ApiResponse<Array<{ name: string; value: number }>>> {
  return get('/system/mem')
}

export function getSystemCpu(): Promise<ApiResponse<Record<string, number>>> {
  return get('/system/cpu')
}
