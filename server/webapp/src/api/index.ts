import axios, { AxiosInstance, AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

const BASE_URL = import.meta.env.DEV ? '' : '/v1'

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 30000
})

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.token = token
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 0 || res.code === 200) {
      return res
    }
    ElMessage.error(res.msg || '请求失败')
    return Promise.reject(new Error(res.msg || '请求失败'))
  },
  (error) => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default service

// 通用响应类型
export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

// 封装 GET 请求
export function get<T = unknown>(url: string, params?: object, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return service.get(url, { params, ...config })
}

// 封装 POST 请求
export function post<T = unknown>(url: string, data?: object, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return service.post(url, data, config)
}

// 封装 PUT 请求
export function put<T = unknown>(url: string, data?: object, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return service.put(url, data, config)
}

// 封装 DELETE 请求
export function del<T = unknown>(url: string, params?: object, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return service.delete(url, { params, ...config })
}
