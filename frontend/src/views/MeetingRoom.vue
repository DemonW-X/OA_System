<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="名称">
        <el-input v-model="query.name" placeholder="请输入会议室名称" clearable />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable style="width:100px">
          <el-option label="可用" :value="1" />
          <el-option label="停用" :value="0" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="会议室名称" min-width="120" />
      <el-table-column prop="location" label="位置" min-width="120" />
      <el-table-column prop="capacity" label="容纳人数" width="100">
        <template #default="{ row }">{{ row.capacity || '-' }} 人</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? '可用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
      <el-table-column label="操作" width="160">
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑会议室' : '新增会议室'" width="560px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="会议室名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入会议室名称" />
        </el-form-item>
        <el-form-item label="位置">
          <el-input v-model="form.location" placeholder="请输入位置（如：3楼A区）" />
        </el-form-item>
        <el-form-item label="容纳人数">
          <el-input-number v-model="form.capacity" :min="0" :max="999" style="width:100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">可用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注信息（选填）" />
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
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMeetingRooms, createMeetingRoom, updateMeetingRoom, deleteMeetingRoom } from '../api/meeting_room'

const list = ref([])
const total = ref(0)
const dialogVisible = ref(false)
const formRef = ref()
const query = ref({ name: '', status: null, page: 1, page_size: 10 })
const defaultForm = () => ({ name: '', location: '', capacity: 10, status: 1, remark: '' })
const form = ref(defaultForm())
const rules = {
  name: [{ required: true, message: '请输入会议室名称', trigger: 'blur' }]
}

const loadData = async () => {
  const params = { ...query.value }
  if (params.status === null) delete params.status
  const res = await getMeetingRooms(params)
  list.value = res.data.data.list || []
  total.value = res.data.data.total || 0
}

const handleSearch = () => { query.value.page = 1; loadData() }
const handleReset = () => { query.value = { name: '', status: null, page: 1, page_size: 10 }; loadData() }

const openDialog = (row = null) => {
  form.value = row ? { ...row } : defaultForm()
  dialogVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value.validate()
  if (form.value.id) {
    await updateMeetingRoom(form.value.id, form.value)
  } else {
    await createMeetingRoom(form.value)
  }
  ElMessage.success('操作成功')
  dialogVisible.value = false
  loadData()
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该会议室？', '提示', { type: 'warning' })
  await deleteMeetingRoom(id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(loadData)
</script>
