import http from './http'

export const getWorkflows = (params) => http.get('/workflows', { params })
export const getWorkflow = (id) => http.get(`/workflows/${id}`)
export const createWorkflow = (data) => http.post('/workflows', data)
export const updateWorkflow = (id, data) => http.put(`/workflows/${id}`, data)
export const deleteWorkflow = (id) => http.delete(`/workflows/${id}`)

export const getBizTypes = () => http.get('/biz-types')
export const createBizType = (data) => http.post('/biz-types', data)
export const deleteBizType = (id) => http.delete(`/biz-types/${id}`)
