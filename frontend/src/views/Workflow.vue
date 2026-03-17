<template>
  <el-card>
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center">
        <span>流程管理（可视化节点拖拽）</span>
        <div style="display:flex;gap:8px">
          <el-button type="success" @click="openGenerateDialog">按职位生成</el-button>
          <el-button type="primary" @click="openDialog()">新增流程</el-button>
        </div>
      </div>
    </template>

    <el-form :inline="true" style="margin-bottom:16px">
      <el-form-item label="流程名称">
        <el-input v-model="query.name" placeholder="请输入流程名称" clearable />
      </el-form-item>
      <el-form-item label="适用业务">
        <el-select v-model="query.biz_type" placeholder="全部" clearable style="width:180px">
          <el-option v-for="b in bizTypes" :key="b.code" :label="b.name" :value="b.code" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="loadData">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="filteredList" stripe>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="流程名称" min-width="180" />
      <el-table-column label="适用业务" width="180">
        <template #default="{ row }">{{ getBizLabel(row.biz_type) }}</template>
      </el-table-column>
      <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
      <el-table-column label="启用" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_active ? 'success' : 'info'">{{ row.is_active ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button size="small" @click="openDialog(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑流程' : '新增流程'" width="1300px" top="4vh">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="流程名称" prop="name">
              <el-input v-model="form.name" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="适用业务" prop="biz_type">
              <el-select v-model="form.biz_type" style="width:100%" placeholder="请选择业务">
                <el-option v-for="b in bizTypes" :key="b.code" :label="b.name" :value="b.code" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="是否启用">
              <el-switch v-model="form.is_active" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="描述">
          <el-input v-model="form.description" />
        </el-form-item>

        <el-divider content-position="left">可视化设计器</el-divider>
        <div class="designer-wrap">
          <div class="toolbox">
            <div class="toolbox-title">节点</div>
            <el-button class="tool-btn" @click="addNode('approve')">+ 审批</el-button>
            <el-button class="tool-btn" @click="addNode('copy')">+ 抄送</el-button>

            <el-divider />
            <el-button type="primary" class="tool-btn" @click="toggleConnectMode">
              {{ connectMode ? '退出连线模式' : '进入连线模式' }}
            </el-button>
            <el-button class="tool-btn" @click="clearSelection">清空选中</el-button>

            <div class="tool-tip">
              <div>· 拖拽节点调整位置</div>
              <div>· 连线模式：先点源节点，再点目标节点</div>
              <div>· 当前选中节点可在右侧编辑</div>
            </div>
          </div>

          <div class="canvas-wrap">
            <div class="canvas" ref="canvasRef" @click="onCanvasClick">
              <svg class="edge-layer" xmlns="http://www.w3.org/2000/svg" width="100%" height="100%">
                <defs>
                  <marker id="arrow" markerWidth="14" markerHeight="10" refX="12" refY="5" orient="auto" markerUnits="userSpaceOnUse">
                    <path d="M0,0 L14,5 L0,10 z" fill="#409eff" />
                  </marker>
                </defs>
                <g v-for="(e, idx) in form.edges" :key="edgeKey(e, idx)" v-show="hasNode(e.from) && hasNode(e.to)">
                  <path
                    :d="edgeGeometry(e).d"
                    fill="none"
                    stroke="#409eff"
                    stroke-width="2.5"
                    marker-end="url(#arrow)"
                    @click.stop="selectEdge(idx)"
                  />
                  <text
                    :x="edgeGeometry(e).lx"
                    :y="edgeGeometry(e).ly - 8"
                    fill="#333"
                    font-size="12"
                  >
                    {{ e.condition || `${e.from} → ${e.to}` }}
                  </text>
                </g>
              </svg>

              <div
                v-for="(n, idx) in form.nodes"
                :key="n.key || idx"
                class="wf-node"
                :class="[
                  `type-${n.type || 'approve'}`,
                  { selected: selectedNodeIndex === idx, source: pendingSourceKey === n.key }
                ]"
                :style="{ left: `${n.x}px`, top: `${n.y}px` }"
                @mousedown.stop="onNodeMouseDown(idx, $event)"
                @click.stop="onNodeClick(idx)"
              >
                <div class="node-title">{{ n.name || n.key || '未命名节点' }}</div>
                <div class="node-sub">{{ nodeTypeLabel(n.type) }} · {{ n.key || '-' }}</div>
              </div>
            </div>
          </div>

          <div class="property-panel">
            <div class="toolbox-title">节点属性</div>
            <template v-if="selectedNode">
              <el-form label-width="86px" size="small">
                <el-form-item label="节点Key">
                  <el-input v-model="selectedNode.key" placeholder="例如 manager" />
                </el-form-item>
                <el-form-item label="节点名称">
                  <el-input v-model="selectedNode.name" placeholder="例如 直属领导审批" />
                </el-form-item>
                <el-form-item label="节点类型">
                  <el-input
                    v-if="selectedNode?.type === 'start' || selectedNode?.type === 'end'"
                    :model-value="nodeTypeLabel(selectedNode?.type)"
                    disabled
                  />
                  <el-select v-else v-model="selectedNode.type" style="width:100%">
                    <el-option label="审批" value="approve" />
                    <el-option label="抄送" value="copy" />
                  </el-select>
                </el-form-item>
                <el-form-item label="审批方式" v-if="selectedNode?.type === 'approve'">
                  <el-select v-model="selectedNode.approve_type" style="width:100%">
                    <el-option label="或签（任一人通过）" value="or" />
                    <el-option label="会签（全部通过）" value="and" />
                  </el-select>
                </el-form-item>

                <el-form-item label="审批人" v-if="selectedNode?.type === 'approve'">
                  <el-select
                    v-model="selectedNode.assignee_mode"
                    style="width:100%; margin-bottom: 8px"
                  >
                    <el-option label="按人员" value="user" />
                    <el-option label="按职位" value="position" />
                  </el-select>

                  <el-select
                    v-if="selectedNode.assignee_mode === 'user'"
                    v-model="selectedNode.approver_user_ids"
                    multiple
                    filterable
                    remote
                    :remote-method="onEmployeeRemoteSearch"
                    :loading="employeeLoading"
                    collapse-tags
                    collapse-tags-tooltip
                    style="width:100%"
                    placeholder="请输入关键字搜索审批人"
                  >
                    <el-option v-for="u in employeeOptions" :key="u.value" :label="u.label" :value="u.value" />
                  </el-select>

                  <el-select
                    v-else
                    v-model="selectedNode.approver_position_ids_arr"
                    multiple
                    filterable
                    remote
                    :remote-method="onPositionRemoteSearch"
                    :loading="positionLoading"
                    collapse-tags
                    collapse-tags-tooltip
                    style="width:100%"
                    placeholder="请输入关键字搜索审批职位"
                  >
                    <el-option v-for="p in positionOptions" :key="p.value" :label="p.label" :value="p.value" />
                  </el-select>
                </el-form-item>

                <el-form-item label="抄送人" v-if="selectedNode?.type === 'copy'">
                  <el-select
                    v-model="selectedNode.copy_user_ids"
                    multiple
                    filterable
                    remote
                    :remote-method="onEmployeeRemoteSearch"
                    :loading="employeeLoading"
                    collapse-tags
                    collapse-tags-tooltip
                    style="width:100%"
                    placeholder="请输入关键字搜索抄送人"
                  >
                    <el-option v-for="u in employeeOptions" :key="u.value" :label="u.label" :value="u.value" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button
                    type="danger"
                    plain
                    :disabled="selectedNode?.type === 'start' || selectedNode?.type === 'end'"
                    @click="removeSelectedNode"
                  >
                    删除节点
                  </el-button>
                </el-form-item>
              </el-form>
            </template>
            <el-empty v-else description="请在画布中选中节点" :image-size="60" />

            <el-divider />
            <div class="toolbox-title">连线属性</div>
            <template v-if="selectedEdgeIndex >= 0 && form.edges[selectedEdgeIndex]">
              <el-form label-width="64px" size="small">
                <el-form-item label="来源">
                  <el-input :model-value="form.edges[selectedEdgeIndex].from" disabled />
                </el-form-item>
                <el-form-item label="目标">
                  <el-input :model-value="form.edges[selectedEdgeIndex].to" disabled />
                </el-form-item>
                <el-form-item label="条件">
                  <el-input v-model="form.edges[selectedEdgeIndex].condition" placeholder="例如：department_id=7 && position_name~经理" />
                </el-form-item>
                <el-form-item>
                  <el-button type="danger" plain @click="removeSelectedEdge">删除连线</el-button>
                  <el-button plain @click="showConditionSyntaxHelp">语法提示</el-button>
                </el-form-item>
              </el-form>
            </template>
          </div>
        </div>

        <el-divider content-position="left">DAG JSON 预览</el-divider>
        <el-input :model-value="dagPreview" type="textarea" :rows="10" readonly />
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="generateDialogVisible" title="按职位生成审批流程" width="500px">
      <el-form label-width="100px">
        <el-form-item label="流程名称">
          <el-input v-model="generateForm.name" placeholder="请输入流程名称" />
        </el-form-item>
        <el-form-item label="适用业务">
          <el-select v-model="generateForm.biz_type" placeholder="请选择业务" style="width:100%">
            <el-option v-for="b in bizTypes" :key="b.code" :label="b.name" :value="b.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择部门">
          <el-select v-model="generateForm.department_id" placeholder="请选择部门" style="width:100%">
            <el-option v-for="d in departments" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="generateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleGenerate">生成并编辑</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getBizTypes, generatePositionWorkflow } from '../api/workflow'
