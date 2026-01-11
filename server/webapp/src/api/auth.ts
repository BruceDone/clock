import { post, ApiResponse } from '.'
import type { LoginRequest, LoginResponse } from '@/types/model'

export function login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
  return post('/login', data)
}

export function logout(): Promise<ApiResponse> {
  return post('/logout')
}
