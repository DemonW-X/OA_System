<template>
  <div>
    <!-- 统计卡片 -->
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:40px;color:#409EFF;font-weight:bold">{{ stats.departments }}</div>
            <div style="color:#999;margin-top:8px">部门总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:40px;color:#67C23A;font-weight:bold">{{ stats.positions }}</div>
            <div style="color:#999;margin-top:8px">职位总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:40px;color:#E6A23C;font-weight:bold">{{ stats.employees }}</div>
            <div style="color:#999;margin-top:8px">员工总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:40px;color:#F56C6C;font-weight:bold">{{ stats.notices }}</div>
            <div style="color:#999;margin-top:8px">公告总数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 待我审核 -->
    <el-card shadow="hover" style="margin-top:20px">
      <template #header>
        <div style="display:flex;align-items:center;gap:8px">
          <el-icon color="#E6A23C"><Bell /></el-icon>
          <span style="font-weight:bold">待我审核</span>
          <el-tag size="small" type="warning" style="margin-left:auto">{{ pendingApprovals.length }} 条</el-tag>
        </div>
      </template>
      <el-empty v-if="pendingApprovals.length === 0" description="暂无待审核事项" :image-size="60" />
      <el-table v-else :data="pendingApprovals" size="small" border>
        <el-table-column prop="title" label="事项" min-width="260" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link type="primary" @click="openApproval(row)">{{ row.title }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="biz_type" label="业务类型" width="120" />
        <el-table-column prop="node_key" label="当前节点" width="120" />
        <el-table-column prop="created_at" label="待办时间" width="180" />
      </el-table>
    </el-card>

    <!-- 公告栏 -->
    <el-card shadow="hover" style="margin-top:20px">
      <template #header>
        <div style="display:flex;align-items:center;gap:8px">
          <el-icon color="#409EFF"><Bell /></el-icon>
          <span style="font-weight:bold">公告栏</span>
          <el-tag size="small" type="info" style="margin-left:auto">最新 {{ noticeList.length }} 条</el-tag>
        </div>
      </template>
      <el-empty v-if="noticeList.length === 0" description="暂无公告" :image-size="60" />
      <ul v-else class="notice-list">
        <li v-for="item in noticeList" :key="item.id" class="notice-item">
          <el-link type="primary" @click="openNotice(item)">{{ item.title }}</el-link>
          <span class="notice-meta">{{ item.author }} · {{ formatDate(item.created_at) }}</span>
        </li>
      </ul>
    </el-card>

    <!-- 公告详情弹窗 -->
    <el-dialog v-model="dialogVisible" width="600px" :title="currentNotice.title">
      <div style="margin-bottom:12px">
        <el-tag type="success" v-if="currentNotice.status === 1">已发布</el-tag>
        <el-tag type="info" v-else>草稿</el-tag>
        <span style="color:#999;font-size:13px;margin-left:10px">
          作者：{{ currentNotice.author }} · {{ formatDate(currentNotice.created_at) }}
        </span>
      </div>
      <el-divider />
      <div v-if="currentNotice.content" class="notice-content" v-html="currentNotice.content" />
      <div v-else style="color:#999;text-align:center;padding:20px">暂无内容</div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell } from '@element-plus/icons-vue'
import { getDepartments } from '../api/department'
import { getEmployees } from '../api/employee'
import { getNotices } from '../api/notice'
import { getPositions } from '../api/position'
import { getMyPendingApprovals } from '../api/orchid_workflow'

const router = useRouter()

const stats = ref({ departments: 0, positions: 0, employees: 0, notices: 0 })
const pendingApprovals = ref([])
const noticeList = ref([])
const dialogVisible = ref(false)
const currentNotice = ref({})

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit'
  })
}

const openNotice = (item) => {
  currentNotice.value = item
  dialogVisible.value = true
}

const openApproval = (item) => {
  if (!item?.detail_path) return
  router.push(item.detail_path)
}

onMounted(async () => {
  const [d, p, e, n, notices, approvals] = await Promise.all([
    getDepartments({ page: 1, page_size: 1 }),
    getPositions({ page: 1, page_size: 1 }),
    getEmployees({ page: 1, page_size: 1 }),
    getNotices({ page: 1, page_size: 1 }),
    getNotices({ page: 1, page_size: 10, status: 1 }),
    getMyPendingApprovals({ limit: 10 })
  ])
  stats.value.departments = d.data.data?.total || 0
  stats.value.positions   = p.data.data?.total || 0
  stats.value.employees   = e.data.data?.total || 0
  stats.value.notices     = n.data.data?.total || 0
  noticeList.value        = notices.data.data?.list || []
  pendingApprovals.value  = approvals.data?.data || []
})
</script>

<style scoped>
.notice-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.notice-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}
.notice-item:last-child {
  border-bottom: none;
}
.notice-meta {
  font-size: 12px;
  color: #999;
  white-space: nowrap;
  margin-left: 12px;
}
.notice-content {
  line-height: 1.8;
  color: #333;
  min-height: 80px;
}
.notice-content :deep(img) {
  max-width: 100%;
  height: auto;
}
.notice-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
}
.notice-content :deep(td),
.notice-content :deep(th) {
  border: 1px solid #ddd;
  padding: 6px 10px;
}
.notice-content :deep(blockquote) {
  border-left: 4px solid #ddd;
  margin: 0;
  padding-left: 16px;
  color: #666;
}
</style>
