<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="员工">
        <el-input v-model="query.employee_name" placeholder="输入姓名搜索" clearable style="width:140px" />
      </el-form-item>
      <el-form-item label="类型">
        <el-select v-model="query.type" placeholder="全部" clearable style="width:110px">
          <el-option v-for="t in leaveTypes" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable style="width:110px">
          <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="员工" width="100">
        <template #default="{ row }">{{ row.employee?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="请假类型" width="100">
        <template #default="{ row }">
          <el-tag :type="getLeaveTypeTag(row.type)">{{ getLeaveTypeLabel(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开始日期" width="120">
        <template #default="{ row }">{{ formatDate(row.start_date) }}</template>
      </el-table-column>
      <el-table-column label="结束日期" width="120">
        <template #default="{ row }">{{ formatDate(row.end_date) }}</template>
      </el-table-column>
      <el-table-column prop="days" label="天数" width="80">
        <template #default="{ row }">{{ row.days }} 天</template>
      </el-table-column>
      <el-table-column prop="reason" label="原因" min-width="120" show-overflow-tooltip />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusTag(row.status)">{{ getStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="approved_by" label="审批人" width="100" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">详情</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)" :disabled="row.status === 'approved' || row.status === 'pending'">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top:16px;justify-content:flex-end;display:flex"
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      :page-sizes="[10, 20, 50]"
      layout="total, sizes, prev, pager, next"
      @change="loadData"
    />

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '请假详情' : '新增请假'" width="860px" @closed="resetForm">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item label="员工" prop="employee_id">
          <el-select v-model="form.employee_id" placeholder="请选择员工" filterable style="width:100%">
            <el-option v-for="e in employeeList" :key="e.id"
              :label="`${e.name}（${e.department?.name || ''}）`" :value="e.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="请假类型" prop="type">
          <el-select v-model="form.type" style="width:100%">
            <el-option v-for="t in leaveTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始日期" prop="start_date">
          <el-date-picker v-model="form.start_date" type="date" placeholder="选择开始日期"
            format="YYYY-MM-DD" value-format="YYYY-MM-DD" style="width:100%"
            @change="calcDays" />
        </el-form-item>
        <el-form-item label="结束日期" prop="end_date">
          <el-date-picker v-model="form.end_date" type="date" placeholder="选择结束日期"
            format="YYYY-MM-DD" value-format="YYYY-MM-DD" style="width:100%"
            @change="calcDays" />
        </el-form-item>
        <el-form-item label="请假天数" prop="days">
          <el-input
            :value="calcLoading ? '计算中...' : (form.days ? `${form.days} 天` : '-')"
            readonly
            style="width:100%"
            placeholder="自动根据日期计算"
          />
        </el-form-item>
        <el-form-item label="请假原因">
          <el-input v-model="form.reason" type="textarea" :rows="3" placeholder="请输入请假原因（选填）" />
        </el-form-item>

        <el-form-item label="流程信息" v-if="form.id">
          <div style="width:100%">
            <div style="max-height:160px;overflow:auto;border:1px solid #ebeef5;border-radius:6px;padding:8px;margin-bottom:8px">
              <div v-if="!workflowLogs.length" style="color:#999">暂无流程记录</div>
              <div v-for="(log,i) in workflowLogs" :key="i" style="font-size:12px;line-height:1.7">
                {{ log.time }}｜{{ log.node }}｜{{ log.action }}｜{{ log.operator }}{{ log.remark ? `（${log.remark}）` : '' }}
              </div>
            </div>
            <div style="font-size:12px;color:#666;margin-bottom:4px">当前待办：</div>
            <div v-if="!openTasks.length" style="font-size:12px;color:#999">无</div>
            <div v-else style="display:flex;gap:6px;flex-wrap:wrap">
              <el-tag v-for="t in openTasks" :key="t.id" size="small" type="warning">
                {{ t.node_key }} / {{ getTaskAssigneeName(t.assignee_id) }}
              </el-tag>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button v-if="!form.id || form.status === 'draft'" type="primary" @click="handleSubmit">保存草稿</el-button>
        <el-button v-if="form.id && form.status === 'draft'" type="success" @click="handleSubmitFlow">提交</el-button>
        <el-button v-if="form.id && form.status === 'pending'" type="success" @click="openApprove(form, 'approved')">通过</el-button>
        <el-button v-if="form.id && form.status === 'pending'" type="warning" @click="openApprove(form, 'rejected')">拒绝</el-button>
        <el-button v-if="form.id && form.status === 'pending'" @click="openTransferDialog">转交</el-button>
        <el-button v-if="form.id && form.status === 'pending'" type="info" @click="handleSkipNode">跳过节点</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="transferVisible" title="转交待办" width="420px">
      <el-form :model="transferForm" label-width="90px">
        <el-form-item label="当前处理人">
          <el-select v-model="transferForm.from_user_id" style="width:100%" filterable>
            <el-option v-for="t in openTasks" :key="t.id" :label="getTaskAssigneeName(t.assignee_id)" :value="t.assignee_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="转交给">
          <el-select v-model="transferForm.to_user_id" style="width:100%" filterable>
            <el-option v-for="e in employeeList" :key="e.id" :label="`${e.name}（${e.department?.name || ''}）`" :value="e.user_id || 0" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="transferForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="transferVisible = false">取消</el-button>
        <el-button type="primary" @click="handleTransfer">确认转交</el-button>
      </template>
    </el-dialog>

    <!-- 审批对话框 -->
    <el-dialog v-model="approveVisible" :title="approveAction === 'approved' ? '审批通过' : '拒绝申请'" width="420px">
      <el-form :model="approveForm" label-width="90px">
        <el-form-item label="员工">{{ approveTarget?.employee?.name }}</el-form-item>
        <el-form-item label="请假类型">{{ getLeaveTypeLabel(approveTarget?.type) }}</el-form-item>
        <el-form-item label="请假天数">{{ approveTarget?.days }} 天</el-form-item>
        <el-form-item label="拒绝原因" v-if="approveAction === 'rejected'">
          <el-input v-model="approveForm.reject_reason" type="textarea" :rows="2" placeholder="请输入拒绝原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveVisible = false">取消</el-button>
        <el-button :type="approveAction === 'approved' ? 'success' : 'warning'" @click="handleApprove">
          确认{{ approveAction === 'approved' ? '通过' : '拒绝' }}
        </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getLeaveRequests, createLeaveRequest, updateLeaveRequest, submitLeaveRequest, approveLeaveRequest, deleteLeaveRequest } from '../api/leave_request'
import { getEmployees } from '../api/employee'
import { getOrchidWorkflowHistories, transferOrchidWorkflowTask, skipOrchidWorkflowNode } from '../api/orchid_workflow'

const leaveTypes = [
  { value: 'annual',   label: '年假',   tag: ''        },
  { value: 'sick',     label: '病假',   tag: 'warning' },
  { value: 'personal', label: '事假',   tag: 'info'    },
  { value: 'other',    label: '其他',   tag: 'info'    },
]
const statusOptions = [
  { value: 'draft',    label: '草稿',   tag: 'info'    },
  { value: 'pending',  label: '待审核', tag: 'warning' },
  { value: 'approved', label: '已通过', tag: 'success' },
  { value: 'rejected', label: '已拒绝', tag: 'danger'  },
]
const getLeaveTypeLabel = (v) => leaveTypes.find(t => t.value === v)?.label || v
const getLeaveTypeTag   = (v) => leaveTypes.find(t => t.value === v)?.tag  || ''
const getStatusLabel    = (v) => statusOptions.find(s => s.value === v)?.label || v
const getStatusTag      = (v) => statusOptions.find(s => s.value === v)?.tag   || ''
const formatDate = (t) => t ? t.slice(0, 10) : ''

const list = ref([])
const total = ref(0)
const employeeList = ref([])
const dialogVisible = ref(false)
const approveVisible = ref(false)
const transferVisible = ref(false)
const workflowLogs = ref([])
const openTasks = ref([])
const approveAction = ref('approved')
const approveTarget = ref(null)
const approveForm = ref({ reject_reason: '' })
const transferForm = ref({ from_user_id: 0, to_user_id: 0, remark: '' })
const formRef = ref()
const query = ref({ employee_name: null, type: null, status: null, page: 1, page_size: 10 })

const defaultForm = () => ({ employee_id: null, type: 'annual', start_date: '', end_date: '', days: 0, reason: '' })
const form = ref(defaultForm())

const calcLoading = ref(false)
// 节假日缓存，key 为年份
const holidayCache = {}

// 获取某年节假日数据（timor.tech 免费 API）
const fetchHolidays = async (year) => {
  if (holidayCache[year]) return holidayCache[year]
  try {
    const res = await fetch(`https://timor.tech/api/holiday/year/${year}`)
    const data = await res.json()
    // data.holiday 是对象，key 为 MM-dd，value.holiday=true 表示节假日，false 表示调休工作日
    holidayCache[year] = data.holiday || {}
  } catch {
    holidayCache[year] = {}
  }
  return holidayCache[year]
}

// 计算两个日期之间的工作日天数（跳过周末和节假日，保留调休工作日）
const calcWorkDays = async (startStr, endStr) => {
  const start = new Date(startStr)
  const end   = new Date(endStr)
  if (end < start) return 0

  // 收集涉及的年份
  const years = new Set()
  for (let y = start.getFullYear(); y <= end.getFullYear(); y++) years.add(y)
  const holidayMaps = {}
  for (const y of years) holidayMaps[y] = await fetchHolidays(y)

  let count = 0
  const cur = new Date(start)
  while (cur <= end) {
    const mm   = String(cur.getMonth() + 1).padStart(2, '0')
    const dd   = String(cur.getDate()).padStart(2, '0')
    const key  = `${mm}-${dd}`
    const year = cur.getFullYear()
    const map  = holidayMaps[year] || {}
    const dow  = cur.getDay() // 0=日,6=六

    if (map[key] !== undefined) {
      // API 有记录：holiday=true 是节假日休息，holiday=false 是调休上班
      if (!map[key].holiday) count++ // 调休工作日计入
    } else {
      // API 无记录：按自然周判断
      if (dow !== 0 && dow !== 6) count++
    }
    cur.setDate(cur.getDate() + 1)
  }
  return count
}

const calcDays = async () => {
  if (!form.value.start_date || !form.value.end_date) return
  if (form.value.end_date < form.value.start_date) {
    ElMessage.warning('结束日期不能早于开始日期')
    form.value.days = 0
    return
  }
  calcLoading.value = true
  try {
    form.value.days = await calcWorkDays(form.value.start_date, form.value.end_date)
  } finally {
    calcLoading.value = false
  }
}

const rules = {
  employee_id: [{ required: true, message: '请选择员工',     trigger: 'change' }],
  type:        [{ required: true, message: '请选择请假类型', trigger: 'change' }],
  start_date:  [{ required: true, message: '请选择开始日期', trigger: 'change' }],
  end_date:    [{ required: true, message: '请选择结束日期', trigger: 'change' }],
  days:        [{ required: true, message: '请输入请假天数', trigger: 'blur'   }],
}

const loadData = async () => {
  const params = { ...query.value }
  if (!params.employee_name) delete params.employee_name
  if (!params.type)        delete params.type
  if (!params.status)      delete params.status
  const res = await getLeaveRequests(params)
  list.value  = res.data.data.list  || []
  total.value = res.data.data.total || 0
}

const loadEmployees = async () => {
  const res = await getEmployees({ page: 1, page_size: 999 })
  employeeList.value = res.data.data.list || []
}

const getTaskAssigneeName = (userId) => {
  const emp = employeeList.value.find(e => e.user_id === userId)
  return emp ? emp.name : `用户#${userId}`
}

const loadWorkflowDetail = async (row) => {
  if (!row?.id) {
    workflowLogs.value = []
    openTasks.value = []
    return
  }
  const res = await getOrchidWorkflowHistories({ biz_type: 'leave_request', biz_id: row.id })
  const data = res.data.data || {}
  workflowLogs.value = (data.histories || []).map(h => ({
    time: h.created_at?.replace('T', ' ').slice(0, 19) || '',
    node: h.node_key,
    action: h.action,
    operator: h.operator,
    remark: h.remark
  }))
  openTasks.value = data.tasks || []
}

const handleSearch = () => { query.value.page = 1; loadData() }
const handleReset  = () => {
  query.value = { employee_name: null, type: null, status: null, page: 1, page_size: 10 }
  loadData()
}

const openDialog = async (row = null) => {
  form.value = row
    ? { ...row, start_date: formatDate(row.start_date), end_date: formatDate(row.end_date) }
    : defaultForm()
  workflowLogs.value = row?.workflow_logs ? JSON.parse(row.workflow_logs) : []
  await loadWorkflowDetail(row)
  dialogVisible.value = true
}

const resetForm = () => { formRef.value?.resetFields() }

const handleSubmit = async () => {
  await formRef.value.validate()
  if (form.value.id) {
    await updateLeaveRequest(form.value.id, form.value)
  } else {
    await createLeaveRequest(form.value)
  }
  ElMessage.success('已保存草稿')
  dialogVisible.value = false
  loadData()
}

const handleSubmitFlow = async () => {
  await submitLeaveRequest(form.value.id, { remark: '' })
  ElMessage.success('提交成功')
  await loadData()
  const latest = list.value.find(x => x.id === form.value.id)
  if (latest) form.value = { ...latest, start_date: formatDate(latest.start_date), end_date: formatDate(latest.end_date) }
  await loadWorkflowDetail(form.value)
}

const openApprove = (row, action) => {
  approveTarget.value = row
  approveAction.value = action
  approveForm.value   = { reject_reason: '' }
  approveVisible.value = true
}

const handleApprove = async () => {
  await approveLeaveRequest(approveTarget.value.id, {
    status: approveAction.value,
    reject_reason: approveForm.value.reject_reason
  })
  ElMessage.success(approveAction.value === 'approved' ? '已通过' : '已拒绝')
  approveVisible.value = false
  await loadData()
  await loadWorkflowDetail(form.value)
}

const openTransferDialog = () => {
  transferForm.value = { from_user_id: openTasks.value[0]?.assignee_id || 0, to_user_id: 0, remark: '' }
  transferVisible.value = true
}

const handleTransfer = async () => {
  await transferOrchidWorkflowTask(
    { biz_type: 'leave_request', biz_id: form.value.id },
    transferForm.value
  )
  ElMessage.success('转交成功')
  transferVisible.value = false
  await loadWorkflowDetail(form.value)
}

const handleSkipNode = async () => {
  await skipOrchidWorkflowNode(
    { biz_type: 'leave_request', biz_id: form.value.id },
    { remark: '手动跳过' }
  )
  ElMessage.success('已跳过当前节点')
  await loadData()
  await loadWorkflowDetail(form.value)
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该请假记录？', '提示', { type: 'warning' })
  await deleteLeaveRequest(id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadData(); loadEmployees() })
</script>
