<template>
  <div>
    <el-row :gutter="20">
      <!-- 待我审核 -->
      <el-col :span="12" class="panel-col">
        <el-card shadow="hover" class="panel-card">
          <template #header>
            <div style="display:flex;align-items:center;gap:8px">
              <el-icon color="#E6A23C"><Bell /></el-icon>
              <span style="font-weight:bold">我的审核</span>
              <el-tag size="small" type="warning" style="margin-left:auto">最新 {{ pendingApprovals.length }} 条</el-tag>
              <el-button link type="primary" @click="openAllApprovalsDialog">更多</el-button>
            </div>
          </template>
          <el-empty v-if="pendingApprovals.length === 0" description="暂无待审核事项" :image-size="60" />
          <ul v-else class="approval-list">
            <li v-for="item in pendingApprovals" :key="item.task_id" class="approval-item">
              <el-link type="primary" @click="openApproval(item)">{{ item.title }}</el-link>
            </li>
          </ul>
        </el-card>
      </el-col>

      <!-- 公告栏 -->
      <el-col :span="12" class="panel-col">
        <el-card shadow="hover" class="panel-card">
          <template #header>
            <div style="display:flex;align-items:center;gap:8px">
              <el-icon color="#409EFF"><Bell /></el-icon>
              <span style="font-weight:bold">公告栏</span>
              <el-tag size="small" type="info" style="margin-left:auto">最新 {{ noticeList.length }} 条</el-tag>
              <el-button link type="primary" @click="openAllNoticesDialog">更多</el-button>
            </div>
          </template>
          <el-empty v-if="noticeList.length === 0" description="暂无公告" :image-size="60" />
          <ul v-else class="notice-list">
            <li v-for="item in noticeList" :key="item.id" class="notice-item">
              <el-link type="primary" @click="openNotice(item)">{{ item.title }}</el-link>
              <span class="notice-meta">{{ item.author }} · {{ formatDate(item.created_at) }}</span>
            </li>
          </ul>
        </el-card>
      </el-col>
    </el-row>

    <!-- 全部待审核弹窗 -->
    <el-dialog v-model="allApprovalsDialogVisible" width="900px" title="全部待审核事项">
      <el-empty v-if="allPendingApprovals.length === 0" description="暂无待审核事项" :image-size="60" />
      <ul v-else class="approval-list all-approval-list">
        <li v-for="item in allPendingApprovals" :key="item.task_id" class="approval-item">
          <el-link type="primary" @click="openApproval(item)">{{ item.title }}</el-link>
        </li>
      </ul>
    </el-dialog>

    <!-- 审核详情弹窗 -->
    <el-dialog v-model="approvalDialogVisible" width="680px" title="审核详情">
      <el-descriptions v-if="approvalDetail" :column="1" border>
        <el-descriptions-item label="事项">{{ approvalDetail.title || '-' }}</el-descriptions-item>
        <el-descriptions-item label="业务类型">{{ getBizTypeLabel(approvalDetail.biz_type, approvalDetail.detail_path) }}</el-descriptions-item>
        <el-descriptions-item label="当前状态">{{ getStatusLabel(approvalDetail.status) }}</el-descriptions-item>
        <el-descriptions-item label="详情信息">{{ approvalDetail.summary || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-form :model="approvalForm" label-width="80px" style="margin-top:16px">
        <el-form-item label="审核动作">
          <el-radio-group v-model="approvalForm.action">
            <el-radio label="approved">通过</el-radio>
            <el-radio label="rejected">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="说明">
          <el-input
            v-model="approvalForm.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入审核说明"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="approvalDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="approvalSubmitting" @click="submitApproval('approved')">审核通过</el-button>
        <el-button type="danger" plain :loading="approvalSubmitting" @click="submitApproval('rejected')">拒绝</el-button>
      </template>
    </el-dialog>

    <!-- 全部公告弹窗 -->
    <el-dialog v-model="allNoticesDialogVisible" width="900px" title="全部公告">
      <el-empty v-if="allNoticeList.length === 0" description="暂无公告" :image-size="60" />
      <el-table v-else :data="allNoticeList" size="small" border max-height="520">
        <el-table-column prop="title" label="标题" min-width="320" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" @click="openNotice(row)">{{ row.title }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="author" label="作者" width="120" />
        <el-table-column prop="created_at" label="发布时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 公告详情弹窗 -->
    <el-dialog v-model="dialogVisible" width="600px" :title="currentNotice.title">
      <div style="margin-bottom:12px">
        <el-tag type="success" v-if="currentNotice.status === 1">已发布</el-tag>
        <el-tag type="info" v-else>草稿</el-tag>
        <span style="color:#999;font-size:13px;margin-left:10px">
          作者：{{ currentNotice.author }} · {{ formatDate(currentNotice.created_at) }}
        </span>
      </div>
      <el-divider />
      <div v-if="currentNotice.content" class="notice-content" v-html="currentNotice.content" />
      <div v-else style="color:#999;text-align:center;padding:20px">暂无内容</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getEmployees, getEmployee, approveEmployee } from '../api/employee'
import { getNotices } from '../api/notice'
import { getMenus } from '../api/menu'
import { getBizTypes } from '../api/workflow'
import { getLeaveRequest, approveLeaveRequest } from '../api/leave_request'
import { getEventBooking, approveEventBooking } from '../api/event_booking'
import { getMyPendingApprovals } from '../api/orchid_workflow'

const router = useRouter()

const pendingApprovals = ref([])
const allPendingApprovals = ref([])
const noticeList = ref([])
const allNoticeList = ref([])
const allApprovalsDialogVisible = ref(false)
const allNoticesDialogVisible = ref(false)
const approvalDialogVisible = ref(false)
const approvalDetail = ref(null)
const approvalForm = ref({ action: 'approved', remark: '' })
const approvalSubmitting = ref(false)
const dialogVisible = ref(false)
const currentNotice = ref({})
const menuPathLabelMap = ref({})
const bizTypeLabelMap = ref({})

const bizTypeMap = {
  employee: '员工管理',
  leave_request: '请假管理',
  event_booking: '事件预定'
}

const statusMap = {
  draft: '草稿',
  pending: '待审批',
  approved: '已通过',
  rejected: '已拒绝',
  withdrawn: '已撤回'
}

const getBizTypeLabel = (bizType, detailPath = '') => {
  const byPath = menuPathLabelMap.value[detailPath]
  if (byPath) return byPath
  const byBiz = bizTypeLabelMap.value[bizType]
  if (byBiz) return byBiz
  return bizTypeMap[bizType] || bizType || '-'
}
const getStatusLabel = (status) => statusMap[status] || status || '-'

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit'
  })
}

