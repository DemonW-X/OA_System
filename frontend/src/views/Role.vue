<template>
  <el-card shadow="never">
    <div style="display:flex;gap:16px;height:calc(100vh - 240px)">
    <!-- 左侧：部门树 -->
    <div style="width:280px;border-right:1px solid #ebeef5;overflow:auto;padding:8px">
      <div style="font-weight:600;padding:8px 12px;border-bottom:1px solid #ebeef5;margin-bottom:8px">部门结构</div>
      <el-tree
        ref="deptTreeRef"
        :data="deptTree"
        node-key="id"
        :props="{ children: 'children', label: 'name' }"
        :expand-on-click-node="false"
        default-expand-all
        highlight-current
        @node-click="handleDeptClick"
      >
        <template #default="{ node, data }">
          <div style="flex:1;display:flex;align-items:center;justify-content:space-between;padding-right:4px">
            <div style="display:flex;align-items:center;gap:6px">
              <span>{{ data.name }}</span>
            </div>
            <span class="dept-tree-actions" @click.stop>
              <el-button :icon="Plus" type="primary" size="small" circle plain @click.stop="openDeptDialog(null, data.id)" />
              <el-button :icon="Edit" type="warning" size="small" circle plain @click.stop="openDeptDialog(data)" />
              <el-button :icon="Minus" type="danger" size="small" circle plain @click.stop="handleDeptDelete(data.id)" />
            </span>
          </div>
        </template>
      </el-tree>
    </div>

    <!-- 右侧：职位列表 -->
    <div style="flex:1;overflow:auto;padding:12px">
      <div v-if="!selectedDept" style="text-align:center;color:#909399;padding:60px 0">
        请在左侧选择部门查看职位
      </div>
      <div v-else>
        <div style="font-weight:600;margin-bottom:12px;display:flex;align-items:center;gap:8px">
          <span>{{ selectedDept.name }}</span>
          <el-tag size="small">{{ positionList.length }} 个职位</el-tag>
          <el-button type="primary" size="small" style="margin-left:auto" @click="openAddDialog">添加关系</el-button>
        </div>
        <el-table :data="positionList" stripe size="small" v-loading="loadingPositions">
          <el-table-column prop="name" label="职位名称" />
          <el-table-column prop="remark" label="备注" />
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="handleRemoveRelation(row)">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
    </div>

    <!-- 添加关系弹窗 -->
    <el-dialog v-model="addDialogVisible" title="添加部门" width="700px">
      <div style="display:flex;gap:16px;height:400px">
        <div style="width:280px;border:1px solid #ebeef5;border-radius:4px;overflow:auto;padding:8px">
          <div style="font-weight:600;padding:4px 8px;margin-bottom:8px">选择部门</div>
          <el-tree
            ref="addDeptTreeRef"
            :data="deptTree"
            node-key="id"
            :props="{ children: 'children', label: 'name' }"
            :expand-on-click-node="false"
            default-expand-all
            highlight-current
            @node-click="handleAddDeptClick"
          >
            <template #default="{ data }">
              <span>{{ data.name }}</span>
            </template>
          </el-tree>
        </div>
        <div style="flex:1;border:1px solid #ebeef5;border-radius:4px;overflow:auto;padding:12px">
          <div v-if="!addSelectedDept" style="text-align:center;color:#909399;padding:60px 0">
            请先在左侧选择部门
          </div>
          <div v-else>
            <div style="font-weight:600;margin-bottom:12px">选择职位（可多选）</div>
            <el-checkbox-group v-model="addSelectedPositions">
              <div v-for="p in allPositions" :key="p.id" style="margin-bottom:8px">
                <el-checkbox :label="p.id">{{ p.name }}</el-checkbox>
              </div>
            </el-checkbox-group>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchAdd" :disabled="!addSelectedDept || !addSelectedPositions.length">确定添加</el-button>
      </template>
    </el-dialog>

    <!-- 部门新增/编辑弹窗 -->
    <el-dialog v-model="deptDialogVisible" :title="deptForm.id ? '编辑部门' : '新增部门'" width="400px">
      <el-form :model="deptForm" label-width="80px">
        <el-form-item label="部门名称">
          <el-input v-model="deptForm.name" placeholder="请输入部门名称" />
        </el-form-item>
        <el-form-item label="上级部门">
          <el-select v-model="deptForm.parent_id" placeholder="一级部门" clearable style="width:100%">
            <el-option v-for="d in flatDepts" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="deptForm.remark" type="textarea" rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deptDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleDeptSubmit">确定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Minus } from '@element-plus/icons-vue'
import { getDepartments, createDepartment, updateDepartment, deleteDepartment } from '../api/department'
import { getPositions } from '../api/position'
import { createDepartmentPosition, deleteDepartmentPosition, getDepartmentPositions } from '../api/department_position'

