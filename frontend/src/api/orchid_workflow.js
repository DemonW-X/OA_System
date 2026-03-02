import http from './http'

export const getOrchidWorkflows = (params) => http.get('/orchid-workflows', { params })
export const getOrchidWorkflowDefinition = (id) => http.get(`/orchid-workflows/${id}`)
export const createOrchidWorkflow = (data) => http.post('/orchid-workflows', data)
export const updateOrchidWorkflow = (id, data) => http.put(`/orchid-workflows/${id}`, data)
export const deleteOrchidWorkflow = (id) => http.delete(`/orchid-workflows/${id}`)

export const getOrchidWorkflowHistories = (params) => http.get('/orchid-workflow-histories', { params })
export const getMyPendingApprovals = (params) => http.get('/approvals/pending', { params })
export const transferOrchidWorkflowTask = (params, data) => http.post('/orchid-workflow-transfer', data, { params })
export const skipOrchidWorkflowNode = (params, data) => http.post('/orchid-workflow-skip', data, { params })
export const seedOrchidWorkflowTemplates = () => http.post('/orchid-workflow-seed')
