import { get, ApiResponse } from '.'

export function getMessages(): Promise<ApiResponse<Array<{ title: string; icon: string; count: number; color: string }>>> {
  return get('/message')
}
