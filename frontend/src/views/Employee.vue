<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="姓名">
        <el-input v-model="query.name" placeholder="请输入姓名" clearable />
      </el-form-item>
      <el-form-item label="部门">
        <el-select v-model="query.department_id" placeholder="全部" clearable style="width:160px">
          <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.approve_status" placeholder="全部" clearable style="width:140px">
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
        <el-button type="primary" @click="openCreateDialog">新增员工</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column label="序号" width="90">
        <template #default="{ $index }">
          {{ seqNo($index) }}
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
          <el-tag :type="approveStatusTagType(row.approve_status)">
            {{ approveStatusLabel(row.approve_status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button
            size="small"
            type="primary"
            link
            :disabled="!canEdit(row)"
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button size="small" type="primary" link @click="openDetail(row.id)">查看详情</el-button>
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

    <el-dialog v-model="detailVisible" title="员工详情" width="760px">
      <template v-if="detailData">
        <el-divider content-position="left">基本信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ detailData.id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="姓名">{{ detailData.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="电话">{{ detailData.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ detailData.email || '-' }}</el-descriptions-item>
          <el-descriptions-item label="部门">{{ detailData.department?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="职位">{{ detailData.position_info?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="入职时间">{{ formatDate(detailData.onboard_date) }}</el-descriptions-item>
          <el-descriptions-item label="入职类型">{{ onboardTypeLabel(detailData.onboard_type) }}</el-descriptions-item>
          <el-descriptions-item label="试用期（月）">{{ detailData.probation_days ?? '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">个人信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="试用截止日期">{{ formatDate(detailData.probation_end) }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ detailData.id_card || '-' }}</el-descriptions-item>
          <el-descriptions-item label="籍贯">{{ detailData.native_place || '-' }}</el-descriptions-item>
          <el-descriptions-item label="现居地址">{{ detailData.address || '-' }}</el-descriptions-item>
          <el-descriptions-item label="紧急联系人">{{ detailData.emergency_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="紧急联系电话">{{ detailData.emergency_phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="备注" :span="2">{{ detailData.remark || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">教育与经历</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="学历">{{ educationLabel(detailData.education) }}</el-descriptions-item>
          <el-descriptions-item label="工作年限">{{ detailData.work_years ?? 0 }} 年</el-descriptions-item>
          <el-descriptions-item label="毕业院校">{{ detailData.school || '-' }}</el-descriptions-item>
          <el-descriptions-item label="专业">{{ detailData.major || '-' }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="createVisible" :title="formMode === 'edit' ? '编辑员工' : '新增员工'" width="760px" @closed="resetCreateForm">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <el-divider content-position="left">基本信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="姓名" prop="name">
              <el-input v-model="createForm.name" placeholder="请输入姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="电话" prop="phone">
              <el-input v-model="createForm.phone" placeholder="请输入手机号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="createForm.email" placeholder="请输入邮箱（选填）" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="部门" prop="department_id">
              <el-select v-model="createForm.department_id" placeholder="请选择部门" clearable style="width:100%">
                <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="职位" prop="position_id">
              <el-select v-model="createForm.position_id" placeholder="请选择职位" clearable style="width:100%">
                <el-option v-for="p in positions" :key="p.id" :label="p.name" :value="p.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="入职时间">
              <el-date-picker
                v-model="createForm.onboard_date"
                type="date"
                placeholder="请选择入职时间"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                style="width:100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="入职类型">
              <el-select v-model="createForm.onboard_type" style="width:100%">
                <el-option label="新员工" value="new" />
                <el-option label="返聘" value="rehire" />
                <el-option label="调入" value="transfer" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="试用期（月）">
              <el-input-number v-model="createForm.probation_days" :min="0" :max="365" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="试用截止日期">
              <el-input
                :model-value="calculatedProbationEnd"
                disabled
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">个人信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="身份证号">
              <el-input v-model="createForm.id_card" placeholder="请输入身份证号" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="籍贯">
              <el-input v-model="createForm.native_place" placeholder="请输入籍贯" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="现居地址">
              <el-input v-model="createForm.address" placeholder="请输入现居地址" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="紧急联系人">
              <el-input v-model="createForm.emergency_name" placeholder="请输入紧急联系人" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="紧急联系电话">
              <el-input v-model="createForm.emergency_phone" placeholder="请输入紧急联系电话" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">教育与经历</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="学历">
              <el-select v-model="createForm.education" placeholder="请选择学历" clearable style="width:100%">
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
              <el-input-number v-model="createForm.work_years" :min="0" :max="50" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="毕业院校">
              <el-input v-model="createForm.school" placeholder="请输入毕业院校" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="专业">
              <el-input v-model="createForm.major" placeholder="请输入专业" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="createForm.remark" type="textarea" :rows="2" placeholder="请输入备注（选填）" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleSubmit">
          {{ formMode === 'edit' ? '保存' : '确定' }}
        </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getEmployees, getEmployee, createEmployee, updateEmployee } from '../api/employee'
import { getDepartments } from '../api/department'
import { getPositions } from '../api/position'

const list = ref([])
const total = ref(0)
const departments = ref([])
const positions = ref([])

const detailVisible = ref(false)
const detailData = ref(null)

const createVisible = ref(false)
const createFormRef = ref()
const createLoading = ref(false)
const formMode = ref('create')
const editingId = ref(null)

const query = ref({ name: '', department_id: null, approve_status: '', page: 1, page_size: 10 })

const defaultCreateForm = () => ({
  name: '',
  phone: '',
  email: '',
  onboard_date: '',
  onboard_type: 'new',
  probation_days: 3,
  probation_end: '',
  id_card: '',
  native_place: '',
  address: '',
  emergency_name: '',
  emergency_phone: '',
  education: '',
  school: '',
  major: '',
  work_years: 0,
  remark: '',
  department_id: null,
  position_id: null,
  status: 1
})
const createForm = ref(defaultCreateForm())

const createRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
  department_id: [{ required: true, message: '请选择部门', trigger: 'change' }],
  position_id: [{ required: true, message: '请选择职位', trigger: 'change' }]
}

const seqNo = (idx) => (query.value.page - 1) * query.value.page_size + idx + 1

const approveStatusLabel = (status) => {
  const key = String(status || '').toLowerCase()
  if (key === 'draft') return '草稿'
  if (key === 'pending') return '待审批'
  if (key === 'approved') return '已通过'
  if (key === 'rejected') return '已拒绝'
  return '-'
}

const approveStatusTagType = (status) => {
  const key = String(status || '').toLowerCase()
  if (key === 'approved') return 'success'
  if (key === 'pending') return 'warning'
  if (key === 'rejected') return 'danger'
  return 'info'
}

const canEdit = (row) => {
  const status = String(row?.approve_status || '').toLowerCase()
  return status === 'draft'
}

const parseDateOnly = (value) => {
  const text = String(value || '')
  const match = text.match(/^(\d{4})-(\d{2})-(\d{2})$/)
  if (!match) return null
  return new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]))
}

const formatDateOnly = (date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const calcProbationEnd = (onboardDate, probationMonths) => {
  const base = parseDateOnly(onboardDate)
  const months = Number(probationMonths)
  if (!base || !Number.isFinite(months) || months <= 0) return ''
  return formatDateOnly(new Date(base.getFullYear(), base.getMonth() + months, base.getDate()))
}

const calculatedProbationEnd = computed(() => calcProbationEnd(createForm.value.onboard_date, createForm.value.probation_days))

const toDateValue = (value) => {
  if (!value) return ''
  return String(value).slice(0, 10)
}

const mapEmployeeToForm = (data) => ({
  ...defaultCreateForm(),
  name: data?.name || '',
  phone: data?.phone || '',
  email: data?.email || '',
  onboard_date: toDateValue(data?.onboard_date),
  onboard_type: data?.onboard_type || 'new',
  probation_days: data?.probation_days ?? 3,
  probation_end: toDateValue(data?.probation_end),
  id_card: data?.id_card || '',
  native_place: data?.native_place || '',
  address: data?.address || '',
  emergency_name: data?.emergency_name || '',
  emergency_phone: data?.emergency_phone || '',
  education: data?.education || '',
  school: data?.school || '',
  major: data?.major || '',
  work_years: data?.work_years ?? 0,
  remark: data?.remark || '',
  department_id: data?.department_id ?? data?.department?.id ?? null,
  position_id: data?.position_id ?? data?.position_info?.id ?? null,
  status: Number.isFinite(Number(data?.status)) ? Number(data.status) : 1
})

const handleEdit = async (row) => {
  if (!canEdit(row)) return
  try {
    const res = await getEmployee(row.id)
    createForm.value = mapEmployeeToForm(res.data?.data || {})
    formMode.value = 'edit'
    editingId.value = row.id
    createVisible.value = true
  } catch {
    ElMessage.error('获取员工详情失败')
  }
}

const formatDate = (value) => {
  if (!value) return '-'
  return String(value).slice(0, 10)
}

const onboardTypeLabel = (type) => ({
  new: '新员工',
  rehire: '返聘',
  transfer: '调入'
}[type] || type || '-')

const educationLabel = (education) => ({
  junior: '初中及以下',
  high: '高中/中专',
  college: '大专',
  bachelor: '本科',
  master: '硕士',
  doctor: '博士'
}[education] || (education || '-'))

const loadData = async () => {
  const params = { ...query.value }
  if (!params.department_id) delete params.department_id
  if (!params.approve_status) delete params.approve_status
  const res = await getEmployees(params)
  list.value = res.data?.data?.list || []
  total.value = res.data?.data?.total || 0
}

const handleSearch = () => {
  query.value.page = 1
  loadData()
}

const handleReset = () => {
  query.value = { name: '', department_id: null, approve_status: '', page: 1, page_size: 10 }
  loadData()
}

const openDetail = async (id) => {
  const res = await getEmployee(id)
  detailData.value = res.data?.data || null
  detailVisible.value = true
}

const openCreateDialog = () => {
  createForm.value = defaultCreateForm()
  formMode.value = 'create'
  editingId.value = null
  createVisible.value = true
}

const resetCreateForm = () => {
  createFormRef.value?.clearValidate()
  createForm.value = defaultCreateForm()
  formMode.value = 'create'
  editingId.value = null
}

const handleSubmit = async () => {
  try {
    await createFormRef.value.validate()
  } catch {
    return
  }

  createLoading.value = true
  try {
    const payload = {
      ...createForm.value,
      probation_end: calculatedProbationEnd.value,
      department_id: createForm.value.department_id,
      position_id: createForm.value.position_id
    }
    if (formMode.value === 'edit' && editingId.value) {
      await updateEmployee(editingId.value, payload)
      ElMessage.success('保存成功')
    } else {
      await createEmployee(payload)
      ElMessage.success('新增成功')
    }
    createVisible.value = false
    query.value.page = 1
    await loadData()
  } catch (error) {
    if (!error?.response?.data?.msg) {
      const fallbackMsg = formMode.value === 'edit' ? '保存失败' : '新增失败'
      ElMessage.error(error?.message || fallbackMsg)
    }
  } finally {
    createLoading.value = false
  }
}

onMounted(async () => {
  const [deptRes, posRes] = await Promise.all([
    getDepartments({ page: 1, page_size: 1000 }),
    getPositions({ page: 1, page_size: 1000 })
  ])
  departments.value = deptRes.data?.data?.list || []
  positions.value = posRes.data?.data?.list || []
  loadData()
})
</script>
