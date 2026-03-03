import http from './http'

export const getResignations = (params) => http.get('/resignations', { params })
export const getResignation = (id) => http.get(`/resignations/${id}`)
export const createResignation = (data) => http.post('/resignations', data)
export const updateResignation = (id, data) => http.put(`/resignations/${id}`, data)
export const deleteResignation = (id) => http.delete(`/resignations/${id}`)
export const submitResignation = (id, data) => http.put(`/resignations/${id}/submit`, data)
export const withdrawResignation = (id) => http.put(`/resignations/${id}/withdraw`, {})
export const approveResignation = (id, data) => http.put(`/resignations/${id}/approve`, data)
export const cancelApproveResignation = (id) => http.put(`/resignations/${id}/cancel-approve`, {})
