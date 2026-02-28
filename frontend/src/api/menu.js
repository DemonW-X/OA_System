import http from './http'

export const getMenus = (params) => http.get('/menus', { params })
export const getMenu = (id) => http.get(`/menus/${id}`)
export const createMenu = (data) => http.post('/menus', data)
export const updateMenu = (id, data) => http.put(`/menus/${id}`, data)
export const deleteMenu = (id) => http.delete(`/menus/${id}`)
