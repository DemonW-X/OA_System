<template>
  <el-card>
    <template #header>
      <span>操作日志</span>
    </template>

    <el-form :inline="true" style="margin-bottom:16px">
      <el-form-item label="操作人">
        <el-input v-model="query.username" placeholder="请输入操作人" clearable />
      </el-form-item>
      <el-form-item label="模块">
        <el-select v-model="query.module" placeholder="全部" clearable style="width:120px">
          <el-option label="部门管理" value="部门管理" />
          <el-option label="职位管理" value="职位管理" />
          <el-option label="员工管理" value="员工管理" />
          <el-option label="公告管理" value="公告管理" />
        </el-select>
      </el-form-item>
      <el-form-item label="操作类型">
        <el-select v-model="query.action" placeholder="全部" clearable style="width:110px">
          <el-option label="查询" value="查询" />
          <el-option label="新增" value="新增" />
          <el-option label="修改" value="修改" />
          <el-option label="删除" value="删除" />
        </el-select>
      </el-form-item>
      <el-form-item label="开始时间">
        <el-date-picker v-model="query.start_time" type="datetime" placeholder="选择开始时间"
          value-format="YYYY-MM-DD HH:mm:ss" style="width:180px" />
      </el-form-item>
      <el-form-item label="结束时间">
        <el-date-picker v-model="query.end_time" type="datetime" placeholder="选择结束时间"
          value-format="YYYY-MM-DD HH:mm:ss" style="width:180px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="username" label="操作人" width="100" />
      <el-table-column prop="module" label="模块" width="100" />
      <el-table-column prop="action" label="操作类型" width="90">
        <template #default="{ row }">
          <el-tag :type="actionTagType(row.action)" size="small">{{ row.action }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="操作描述" />
      <el-table-column prop="method" label="请求方法" width="90" />
      <el-table-column prop="path" label="请求路径" width="180" />
      <el-table-column prop="ip" label="IP" width="130" />
      <el-table-column prop="created_at" label="操作时间" width="170">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
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
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getLogs } from '../api/log'

const list = ref([])
const total = ref(0)
const query = ref({
  username: '', module: '', action: '',
  start_time: '', end_time: '',
  page: 1, page_size: 10
})

const actionTagType = (action) => {
  const map = { '新增': 'success', '修改': 'warning', '删除': 'danger', '查询': 'info' }
  return map[action] || ''
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit'
  })
}

const loadData = async () => {
  const res = await getLogs(query.value)
  list.value = res.data.data.list || []
  total.value = res.data.data.total || 0
}

const handleSearch = () => { query.value.page = 1; loadData() }
const handleReset = () => {
  query.value = { username: '', module: '', action: '', start_time: '', end_time: '', page: 1, page_size: 10 }
  loadData()
}

onMounted(loadData)
</script>
