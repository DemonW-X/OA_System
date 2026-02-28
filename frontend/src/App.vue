<template>
  <router-view v-if="$route.path === '/login'" />
  <el-container v-else style="height: 100vh">
    <el-aside :width="isCollapsed ? '64px' : '200px'" class="sidebar-aside">
      <div class="sidebar-logo">
        {{ isCollapsed ? 'OA' : 'OA系统' }}
      </div>
      <el-menu
        router
        :default-active="route.path"
        :collapse="isCollapsed"
        :collapse-transition="false"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        class="sidebar-menu"
      >
        <template v-for="m in menus" :key="m.id">
          <el-sub-menu v-if="(m.children || []).length" :index="m.path || `menu-${m.id}`">
            <template #title>
              <el-tooltip :content="m.name" placement="right" :disabled="!isCollapsed">
                <el-icon><component :is="resolveIcon(m.icon)" /></el-icon>
              </el-tooltip>
              <span>{{ m.name }}</span>
            </template>
            <el-menu-item v-for="c in m.children" :key="c.id" :index="c.path || `menu-${c.id}`">
              <el-tooltip :content="c.name" placement="right" :disabled="!isCollapsed">
                <el-icon><component :is="resolveIcon(c.icon)" /></el-icon>
              </el-tooltip>
              <span>{{ c.name }}</span>
            </el-menu-item>
          </el-sub-menu>

          <el-menu-item v-else :index="m.path || `menu-${m.id}`">
            <el-tooltip :content="m.name" placement="right" :disabled="!isCollapsed">
              <el-icon><component :is="resolveIcon(m.icon)" /></el-icon>
            </el-tooltip>
            <span>{{ m.name }}</span>
          </el-menu-item>
        </template>
      </el-menu>
      <div class="sidebar-bottom">
        <el-tooltip :content="isCollapsed ? '展开菜单' : '折叠菜单'" placement="right">
          <div class="sidebar-collapse-btn" @click="toggleAside">
            <el-icon class="sidebar-collapse-icon"><DArrowLeft v-if="!isCollapsed" /><DArrowRight v-else /></el-icon>
          </div>
        </el-tooltip>
      </div>
    </el-aside>
    <el-container>
      <el-header style="background:#fff;border-bottom:1px solid #eee;display:flex;align-items:center;justify-content:space-between">
        <span style="font-size:16px">OA 办公系统</span>
        <el-dropdown @command="handleCommand">
          <span style="cursor:pointer;display:flex;align-items:center;gap:6px">
            <el-avatar :size="28" style="background:#409EFF">{{ userInfo.real_name?.[0] || 'U' }}</el-avatar>
            {{ userInfo.real_name || userInfo.username }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="settings">
                <el-icon><Setting /></el-icon>设置
              </el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main style="background:#f0f2f5">
        <router-view />
      </el-main>
    </el-container>
  </el-container>

  <!-- 设置弹窗 -->
  <el-dialog v-model="settingsVisible" title="个人设置" width="560px" @open="onSettingsOpen">
    <el-tabs v-model="activeTab" tab-position="left" style="min-height:200px">

      <!-- 基本信息 -->
      <el-tab-pane label="基本信息" name="profile">
        <el-form :model="profileForm" :rules="profileRules" ref="profileFormRef" label-width="90px" style="margin-top:10px">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="profileForm.username" placeholder="请输入用户名" />
          </el-form-item>
          <el-form-item label="真实姓名" prop="real_name">
            <el-input v-model="profileForm.real_name" placeholder="请输入真实姓名" />
          </el-form-item>
        </el-form>
        <div style="text-align:right;margin-top:8px">
          <el-button @click="settingsVisible = false">取消</el-button>
          <el-button type="primary" :loading="savingProfile" @click="handleSaveProfile">保存</el-button>
        </div>
      </el-tab-pane>

      <!-- 修改密码 -->
      <el-tab-pane label="修改密码" name="password">
        <el-form :model="pwdForm" :rules="pwdRules" ref="pwdFormRef" label-width="90px" style="margin-top:10px">
          <el-form-item label="原密码" prop="old_password">
            <el-input v-model="pwdForm.old_password" type="password" show-password placeholder="请输入原密码" />
          </el-form-item>
          <el-form-item label="新密码" prop="new_password">
            <el-input v-model="pwdForm.new_password" type="password" show-password placeholder="至少8位，含英文和数字" />
          </el-form-item>
          <el-form-item label="确认密码" prop="confirm_password">
            <el-input v-model="pwdForm.confirm_password" type="password" show-password placeholder="请再次输入新密码" />
          </el-form-item>
          <el-form-item label="验证码" prop="captcha">
            <div style="display:flex;gap:10px;align-items:center">
              <el-input v-model="pwdForm.captcha" placeholder="请输入验证码" maxlength="4" style="width:160px" />
              <div class="captcha-box" @click="refreshCaptcha" title="点击刷新">{{ captchaCode }}</div>
            </div>
          </el-form-item>
        </el-form>
        <div style="text-align:right;margin-top:8px">
          <el-button @click="settingsVisible = false">取消</el-button>
          <el-button type="primary" :loading="savingPwd" @click="handleChangePassword">确认修改</el-button>
        </div>
      </el-tab-pane>

    </el-tabs>
  </el-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as Icons from '@element-plus/icons-vue'
import { updateProfile, changePassword } from './api/auth'
import { getMenus } from './api/menu'

const router = useRouter()
const route = useRoute()

const isCollapsed = ref(false)
const menus = ref([])

const resolveIcon = (iconName) => {
  if (iconName && Icons[iconName]) return Icons[iconName]
  return Icons.Menu
}

const loadMenus = async () => {
  try {
    const res = await getMenus({ tree: 1 })
    const list = res.data?.data || []
    const walk = (arr = []) => arr
      .filter(i => i.visible !== false)
      .map(i => ({ ...i, children: walk(i.children || []) }))
    menus.value = walk(list)
  } catch {
    menus.value = []
  }
}

const toggleAside = () => {
  isCollapsed.value = !isCollapsed.value
}

const userInfo = computed(() => {
  try { return JSON.parse(localStorage.getItem('userInfo') || '{}') } catch { return {} }
})

// 弹窗状态
const settingsVisible = ref(false)
const activeTab = ref('profile')

// 基本信息
const profileFormRef = ref()
const savingProfile = ref(false)
const profileForm = ref({ username: '', real_name: '' })
const profileRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
}

