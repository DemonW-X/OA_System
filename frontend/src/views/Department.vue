<template>
  <!-- 职位管理 -->
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
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPositions, createPosition, updatePosition, deletePosition } from '../api/position'

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
  loadPositions()
})
</script>
