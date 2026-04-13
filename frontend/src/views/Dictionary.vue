<template>
  <el-card shadow="never" style="border:none">
    <el-form :inline="true" style="margin-bottom:16px;display:flex;align-items:center;flex-wrap:wrap">
      <el-form-item label="关键字">
        <el-input v-model="query.keyword" placeholder="编码/名称/备注" clearable />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="listLoading" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
      <el-form-item style="margin-left:auto;margin-right:0">
        <el-button type="primary" @click="openMainDialog()">新增字典</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" stripe v-loading="listLoading">
      <el-table-column label="序号" width="80">
        <template #default="{ $index }">
          {{ (query.page - 1) * query.page_size + $index + 1 }}
        </template>
      </el-table-column>
      <el-table-column prop="code" label="编码" min-width="180" />
      <el-table-column prop="name" label="名称" min-width="180" />
      <el-table-column prop="remark" label="备注" min-width="220" />
      <el-table-column label="操作" width="220" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" link @click="openDetailDialog(row)">详情</el-button>
          <el-button size="small" type="primary" link @click="openMainDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" link @click="handleDeleteMain(row)">删除</el-button>
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
      @change="loadList"
    />
  </el-card>

  <el-dialog v-model="mainDialogVisible" :title="mainDialogMode === 'edit' ? '编辑字典主表' : '新增字典主表'" width="520px">
    <el-form :model="mainForm" label-width="110px">
      <el-form-item label="编码">
        <el-input v-model="mainForm.code" placeholder="请输入编码" />
      </el-form-item>
      <el-form-item label="名称">
        <el-input v-model="mainForm.name" placeholder="请输入名称" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="mainForm.remark" type="textarea" :rows="3" placeholder="请输入备注" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="mainDialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="mainSubmitting" @click="submitMainDialog">确定</el-button>
    </template>
  </el-dialog>

  <el-dialog
    v-model="detailDialogVisible"
    title="数据字典详情"
    width="92%"
    top="4vh"
    destroy-on-close
  >
    <el-card shadow="never" v-loading="detailLoading">
      <template #header>
        <span>明细信息</span>
      </template>

      <el-table
        ref="detailTableRef"
        :data="detailPagedItems"
        stripe
        @selection-change="handleDetailSelectionChange"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column label="ID" width="100">
          <template #default="{ row }">
            {{ row.id || '待保存' }}
          </template>
        </el-table-column>
        <el-table-column label="扩展字段" min-width="140">
          <template #default="{ row }">
            <el-input v-model="row.extfield" placeholder="请输入" />
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="180">
          <template #default="{ row }">
            <el-input v-model="row.remark" placeholder="请输入" />
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        style="margin-top:12px;justify-content:flex-end;display:flex"
        v-model:current-page="detailPage"
        :page-size="detailPageSize"
        :total="detailItems.length"
        layout="total, prev, pager, next"
        @current-change="handleDetailPageChange"
      />

      <div style="margin-top:12px;display:flex;justify-content:flex-start;gap:8px">
        <el-button type="primary" @click="handleAddDetailRow">新增</el-button>
        <el-button type="primary" plain @click="openImportDialog">导入</el-button>
        <el-button type="danger" :disabled="selectedDetailRows.length === 0" @click="handleBatchDeleteDetailRows">
          删除
        </el-button>
      </div>

    </el-card>

    <template #footer>
      <el-button @click="detailDialogVisible = false">关闭</el-button>
      <el-button type="primary" :loading="detailSaving" @click="saveDetailAll">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="importDialogVisible" title="导入明细" width="640px">
    <el-alert
      title="支持 xlsx/xls/csv/txt 文件，或直接粘贴文本；每行一条，格式：extfield,备注（备注可选）。导入后为新增数据，保存后系统生成ID。"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom:12px"
    />
    <el-upload
      :auto-upload="false"
      :limit="1"
      :file-list="importFileList"
      accept=".xlsx,.xls,.csv,.txt,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,text/plain,text/csv"
      :on-change="handleImportFileChange"
      :on-remove="handleImportFileRemove"
    >
      <el-button type="primary" plain>选择文件</el-button>
    </el-upload>
    <div style="margin:12px 0;font-size:12px;color:#909399">或直接粘贴内容：</div>
    <el-input
      v-model="importText"
      type="textarea"
      :rows="10"
      placeholder="示例：&#10;北京市,省会&#10;上海市&#10;浙江省,华东地区"
    />
    <template #footer>
      <el-button @click="closeImportDialog">取消</el-button>
      <el-button type="primary" @click="applyImport">导入到列表</el-button>
    </template>
  </el-dialog>