const deptTreeRef = ref()
const addDeptTreeRef = ref()
const deptTree = ref([])
const selectedDept = ref(null)
const positionList = ref([])
const loadingPositions = ref(false)
const allPositions = ref([])

const addDialogVisible = ref(false)
const addSelectedDept = ref(null)
const addSelectedPositions = ref([])

const deptDialogVisible = ref(false)
const deptForm = ref({ id: null, name: '', parent_id: null, remark: '' })

const flatDepts = computed(() => {
  const result = []
  const flatten = (nodes) => nodes.forEach(n => { result.push(n); if (n.children) flatten(n.children) })
  flatten(deptTree.value)
  return result
})

const openDeptDialog = (dept = null, parentId = null) => {
  if (dept) {
    deptForm.value = { id: dept.id, name: dept.name, parent_id: dept.parent_id ?? null, remark: dept.remark || '' }
  } else {
    deptForm.value = { id: null, name: '', parent_id: parentId ?? null, remark: '' }
  }
  deptDialogVisible.value = true
}

const handleDeptSubmit = async () => {
  if (!deptForm.value.name) return ElMessage.warning('请输入部门名称')
  const payload = { name: deptForm.value.name, parent_id: deptForm.value.parent_id, remark: deptForm.value.remark }
  if (deptForm.value.id) {
    await updateDepartment(deptForm.value.id, payload)
    ElMessage.success('修改成功')
  } else {
    await createDepartment(payload)
    ElMessage.success('新增成功')
  }
  deptDialogVisible.value = false
  await loadDeptTree()
}

const handleDeptDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该部门？', '提示', { type: 'warning' })
  await deleteDepartment(id)
  ElMessage.success('删除成功')
  if (selectedDept.value?.id === id) selectedDept.value = null
  await loadDeptTree()
}

const buildDeptTree = (depts) => {
  const map = {}
  const roots = []
  for (const d of depts) {
    map[d.id] = { ...d, children: [] }
  }
  for (const d of depts) {
    const node = map[d.id]
    if (d.parent_id && map[d.parent_id]) {
      map[d.parent_id].children.push(node)
    } else {
      roots.push(node)
    }
  }
  return roots
}

const loadDeptTree = async () => {
  const res = await getDepartments({ page: 1, page_size: 1000 })
  const list = res.data?.data?.list || []
  deptTree.value = buildDeptTree(list)
}

const loadAllPositions = async () => {
  const res = await getPositions({ page: 1, page_size: 1000 })
  allPositions.value = res.data?.data?.list || []
}

const handleDeptClick = async (dept) => {
  selectedDept.value = dept
  loadingPositions.value = true
  try {
    const res = await getDepartmentPositions({ department_id: dept.id, page: 1, page_size: 1000 })
    const rels = res.data?.data?.list || []
    positionList.value = rels.map(r => ({
      ...r.position,
      relation_id: r.id
    })).sort((a, b) => a.id - b.id)
  } finally {
    loadingPositions.value = false
  }
}

const handleRemoveRelation = async (row) => {
  if (!row.relation_id) return
  await ElMessageBox.confirm(`确认移除职位「${row.name}」？`, '提示', { type: 'warning' })
  await deleteDepartmentPosition(row.relation_id)
  ElMessage.success('移除成功')
  handleDeptClick(selectedDept.value)
}

const openAddDialog = () => {
  addSelectedDept.value = null
  addSelectedPositions.value = []
  addDialogVisible.value = true
}

const handleAddDeptClick = (dept) => {
  addSelectedDept.value = dept
  addSelectedPositions.value = []
}

const handleBatchAdd = async () => {
  if (!addSelectedDept.value || !addSelectedPositions.value.length) return
  const deptId = addSelectedDept.value.id
  const posIds = addSelectedPositions.value

  let successCount = 0
  for (const posId of posIds) {
    try {
      await createDepartmentPosition({ department_id: deptId, position_id: posId })
      successCount++
    } catch (e) {
      console.warn(`职位 ${posId} 添加失败:`, e)
    }
  }

  ElMessage.success(`成功添加 ${successCount} 个关系`)
  addDialogVisible.value = false
  if (selectedDept.value?.id === deptId) {
    handleDeptClick(selectedDept.value)
  }
}

onMounted(async () => {
  await loadDeptTree()
  await loadAllPositions()
})
</script>

<style scoped>
.dept-tree-actions {
  visibility: hidden;
  display: inline-flex;
  gap: 4px;
}

:deep(.el-tree-node__content:hover) .dept-tree-actions {
  visibility: visible;
}

.dept-tree-actions .el-button {
  background: transparent;
  border-color: transparent;
}

.dept-tree-actions .el-button:hover {
  background: var(--el-button-bg-color);
  border-color: var(--el-button-border-color);
}
</style>
