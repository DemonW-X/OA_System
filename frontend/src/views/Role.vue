<template>
  <el-card shadow="never">
    <div class="role-page">
      <div class="dept-panel">
        <div class="dept-panel-header">
          <span>部门结构</span>
          <el-button :icon="Plus" type="primary" size="small" plain @click="openDeptDialog()">新增根部门</el-button>
        </div>

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
            <div class="dept-node">
              <span class="dept-node-name">{{ data.name }}</span>
              <span class="dept-node-actions" @click.stop>
                <el-button :icon="Plus" type="primary" size="small" circle plain @click.stop="openDeptDialog(null, data.id)" />
                <el-button :icon="Edit" type="warning" size="small" circle plain @click.stop="openDeptDialog(data)" />
                <el-button :icon="Setting" type="success" size="small" circle plain @click.stop="openDeptPermDrawer(data)" />
                <el-button :icon="Minus" type="danger" size="small" circle plain @click.stop="handleDeptDelete(data.id)" />
              </span>
            </div>
          </template>
        </el-tree>
      </div>

      <div class="role-panel">
        <div v-if="!selectedDept" class="empty-placeholder">
          请在左侧选择部门后管理角色
        </div>
        <template v-else>
          <div class="role-toolbar">
            <div class="role-toolbar-title">
              <span>{{ selectedDept.name }}</span>
              <el-tag size="small">{{ positionList.length }} 个角色</el-tag>
            </div>

            <div class="role-toolbar-actions">
              <el-button :icon="Plus" type="primary" size="small" plain @click="openBindDialog">关联已有角色</el-button>
              <el-button :icon="Plus" type="primary" size="small" plain @click="openPositionDialog()">新建角色并关联</el-button>
            </div>
          </div>

          <el-table
            :data="positionList"
            stripe
            size="small"
            row-key="id"
            v-loading="loadingPositions"
            highlight-current-row
            :row-class-name="getRoleRowClassName"
            @row-click="handleRoleRowClick"
          >
            <el-table-column prop="name" label="角色名称" min-width="150" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <span class="role-row-actions" @click.stop>
                  <el-button :icon="Setting" type="primary" size="small" circle plain title="权限设置" @click.stop="openPermDrawer(row)" />
                  <el-button :icon="Edit" type="warning" size="small" circle plain title="编辑" @click.stop="openPositionDialog(row)" />
                  <el-button :icon="Minus" type="info" size="small" circle plain title="移除关联" @click.stop="handleRemoveRelation(row)" />
                  <el-button :icon="Delete" type="danger" size="small" circle plain title="删除角色" @click.stop="handleDeletePosition(row)" />
                </span>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </div>

      <div class="employee-panel">
        <div v-if="!selectedDept" class="empty-placeholder">
          请先在左侧选择部门
        </div>

        <div v-else-if="!selectedRole" class="empty-placeholder">
          请点击中间角色，查看并配置该角色的人员关联
        </div>
        <template v-else>
          <div class="employee-toolbar">
            <div class="employee-toolbar-title">
              <span>{{ selectedRole.name }}</span>
              <el-tag type="success" size="small">已关联 {{ roleEmployeeList.length }} 人</el-tag>
            </div>
            <div class="employee-toolbar-actions">
              <el-button :icon="Plus" type="primary" size="small" plain @click="openEmployeeBindDrawer">关联人员</el-button>
            </div>
          </div>

          <el-table
            :data="roleEmployeeList"
            stripe
            size="small"
            row-key="id"
            height="calc(100% - 70px)"
            v-loading="loadingRoleEmployees"
          >
            <el-table-column prop="name" label="员工姓名" min-width="120" />
            <el-table-column label="操作" width="92" fixed="right">
              <template #default="{ row }">
                <span class="employee-row-actions" @click.stop>
                  <el-button :icon="Setting" type="warning" size="small" circle plain title="权限设置" @click.stop="openEmployeePermEdit(row)" />
                </span>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </div>
    </div>

    <el-dialog
      v-model="bindDialogVisible"
      :title="selectedDept ? `关联已有角色到【${selectedDept.name}】` : '关联已有角色'"
      width="520px"
    >
      <div class="bind-dialog-body">
        <div v-if="!bindPositionOptions.length" class="empty-placeholder" style="padding: 40px 0;">
          暂无可关联角色，请先新建角色
        </div>
        <el-checkbox-group v-else v-model="bindSelectedPositionIDs">
          <div v-for="pos in bindPositionOptions" :key="pos.id" class="bind-item">
            <el-checkbox :label="pos.id" :disabled="pos.disabled">
              <span>{{ pos.name }}</span>
              <el-tag v-if="pos.disabled" size="small" type="info" style="margin-left: 8px;">已关联</el-tag>
            </el-checkbox>
          </div>
        </el-checkbox-group>
      </div>
      <template #footer>
        <el-button @click="bindDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!bindSelectedPositionIDs.length" @click="handleBatchBind">确认关联</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="deptDialogVisible" :title="deptForm.id ? '编辑部门' : '新增部门'" width="420px">
      <el-form :model="deptForm" label-width="90px">
        <el-form-item label="部门名称" required>
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

    <el-dialog v-model="positionDialogVisible" :title="positionForm.id ? '编辑角色' : '新建角色'" width="420px">
      <el-form :model="positionForm" label-width="90px">
        <el-form-item label="角色名称" required>
          <el-input v-model="positionForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="positionForm.remark" type="textarea" rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="positionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="positionSubmitting" @click="handlePositionSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="employeeBindDrawerVisible"
      :title="`关联人员 - ${selectedRole?.name || ''}`"
      direction="rtl"
      size="420px"
      destroy-on-close
    >
      <div v-loading="employeeBindLoading" class="employee-bind-drawer-body">
        <div v-if="!employeeBindOptions.length" class="empty-placeholder" style="padding: 40px 0;">
          暂无可关联人员
        </div>
        <el-checkbox-group v-else v-model="employeeBindSelectedIDs" class="employee-bind-list">
          <div v-for="emp in employeeBindOptions" :key="emp.id" class="employee-bind-item">
            <el-checkbox :label="emp.id">
              <span>{{ emp.name }}</span>
            </el-checkbox>
          </div>
        </el-checkbox-group>

        <div class="employee-bind-footer">
          <el-button @click="employeeBindDrawerVisible = false">取消</el-button>
          <el-button type="primary" size="small" :loading="employeeBindSaving" @click="handleSaveEmployeeBind">保存</el-button>
        </div>
      </div>
    </el-drawer>

    <el-drawer
      v-model="permDrawerVisible"
      :title="permDrawerTitle"
      direction="rtl"
      size="460px"
      destroy-on-close
    >
      <div v-loading="permLoading" class="perm-drawer-body">
        <div v-if="!permReadOnly" class="perm-header-actions">
          <el-button size="small" @click="handleCheckAllMenus">全选</el-button>
          <el-button size="small" @click="handleClearAllMenus">清空</el-button>
        </div>
        <el-tree
          ref="menuTreeRef"
          :data="permMenuTree"
          :show-checkbox="!permReadOnly"
          node-key="id"
          default-expand-all
          :props="{ label: 'name', children: 'children' }"
          class="perm-tree"
        />
        <div class="perm-footer-actions">
          <el-button @click="permDrawerVisible = false">取消</el-button>
          <el-button v-if="!permReadOnly" type="primary" :loading="permSaving" @click="handleSavePermissions">保存</el-button>
        </div>
      </div>
    </el-drawer>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Minus, Setting, Delete } from '@element-plus/icons-vue'