import { getEmployees } from '../api/employee'
import { getPositions } from '../api/position'
import { getDepartments } from '../api/department'
import { getOrchidWorkflows, createOrchidWorkflow, updateOrchidWorkflow, deleteOrchidWorkflow } from '../api/orchid_workflow'

const NODE_W = 170
const NODE_H = 64

const list = ref([])
const bizTypes = ref([])
const departments = ref([])
const employeeOptions = ref([])
const employeeLoading = ref(false)
const positionOptions = ref([])
const positionLoading = ref(false)
const dialogVisible = ref(false)
const generateDialogVisible = ref(false)
const generateForm = ref({ name: '', biz_type: '', department_id: null })
const formRef = ref()
const canvasRef = ref()

const query = ref({ name: '', biz_type: '' })
const connectMode = ref(false)
const pendingSourceKey = ref('')
const selectedNodeIndex = ref(-1)
const selectedEdgeIndex = ref(-1)

const dragging = ref({ active: false, index: -1, offsetX: 0, offsetY: 0 })

const rules = {
  name: [{ required: true, message: '请输入流程名称', trigger: 'blur' }],
  biz_type: [{ required: true, message: '请选择适用业务', trigger: 'change' }],
}

const defaultForm = () => ({
  id: null,
  name: '',
  biz_type: '',
  description: '',
  is_active: true,
  nodes: [],
  edges: []
})

