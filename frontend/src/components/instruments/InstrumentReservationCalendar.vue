<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ChevronLeft, ChevronRight, Calendar as CalendarIcon, User } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Dialog from '@/components/ui/Dialog.vue'
import Input from '@/components/ui/Input.vue'


// Basic types
import { fetchReservations, createReservation, cancelReservation, type Reservation as ApiReservation } from '@/api/reservations'

// Basic types
interface Reservation {
  id: string
  userId: string
  userName: string
  startTime: Date
  endTime: Date
  type: 'usage' | 'maintenance'
  description?: string
}

const props = defineProps<{
  instrumentId?: string | number
}>()

// State
const currentDate = ref(new Date())
const reservations = ref<Reservation[]>([])
const showReservationDialog = ref(false)
const selectedSlot = ref<{ start: Date, end: Date } | null>(null)
const newReservationForm = ref({ project: '', description: '' })
const hoverReservation = ref<Reservation | null>(null)

// Constants
const HOURS = Array.from({ length: 24 }, (_, i) => i) // 0:00 - 23:00
const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
const CELL_HEIGHT = 28 // px per hour

// Helper: Get start of current week (Monday)
const startOfWeek = computed(() => {
    const d = new Date(currentDate.value)
    const day = d.getDay()
    const diff = d.getDate() - day + (day === 0 ? -6 : 1) // adjust when day is sunday
    d.setDate(diff)
    d.setHours(0, 0, 0, 0)
    return d
})

// Helper: Get dates for the week
const weekDates = computed(() => {
    return DAYS.map((_, i) => {
        const d = new Date(startOfWeek.value)
        d.setDate(d.getDate() + i)
        return d
    })
})

const currentMonthYear = computed(() => {
    return startOfWeek.value.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long' })
})

// Navigation
const prevWeek = () => {
    const d = new Date(currentDate.value)
    d.setDate(d.getDate() - 7)
    currentDate.value = d
}

const nextWeek = () => {
    const d = new Date(currentDate.value)
    d.setDate(d.getDate() + 7)
    currentDate.value = d
}

const today = () => {
    currentDate.value = new Date()
}

// Data Loading
const loadReservations = async () => {
    if (!props.instrumentId) return
    
    try {
        const start = new Date(startOfWeek.value)
        const end = new Date(start)
        end.setDate(end.getDate() + 7)
        
        const apiData = await fetchReservations(props.instrumentId, start, end)
        
        // Map API data to local format
        reservations.value = (apiData || []).map((r: ApiReservation) => ({
            id: r.ID.toString(),
            userId: r.user_id,
            userName: r.user_name,
            startTime: new Date(r.start_time),
            endTime: new Date(r.end_time),
            type: r.type,
            description: r.description
        }))
    } catch (e) {
        console.error("Failed to load reservations", e)
        reservations.value = []
    }
}

// Watchers
watch(startOfWeek, loadReservations, { immediate: true })
watch(() => props.instrumentId, loadReservations, { immediate: true })

// Calendar Rendering Logic
const getReservationStyle = (res: Reservation) => {
    const startHour = res.startTime.getHours() + res.startTime.getMinutes() / 60
    const endHour = res.endTime.getHours() + res.endTime.getMinutes() / 60
    const duration = endHour - startHour
    
    // Relative to grid start (0:00)
    const top = startHour * CELL_HEIGHT
    const height = duration * CELL_HEIGHT
    
    return {
        top: `${top}px`,
        height: `${height}px`,
        left: '2px',
        right: '2px'
    }
}

const isSameDay = (d1: Date, d2: Date) => {
    return d1.getFullYear() === d2.getFullYear() &&
           d1.getMonth() === d2.getMonth() &&
           d1.getDate() === d2.getDate()
}

const getDailyReservations = (date: Date) => {
    return reservations.value.filter(res => isSameDay(res.startTime, date))
}


const isDragging = ref(false)
const dragSelection = ref<{ date: Date, start: Date, end: Date, startHour: number, endHour: number } | null>(null)

const startDrag = (date: Date, hour: number) => {
    isDragging.value = true
    const start = new Date(date)
    start.setHours(hour, 0, 0, 0)
    const end = new Date(start)
    end.setHours(hour + 1, 0, 0, 0)
    
    dragSelection.value = {
        date: new Date(date),
        start,
        end,
        startHour: hour,
        endHour: hour + 1
    }
}