import {
  getDepartments,
  createDepartment,
  updateDepartment,
  deleteDepartment,
  getDepartmentMenuPermissions,
  setDepartmentMenuPermissions
} from '../api/department'
import {
  getPositions,
  createPosition,
  updatePosition,
  deletePosition,
  setPositionEmployees,
  getPositionMenuPermissions,
  setPositionMenuPermissions
} from '../api/position'
import { createDepartmentPosition, deleteDepartmentPosition, getDepartmentPositions } from '../api/department_position'
import { getEmployees } from '../api/employee'

const deptTreeRef = ref()
const deptTree = ref([])
const selectedDept = ref(null)

const positionList = ref([])
const allPositions = ref([])
const loadingPositions = ref(false)
const selectedRole = ref(null)

const roleEmployeeList = ref([])
const loadingRoleEmployees = ref(false)
const employeeBindDrawerVisible = ref(false)
const employeeBindLoading = ref(false)
const employeeBindSaving = ref(false)
const employeeBindOptions = ref([])
const employeeBindSelectedIDs = ref([])

const bindDialogVisible = ref(false)
const bindSelectedPositionIDs = ref([])

const deptDialogVisible = ref(false)
const deptForm = ref({ id: null, name: '', parent_id: null, remark: '' })

const positionDialogVisible = ref(false)
const positionSubmitting = ref(false)
const positionForm = ref({ id: null, name: '', remark: '' })

