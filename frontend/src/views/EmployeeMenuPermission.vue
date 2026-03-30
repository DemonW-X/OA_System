<template>
  <el-card shadow="never" style="border:none">
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
        <template #default="{ $index }">{{ seqNo($index) }}</template>
      </el-table-column>
      <el-table-column prop="name" label="姓名" min-width="120" />
      <el-table-column prop="phone" label="电话" min-width="140" />
      <el-table-column label="部门" min-width="120">
        <template #default="{ row }">{{ row.department?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="职位" min-width="120">
        <template #default="{ row }">{{ row.position_info?.name || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '在职' : '离职' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="openPermissionDialog(row)">编辑</el-button>
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

    <el-dialog v-model="dialogVisible" :title="`菜单权限分配 - ${activeEmployee.name || ''}`" width="620px" destroy-on-close>
      <el-skeleton :loading="dialogLoading" animated :rows="6">
        <template #default>
          <el-alert
            title="勾选子菜单时会自动包含其父菜单"
            type="info"
            :closable="false"
            style="margin-bottom:10px"
          />
          <div style="display:flex;justify-content:flex-end;gap:8px;margin-bottom:10px">
            <el-button size="small" @click="handleCheckAll">全选</el-button>
            <el-button size="small" @click="handleClearAll">清空</el-button>
          </div>
          <el-tree
            ref="treeRef"
            :data="menuTree"
            node-key="id"
            show-checkbox
            default-expand-all
            :props="{ label: 'name', children: 'children' }"
            @check="handleTreeCheck"
          >
            <template #default="{ data }">
              <div style="display:flex;align-items:center;gap:8px">
                <span>{{ data.name }}</span>
                <el-tag size="small" type="info">{{ data.path || '-' }}</el-tag>
              </div>
            </template>
          </el-tree>
        </template>
      </el-skeleton>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { getEmployees, getEmployeeMenuPermissions, setEmployeeMenuPermissions } from '../api/employee'
import { getDepartments } from '../api/department'

const list = ref([])
const total = ref(0)
const departments = ref([])
const query = ref({ name: '', department_id: null, status: 1, page: 1, page_size: 10 })

const dialogVisible = ref(false)
const dialogLoading = ref(false)
const saving = ref(false)
const activeEmployee = ref({ id: null, name: '' })
const menuTree = ref([])
const treeRef = ref()
const parentMap = ref({})
const allMenuIds = ref([])
const loadedCheckedKeys = ref([])

const seqNo = (idx) => (query.value.page - 1) * query.value.page_size + idx + 1

const loadData = async () => {
  const res = await getEmployees(query.value)
  list.value = res.data?.data?.list || []
  total.value = res.data?.data?.total || 0
}

const handleSearch = () => {
  query.value.page = 1
  loadData()
}

const handleReset = () => {
  query.value = { name: '', department_id: null, status: 1, page: 1, page_size: 10 }
  loadData()
}

const buildParentMap = (nodes = [], parentId = 0) => {
  for (const n of nodes) {
    parentMap.value[n.id] = parentId
    if (n.children?.length) buildParentMap(n.children, n.id)
  }
}

const collectAllMenuIds = (nodes = []) => {
  const ids = []
  const walk = (arr = []) => {
    for (const n of arr) {
      ids.push(n.id)
      if (n.children?.length) walk(n.children)
    }
  }
  walk(nodes)
  return ids
}

const openPermissionDialog = async (row) => {
  activeEmployee.value = { id: row.id, name: row.name }
  dialogVisible.value = true
  dialogLoading.value = true
  menuTree.value = []
  parentMap.value = {}
  allMenuIds.value = []
  loadedCheckedKeys.value = []

  try {
    const res = await getEmployeeMenuPermissions(row.id)
    const data = res.data?.data || {}
    menuTree.value = data.menu_tree || []
    allMenuIds.value = collectAllMenuIds(menuTree.value)
    buildParentMap(menuTree.value)
    loadedCheckedKeys.value = data.checked_menu_ids || []
  } finally {
    dialogLoading.value = false
  }

  // 注意：需要等 skeleton 渲染出 el-tree 后再回显勾选
  await nextTick()
  const checked = includeParentKeys(loadedCheckedKeys.value)
  treeRef.value?.setCheckedKeys(checked)
}

const includeParentKeys = (keys = []) => {
  const set = new Set(keys)
  for (const id of keys) {
    let pid = parentMap.value[id]
    while (pid && pid > 0) {
      set.add(pid)
      pid = parentMap.value[pid]
    }
  }
  return Array.from(set)
}

const handleTreeCheck = () => {
  const keys = treeRef.value?.getCheckedKeys(false) || []
  const merged = includeParentKeys(keys)
  treeRef.value?.setCheckedKeys(merged)
}

const handleCheckAll = () => {
  treeRef.value?.setCheckedKeys(allMenuIds.value)
}

const handleClearAll = () => {
  treeRef.value?.setCheckedKeys([])
}

const handleSave = async () => {
  if (!activeEmployee.value.id) return
  const checked = treeRef.value?.getCheckedKeys(false) || []
  const menuIds = includeParentKeys(checked)
  saving.value = true
  try {
    await setEmployeeMenuPermissions(activeEmployee.value.id, { menu_ids: menuIds })
    ElMessage.success('保存成功')
    dialogVisible.value = false
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const deptRes = await getDepartments({ page: 1, page_size: 100 })
  departments.value = deptRes.data?.data?.list || []
  loadData()
})
</script>