const updateDrag = (date: Date, hour: number) => {
    if (!isDragging.value || !dragSelection.value) return
    
    // Only allow dragging within the same day for simplicity v1
    if (!isSameDay(date, dragSelection.value.date)) return

    const currentHour = hour
    const startH = dragSelection.value.startHour
    
    // Update end time based on min/max
    const minH = Math.min(startH, currentHour)
    const maxH = Math.max(startH, currentHour)
    
    const start = new Date(dragSelection.value.date)
    start.setHours(minH, 0, 0, 0)
    
    const end = new Date(dragSelection.value.date)
    end.setHours(maxH + 1, 0, 0, 0) // +1 because selection includes the hovered slot
    
    dragSelection.value.start = start
    dragSelection.value.end = end
    dragSelection.value.endHour = maxH + 1
}

const endDrag = () => {
    if (!isDragging.value || !dragSelection.value) return
    isDragging.value = false
    
    selectedSlot.value = {
        start: dragSelection.value.start,
        end: dragSelection.value.end
    }
    
    newReservationForm.value = { project: '', description: '' }
    showReservationDialog.value = true
    dragSelection.value = null
}

// Reservation Details
const showDetailDialog = ref(false)
const selectedReservation = ref<Reservation | null>(null)

const openReservationDetails = (res: Reservation) => {
    selectedReservation.value = res
    showDetailDialog.value = true
}

const handleCancelReservation = async () => {
    if (!selectedReservation.value) return
    
    try {
        // ID is string in local view, but API needs number
        await cancelReservation(Number(selectedReservation.value.id))
        showDetailDialog.value = false
        selectedReservation.value = null
        loadReservations() // reload
    } catch (e) {
        alert("取消失败")
    }
}

const getDragSelectionStyle = () => {
    if (!dragSelection.value) return {}
    const startH = dragSelection.value.start.getHours()
    const endH = dragSelection.value.end.getHours()
    const duration = endH - startH
    
    const top = startH * CELL_HEIGHT
    const height = duration * CELL_HEIGHT
    
    return {
        top: `${top}px`,
        height: `${height}px`,
        left: '4px',
        right: '4px'
    }
}

const confirmReservation = async () => {
    if (!selectedSlot.value || !props.instrumentId) return
    
    try {
        await createReservation({
            instrument_id: Number(props.instrumentId),
            start_time: selectedSlot.value.start.toISOString(),
            end_time: selectedSlot.value.end.toISOString(),
            type: 'usage',
            description: newReservationForm.value.project + (newReservationForm.value.description ? ` (${newReservationForm.value.description})` : '')
        })
        
        showReservationDialog.value = false
        loadReservations() // reload
    } catch (e: any) {
        console.error(e)
        alert(e.response?.data?.error || "预约失败")
    }
}

