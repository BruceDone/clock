import { get, put, del, ApiResponse } from '.'
import type { Container, ListResponse } from '@/types/model'

export function getContainers(params?: { count?: number; index?: number }): Promise<ApiResponse<ListResponse<Container>>> {
  return get('/container', params)
}

export function getContainer(cid: number): Promise<ApiResponse<Container>> {
  return get(`/container/${cid}`)
}

export function putContainer(data: Partial<Container> & { cid?: number }): Promise<ApiResponse<{ cid: number }>> {
  return put('/container', data)
}

export function deleteContainer(cid: number): Promise<ApiResponse> {
  return del(`/container/${cid}`)
}

export function runContainer(cid: number): Promise<ApiResponse> {
  return get('/container/run', { cid })
}
