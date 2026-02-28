import http from './http'

export const getResignations = (params) => http.get('/resignations', { params })
export const getResignation = (id) => http.get(`/resignations/${id}`)
export const createResignation = (data) => http.post('/resignations', data)
export const updateResignation = (id, data) => http.put(`/resignations/${id}`, data)
export const deleteResignation = (id) => http.delete(`/resignations/${id}`)