const formatValues = (d: Date) => {
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

</script>

<template>
  <div class="flex flex-col bg-white rounded-xl select-none">
    <!-- Header -->
    <div class="flex items-center justify-between p-4 border-b">
        <div class="flex items-center gap-4">
            <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                <CalendarIcon class="w-5 h-5 text-gray-500" />
                {{ currentMonthYear }}
            </h2>
            <div class="flex items-center bg-gray-100 rounded-lg p-1">
                <button @click="prevWeek" class="p-1 hover:bg-white rounded shadow-sm transition-all text-gray-600"><ChevronLeft class="w-4 h-4" /></button>
                <button @click="today" class="px-3 text-xs font-medium text-gray-600 hover:bg-white rounded mx-1 transition-all">本周</button>
                <button @click="nextWeek" class="p-1 hover:bg-white rounded shadow-sm transition-all text-gray-600"><ChevronRight class="w-4 h-4" /></button>
            </div>
        </div>
        <div class="flex gap-4 text-xs font-medium text-gray-500">
            <div class="flex items-center gap-2"><div class="w-2.5 h-2.5 bg-blue-500 rounded-sm"></div> 我的预约</div>
            <div class="flex items-center gap-2"><div class="w-2.5 h-2.5 bg-gray-400 rounded-sm opacity-50"></div>他人预约</div>
            <div class="flex items-center gap-2"><div class="w-2.5 h-2.5 bg-orange-400 rounded-sm"></div> 维保</div>
        </div>
    </div>

    <!-- Calendar Body -->
    <div class="flex-1 relative flex">
        <!-- Time Column -->
        <div class="flex-none flex flex-col w-12 border-r bg-gray-50/50 sticky left-0 z-20">
             <!-- Sticky top-left corner -->
             <div class="h-10 border-b bg-gray-50 sticky top-0 z-30 shadow-sm"></div> 
             
             <div class="relative">
                 <!-- Time Slots -->
                 <div v-for="h in HOURS" :key="h" 
                      class="flex items-center justify-center text-xs text-gray-400 font-medium border-b border-dashed border-gray-100/0 box-border relative"
                      :style="{ height: `${CELL_HEIGHT}px` }">
                     <span class="-translate-y-1/2 absolute top-0 bg-white/80 px-1 rounded text-[10px]">{{ h }}:00</span>
                 </div>
                 <!-- Final End Label (23:00) -->
                 <div class="absolute bottom-0 w-full flex justify-center">
                    <span class="translate-y-1/2 text-xs text-gray-400 font-medium bg-white/80 px-1 rounded text-[10px]">{{ (HOURS[HOURS.length-1] ?? 22) + 1 }}:00</span>
                 </div>
             </div>
        </div>

        <!-- Days Columns -->
        <div class="flex-1 flex min-w-[600px]">
            <div v-for="(date, idx) in weekDates" :key="idx" class="flex-1 border-r min-w-[100px] flex flex-col">
                <!-- Day Header -->
                <div class="h-10 border-b flex items-center justify-center gap-2 sticky top-0 bg-white z-10 shadow-sm"
                     :class="{'bg-blue-50/50 text-blue-600': isSameDay(date, new Date())}">
                    <span class="text-xs font-medium uppercase text-gray-500" :class="{'text-blue-500': isSameDay(date, new Date())}">{{ DAYS[idx] }}</span>
                    <span class="text-sm font-bold w-6 h-6 flex items-center justify-center rounded-full"
                          :class="{'bg-blue-600 text-white shadow-sm': isSameDay(date, new Date()), 'text-gray-900': !isSameDay(date, new Date())}">
                          {{ date.getDate() }}
                    </span>
                </div>

                <!-- Day Slots Container -->
                <div class="relative flex-1 bg-white" 
                     @mouseleave="endDrag"
                     @mouseup="endDrag">
                    <!-- Background Grid -->
                    <div v-for="h in HOURS" :key="h" 
                         class="border-b border-gray-100 group hover:bg-gray-50/30 transition-colors cursor-pointer select-none"
                         :style="{ height: `${CELL_HEIGHT}px` }"
                         @mousedown="startDrag(date, h)"
                         @mouseenter="updateDrag(date, h)">
                    </div>

                    <!-- Drag Preview (Ghost) -->
                    <div v-if="isDragging && dragSelection && isSameDay(dragSelection.date, date)"
                         class="absolute bg-blue-100/50 border border-blue-300 border-dashed rounded z-10 pointer-events-none transition-all duration-75"
                         :style="getDragSelectionStyle()">
                         <div class="text-[10px] p-1 font-mono text-blue-600 font-bold">
                             {{ formatValues(dragSelection.start) }} - {{ formatValues(dragSelection.end) }}
                         </div>
                    </div>

                    <!-- Reservations Overlay -->
                    <div v-for="res in getDailyReservations(date)" :key="res.id"
                         class="absolute rounded-md border p-1 flex flex-col overflow-hidden transition-all hover:scale-[1.02] hover:shadow-md hover:z-20 cursor-pointer group/card"
                         :class="{
                             'bg-blue-50 border-blue-200 text-blue-700': res.type === 'usage' && res.userId === 'current',
                             'bg-gray-50 border-gray-200 text-gray-500': res.type === 'usage' && res.userId !== 'current',
                             'bg-orange-50 border-orange-200 text-orange-700': res.type === 'maintenance'
                         }"
                         :style="getReservationStyle(res)"
                         @mouseenter="hoverReservation = res"
                         @mouseleave="hoverReservation = null"
                         @click.stop="openReservationDetails(res)">
                         
                         <!-- Header Line: User Name -->
                         <div class="font-bold text-[11px] truncate flex items-center gap-1 shrink-0 h-4 leading-none">
                             <User class="w-2.5 h-2.5" v-if="res.userId !== 'current'" />
                             <span v-if="res.userId === 'current'">我的预约</span>
                             <span v-else>{{ res.userName }}</span>
                         </div>
                         
                         <!-- Middle: Description (Hidden if too short) -->
                         <div class="truncate opacity-80 text-[10px] leading-tight min-h-0 flex-1 py-0.5"
                              v-if="(res.endTime.getTime() - res.startTime.getTime()) > 3600000"> <!-- Only show desc if > 1 hour -->
                             {{ res.description }}
                         </div>
                         
                         <!-- Footer: Time -->
                         <div class="mt-auto text-[9px] opacity-70 font-mono leading-none shrink-0"
                              v-if="(res.endTime.getTime() - res.startTime.getTime()) >= 3600000">
                             {{ formatValues(res.startTime) }} - {{ formatValues(res.endTime) }}
                         </div>
                    </div>
                    
                    <!-- Current Time Indicator (if same day) -->
                    <div v-if="isSameDay(date, new Date())" 
                         class="absolute w-full border-t-2 border-red-500 z-0 pointer-events-none opacity-50"
                         :style="{ top: `${(new Date().getHours() + new Date().getMinutes()/60) * CELL_HEIGHT}px` }">
                         <div class="absolute -left-1.5 -top-1 w-2 h-2 rounded-full bg-red-500"></div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- New Reservation Dialog -->
    <Dialog :open="showReservationDialog" @close="showReservationDialog = false" title="新建预约申请" maxWidth="max-w-sm">
        <div class="space-y-4" v-if="selectedSlot">
            <div class="bg-blue-50 rounded-lg p-3 text-blue-800 text-sm flex gap-4">
                <div>
                    <span class="block text-xs text-blue-500 uppercase font-bold">日期</span>
                    <span class="font-medium">{{ selectedSlot.start.toLocaleDateString() }}</span>
                </div>
                <div>
                    <span class="block text-xs text-blue-500 uppercase font-bold">时间</span>
                    <span class="font-medium font-mono">{{ formatValues(selectedSlot.start) }} - {{ formatValues(selectedSlot.end) }}</span>
                </div>
            </div>
            
            <div>
                <label class="text-sm font-medium mb-1.5 block">实验项目名称</label>
                <Input v-model="newReservationForm.project" placeholder="例如: 细胞培养观察" />
            </div>
            
            <div>
                <label class="text-sm font-medium mb-1.5 block">备注说明</label>
                <Input v-model="newReservationForm.description" placeholder="可选填" />
            </div>
        </div>
        <template #footer>
            <Button variant="ghost" @click="showReservationDialog = false">取消</Button>
            <Button @click="confirmReservation">提交预约</Button>
        </template>
    </Dialog>

    <!-- Reservation Detail Dialog -->
    <Dialog :open="showDetailDialog" @close="showDetailDialog = false" title="预约详情" maxWidth="max-w-sm">
        <div class="space-y-4" v-if="selectedReservation">
            <div class="flex items-center gap-3 pb-4 border-b">
                <div class="w-10 h-10 rounded-full flex items-center justify-center"
                     :class="selectedReservation.userId === 'current' ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-600'">
                    <User class="w-5 h-5" />
                </div>
                <div>
                    <div class="font-bold text-gray-900">{{ selectedReservation.userId === 'current' ? '我的预约' : selectedReservation.userName }}</div>
                    <div class="text-xs text-gray-500 uppercase">{{ selectedReservation.type === 'usage' ? '常规使用' : '设备维护' }}</div>
                </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
                 <div>
                    <span class="block text-xs text-gray-500 uppercase font-bold mb-0.5">开始时间</span>
                    <span class="text-sm font-medium font-mono text-gray-900">{{ formatValues(selectedReservation.startTime) }}</span>
                </div>
                <div>
                    <span class="block text-xs text-gray-500 uppercase font-bold mb-0.5">结束时间</span>
                    <span class="text-sm font-medium font-mono text-gray-900">{{ formatValues(selectedReservation.endTime) }}</span>
                </div>
            </div>

            <div>
                <span class="block text-xs text-gray-500 uppercase font-bold mb-0.5">实验项目</span>
                <p class="text-sm text-gray-900">{{ selectedReservation.description || '无描述' }}</p>
            </div>

            <div v-if="selectedReservation.userId === 'current'" class="pt-4 border-t flex justify-end">
                 <Button variant="destructive" size="sm" @click="handleCancelReservation">
                    取消预约
                 </Button>
            </div>
        </div>
        <template #footer>
             <Button variant="secondary" @click="showDetailDialog = false">关闭</Button>
        </template>
    </Dialog>
  </div>
</template>
