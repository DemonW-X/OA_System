<template>
  <div>
    <el-row :gutter="20">
      <!-- 我的审核 -->
      <el-col :span="12" class="panel-col">
        <el-card shadow="hover" class="panel-card">
          <template #header>
            <div style="display:flex;align-items:center;gap:8px">
              <el-icon color="#E6A23C"><Bell /></el-icon>
              <span style="font-weight:bold">我的审核</span>
              <el-button link type="primary" style="margin-left:auto" @click="openAllAuditDialog">更多</el-button>
            </div>
          </template>
            <el-tabs v-model="approvalActiveTab" @tab-change="onApprovalTabChange" style="margin-top:-8px">
              <el-tab-pane name="pending">
                <template #label>
                  我的待审
                  <el-badge v-if="pendingApprovals.length > 0" :value="pendingApprovals.length" :max="99" style="margin-left:4px" />
                </template>
                <el-empty v-if="pendingApprovals.length === 0" description="暂无待审核事项" :image-size="50" />
                <ul v-else class="approval-list">
                  <li v-for="item in pendingApprovals" :key="item.task_id" class="approval-item">
                    <el-link type="primary" @click="openApproval(item)">{{ item.title }}</el-link>
                    <span class="approval-meta">{{ item.created_at }}</span>
                  </li>
                </ul>
                <div class="tab-pagination">
                  <el-pagination
                    v-model:current-page="pendingPage"
                    :page-size="pageSize"
                    :total="pendingTotal"
                    layout="prev, pager, next"
                    small
                    @current-change="loadPending"
                  />
                </div>
              </el-tab-pane>

              <el-tab-pane name="approved" label="我的已审">
                <el-empty v-if="approvedList.length === 0" description="暂无已审核事项" :image-size="50" />
                <ul v-else class="approval-list">
                  <li v-for="item in approvedList" :key="item.task_id" class="approval-item">
                    <el-link type="primary" @click="openApprovalReadonly(item)">{{ item.title }}</el-link>
                    <span class="approval-meta">{{ item.created_at }}</span>
                  </li>
                </ul>
                <div class="tab-pagination">
                  <el-pagination
                    v-model:current-page="approvedPage"
                    :page-size="pageSize"
                    :total="approvedTotal"
                    layout="prev, pager, next"
                    small
                    @current-change="loadApproved"
                  />
                </div>
              </el-tab-pane>

              <el-tab-pane name="pending-read">
                <template #label>
                  我的待阅
                  <el-badge v-if="pendingReadList.length > 0" :value="pendingReadList.length" :max="99" style="margin-left:4px" />
                </template>
                <el-empty v-if="pendingReadList.length === 0" description="暂无待阅事项" :image-size="50" />
                <ul v-else class="approval-list">
                  <li v-for="item in pendingReadList" :key="item.task_id" class="approval-item">
                    <el-link type="primary" @click="openApprovalReadonly(item)">{{ item.title }}</el-link>
                    <span class="approval-meta">{{ item.created_at }}</span>
                  </li>
                </ul>
                <div class="tab-pagination">
                  <el-pagination
                    v-model:current-page="pendingReadPage"
                    :page-size="pageSize"
                    :total="pendingReadTotal"
                    layout="prev, pager, next"
                    small
                    @current-change="loadPendingRead"
                  />
                </div>
              </el-tab-pane>

              <el-tab-pane name="read" label="我的已阅">
                <el-empty v-if="readList.length === 0" description="暂无已阅事项" :image-size="50" />
                <ul v-else class="approval-list">
                  <li v-for="item in readList" :key="item.task_id" class="approval-item">
                    <el-link type="primary" @click="openApprovalReadonly(item)">{{ item.title }}</el-link>
                    <span class="approval-meta">{{ item.created_at }}</span>
                  </li>
                </ul>
                <div class="tab-pagination">
                  <el-pagination
                    v-model:current-page="readPage"
                    :page-size="pageSize"
                    :total="readTotal"
                    layout="prev, pager, next"
                    small
                    @current-change="loadRead"
                  />
                </div>
              </el-tab-pane>
            </el-tabs>
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

    <!-- 全部审核弹窗 -->
    <el-dialog v-model="allAuditDialogVisible" width="860px" title="我的审核" destroy-on-close>
      <el-tabs v-model="allAuditActiveTab" @tab-change="onAllAuditTabChange">
        <el-tab-pane name="pending">
          <template #label>我的待审<el-badge v-if="allAuditPendingTotal > 0" :value="allAuditPendingTotal" :max="99" style="margin-left:4px" /></template>
          <el-empty v-if="allAuditPendingList.length === 0" description="暂无待审核事项" :image-size="50" />
          <ul v-else class="approval-list all-audit-list">
            <li v-for="item in allAuditPendingList" :key="item.task_id" class="approval-item">
              <el-link type="primary" @click="openApproval(item)">{{ item.title }}</el-link>
              <span class="approval-meta">{{ item.created_at }}</span>
            </li>
          </ul>
          <div class="tab-pagination">
            <el-pagination v-model:current-page="allAuditPendingPage" :page-size="allAuditPageSize" :total="allAuditPendingTotal" layout="prev, pager, next" small @current-change="loadAllAuditPending" />
          </div>
        </el-tab-pane>
        <el-tab-pane name="approved" label="我的已审">
          <el-empty v-if="allAuditApprovedList.length === 0" description="暂无已审核事项" :image-size="50" />
          <ul v-else class="approval-list all-audit-list">
            <li v-for="item in allAuditApprovedList" :key="item.task_id" class="approval-item">
              <el-link type="primary" @click="openApprovalReadonly(item)">{{ item.title }}</el-link>
              <span class="approval-meta">{{ item.created_at }}</span>
            </li>
          </ul>
          <div class="tab-pagination">
            <el-pagination v-model:current-page="allAuditApprovedPage" :page-size="allAuditPageSize" :total="allAuditApprovedTotal" layout="prev, pager, next" small @current-change="loadAllAuditApproved" />
          </div>
        </el-tab-pane>
        <el-tab-pane name="pending-read">
          <template #label>我的待阅<el-badge v-if="allAuditPendingReadTotal > 0" :value="allAuditPendingReadTotal" :max="99" style="margin-left:4px" /></template>
          <el-empty v-if="allAuditPendingReadList.length === 0" description="暂无待阅事项" :image-size="50" />
          <ul v-else class="approval-list all-audit-list">
            <li v-for="item in allAuditPendingReadList" :key="item.task_id" class="approval-item">
              <el-link type="primary" @click="openApprovalReadonly(item)">{{ item.title }}</el-link>
              <span class="approval-meta">{{ item.created_at }}</span>
            </li>
          </ul>
          <div class="tab-pagination">
            <el-pagination v-model:current-page="allAuditPendingReadPage" :page-size="allAuditPageSize" :total="allAuditPendingReadTotal" layout="prev, pager, next" small @current-change="loadAllAuditPendingRead" />
          </div>
        </el-tab-pane>
        <el-tab-pane name="read" label="我的已阅">
          <el-empty v-if="allAuditReadList.length === 0" description="暂无已阅事项" :image-size="50" />
          <ul v-else class="approval-list all-audit-list">
            <li v-for="item in allAuditReadList" :key="item.task_id" class="approval-item">
              <el-link type="primary" @click="openApprovalReadonly(item)">{{ item.title }}</el-link>
              <span class="approval-meta">{{ item.created_at }}</span>
            </li>
          </ul>
          <div class="tab-pagination">
            <el-pagination v-model:current-page="allAuditReadPage" :page-size="allAuditPageSize" :total="allAuditReadTotal" layout="prev, pager, next" small @current-change="loadAllAuditRead" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

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

      <el-form v-if="!approvalDetail?.readonly" :model="approvalForm" label-width="80px" style="margin-top:16px">
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
        <el-button @click="approvalDialogVisible = false">关闭</el-button>
        <template v-if="!approvalDetail?.readonly">
          <el-button type="primary" :loading="approvalSubmitting" @click="submitApproval('approved')">审核通过</el-button>
          <el-button type="danger" plain :loading="approvalSubmitting" @click="submitApproval('rejected')">拒绝</el-button>
        </template>
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
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getEmployees, getEmployee, approveEmployee } from '../api/employee'
import { getNotices } from '../api/notice'
import { getMenus } from '../api/menu'
import { getBizTypes } from '../api/workflow'
import { getLeaveRequest, approveLeaveRequest } from '../api/leave_request'
import { getEventBooking, approveEventBooking } from '../api/event_booking'
import { getMyPendingApprovals, getMyApprovedApprovals, getMyPendingReads, getMyReadItems } from '../api/orchid_workflow'

