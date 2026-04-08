<template>
  <router-view v-if="$route.path === '/login'" />
  <el-container v-else style="height: 100vh">
    <el-aside :width="isCollapsed ? '64px' : '200px'" class="sidebar-aside">
      <div class="sidebar-logo">
        {{ isCollapsed ? 'OA' : 'OA系统' }}
      </div>
      <el-menu
        :default-active="currentTab"
        :collapse="isCollapsed"
        :collapse-transition="false"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        class="sidebar-menu"
        @select="onMenuSelect"
      >
        <MenuNode v-for="m in menus" :key="m.id" :node="m" :resolve-icon="resolveIcon" :is-collapsed="isCollapsed" />
      </el-menu>
      <div class="sidebar-bottom">
        <el-tooltip :content="isCollapsed ? '展开菜单' : '折叠菜单'" placement="right">
          <div class="sidebar-collapse-btn" @click="toggleAside">
            <el-icon class="sidebar-collapse-icon"><DArrowLeft v-if="!isCollapsed" /><DArrowRight v-else /></el-icon>
          </div>
        </el-tooltip>
      </div>
    </el-aside>
    <el-container style="overflow:hidden">
      <el-header style="background:#fff;border-bottom:1px solid #eee;display:flex;align-items:center;justify-content:flex-end;padding:0 16px">
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

      <!-- 多页签栏 -->
      <div class="tabs-bar">
        <el-tabs
          v-model="currentTab"
          type="card"
          closable
          @tab-click="onTabClick"
          @tab-remove="onTabRemove"
          class="main-tabs"
        >
          <el-tab-pane
            v-for="tab in openTabs"
            :key="tab.path"
            :label="tab.title"
            :name="tab.path"
            :closable="tab.path !== '/dashboard'"
          />
        </el-tabs>
      </div>

      <el-main style="background:#f0f2f5;overflow:auto">
        <keep-alive>
          <component :is="currentComponent" :key="currentTab" />
        </keep-alive>
      </el-main>
    </el-container>
  </el-container>

  <!-- 设置弹窗 -->
  <el-dialog v-model="settingsVisible" title="个人设置" width="560px" @open="onSettingsOpen">
    <el-tabs v-model="activeTab" tab-position="left" style="min-height:200px">
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
import { ref, computed, onMounted, watch, defineComponent, h, resolveComponent, shallowRef } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as Icons from '@element-plus/icons-vue'
import { updateProfile, changePassword } from './api/auth'
import { getMenus } from './api/menu'

const router = useRouter()
const route = useRoute()

const isCollapsed = ref(false)
const menus = ref([])

// 路由path -> 组件 映射
const routeComponentMap = {
  '/dashboard':                 () => import('./views/Dashboard.vue'),
  '/employee':                  () => import('./views/Employee.vue'),
  '/notice':                    () => import('./views/Notice.vue'),
  '/log':                       () => import('./views/OperationLog.vue'),
  '/meeting-room':              () => import('./views/MeetingRoom.vue'),
  '/event-booking':             () => import('./views/EventBooking.vue'),
  '/leave-request':             () => import('./views/LeaveRequest.vue'),
  '/resignation':               () => import('./views/Resignation.vue'),
  '/workflow':                  () => import('./views/Workflow.vue'),
  '/menu':                      () => import('./views/Menu.vue'),
  '/schedule':                  () => import('./views/Schedule.vue'),
  '/role':                      () => import('./views/Role.vue'),
}

// 多页签状态
const openTabs = ref([{ path: '/dashboard', title: '首页' }])
const currentTab = ref('/dashboard')
const componentCache = shallowRef({})

const currentComponent = computed(() => componentCache.value[currentTab.value] || null)

const loadComponent = async (path) => {
  if (componentCache.value[path]) return
  const loader = routeComponentMap[path]
  if (!loader) return
  const mod = await loader()
  componentCache.value = { ...componentCache.value, [path]: mod.default }
}

const onMenuSelect = async (path) => {
  if (!routeComponentMap[path]) return
  const exists = openTabs.value.find(t => t.path === path)
  if (!exists) {
    // 从菜单树中找标题
    const title = findMenuTitle(menus.value, path) || path
    openTabs.value.push({ path, title })
  }
  currentTab.value = path
  await loadComponent(path)
}

const findMenuTitle = (nodes, path) => {
  for (const n of nodes) {
    if (n.path === path) return n.name
    if (n.children?.length) {
      const found = findMenuTitle(n.children, path)
      if (found) return found
    }
  }
  return null
}