const form = ref(defaultForm())

const selectedNode = computed(() => {
  if (selectedNodeIndex.value < 0) return null
  return form.value.nodes[selectedNodeIndex.value] || null
})

const filteredList = computed(() => {
  return list.value.filter(x => {
    if (query.value.name && !x.name?.includes(query.value.name)) return false
    if (query.value.biz_type && x.biz_type !== query.value.biz_type) return false
    return true
  })
})

const dagPreview = computed(() => {
  const nodes = {}
  form.value.nodes.forEach((n, idx) => {
    if (!n.key) return

    const config = {
      approve_type: n.approve_type || 'or',
      approvers: n.assignee_mode === 'position' ? [] : (n.approver_user_ids || []),
      approver_position_ids: n.assignee_mode === 'position' ? (n.approver_position_ids_arr || []) : [],
      copy_user_ids: n.type === 'copy' ? (n.copy_user_ids || []) : [],
      _ui: {
        x: Number(n.x || 0),
        y: Number(n.y || 0),
        type: n.type || 'approve',
        name: n.name || ''
      }
    }

    nodes[n.key] = {
      id: idx + 1,
      activity: n.key,
      config,
    }
  })
  const edges = (form.value.edges || []).filter(e => e.from && e.to).map(e => {
    const obj = { from: e.from, to: e.to }
    if (e.condition) obj.condition = e.condition
    return obj
  })
  return JSON.stringify({ name: form.value.name, nodes, edges }, null, 2)
})

