import { get, ApiResponse } from '.'

export function getSystemLoad(): Promise<ApiResponse<number[]>> {
  return get('/system/load')
}

export function getSystemMem(): Promise<ApiResponse<number>> {
  return get('/system/mem')
}

export function getSystemCpu(): Promise<ApiResponse<number>> {
  return get('/system/cpu')
}