const permDrawerVisible = ref(false)
const permLoading = ref(false)
const permSaving = ref(false)
const permReadOnly = ref(false)
const permDrawerTitle = ref('')
const activePermissionTarget = ref({ type: 'position', id: null, name: '' })
const permMenuTree = ref([])
const menuTreeRef = ref()
const permParentMap = ref({})
const allMenuIDs = ref([])

const flatDepts = computed(() => {
  const result = []
  const flatten = (nodes = []) => {
    nodes.forEach((node) => {
      result.push(node)
      if (node.children?.length) flatten(node.children)
    })
  }
  flatten(deptTree.value)
  return result
})

const boundPositionIDSet = computed(() => {
  const set = new Set()
  for (const item of positionList.value) {
    if (item.id) set.add(item.id)
  }
  return set
})

const bindPositionOptions = computed(() => {
  return [...allPositions.value]
    .sort((a, b) => (a.id || 0) - (b.id || 0))
    .map((item) => ({
      ...item,
      disabled: boundPositionIDSet.value.has(item.id)
    }))
})

const extractErrorMessage = (error, fallback = '操作失败') => {
  return error?.response?.data?.msg || error?.message || fallback
}

const isDialogCancel = (error) => {
  return error === 'cancel' || error === 'close' || error?.message === 'cancel' || error?.message === 'close'
}

const buildDeptTree = (depts) => {
  const map = {}
  const roots = []
  for (const dept of depts) {
    map[dept.id] = { ...dept, children: [] }
  }
  for (const dept of depts) {
    const node = map[dept.id]
    if (dept.parent_id && map[dept.parent_id]) {
      map[dept.parent_id].children.push(node)
    } else {
      roots.push(node)
    }
  }
  return roots
}

const findDeptByID = (nodes, id) => {
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children?.length) {
      const found = findDeptByID(node.children, id)
      if (found) return found
    }
  }
  return null
}

const getFirstDept = (nodes) => {
  if (!nodes.length) return null
  return nodes[0]
}

const fetchAllEmployees = async (params = {}) => {
  const pageSize = 500
  let page = 1
  let total = 0
  const result = []

  while (true) {
    const res = await getEmployees({
      ...params,
      name_only: 1,
      page,
      page_size: pageSize
    })
    const data = res.data?.data || {}
    const list = data.list || []
    total = Number(data.total || 0)
    result.push(...list)
    if (!list.length || result.length >= total || list.length < pageSize) break
    page += 1
  }

  return result.sort((a, b) => (a.id || 0) - (b.id || 0))
}

const clearSelectedRoleEmployees = () => {
  selectedRole.value = null
  roleEmployeeList.value = []
  employeeBindDrawerVisible.value = false
  employeeBindOptions.value = []
  employeeBindSelectedIDs.value = []
}