const getBizLabel = (v) => bizTypes.value.find(b => b.code === v)?.name || v || '-'
const nodeTypeLabel = (t) => ({ start: '开始', approve: '审批', copy: '抄送', end: '结束' }[t] || '审批')
const edgeKey = (e, idx) => `${e.from}-${e.to}-${idx}`
const hasNode = (key) => !!form.value.nodes.find(n => n.key === key)

const clearSelection = () => {
  selectedNodeIndex.value = -1
  selectedEdgeIndex.value = -1
  pendingSourceKey.value = ''
}

const addNode = (type = 'approve') => {
  if (type !== 'approve' && type !== 'copy') {
    type = 'approve'
  }

  const businessNodes = form.value.nodes.filter(n => n.type !== 'start' && n.type !== 'end')
  const index = businessNodes.length + 1
  const key = `${type}_${index}`
  const endNode = form.value.nodes.find(n => n.type === 'end')
  const baseY = endNode ? Math.max(40, Number(endNode.y) - 120) : 140

  form.value.nodes.push({
    key,
    name: type === 'copy' ? `抄送${index}` : `审批${index}`,
    type,
    approve_type: 'or',
    assignee_mode: 'user',
    approver_user_ids: [],
    approver_position_ids_arr: [],
    copy_user_ids: [],
    x: 120 + ((index - 1) % 4) * 190,
    y: baseY + Math.floor((index - 1) / 4) * 110,
  })
}

const findNode = (key) => form.value.nodes.find(n => n.key === key)
const nodeCenterX = (n) => (n ? Number(n.x) + NODE_W / 2 : 0)
const nodeCenterY = (n) => (n ? Number(n.y) + NODE_H / 2 : 0)

const edgePoints = (e) => {
  const from = findNode(e.from)
  const to = findNode(e.to)
  if (!from || !to) return { x1: 0, y1: 0, x2: 0, y2: 0, mx: 0, my: 0 }

  const fromCenterX = nodeCenterX(from)
  const fromCenterY = nodeCenterY(from)
  const toCenterX = nodeCenterX(to)
  const toCenterY = nodeCenterY(to)

  const dx = toCenterX - fromCenterX
  const dy = toCenterY - fromCenterY

  // 横向为主：从左右边连线；纵向为主：从上下边连线
  let x1 = fromCenterX
  let y1 = fromCenterY
  let x2 = toCenterX
  let y2 = toCenterY

  if (Math.abs(dx) >= Math.abs(dy)) {
    x1 = fromCenterX + (dx >= 0 ? NODE_W / 2 : -NODE_W / 2)
    y1 = fromCenterY
    x2 = toCenterX + (dx >= 0 ? -NODE_W / 2 : NODE_W / 2)
    y2 = toCenterY
  } else {
    x1 = fromCenterX
    y1 = fromCenterY + (dy >= 0 ? NODE_H / 2 : -NODE_H / 2)
    x2 = toCenterX
    y2 = toCenterY + (dy >= 0 ? -NODE_H / 2 : NODE_H / 2)
  }

  return {
    x1,
    y1,
    x2,
    y2,
    mx: (x1 + x2) / 2,
    my: (y1 + y2) / 2,
    dx,
    dy,
  }
}

