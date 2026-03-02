import http from './http'

export const getLogs = (params) => http.get('/logs', { params })
export const getLogModules = () => http.get('/logs/modules')
