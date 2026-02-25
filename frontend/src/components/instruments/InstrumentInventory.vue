<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import StatusUpdateDialog from './StatusUpdateDialog.vue'
import InstrumentCreationDialog from './InstrumentCreationDialog.vue'
import InstrumentDetailDialog from './InstrumentDetailDialog.vue'
import { fetchInstruments, type Instrument } from '@/api/instruments'
import { Search, RefreshCcw, Plus } from 'lucide-vue-next'

// State
const instruments = ref<Instrument[]>([])
const loading = ref(false)
const searchQuery = ref('')

const detailOpen = ref(false)
const statusOpen = ref(false)
const creationOpen = ref(false)
const selectedInstrument = ref<Instrument | null>(null)

const loadData = async () => {
    loading.value = true
    try {
        instruments.value = await fetchInstruments()
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

onMounted(loadData)

const filteredInstruments = computed(() => {
    return instruments.value.filter(i => 
        i.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
        i.model.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
})

const openDetail = (ins: Instrument) => {
    selectedInstrument.value = ins
    detailOpen.value = true
    console.log('Opening detail dialog for:', ins)
}

const openStatus = (ins: Instrument) => {
    selectedInstrument.value = ins
    statusOpen.value = true
}

const handleStatusUpdate = () => {
    statusOpen.value = false
    loadData() // Refresh list
}

const statusMap: Record<string, { label: string, variant: 'success' | 'warning' | 'destructive' | 'secondary' | 'default' }> = {
  active: { label: '空闲/可用', variant: 'success' }, // Green
  in_use: { label: '使用中', variant: 'default' }, // Blue (Primary)
  maintenance: { label: '维护中', variant: 'warning' }, // Orange
  fault: { label: '故障/维修', variant: 'destructive' }, // Red
  arrival: { label: '待安装', variant: 'secondary' }, // Gray
  retired: { label: '已报废', variant: 'destructive' }, // Red
  planning: { label: '规划中', variant: 'secondary' },
  procurement: { label: '采购中', variant: 'secondary' },
}

</script>

<template>
  <div class="space-y-4">
    <!-- Toolbar -->
    <div class="flex items-center justify-between">
       <div class="flex items-center gap-2 w-full max-w-sm">
           <div class="relative w-full">
               <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
               <input 
                  v-model="searchQuery"
                  type="text" 
                  placeholder="搜索仪器名称、型号..." 
                  class="flex h-10 w-full rounded-lg border border-input bg-background pl-9 pr-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
               />
           </div>
           <Button variant="outline" size="icon" @click="loadData"><RefreshCcw class="h-4 w-4" :class="{'animate-spin': loading}" /></Button>
       </div>
       
       <Button @click="creationOpen = true" class="shadow-sm">
           <Plus class="w-4 h-4 mr-2" /> 新增仪器
       </Button>
    </div>

    <!-- Table -->
    <div class="rounded-xl border bg-white shadow-sm overflow-hidden">
        <table class="w-full text-sm text-left">
            <thead class="bg-gray-50/75 border-b text-gray-500">
                <tr>
                    <th class="px-6 py-4 font-medium">状态</th>
                    <th class="px-6 py-4 font-medium">仪器信息</th>
                    <th class="px-6 py-4 font-medium">位置</th>
                    <th class="px-6 py-4 font-medium">利用率</th>
                    <th class="px-6 py-4 font-medium">健康度</th>
                    <th class="px-6 py-4 font-medium text-right">操作</th>
                </tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
                <tr v-for="ins in filteredInstruments" :key="ins.ID" class="hover:bg-gray-50/50 transition-colors">
                    <td class="px-6 py-4">
                        <Badge :variant="statusMap[ins.status]?.variant || 'secondary'">{{ statusMap[ins.status]?.label || ins.status }}</Badge>
                    </td>
                    <td class="px-6 py-4">
                        <div class="font-medium text-gray-900">{{ ins.name }}</div>
                        <div class="text-xs text-gray-500">{{ ins.model }} (ID: {{ ins.ID }})</div>
                    </td>
                    <td class="px-6 py-4 text-gray-600">{{ ins.location || '待定' }}</td>
                    <td class="px-6 py-4">
                        <div class="flex items-center gap-2">
                             <div class="h-2 w-24 rounded-full bg-gray-100 overflow-hidden">
                                 <div class="h-full bg-blue-500" :style="{ width: Math.min(ins.run_time / 20, 100) + '%' }"></div>
                             </div>
                             <span class="text-xs text-gray-500">{{ ins.run_time }}h</span>
                        </div>
                    </td>
                    <td class="px-6 py-4 text-gray-600">{{ ins.health }}%</td>
                    <td class="px-6 py-4 text-right space-x-2">
                        <Button variant="outline" size="sm" @click="openDetail(ins)">详情</Button>
                        <Button variant="ghost" size="sm" class="text-blue-600" @click="openStatus(ins)">状态</Button>
                    </td>
                </tr>
            </tbody>
        </table>
        
        <div v-if="filteredInstruments.length === 0 && !loading" class="p-8 text-center text-gray-500">
            暂无数据
        </div>
    </div>

    <!-- Dialogs -->
    <InstrumentDetailDialog 
        :open="detailOpen" 
        :instrument-data="selectedInstrument" 
        @close="detailOpen = false" 
        @refresh="loadData"
    />
    <StatusUpdateDialog 
        :open="statusOpen" 
        :instrument-id="selectedInstrument?.ID"
        :current-status="statusMap[selectedInstrument?.status || '']?.label || '未知'" 
        @close="statusOpen = false" 
        @update="handleStatusUpdate" 
    />
    <InstrumentCreationDialog
        :open="creationOpen"
        @close="creationOpen = false"
        @success="loadData"
    />
  </div>
</template>
