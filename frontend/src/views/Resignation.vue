<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>离职管理</span>
        <el-button type="primary" @click="openDialog()">新增离职</el-button>
      </div>
    </template>

    <el-form :inline="true" style="margin-bottom:16px">
      <el-form-item label="员工">
        <el-select v-model="query.employee_id" placeholder="全部员工" clearable filterable style="width:180px">
          <el-option v-for="e in employeeList" :key="e.id" :label="e.name" :value="e.id" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
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
      <el-table-column label="操作" width="170">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
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
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEmployees } from '../api/employee'
import { getResignations, createResignation, updateResignation, deleteResignation } from '../api/resignation'

const list = ref([])
const total = ref(0)
const employeeList = ref([])
const dialogVisible = ref(false)
const formRef = ref()
const query = ref({ employee_id: null, page: 1, page_size: 10 })

const defaultForm = () => ({ employee_id: null, resign_date: '', reason: '', remark: '' })
const form = ref(defaultForm())

const rules = {
  employee_id: [{ required: true, message: '请选择员工', trigger: 'change' }],
  resign_date: [{ required: true, message: '请选择离职日期', trigger: 'change' }],
}

const formatDate = (v) => (v ? String(v).slice(0, 10) : '-')

const activeEmployeeOptions = computed(() => {
  if (form.value.id) {
    // 编辑时允许保留原员工（即便已是离职）
    return employeeList.value
  }
  return employeeList.value.filter(e => e.status === 1)
})

const loadEmployees = async () => {
  const res = await getEmployees({ page: 1, page_size: 1000 })
  employeeList.value = res.data?.data?.list || []
}

const loadData = async () => {
  const params = { ...query.value }
  if (!params.employee_id) delete params.employee_id
  const res = await getResignations(params)
  list.value = res.data?.data?.list || []
  total.value = res.data?.data?.total || 0
}

const handleSearch = () => {
  query.value.page = 1
  loadData()
}

const handleReset = () => {
  query.value = { employee_id: null, page: 1, page_size: 10 }
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

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该离职记录？删除后可能恢复员工为在职。', '提示', { type: 'warning' })
  await deleteResignation(id)
  ElMessage.success('删除成功')
  await Promise.all([loadData(), loadEmployees()])
}

onMounted(async () => {
  await Promise.all([loadEmployees(), loadData()])
})
</script>