const router = useRouter()

const pendingApprovals = ref([])
const pendingTotal = ref(0)
const pendingPage = ref(1)
const approvedList = ref([])
const approvedTotal = ref(0)
const approvedPage = ref(1)
const pendingReadList = ref([])
const pendingReadTotal = ref(0)
const pendingReadPage = ref(1)
const readList = ref([])
const readTotal = ref(0)
const readPage = ref(1)
const pageSize = 10
const approvalActiveTab = ref('pending')
const allPendingApprovals = ref([])
const noticeList = ref([])
const allNoticeList = ref([])
const allApprovalsDialogVisible = ref(false)
const allNoticesDialogVisible = ref(false)
const allAuditDialogVisible = ref(false)
const allAuditActiveTab = ref('pending')
const allAuditPageSize = 20
const allAuditPendingList = ref([])
const allAuditPendingTotal = ref(0)
const allAuditPendingPage = ref(1)
const allAuditApprovedList = ref([])
const allAuditApprovedTotal = ref(0)
const allAuditApprovedPage = ref(1)
const allAuditPendingReadList = ref([])
const allAuditPendingReadTotal = ref(0)
const allAuditPendingReadPage = ref(1)
const allAuditReadList = ref([])
const allAuditReadTotal = ref(0)
const allAuditReadPage = ref(1)
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