const loadRoleEmployees = async (role) => {
  if (!role?.id || !selectedDept.value?.id) {
    clearSelectedRoleEmployees()
    return
  }

  selectedRole.value = role
  employeeBindDrawerVisible.value = false
  employeeBindOptions.value = []
  employeeBindSelectedIDs.value = []
  loadingRoleEmployees.value = true
  try {
    roleEmployeeList.value = await fetchAllEmployees({
      department_id: selectedDept.value.id,
      position_id: role.id
    })
  } finally {
    loadingRoleEmployees.value = false
  }
}

const loadDeptTree = async () => {
  const previousID = selectedDept.value?.id || null
  const res = await getDepartments({ page: 1, page_size: 1000 })
  const list = res.data?.data?.list || []
  deptTree.value = buildDeptTree(list)

  if (previousID) {
    const refreshed = findDeptByID(deptTree.value, previousID)
    if (refreshed) {
      selectedDept.value = refreshed
      await nextTick()
      deptTreeRef.value?.setCurrentKey(refreshed.id)
      return
    }
  }

  if (!selectedDept.value) {
    const first = getFirstDept(deptTree.value)
    if (first) {
      selectedDept.value = first
      await nextTick()
      deptTreeRef.value?.setCurrentKey(first.id)
    }
  }
}

const loadAllPositions = async () => {
  const res = await getPositions({ page: 1, page_size: 1000 })
  allPositions.value = res.data?.data?.list || []
}

const loadDeptPositions = async (dept) => {
  if (!dept?.id) {
    positionList.value = []
    clearSelectedRoleEmployees()
    return
  }
  const previousRoleID = selectedRole.value?.id || null
  selectedDept.value = dept
  loadingPositions.value = true
  try {
    const res = await getDepartmentPositions({ department_id: dept.id, page: 1, page_size: 1000 })
    const rels = res.data?.data?.list || []
    positionList.value = rels
      .map((rel) => ({
        ...rel.position,
        relation_id: rel.id
      }))
      .sort((a, b) => (a.id || 0) - (b.id || 0))
  } finally {
    loadingPositions.value = false
  }

  if (!positionList.value.length) {
    clearSelectedRoleEmployees()
    return
  }

  const nextRole = previousRoleID
    ? positionList.value.find((item) => item.id === previousRoleID) || positionList.value[0]
    : positionList.value[0]
  await loadRoleEmployees(nextRole)
}

const handleDeptClick = async (dept) => {
  await loadDeptPositions(dept)
}

const getRoleRowClassName = ({ row }) => {
  if (selectedRole.value?.id === row?.id) return 'role-row-active'
  return ''
}

const handleRoleRowClick = async (row) => {
  await loadRoleEmployees(row)
}

const openEmployeeBindDrawer = async () => {
  const roleID = selectedRole.value?.id
  const departmentID = selectedDept.value?.id
  if (!roleID || !departmentID) return

  employeeBindDrawerVisible.value = true
  employeeBindLoading.value = true
  try {
    const [list, related] = await Promise.all([
      fetchAllEmployees(),
      fetchAllEmployees({
        department_id: departmentID,
        position_id: roleID
      })
    ])
    employeeBindOptions.value = list
    employeeBindSelectedIDs.value = related.map((item) => item.id)
  } finally {
    employeeBindLoading.value = false
  }
}

const handleSaveEmployeeBind = async () => {
  const roleID = selectedRole.value?.id
  const departmentID = selectedDept.value?.id
  if (!roleID || !departmentID) return

  employeeBindSaving.value = true
  try {
    await setPositionEmployees(roleID, {
      department_id: departmentID,
      employee_ids: employeeBindSelectedIDs.value
    })
    ElMessage.success('人员关联保存成功')
    employeeBindDrawerVisible.value = false
    await loadRoleEmployees(selectedRole.value)
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '人员关联保存失败'))
  } finally {
    employeeBindSaving.value = false
  }
}

const openDeptDialog = (dept = null, parentID = null) => {
  if (dept) {
    deptForm.value = {
      id: dept.id,
      name: dept.name,
      parent_id: dept.parent_id ?? null,
      remark: dept.remark || ''
    }
  } else {
    deptForm.value = {
      id: null,
      name: '',
      parent_id: parentID ?? null,
      remark: ''
    }
  }
  deptDialogVisible.value = true
}

