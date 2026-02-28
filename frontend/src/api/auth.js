import http from './http'

export const login = (data) => http.post('/login', data)
export const getProfile = () => http.get('/profile')
export const updateProfile = (data) => http.put('/profile', data)
export const changePassword = (data) => http.put('/profile/password', data)