</template>

<script setup>
import { computed, ref, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as XLSX from 'xlsx'
import {
  getDataDictionaries,
  getDataDictionary,
  getDataDictionaryItems,
  createDataDictionary,
  updateDataDictionary,
  deleteDataDictionary,
  createDataDictionaryItem,
  updateDataDictionaryItem,
  deleteDataDictionaryItem
} from '../api/dictionary'

const query = ref({ keyword: '', page: 1, page_size: 10 })
const list = ref([])
const total = ref(0)
const listLoading = ref(false)

const mainDialogVisible = ref(false)
const mainDialogMode = ref('create')
const mainSubmitting = ref(false)
const mainForm = ref({ id: null, code: '', name: '', remark: '' })

const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const detailSaving = ref(false)
const detailForm = ref({ id: null })
const detailItems = ref([])
const originalDetailItems = ref([])
const detailTableRef = ref()
const selectedDetailRows = ref([])
const detailPage = ref(1)
const detailPageSize = 20
const importDialogVisible = ref(false)
const importText = ref('')
const importFileList = ref([])

const detailPagedItems = computed(() => {
  const start = (detailPage.value - 1) * detailPageSize
  return detailItems.value.slice(start, start + detailPageSize)
})

const clampDetailPage = () => {
  const maxPage = Math.max(1, Math.ceil(detailItems.value.length / detailPageSize))
  if (detailPage.value > maxPage) detailPage.value = maxPage
  if (detailPage.value < 1) detailPage.value = 1
}

watch(
  () => detailItems.value.length,
  () => {
    clampDetailPage()
  }
)

const loadList = async () => {
  listLoading.value = true
  try {
    const res = await getDataDictionaries({ ...query.value })
    list.value = res.data?.data?.list || []
    total.value = res.data?.data?.total || 0
  } finally {
    listLoading.value = false
  }
}

const handleSearch = () => {
  query.value.page = 1
  loadList()
}

const handleReset = () => {
  query.value = { keyword: '', page: 1, page_size: 10 }
  loadList()
}

const openMainDialog = (row = null) => {
  if (row) {
    mainDialogMode.value = 'edit'
    mainForm.value = {
      id: row.id,
      code: row.code || '',
      name: row.name || '',
      remark: row.remark || ''
    }
  } else {
    mainDialogMode.value = 'create'
    mainForm.value = { id: null, code: '', name: '', remark: '' }
  }
  mainDialogVisible.value = true
}

const submitMainDialog = async () => {
  const payload = {
    code: (mainForm.value.code || '').trim(),
    name: (mainForm.value.name || '').trim(),
    remark: mainForm.value.remark || ''
  }
  if (!payload.code || !payload.name) {
    ElMessage.warning('编码和名称不能为空')
    return
  }

  mainSubmitting.value = true
  try {
    if (mainDialogMode.value === 'edit' && mainForm.value.id) {
      await updateDataDictionary(mainForm.value.id, payload)
      ElMessage.success('主表更新成功')
    } else {
      await createDataDictionary(payload)
      ElMessage.success('主表创建成功')
    }
    mainDialogVisible.value = false
    await loadList()
  } finally {
    mainSubmitting.value = false
  }
}

const loadDetail = async (id) => {
  if (!id) return
  detailLoading.value = true
  try {
    const res = await getDataDictionary(id)
    const data = res.data?.data || {}
    let items = Array.isArray(data.items) ? data.items : []
    if (items.length === 0) {
      const listRes = await getDataDictionaryItems(id, { page: 1, page_size: 1000 })
      items = listRes.data?.data?.list || []
    }
    detailForm.value = { id: data.id || id }
    detailItems.value = (items || []).map((item) => ({
      id: item.id || null,
      extfield: resolveExtField(item),
      remark: item.remark || ''
    }))
    originalDetailItems.value = detailItems.value.map((item) => ({ ...item }))
    detailPage.value = 1
    selectedDetailRows.value = []
    detailTableRef.value?.clearSelection()
  } finally {
    detailLoading.value = false
  }
}

const openDetailDialog = async (row) => {
  detailDialogVisible.value = true
  await loadDetail(row.id)
}

const resolveExtField = (item) => {
  if (!item || typeof item !== 'object') return ''
  const direct =
    item.extfield ??
    item.extField ??
    item.EXTField ??
    item.ext_field ??
    item.ExtField ??
    item.ext1 ??
    item.Ext1 ??
    item.EXT1
  if (direct != null && direct !== '') return String(direct)

  const matchedKey = Object.keys(item).find((key) => {
    const normalized = String(key).replace(/[_-]/g, '').toLowerCase()
    return normalized === 'extfield' || normalized === 'ext1'
  })
  if (!matchedKey) return ''

  const value = item[matchedKey]
  return value == null ? '' : String(value)
}

const saveDetailAll = async () => {
  if (!detailForm.value.id) return

  const normalizeItemPayload = (item) => {
    const extfield = resolveExtField(item).trim()
    return {
      extfield,
      extField: extfield,
      EXTField: extfield,
      ext_field: extfield,
      ext1: extfield,
      remark: item.remark || ''
    }
  }
  const isChanged = (oldItem, currentPayload) => {
    const oldPayload = normalizeItemPayload(oldItem)
    return Object.keys(oldPayload).some((key) => oldPayload[key] !== currentPayload[key])
  }

  detailSaving.value = true
  try {
    const originalMap = new Map(
      originalDetailItems.value
        .filter((item) => Number.isInteger(Number(item.id)) && Number(item.id) > 0)
        .map((item) => [Number(item.id), item])
    )
    const currentIDs = new Set()
    const updateQueue = []
    const createQueue = []

    for (const item of detailItems.value) {
      const itemID = Number(item.id)
      const itemPayload = normalizeItemPayload(item)
      if (Number.isInteger(itemID) && itemID > 0) {
        currentIDs.add(itemID)
        const oldItem = originalMap.get(itemID)
        if (!oldItem || isChanged(oldItem, itemPayload)) {
          updateQueue.push({ id: itemID, payload: itemPayload })
        }
      } else {
        createQueue.push(itemPayload)
      }
    }

    const deleteQueue = []
    for (const [id] of originalMap.entries()) {
      if (!currentIDs.has(id)) {
        deleteQueue.push(id)
      }
    }

    for (const item of updateQueue) {
      await updateDataDictionaryItem(item.id, item.payload)
    }
    for (const item of createQueue) {
      await createDataDictionaryItem(detailForm.value.id, item)
    }
    for (const id of deleteQueue) {
      await deleteDataDictionaryItem(id)
    }

    ElMessage.success('保存成功')
    await Promise.all([loadDetail(detailForm.value.id), loadList()])
  } finally {
    detailSaving.value = false
  }
}

const handleDetailSelectionChange = (rows) => {
  selectedDetailRows.value = rows || []
}

const handleDetailPageChange = () => {
  selectedDetailRows.value = []
  detailTableRef.value?.clearSelection()
}

const handleAddDetailRow = () => {
  detailItems.value.push({
    id: null,
    extfield: '',
    remark: ''
  })
  detailPage.value = Math.max(1, Math.ceil(detailItems.value.length / detailPageSize))
}

const openImportDialog = () => {
  if (!detailForm.value.id) {
    ElMessage.warning('请先打开字典详情')
    return
  }
  importDialogVisible.value = true
}

const closeImportDialog = () => {
  importDialogVisible.value = false
  importText.value = ''
  importFileList.value = []
}

const readRawFileText = (raw) => {
  if (!raw) return Promise.resolve('')
  if (typeof raw.text === 'function') return raw.text()
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('read-failed'))
    reader.readAsText(raw)
  })
}