const edgeGeometry = (e) => {
  const p = edgePoints(e)
  const reverseExists = form.value.edges.some(ed => ed !== e && ed.from === e.to && ed.to === e.from)

  if (!reverseExists) {
    return {
      d: `M ${p.x1} ${p.y1} L ${p.x2} ${p.y2}`,
      lx: p.mx,
      ly: p.my,
    }
  }

  // 互指边：固定左右出入点 + 使用“规范方向法向量”分居两侧，避免重合
  const fromNode = findNode(e.from)
  const toNode = findNode(e.to)
  if (!fromNode || !toNode) {
    return {
      d: `M ${p.x1} ${p.y1} L ${p.x2} ${p.y2}`,
      lx: p.mx,
      ly: p.my,
    }
  }

  const fx = nodeCenterX(fromNode)
  const fy = nodeCenterY(fromNode)
  const tx = nodeCenterX(toNode)
  const ty = nodeCenterY(toNode)

  // 当前边的出入点：固定左右锚点
  let x1 = fx + (tx >= fx ? NODE_W / 2 : -NODE_W / 2)
  let y1 = fy
  let x2 = tx + (tx >= fx ? -NODE_W / 2 : NODE_W / 2)
  let y2 = ty

  // 规范方向（按 key 排序）用于统一法向量，防止反向边计算后抵消成同一条线
  const aKey = e.from < e.to ? e.from : e.to
  const bKey = e.from < e.to ? e.to : e.from
  const aNode = findNode(aKey)
  const bNode = findNode(bKey)
  if (!aNode || !bNode) {
    return {
      d: `M ${x1} ${y1} L ${x2} ${y2}`,
      lx: (x1 + x2) / 2,
      ly: (y1 + y2) / 2,
    }
  }

  const avx = nodeCenterX(bNode) - nodeCenterX(aNode)
  const avy = nodeCenterY(bNode) - nodeCenterY(aNode)
  const alen = Math.hypot(avx, avy) || 1
  const nx = -avy / alen
  const ny = avx / alen

  // a->b 在一侧，b->a 在另一侧
  const side = e.from === aKey ? 1 : -1
  const offset = 24

  x1 += nx * offset * side
  y1 += ny * offset * side
  x2 += nx * offset * side
  y2 += ny * offset * side

  return {
    d: `M ${x1} ${y1} L ${x2} ${y2}`,
    lx: (x1 + x2) / 2,
    ly: (y1 + y2) / 2,
  }
}

const onNodeClick = (idx) => {
  const target = form.value.nodes[idx]
  if (!target?.key) return

  // 非连线模式：普通选中节点
  if (!connectMode.value) {
    selectedNodeIndex.value = idx
    selectedEdgeIndex.value = -1
    return
  }

  // 连线模式：第一次点击选中源节点
  if (!pendingSourceKey.value) {
    pendingSourceKey.value = target.key
    selectedNodeIndex.value = idx
    selectedEdgeIndex.value = -1
    return
  }

  // 连线模式：第二次点击作为目标节点，创建连线后清空选中
  const from = pendingSourceKey.value
  const to = target.key

  if (from === to) {
    ElMessage.warning('不能连接到自身')
    clearSelection()
    return
  }

  const exists = form.value.edges.some(e => e.from === from && e.to === to)
  if (exists) {
    ElMessage.warning('该连线已存在')
    clearSelection()
    return
  }

  form.value.edges.push({ from, to, condition: '' })
  clearSelection()
}

const onCanvasClick = () => {
  selectedNodeIndex.value = -1
}

const onNodeMouseDown = (idx, evt) => {
  if (!canvasRef.value) return
  const rect = canvasRef.value.getBoundingClientRect()
  const node = form.value.nodes[idx]
  dragging.value = {
    active: true,
    index: idx,
    offsetX: evt.clientX - rect.left - Number(node.x || 0),
    offsetY: evt.clientY - rect.top - Number(node.y || 0),
  }
}

const onMouseMove = (evt) => {
  if (!dragging.value.active || !canvasRef.value) return
  const rect = canvasRef.value.getBoundingClientRect()
  const idx = dragging.value.index
  const node = form.value.nodes[idx]
  if (!node) return

  const maxX = rect.width - NODE_W
  const maxY = rect.height - NODE_H
  const x = evt.clientX - rect.left - dragging.value.offsetX
  const y = evt.clientY - rect.top - dragging.value.offsetY

  node.x = Math.max(0, Math.min(maxX, x))
  node.y = Math.max(0, Math.min(maxY, y))
}

