import http from './http'

export const getDepartments = (params) => http.get('/departments', { params })
export const createDepartment = (data) => http.post('/departments', data)
export const updateDepartment = (id, data) => http.put(`/departments/${id}`, data)
export const deleteDepartment = (id) => http.delete(`/departments/${id}`)
