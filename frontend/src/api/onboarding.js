import http from './http'

export const getOnboardings = (params) => http.get('/onboardings', { params })
export const getOnboarding = (id) => http.get(`/onboardings/${id}`)
export const createOnboarding = (data) => http.post('/onboardings', data)
export const updateOnboarding = (id, data) => http.put(`/onboardings/${id}`, data)
export const deleteOnboarding = (id) => http.delete(`/onboardings/${id}`)
export const submitOnboarding = (id) => http.put(`/onboardings/${id}/submit`, {})
export const withdrawOnboarding = (id) => http.put(`/onboardings/${id}/withdraw`, {})
export const approveOnboarding = (id, data) => http.put(`/onboardings/${id}/approve`, data)
export const cancelApproveOnboarding = (id) => http.put(`/onboardings/${id}/cancel-approve`, {})
