import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

let isHandlingAuthError = false

const http = axios.create({
  baseURL: '/api',
  timeout: 5000
})

// 请求拦截：自动带上 token
http.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  const isLoginRequest = typeof config.url === 'string' && config.url.includes('/login')
  if (token && !isLoginRequest) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：统一处理 401
http.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401 && !err.config.url.includes('/login')) {
      if (!isHandlingAuthError) {
        isHandlingAuthError = true
        localStorage.removeItem('token')
        localStorage.removeItem('userInfo')
        router.push('/login')
        ElMessage.error('登录已过期，请重新登录')
        setTimeout(() => {
          isHandlingAuthError = false
        }, 800)
      }
      err.__authExpired = true
    } else {
      ElMessage.error(err.response?.data?.msg || '请求失败')
    }
    return Promise.reject(err)
  }
)

export default http
