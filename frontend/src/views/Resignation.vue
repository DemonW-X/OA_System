<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="员工">
        <el-input v-model="query.employee_name" placeholder="输入姓名搜索" clearable style="width:180px" />
      </el-form-item>
      <el-form-item label="审批状态">
        <el-select v-model="query.approve_status" placeholder="全部" clearable style="width:130px">
          <el-option label="草稿" value="draft" />
          <el-option label="待审批" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
      <el-form-item style="margin-left:auto;margin-right:0">
        <el-button type="primary" @click="openDialog()">新增离职</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="员工" min-width="120">
        <template #default="{ row }">{{ row.employee?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="部门" min-width="120">
        <template #default="{ row }">{{ row.employee?.department?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="职位" min-width="120">
        <template #default="{ row }">{{ row.employee?.position_info?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="离职日期" width="120">
        <template #default="{ row }">{{ formatDate(row.resign_date) }}</template>
      </el-table-column>
      <el-table-column prop="reason" label="离职原因" min-width="150" show-overflow-tooltip />
      <el-table-column prop="remark" label="备注" min-width="150" show-overflow-tooltip />
      <el-table-column label="员工状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.employee?.status === 0 ? 'info' : 'success'">
            {{ row.employee?.status === 0 ? '离职' : '在职' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="审批状态" width="100">
        <template #default="{ row }">
          <el-tag :type="approveTagType(row.approve_status)">{{ approveStatusLabel(row.approve_status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" :disabled="!isEditable(row)" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" :disabled="!isEditable(row)" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
      <el-table-column label="审核" width="280">
        <template #default="{ row }">
          <el-button size="small" type="success" :disabled="!isEditable(row)" @click="handleSubmitAudit(row)">提交审核</el-button>
          <el-button size="small" type="warning" :disabled="row.approve_status !== 'pending'" @click="handleWithdraw(row)">撤回</el-button>
          <el-button size="small" type="info" :disabled="row.approve_status !== 'approved'" @click="handleCancelApprove(row)">弃审</el-button>
        </template>
      </el-table-column>
      <el-table-column label="查看记录" width="120">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openFlowDialog(row)">流程记录</el-button>
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑离职记录' : '新增离职记录'" width="520px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="90px">
        <el-form-item label="员工" prop="employee_id">
          <el-select v-model="form.employee_id" placeholder="请选择员工" filterable style="width:100%">
            <el-option
              v-for="e in activeEmployeeOptions"
              :key="e.id"
              :label="`${e.name}（${e.department?.name || ''}）`"
              :value="e.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="离职日期" prop="resign_date">
          <el-date-picker
            v-model="form.resign_date"
            type="date"
            placeholder="请选择离职日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="离职原因">
          <el-input v-model="form.reason" type="textarea" :rows="3" placeholder="请输入离职原因（选填）" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="请输入备注（选填）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="flowVisible" :title="`审批流程记录 - ${flowTarget.name || ''}`" width="860px">
      <el-skeleton :loading="flowLoading" animated :rows="6">
        <template #default>
          <el-empty v-if="!(flowData.instances || []).length" description="暂无流程记录" :image-size="72" />
          <el-tabs v-else type="border-card">
            <el-tab-pane label="当前流程">
              <div style="margin-bottom:12px;display:flex;gap:12px;align-items:center">
                <el-tag :type="flowInstanceTagType(flowData.instance?.status)">{{ flowInstanceStatusLabel(flowData.instance?.status) }}</el-tag>
                <span style="color:#606266">当前节点：{{ nodeKeyToName(flowData.instance?.current_node) || '-' }}</span>
                <span style="color:#909399;font-size:12px">提交时间：{{ formatTime(flowData.instance?.started_at) }}</span>
              </div>

              <el-empty v-if="!(flowData.instances?.[flowData.instances.length-1]?.histories || []).length" description="暂无操作记录" :image-size="48" />
              <el-timeline v-else>
                <el-timeline-item
                  v-for="(h, idx) in flowData.instances[flowData.instances.length-1].histories"
                  :key="idx"
                  :timestamp="formatTime(h.created_at)"
                  :type="timelineItemType(h.action)"
                  placement="top"
                >
                  <div style="font-weight:600">{{ timelineTitle(h) }}</div>
                  <div style="color:#666;margin-top:4px">处理人：{{ h.operator || '-' }}</div>
                  <div v-if="h.remark" style="color:#909399;margin-top:4px">意见：{{ h.remark }}</div>
                </el-timeline-item>
              </el-timeline>

              <template v-if="(flowData.tasks || []).length">
                <el-divider content-position="left">当前待办</el-divider>
                <el-table :data="flowData.tasks" size="small" border>
                  <el-table-column label="节点" min-width="130">
                    <template #default="{ row }">{{ nodeKeyToName(row.node_key) }}</template>
                  </el-table-column>
                  <el-table-column label="待审批人" min-width="130">
                    <template #default="{ row }">{{ assigneeIdToName(row.assignee_id) }}</template>
                  </el-table-column>
                  <el-table-column label="状态" width="90">
                    <template #default="{ row }">
                      <el-tag size="small" type="warning">待处理</el-tag>
                    </template>
                  </el-table-column>
                </el-table>
              </template>
            </el-tab-pane>

            <el-tab-pane :label="`提交历史（${flowData.instances.length} 次）`">
              <el-table :data="flowHistoryTableData" border size="small" style="width:100%" row-key="id" :expand-row-keys="flowExpandedRows">
                <el-table-column type="expand">
                  <template #default="{ row }">
                    <div style="padding:8px 24px">
                      <el-table :data="row.histories" size="small" border>
                        <el-table-column label="时间" width="160" prop="created_at">
                          <template #default="{ row: h }">{{ formatTime(h.created_at) }}</template>
                        </el-table-column>
                        <el-table-column label="节点" min-width="120">
                          <template #default="{ row: h }">{{ nodeKeyToName(h.node_key) }}</template>
                        </el-table-column>
                        <el-table-column label="操作" width="110">
                          <template #default="{ row: h }">
                            <el-tag size="small" :type="timelineItemType(h.action)">{{ actionLabel(h.action) }}</el-tag>
                          </template>
                        </el-table-column>
                        <el-table-column label="处理人" width="110" prop="operator">
                          <template #default="{ row: h }">{{ h.operator || '-' }}</template>
                        </el-table-column>
                        <el-table-column label="意见" min-width="140" prop="remark">
                          <template #default="{ row: h }">{{ h.remark || '-' }}</template>
                        </el-table-column>
                      </el-table>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="次数" width="70">
                  <template #default="{ $index }">第 {{ $index + 1 }} 次</template>
                </el-table-column>
                <el-table-column label="提交时间" width="160">
                  <template #default="{ row }">{{ formatTime(row.instance.started_at) }}</template>
                </el-table-column>
                <el-table-column label="提交人" width="110">
                  <template #default="{ row }">{{ row.instance.started_by || '-' }}</template>
                </el-table-column>
                <el-table-column label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag size="small" :type="flowInstanceTagType(row.instance.status)">{{ flowInstanceStatusLabel(row.instance.status) }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作步骤数" width="100">
                  <template #default="{ row }">{{ (row.histories || []).length }} 步</template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </template>
      </el-skeleton>
      <template #footer>
        <el-button @click="flowVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEmployees } from '../api/employee'
import {
  getResignations, createResignation, updateResignation, deleteResignation,
  submitResignation, withdrawResignation, cancelApproveResignation
} from '../api/resignation'
import { getOrchidWorkflowHistories } from '../api/orchid_workflow'

const list = ref([])
const total = ref(0)
const employeeList = ref([])
const dialogVisible = ref(false)
const formRef = ref()
const query = ref({ employee_name: null, approve_status: null, page: 1, page_size: 10 })

const flowVisible = ref(false)
const flowLoading = ref(false)
const flowData = ref({ instance: null, histories: [], tasks: [], instances: [] })
const flowTarget = ref({ id: null, name: '' })
const flowUserMap = ref({})
const flowNodeMap = ref({})

const defaultForm = () => ({ employee_id: null, resign_date: '', reason: '', remark: '' })
const form = ref(defaultForm())

const rules = {
  employee_id: [{ required: true, message: '请选择员工', trigger: 'change' }],
  resign_date: [{ required: true, message: '请选择离职日期', trigger: 'change' }],
}

const approveStatusLabel = (s) => ({ draft: '草稿', pending: '待审批', approved: '已通过', rejected: '已拒绝' }[s] || '草稿')
const approveTagType = (s) => ({ draft: 'info', pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info')
const isEditable = (row) => ['draft', 'rejected'].includes(row?.approve_status)

const formatDate = (v) => (v ? String(v).slice(0, 10) : '-')
const formatTime = (t) => t ? String(t).replace('T', ' ').slice(0, 19) : '-'

const activeEmployeeOptions = computed(() => {
  if (form.value.id) return employeeList.value
  return employeeList.value.filter(e => e.status === 1)
})

const loadEmployees = async () => {
  const res = await getEmployees({ page: 1, page_size: 1000 })
  employeeList.value = res.data?.data?.list || []
}

const loadData = async () => {
  const params = { ...query.value }
  if (!params.employee_name) delete params.employee_name
  if (!params.approve_status) delete params.approve_status
  const res = await getResignations(params)
  list.value = res.data?.data?.list || []
  total.value = res.data?.data?.total || 0
}

const handleSearch = () => {
  query.value.page = 1
  loadData()
}

const handleReset = () => {
  query.value = { employee_name: null, approve_status: null, page: 1, page_size: 10 }
  loadData()
}

const openDialog = (row = null) => {
  if (row) {
    form.value = {
      id: row.id,
      employee_id: row.employee_id,
      resign_date: formatDate(row.resign_date),
      reason: row.reason || '',
      remark: row.remark || '',
    }
  } else {
    form.value = defaultForm()
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value.validate()
  const payload = {
    employee_id: form.value.employee_id,
    resign_date: form.value.resign_date,
    reason: form.value.reason,
    remark: form.value.remark,
  }
  if (form.value.id) {
    await updateResignation(form.value.id, payload)
  } else {
    await createResignation(payload)
  }
  ElMessage.success('操作成功')
  dialogVisible.value = false
  await Promise.all([loadData(), loadEmployees()])
}

const handleSubmitAudit = async (row) => {
  await ElMessageBox.confirm('确认提交审核？提交后将进入审批流程且不可编辑。', '提交审核', { type: 'warning' })
  await submitResignation(row.id, {})
  ElMessage.success('已提交审核')
  loadData()
}

const handleWithdraw = async (row) => {
  await ElMessageBox.confirm('确认撤回审核？撤回后将恢复为草稿状态。', '撤回审核', { type: 'warning' })
  await withdrawResignation(row.id)
  ElMessage.success('已撤回')
  await Promise.all([loadData(), loadEmployees()])
}

const handleCancelApprove = async (row) => {
  await ElMessageBox.confirm('确认取消审核？操作后单据将恢复为草稿状态。', '取消审核', { type: 'warning' })
  await cancelApproveResignation(row.id)
  ElMessage.success('已取消审核，恢复为草稿')
  await Promise.all([loadData(), loadEmployees()])
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该离职记录？删除后可能恢复员工为在职。', '提示', { type: 'warning' })
  await deleteResignation(id)
  ElMessage.success('删除成功')
  await Promise.all([loadData(), loadEmployees()])
}

const flowInstanceStatusLabel = (s) => ({ pending: '审批中', approved: '已通过', rejected: '已拒绝', withdrawn: '已撤回' }[s] || '未开始')
const flowInstanceTagType = (s) => ({ pending: 'warning', approved: 'success', rejected: 'danger', withdrawn: 'info' }[s] || 'info')
const timelineItemType = (action) => ({ approved: 'success', rejected: 'danger', pending: 'warning', withdraw: 'info' }[action] || 'primary')

const actionLabel = (action) => ({
  submit: '发起申请', pending: '进入待办', approved: '审批通过', approved_partial: '部分通过',
  rejected: '审批拒绝', transfer: '转交', skip: '跳过', withdraw: '撤回',
}[action] || action)

const flowHistoryTableData = computed(() => (flowData.value.instances || []).map((r, i) => ({ ...r, _idx: i })))
const flowExpandedRows = computed(() => flowHistoryTableData.value.map(r => r.instance?.id?.toString()))

const nodeKeyToName = (key) => flowNodeMap.value[key] || key
const assigneeIdToName = (id) => flowUserMap.value[id] || `用户#${id}`

const timelineTitle = (h) => {
  const nodeName = nodeKeyToName(h.node_key)
  const map = {
    submit: '发起申请', pending: `进入节点：${nodeName}`, approved_partial: `部分通过：${nodeName}`,
    approved: `审批通过：${nodeName}`, rejected: `审批拒绝：${nodeName}`, transfer: `转交：${nodeName}`,
    skip: `跳过：${nodeName}`,
  }
  return map[h.action] || `${h.action || '处理'}：${nodeName}`
}

const openFlowDialog = async (row) => {
  flowTarget.value = { id: row.id, name: row.employee?.name || `#${row.id}` }
  flowVisible.value = true
  flowLoading.value = true
  flowUserMap.value = {}
  flowNodeMap.value = {}
  try {
    const res = await getOrchidWorkflowHistories({ biz_type: 'resignation', biz_id: row.id })
    const data = res.data?.data || {}
    flowData.value = {
      instance: data.instance || null,
      instances: data.instances || [],
      histories: data.histories || [],
      tasks: data.tasks || [],
    }

    const userMap = {}
    for (const emp of employeeList.value) {
      if (emp.user_id) userMap[emp.user_id] = emp.name
    }
    flowUserMap.value = userMap

    if (data.instance?.definition_id) {
      try {
        const { getOrchidWorkflowDefinition } = await import('../api/orchid_workflow')
        const defRes = await getOrchidWorkflowDefinition(data.instance.definition_id)
        const dag = JSON.parse(defRes.data?.data?.dag_json || '{}')
        const nodeMap = {}
        for (const [key, node] of Object.entries(dag.nodes || {})) {
          nodeMap[key] = node?.config?._ui?.name || key
        }
        flowNodeMap.value = nodeMap
      } catch {}
    }
  } finally {
    flowLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadEmployees(), loadData()])
})
</script>
