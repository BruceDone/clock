// API 参数类型定义
export type GetContainersParams = { count?: number; index?: number }
export type GetTasksParams = { count?: number; index?: number; cid?: number }
export type GetRelationsParams = { cid: number }
export type PutNodesRequest = Array<{ id: number; name: string; status: number; x: number; y: number }>
export type GetLogsParams = { count?: number; index?: number; left_ts?: number; right_ts?: number; tid?: number; cid?: number }
