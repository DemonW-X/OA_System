import http from './http'

export const getCalendarEvents = (params) => http.get('/calendar', { params })
export const createCalendarEvent = (data) => http.post('/calendar', data)
export const updateCalendarEvent = (id, data) => http.put(`/calendar/${id}`, data)
export const deleteCalendarEvent = (id) => http.delete(`/calendar/${id}`)
