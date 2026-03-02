import http from './http'

export const getLeaveRequests = (params) => http.get('/leave-requests', { params })
export const getLeaveRequest = (id) => http.get(`/leave-requests/${id}`)
export const createLeaveRequest = (data) => http.post('/leave-requests', data)
export const updateLeaveRequest = (id, data) => http.put(`/leave-requests/${id}`, data)
export const submitLeaveRequest = (id, data) => http.put(`/leave-requests/${id}/submit`, data)
export const approveLeaveRequest = (id, data) => http.put(`/leave-requests/${id}/approve`, data)
export const deleteLeaveRequest = (id) => http.delete(`/leave-requests/${id}`)
