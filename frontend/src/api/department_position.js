import http from './http'

export const getDepartmentPositions = (params) => http.get('/department-positions', { params })
export const createDepartmentPosition = (data) => http.post('/department-positions', data)
export const deleteDepartmentPosition = (id) => http.delete(`/department-positions/${id}`)
export const getDepartmentPositionTree = () => http.get('/department-positions/tree')
