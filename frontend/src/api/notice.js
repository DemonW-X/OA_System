import http from './http'

export const getNotices = (params) => http.get('/notices', { params })
export const getNotice = (id) => http.get(`/notices/${id}`)
export const createNotice = (data) => http.post('/notices', data)
export const updateNotice = (id, data) => http.put(`/notices/${id}`, data)
export const deleteNotice = (id) => http.delete(`/notices/${id}`)

export const uploadImage = (formData) => http.post('/upload/image', formData, {
  headers: { 'Content-Type': 'multipart/form-data' }
})
export const uploadAttachment = (formData) => http.post('/upload/attachment', formData, {
  headers: { 'Content-Type': 'multipart/form-data' }
})
