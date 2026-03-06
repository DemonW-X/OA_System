<template>
  <el-card shadow="never">
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>角色管理</span>
        <el-button type="primary" @click="openAddDialog">添加关系</el-button>
      </div>
    </template>

    <div style="display:flex;gap:16px;height:calc(100vh - 240px)">
      <!-- 左侧：部门树 -->
      <div style="width:280px;border:1px solid #ebeef5;border-radius:4px;overflow:auto;padding:8px">
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
          <template #default="{ data }">
            <div style="display:flex;align-items:center;gap:6px">
              <el-tag size="small" type="success">L{{ data.level || 1 }}</el-tag>
              <span>{{ data.name }}</span>
            </div>
          </template>
        </el-tree>
      </div>

      <!-- 右侧：职位列表 -->
      <div style="flex:1;border:1px solid #ebeef5;border-radius:4px;overflow:auto;padding:12px">
        <div v-if="!selectedDept" style="text-align:center;color:#909399;padding:60px 0">
          请在左侧选择部门查看职位
        </div>
        <div v-else>
          <div style="font-weight:600;margin-bottom:12px;display:flex;align-items:center;gap:8px">
            <span>{{ selectedDept.name }}</span>
            <el-tag size="small">{{ positionList.length }} 个职位</el-tag>
          </div>
          <el-table :data="positionList" stripe size="small" v-loading="loadingPositions">
            <el-table-column prop="name" label="职位名称" />
            <el-table-column prop="sort_order" label="排序" width="90" />
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
        <!-- 左：部门选择 -->
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

        <!-- 右：职位多选 -->
        <div style="flex:1;border:1px solid #ebeef5;border-radius:4px;overflow:auto;padding:12px">
          <div v-if="!addSelectedDept" style="text-align:center;color:#909399;padding:60px 0">
            请先在左侧选择部门
          </div>
          <div v-else>
            <div style="font-weight:600;margin-bottom:12px">选择职位（可多选）</div>
            <el-checkbox-group v-model="addSelectedPositions">
              <div v-for="p in allPositions" :key="p.id" style="margin-bottom:8px">
                <el-checkbox :label="p.id">{{ p.name }}（排序: {{ p.sort_order || 0 }}）</el-checkbox>
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
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDepartments } from '../api/department'
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
    })).sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
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