const AUTO_REFRESH_INTERVAL = 5 * 60 * 1000
let autoRefreshTimer = null
const isAutoRefreshing = ref(false)

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

const isAuthExpiredError = (error) => {
  return error?.response?.status === 401 || error?.__authExpired
}

const showErrorIfNeeded = (error, message) => {
  if (isAuthExpiredError(error)) return
  ElMessage.error(message)
}

const loadNoticeCard = async (silent = false) => {
  try {
    const res = await getNotices({ page: 1, page_size: 20, status: 1 })
    noticeList.value = res.data?.data?.list || []
  } catch (error) {
    if (!silent) showErrorIfNeeded(error, '【公告栏】获取数据失败')
  }
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
  } catch (error) {
    showErrorIfNeeded(error, '【待我审核】获取详情失败')
  }
}

const openApprovalReadonly = async (item) => {
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
    }
    approvalDetail.value = { ...item, detail, summary }
    approvalDialogVisible.value = true
  } catch (error) {
    showErrorIfNeeded(error, '获取详情失败')
  }
}

const openAllApprovalsDialog = async () => {
  try {
    const res = await getMyPendingApprovals({ limit: 1000 })
    allPendingApprovals.value = res.data?.data || []
    allApprovalsDialogVisible.value = true
  } catch (error) {
    showErrorIfNeeded(error, '【待我审核】获取数据失败')
  }
}

const reloadApprovals = async () => {
  await loadPending(pendingPage.value)
}

const loadPending = async (page = 1, options = {}) => {
  const { silent = false } = options
  pendingPage.value = page
  try {
    const res = await getMyPendingApprovals({ page, page_size: pageSize })
    pendingApprovals.value = res.data?.data?.list || res.data?.data || []
    pendingTotal.value = res.data?.data?.total || pendingApprovals.value.length
  } catch (error) {
    if (!silent) showErrorIfNeeded(error, '【我的待审】获取数据失败')
  }
}

const loadApproved = async (page = 1, options = {}) => {
  const { silent = false } = options
  approvedPage.value = page
  try {
    const res = await getMyApprovedApprovals({ page, page_size: pageSize })
    approvedList.value = res.data?.data?.list || res.data?.data || []
    approvedTotal.value = res.data?.total || res.data?.data?.total || approvedList.value.length
  } catch (error) {
    if (!silent) showErrorIfNeeded(error, '【我的已审】获取数据失败')
  }
}

const loadPendingRead = async (page = 1, options = {}) => {
  const { silent = false } = options
  pendingReadPage.value = page
  try {
    const res = await getMyPendingReads({ page, page_size: pageSize })
    pendingReadList.value = res.data?.data?.list || res.data?.data || []
    pendingReadTotal.value = res.data?.total || res.data?.data?.total || pendingReadList.value.length
  } catch (error) {
    if (!silent) showErrorIfNeeded(error, '【我的待阅】获取数据失败')
  }
}

