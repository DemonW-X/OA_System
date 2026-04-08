<template>
  <div class="login-wrap">
    <el-card class="login-card">
      <div class="login-title">OA 办公系统</div>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="0">
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            placeholder="密码"
            size="large"
            type="password"
            prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-button type="primary" size="large" style="width:100%" :loading="loading" @click="handleLogin">
          登录
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '../api/auth'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const form = ref({ username: '', password: '' })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const getLoginErrorMsg = (payload) => {
  if (!payload) return '登录失败'
  if (typeof payload.msg === 'string' && payload.msg.trim()) return payload.msg
  if (payload.msg && typeof payload.msg === 'object') {
    if (typeof payload.msg.msg === 'string' && payload.msg.msg.trim()) return payload.msg.msg
    if (typeof payload.msg.message === 'string' && payload.msg.message.trim()) return payload.msg.message
  }
  if (typeof payload.message === 'string' && payload.message.trim()) return payload.message
  return '登录失败'
}

const handleLogin = async () => {
  await formRef.value.validate()
  loading.value = true
  try {
    const res = await login(form.value)
    const payload = res?.data || {}
    if (payload.code !== 0 || !payload.data?.token) {
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      ElMessage.error(getLoginErrorMsg(payload))
      return
    }
    const { token, username, real_name, role } = payload.data
    localStorage.setItem('token', token)
    localStorage.setItem('userInfo', JSON.stringify({ username, real_name, role }))
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (err) {
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
    ElMessage.error(getLoginErrorMsg(err?.response?.data))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
}

.login-card {
  width: 380px;
}

.login-title {
  text-align: center;
  font-size: 22px;
  font-weight: bold;
  margin-bottom: 28px;
  color: #303133;
}
</style>