const onMouseUp = () => {
  dragging.value.active = false
}

const selectEdge = (idx) => {
  selectedEdgeIndex.value = idx
  selectedNodeIndex.value = -1
}

const removeSelectedEdge = () => {
  if (selectedEdgeIndex.value < 0) return
  form.value.edges.splice(selectedEdgeIndex.value, 1)
  selectedEdgeIndex.value = -1
}

const removeSelectedNode = () => {
  if (selectedNodeIndex.value < 0) return
  const target = form.value.nodes[selectedNodeIndex.value]
  if (target?.type === 'start' || target?.type === 'end') {
    ElMessage.warning('开始节点和结束节点不允许删除')
    return
  }
  form.value.nodes.splice(selectedNodeIndex.value, 1)
  form.value.edges = form.value.edges.filter(e => e.from !== target.key && e.to !== target.key)
  clearSelection()
}

const toggleConnectMode = () => {
  connectMode.value = !connectMode.value
  pendingSourceKey.value = ''
}

const makeDefaultNodes = () => ([
  {
    key: 'start',
    name: '开始',
    type: 'start',
    approve_type: 'or',
    assignee_mode: 'user',
    approver_user_ids: [],
    approver_position_ids_arr: [],
    copy_user_ids: [],
    x: 140,
    y: 220,
  },
  {
    key: 'end',
    name: '结束',
    type: 'end',
    approve_type: 'or',
    assignee_mode: 'user',
    approver_user_ids: [],
    approver_position_ids_arr: [],
    copy_user_ids: [],
    x: 500,
    y: 220,
  }
])

const parseDagToEditor = (dagJSON) => {
  try {
    const dag = JSON.parse(dagJSON || '{}')
    const nodes = Object.entries(dag.nodes || {}).map(([key, node], idx) => {
      const rawType = node?.config?._ui?.type || 'approve'
      const safeType = ['start', 'end', 'approve', 'copy'].includes(rawType) ? rawType : 'approve'
      const userIDs = (node?.config?.approvers || []).map(v => Number(v)).filter(v => !Number.isNaN(v) && v > 0)
      const positionIDs = (node?.config?.approver_position_ids || []).map(v => Number(v)).filter(v => !Number.isNaN(v) && v > 0)
      const copyIDs = (node?.config?.copy_user_ids || []).map(v => Number(v)).filter(v => !Number.isNaN(v) && v > 0)

      return {
        key,
        name: node?.config?._ui?.name || key,
        type: safeType,
        approve_type: node?.config?.approve_type || 'or',
        assignee_mode: positionIDs.length ? 'position' : 'user',
        approver_user_ids: userIDs,
        approver_position_ids_arr: positionIDs,
        copy_user_ids: copyIDs,
        x: Number(node?.config?._ui?.x ?? (80 + (idx % 4) * 200)),
        y: Number(node?.config?._ui?.y ?? (50 + Math.floor(idx / 4) * 110)),
      }
    })
    const edges = (dag.edges || []).map(e => ({ from: e.from || '', to: e.to || '', condition: e.condition || e.label || '' }))
    return { nodes, edges }
  } catch {
    return { nodes: [], edges: [] }
  }
}