const readRawFileArrayBuffer = (raw) => {
  if (!raw) return Promise.resolve(new ArrayBuffer(0))
  if (typeof raw.arrayBuffer === 'function') return raw.arrayBuffer()
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = () => reject(new Error('read-array-buffer-failed'))
    reader.readAsArrayBuffer(raw)
  })
}

const isExcelFile = (fileLike) => {
  const name = String(fileLike?.name || fileLike?.raw?.name || '').toLowerCase()
  const type = String(fileLike?.type || fileLike?.raw?.type || '').toLowerCase()
  return /\.xlsx?$/.test(name) || type.includes('spreadsheetml') || type.includes('ms-excel')
}

const handleImportFileChange = async (file, fileList) => {
  importFileList.value = (fileList || []).slice(-1)
  const current = importFileList.value[0]
  if (!current?.raw) return
  if (isExcelFile(current)) {
    importText.value = ''
    return
  }
  try {
    importText.value = await readRawFileText(current.raw)
  } catch {
    ElMessage.error('读取文件失败，请重试')
  }
}

const handleImportFileRemove = () => {
  importFileList.value = []
}

const parseImportLines = (text) => {
  const lines = String(text || '')
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)

  const rows = []
  for (const line of lines) {
    const parts = line.split(/[,，\t]/).map((p) => p.trim())
    const extfield = parts[0] || ''
    if (!extfield) continue
    rows.push({
      id: null,
      extfield,
      remark: parts.length > 1 ? parts.slice(1).join(' ') : ''
    })
  }
  return rows
}

