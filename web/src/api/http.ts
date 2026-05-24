import axios from 'axios'

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 60000
})

http.interceptors.request.use(cfg => {
  const t = localStorage.getItem('token')
  if (t) cfg.headers.Authorization = `Bearer ${t}`
  return cfg
})

http.interceptors.response.use(
  resp => resp,
  err => {
    if (err?.response?.status === 401) {
      localStorage.removeItem('token')
      if (location.pathname !== '/login') location.replace('/login')
    }
    return Promise.reject(err)
  }
)