const validateBeforeSubmit = () => {
  if (!form.value.nodes.length) return '请至少添加一个节点'

  const keys = form.value.nodes.map(n => (n.key || '').trim()).filter(Boolean)
  if (keys.length !== form.value.nodes.length) return '所有节点必须填写节点Key'
  if (new Set(keys).size !== keys.length) return '节点Key不能重复'

  for (const n of form.value.nodes) {
    if (!['start', 'end', 'approve', 'copy'].includes(n.type)) {
      return '节点类型仅支持：审批、抄送（系统保留开始/结束）'
    }
  }

  for (const n of form.value.nodes) {
    if (n.type === 'approve') {
      if (!['or', 'and'].includes(n.approve_type || 'or')) {
        return `节点【${n.name || n.key}】审批方式不合法`
      }
      if (n.assignee_mode === 'position') {
        if (!(n.approver_position_ids_arr || []).length) return `节点【${n.name || n.key}】请选择审批职位`
      } else {
        if (!(n.approver_user_ids || []).length) return `节点【${n.name || n.key}】请选择审批人`
      }
    }
    if (n.type === 'copy') {
      if (!(n.copy_user_ids || []).length) return `节点【${n.name || n.key}】请选择抄送人`
    }
  }

  for (const e of form.value.edges) {
    if (!keys.includes(e.from) || !keys.includes(e.to)) return '连线存在无效节点，请检查'
    if (e.from === e.to) return '不允许节点自环连线'
  }

  const startCount = form.value.nodes.filter(n => n.type === 'start').length
  const endCount = form.value.nodes.filter(n => n.type === 'end').length
  if (startCount !== 1) return '必须且仅能有1个开始节点'
  if (endCount !== 1) return '必须且仅能有1个结束节点'

  return ''
}

const loadData = async () => {
  const res = await getOrchidWorkflows({})
  list.value = res.data.data || []
}

const loadBizTypes = async () => {
  const res = await getBizTypes()
  bizTypes.value = res.data.data || []
}

const loadDepartments = async () => {
  const res = await getDepartments({ page: 1, page_size: 1000 })
  departments.value = res.data.data.list || []
}

const loadEmployeeOptions = async (keyword = '') => {
  employeeLoading.value = true
  try {
    const params = { page: 1, page_size: 100, status: 1 }
    if ((keyword || '').trim()) params.keyword = keyword.trim()
    const res = await getEmployees(params)
    const rows = res.data?.data?.list || []
    // 审批按“人员”时，后端实际使用的是 user_id（登录账号ID）
    employeeOptions.value = rows
      .filter(e => Number(e.user_id) > 0)
      .map(e => ({
        value: Number(e.user_id),
        label: `${e.name}${e.position_info?.name ? `（${e.position_info.name}` : ''}${e.department?.name ? ` / ${e.department.name}` : ''}${e.position_info?.name ? '）' : ''}`,
      }))
  } finally {
    employeeLoading.value = false
  }
}

const onEmployeeRemoteSearch = (keyword) => {
  loadEmployeeOptions(keyword)
}

const loadPositionOptions = async (keyword = '') => {
  positionLoading.value = true
  try {
    const params = { page: 1, page_size: 100 }
    if ((keyword || '').trim()) params.keyword = keyword.trim()
    const res = await getPositions(params)
    const rows = res.data?.data?.list || []
    positionOptions.value = rows.map(p => ({
      value: p.id,
      label: `${p.name}${p.department?.name ? `（${p.department.name}）` : ''}`,
    }))
  } finally {
    positionLoading.value = false
  }
}

const onPositionRemoteSearch = (keyword) => {
  loadPositionOptions(keyword)
}

const handleReset = () => {
  query.value = { name: '', biz_type: '' }
}

const openDialog = async (row = null) => {
  // 每次打开都重新拉取，保证新建的员工/职位可立即在下拉里看到
  await Promise.all([loadEmployeeOptions(), loadPositionOptions()])

  if (!row) {
    form.value = defaultForm()
    form.value.nodes = makeDefaultNodes()
    form.value.edges = [{ from: 'start', to: 'end', condition: '' }]
  } else {
    const parsed = parseDagToEditor(row.dag_json)
    form.value = {
      id: row.id,
      name: row.name,
      biz_type: row.biz_type,
      description: row.description,
      is_active: !!row.is_active,
      nodes: parsed.nodes,
      edges: parsed.edges
    }
  }
  clearSelection()
  connectMode.value = false
  dialogVisible.value = true
}

const openGenerateDialog = () => {
  generateForm.value = { name: '', biz_type: '', department_id: null }
  generateDialogVisible.value = true
}

