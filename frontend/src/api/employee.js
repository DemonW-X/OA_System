import http from './http'

export const getEmployees = (params) => http.get('/employees', { params })
export const getEmployee = (id) => http.get(`/employees/${id}`)
export const createEmployee = (data) => http.post('/employees', data)
export const updateEmployee = (id, data) => http.put(`/employees/${id}`, data)
export const deleteEmployee = (id) => http.delete(`/employees/${id}`)
export const submitEmployee = (id, data) => http.put(`/employees/${id}/submit`, data)
export const withdrawEmployee = (id) => http.put(`/employees/${id}/withdraw`, {})
export const approveEmployee = (id, data) => http.put(`/employees/${id}/approve`, data)
export const cancelApproveEmployee = (id) => http.put(`/employees/${id}/cancel-approve`, {})

export const getEmployeeMenuPermissions = (id) => http.get(`/employees/${id}/menu-permissions`)
export const setEmployeeMenuPermissions = (id, data) => http.put(`/employees/${id}/menu-permissions`, data)
