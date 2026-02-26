<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import axios from 'axios'
import { toast } from 'vue-sonner'
import {
  AlertTriangle, Package, MapPin, 
  CalendarClock, QrCode, ShoppingCart, TestTube2,
  ChevronRight, History
} from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Dialog from '@/components/ui/Dialog.vue'
import { formatAmount, formatNumber, normalizeUnit } from '@/lib/quantity'

const props = defineProps<{
  isOpen: boolean
  itemUuid: string | null
  userRole?: 'researcher' | 'procurement' | 'leader'
}>()

const emit = defineEmits(['close', 'refresh-needed'])

const loading = ref(false)
const itemData = ref<any>(null)
const actionProcessing = ref(false)
const consumeVolume = ref(1)
const consumeRemarks = ref('')

const consumeUnit = computed(() => {
  return normalizeUnit(itemData.value?.reagent_catalog?.unit, 'ml')
})

watch(() => props.isOpen, (newVal) => {
  if (newVal && props.itemUuid) {
    fetchItemDetails()
  } else {
    itemData.value = null
  }
})

const fetchItemDetails = async () => {
  loading.value = true
  try {
    const res = await axios.get(`/api/reagents/items/${props.itemUuid}`)
    itemData.value = res.data
    const remaining = Number(itemData.value?.remaining_volume || 0)
    consumeVolume.value = remaining > 0 ? Math.min(1, remaining) : 0
    consumeRemarks.value = ''
  } catch (e: any) {
    console.error(e)
    toast.error('无法提取物资档案')
    emit('close')
  } finally {
    loading.value = false
  }
}

const getStatusVariant = (status: string): any => {
  const map: Record<string, string> = {
    '在库': 'success',
    '已到货': 'info',
    '待收货': 'warning',
    '部分收货': 'warning',
    '已耗尽': 'default',
  }
  return map[status] || 'default'
}

const getLogDotColor = (action: string) => {
  const map: Record<string, string> = {
    '扫码入库': 'bg-emerald-500',
    '入库登记': 'bg-emerald-500',
    '确认领回': 'bg-blue-500',
    '扫码领用': 'bg-indigo-500',
    '领用消耗': 'bg-indigo-500',
    '空瓶核销': 'bg-orange-500',
  }
  return map[action] || 'bg-gray-400'
}

const handleAction = async (actionType: 'consume' | 'dispose' | 'receive') => {
  if (!itemData.value) return
  actionProcessing.value = true
  
  try {
    if (actionType === 'receive') {
      await axios.put(`/api/reagents/items/${itemData.value.uuid}/status`, {
        status: '在库',
        location: itemData.value.location || '已领回本人存放点',
        cabinet_id: itemData.value.cabinet_id
      })
      toast.success('已确认领回并入库')
    } else if (actionType === 'consume') {
      const remaining = Number(itemData.value.remaining_volume || 0)
      if (consumeVolume.value <= 0) {
        toast.error('消耗量必须大于 0')
        actionProcessing.value = false
        return
      }
      if (consumeVolume.value > remaining) {
        toast.error('消耗量不能超过当前余量')
        actionProcessing.value = false
        return
      }
      await axios.put(`/api/reagents/items/${itemData.value.uuid}/consume`, {
        consume_volume: Number(consumeVolume.value),
        remarks: consumeRemarks.value || `详情页登记消耗 ${formatAmount(consumeVolume.value, consumeUnit.value)}`
      })
      toast.success('已登记试剂消耗')
    } else if (actionType === 'dispose') {
      if (!confirm('确认该试剂已经彻底使用完毕并需空瓶核销吗？')) {
        actionProcessing.value = false; return;
      }
      await axios.post(`/api/reagents/items/${itemData.value.uuid}/deplete`, {
        remarks: '试剂详情执行耗尽核销'
      })
      toast.success('已空瓶核销该试剂')
    }
    
    await fetchItemDetails()
    emit('refresh-needed')

    } catch (e: any) {
    toast.error('操作失败: ' + (e.response?.data?.error || '服务器错误'))
  } finally {
    actionProcessing.value = false
  }
}
</script>

