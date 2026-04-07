<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="标题">
        <el-input v-model="query.title" placeholder="请输入标题" clearable />
      </el-form-item>
      <el-form-item label="类型">
        <el-select v-model="query.type" placeholder="全部" clearable style="width:110px">
          <el-option v-for="t in eventTypes" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="日期">
        <el-date-picker
          v-model="query.date"
          type="date"
          placeholder="选择日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          clearable
          style="width:150px"
        />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" min-width="130" />
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag :color="getTypeColor(row.type)" style="color:#fff;border:none">
            {{ getTypeLabel(row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="开始时间" width="160">
        <template #default="{ row }">{{ formatTime(row.start_time) }}</template>
      </el-table-column>
      <el-table-column label="结束时间" width="160">
        <template #default="{ row }">{{ formatTime(row.end_time) }}</template>
      </el-table-column>
      <el-table-column label="会议室" width="120">
        <template #default="{ row }">{{ row.meeting_room?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="参与人员" min-width="160">
        <template #default="{ row }">
          <span v-if="row.participants && JSON.parse(row.participants).length">
            <el-tag
              v-for="id in JSON.parse(row.participants)"
              :key="id"
              size="small"
              style="margin:2px"
            >{{ getEmployeeName(id) }}</el-tag>
          </span>
          <span v-else style="color:#999">-</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusTag(row.status)">{{ getStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_by" label="创建人" width="100" />
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

    <!-- 新增/详情对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '预定详情' : '新增预定'" width="920px" @closed="resetForm">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item label="事件标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入事件标题" />
        </el-form-item>
        <el-form-item label="事件类型" prop="type">
          <el-select v-model="form.type" style="width:100%">
            <el-option v-for="t in eventTypes" :key="t.value" :label="t.label" :value="t.value">
              <span style="display:flex;align-items:center;gap:8px">
                <span :style="{width:'10px',height:'10px',borderRadius:'50%',background:t.color,display:'inline-block'}" />
                {{ t.label }}
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间" prop="start_time">
          <el-date-picker
            v-model="form.start_time"
            type="datetime"
            placeholder="选择开始时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="结束时间" prop="end_time">
          <el-date-picker
            v-model="form.end_time"
            type="datetime"
            placeholder="选择结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="会议室">
          <el-select v-model="form.meeting_room_id" placeholder="不需要会议室" clearable style="width:100%">
            <el-option
              v-for="r in availableRooms"
              :key="r.id"
              :label="`${r.name}（${r.location || ''}，${r.capacity}人）`"
              :value="r.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="参与人员">
          <el-select
            v-model="form.participant_ids"
            multiple
            filterable
            placeholder="请选择参与人员（可多选）"
            style="width:100%"
          >
            <el-option
              v-for="emp in employeeList"
              :key="emp.id"
              :label="`${emp.name}（${emp.department?.name || ''}）`"
              :value="emp.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="事件描述（选填）" />
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

    <el-dialog v-model="approveVisible" :title="approveAction === 'approved' ? '审批通过' : '拒绝申请'" width="420px">
      <el-form :model="approveForm" label-width="90px">
        <el-form-item label="标题">{{ approveTarget?.title }}</el-form-item>
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
import { getEventBookings, createEventBooking, updateEventBooking, submitEventBooking, approveEventBooking, deleteEventBooking } from '../api/event_booking'
import { getMeetingRooms } from '../api/meeting_room'
import { getEmployees } from '../api/employee'
import { getOrchidWorkflowHistories, transferOrchidWorkflowTask, skipOrchidWorkflowNode } from '../api/orchid_workflow'

const eventTypes = [
  { value: 'meeting',  label: '会议',   color: '#409EFF' },
  { value: 'activity', label: '活动',   color: '#67C23A' },
  { value: 'other',    label: '其他',   color: '#909399' },
]
const statusOptions = [
  { value: 'draft', label: '草稿', tag: 'info' },
  { value: 'pending', label: '待审核', tag: 'warning' },
  { value: 'approved', label: '已通过', tag: 'success' },
  { value: 'rejected', label: '已拒绝', tag: 'danger' },
]
const getTypeColor = (type) => eventTypes.find(t => t.value === type)?.color || '#909399'
const getTypeLabel = (type) => eventTypes.find(t => t.value === type)?.label || '其他'
const getStatusLabel = (s) => statusOptions.find(x => x.value === s)?.label || s
const getStatusTag = (s) => statusOptions.find(x => x.value === s)?.tag || 'info'
const getEmployeeName = (id) => employeeList.value.find(e => e.id === id)?.name || id
const formatTime = (t) => t ? t.replace('T', ' ').slice(0, 19) : ''

const list = ref([])
const total = ref(0)
const availableRooms = ref([])
const employeeList = ref([])
const dialogVisible = ref(false)
const approveVisible = ref(false)
const transferVisible = ref(false)
const approveAction = ref('approved')
const approveTarget = ref(null)
const approveForm = ref({ reject_reason: '' })
const transferForm = ref({ from_user_id: 0, to_user_id: 0, remark: '' })
const workflowLogs = ref([])
const openTasks = ref([])
const formRef = ref()
const query = ref({ title: '', type: null, date: null, page: 1, page_size: 10 })

const defaultForm = () => ({
  title: '', type: 'other', start_time: '', end_time: '',
  meeting_room_id: null, participant_ids: [], description: '', status: 'draft'
})
const form = ref(defaultForm())

const rules = {
  title:      [{ required: true, message: '请输入事件标题', trigger: 'blur' }],
  type:       [{ required: true, message: '请选择事件类型', trigger: 'change' }],
  start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_time:   [{ required: true, message: '请选择结束时间', trigger: 'change' }],
}

const loadData = async () => {
  const params = { ...query.value }
  if (!params.type) delete params.type
  if (!params.date) delete params.date
  const res = await getEventBookings(params)
  list.value = res.data.data.list || []
  total.value = res.data.data.total || 0
}

const loadRooms = async () => {
  const res = await getMeetingRooms({ page: 1, page_size: 999, status: 1 })
  availableRooms.value = res.data.data.list || []
}

const loadEmployees = async () => {
  const res = await getEmployees({ page: 1, page_size: 999, status: 1 })
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
  const res = await getOrchidWorkflowHistories({ biz_type: 'event_booking', biz_id: row.id })
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
const handleReset = () => {
  query.value = { title: '', type: null, date: null, page: 1, page_size: 10 }
  loadData()
}

const openDialog = async (row = null) => {
  if (row) {
    const participantIds = row.participants ? JSON.parse(row.participants) : []
    form.value = {
      ...row,
      start_time: formatTime(row.start_time),
      end_time: formatTime(row.end_time),
      meeting_room_id: row.meeting_room_id || null,
      participant_ids: participantIds,
      status: row.status || 'draft'
    }
    workflowLogs.value = row.workflow_logs ? JSON.parse(row.workflow_logs) : []
    await loadWorkflowDetail(row)
  } else {
    form.value = defaultForm()
    workflowLogs.value = []
    openTasks.value = []
  }
  dialogVisible.value = true
}

const resetForm = () => { formRef.value?.resetFields() }

const handleSubmit = async () => {
  await formRef.value.validate()
  const payload = {
    ...form.value,
    meeting_room_id: form.value.meeting_room_id || 0,
    participants: form.value.participant_ids?.length
      ? JSON.stringify(form.value.participant_ids) : ''
  }
  if (form.value.id) {
    await updateEventBooking(form.value.id, payload)
  } else {
    await createEventBooking(payload)
  }
  ElMessage.success('已保存草稿')
  dialogVisible.value = false
  loadData()
}

const handleSubmitFlow = async () => {
  await submitEventBooking(form.value.id, { remark: '' })
  ElMessage.success('提交成功')
  await loadData()
  const latest = list.value.find(x => x.id === form.value.id)
  if (latest) {
    const participantIds = latest.participants ? JSON.parse(latest.participants) : []
    form.value = {
      ...latest,
      start_time: formatTime(latest.start_time),
      end_time: formatTime(latest.end_time),
      meeting_room_id: latest.meeting_room_id || null,
      participant_ids: participantIds,
      status: latest.status || 'draft'
    }
  }
  await loadWorkflowDetail(form.value)
}

const openApprove = (row, action) => {
  approveTarget.value = row
  approveAction.value = action
  approveForm.value = { reject_reason: '' }
  approveVisible.value = true
}

const handleApprove = async () => {
  await approveEventBooking(approveTarget.value.id, {
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
    { biz_type: 'event_booking', biz_id: form.value.id },
    transferForm.value
  )
  ElMessage.success('转交成功')
  transferVisible.value = false
  await loadWorkflowDetail(form.value)
}

const handleSkipNode = async () => {
  await skipOrchidWorkflowNode(
    { biz_type: 'event_booking', biz_id: form.value.id },
    { remark: '手动跳过' }
  )
  ElMessage.success('已跳过当前节点')
  await loadData()
  await loadWorkflowDetail(form.value)
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该预定？', '提示', { type: 'warning' })
  await deleteEventBooking(id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadData(); loadRooms(); loadEmployees() })
</script>