const loadRead = async (page = 1, options = {}) => {
  const { silent = false } = options
  readPage.value = page
  try {
    const res = await getMyReadItems({ page, page_size: pageSize })
    readList.value = res.data?.data?.list || res.data?.data || []
    readTotal.value = res.data?.total || res.data?.data?.total || readList.value.length
  } catch (error) {
    if (!silent) showErrorIfNeeded(error, '【我的已阅】获取数据失败')
  }
}

const refreshHomeCards = async (silent = true) => {
  if (isAutoRefreshing.value) return
  isAutoRefreshing.value = true
  try {
    await Promise.all([
      loadNoticeCard(silent),
      loadPending(pendingPage.value, { silent }),
      loadApproved(approvedPage.value, { silent }),
      loadPendingRead(pendingReadPage.value, { silent }),
      loadRead(readPage.value, { silent })
    ])
  } finally {
    isAutoRefreshing.value = false
  }
}

const onApprovalTabChange = (tab) => {
  if (tab === 'pending' && pendingApprovals.value.length === 0) loadPending()
  else if (tab === 'approved' && approvedList.value.length === 0) loadApproved()
  else if (tab === 'pending-read' && pendingReadList.value.length === 0) loadPendingRead()
  else if (tab === 'read' && readList.value.length === 0) loadRead()
}

const loadAllAuditPending = async (page = 1) => {
  allAuditPendingPage.value = page
  try {
    const res = await getMyPendingApprovals({ page, page_size: allAuditPageSize })
    allAuditPendingList.value = res.data?.data?.list || res.data?.data || []
    allAuditPendingTotal.value = res.data?.data?.total || allAuditPendingList.value.length
  } catch (error) { showErrorIfNeeded(error, '获取待审数据失败') }
}

const loadAllAuditApproved = async (page = 1) => {
  allAuditApprovedPage.value = page
  try {
    const res = await getMyApprovedApprovals({ page, page_size: allAuditPageSize })
    allAuditApprovedList.value = res.data?.data?.list || res.data?.data || []
    allAuditApprovedTotal.value = res.data?.data?.total || allAuditApprovedList.value.length
  } catch (error) { showErrorIfNeeded(error, '获取已审数据失败') }
}

const loadAllAuditPendingRead = async (page = 1) => {
  allAuditPendingReadPage.value = page
  try {
    const res = await getMyPendingReads({ page, page_size: allAuditPageSize })
    allAuditPendingReadList.value = res.data?.data?.list || res.data?.data || []
    allAuditPendingReadTotal.value = res.data?.data?.total || allAuditPendingReadList.value.length
  } catch (error) { showErrorIfNeeded(error, '获取待阅数据失败') }
}

const loadAllAuditRead = async (page = 1) => {
  allAuditReadPage.value = page
  try {
    const res = await getMyReadItems({ page, page_size: allAuditPageSize })
    allAuditReadList.value = res.data?.data?.list || res.data?.data || []
    allAuditReadTotal.value = res.data?.data?.total || allAuditReadList.value.length
  } catch (error) { showErrorIfNeeded(error, '获取已阅数据失败') }
}

const openAllAuditDialog = async () => {
  allAuditActiveTab.value = 'pending'
  allAuditDialogVisible.value = true
  await loadAllAuditPending(1)
}

const onAllAuditTabChange = (tab) => {
  if (tab === 'pending') loadAllAuditPending(1)
  else if (tab === 'approved') loadAllAuditApproved(1)
  else if (tab === 'pending-read') loadAllAuditPendingRead(1)
  else if (tab === 'read') loadAllAuditRead(1)
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
  } catch (error) {
    showErrorIfNeeded(error, '审核操作失败')
  } finally {
    approvalSubmitting.value = false
  }
}

const openAllNoticesDialog = async () => {
  try {
    const res = await getNotices({ page: 1, page_size: 1000 })
    allNoticeList.value = res.data?.data?.list || []
    allNoticesDialogVisible.value = true
  } catch (error) {
    showErrorIfNeeded(error, '【公告栏】获取数据失败')
  }
}

onMounted(async () => {
  loadBizLabels()
  await refreshHomeCards(false)
  autoRefreshTimer = window.setInterval(() => {
    refreshHomeCards(true)
  }, AUTO_REFRESH_INTERVAL)
})

onUnmounted(() => {
  if (autoRefreshTimer) {
    clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
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

.all-audit-list {
  min-height: 200px;
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
