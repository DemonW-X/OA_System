import http from './http'

export const getPositions = (params) => http.get('/positions', { params })
export const getPosition = (id) => http.get(`/positions/${id}`)
export const createPosition = (data) => http.post('/positions', data)
export const updatePosition = (id, data) => http.put(`/positions/${id}`, data)
export const deletePosition = (id) => http.delete(`/positions/${id}`)
export const getPositionMenuPermissions = (id) => http.get(`/positions/${id}/menu-permissions`)
export const setPositionMenuPermissions = (id, data) => http.put(`/positions/${id}/menu-permissions`, data)
