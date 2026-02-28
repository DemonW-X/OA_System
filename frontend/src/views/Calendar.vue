<template>
  <el-card>
    <template #header>
      <div style="display:flex;align-items:center">
        <span>企业行事历</span>
      </div>
    </template>

    <!-- 月份切换 -->
    <div style="display:flex;align-items:center;justify-content:center;gap:16px;margin-bottom:16px">
      <el-button :icon="ArrowLeft" circle @click="prevMonth" />
      <span style="font-size:18px;font-weight:bold;min-width:120px;text-align:center">
        {{ currentYear }} 年 {{ currentMonth + 1 }} 月
      </span>
      <el-button :icon="ArrowRight" circle @click="nextMonth" />
      <el-button size="small" @click="goToday">今天</el-button>
    </div>

    <!-- 图例 -->
    <div style="display:flex;gap:16px;margin-bottom:12px;flex-wrap:wrap">
      <span v-for="t in eventTypes" :key="t.value" style="display:flex;align-items:center;gap:4px;font-size:13px">
        <span :style="{width:'12px',height:'12px',borderRadius:'50%',background:t.color,display:'inline-block'}" />
        {{ t.label }}
      </span>
    </div>

    <!-- 日历主体 -->
    <div class="calendar-grid">
      <!-- 星期头 -->
      <div v-for="w in weekDays" :key="w" class="calendar-weekday">{{ w }}</div>
      <!-- 日期格子 -->
      <div
        v-for="(day, idx) in calendarDays"
        :key="idx"
        class="calendar-cell"
        :class="{
          'other-month': !day.currentMonth,
          'today': day.isToday,
          'has-events': day.events.length > 0
        }"
      >
        <div class="day-number">{{ day.date }}</div>
        <div class="event-list">
          <div
            v-for="ev in day.events.slice(0, 3)"
            :key="ev.id"
            class="event-tag"
            :style="{ background: getTypeColor(ev.type) }"
            @click.stop="openDetail(ev)"
            :title="ev.title"
          >
            {{ ev.title }}
          </div>
          <div
            v-if="day.events.length > 3"
            class="event-more"
            @click.stop="openDayEvents(day)"
          >
            +{{ day.events.length - 3 }} 更多
          </div>
        </div>
      </div>
    </div>

    <!-- 事件详情对话框 -->
    <el-dialog v-model="detailVisible" title="事件详情" width="420px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="标题">{{ detailEvent.title }}</el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag :color="getTypeColor(detailEvent.type)" style="color:#fff;border:none">
            {{ getTypeLabel(detailEvent.type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(detailEvent.start_time) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(detailEvent.end_time) }}</el-descriptions-item>
        <el-descriptions-item label="描述" v-if="detailEvent.description">{{ detailEvent.description }}</el-descriptions-item>
        <el-descriptions-item label="会议室">{{ detailEvent.meeting_room?.name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="参与人员">
          <span v-if="detailParticipants.length">
            <el-tag v-for="p in detailParticipants" :key="p.id" style="margin:2px">{{ p.name }}</el-tag>
          </span>
          <span v-else style="color:#999">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="创建人">{{ detailEvent.created_by }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 当天全部事件 -->
    <el-dialog v-model="dayEventsVisible" :title="`${dayEventsDate} 的事件`" width="420px">
      <div v-for="ev in dayEventsList" :key="ev.id" class="day-event-item" @click="openDetail(ev); dayEventsVisible = false">
        <span :style="{width:'10px',height:'10px',borderRadius:'50%',background:getTypeColor(ev.type),display:'inline-block',marginRight:'8px',flexShrink:0}" />
        <span style="flex:1">{{ ev.title }}</span>
        <span style="color:#999;font-size:12px">{{ formatTime(ev.start_time).slice(11, 16) }}</span>
      </div>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { getEventBookings } from '../api/event_booking'
import { getEmployees } from '../api/employee'

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

const eventTypes = [
  { value: 'holiday',  label: '节假日', color: '#F56C6C' },
  { value: 'meeting',  label: '会议',   color: '#409EFF' },
  { value: 'activity', label: '活动',   color: '#67C23A' },
  { value: 'other',    label: '其他',   color: '#909399' },
]

const getTypeColor = (type) => eventTypes.find(t => t.value === type)?.color || '#909399'
const getTypeLabel = (type) => eventTypes.find(t => t.value === type)?.label || '其他'

const now = new Date()
const currentYear  = ref(now.getFullYear())
const currentMonth = ref(now.getMonth())

const events = ref([])
const employeeList = ref([])

const loadEmployees = async () => {
  const res = await getEmployees({ page: 1, page_size: 999 })
  employeeList.value = res.data.data?.list || []
}

const currentMonthStr = computed(() => {
  const m = String(currentMonth.value + 1).padStart(2, '0')
  return `${currentYear.value}-${m}`
})

const loadEvents = async () => {
  const res = await getEventBookings({ page: 1, page_size: 999 })
  const all = res.data.data?.list || []
  const monthStr = currentMonthStr.value
  events.value = all.filter(ev => {
    const start = ev.start_time?.slice(0, 7)
    const end   = ev.end_time?.slice(0, 7)
    return start <= monthStr && monthStr <= end
  })
}

const prevMonth = () => {
  if (currentMonth.value === 0) { currentYear.value--; currentMonth.value = 11 }
  else currentMonth.value--
  loadEvents()
}
const nextMonth = () => {
  if (currentMonth.value === 11) { currentYear.value++; currentMonth.value = 0 }
  else currentMonth.value++
  loadEvents()
}
const goToday = () => {
  currentYear.value  = now.getFullYear()
  currentMonth.value = now.getMonth()
  loadEvents()
}

// 构建日历格子
const calendarDays = computed(() => {
  const year  = currentYear.value
  const month = currentMonth.value
  const firstDay = new Date(year, month, 1).getDay()
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const daysInPrev  = new Date(year, month, 0).getDate()
  const todayStr = `${now.getFullYear()}-${String(now.getMonth()+1).padStart(2,'0')}-${String(now.getDate()).padStart(2,'0')}`

  const days = []

  // 上月补位
  for (let i = firstDay - 1; i >= 0; i--) {
    const d = daysInPrev - i
    const m = month === 0 ? 12 : month
    const y = month === 0 ? year - 1 : year
    const dateStr = `${y}-${String(m).padStart(2,'0')}-${String(d).padStart(2,'0')}`
    days.push({ date: d, currentMonth: false, isToday: false, dateStr, events: getEventsForDay(dateStr) })
  }

  // 本月
  for (let d = 1; d <= daysInMonth; d++) {
    const dateStr = `${year}-${String(month+1).padStart(2,'0')}-${String(d).padStart(2,'0')}`
    days.push({ date: d, currentMonth: true, isToday: dateStr === todayStr, dateStr, events: getEventsForDay(dateStr) })
  }

  // 下月补位（补满6行42格）
  const remaining = 42 - days.length
  for (let d = 1; d <= remaining; d++) {
    const m = month === 11 ? 1 : month + 2
    const y = month === 11 ? year + 1 : year
    const dateStr = `${y}-${String(m).padStart(2,'0')}-${String(d).padStart(2,'0')}`
    days.push({ date: d, currentMonth: false, isToday: false, dateStr, events: getEventsForDay(dateStr) })
  }

  return days
})

const getEventsForDay = (dateStr) => {
  return events.value.filter(ev => {
    const start = ev.start_time?.slice(0, 10)
    const end   = ev.end_time?.slice(0, 10)
    return start <= dateStr && dateStr <= end
  })
}

const formatTime = (t) => {
  if (!t) return ''
  return t.replace('T', ' ').slice(0, 19)
}

// 详情
const detailVisible = ref(false)
const detailEvent   = ref({})
const detailParticipants = ref([])

const openDetail = (ev) => {
  detailEvent.value = ev
  const ids = ev.participants ? JSON.parse(ev.participants) : []
  detailParticipants.value = ids.map(id => employeeList.value.find(e => e.id === id)).filter(Boolean)
  detailVisible.value = true
}

// 当天更多事件
const dayEventsVisible = ref(false)
const dayEventsDate    = ref('')
const dayEventsList    = ref([])
const openDayEvents = (day) => {
  dayEventsDate.value = day.dateStr
  dayEventsList.value = day.events
  dayEventsVisible.value = true
}

onMounted(() => { loadEmployees(); loadEvents() })
</script>

<style scoped>
.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  border-top: 1px solid #eee;
  border-left: 1px solid #eee;
}
.calendar-weekday {
  text-align: center;
  padding: 8px 0;
  font-weight: 600;
  font-size: 13px;
  color: #606266;
  background: #fafafa;
  border-right: 1px solid #eee;
  border-bottom: 1px solid #eee;
}
.calendar-cell {
  min-height: 100px;
  padding: 4px;
  border-right: 1px solid #eee;
  border-bottom: 1px solid #eee;
  vertical-align: top;
  cursor: default;
}
.calendar-cell.other-month { background: #fafafa; }
.calendar-cell.other-month .day-number { color: #c0c4cc; }
.calendar-cell.today { background: #ecf5ff; }
.calendar-cell.today .day-number {
  background: #409EFF;
  color: #fff;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  line-height: 24px;
  text-align: center;
}
.day-number {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 2px;
  width: 24px;
  height: 24px;
  line-height: 24px;
  text-align: center;
}
.event-list { display: flex; flex-direction: column; gap: 2px; }
.event-tag {
  font-size: 11px;
  color: #fff;
  padding: 1px 5px;
  border-radius: 3px;
  cursor: pointer;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.event-tag:hover { opacity: 0.85; }
.event-more {
  font-size: 11px;
  color: #409EFF;
  cursor: pointer;
  padding-left: 4px;
}
.day-event-item {
  display: flex;
  align-items: center;
  padding: 8px 4px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
}
.day-event-item:hover { background: #f5f7fa; }
.day-event-item:last-child { border-bottom: none; }
</style>
