<template>
  <el-row :gutter="16">
    <el-col :span="8" class="schedule-left-col">
      <el-card shadow="never" class="schedule-card schedule-calendar-card" style="margin-bottom:12px">
        <div class="mini-calendar-wrap">
          <el-calendar v-model="selectedDate">
            <template #header>
              <div class="calendar-head-custom">
                <el-button text class="calendar-nav-btn" @click="changeMonth(-1)"><</el-button>
                <div class="calendar-head-center" @mouseenter="monthPanelVisible = true" @mouseleave="monthPanelVisible = false">
                  <span class="calendar-head-title" @click="monthPanelVisible = !monthPanelVisible">{{ calendarHeaderText }}</span>
                  <div v-show="monthPanelVisible" class="calendar-month-panel">
                    <div class="calendar-month-panel-title">{{ currentYear }}</div>
                    <div class="calendar-month-grid">
                      <el-button
                        v-for="(m, idx) in monthNames"
                        :key="m"
                        size="small"
                        text
                        class="calendar-month-btn"
                        :class="{ active: idx === selectedMonthIndex }"
                        @click="chooseMonth(idx)"
                      >
                        {{ m }}
                      </el-button>
                    </div>
                  </div>
                </div>
                <el-button text class="calendar-nav-btn" @click="changeMonth(1)">></el-button>
              </div>
            </template>
          </el-calendar>
        </div>
      </el-card>

      <el-card shadow="never" class="schedule-card schedule-today-card" style="height:100%">
        <template #header>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span>今日事项</span>
            <el-tag type="primary">{{ todayItems.length }} 项</el-tag>
          </div>
        </template>
        <el-empty v-if="!todayItems.length" description="今日暂无事项" :image-size="64" />
        <div v-else class="today-list">
          <div v-for="item in todayItems" :key="item.id" class="today-item">
            <div style="font-weight:600">{{ item.title }}</div>
            <div style="font-size:12px;color:#909399">{{ item.timeText }}</div>
            <el-tag size="small" :type="statusTag(item.timeStatus)">{{ statusText(item.timeStatus) }}</el-tag>
          </div>
        </div>
      </el-card>
    </el-col>

    <el-col :span="16">
      <el-card shadow="never" class="schedule-card schedule-right-card" style="height:100%">
        <template #header>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span>{{ currentMonthEnglish }}</span>
            <el-button size="small" @click="loadData">刷新</el-button>
          </div>
        </template>

        <el-empty v-if="!rangeItems.length" description="近7日暂无行程" :image-size="72" />

        <div v-else class="gantt-v-wrap">
          <div class="gantt-v-header">
            <div class="gantt-v-time-col"></div>
            <div v-for="d in ganttColumns" :key="d.date" class="gantt-v-day-head">{{ d.label }}</div>
          </div>

          <div class="gantt-v-body">
            <div class="gantt-v-time-axis">
              <div v-for="h in hourMarks" :key="h" class="gantt-v-time-mark" :style="{ top: `${(h / 24) * 100}%` }">
                {{ String(h).padStart(2, '0') }}:00
              </div>
            </div>

            <div class="gantt-v-grid">
              <div v-for="d in ganttColumns" :key="d.date" class="gantt-v-day-col">
                <div
                  v-for="s in d.segments"
                  :key="s.id"
                  class="gantt-v-bar"
                  :class="s.timeStatus"
                  :style="{ top: s.top + '%', height: s.height + '%' }"
                  :title="`${s.title} ${s.timeText}`"
                >
                  <span class="gantt-bar-text">{{ s.title }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { getEventBookings } from '../api/event_booking'

const selectedDate = ref(new Date())
const allEvents = ref([])
const monthPanelVisible = ref(false)
const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
let refreshTimer = null

const calendarHeaderText = computed(() => {
  const d = selectedDate.value || new Date()
  const dt = d instanceof Date ? d : new Date(d)
  const m = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'][dt.getMonth()]
  return `${dt.getFullYear()} ${m}`
})

const currentYear = computed(() => {
  const d = selectedDate.value || new Date()
  const dt = d instanceof Date ? d : new Date(d)
  return dt.getFullYear()
})

const selectedMonthIndex = computed(() => {
  const d = selectedDate.value || new Date()
  const dt = d instanceof Date ? d : new Date(d)
  return dt.getMonth()
})

const changeMonth = (offset) => {
  const d = selectedDate.value || new Date()
  const dt = new Date(d)
  dt.setMonth(dt.getMonth() + offset)
  selectedDate.value = dt
}

const chooseMonth = (monthIndex) => {
  const d = selectedDate.value || new Date()
  const dt = new Date(d)
  dt.setMonth(monthIndex)
  dt.setDate(1)
  selectedDate.value = dt
  monthPanelVisible.value = false
}

const pad = (n) => String(n).padStart(2, '0')
const toDate = (v) => {
  if (!v) return null
  const s = String(v).replace('T', ' ')
  return new Date(s)
}
const formatDate = (d, mode = 'YYYY-MM-DD') => {
  if (!d) return ''
  const dt = d instanceof Date ? d : new Date(d)
  const y = dt.getFullYear()
  const m = pad(dt.getMonth() + 1)
  const day = pad(dt.getDate())
  const hh = pad(dt.getHours())
  const mm = pad(dt.getMinutes())
  if (mode === 'YYYY-MM-DD') return `${y}-${m}-${day}`
  return `${y}-${m}-${day} ${hh}:${mm}`
}
const startOfDay = (d) => new Date(d.getFullYear(), d.getMonth(), d.getDate(), 0, 0, 0)
const endOfDay = (d) => new Date(d.getFullYear(), d.getMonth(), d.getDate(), 23, 59, 59)
const addDays = (d, n) => new Date(d.getFullYear(), d.getMonth(), d.getDate() + n)

const now = () => new Date()

const currentMonthEnglish = computed(() => {
  const m = now().getMonth()
  return ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'][m]
})

const formatDayLabelEnglish = (d) => {
  if (!d) return '-'
  const dt = d instanceof Date ? d : new Date(d)
  const month = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'][dt.getMonth()]
  const day = pad(dt.getDate())
  const week = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][dt.getDay()]
  return `${month}-${day} ${week}`
}