// 修改密码
const pwdFormRef = ref()
const savingPwd = ref(false)
const captchaCode = ref('')
const pwdForm = ref({ old_password: '', new_password: '', confirm_password: '', captcha: '' })

const refreshCaptcha = () => {
  captchaCode.value = String(Math.floor(1000 + Math.random() * 9000))
}
const pwdRules = {
  old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        const minLen = /^.{8,}$/.test(value)
        const hasLetter = /[a-zA-Z]/.test(value)
        const hasDigit = /[0-9]/.test(value)
        if (!minLen || !hasLetter || !hasDigit) {
          callback(new Error('密码不少于8位，且必须包含英文字母和数字'))
        } else if (value === pwdForm.value.old_password) {
          callback(new Error('新密码不能与原密码相同'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== pwdForm.value.new_password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  captcha: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== captchaCode.value) {
          callback(new Error('验证码不正确'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const onSettingsOpen = () => {
  activeTab.value = 'profile'
  profileForm.value = {
    username: userInfo.value.username || '',
    real_name: userInfo.value.real_name || ''
  }
  pwdForm.value = { old_password: '', new_password: '', confirm_password: '', captcha: '' }
  refreshCaptcha()
}

const handleSaveProfile = async () => {
  await profileFormRef.value.validate()
  savingProfile.value = true
  try {
    const res = await updateProfile(profileForm.value)
    const { username, real_name } = res.data.data
    localStorage.setItem('userInfo', JSON.stringify({ ...userInfo.value, username, real_name }))
    ElMessage.success('保存成功')
    settingsVisible.value = false
  } finally {
    savingProfile.value = false
  }
}

const handleChangePassword = async () => {
  await pwdFormRef.value.validate()
  savingPwd.value = true
  try {
    await changePassword({
      old_password: pwdForm.value.old_password,
      new_password: pwdForm.value.new_password
    })
    ElMessage.success('密码修改成功，请重新登录')
    settingsVisible.value = false
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
    router.push('/login')
  } catch {
    refreshCaptcha()
    pwdForm.value.captcha = ''
  } finally {
    savingPwd.value = false
  }
}

const handleCommand = async (cmd) => {
  if (cmd === 'settings') {
    settingsVisible.value = true
  } else if (cmd === 'logout') {
    await ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
    router.push('/login')
  }
}

onMounted(() => {
  loadMenus()
})
</script>

<style scoped>
.sidebar-aside {
  background: #304156;
  transition: width .2s;
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.sidebar-logo {
  color: #fff;
  font-size: 18px;
  padding: 20px 8px;
  text-align: center;
  font-weight: bold;
  white-space: nowrap;
  overflow: hidden;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
}

.sidebar-bottom {
  padding: 10px 8px 12px;
  border-top: 1px solid rgba(255,255,255,.08);
}

.sidebar-collapse-btn {
  height: 36px;
  border-radius: 6px;
  color: #bfcbd9;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
  background: rgba(255, 255, 255, 0.04);
}

.sidebar-collapse-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.sidebar-collapse-icon {
  font-size: 16px;
}

.sidebar-collapse-text {
  font-size: 13px;
  white-space: nowrap;
}

/* 折叠菜单时仅保留菜单图标，彻底隐藏菜单文字（仅作用于菜单区域） */
:deep(.sidebar-menu.el-menu--collapse .el-menu-item span),
:deep(.sidebar-menu.el-menu--collapse .el-sub-menu__title span) {
  display: none !important;
}

:deep(.el-menu--collapse .el-menu-item),
:deep(.el-menu--collapse .el-sub-menu__title) {
  justify-content: center;
}

.captcha-box {
  width: 100px;
  height: 32px;
  line-height: 32px;
  text-align: center;
  font-size: 20px;
  font-weight: bold;
  letter-spacing: 6px;
  color: #409EFF;
  background: #f0f7ff;
  border: 1px dashed #409EFF;
  border-radius: 4px;
  cursor: pointer;
  user-select: none;
}
.captcha-box:hover {
  background: #e0f0ff;
}
</style>
