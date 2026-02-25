<script setup lang="ts">
import { ref } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Input from '@/components/ui/Input.vue'
import Dialog from '@/components/ui/Dialog.vue'
import type { Instrument } from '@/api/instruments'

const props = defineProps<{
    instrument: Instrument
}>()

// Reservation Logic
const viewDate = ref(new Date())
const timelineRef = ref<HTMLElement | null>(null)
const isDragging = ref(false)
const dragStartX = ref(0)
const selectionRange = ref<{ start: number, end: number } | null>(null)
const showReservationDialog = ref(false)
const newReservationForm = ref({ project: '', startTime: '', endTime: '' })

// Mock existing reservations (0-14 scale relative to 8:00)
const reservations = ref([
    { id: 1, start: 2, end: 4, user: '李四', project: '天然产物分析' }, // 10:00 - 12:00
    { id: 2, start: 6, end: 7, user: '王五', project: '材料合成' },     // 14:00 - 15:00
])

const timeToPx = (hours: number) => {
   return `${(hours / 14) * 100}%`
}

const pxToHours = (px: number, totalWidth: number) => {
    const ratio = Math.max(0, Math.min(1, px / totalWidth))
    return ratio * 14
}

const formatTime = (offsetHours: number) => {
    const base = 8 // Starts at 8:00
    const totalMinutes = (base + offsetHours) * 60
    const h = Math.floor(totalMinutes / 60)
    const m = Math.floor(totalMinutes % 60)
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`
}

const handleMouseDown = (e: MouseEvent) => {
    if (!timelineRef.value) return
    const rect = timelineRef.value.getBoundingClientRect()
    const x = e.clientX - rect.left
    const h = pxToHours(x, rect.width)
    
    isDragging.value = true
    dragStartX.value = h
    selectionRange.value = { start: h, end: h }
}

const handleMouseMove = (e: MouseEvent) => {
    if (!isDragging.value || !timelineRef.value || !selectionRange.value) return
    const rect = timelineRef.value.getBoundingClientRect()
    const x = e.clientX - rect.left
    const h = pxToHours(x, rect.width)
    
    // Snap to 15m (0.25h)
    const snappedH = Math.round(h * 4) / 4
    
    const start = Math.min(dragStartX.value, snappedH)
    const end = Math.max(dragStartX.value, snappedH)
    
    selectionRange.value = { start, end }
}

const handleMouseUp = () => {
    if (!isDragging.value || !selectionRange.value) return
    isDragging.value = false
    
    // If range is too small (click), default to 1 hour
    if (selectionRange.value.end - selectionRange.value.start < 0.25) {
        selectionRange.value.end = selectionRange.value.start + 1
    }

    // Open Dialog
    newReservationForm.value = {
        project: '',
        startTime: formatTime(selectionRange.value.start),
        endTime: formatTime(selectionRange.value.end)
    }
    showReservationDialog.value = true
}

const handleReservationSubmit = () => {
    if (selectionRange.value) {
        reservations.value.push({
            id: Date.now(),
            start: selectionRange.value.start,
            end: selectionRange.value.end,
            user: '当前用户',
            project: newReservationForm.value.project
        })
    }
    showReservationDialog.value = false
    selectionRange.value = null
    alert('预约成功')
}
</script>

<template>
  <div class="animate-in fade-in">
     <div class="flex items-center justify-between mb-6">
         <div class="flex items-center gap-4">
             <Button variant="outline" size="icon" class="h-8 w-8"><ChevronLeft class="w-4 h-4" /></Button>
             <span class="font-medium text-lg text-gray-900">{{ viewDate.toLocaleDateString() }}</span>
             <Button variant="outline" size="icon" class="h-8 w-8"><ChevronRight class="w-4 h-4" /></Button>
             <Badge variant="secondary">今日</Badge>
         </div>
         <div class="flex gap-4 text-sm text-gray-500">
             <div class="flex items-center gap-2"><div class="w-3 h-3 bg-blue-200 rounded"></div> 已预约</div>
             <div class="flex items-center gap-2"><div class="w-3 h-3 bg-blue-500 rounded"></div> 当前选择</div>
         </div>
     </div>

     <!-- Timeline -->
     <div class="relative select-none" 
          @mousemove="handleMouseMove" 
          @mouseup="handleMouseUp" 
          @mouseleave="handleMouseUp">
          
         <!-- Time Labels -->
         <div class="flex justify-between text-xs text-gray-400 mb-2 px-1">
             <span v-for="h in 15" :key="h">{{ (h + 7).toString().padStart(2,'0') }}:00</span>
         </div>

         <!-- Track -->
         <div ref="timelineRef" 
              class="h-16 bg-gray-50 rounded-lg border relative cursor-crosshair overflow-hidden"
              @mousedown="handleMouseDown">
              
              <!-- Grid Lines -->
              <div class="absolute inset-0 flex justify-between pointer-events-none">
                  <div v-for="h in 15" :key="h" class="h-full w-px bg-gray-200/50 first:bg-transparent last:bg-transparent"></div>
              </div>

              <!-- Existing Reservations -->
              <div v-for="res in reservations" :key="res.id"
                   class="absolute top-2 bottom-2 bg-blue-100 border border-blue-200 rounded-md flex items-center justify-center text-xs text-blue-700 font-medium z-10 overflow-hidden px-2 whitespace-nowrap"
                   :style="{ left: timeToPx(res.start), width: timeToPx(res.end - res.start) }">
                   {{ res.user }} - {{ res.project }}
              </div>

              <!-- Selection Overlay -->
              <div v-if="selectionRange"
                   class="absolute top-0 bottom-0 bg-blue-500/20 border-x-2 border-blue-500 z-20 pointer-events-none transition-all duration-75"
                   :style="{ left: timeToPx(selectionRange.start), width: timeToPx(selectionRange.end - selectionRange.start || 0.01) }">
                   <div class="absolute -top-6 left-1/2 -translate-x-1/2 bg-blue-600 text-white text-[10px] px-1.5 py-0.5 rounded shadow-sm whitespace-nowrap">
                       {{ formatTime(selectionRange.start) }} - {{ formatTime(selectionRange.end) }}
                   </div>
              </div>
         </div>
         <p class="text-xs text-gray-400 mt-2 text-center">拖拽时间轴空白区域进行预约</p>
     </div>

     <!-- Dialog -->
     <Dialog :open="showReservationDialog" @close="showReservationDialog = false" title="新建预约" maxWidth="max-w-sm">
          <div class="space-y-4">
              <div class="grid grid-cols-2 gap-4">
                  <div>
                      <label class="text-sm font-medium mb-1 block">开始时间</label>
                      <Input v-model="newReservationForm.startTime" disabled class="bg-gray-50" />
                  </div>
                  <div>
                      <label class="text-sm font-medium mb-1 block">结束时间</label>
                      <Input v-model="newReservationForm.endTime" disabled class="bg-gray-50" />
                  </div>
              </div>
              <div>
                  <label class="text-sm font-medium mb-1 block">实验项目</label>
                  <Input v-model="newReservationForm.project" placeholder="请输入实验项目名称" />
              </div>
          </div>
          <template #footer>
                <Button variant="ghost" @click="showReservationDialog = false">取消</Button>
                <Button @click="handleReservationSubmit">确认预约</Button>
          </template>
      </Dialog>
  </div>
</template>
