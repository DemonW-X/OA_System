<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>公告管理</span>
        <el-button type="primary" @click="openDialog()">新增公告</el-button>
      </div>
    </template>

    <!-- 搜索栏 -->
    <el-form :inline="true" style="margin-bottom:16px">
      <el-form-item label="标题">
        <el-input v-model="query.title" placeholder="请输入标题" clearable />
      </el-form-item>
      <el-form-item label="部门">
        <el-select v-model="query.department_id" placeholder="全部部门" clearable style="width:130px">
          <el-option
            v-for="d in deptList"
            :key="d.id"
            :label="d.name"
            :value="d.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable style="width:100px">
          <el-option label="已发布" :value="1" />
          <el-option label="草稿" :value="0" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <!-- 列表 -->
    <el-table :data="list" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" min-width="160" />
      <el-table-column label="部门" width="130">
        <template #default="{ row }">
          {{ row.department ? row.department.name : '全部部门' }}
        </template>
      </el-table-column>
      <el-table-column prop="author" label="作者" width="110" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '已发布' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="openPreview(row)">预览</el-button>
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

    <!-- 编辑/新增对话框 -->
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
            <el-option
              v-for="d in deptList"
              :key="d.id"
              :label="d.name"
              :value="d.id"
            />
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
            <Toolbar
              :editor="editorRef"
              :defaultConfig="toolbarConfig"
              style="border-bottom:1px solid #ccc"
            />
            <Editor
              v-model="form.content"
              :defaultConfig="editorConfig"
              style="height:360px;overflow-y:hidden"
              @onCreated="handleEditorCreated"
            />
          </div>
        </el-form-item>
        <el-form-item label="附件">
          <div style="width:100%">
            <el-upload
              :http-request="handleAttachmentUpload"
              :show-file-list="false"
              :accept="attachmentAccept"
              multiple
            >
              <el-button size="small">上传附件</el-button>
              <template #tip>
                <div style="color:#999;font-size:12px;margin-top:4px">
                  支持 PDF、Word、Excel、PPT、TXT、ZIP、RAR，单文件不超过 20MB
                </div>
              </template>
            </el-upload>
            <div v-if="attachments.length" style="margin-top:8px">
              <el-tag
                v-for="(att, idx) in attachments"
                :key="idx"
                closable
                style="margin:4px"
                @close="removeAttachment(idx)"
              >
                <a :href="att.url" target="_blank" style="text-decoration:none;color:inherit">
                  {{ att.name }}
                </a>
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

    <!-- 预览对话框 -->
    <el-dialog v-model="previewVisible" :title="previewData.title" width="800px">
      <div style="color:#999;font-size:13px;margin-bottom:12px">
        作者：{{ previewData.author }}
        <span v-if="previewData.department" style="margin-left:16px">
          部门：{{ previewData.department.name }}
        </span>
      </div>
      <div class="notice-preview" v-html="previewData.content" />
      <div v-if="previewAttachments.length" style="margin-top:16px;border-top:1px solid #eee;padding-top:12px">
        <div style="font-weight:600;margin-bottom:8px">附件：</div>
        <div v-for="(att, idx) in previewAttachments" :key="idx" style="margin-bottom:4px">
          <a :href="att.url" target="_blank">{{ att.name }}</a>
        </div>
      </div>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onBeforeUnmount, onMounted, shallowRef } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import { getNotices, createNotice, updateNotice, deleteNotice, uploadAttachment } from '../api/notice'
import { getDepartments } from '../api/department'

const token = localStorage.getItem('token') || ''
const list = ref([])
const total = ref(0)
const deptList = ref([])
const dialogVisible = ref(false)
const previewVisible = ref(false)
const previewData = ref({ title: '', content: '', author: '', department: null })
const previewAttachments = ref([])
const query = ref({ title: '', status: null, department_id: null, page: 1, page_size: 10 })
const attachments = ref([])

const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{}')
const currentUser = userInfo.real_name || userInfo.username || ''

const defaultForm = () => ({ title: '', author: currentUser, content: '', status: 1, attachments: '', department_id: null })
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

const loadDepts = async () => {
  const res = await getDepartments({ page: 1, page_size: 999 })
  deptList.value = res.data.data.list || []
}

const loadData = async () => {
  const params = { ...query.value }
  if (!params.department_id) delete params.department_id
  const res = await getNotices(params)
  list.value = res.data.data.list || []
  total.value = res.data.data.total || 0
}

const handleSearch = () => { query.value.page = 1; loadData() }
const handleReset = () => {
  query.value = { title: '', status: null, department_id: null, page: 1, page_size: 10 }
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
  if (form.value.id) {
    await updateNotice(form.value.id, payload)
  } else {
    await createNotice(payload)
  }
  ElMessage.success('操作成功')
  dialogVisible.value = false
  loadData()
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  await deleteNotice(id)
  ElMessage.success('删除成功')
  loadData()
}

onMounted(() => { loadDepts(); loadData() })
</script>

<style scoped>
.notice-preview :deep(img) { max-width: 100%; height: auto; }
.notice-preview :deep(table) { border-collapse: collapse; width: 100%; }
.notice-preview :deep(td), .notice-preview :deep(th) { border: 1px solid #ddd; padding: 6px 10px; }
.notice-preview :deep(blockquote) { border-left: 4px solid #ddd; margin: 0; padding-left: 16px; color: #666; }
</style>
