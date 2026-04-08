<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="标题">
        <el-input v-model="query.title" placeholder="请输入标题" clearable />
      </el-form-item>
      <el-form-item label="部门">
        <el-select v-model="query.department_id" placeholder="全部部门" clearable style="width:130px">
          <el-option v-for="d in deptList" :key="d.id" :label="d.name" :value="d.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable style="width:100px">
          <el-option label="已发布" :value="1" />
          <el-option label="草稿" :value="0" />
        </el-select>
      </el-form-item>
      <el-form-item label="审批状态">
        <el-select v-model="query.approve_status" placeholder="全部" clearable style="width:130px">
          <el-option label="草稿" value="draft" />
          <el-option label="待审批" value="pending" />
          <el-option label="已通过" value="approved" />
          <el-option label="已拒绝" value="rejected" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
      <el-form-item style="margin-left:auto;margin-right:0">
        <el-button type="primary" @click="openDialog()">新增公告</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" min-width="160" />
      <el-table-column label="部门" width="130">
        <template #default="{ row }">{{ row.department ? row.department.name : '全部部门' }}</template>
      </el-table-column>
      <el-table-column prop="author" label="作者" width="110" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '已发布' : '草稿' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="审批状态" width="100">
        <template #default="{ row }">
          <el-tag :type="approveTagType(row.approve_status)">{{ approveStatusLabel(row.approve_status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="220">
        <template #default="{ row }">
          <el-button size="small" @click="openPreview(row)">预览</el-button>
          <el-button size="small" :disabled="!isEditable(row)" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" :disabled="!isEditable(row)" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
      <el-table-column label="审核" width="280">
        <template #default="{ row }">
          <el-button size="small" type="success" :disabled="!isEditable(row)" @click="handleSubmitAudit(row)">提交审核</el-button>
          <el-button size="small" type="warning" :disabled="row.approve_status !== 'pending'" @click="handleWithdraw(row)">撤回</el-button>
          <el-button size="small" type="info" :disabled="row.approve_status !== 'approved'" @click="handleCancelApprove(row)">弃审</el-button>
        </template>
      </el-table-column>
      <el-table-column label="查看记录" width="120">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openFlowDialog(row)">流程记录</el-button>
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

    <el-dialog
      v-model="dialogVisible"
      :title="form.id ? '编辑公告' : '新增公告'"
      width="860px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="请输入公告标题" />
        </el-form-item>
        <el-form-item label="作者">
          <el-input v-model="form.author" disabled />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="form.department_id" placeholder="全部部门（不限）" clearable style="width:200px">
            <el-option v-for="d in deptList" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
          <span style="margin-left:8px;color:#999;font-size:12px">不选则对全部部门可见</span>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option label="已发布" :value="1" />
            <el-option label="草稿" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容">
          <div style="border:1px solid #ccc;width:100%;border-radius:4px;overflow:hidden">
            <Toolbar :editor="editorRef" :defaultConfig="toolbarConfig" style="border-bottom:1px solid #ccc" />
            <Editor v-model="form.content" :defaultConfig="editorConfig" style="height:360px;overflow-y:hidden" @onCreated="handleEditorCreated" />
          </div>
        </el-form-item>
        <el-form-item label="附件">
          <div style="width:100%">
            <el-upload :http-request="handleAttachmentUpload" :show-file-list="false" :accept="attachmentAccept" multiple>
              <el-button size="small">上传附件</el-button>
              <template #tip>
                <div style="color:#999;font-size:12px;margin-top:4px">支持 PDF、Word、Excel、PPT、TXT、ZIP、RAR，单文件不超过 20MB</div>
              </template>
            </el-upload>
            <div v-if="attachments.length" style="margin-top:8px">
              <el-tag v-for="(att, idx) in attachments" :key="idx" closable style="margin:4px" @close="removeAttachment(idx)">
                <a :href="att.url" target="_blank" style="text-decoration:none;color:inherit">{{ att.name }}</a>
              </el-tag>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewVisible" :title="previewData.title" width="800px">
      <div style="color:#999;font-size:13px;margin-bottom:12px">
        作者：{{ previewData.author }}
        <span v-if="previewData.department" style="margin-left:16px">部门：{{ previewData.department.name }}</span>
      </div>
      <div class="notice-preview" v-html="previewData.content" />
      <div v-if="previewAttachments.length" style="margin-top:16px;border-top:1px solid #eee;padding-top:12px">
        <div style="font-weight:600;margin-bottom:8px">附件：</div>
        <div v-for="(att, idx) in previewAttachments" :key="idx" style="margin-bottom:4px">
          <a :href="att.url" target="_blank">{{ att.name }}</a>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="flowVisible" :title="`审批流程记录 - ${flowTarget.name || ''}`" width="860px">
      <el-skeleton :loading="flowLoading" animated :rows="6">
        <template #default>
          <el-empty v-if="!(flowData.instances || []).length" description="暂无流程记录" :image-size="72" />
          <el-tabs v-else type="border-card">
            <el-tab-pane label="当前流程">
              <div style="margin-bottom:12px;display:flex;gap:12px;align-items:center">
                <el-tag :type="flowInstanceTagType(flowData.instance?.status)">{{ flowInstanceStatusLabel(flowData.instance?.status) }}</el-tag>
                <span style="color:#606266">当前节点：{{ nodeKeyToName(flowData.instance?.current_node) || '-' }}</span>
                <span style="color:#909399;font-size:12px">提交时间：{{ formatTime(flowData.instance?.started_at) }}</span>
              </div>

              <el-empty v-if="!(flowData.instances?.[flowData.instances.length-1]?.histories || []).length" description="暂无操作记录" :image-size="48" />
              <el-timeline v-else>
                <el-timeline-item
                  v-for="(h, idx) in flowData.instances[flowData.instances.length-1].histories"
                  :key="idx"
                  :timestamp="formatTime(h.created_at)"
                  :type="timelineItemType(h.action)"
                  placement="top"
                >
                  <div style="font-weight:600">{{ timelineTitle(h) }}</div>
                  <div style="color:#666;margin-top:4px">处理人：{{ h.operator || '-' }}</div>
                  <div v-if="h.remark" style="color:#909399;margin-top:4px">意见：{{ h.remark }}</div>
                </el-timeline-item>
              </el-timeline>

              <template v-if="(flowData.tasks || []).length">
                <el-divider content-position="left">当前待办</el-divider>
                <el-table :data="flowData.tasks" size="small" border>
                  <el-table-column label="节点" min-width="130">
                    <template #default="{ row }">{{ nodeKeyToName(row.node_key) }}</template>
                  </el-table-column>
                  <el-table-column label="待审批人" min-width="130">
                    <template #default="{ row }">{{ assigneeIdToName(row.assignee_id) }}</template>
                  </el-table-column>
                  <el-table-column label="状态" width="90">
                    <template #default="{ row }"><el-tag size="small" type="warning">待处理</el-tag></template>
                  </el-table-column>
                </el-table>
              </template>
            </el-tab-pane>

            <el-tab-pane :label="`提交历史（${flowData.instances.length} 次）`">
              <el-table :data="flowHistoryTableData" border size="small" style="width:100%" row-key="id" :expand-row-keys="flowExpandedRows">
                <el-table-column type="expand">
                  <template #default="{ row }">
                    <div style="padding:8px 24px">
                      <el-table :data="row.histories" size="small" border>
                        <el-table-column label="时间" width="160" prop="created_at">
                          <template #default="{ row: h }">{{ formatTime(h.created_at) }}</template>
                        </el-table-column>
                        <el-table-column label="节点" min-width="120">
                          <template #default="{ row: h }">{{ nodeKeyToName(h.node_key) }}</template>
                        </el-table-column>
                        <el-table-column label="操作" width="110">
                          <template #default="{ row: h }"><el-tag size="small" :type="timelineItemType(h.action)">{{ actionLabel(h.action) }}</el-tag></template>
                        </el-table-column>
                        <el-table-column label="处理人" width="110" prop="operator">
                          <template #default="{ row: h }">{{ h.operator || '-' }}</template>
                        </el-table-column>
                        <el-table-column label="意见" min-width="140" prop="remark">
                          <template #default="{ row: h }">{{ h.remark || '-' }}</template>
                        </el-table-column>
                      </el-table>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column label="次数" width="70"><template #default="{ $index }">第 {{ $index + 1 }} 次</template></el-table-column>
                <el-table-column label="提交时间" width="160"><template #default="{ row }">{{ formatTime(row.instance.started_at) }}</template></el-table-column>
                <el-table-column label="提交人" width="110"><template #default="{ row }">{{ row.instance.started_by || '-' }}</template></el-table-column>
                <el-table-column label="状态" width="100">
                  <template #default="{ row }"><el-tag size="small" :type="flowInstanceTagType(row.instance.status)">{{ flowInstanceStatusLabel(row.instance.status) }}</el-tag></template>
                </el-table-column>
                <el-table-column label="操作步骤数" width="100"><template #default="{ row }">{{ (row.histories || []).length }} 步</template></el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </template>
      </el-skeleton>
      <template #footer>
        <el-button @click="flowVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onBeforeUnmount, onMounted, shallowRef, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import {
  getNotices, createNotice, updateNotice, deleteNotice, uploadAttachment,
  submitNotice, withdrawNotice, approveNotice, cancelApproveNotice
} from '../api/notice'
import { getDepartments } from '../api/department'
import { getOrchidWorkflowHistories } from '../api/orchid_workflow'
import { getEmployees } from '../api/employee'

const token = localStorage.getItem('token') || ''
const list = ref([])
const total = ref(0)
const deptList = ref([])
const dialogVisible = ref(false)
const previewVisible = ref(false)
const previewData = ref({ title: '', content: '', author: '', department: null })
const previewAttachments = ref([])
const query = ref({ title: '', status: null, approve_status: null, department_id: null, page: 1, page_size: 10 })
const attachments = ref([])

const flowVisible = ref(false)
const flowLoading = ref(false)
const flowData = ref({ instance: null, histories: [], tasks: [], instances: [] })
const flowTarget = ref({ id: null, name: '' })
const flowUserMap = ref({})
const flowNodeMap = ref({})

const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{}')
const currentUser = userInfo.real_name || userInfo.username || ''

const defaultForm = () => ({ title: '', author: currentUser, content: '', status: 1, attachments: '', department_id: null, approve_status: 'draft' })
const form = ref(defaultForm())

const editorRef = shallowRef()
const toolbarConfig = { excludeKeys: ['fullScreen'] }
const editorConfig = {
  placeholder: '请输入公告内容...',
  MENU_CONF: {
    uploadImage: {
      server: '/api/upload/image',
      fieldName: 'file',
      maxFileSize: 5 * 1024 * 1024,
      maxNumberOfFiles: 10,
      allowedFileTypes: ['image/*'],
      headers: { Authorization: 'Bearer ' + token }
    }
  }
}

const attachmentAccept = '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar'
const handleEditorCreated = (editor) => { editorRef.value = editor }
const handleDialogClosed = () => {
  if (editorRef.value) {
    editorRef.value.destroy()
    editorRef.value = null
  }
}
onBeforeUnmount(() => { if (editorRef.value) editorRef.value.destroy() })

const approveStatusLabel = (s) => ({ draft: '草稿', pending: '待审批', approved: '已通过', rejected: '已拒绝' }[s] || '草稿')
const approveTagType = (s) => ({ draft: 'info', pending: 'warning', approved: 'success', rejected: 'danger' }[s] || 'info')
const isEditable = (row) => ['draft', 'rejected'].includes(row?.approve_status)

const loadDepts = async () => {
  const res = await getDepartments({ page: 1, page_size: 999 })
  deptList.value = res.data.data.list || []
}

const loadData = async () => {
  const params = { ...query.value }
  if (!params.department_id) delete params.department_id
  if (!params.approve_status) delete params.approve_status
  const res = await getNotices(params)
  list.value = res.data.data.list || []
  total.value = res.data.data.total || 0
}

const handleSearch = () => { query.value.page = 1; loadData() }
const handleReset = () => {
  query.value = { title: '', status: null, approve_status: null, department_id: null, page: 1, page_size: 10 }
  loadData()
}

const openDialog = (row = null) => {
  if (row) {
    form.value = { ...row, department_id: row.department_id || null }
    attachments.value = row.attachments ? JSON.parse(row.attachments) : []
  } else {
    form.value = defaultForm()
    attachments.value = []
  }
  dialogVisible.value = true
}

const openPreview = (row) => {
  previewData.value = row
  previewAttachments.value = row.attachments ? JSON.parse(row.attachments) : []
  previewVisible.value = true
}

const handleAttachmentUpload = async ({ file }) => {
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await uploadAttachment(fd)
    if (res.data.code === 0) {
      attachments.value.push({ name: file.name, url: res.data.data.url })
      ElMessage.success('附件上传成功')
    } else {
      ElMessage.error(res.data.msg || '上传失败')
    }
  } catch {
    ElMessage.error('附件上传失败')
  }
}

const removeAttachment = (idx) => { attachments.value.splice(idx, 1) }

const handleSubmit = async () => {
  if (!form.value.title) { ElMessage.warning('请输入公告标题'); return }
  if (!form.value.content || form.value.content === '<p><br></p>') {
    ElMessage.warning('请输入公告内容'); return
  }
  const payload = {
    ...form.value,
    department_id: form.value.department_id || 0,
    attachments: attachments.value.length ? JSON.stringify(attachments.value) : ''
  }
  if (form.value.id) await updateNotice(form.value.id, payload)
  else await createNotice(payload)

  ElMessage.success('操作成功')
  dialogVisible.value = false
  loadData()
}

const handleSubmitAudit = async (row) => {
  await ElMessageBox.confirm('确认提交审核？提交后将进入审批流程且不可编辑。', '提交审核', { type: 'warning' })
  await submitNotice(row.id, {})
  ElMessage.success('已提交审核')
  loadData()
}

const handleWithdraw = async (row) => {
  await ElMessageBox.confirm('确认撤回审核？撤回后将恢复为草稿状态。', '撤回审核', { type: 'warning' })
  await withdrawNotice(row.id)
  ElMessage.success('已撤回')
  loadData()
}

const handleCancelApprove = async (row) => {
  await ElMessageBox.confirm('确认取消审核？操作后单据将恢复为草稿状态。', '取消审核', { type: 'warning' })
  await cancelApproveNotice(row.id)
  ElMessage.success('已取消审核，恢复为草稿')
  loadData()
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  await deleteNotice(id)
  ElMessage.success('删除成功')
  loadData()
}

const flowInstanceStatusLabel = (s) => ({ pending: '审批中', approved: '已通过', rejected: '已拒绝', withdrawn: '已撤回' }[s] || '未开始')
const flowInstanceTagType = (s) => ({ pending: 'warning', approved: 'success', rejected: 'danger', withdrawn: 'info' }[s] || 'info')
const timelineItemType = (action) => ({ approved: 'success', rejected: 'danger', pending: 'warning', withdraw: 'info' }[action] || 'primary')
const actionLabel = (action) => ({
  submit: '发起申请', pending: '进入待办', approved: '审批通过', approved_partial: '部分通过',
  rejected: '审批拒绝', transfer: '转交', skip: '跳过', withdraw: '撤回',
}[action] || action)

const flowHistoryTableData = computed(() => (flowData.value.instances || []).map((r, i) => ({ ...r, _idx: i })))
const flowExpandedRows = computed(() => flowHistoryTableData.value.map(r => r.instance?.id?.toString()))
const formatTime = (t) => t ? String(t).replace('T', ' ').slice(0, 19) : '-'
const nodeKeyToName = (key) => flowNodeMap.value[key] || key
const assigneeIdToName = (id) => flowUserMap.value[id] || `用户#${id}`

const timelineTitle = (h) => {
  const nodeName = nodeKeyToName(h.node_key)
  const map = {
    submit: '发起申请', pending: `进入节点：${nodeName}`, approved_partial: `部分通过：${nodeName}`,
    approved: `审批通过：${nodeName}`, rejected: `审批拒绝：${nodeName}`, transfer: `转交：${nodeName}`,
    skip: `跳过：${nodeName}`,
  }
  return map[h.action] || `${h.action || '处理'}：${nodeName}`
}

const openFlowDialog = async (row) => {
  flowTarget.value = { id: row.id, name: row.title || `#${row.id}` }
  flowVisible.value = true
  flowLoading.value = true
  flowUserMap.value = {}
  flowNodeMap.value = {}
  try {
    const res = await getOrchidWorkflowHistories({ biz_type: 'notice', biz_id: row.id })
    const data = res.data?.data || {}
    flowData.value = {
      instance: data.instance || null,
      instances: data.instances || [],
      histories: data.histories || [],
      tasks: data.tasks || [],
    }

    const empRes = await getEmployees({ page: 1, page_size: 1000 })
    const userMap = {}
    for (const emp of empRes.data?.data?.list || []) {
      if (emp.user_id) userMap[emp.user_id] = emp.name
    }
    flowUserMap.value = userMap

    if (data.instance?.definition_id) {
      try {
        const { getOrchidWorkflowDefinition } = await import('../api/orchid_workflow')
        const defRes = await getOrchidWorkflowDefinition(data.instance.definition_id)
        const dag = JSON.parse(defRes.data?.data?.dag_json || '{}')
        const nodeMap = {}
        for (const [key, node] of Object.entries(dag.nodes || {})) {
          nodeMap[key] = node?.config?._ui?.name || key
        }
        flowNodeMap.value = nodeMap
      } catch {}
    }
  } finally {
    flowLoading.value = false
  }
}

onMounted(() => { loadDepts(); loadData() })
</script>

<style scoped>
.notice-preview :deep(img) { max-width: 100%; height: auto; }
.notice-preview :deep(table) { border-collapse: collapse; width: 100%; }
.notice-preview :deep(td), .notice-preview :deep(th) { border: 1px solid #ddd; padding: 6px 10px; }
.notice-preview :deep(blockquote) { border-left: 4px solid #ddd; margin: 0; padding-left: 16px; color: #666; }
</style>
