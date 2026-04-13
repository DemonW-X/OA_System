import http from './http'

export const getDataDictionaries = (params) => http.get('/data-dictionaries', { params })
export const getDataDictionary = (id) => http.get(`/data-dictionaries/${id}`)
export const createDataDictionary = (data) => http.post('/data-dictionaries', data)
export const updateDataDictionary = (id, data) => http.put(`/data-dictionaries/${id}`, data)
export const deleteDataDictionary = (id) => http.delete(`/data-dictionaries/${id}`)

export const getDataDictionaryItems = (id, params) => http.get(`/data-dictionaries/${id}/items`, { params })
export const createDataDictionaryItem = (id, data) => http.post(`/data-dictionaries/${id}/items`, data)
export const updateDataDictionaryItem = (id, data) => http.put(`/data-dictionary-items/${id}`, data)
export const deleteDataDictionaryItem = (id) => http.delete(`/data-dictionary-items/${id}`)

