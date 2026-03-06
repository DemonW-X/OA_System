<template>
  <el-tabs v-model="activeTab" type="border-card">
    <!-- 部门管理 -->
    <el-tab-pane label="部门管理" name="department">
      <el-card shadow="never" style="border:none">
        <template #header>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span>部门列表</span>
            <el-button type="primary" @click="openDeptDialog()">新增部门</el-button>
          </div>
        </template>

        <el-form :inline="true" style="margin-bottom:16px">
          <el-form-item label="部门名称">
            <el-input v-model="deptQuery.name" placeholder="请输入部门名称" clearable />
          </el-form-item>
          <el-form-item label="上级部门">
            <el-select v-model="deptQuery.parent_id" placeholder="全部" clearable style="width:180px">
              <el-option label="一级部门" :value="0" />
              <el-option v-for="d in allDepts" :key="d.id" :label="d.name" :value="d.id" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleDeptSearch">搜索</el-button>
            <el-button @click="handleDeptReset">重置</el-button>
          </el-form-item>
        </el-form>

        <el-table :data="deptList" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="部门名称" />
          <el-table-column prop="level" label="层级" width="90" />
          <el-table-column label="上级部门">
            <template #default="{ row }">{{ row.parent?.name || '-' }}</template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" />
          <el-table-column label="操作" width="160">
            <template #default="{ row }">
              <el-button size="small" @click="openDeptDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeptDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          style="margin-top:16px;justify-content:flex-end;display:flex"
          v-model:current-page="deptQuery.page"
          v-model:page-size="deptQuery.page_size"
          :total="deptTotal"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadDepts"
        />
      </el-card>
    </el-tab-pane>

    <!-- 职位管理（不绑定部门） -->
    <el-tab-pane label="职位管理" name="position">
      <el-card shadow="never" style="border:none">
        <template #header>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span>职位列表</span>
            <el-button type="primary" @click="openPosDialog()">新增职位</el-button>
          </div>
        </template>

        <el-form :inline="true" style="margin-bottom:16px">
          <el-form-item label="职位名称">
            <el-input v-model="posQuery.name" placeholder="请输入职位名称" clearable />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handlePosSearch">搜索</el-button>
            <el-button @click="handlePosReset">重置</el-button>
          </el-form-item>
        </el-form>

        <el-table :data="posList" stripe>
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="职位名称" />
          <el-table-column prop="sort_order" label="排序" width="90" />
          <el-table-column prop="remark" label="备注" />
          <el-table-column label="操作" width="160">
            <template #default="{ row }">
              <el-button size="small" @click="openPosDialog(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handlePosDelete(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-pagination
          style="margin-top:16px;justify-content:flex-end;display:flex"
          v-model:current-page="posQuery.page"
          v-model:page-size="posQuery.page_size"
          :total="posTotal"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadPositions"
        />
      </el-card>
    </el-tab-pane>
  </el-tabs>

  <el-dialog v-model="deptDialogVisible" :title="deptForm.id ? '编辑部门' : '新增部门'" width="460px">
    <el-form :model="deptForm" label-width="90px">
      <el-form-item label="部门名称">
        <el-input v-model="deptForm.name" />
      </el-form-item>
      <el-form-item label="上级部门">
        <el-select v-model="deptForm.parent_id" placeholder="不选则为一级部门" clearable style="width:100%">
          <el-option v-for="d in parentDeptOptions" :key="d.id" :label="`${d.name}（L${d.level || 1}）`" :value="d.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="deptForm.remark" type="textarea" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="deptDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handleDeptSubmit">确定</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="posDialogVisible" :title="posForm.id ? '编辑职位' : '新增职位'" width="400px">
    <el-form :model="posForm" label-width="80px">
      <el-form-item label="职位名称">
        <el-input v-model="posForm.name" />
      </el-form-item>
      <el-form-item label="排序">
        <el-input-number v-model="posForm.sort_order" :min="0" style="width:100%" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="posForm.remark" type="textarea" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="posDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handlePosSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDepartments, createDepartment, updateDepartment, deleteDepartment } from '../api/department'
import { getPositions, createPosition, updatePosition, deletePosition } from '../api/position'

const activeTab = ref('department')

const deptList = ref([])
const deptTotal = ref(0)
const deptDialogVisible = ref(false)
const deptForm = ref({ name: '', parent_id: null, remark: '' })
const deptQuery = ref({ name: '', parent_id: null, page: 1, page_size: 10 })
const allDepts = ref([])

const parentDeptOptions = computed(() => allDepts.value.filter(d => (deptForm.value.id ? d.id !== deptForm.value.id : true)))

const loadDepts = async () => {
  const params = { ...deptQuery.value }
  if (params.parent_id === null || params.parent_id === undefined || params.parent_id === '') delete params.parent_id
  const res = await getDepartments(params)
  deptList.value = res.data.data.list || []
  deptTotal.value = res.data.data.total || 0
}

const loadAllDepts = async () => {
  const res = await getDepartments({ page: 1, page_size: 1000 })
  allDepts.value = res.data.data.list || []
}

const handleDeptSearch = () => { deptQuery.value.page = 1; loadDepts() }
const handleDeptReset = () => { deptQuery.value = { name: '', parent_id: null, page: 1, page_size: 10 }; loadDepts() }

const openDeptDialog = (row = null) => {
  deptForm.value = row ? { id: row.id, name: row.name, parent_id: row.parent_id ?? null, remark: row.remark } : { name: '', parent_id: null, remark: '' }
  deptDialogVisible.value = true
}

const handleDeptSubmit = async () => {
  if (deptForm.value.id) await updateDepartment(deptForm.value.id, deptForm.value)
  else await createDepartment(deptForm.value)
  ElMessage.success('操作成功')
  deptDialogVisible.value = false
  await Promise.all([loadDepts(), loadAllDepts()])
}

const handleDeptDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该部门？', '提示', { type: 'warning' })
  await deleteDepartment(id)
  ElMessage.success('删除成功')
  await Promise.all([loadDepts(), loadAllDepts()])
}

const posList = ref([])
const posTotal = ref(0)
const posDialogVisible = ref(false)
const posForm = ref({ name: '', sort_order: 0, remark: '' })
const posQuery = ref({ name: '', page: 1, page_size: 10 })

const loadPositions = async () => {
  const res = await getPositions(posQuery.value)
  posList.value = res.data.data.list || []
  posTotal.value = res.data.data.total || 0
}

const handlePosSearch = () => { posQuery.value.page = 1; loadPositions() }
const handlePosReset = () => { posQuery.value = { name: '', page: 1, page_size: 10 }; loadPositions() }

const openPosDialog = (row = null) => {
  posForm.value = row ? { id: row.id, name: row.name, sort_order: row.sort_order || 0, remark: row.remark } : { name: '', sort_order: 0, remark: '' }
  posDialogVisible.value = true
}

const handlePosSubmit = async () => {
  if (posForm.value.id) await updatePosition(posForm.value.id, posForm.value)
  else await createPosition(posForm.value)
  ElMessage.success('操作成功')
  posDialogVisible.value = false
  loadPositions()
}

const handlePosDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该职位？', '提示', { type: 'warning' })
  await deletePosition(id)
  ElMessage.success('删除成功')
  loadPositions()
}

onMounted(() => {
  loadDepts()
  loadAllDepts()
  loadPositions()
})
</script>
