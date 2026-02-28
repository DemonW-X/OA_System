import http from './http'

export const getMeetingRooms = (params) => http.get('/meeting-rooms', { params })
export const createMeetingRoom = (data) => http.post('/meeting-rooms', data)
export const updateMeetingRoom = (id, data) => http.put(`/meeting-rooms/${id}`, data)
export const deleteMeetingRoom = (id) => http.delete(`/meeting-rooms/${id}`)