const parseExcelRows = async (raw) => {
  const buffer = await readRawFileArrayBuffer(raw)
  const workbook = XLSX.read(buffer, { type: 'array' })
  const sheetName = workbook.SheetNames?.[0]
  if (!sheetName) return []
  const worksheet = workbook.Sheets[sheetName]
  const table = XLSX.utils.sheet_to_json(worksheet, { header: 1, defval: '' })
  if (!Array.isArray(table) || table.length === 0) return []

  const lines = table
    .map((row) => (Array.isArray(row) ? row.map((cell) => String(cell ?? '').trim()) : []))
    .filter((row) => row.some(Boolean))
  if (lines.length === 0) return []

  const header = lines[0].map((v) => v.toLowerCase().replace(/[_\s-]/g, ''))
  const extIdx = header.findIndex((v) => v === 'extfield' || v === '扩展字段' || v === 'ext1')
  const remarkIdx = header.findIndex((v) => v === 'remark' || v === '备注')
  const hasHeader = extIdx >= 0 || remarkIdx >= 0
  const startIndex = hasHeader ? 1 : 0
  const useExtIdx = extIdx >= 0 ? extIdx : 0
  const useRemarkIdx = remarkIdx >= 0 ? remarkIdx : 1

  const rows = []
  for (let i = startIndex; i < lines.length; i += 1) {
    const row = lines[i]
    const extfield = (row[useExtIdx] || '').trim()
    if (!extfield) continue
    rows.push({
      id: null,
      extfield,
      remark: (row[useRemarkIdx] || '').trim()
    })
  }
  return rows
}

const applyImport = async () => {
  const current = importFileList.value[0]
  let rows = []
  if (current?.raw) {
    try {
      if (isExcelFile(current)) {
        rows = await parseExcelRows(current.raw)
      } else {
        const text = importText.value || (await readRawFileText(current.raw))
        rows = parseImportLines(text)
      }
    } catch {
      ElMessage.error('解析导入文件失败，请检查文件格式')
      return
    }
  } else {
    rows = parseImportLines(importText.value)
  }
  if (rows.length === 0) {
    ElMessage.warning('未识别到可导入的数据')
    return
  }
  detailItems.value.push(...rows)
  detailPage.value = Math.max(1, Math.ceil(detailItems.value.length / detailPageSize))
  ElMessage.success(`已导入 ${rows.length} 条，请点击保存`)
  closeImportDialog()
}

const handleBatchDeleteDetailRows = async () => {
  if (selectedDetailRows.value.length === 0) {
    ElMessage.warning('请先勾选要删除的明细')
    return
  }
  await ElMessageBox.confirm(`确认删除已勾选的 ${selectedDetailRows.value.length} 条明细？`, '提示', { type: 'warning' })
  const selectedSet = new Set(selectedDetailRows.value)
  detailItems.value = detailItems.value.filter((item) => !selectedSet.has(item))
  selectedDetailRows.value = []
  detailTableRef.value?.clearSelection()
  ElMessage.success('已删除勾选行，点击保存后生效')
}

const handleDeleteMain = async (row) => {
  await ElMessageBox.confirm(`确认删除字典「${row.name || row.code}」？`, '提示', { type: 'warning' })
  await deleteDataDictionary(row.id)
  ElMessage.success('删除成功')
  await loadList()
}

onMounted(loadList)
</script>
