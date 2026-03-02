import http from './http'

export const getEventBookings = (params) => http.get('/event-bookings', { params })
export const getEventBooking = (id) => http.get(`/event-bookings/${id}`)
export const createEventBooking = (data) => http.post('/event-bookings', data)
export const updateEventBooking = (id, data) => http.put(`/event-bookings/${id}`, data)
export const submitEventBooking = (id, data) => http.put(`/event-bookings/${id}/submit`, data)
export const approveEventBooking = (id, data) => http.put(`/event-bookings/${id}/approve`, data)
export const deleteEventBooking = (id) => http.delete(`/event-bookings/${id}`)
