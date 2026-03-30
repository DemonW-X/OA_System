<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="关键字">
        <el-input v-model="query.keyword" placeholder="菜单名称/路径" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadData">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="treeData" row-key="id" border default-expand-all>
      <el-table-column prop="name" label="菜单名称" min-width="180" />
      <el-table-column label="图标" width="140">
        <template #default="{ row }">
          <div style="display:flex;align-items:center;gap:6px">
            <el-icon><component :is="resolveIcon(row.icon)" /></el-icon>
            <span>{{ row.icon || 'Menu' }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="path" label="页面URL" min-width="200" />
      <el-table-column prop="sort_code" label="排序编码" width="100" />
      <el-table-column label="是否显示" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.visible ? 'success' : 'info'">{{ row.visible ? '显示' : '隐藏' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="160" />
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑菜单' : '新增菜单'" width="520px">
      <el-form :model="form" label-width="96px">
        <el-form-item label="菜单名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="菜单图标">
          <el-input v-model="form.icon" placeholder="请选择图标" readonly>
            <template #prepend>
              <el-icon><component :is="resolveIcon(form.icon)" /></el-icon>
            </template>
            <template #append>
              <el-button @click="iconDialogVisible = true">选择</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="页面URL">
          <el-input v-model="form.path" placeholder="如 /employee" />
        </el-form-item>
        <el-form-item label="父级菜单">
          <el-tree-select
            v-model="form.parent_id"
            :data="parentOptions"
            :props="{ label: 'name', value: 'id', children: 'children' }"
            clearable
            check-strictly
            node-key="id"
            placeholder="顶级菜单可留空"
            style="width:100%"
          />
        </el-form-item>
        <el-form-item label="排序编码">
          <el-input-number v-model="form.sort_code" :min="0" style="width:100%" />
        </el-form-item>
        <el-form-item label="是否显示">
          <div style="display:flex;align-items:center;gap:24px">
            <el-switch v-model="form.visible" />
            <span style="color:#606266;font-size:13px">启用审批流</span>
            <el-switch v-model="form.enable_workflow" />
          </div>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" />
        </el-form-item>
        <template v-if="form.enable_workflow">
          <el-form-item label="业务编码" required>
            <el-input v-model="form.biz_code" placeholder="如 leave_request、event_booking" />
          </el-form-item>
          <el-form-item label="业务名称" required>
            <el-input v-model="form.biz_name" placeholder="如 请假审批" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="form.biz_sort" :min="0" style="width:100%" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="iconDialogVisible" title="选择菜单图标" width="680px">
      <el-input v-model="iconKeyword" placeholder="输入图标名搜索，如 User / Menu / Setting" clearable style="margin-bottom:12px" />
      <div class="icon-grid">
        <div
          v-for="item in filteredIcons"
          :key="item"
          class="icon-item"
          :class="{ active: form.icon === item }"
          @click="selectIcon(item)"
        >
          <el-icon><component :is="resolveIcon(item)" /></el-icon>
          <div class="icon-name">{{ item }}</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="iconDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import * as Icons from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMenus, createMenu, updateMenu, deleteMenu } from '../api/menu'

const query = ref({ keyword: '' })
const treeData = ref([])
const dialogVisible = ref(false)
const iconDialogVisible = ref(false)
const iconKeyword = ref('')
const form = ref({ id: null, name: '', icon: 'Menu', path: '', parent_id: null, sort_code: 0, visible: true, remark: '', enable_workflow: false, biz_code: '', biz_name: '', biz_sort: 0 })

const iconList = Object.keys(Icons).sort((a, b) => a.localeCompare(b))

const filteredIcons = computed(() => {
  const kw = (iconKeyword.value || '').trim().toLowerCase()
  if (!kw) return iconList
  return iconList.filter(n => n.toLowerCase().includes(kw))
})

const resolveIcon = (iconName) => {
  if (iconName && Icons[iconName]) return Icons[iconName]
  return Icons.Menu
}

const selectIcon = (name) => {
  form.value.icon = name
  iconDialogVisible.value = false
}

const parentOptions = computed(() => {
  const clone = JSON.parse(JSON.stringify(treeData.value || []))
  const walk = (arr) => {
    for (const n of arr) {
      if (form.value.id && n.id === form.value.id) {
        n.disabled = true
      }
      if (n.children?.length) walk(n.children)
    }
  }
  walk(clone)
  return [{ id: 0, name: '顶级菜单', children: clone }]
})

const loadData = async () => {
  const res = await getMenus({ tree: 1, keyword: query.value.keyword || '' })
  treeData.value = res.data?.data || []
}

const handleReset = () => {
  query.value.keyword = ''
  loadData()
}

const openDialog = (row = null) => {
  if (!row) {
    form.value = { id: null, name: '', icon: 'Menu', path: '', parent_id: 0, sort_code: 0, visible: true, remark: '', enable_workflow: false, biz_code: '', biz_name: '', biz_sort: 0 }
  } else {
    form.value = {
      id: row.id,
      name: row.name,
      icon: row.icon || 'Menu',
      path: row.path || '',
      parent_id: row.parent_id ?? 0,
      sort_code: row.sort_code || 0,
      visible: row.visible !== false,
      remark: row.remark || '',
      enable_workflow: !!row.enable_workflow,
      biz_code: row.biz_code || '',
      biz_name: row.biz_name || '',
      biz_sort: row.biz_sort || 0
    }
  }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.value.name) return ElMessage.warning('请输入菜单名称')
  if (form.value.enable_workflow && (!form.value.biz_code || !form.value.biz_name)) {
    return ElMessage.warning('启用审批流时，业务编码和业务名称不能为空')
  }
  const payload = {
    name: form.value.name,
    icon: form.value.icon,
    path: form.value.path,
    parent_id: Number(form.value.parent_id || 0),
    sort_code: Number(form.value.sort_code || 0),
    visible: !!form.value.visible,
    remark: form.value.remark,
    enable_workflow: !!form.value.enable_workflow,
    biz_code: form.value.biz_code || '',
    biz_name: form.value.biz_name || '',
    biz_sort: Number(form.value.biz_sort || 0)
  }
  if (form.value.id) await updateMenu(form.value.id, payload)
  else await createMenu(payload)

  ElMessage.success('操作成功')
  dialogVisible.value = false
  loadData()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm(`确认删除菜单「${row.name}」？`, '提示', { type: 'warning' })
  await deleteMenu(row.id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(loadData)
</script>

<style scoped>
.icon-grid {
  max-height: 420px;
  overflow: auto;
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 10px;
}

.icon-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 10px 6px;
  text-align: center;
  cursor: pointer;
  user-select: none;
}

.icon-item:hover,
.icon-item.active {
  border-color: #409eff;
  background: #ecf5ff;
}

.icon-name {
  margin-top: 6px;
  font-size: 12px;
  color: #606266;
  word-break: break-all;
}
</style>