const openNotice = (item) => {
  currentNotice.value = item
  dialogVisible.value = true
}

const loadBizLabels = async () => {
  try {
    const [menuRes, bizRes] = await Promise.all([
      getMenus({ tree: 1 }),
      getBizTypes()
    ])

    const menuMap = {}
    const walk = (arr = []) => {
      arr.forEach((m) => {
        if (m?.path && m?.name) menuMap[m.path] = m.name
        if (Array.isArray(m?.children) && m.children.length) walk(m.children)
      })
    }
    walk(menuRes.data?.data || [])
    menuPathLabelMap.value = menuMap

    const bizMap = {}
    ;(bizRes.data?.data || []).forEach((b) => {
      if (b?.code && b?.name) bizMap[b.code] = b.name
    })
    bizTypeLabelMap.value = bizMap
  } catch {
    // 忽略，使用本地兜底映射
  }
}

const openApproval = async (item) => {
  if (!item?.biz_type || !item?.biz_id) return

  try {
    let detail = null
    let summary = ''

    if (item.biz_type === 'employee') {
      const res = await getEmployee(item.biz_id)
      detail = res.data?.data
      summary = `姓名：${detail?.name || '-'}；部门：${detail?.department?.name || '-'}；职位：${detail?.position_info?.name || '-'}；手机号：${detail?.phone || '-'}`
    } else if (item.biz_type === 'leave_request') {
      const res = await getLeaveRequest(item.biz_id)
      detail = res.data?.data
      summary = `请假人：${detail?.employee?.name || '-'}；类型：${detail?.type || '-'}；时间：${formatDate(detail?.start_date)} ~ ${formatDate(detail?.end_date)}；天数：${detail?.days ?? '-'}`
    } else if (item.biz_type === 'event_booking') {
      const res = await getEventBooking(item.biz_id)
      detail = res.data?.data
      summary = `主题：${detail?.title || '-'}；类型：${detail?.type || '-'}；时间：${formatDate(detail?.start_time)} ~ ${formatDate(detail?.end_time)}；会议室：${detail?.meeting_room?.name || '-'}`
    } else {
      ElMessage.warning('暂不支持该业务类型的详情审核')
      return
    }

    approvalDetail.value = {
      ...item,
      detail,
      summary
    }
    approvalForm.value = { action: 'approved', remark: '' }
    allApprovalsDialogVisible.value = false
    approvalDialogVisible.value = true
  } catch {
    ElMessage.error('【待我审核】获取详情失败')
  }
}