const handleDeptSubmit = async () => {
  const name = (deptForm.value.name || '').trim()
  if (!name) {
    ElMessage.warning('请输入部门名称')
    return
  }

  const payload = {
    name,
    parent_id: deptForm.value.parent_id || null,
    remark: deptForm.value.remark || ''
  }

  try {
    if (deptForm.value.id) {
      await updateDepartment(deptForm.value.id, payload)
      ElMessage.success('部门修改成功')
    } else {
      await createDepartment(payload)
      ElMessage.success('部门新增成功')
    }
    deptDialogVisible.value = false
    await loadDeptTree()
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存部门失败'))
  }
}

const handleDeptDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确认删除该部门？', '提示', { type: 'warning' })
    await deleteDepartment(id)
    ElMessage.success('部门删除成功')

    if (selectedDept.value?.id === id) {
      selectedDept.value = null
      positionList.value = []
      clearSelectedRoleEmployees()
    }

    await loadDeptTree()
    if (selectedDept.value?.id) {
      await loadDeptPositions(selectedDept.value)
    }
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(extractErrorMessage(error, '删除部门失败'))
    }
  }
}

const openBindDialog = () => {
  if (!selectedDept.value?.id) {
    ElMessage.warning('请先选择部门')
    return
  }
  bindSelectedPositionIDs.value = []
  bindDialogVisible.value = true
}

const handleBatchBind = async () => {
  if (!selectedDept.value?.id || !bindSelectedPositionIDs.value.length) return

  const departmentID = selectedDept.value.id
  let successCount = 0
  const failedNames = []

  for (const positionID of bindSelectedPositionIDs.value) {
    try {
      await createDepartmentPosition({ department_id: departmentID, position_id: positionID })
      successCount++
    } catch {
      const matched = allPositions.value.find((item) => item.id === positionID)
      failedNames.push(matched?.name || String(positionID))
    }
  }

  if (successCount > 0) {
    ElMessage.success(`成功关联 ${successCount} 个角色`)
  }
  if (failedNames.length > 0) {
    ElMessage.warning(`以下角色关联失败：${failedNames.join('、')}`)
  }

  bindDialogVisible.value = false
  await loadDeptPositions(selectedDept.value)
}

const openPositionDialog = (row = null) => {
  if (row) {
    positionForm.value = {
      id: row.id,
      name: row.name,
      remark: row.remark || ''
    }
  } else {
    positionForm.value = {
      id: null,
      name: '',
      remark: ''
    }
  }
  positionDialogVisible.value = true
}

const handlePositionSubmit = async () => {
  const name = (positionForm.value.name || '').trim()
  if (!name) {
    ElMessage.warning('请输入角色名称')
    return
  }
  if (!positionForm.value.id && !selectedDept.value?.id) {
    ElMessage.warning('请先选择部门')
    return
  }

  positionSubmitting.value = true
  try {
    const payload = {
      name,
      remark: positionForm.value.remark || ''
    }

    if (positionForm.value.id) {
      await updatePosition(positionForm.value.id, payload)
      ElMessage.success('角色修改成功')
    } else {
      await createPosition({
        ...payload,
        department_id: selectedDept.value.id
      })
      ElMessage.success('角色创建成功')
    }

    positionDialogVisible.value = false
    await Promise.all([loadAllPositions(), loadDeptPositions(selectedDept.value)])
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存角色失败'))
  } finally {
    positionSubmitting.value = false
  }
}

const handleRemoveRelation = async (row) => {
  if (!row.relation_id) return
  try {
    await ElMessageBox.confirm(`确认移除角色【${row.name}】与当前部门的关系？`, '提示', { type: 'warning' })
    await deleteDepartmentPosition(row.relation_id)
    ElMessage.success('关系移除成功')
    await loadDeptPositions(selectedDept.value)
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(extractErrorMessage(error, '移除关系失败'))
    }
  }
}