const handleGenerate = async () => {
  if (!generateForm.value.name || !generateForm.value.biz_type || !generateForm.value.department_id) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    const res = await generatePositionWorkflow(generateForm.value)
    const data = res.data.data
    
    generateDialogVisible.value = false
    
    await Promise.all([loadEmployeeOptions(), loadPositionOptions()])
    
    const parsed = parseDagToEditor(data.dag_json)
    form.value = {
      name: data.name,
      biz_type: data.biz_type,
      description: data.description,
      is_active: true,
      nodes: parsed.nodes,
      edges: parsed.edges
    }
    clearSelection()
    connectMode.value = false
    dialogVisible.value = true
    
    ElMessage.success('已生成流程，请检查并保存')
  } catch (e) {
    ElMessage.error(e.response?.data?.msg || '生成失败')
  }
}

const handleSubmit = async () => {
  await formRef.value.validate()

  const errMsg = validateBeforeSubmit()
  if (errMsg) {
    ElMessage.warning(errMsg)
    return
  }

  const payload = {
    name: form.value.name,
    biz_type: form.value.biz_type,
    description: form.value.description,
    is_active: form.value.is_active,
    dag_json: dagPreview.value
  }

  if (form.value.id) await updateOrchidWorkflow(form.value.id, payload)
  else await createOrchidWorkflow(payload)

  ElMessage.success('保存成功')
  dialogVisible.value = false
  loadData()
}

const handleDelete = async (id) => {
  await ElMessageBox.confirm('确认删除该流程定义？', '提示', { type: 'warning' })
  await deleteOrchidWorkflow(id)
  ElMessage.success('删除成功')
  loadData()
}

const showConditionSyntaxHelp = () => {
  ElMessageBox.alert(
    [
      '支持运算符：=  :  !=  >  <  >=  <=  ~（包含）  &&  ||',
      '',
      '可用字段：当前业务表中的所有字段，均可直接用于条件判断。',
      '建议写法：优先使用表前缀，避免同名字段歧义（如 employee.status 或 employees.status）。',
      '',
      '示例1：employee.department_id=7 && position.name~经理',
      '示例2：employees.status=1 && employees.email~@company.com',
      '示例3：leave_requests.id>=100 && leave_requests.id<=200'
    ].join('<br/>'),
    '条件语法提示',
    {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '我知道了'
    }
  )
}

onMounted(() => {
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', onMouseUp)
  loadData()
  loadBizTypes()
  loadEmployeeOptions()
  loadPositionOptions()
  loadDepartments()
})
</script>

<style scoped>
.designer-wrap {
  display: grid;
  grid-template-columns: 190px 1fr 280px;
  gap: 12px;
}

.toolbox,
.property-panel {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 10px;
  height: 560px;
  overflow: auto;
  background: #fff;
}

.toolbox-title {
  font-weight: 600;
  margin-bottom: 8px;
  color: #303133;
}

.tool-btn {
  width: 100%;
  margin: 0 0 8px 0;
}

.tool-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.7;
}

.canvas-wrap {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  height: 560px;
  overflow: hidden;
}

.canvas {
  position: relative;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, #f8f9fc 1px, transparent 1px), linear-gradient(#f8f9fc 1px, transparent 1px);
  background-size: 24px 24px;
}

.edge-layer {
  position: absolute;
  inset: 0;
  pointer-events: auto;
}

.wf-node {
  position: absolute;
  width: 170px;
  height: 64px;
  border-radius: 8px;
  border: 1px solid #d9ecff;
  background: #ecf5ff;
  padding: 8px 10px;
  cursor: move;
  user-select: none;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
}

.wf-node.selected {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.wf-node.source {
  border-color: #67c23a;
  box-shadow: 0 0 0 2px rgba(103, 194, 58, 0.2);
}

.wf-node.type-start {
  background: #f0f9eb;
  border-color: #c2e7b0;
}

.wf-node.type-end {
  background: #fef0f0;
  border-color: #fbc4c4;
}

.wf-node.type-condition {
  background: #fdf6ec;
  border-color: #f5dab1;
}

.wf-node.type-copy {
  background: #f4f4f5;
  border-color: #d3d4d6;
}

.node-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-sub {
  margin-top: 6px;
  color: #606266;
  font-size: 12px;
}
</style>