const mapEvent = (ev) => {
  const start = toDate(ev.start_time)
  const end = toDate(ev.end_time)
  const n = now()
  let timeStatus = 'upcoming'
  if (start && end) {
    if (n < start) timeStatus = 'upcoming'
    else if (n > end) timeStatus = 'finished'
    else timeStatus = 'ongoing'
  }
  return {
    ...ev,
    _start: start,
    _end: end,
    timeStatus,
    timeText: `${start ? formatDate(start, 'HH:mm') : '--:--'}-${end ? formatDate(end, 'HH:mm') : '--:--'}`,
    fullRangeText: `${start ? formatDate(start) : '-'} ~ ${end ? formatDate(end) : '-'}`,
    dayLabel: start ? formatDayLabelEnglish(start) : '-',
  }
}

const todayItems = computed(() => {
  const t = new Date()
  const st = startOfDay(t)
  const et = endOfDay(t)
  return allEvents.value
    .filter(ev => ev._start && ev._end && ev._start <= et && ev._end >= st)
    .sort((a, b) => (a._start?.getTime() || 0) - (b._start?.getTime() || 0))
})

const rangeItems = computed(() => {
  const t = new Date()
  const rangeStart = startOfDay(addDays(t, -3))
  const rangeEnd = endOfDay(addDays(t, 3))
  return allEvents.value
    .filter(ev => ev._start && ev._end && ev._start <= rangeEnd && ev._end >= rangeStart)
    .sort((a, b) => (a._start?.getTime() || 0) - (b._start?.getTime() || 0))
})

const hourMarks = [0, 4, 8, 12, 16, 20, 24]

const ganttColumns = computed(() => {
  const t = new Date()
  const days = []
  for (let i = -3; i <= 3; i++) {
    const d = addDays(t, i)
    const dayStart = startOfDay(d)
    const dayEnd = endOfDay(d)
    const label = formatDayLabelEnglish(d)

    const segments = rangeItems.value
      .filter(ev => ev._start && ev._end && ev._start <= dayEnd && ev._end >= dayStart)
      .map(ev => {
        const segStart = ev._start < dayStart ? dayStart : ev._start
        const segEnd = ev._end > dayEnd ? dayEnd : ev._end
        const startMin = segStart.getHours() * 60 + segStart.getMinutes()
        const endMin = segEnd.getHours() * 60 + segEnd.getMinutes()
        const top = (startMin / 1440) * 100
        const height = Math.max(((endMin - startMin) / 1440) * 100, 1.2)
        return {
          ...ev,
          top,
          height,
        }
      })

    days.push({
      date: formatDate(d, 'YYYY-MM-DD'),
      label,
      segments,
    })
  }
  return days
})

const statusText = (s) => ({ upcoming: '未开始', ongoing: '进行中', finished: '已结束' }[s] || '未知')
const statusTag = (s) => ({ upcoming: 'info', ongoing: 'success', finished: 'warning' }[s] || 'info')

const loadData = async () => {
  const res = await getEventBookings({ page: 1, page_size: 1000 })
  const list = res.data?.data?.list || []
  allEvents.value = list.map(mapEvent)
}