const handleDeletePosition = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确认删除角色【${row.name}】？删除后会清理该角色在全部部门中的关联关系。`,
      '高风险操作',
      { type: 'warning' }
    )
    await deletePosition(row.id)
    ElMessage.success('角色删除成功')
    await Promise.all([loadAllPositions(), loadDeptPositions(selectedDept.value)])
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(extractErrorMessage(error, '删除角色失败'))
    }
  }
}

const buildPermParentMap = (nodes = [], parentID = 0) => {
  for (const node of nodes) {
    permParentMap.value[node.id] = parentID
    if (node.children?.length) buildPermParentMap(node.children, node.id)
  }
}

const collectMenuIDs = (nodes = []) => {
  const ids = []
  const walk = (arr) => {
    for (const node of arr) {
      ids.push(node.id)
      if (node.children?.length) walk(node.children)
    }
  }
  walk(nodes)
  return ids
}

const collectLeafMenuIDs = (nodes = []) => {
  const ids = []
  const walk = (arr) => {
    for (const node of arr) {
      if (node.children?.length) {
        walk(node.children)
      } else {
        ids.push(node.id)
      }
    }
  }
  walk(nodes)
  return ids
}

const includeParentMenuIDs = (keys = []) => {
  const set = new Set(keys)
  for (const id of keys) {
    let parentID = permParentMap.value[id]
    while (parentID && parentID > 0) {
      set.add(parentID)
      parentID = permParentMap.value[parentID]
    }
  }
  return Array.from(set)
}

const cloneMenuTreeWithDisabled = (nodes = [], disabled = false) => {
  return (nodes || []).map((node) => ({
    ...node,
    disabled,
    children: cloneMenuTreeWithDisabled(node.children || [], disabled)
  }))
}

const openPermDrawer = async (row, options = {}) => {
  const readOnly = options?.readOnly === true
  const employeeName = options?.employeeName || ''
  const targetType = options?.targetType || 'position'
  const targetTypeText = targetType === 'department' ? '部门' : '菜单'

  permReadOnly.value = readOnly
  permDrawerTitle.value = employeeName
    ? `${readOnly ? '权限查看' : '权限设置'} - ${employeeName}`
    : `${readOnly ? `${targetTypeText}权限查看` : `${targetTypeText}权限设置`} - ${row.name || ''}`

  activePermissionTarget.value = { type: targetType, id: row.id, name: row.name }
  permMenuTree.value = []
  permParentMap.value = {}
  allMenuIDs.value = []
  permDrawerVisible.value = true
  permLoading.value = true

  try {
    const res = targetType === 'department'
      ? await getDepartmentMenuPermissions(row.id)
      : await getPositionMenuPermissions(row.id)
    const data = res.data?.data || {}
    const menuTree = data.menu_tree || []
    permMenuTree.value = readOnly ? cloneMenuTreeWithDisabled(menuTree, true) : menuTree
    buildPermParentMap(permMenuTree.value)
    allMenuIDs.value = collectMenuIDs(permMenuTree.value)

    const checkedIDs = data.checked_menu_ids || []
    const leafIDSet = new Set(collectLeafMenuIDs(permMenuTree.value))
    const checkedLeafIDs = checkedIDs.filter((id) => leafIDSet.has(id))
    await nextTick()
    menuTreeRef.value?.setCheckedKeys(checkedLeafIDs)
  } catch (error) {
    permDrawerVisible.value = false
    ElMessage.error(extractErrorMessage(error, '加载权限失败'))
  } finally {
    permLoading.value = false
  }
}

const openDeptPermDrawer = async (dept) => {
  if (!dept?.id) return
  await openPermDrawer(dept, { targetType: 'department' })
}


const openEmployeePermEdit = async (employee) => {
  if (!selectedRole.value?.id) return
  await openPermDrawer(selectedRole.value, { readOnly: false, employeeName: employee?.name || '' })
}

const handleCheckAllMenus = () => {
  menuTreeRef.value?.setCheckedKeys(allMenuIDs.value)
}

const handleClearAllMenus = () => {
  menuTreeRef.value?.setCheckedKeys([])
}

const handleSavePermissions = async () => {
  if (!activePermissionTarget.value.id) return

  permSaving.value = true
  try {
    const checkedKeys = menuTreeRef.value?.getCheckedKeys(false) || []
    const menuIDs = includeParentMenuIDs(checkedKeys)
    if (activePermissionTarget.value.type === 'department') {
      await setDepartmentMenuPermissions(activePermissionTarget.value.id, { menu_ids: menuIDs })
    } else {
      await setPositionMenuPermissions(activePermissionTarget.value.id, { menu_ids: menuIDs })
    }
    ElMessage.success('权限保存成功')
    permDrawerVisible.value = false
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '保存权限失败'))
  } finally {
    permSaving.value = false
  }
}

onMounted(async () => {
  try {
    await Promise.all([loadDeptTree(), loadAllPositions()])
    if (selectedDept.value?.id) {
      await loadDeptPositions(selectedDept.value)
    }
  } catch (error) {
    ElMessage.error(extractErrorMessage(error, '角色管理页面初始化失败'))
  }
})
</script>

<style scoped>
.role-page {
  display: flex;
  gap: 16px;
  height: calc(100vh - 240px);
}

.dept-panel {
  width: 300px;
  border-right: 1px solid #ebeef5;
  overflow: auto;
  padding: 8px 10px;
}

.dept-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  padding: 8px 4px 10px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 8px;
}

.dept-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding-right: 4px;
}

.dept-node-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dept-node-actions {
  visibility: hidden;
  display: inline-flex;
  gap: 4px;
}

:deep(.el-tree-node__content:hover) .dept-node-actions {
  visibility: visible;
}

.dept-node-actions .el-button {
  background: transparent;
  border-color: transparent;
}

.dept-node-actions .el-button:hover {
  background: var(--el-button-bg-color);
  border-color: var(--el-button-border-color);
}

.role-row-actions,
.employee-row-actions {
  visibility: hidden;
  display: inline-flex;
  gap: 4px;
}

:deep(.el-table__row:hover) .role-row-actions,
:deep(.el-table__row:hover) .employee-row-actions,
:deep(.role-row-active) .role-row-actions {
  visibility: visible;
}

.role-row-actions .el-button,
.employee-row-actions .el-button {
  background: transparent;
  border-color: transparent;
}

.role-row-actions .el-button:hover,
.employee-row-actions .el-button:hover {
  background: var(--el-button-bg-color);
  border-color: var(--el-button-border-color);
}

.role-panel {
  width: 520px;
  overflow: auto;
  padding: 12px;
  border-right: 1px solid #ebeef5;
}

.employee-panel {
  flex: 1;
  overflow: auto;
  padding: 12px;
}

.role-toolbar,
.employee-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.role-toolbar-title,
.employee-toolbar-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.role-toolbar-actions,
.employee-toolbar-actions {
  display: inline-flex;
  gap: 8px;
}

.empty-placeholder {
  text-align: center;
  color: #909399;
  padding: 60px 0;
}

.bind-dialog-body {
  max-height: 420px;
  overflow: auto;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 10px 12px;
}

.bind-item {
  padding: 6px 0;
  border-bottom: 1px dashed #f0f2f5;
}

.bind-item:last-child {
  border-bottom: none;
}

:deep(.role-row-active > td) {
  background: #ecf5ff !important;
}

.employee-bind-drawer-body {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.employee-bind-list {
  flex: 1;
  overflow: auto;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 10px 12px;
}

.employee-bind-item {
  padding: 6px 0;
  border-bottom: 1px dashed #f0f2f5;
}

.employee-bind-item:last-child {
  border-bottom: none;
}

.employee-bind-footer {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.perm-drawer-body {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.perm-header-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-bottom: 10px;
}

.perm-tree {
  flex: 1;
  overflow: auto;
}

.perm-footer-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>

