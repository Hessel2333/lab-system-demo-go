<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, CheckSquare, Clock, CreditCard, Truck } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import ProcurementDialog from './ProcurementDialog.vue'
import ArrivalWizard from './ArrivalWizard.vue'
import { fetchInstruments, createInstrument, updateInstrumentStatus, type Instrument } from '@/api/instruments'

const requestOpen = ref(false)
const arrivalOpen = ref(false)
const selectedItem = ref<Instrument | null>(null)
const items = ref<Instrument[]>([])

const loadData = async () => {
    try {
        const all = await fetchInstruments()
        // Filter only those in procurement/arrival workflow
        items.value = all.filter(i => 
            ['planning', 'procurement', 'arrival', 'active'].includes(i.lifecycle_stage)
        )
    } catch (e) {
        console.error(e)
    }
}

onMounted(loadData)

const columns = [
    { id: 'planning', label: '申请中', icon: Clock, color: 'text-gray-500', bg: 'bg-gray-50/50' },
    { id: 'procurement', label: '采购中', icon: CreditCard, color: 'text-orange-500', bg: 'bg-orange-50/30' },
    { id: 'arrival', label: '已到货', icon: Truck, color: 'text-green-500', bg: 'bg-green-50/30' },
]

const getItemsByStatus = (stage: string) => items.value.filter(i => i.lifecycle_stage === stage)

const handleRequestSubmit = async (data: any) => {
    try {
        await createInstrument({
            name: data.name,
            model: data.specs,
            budget: parseFloat(data.budget),
            application_reason: data.reason,
            lifecycle_stage: 'planning',
            status: 'planning'
        })
        requestOpen.value = false
        loadData()
    } catch (e) {
        console.error(e)
    }
}

const openArrival = (item: Instrument) => {
    selectedItem.value = item
    arrivalOpen.value = true
}

const handleArrivalComplete = async () => {
    if (!selectedItem.value) return
    try {
        // Update to arrival/active
        await updateInstrumentStatus(selectedItem.value.ID, 'active') 
        // Note: Real world might have more fields update (SN, Assets No)
        arrivalOpen.value = false
        loadData()
    } catch (e) {
        console.error(e)
    }
}
</script>

<template>
  <div class="space-y-4">
      <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold">采购看板</h2>
          <Button @click="requestOpen = true"><Plus class="w-4 h-4 mr-2" /> 发起采购</Button>
      </div>
      
      <div class="grid grid-cols-3 gap-4 h-[calc(100vh-240px)] min-h-[500px]">
          <div v-for="col in columns" :key="col.id" 
               class="rounded-xl p-4 border border-dashed border-gray-200 flex flex-col gap-3 transition-colors"
               :class="col.bg">
              
              <!-- Header -->
              <h3 class="font-medium flex items-center justify-between" :class="col.color">
                  <span class="flex items-center gap-2">
                       <component :is="col.icon" class="w-4 h-4" />
                       {{ col.label }}
                  </span>
                  <span class="text-xs bg-white px-2 py-0.5 rounded-full border shadow-sm text-gray-600">
                      {{ getItemsByStatus(col.id).length }}
                  </span>
              </h3>

              <!-- Cards -->
              <div class="space-y-3 overflow-y-auto flex-1 pr-1">
                  <div v-for="item in getItemsByStatus(col.id)" :key="item.ID" 
                       class="bg-white p-4 rounded-xl shadow-sm border border-gray-100 cursor-pointer hover:shadow-md hover:border-blue-200 transition-all group relative">
                      
                      <div class="font-medium text-gray-900">{{ item.name }}</div>
                      
                      <div class="flex justify-between items-center mt-2 text-xs text-gray-500">
                          <span>预算: ¥{{ item.budget || 0 }}</span>
                          <span>{{ item.model }}</span>
                      </div>
                      
                      <div v-if="item.application_reason" class="mt-2 text-xs bg-gray-50 text-gray-500 p-2 rounded">
                          {{ item.application_reason }}
                      </div>

                      <!-- Actions for Arrived -->
                      <div v-if="col.id === 'arrival'" class="mt-3">
                          <Button size="sm" class="w-full bg-green-600 hover:bg-green-700 h-8 shadow-sm" @click="openArrival(item)">
                              <CheckSquare class="w-3 h-3 mr-1" /> 验收
                          </Button>
                      </div>
                  </div>
              </div>
          </div>
      </div>

      <!-- Dialogs -->
      <ProcurementDialog :open="requestOpen" @close="requestOpen = false" @submit="handleRequestSubmit" />
      <ArrivalWizard :open="arrivalOpen" @close="arrivalOpen = false" @complete="handleArrivalComplete" />
  </div>
</template>