const openAllApprovalsDialog = async () => {
  try {
    const res = await getMyPendingApprovals({ limit: 1000 })
    allPendingApprovals.value = res.data?.data || []
    allApprovalsDialogVisible.value = true
  } catch {
    ElMessage.error('【待我审核】获取数据失败')
  }
}

const reloadApprovals = async () => {
  const [homeRes, allRes] = await Promise.all([
    getMyPendingApprovals({ limit: 20 }),
    getMyPendingApprovals({ limit: 1000 })
  ])
  pendingApprovals.value = homeRes.data?.data || []
  allPendingApprovals.value = allRes.data?.data || []
}

const submitApproval = async (action) => {
  if (!approvalDetail.value?.biz_type || !approvalDetail.value?.biz_id) return

  approvalSubmitting.value = true
  try {
    const remark = (approvalForm.value.remark || '').trim()
    const bizType = approvalDetail.value.biz_type
    const bizID = approvalDetail.value.biz_id

    if (bizType === 'employee') {
      await approveEmployee(bizID, { action, remark })
    } else if (bizType === 'leave_request') {
      await approveLeaveRequest(bizID, { status: action, reject_reason: remark })
    } else if (bizType === 'event_booking') {
      await approveEventBooking(bizID, { status: action, reject_reason: remark })
    } else {
      ElMessage.warning('暂不支持该业务类型的审批操作')
      return
    }

    ElMessage.success(action === 'approved' ? '审核通过成功' : '拒绝成功')
    approvalDialogVisible.value = false
    approvalDetail.value = null
    await reloadApprovals()
  } catch {
    ElMessage.error('审核操作失败')
  } finally {
    approvalSubmitting.value = false
  }
}

const openAllNoticesDialog = async () => {
  try {
    const res = await getNotices({ page: 1, page_size: 1000 })
    allNoticeList.value = res.data?.data?.list || []
    allNoticesDialogVisible.value = true
  } catch {
    ElMessage.error('【公告栏】获取数据失败')
  }
}

onMounted(async () => {
  loadBizLabels()

  const tasks = [
    { key: 'noticesList', label: '公告栏', req: getNotices({ page: 1, page_size: 20, status: 1 }) },
    { key: 'approvals', label: '待我审核', req: getMyPendingApprovals({ limit: 20 }) }
  ]

  const results = await Promise.allSettled(tasks.map(t => t.req))

  results.forEach((result, idx) => {
    const task = tasks[idx]
    if (result.status === 'rejected') {
      ElMessage.error(`【${task.label}】获取数据失败`)
      return
    }

    const data = result.value?.data?.data
    switch (task.key) {
      case 'noticesList':
        noticeList.value = data?.list || []
        break
      case 'approvals':
        pendingApprovals.value = data || []
        break
    }
  })
})
</script>

<style scoped>
.panel-col {
  display: flex;
}

.panel-card {
  width: 100%;
  min-height: 320px;
}

.approval-list,
.notice-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.approval-item,
.notice-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.approval-item:last-child,
.notice-item:last-child {
  border-bottom: none;
}

.all-approval-list {
  max-height: 520px;
  overflow: auto;
}

.notice-meta {
  font-size: 12px;
  color: #999;
  white-space: nowrap;
  margin-left: 12px;
}
.notice-content {
  line-height: 1.8;
  color: #333;
  min-height: 80px;
}
.notice-content :deep(img) {
  max-width: 100%;
  height: auto;
}
.notice-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
}
.notice-content :deep(td),
.notice-content :deep(th) {
  border: 1px solid #ddd;
  padding: 6px 10px;
}
.notice-content :deep(blockquote) {
  border-left: 4px solid #ddd;
  margin: 0;
  padding-left: 16px;
  color: #666;
}
</style>
