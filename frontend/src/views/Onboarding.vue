<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>入职管理</span>
        <el-button type="primary" @click="openDialog()">新增入职</el-button>
      </div>
    </template>

    <el-form :inline="true" style="margin-bottom:16px">
      <el-form-item label="员工">
        <el-input v-model="query.employee_name" placeholder="输入姓名搜索" clearable style="width:180px" />
      </el-form-item>
      <el-form-item label="入职类型">
        <el-select v-model="query.onboard_type" placeholder="全部" clearable style="width:130px">
          <el-option label="新员工" value="new" />
          <el-option label="返聘" value="rehire" />
          <el-option label="调入" value="transfer" />
        </el-select>
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
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="employee_name" label="员工姓名" min-width="100" />
      <el-table-column label="入职日期" width="110">
        <template #default="{ row }">{{ formatDate(row.onboard_date) }}</template>
      </el-table-column>
      <el-table-column label="入职类型" width="90">
        <template #default="{ row }">
          <el-tag :type="onboardTypeTag(row.onboard_type)" size="small">{{ onboardTypeLabel(row.onboard_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="联系电话" width="130">
        <template #default="{ row }">{{ maskPhone(row.phone) }}</template>
      </el-table-column>
      <el-table-column prop="id_card" label="身份证号" width="180" show-overflow-tooltip />
      <el-table-column prop="native_place" label="籍贯" width="100" show-overflow-tooltip />
      <el-table-column label="试用天数" width="85" prop="probation_days" />
      <el-table-column label="试用截止" width="110">
        <template #default="{ row }">{{ formatDate(row.probation_end) }}</template>
      </el-table-column>
      <el-table-column label="部门" width="110">
        <template #default="{ row }">{{ departmentName(row.department_id) }}</template>
      </el-table-column>
      <el-table-column label="职位" width="110">
        <template #default="{ row }">{{ positionName(row.position_id) }}</template>
      </el-table-column>
      <el-table-column label="审批状态" width="95">
        <template #default="{ row }">
          <el-tag :type="approveTagType(row.approve_status)">{{ approveStatusLabel(row.approve_status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openDetail(row)">详情</el-button>
          <el-button size="small" :disabled="!isEditable(row)" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" :disabled="!isDeletable(row)" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
      <el-table-column label="审核" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="success" :disabled="!isEditable(row)" @click="handleSubmitAudit(row)">提交审核</el-button>
          <el-button size="small" type="warning" :disabled="row.approve_status !== 'pending'" @click="handleWithdraw(row)">撤回</el-button>
          <el-button size="small" type="info" :disabled="row.approve_status !== 'approved'" @click="handleCancelApprove(row)">弃审</el-button>
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

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑入职记录' : '新增入职记录'" width="680px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-divider content-position="left">基本信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="员工姓名" prop="employee_name">
              <el-input v-model="form.employee_name" placeholder="请输入员工姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="入职日期" prop="onboard_date">
              <el-date-picker v-model="form.onboard_date" type="date" placeholder="请选择" format="YYYY-MM-DD" value-format="YYYY-MM-DD" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="入职类型" prop="onboard_type">
              <el-select v-model="form.onboard_type" style="width:100%">
                <el-option label="新员工" value="new" />
                <el-option label="返聘" value="rehire" />
                <el-option label="调入" value="transfer" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="试用天数" prop="probation_days">
              <el-input-number v-model="form.probation_days" :min="0" :max="365" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">个人信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="联系电话" prop="phone">
              <el-input
                :model-value="maskedPhone"
                placeholder="请输入手机号"
                maxlength="11"
                @input="onPhoneInput"
                @focus="phoneFocused = true"
                @blur="phoneFocused = false"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="电子邮箱" prop="email">
              <el-input v-model="form.email" placeholder="请输入邮箱" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="身份证号" prop="id_card">
              <el-input
                :model-value="maskedIdCard"
                placeholder="请输入身份证号"
                maxlength="18"
                @input="onIdCardInput"
                @focus="idCardFocused = true"
                @blur="idCardFocused = false"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="籍贯">
              <el-input v-model="form.native_place" placeholder="如：湖南长沙" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="现居地址">
              <el-input v-model="form.address" placeholder="请输入现居住地址" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="紧急联系人">
              <el-input v-model="form.emergency_name" placeholder="请输入姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="紧急电话">
              <el-input v-model="form.emergency_phone" placeholder="请输入电话" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">教育与经历</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="学历">
              <el-select v-model="form.education" placeholder="请选择" clearable style="width:100%">
                <el-option label="初中及以下" value="junior" />
                <el-option label="高中/中专" value="high" />
                <el-option label="大专" value="college" />
                <el-option label="本科" value="bachelor" />
                <el-option label="硕士" value="master" />
                <el-option label="博士" value="doctor" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="工作年限">
              <el-input-number v-model="form.work_years" :min="0" :max="50" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="毕业院校">
              <el-input v-model="form.school" placeholder="请输入毕业院校" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="专业">
              <el-input v-model="form.major" placeholder="请输入专业" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="入职部门">
              <el-select v-model="form.department_id" placeholder="请选择部门" clearable filterable style="width:100%" @change="onDepartmentChange">
                <el-option v-for="d in departmentList" :key="d.id" :label="d.name" :value="d.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="入职职位">
              <el-select v-model="form.position_id" placeholder="请先选择部门" clearable filterable style="width:100%" :disabled="!form.department_id">
                <el-option v-for="p in filteredPositionList" :key="p.id" :label="p.name" :value="p.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="请输入备注（选填）" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :disabled="submitting" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="入职详情" width="640px">
      <el-descriptions :column="2" border v-if="detailData">
        <el-descriptions-item label="员工姓名">{{ detailData.employee_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="入职日期">{{ formatDate(detailData.onboard_date) }}</el-descriptions-item>
        <el-descriptions-item label="入职类型">
          <el-tag :type="onboardTypeTag(detailData.onboard_type)" size="small">{{ onboardTypeLabel(detailData.onboard_type) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="试用天数">{{ detailData.probation_days }} 天</el-descriptions-item>
        <el-descriptions-item label="试用截止">{{ formatDate(detailData.probation_end) }}</el-descriptions-item>
        <el-descriptions-item label="审批状态">
          <el-tag :type="approveTagType(detailData.approve_status)">{{ approveStatusLabel(detailData.approve_status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="入职部门">{{ departmentName(detailData.department_id) }}</el-descriptions-item>
        <el-descriptions-item label="入职职位">{{ positionName(detailData.position_id) }}</el-descriptions-item>
        <el-descriptions-item label="审批人">{{ detailData.approved_by || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审批意见">{{ detailData.approve_remark || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ maskPhone(detailData.phone) }}</el-descriptions-item>
        <el-descriptions-item label="电子邮箱">{{ detailData.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="身份证号" :span="2">{{ maskIdCard(detailData.id_card) }}</el-descriptions-item>
        <el-descriptions-item label="籍贯">{{ detailData.native_place || '-' }}</el-descriptions-item>
        <el-descriptions-item label="工作年限">{{ detailData.work_years ?? 0 }} 年</el-descriptions-item>
        <el-descriptions-item label="现居地址" :span="2">{{ detailData.address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="紧急联系人">{{ detailData.emergency_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="紧急联系电话">{{ detailData.emergency_phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="学历">{{ educationLabel(detailData.education) }}</el-descriptions-item>
        <el-descriptions-item label="毕业院校">{{ detailData.school || '-' }}</el-descriptions-item>
        <el-descriptions-item label="专业">{{ detailData.major || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ detailData.remark || '-' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { phoneRules, emailRules, idCardRules } from '../utils/validators'
import {
  getOnboardings, getOnboarding, createOnboarding, updateOnboarding, deleteOnboarding,
  submitOnboarding, withdrawOnboarding, cancelApproveOnboarding
} from '../api/onboarding'
import { getDepartments } from '../api/department'
import { getPositions } from '../api/position'
import { getDepartmentPositions } from '../api/department_position'

const list = ref([])
const total = ref(0)
const dialogVisible = ref(false)
const detailVisible = ref(false)
const detailData = ref(null)
const formRef = ref()
const query = ref({ employee_name: null, onboard_type: null, approve_status: null, page: 1, page_size: 10 })
const departmentList = ref([])
const positionList = ref([])
const filteredPositionList = ref([])

const departmentName = (id) => departmentList.value.find(d => d.id === id)?.name || '-'
const positionName = (id) => positionList.value.find(p => p.id === id)?.name || '-'

const onDepartmentChange = async (deptId) => {
  form.value.position_id = null
  if (!deptId) {
    filteredPositionList.value = []
    return
  }
  const res = await getDepartmentPositions({ department_id: deptId, page: 1, page_size: 999 })
  const relations = res.data?.data?.list || []
  filteredPositionList.value = relations.map(r => ({ id: r.position_id, name: r.position?.name || '' }))
}

const defaultForm = () => ({
  employee_name: '', onboard_date: '', onboard_type: 'new', probation_days: 90,
  phone: '', email: '', id_card: '', native_place: '', address: '',
  emergency_name: '', emergency_phone: '', education: '', school: '', major: '',
  work_years: 0, department_id: null, position_id: null, remark: ''
})
const form = ref(defaultForm())

const rules = {
  employee_name: [{ required: true, message: '请输入员工姓名', trigger: 'blur' }],
  onboard_date: [{ required: true, message: '请选择入职日期', trigger: 'change' }],
  onboard_type: [{ required: true, message: '请选择入职类型', trigger: 'change' }],
  phone: phoneRules,
  email: emailRules,
  id_card: idCardRules,
}

const formatDate = (v) => (v ? String(v).slice(0, 10) : '-')
const onboardTypeLabel = (t) => ({ new: '新员工', rehire: '返聘', transfer: '调入' }[t] || t)
const onboardTypeTag = (t) => ({ new: 'primary', rehire: 'warning', transfer: 'success' }[t] || 'info')
const approveStatusLabel = (s) => ({ draft: '草稿', pending: '待审批', approved: '已通过', rejected: '已拒绝' }[s] || '草稿')
const approveTagType = (s) => ({ draft: 'info', pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info')
const educationLabel = (e) => ({ junior: '初中及以下', high: '高中/中专', college: '大专', bachelor: '本科', master: '硕士', doctor: '博士' }[e] || (e || '-'))
const isEditable = (row) => ['draft', 'rejected'].includes(row?.approve_status)
const isDeletable = (row) => ['draft', 'rejected'].includes(row?.approve_status)

// 电话号码脱敏工具函数（第4-7位显示为*）
const maskPhone = (val) => {
  if (!val) return '-'
  if (val.length <= 3) return val
  const end = Math.min(7, val.length)
  return val.slice(0, 3) + '*'.repeat(end - 3) + val.slice(end)
}

// 电话号码脱敏显示（表单输入框）
const phoneFocused = ref(false)
const maskedPhone = computed(() => {
  const val = form.value.phone || ''
  if (!val || phoneFocused.value) return val
  if (val.length <= 3) return val
  const end = Math.min(7, val.length)
  return val.slice(0, 3) + '*'.repeat(end - 3) + val.slice(end)
})
const onPhoneInput = (val) => {
  form.value.phone = val
}

// 身份证号脱敏显示（第7-14位显示为*）
const idCardFocused = ref(false)
const maskedIdCard = computed(() => {
  const val = form.value.id_card || ''
  if (!val || idCardFocused.value) return val
  if (val.length <= 6) return val
  const end = Math.min(14, val.length)
  return val.slice(0, 6) + '*'.repeat(end - 6) + val.slice(end)
})
const onIdCardInput = (val) => {
  form.value.id_card = val
}

// 身份证号脱敏工具函数（第7-14位显示为*）
const maskIdCard = (val) => {
  if (!val) return '-'
  if (val.length <= 6) return val
  const end = Math.min(14, val.length)
  return val.slice(0, 6) + '*'.repeat(end - 6) + val.slice(end)
}

const loadData = async () => {
  const params = { ...query.value }
  if (!params.employee_name) delete params.employee_name
  if (!params.onboard_type) delete params.onboard_type
  if (!params.approve_status) delete params.approve_status
  const res = await getOnboardings(params)
  list.value = res.data?.data?.list || []
  total.value = res.data?.data?.total || 0
}

const handleSearch = () => { query.value.page = 1; loadData() }
const handleReset = () => {
  query.value = { employee_name: null, onboard_type: null, approve_status: null, page: 1, page_size: 10 }
  loadData()
}

const openDialog = async (row = null) => {
  if (row) {
    form.value = {
      id: row.id,
      employee_name: row.employee_name || '',
      onboard_date: formatDate(row.onboard_date),
      onboard_type: row.onboard_type || 'new',
      probation_days: row.probation_days || 90,
      phone: row.phone || '',
      email: row.email || '',
      id_card: row.id_card || '',
      native_place: row.native_place || '',
      address: row.address || '',
      emergency_name: row.emergency_name || '',
      emergency_phone: row.emergency_phone || '',
      education: row.education || '',
      school: row.school || '',
      major: row.major || '',
      work_years: row.work_years || 0,
      department_id: row.department_id || null,
      position_id: row.position_id || null,
      remark: row.remark || '',
    }
  } else {
    form.value = defaultForm()
    filteredPositionList.value = []
  }
  // 编辑时预加载部门对应职位
  if (form.value.department_id) {
    const res = await getDepartmentPositions({ department_id: form.value.department_id, page: 1, page_size: 999 })
    const relations = res.data?.data?.list || []
    filteredPositionList.value = relations.map(r => ({ id: r.position_id, name: r.position?.name || '' }))
  }
  dialogVisible.value = true
}

const openDetail = async (row) => {
  const res = await getOnboarding(row.id)
  detailData.value = res.data?.data
  detailVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value.validate()
  const payload = { ...form.value }
  const id = payload.id
  delete payload.id
  // null 转 0，避免 Go int 类型解析失败
  if (!payload.department_id) payload.department_id = 0
  if (!payload.position_id) payload.position_id = 0
    if (id) {
      await updateOnboarding(id, payload)
    } else {
      await createOnboarding(payload)
    }
    ElMessage.success('操作成功')
    dialogVisible.value = false
    loadData()
}

const handleSubmitAudit = async (row) => {
  await ElMessageBox.confirm('确认提交审核？提交后将进入审批流程且不可编辑。', '提交审核', { type: 'warning' })
  await submitOnboarding(row.id)
  ElMessage.success('已提交审核')
  loadData()
}

const handleWithdraw = async (row) => {
  await ElMessageBox.confirm('确认撤回审核？撤回后将恢复为草稿状态。', '撤回审核', { type: 'warning' })
  await withdrawOnboarding(row.id)
  ElMessage.success('已撤回')
  loadData()
}

const handleCancelApprove = async (row) => {
  await ElMessageBox.confirm('确认取消审核？操作后单据将恢复为草稿状态。', '取消审核', { type: 'warning' })
  await cancelApproveOnboarding(row.id)
  ElMessage.success('已取消审核，恢复为草稿')
  loadData()
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该入职记录？', '提示', { type: 'warning' })
  await deleteOnboarding(id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(async () => {
  const [deptRes, posRes] = await Promise.all([
    getDepartments({ page: 1, page_size: 999 }),
    getPositions({ page: 1, page_size: 999 })
  ])
  departmentList.value = deptRes.data?.data?.list || []
  positionList.value = posRes.data?.data?.list || []
  await loadData()
})
</script>