onMounted(() => {
  loadData()
  // 每10分钟自动刷新一次
  refreshTimer = setInterval(() => {
    loadData()
  }, 10 * 60 * 1000)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<style scoped>
.schedule-left-col {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 106px);
}

.schedule-card :deep(.el-card__header) {
  padding: 8px 10px;
}

.schedule-card :deep(.el-card__body) {
  padding: 8px 10px;
}

.schedule-calendar-card :deep(.el-card__body) {
  padding: 8px 10px;
}

.schedule-today-card {
  flex: 1;
  margin-top: 0;
}

.schedule-today-card :deep(.el-card__body) {
  height: calc(100% - 48px);
  overflow: auto;
}

.schedule-right-card :deep(.el-card__body) {
  height: calc(100vh - 160px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.mini-calendar-wrap {
  width: 100%;
  margin: 0;
}

.calendar-head-custom {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
}

.calendar-head-center {
  position: relative;
}

.calendar-head-title {
  font-weight: 600;
  font-size: 14px;
  min-width: 120px;
  text-align: center;
  cursor: pointer;
  user-select: none;
}

.calendar-head-title:hover {
  color: #409EFF;
}

.calendar-nav-btn {
  min-width: 24px;
  padding: 2px 6px;
  font-size: 16px;
  line-height: 1;
}

.calendar-month-panel {
  position: absolute;
  top: 28px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 20;
  width: 240px;
  height: 240px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  box-shadow: 0 6px 16px rgba(0, 0, 0, .12);
  padding: 8px;
  box-sizing: border-box;
}

.calendar-month-panel-title {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
  text-align: center;
}

.calendar-month-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  grid-template-rows: repeat(3, 1fr);
  gap: 4px;
  height: calc(100% - 24px);
}

.calendar-month-btn {
  justify-content: center;
  height: 100%;
  width: 100%;
  margin: 0 !important;
  border-radius: 6px;
  color: #606266;
}

:deep(.calendar-month-btn:hover) {
  background: #ecf5ff;
  color: #409EFF;
}

:deep(.calendar-month-btn.active) {
  background: #409EFF;
  color: #fff;
}

:deep(.mini-calendar-wrap .el-calendar-table th) {
  padding: 4px 0;
  font-size: 12px;
}

:deep(.mini-calendar-wrap .el-calendar-table td),
:deep(.mini-calendar-wrap .el-calendar-table th) {
  border-color: #ffffff !important;
}

:deep(.mini-calendar-wrap .el-calendar-table .el-calendar-day) {
  min-height: 34px;
  height: 34px;
  padding: 0;
  font-size: 13px;
  line-height: 1.2;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
}

:deep(.mini-calendar-wrap .el-calendar-table td.is-today) {
  background: transparent !important;
}

:deep(.mini-calendar-wrap .el-calendar-table .is-today .el-calendar-day) {
  width: 28px;
  height: 28px;
  min-height: 28px;
  margin: 0 auto;
  border-radius: 50%;
  background: #409EFF;
  color: #fff;
  font-weight: 600;
}

.gantt-v-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
  min-height: 0;
}

.gantt-v-header {
  display: grid;
  grid-template-columns: 56px repeat(7, 1fr);
  gap: 6px;
}

.gantt-v-time-col {
  height: 24px;
}

.gantt-v-day-head {
  text-align: center;
  font-size: 12px;
  color: #606266;
  font-weight: 600;
}

.gantt-v-body {
  display: grid;
  grid-template-columns: 56px 1fr;
  gap: 6px;
  flex: 1;
  min-height: 0;
}

.gantt-v-time-axis {
  position: relative;
  border-right: 1px solid #ebeef5;
}

.gantt-v-time-mark {
  position: absolute;
  left: 0;
  transform: translateY(-50%);
  font-size: 11px;
  color: #909399;
}

.gantt-v-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 6px;
}

.gantt-v-day-col {
  position: relative;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: repeating-linear-gradient(
    to bottom,
    #f5f7fa 0,
    #f5f7fa calc(16.666% - 1px),
    #ffffff calc(16.666% - 1px),
    #ffffff 16.666%
  );
  overflow: hidden;
}

.gantt-v-bar {
  position: absolute;
  left: 6px;
  right: 6px;
  border-radius: 4px;
  padding: 2px 4px;
  display: flex;
  align-items: center;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 11px;
  color: #303133;
  border: 1px solid transparent;
}

.gantt-v-bar.upcoming {
  background: #d9ecff;
  border-color: #b3d8ff;
}
.gantt-v-bar.ongoing {
  background: #e1f3d8;
  border-color: #b3e19d;
}
.gantt-v-bar.finished {
  background: #faecd8;
  border-color: #f3d19e;
}

.gantt-bar-text {
  overflow: hidden;
  text-overflow: ellipsis;
}

.today-list { display: flex; flex-direction: column; gap: 6px; }
.today-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 6px 8px;
}
</style>
