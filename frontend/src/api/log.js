import http from './http'

export const getLogs = (params) => http.get('/logs', { params })
