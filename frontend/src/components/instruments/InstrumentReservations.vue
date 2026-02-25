<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick } from 'vue'
import { fetchInstruments, type Instrument } from '@/api/instruments'
import InstrumentReservationCalendar from './InstrumentReservationCalendar.vue'

const instruments = ref<Instrument[]>([])
const selectedInstrumentId = ref<string>('')
const loading = ref(false)
const schedulerContainer = ref<HTMLElement | null>(null)

const loadInstruments = async () => {
    loading.value = true
    try {
        instruments.value = await fetchInstruments()
        if (instruments.value && instruments.value.length > 0) {
           const first = instruments.value[0]
           if (first) {
               selectedInstrumentId.value = first.ID.toString()
           }
        }
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

onMounted(loadInstruments)

const currentInstrument = computed(() => {
    return instruments.value.find(i => i.ID.toString() === selectedInstrumentId.value)
})

watch(selectedInstrumentId, async () => {
    await nextTick()
    if (schedulerContainer.value) {
        schedulerContainer.value.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
})



</script>

<template>
  <div class="h-full flex flex-col space-y-6">
      <!-- Toolbar -->
      <div class="flex items-center justify-between bg-white p-4 rounded-xl border shadow-sm">
          <div class="flex items-center gap-4">
              <span class="font-medium text-gray-700">选择仪器:</span>
              <select v-model="selectedInstrumentId" class="border rounded-md px-3 py-1.5 min-w-[200px] text-sm bg-gray-50 focus:ring-2 focus:ring-blue-500 outline-none">
                  <option v-for="ins in instruments" :key="ins.ID" :value="ins.ID.toString()">
                      {{ ins.name }} ({{ ins.model }})
                  </option>
              </select>
          </div>
          
          <div v-if="currentInstrument" class="flex gap-4 text-sm text-gray-500">
             <span>位置: <span class="text-gray-900">{{ currentInstrument.location || '未知' }}</span></span>
             <span>状态: <span class="capitalize" :class="{'text-green-600': currentInstrument.status === 'active'}">{{ currentInstrument.status }}</span></span>
          </div>
      </div>

      <!-- Scheduler Area -->
      <div ref="schedulerContainer" class="flex-1 bg-white rounded-xl border shadow-sm p-6 flex flex-col">
          <div v-if="currentInstrument" class="h-full flex flex-col">
              <h2 class="text-lg font-bold mb-6 flex items-center gap-2 flex-shrink-0">
                  {{ currentInstrument.name }} 预约表
              </h2>
              <div class="flex-1 min-h-0">
                  <InstrumentReservationCalendar :instrument-id="selectedInstrumentId" />
              </div>
          </div>
          <div v-else class="h-full flex items-center justify-center text-gray-400">
              <div v-if="loading">加载仪器列表...</div>
              <div v-else>请选择一台仪器以查看预约</div>
          </div>
      </div>
  </div>
</template>
