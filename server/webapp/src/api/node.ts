import { put, ApiResponse } from '.'

export function putNodes(data: Array<{ id: number; name: string; status: number; x: number; y: number }>): Promise<ApiResponse> {
  return put('/node', data)
}
