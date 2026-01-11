import { get, post, del, ApiResponse } from '.'
import type { RelationResponse } from '@/types/model'

export function getRelations(params: { cid: number }): Promise<ApiResponse<RelationResponse>> {
  return get('/relation', params)
}

export function addRelation(data: { cid: number; tid: number; next_tid: number }): Promise<ApiResponse<{ rid: number }>> {
  return post('/relation', data)
}

export function deleteRelation(rid: number): Promise<ApiResponse> {
  return del(`/relation/${rid}`)
}
