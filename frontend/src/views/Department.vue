<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="职位名称">
        <el-input v-model="posQuery.name" placeholder="请输入职位名称" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handlePosSearch">搜索</el-button>
        <el-button @click="handlePosReset">重置</el-button>
      </el-form-item>
      <el-form-item style="margin-left:auto;margin-right:0">
        <el-button type="primary" @click="openPosDialog()">新增职位</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="posList" stripe>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="职位名称" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button size="small" @click="openPosDialog(row)">编辑</el-button>
          <el-button size="small" type="warning" @click="openPermDialog(row)">权限设置</el-button>
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

  <el-dialog v-model="posDialogVisible" :title="posForm.id ? '编辑职位' : '新增职位'" width="400px">
    <el-form :model="posForm" label-width="80px">
      <el-form-item label="职位名称">
        <el-input v-model="posForm.name" />
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

  <el-drawer
    v-model="permDialogVisible"
    :title="`权限设置 - ${activePosition.name || ''}`"
    direction="rtl"
    size="420px"
    destroy-on-close
  >
    <div v-loading="permLoading" style="height:100%;display:flex;flex-direction:column">
      <div style="display:flex;justify-content:flex-end;gap:8px;margin-bottom:12px">
        <el-button size="small" @click="handleCheckAll">全选</el-button>
        <el-button size="small" @click="handleClearAll">清空</el-button>
      </div>
      <el-tree
        ref="treeRef"
        :data="menuTree"
        show-checkbox
        node-key="id"
        default-expand-all
        :props="{ label: 'name', children: 'children' }"
        style="flex:1;overflow-y:auto"
      />
      <div style="padding:16px 0;display:flex;justify-content:flex-end;gap:8px;border-top:1px solid #eee;margin-top:12px">
        <el-button @click="permDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="permSaving" @click="handlePermSave">保存</el-button>
      </div>
    </div>
  </el-drawer>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getPositions,
  createPosition,
  updatePosition,
  deletePosition,
  getPositionMenuPermissions,
  setPositionMenuPermissions
} from '../api/position'

const posList = ref([])
const posTotal = ref(0)
const posDialogVisible = ref(false)
const posForm = ref({ name: '', remark: '' })
const posQuery = ref({ name: '', page: 1, page_size: 10 })

const permDialogVisible = ref(false)
const permLoading = ref(false)
const permSaving = ref(false)
const activePosition = ref({ id: null, name: '' })
const menuTree = ref([])
const treeRef = ref()
const parentMap = ref({})
const allMenuIds = ref([])

const loadPositions = async () => {
  const res = await getPositions(posQuery.value)
  posList.value = res.data?.data?.list || []
  posTotal.value = res.data?.data?.total || 0
}

const handlePosSearch = () => {
  posQuery.value.page = 1
  loadPositions()
}

const handlePosReset = () => {
  posQuery.value = { name: '', page: 1, page_size: 10 }
  loadPositions()
}

const openPosDialog = (row = null) => {
  posForm.value = row
    ? { id: row.id, name: row.name, remark: row.remark || '' }
    : { name: '', remark: '' }
  posDialogVisible.value = true
}

const handlePosSubmit = async () => {
  const name = (posForm.value.name || '').trim()
  if (!name) {
    ElMessage.warning('请输入职位名称')
    return
  }

  const payload = { ...posForm.value, name }
  if (payload.id) {
    await updatePosition(payload.id, payload)
  } else {
    await createPosition(payload)
  }
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

const openPermDialog = async (row) => {
  activePosition.value = { id: row.id, name: row.name }
  permDialogVisible.value = true
  permLoading.value = true
  menuTree.value = []
  parentMap.value = {}
  allMenuIds.value = []

  try {
    const res = await getPositionMenuPermissions(row.id)
    const data = res.data?.data || {}
    menuTree.value = data.menu_tree || []
    buildParentMap(menuTree.value)
    allMenuIds.value = collectAllMenuIds(menuTree.value)

    const checked = data.checked_menu_ids || []
    await nextTick()
    treeRef.value?.setCheckedKeys(includeParentKeys(checked))
  } finally {
    permLoading.value = false
  }
}

const handleCheckAll = () => {
  treeRef.value?.setCheckedKeys(allMenuIds.value)
}

const handleClearAll = () => {
  treeRef.value?.setCheckedKeys([])
}

const handlePermSave = async () => {
  if (!activePosition.value.id) return

  permSaving.value = true
  try {
    const checked = treeRef.value?.getCheckedKeys(false) || []
    const menuIds = includeParentKeys(checked)
    await setPositionMenuPermissions(activePosition.value.id, { menu_ids: menuIds })
    ElMessage.success('权限保存成功')
    permDialogVisible.value = false
  } finally {
    permSaving.value = false
  }
}

onMounted(() => {
  loadPositions()
})
</script>
