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
        <el-select v-model="query.status" placeholder="全部" clearable style="width:100px">
          <el-option label="在职" :value="1" />
          <el-option label="离职" :value="0" />
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
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
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

    <el-dialog v-model="detailVisible" title="员工详情" width="520px">
      <el-descriptions :column="1" border v-if="detailData">
        <el-descriptions-item label="ID">{{ detailData.id }}</el-descriptions-item>
        <el-descriptions-item label="姓名">{{ detailData.name }}</el-descriptions-item>
        <el-descriptions-item label="电话">{{ detailData.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ detailData.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="部门">{{ detailData.department?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="职位">{{ detailData.position_info?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ detailData.status === 1 ? '在职' : '离职' }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="createVisible" title="新增员工" width="520px" @closed="resetCreateForm">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="90px">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="createForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="请输入邮箱（选填）" />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="createForm.department_id" placeholder="请选择部门" clearable style="width:100%">
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="职位">
          <el-select v-model="createForm.position_id" placeholder="请选择职位" clearable style="width:100%">
            <el-option v-for="p in positions" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="createForm.status">
            <el-radio :value="1">在职</el-radio>
            <el-radio :value="0">离职</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">确定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getEmployees, getEmployee, createEmployee } from '../api/employee'
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

const query = ref({ name: '', department_id: null, status: null, page: 1, page_size: 10 })

const defaultCreateForm = () => ({
  name: '',
  phone: '',
  email: '',
  department_id: null,
  position_id: null,
  status: 1
})
const createForm = ref(defaultCreateForm())

const createRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }]
}

const seqNo = (idx) => (query.value.page - 1) * query.value.page_size + idx + 1

const loadData = async () => {
  const params = { ...query.value }
  if (!params.department_id) delete params.department_id
  if (params.status === null || params.status === undefined || params.status === '') delete params.status
  const res = await getEmployees(params)
  list.value = res.data?.data?.list || []
  total.value = res.data?.data?.total || 0
}

const handleSearch = () => {
  query.value.page = 1
  loadData()
}

const handleReset = () => {
  query.value = { name: '', department_id: null, status: null, page: 1, page_size: 10 }
  loadData()
}

const openDetail = async (id) => {
  const res = await getEmployee(id)
  detailData.value = res.data?.data || null
  detailVisible.value = true
}

const openCreateDialog = () => {
  createForm.value = defaultCreateForm()
  createVisible.value = true
}

const resetCreateForm = () => {
  createFormRef.value?.clearValidate()
  createForm.value = defaultCreateForm()
}

const handleCreate = async () => {
  await createFormRef.value.validate()
  createLoading.value = true
  try {
    const payload = {
      ...createForm.value,
      department_id: createForm.value.department_id || 0,
      position_id: createForm.value.position_id || 0
    }
    await createEmployee(payload)
    ElMessage.success('新增成功')
    createVisible.value = false
    query.value.page = 1
    await loadData()
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
