<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>员工管理</span>
        <el-button type="primary" @click="openDialog()">新增员工</el-button>
      </div>
    </template>

    <el-form :inline="true" style="margin-bottom:16px">
      <el-form-item label="姓名">
        <el-input v-model="query.name" placeholder="请输入姓名" clearable />
      </el-form-item>
      <el-form-item label="部门">
        <el-select v-model="query.department_id" placeholder="全部" clearable style="width:160px">
          <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable style="width:100px">
          <el-option label="在职" :value="1" />
          <el-option label="离职" :value="0" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column label="序号" width="90">
        <template #default="{ row, $index }">
          <el-link type="primary" :underline="false" @click="openDetail(row.id)">
            {{ seqNo($index) }}
          </el-link>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="姓名" />
      <el-table-column prop="phone" label="电话" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column label="职位">
        <template #default="{ row }">{{ row.position_info?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="部门">
        <template #default="{ row }">{{ row.department?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '在职' : '离职' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="审批状态" width="100">
        <template #default="{ row }">
          <el-tag :type="approveTagType(row.approve_status)">
            {{ approveStatusLabel(row.approve_status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" :disabled="!isDraft(row)" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" :disabled="!isDraft(row)" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
      <el-table-column label="审核" width="180">
        <template #default="{ row }">
          <el-button size="small" type="success" :disabled="!isDraft(row)" @click="handleSubmitAudit(row)">提交审核</el-button>
          <el-button size="small" type="warning" :disabled="row.approve_status !== 'pending'" @click="handleWithdraw(row)">撤回</el-button>
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑员工' : '新增员工'" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="电话" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入11位手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱地址" />
        </el-form-item>
        <el-form-item label="部门" prop="department_id">
          <el-select v-model="form.department_id" placeholder="请选择部门" style="width:100%" @change="onDeptChange">
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="职位" prop="position_id">
          <el-select v-model="form.position_id" placeholder="请先选择部门" style="width:100%" :disabled="!form.department_id">
            <el-option v-for="p in filteredPositions" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="在职" :value="1" />
            <el-option label="离职" :value="0" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="detailVisible" title="员工详情" width="520px">
      <el-descriptions :column="1" border v-if="detailData">
        <el-descriptions-item label="ID">{{ detailData.id }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detailData.name }}</el-descriptions-item>
        <el-descriptions-item label="电话">{{ detailData.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ detailData.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="部门">{{ detailData.department?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="职位">{{ detailData.position_info?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detailData.status === 1 ? '在职' : '离职' }}</el-descriptions-item>
        <el-descriptions-item label="审批状态">{{ approveStatusLabel(detailData.approve_status) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button type="success" :disabled="!isDraft(detailData)" @click="handleSubmitAudit(detailData)">提交审核</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="flowVisible" :title="`审批流程记录 - ${flowTarget.name || ''}`" width="860px">
      <el-skeleton :loading="flowLoading" animated :rows="6">
        <template #default>
          <el-empty v-if="!(flowData.instances || []).length" description="暂无流程记录" :image-size="72" />
          <el-tabs v-else type="border-card">

            <!-- 当前流程 tab -->
            <el-tab-pane label="当前流程">
              <div style="margin-bottom:12px;display:flex;gap:12px;align-items:center">
                <el-tag :type="flowInstanceTagType(flowData.instance?.status)">
                  {{ flowInstanceStatusLabel(flowData.instance?.status) }}
                </el-tag>
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

            <!-- 历史提交 tab -->
            <el-tab-pane :label="`提交历史（${flowData.instances.length} 次）`">
              <el-table
                :data="flowHistoryTableData"
                border
                size="small"
                style="width:100%"
                row-key="id"
                :expand-row-keys="flowExpandedRows"
              >
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
                    <el-tag size="small" :type="flowInstanceTagType(row.instance.status)">
                      {{ flowInstanceStatusLabel(row.instance.status) }}
                    </el-tag>
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
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEmployees, getEmployee, createEmployee, updateEmployee, deleteEmployee, submitEmployee, withdrawEmployee } from '../api/employee'
import { getDepartments } from '../api/department'
import { getPositions } from '../api/position'
import { getOrchidWorkflowHistories } from '../api/orchid_workflow'

const list = ref([])
const total = ref(0)
const departments = ref([])
const allPositions = ref([])
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref(null)

const flowVisible = ref(false)
const flowLoading = ref(false)
const flowData = ref({ instance: null, histories: [], tasks: [] })
const flowTarget = ref({ id: null, name: '' })
const formRef = ref()
const form = ref({ name: '', phone: '', email: '', department_id: null, position_id: null, status: 1 })
const query = ref({ name: '', department_id: null, status: null, page: 1, page_size: 10 })

const filteredPositions = computed(() =>
  form.value.department_id
    ? allPositions.value.filter(p => p.department_id === form.value.department_id)
    : []
)

const rules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{
    validator: (rule, value, callback) => {
      if (value && !/^1[3-9]\d{9}$/.test(value)) callback(new Error('手机号格式不正确'))
      else callback()
    },
    trigger: 'blur'
  }],
  email: [{
    validator: (rule, value, callback) => {
      if (value && !/^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/.test(value)) callback(new Error('邮箱格式不正确'))
      else callback()
    },
    trigger: 'blur'
  }]
}

const approveStatusLabel = (s) => ({ draft: '草稿', pending: '待审批', approved: '已通过', rejected: '已拒绝' }[s] || '待审批')
const approveTagType = (s) => ({ draft: 'info', pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'warning')
const isDraft = (row) => row?.approve_status === 'draft'
const seqNo = (idx) => (query.value.page - 1) * query.value.page_size + idx + 1

const loadData = async () => {
  const res = await getEmployees(query.value)
  list.value = res.data.data.list || []
  total.value = res.data.data.total || 0
}

const handleSearch = () => { query.value.page = 1; loadData() }
const handleReset = () => {
  query.value = { name: '', department_id: null, status: null, page: 1, page_size: 10 }
  loadData()
}

const onDeptChange = () => {
  const ok = filteredPositions.value.some(p => p.id === form.value.position_id)
  if (!ok) form.value.position_id = null
}

const openDialog = (row = null) => {
  form.value = row
    ? {
        id: row.id,
        name: row.name,
        phone: row.phone,
        email: row.email,
        department_id: row.department_id,
        position_id: row.position_id,
        status: row.status,
      }
    : { name: '', phone: '', email: '', department_id: null, position_id: null, status: 1 }
  dialogVisible.value = true
}

const openDetail = async (id) => {
  const res = await getEmployee(id)
  detailData.value = res.data.data
  detailVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value.validate()
  if (form.value.id) await updateEmployee(form.value.id, form.value)
  else await createEmployee(form.value)
  ElMessage.success('操作成功')
  dialogVisible.value = false
  loadData()
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  await deleteEmployee(id)
  ElMessage.success('删除成功')
  loadData()
}

const handleSubmitAudit = async (row) => {
  if (!row?.id) return
  await ElMessageBox.confirm('确认提交审核？提交后将进入审批流程且不可编辑。', '提交审核', { type: 'warning' })
  await submitEmployee(row.id, {})
  ElMessage.success('已提交审核')
  if (detailVisible.value && detailData.value?.id === row.id) {
    const res = await getEmployee(row.id)
    detailData.value = res.data.data
  }
  loadData()
}

const handleWithdraw = async (row) => {
  if (!row?.id) return
  await ElMessageBox.confirm('确认撤回审核？撤回后将恢复为草稿状态，可重新编辑。', '撤回审核', { type: 'warning' })
  try {
    await withdrawEmployee(row.id)
    ElMessage.success('已撤回，恢复为草稿')
    if (detailVisible.value && detailData.value?.id === row.id) {
      const res = await getEmployee(row.id)
      detailData.value = res.data.data
    }
    loadData()
  } catch (e) {
    ElMessage.error(e?.response?.data?.msg || '撤回失败，可能已有节点审批通过')
  }
}

const flowInstanceStatusLabel = (s) => ({ pending: '审批中', approved: '已通过', rejected: '已拒绝', withdrawn: '已撤回' }[s] || '未开始')
const flowInstanceTagType = (s) => ({ pending: 'warning', approved: 'success', rejected: 'danger', withdrawn: 'info' }[s] || 'info')
const timelineItemType = (action) => ({ approved: 'success', rejected: 'danger', pending: 'warning', withdraw: 'info' }[action] || 'primary')

const actionLabel = (action) => ({
  submit: '发起申请',
  pending: '进入待办',
  approved: '审批通过',
  approved_partial: '部分通过',
  rejected: '审批拒绝',
  transfer: '转交',
  skip: '跳过',
  withdraw: '撤回',
}[action] || action)

const flowHistoryTableData = computed(() =>
  (flowData.value.instances || []).map((r, i) => ({ ...r, _idx: i }))
)
const flowExpandedRows = computed(() =>
  flowHistoryTableData.value.map(r => r.instance?.id?.toString())
)
const formatTime = (t) => t ? String(t).replace('T', ' ').slice(0, 19) : '-'

const timelineTitle = (h) => {
  const nodeName = nodeKeyToName(h.node_key)
  const map = {
    submit: '发起申请',
    pending: `进入节点：${nodeName}`,
    approved_partial: `部分通过：${nodeName}`,
    approved: `审批通过：${nodeName}`,
    rejected: `审批拒绝：${nodeName}`,
    transfer: `转交：${nodeName}`,
    skip: `跳过：${nodeName}`,
  }
  return map[h.action] || `${h.action || '处理'}：${nodeName}`
}

// 用于流程记录弹窗的姓名映射
const flowUserMap = ref({})   // { userId: name }
const flowNodeMap = ref({})   // { nodeKey: nodeName }

const nodeKeyToName = (key) => flowNodeMap.value[key] || key
const assigneeIdToName = (id) => flowUserMap.value[id] || `用户#${id}`

const openFlowDialog = async (row) => {
  flowTarget.value = { id: row.id, name: row.name }
  flowVisible.value = true
  flowLoading.value = true
  flowUserMap.value = {}
  flowNodeMap.value = {}
  try {
    const res = await getOrchidWorkflowHistories({ biz_type: 'employee', biz_id: row.id })
    const data = res.data?.data || {}
    flowData.value = {
      instance: data.instance || null,
      instances: data.instances || [],
      histories: data.histories || [],
      tasks: data.tasks || [],
    }

    // 构建员工 userId -> name 映射（用已加载的 list）
    const userMap = {}
    for (const emp of list.value) {
      if (emp.user_id) userMap[emp.user_id] = emp.name
    }
    // 如果 list 里没有，补充查一次全量
    if (!Object.keys(userMap).length) {
      const empRes = await getEmployees({ page: 1, page_size: 1000 })
      for (const emp of empRes.data?.data?.list || []) {
        if (emp.user_id) userMap[emp.user_id] = emp.name
      }
    }
    flowUserMap.value = userMap

    // 构建 nodeKey -> nodeName 映射（从流程定义 dag_json 解析）
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
  } catch (e) {
    flowData.value = { instance: null, histories: [], tasks: [] }
  } finally {
    flowLoading.value = false
  }
}

onMounted(async () => {
  const [deptRes, posRes] = await Promise.all([
    getDepartments({ page: 1, page_size: 100 }),
    getPositions({ page: 1, page_size: 100 })
  ])
  departments.value = deptRes.data.data.list || []
  allPositions.value = posRes.data.data.list || []
  loadData()
})
</script>