<template>
  <Dialog :open="isOpen" size="lg" @close="emit('close')">
    <template #header>
      <div class="flex flex-col">
        <span class="text-xl font-bold text-gray-900 tracking-tight flex items-center gap-2">
          <History class="w-5 h-5 text-blue-600" />
          试剂个体防伪档案
        </span>
        <span class="text-xs text-gray-400 font-mono mt-0.5 uppercase tracking-wider">UUID: {{ itemUuid }}</span>
      </div>
    </template>

    <div v-if="loading" class="p-12">
       <div class="flex justify-center items-center h-48">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
       </div>
    </div>

    <!-- Content -->
    <div v-else-if="itemData" class="p-6 space-y-6">
      <div class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
         <div class="p-6 flex gap-8">
           <div class="flex-1">
             <div class="flex items-center gap-3 mb-3">
                 <Badge :variant="getStatusVariant(itemData.status)" class="px-3 h-6 font-bold uppercase tracking-widest">
                   {{ itemData.status }}
                 </Badge>
                 <Badge v-if="itemData.reagent_catalog?.is_controlled" variant="destructive" class="px-2 h-5 text-[10px] font-bold">
                   管控品
                 </Badge>
             </div>
             <h1 class="text-2xl font-bold text-gray-900 tracking-tight">{{ itemData.reagent_catalog?.name }}</h1>
             <p class="text-sm font-mono text-gray-500 mt-1">CAS 号: {{ itemData.reagent_catalog?.cas_number || '未知' }}</p>
             
             <div class="grid grid-cols-2 gap-4 mt-6">
               <div>
                  <label class="text-[10px] text-gray-400 uppercase font-bold tracking-widest px-0.5">剩余余量</label>
                  <div class="flex items-end gap-1.5 mt-1">
                    <span class="text-2xl font-bold font-mono tracking-tight" :class="itemData.remaining_volume > 0 ? 'text-blue-600' : 'text-red-500'">{{ formatNumber(itemData.remaining_volume) }}</span>
                    <span class="text-sm text-gray-400 font-medium mb-1">/ {{ formatNumber(itemData.capacity) }} {{ normalizeUnit(itemData.reagent_catalog?.unit, 'ml') }}</span>
                  </div>
               </div>
               <div>
                  <label class="text-[10px] text-gray-400 uppercase font-bold tracking-widest px-0.5">批次来源</label>
                  <p class="text-sm font-mono font-bold text-gray-700 mt-1 h-8 flex items-center">{{ itemData.batch_number }}</p>
               </div>
             </div>
           </div>

           <div class="w-28 shrink-0 flex flex-col items-center justify-center bg-gray-50 rounded-2xl border border-gray-100 p-4">
             <QrCode class="w-12 h-12 text-gray-800" stroke-width="1.5" />
             <span class="text-[10px] text-gray-400 font-mono mt-3 tracking-widest font-bold">{{ itemUuid?.substring(0,8).toUpperCase() }}</span>
             <span class="text-[9px] text-gray-400 mt-1 opacity-60">扫码防伪溯源</span>
           </div>
         </div>
      </div>

      <div class="grid grid-cols-2 gap-4">
         <div class="bg-gray-50/50 rounded-2xl border border-gray-100/60 p-5">
            <h3 class="text-[11px] font-bold text-gray-400 uppercase tracking-widest mb-4 flex items-center gap-2"><MapPin class="h-4 w-4 text-cyan-500"/> 库位坐标记录</h3>
            <div class="space-y-3 text-sm">
              <div class="flex items-center justify-between"><span class="text-gray-500">实物位置</span> <span class="font-bold text-gray-800">{{ itemData.location || '--' }}</span></div>
              <div class="flex items-center justify-between"><span class="text-gray-500">归属柜位</span> <span class="font-bold text-gray-800">{{ itemData.cabinet?.name || '公用区' }}</span></div>
            </div>
         </div>
         
         <div class="bg-gray-50/50 rounded-2xl border border-gray-100/60 p-5">
            <h3 class="text-[11px] font-bold text-gray-400 uppercase tracking-widest mb-4 flex items-center gap-2"><CalendarClock class="h-4 w-4 text-amber-500"/> 存储有效期档案</h3>
            <div class="space-y-3 text-sm">
              <div class="flex items-center justify-between"><span class="text-gray-500">存储要求</span> <span class="font-bold text-gray-800 text-xs">{{ itemData.reagent_catalog?.storage_condition || '常温避光' }}</span></div>
              <div class="flex items-center justify-between"><span class="text-gray-500">失效日期</span> <span :class="itemData.expiry_date && new Date(itemData.expiry_date) < new Date() ? 'text-red-600 font-bold' : 'font-bold text-gray-800'">{{ itemData.expiry_date ? new Date(itemData.expiry_date).toLocaleDateString() : '长期有效' }}</span></div>
            </div>
         </div>
      </div>

      <div class="bg-blue-50/30 rounded-2xl border border-blue-100/50 p-5" v-if="itemData.reagent_request">
         <h3 class="text-[11px] font-bold text-blue-800 uppercase tracking-widest mb-4 flex items-center gap-2"><ShoppingCart class="h-4 w-4"/> 原始申购记录溯源</h3>
         <div class="flex justify-between items-center bg-white p-4 rounded-xl border border-blue-100 shadow-sm transition-hover hover:border-blue-300">
           <div>
              <p class="text-sm text-gray-900 font-bold">申购人：{{ itemData.reagent_request?.requestor?.real_name || 'System' }}</p>
              <p class="text-xs text-gray-500 mt-1 line-clamp-1 italic text-blue-900/40">"{{ itemData.reagent_request?.remarks || '日常实验储备需求' }}"</p>
           </div>
           <ChevronRight class="h-5 w-5 text-blue-200" />
         </div>
      </div>

      <div>
        <h3 class="text-sm font-bold text-gray-900 mb-6 flex items-center gap-2 px-1">
          <History class="h-4 w-4 text-indigo-500" />
          全生命周期操作流转
        </h3>
        
        <div class="relative pl-6 space-y-8 before:absolute before:inset-0 before:left-[11px] before:h-full before:w-0.5 before:bg-gray-100">
           <div v-if="!itemData.logs || itemData.logs.length === 0" class="text-xs text-gray-400 py-4 font-medium italic">暂无流转历史...</div>
           <div v-else>
               <div v-for="log in itemData.logs" :key="log.id" class="relative pl-10 mb-8 last:mb-0">
                  <div class="absolute -left-[19px] top-0 flex items-center justify-center w-5 h-5 rounded-full bg-white border-2 border-gray-200 z-10" v-if="log">
                    <div class="w-1.5 h-1.5 rounded-full" :class="getLogDotColor(log.action)"></div>
                  </div>
                  <div class="bg-gray-50/80 rounded-2xl p-4 border border-gray-100 hover:bg-white hover:shadow-md transition-all duration-300">
                    <div class="flex items-center justify-between mb-2">
                      <span class="font-bold text-sm text-gray-900">{{ log.action || '系统操作' }}</span>
                      <span class="text-[10px] text-gray-400 font-mono">{{ log.created_at ? new Date(log.created_at).toLocaleString('zh-CN', { hour12: false }) : '--' }}</span>
                    </div>
                    <div class="text-xs text-gray-500 font-medium flex gap-4">
                       <span>👤 {{ log.user?.real_name || 'System' }}</span>
                       <span v-if="log.quantity" class="text-blue-600 font-bold">数量变动: {{ formatNumber(log.quantity) }}</span>
                    </div>
                    <p v-if="log.remarks" class="text-[11px] text-gray-400 mt-3 p-3 bg-white/60 rounded-xl border border-dashed border-gray-200 italic">"{{ log.remarks }}"</p>
                  </div>
               </div>
           </div>
        </div>
      </div>
    </div>

    <template #footer>
       <div v-if="itemData" class="flex flex-wrap gap-3 w-full items-center">
          <Button variant="secondary" @click="emit('close')">关闭详情</Button>
          
          <div v-if="itemData.status === '已耗尽'" class="flex-1 flex items-center justify-center text-xs font-bold text-gray-400 uppercase tracking-widest bg-gray-50 rounded-xl">
             Lifecycle Terminated
          </div>
          
          <template v-else>
            <Button v-if="itemData.status === '已到货'" class="flex-1" variant="primary" @click="handleAction('receive')" :disabled="actionProcessing">
              <Package class="w-4 h-4 mr-2" /> 确认领回入库
            </Button>

            <template v-else-if="itemData.status === '在库'">
              <div class="flex-1 min-w-[260px] rounded-xl border border-gray-200 bg-gray-50 p-2.5">
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-2 items-center">
                  <div class="sm:col-span-1">
                    <label class="block text-[10px] text-gray-500 mb-1">本次消耗量</label>
                    <div class="relative">
                      <input
                        v-model.number="consumeVolume"
                        type="number"
                        min="0"
                        :max="itemData.remaining_volume"
                        step="0.1"
                        class="w-full h-9 rounded-lg border border-gray-200 bg-white px-2 pr-9 text-sm"
                      />
                      <span class="absolute right-2 top-2 text-xs text-gray-400">{{ consumeUnit }}</span>
                    </div>
                  </div>
                  <div class="sm:col-span-2">
                    <label class="block text-[10px] text-gray-500 mb-1">用途/备注（选填）</label>
                    <input
                      v-model="consumeRemarks"
                      type="text"
                      class="w-full h-9 rounded-lg border border-gray-200 bg-white px-2 text-sm"
                      placeholder="如：滴定实验A组"
                    />
                  </div>
                </div>
              </div>
              <Button class="flex-1" variant="outline" @click="handleAction('dispose')" :disabled="actionProcessing">
                <AlertTriangle class="w-4 h-4 mr-2 text-orange-500" /> 空瓶核销
              </Button>
              <Button class="flex-1 shadow-blue-100 shadow-lg" variant="primary" @click="handleAction('consume')" :disabled="actionProcessing || consumeVolume <= 0">
                <TestTube2 class="w-4 h-4 mr-2" /> 确认消耗
              </Button>
            </template>
          </template>
       </div>
    </template>
  </Dialog>
</template>