const onTabClick = (tab) => {
  currentTab.value = tab.props.name
}

const onTabRemove = (path) => {
  const idx = openTabs.value.findIndex(t => t.path === path)
  openTabs.value.splice(idx, 1)
  if (currentTab.value === path) {
    currentTab.value = openTabs.value[Math.max(0, idx - 1)].path
  }
  // 清理缓存
  const cache = { ...componentCache.value }
  delete cache[path]
  componentCache.value = cache
}

const resolveIcon = (iconName) => {
  if (iconName && Icons[iconName]) return Icons[iconName]
  return Icons.Menu
}

const MenuNode = defineComponent({
  name: 'MenuNode',
  props: {
    node: { type: Object, required: true },
    resolveIcon: { type: Function, required: true },
    isCollapsed: { type: Boolean, default: false }
  },
  setup(props) {
    return () => {
      const n = props.node || {}
      const hasChildren = Array.isArray(n.children) && n.children.length > 0
      const idx = n.path || `menu-${n.id}`

      const ElSubMenu = resolveComponent('el-sub-menu')
      const ElMenuItem = resolveComponent('el-menu-item')
      const ElTooltip = resolveComponent('el-tooltip')
      const ElIcon = resolveComponent('el-icon')

      if (hasChildren) {
        return h(ElSubMenu, { index: idx }, {
          title: () => [
            h(ElTooltip, { content: n.name, placement: 'right', disabled: !props.isCollapsed }, {
              default: () => h(ElIcon, null, { default: () => h(props.resolveIcon(n.icon)) })
            }),
            h('span', null, n.name)
          ],
          default: () => (n.children || []).map((c) => h(MenuNode, {
            key: c.id,
            node: c,
            resolveIcon: props.resolveIcon,
            isCollapsed: props.isCollapsed
          }))
        })
      }

      return h(ElMenuItem, { index: idx }, {
        default: () => [
          h(ElTooltip, { content: n.name, placement: 'right', disabled: !props.isCollapsed }, {
            default: () => h(ElIcon, null, { default: () => h(props.resolveIcon(n.icon)) })
          }),
          h('span', null, n.name)
        ]
      })
    }
  }
})

const loadMenus = async () => {
  if (!localStorage.getItem('token')) return
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

const parseUserInfo = () => {
  try {
    return JSON.parse(localStorage.getItem('userInfo') || '{}')
  } catch {
    return {}
  }
}

const userInfo = ref(parseUserInfo())
const syncUserInfo = () => {
  userInfo.value = parseUserInfo()
}

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
    const nextUserInfo = { ...userInfo.value, username, real_name }
    localStorage.setItem('userInfo', JSON.stringify(nextUserInfo))
    userInfo.value = nextUserInfo
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
    userInfo.value = {}
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
    userInfo.value = {}
    router.push('/login')
  }
}

onMounted(() => {
  syncUserInfo()
  if (route.path !== '/login') {
    loadMenus()
    // 恢复所有已保存的页签组件
    openTabs.value.forEach(t => loadComponent(t.path))
  }
})

watch(() => route.path, (newPath, oldPath) => {
  syncUserInfo()
  if (oldPath === '/login' && newPath !== '/login') loadMenus()
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
  padding: 20px 12px;
  text-align: center;
  font-weight: bold;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
  overflow-x: hidden;
}

.sidebar-menu::-webkit-scrollbar {
  width: 4px;
}

.sidebar-menu::-webkit-scrollbar-thumb {
  background: rgba(255,255,255,0.15);
  border-radius: 4px;
}

.sidebar-menu::-webkit-scrollbar-track {
  background: transparent;
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

.tabs-bar {
  background: #fff;
  border-bottom: 1px solid #eee;
  padding: 0 8px;
}

.main-tabs {
  height: 40px;
}

:deep(.main-tabs .el-tabs__header) {
  margin: 0;
  border-bottom: none;
}

:deep(.main-tabs .el-tabs__nav) {
  border: none;
}

:deep(.main-tabs .el-tabs__item) {
  height: 40px;
  line-height: 40px;
  border: 1px solid #e4e7ed !important;
  margin-right: 4px;
  border-radius: 4px 4px 0 0;
}

:deep(.main-tabs .el-tabs__item.is-active) {
  background: #f0f2f5;
  border-bottom-color: #f0f2f5 !important;
}
</style>
