/**
 * 公共表单验证规则（后端校验版）
 * 格式校验统一调用后端 POST /api/validate/format 接口。
 * 使用方式：import { phoneRules, emailRules, idCardRules } from '../utils/validators'
 */

import http from '../api/http'

/**
 * 调用后端公共格式校验接口
 * @param {'phone'|'email'|'id_card'} type
 * @param {string} value
 * @returns {Promise<{valid: boolean, msg?: string}>}
 */
export function validateFormat(type, value) {
  return http.post('/validate/format', { type, value }).then(res => res.data)
}

/**
 * 生成调用后端校验的 Element Plus 异步 validator
 */
function makeRemoteValidator(type) {
  return (rule, value, callback) => {
    if (!value) return callback()
    validateFormat(type, value)
      .then(data => {
        if (data.valid) {
          callback()
        } else {
          callback(new Error(data.msg || '格式不正确'))
        }
      })
      .catch(() => {
        callback(new Error('网络异常，格式校验失败，请稍后重试'))
      })
  }
}

/** 手机号验证（必填 + 后端格式校验） */
export const phoneRules = [
  { required: true, message: '请输入联系电话', trigger: 'blur' },
  { validator: makeRemoteValidator('phone'), trigger: 'blur' }
]

/** 邮箱验证（必填 + 后端格式校验） */
export const emailRules = [
  { required: true, message: '请输入电子邮箱', trigger: 'blur' },
  { validator: makeRemoteValidator('email'), trigger: 'blur' }
]

/** 身份证号验证（必填 + 后端格式校验） */
export const idCardRules = [
  { required: true, message: '请输入身份证号', trigger: 'blur' },
  { validator: makeRemoteValidator('id_card'), trigger: 'blur' }
]
