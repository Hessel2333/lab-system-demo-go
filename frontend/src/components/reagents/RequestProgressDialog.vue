<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import axios from 'axios'
import { CheckCircle2, Clock, XCircle, RefreshCw, FileText, Package, ShoppingCart } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Dialog from '@/components/ui/Dialog.vue'
import ItemLifecycleDialog from '@/components/reagents/ItemLifecycleDialog.vue'
import { formatAmount, formatNumber } from '@/lib/quantity'

const props = defineProps<{ open: boolean, request: any }>()
const emit = defineEmits(['close', 'refresh'])

// BPM-B 关联的实体（试剂瓶）
const items = ref<any[]>([])
const isLoadingItems = ref(false)

const arrivedItems = computed(() => items.value.filter(i => i.status === '已到货'))
const storedItems = computed(() => items.value.filter(i => i.status === '在库'))
// consumedItems 不再被显式引用，故不再定义以消除 ts 警告

const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const toast = (msg: string, type: 'success' | 'error' = 'success') => {
    toastMessage.value = msg; toastType.value = type; showToast.value = true
    setTimeout(() => { showToast.value = false }, 3000)
}

const fetchItems = async () => {
    if (!props.request?.id) return
    isLoadingItems.value = true
    try {
        const res = await axios.get(`/api/reagents/items?request_id=${props.request.id}`)
        items.value = res.data
    } catch (e) {
        toast('加载实物信息失败', 'error')
        console.error('Failed to load items', e)
    } finally {
        isLoadingItems.value = false
    }
}

// --- 档案弹窗 ---
const lifecycleDialog = ref({
  isOpen: false,
  itemUuid: null as string | null
})
const openLifecycleDialog = (uuid: string) => {
  lifecycleDialog.value = { isOpen: true, itemUuid: uuid }
}

watch(() => [props.open, props.request?.id], ([open]) => {
    if (open) fetchItems()
}, { immediate: true })

// ── BPM-A 状态机 ──
const bpmASteps = computed(() => {
    if (!props.request) return []
    const status = props.request.status
    const isControlled = props.request.is_controlled

    const fmt = (dateStr: string | null) => {
        if (!dateStr) return ''
        return new Date(dateStr).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
    }

    const step1 = {
        key: 'submitted',
        label: '提交申购申请',
        desc: `${props.request.requestor?.real_name || '研发人员'} 提交了申购请求。`,
        state: 'completed',
        operator: props.request.requestor?.real_name || '申购人',
        time: fmt(props.request.created_at),
    }

    const step2_controlled = isControlled ? (() => {
        let state: string
        if (status === '待审批') state = 'current'
        else if (['待采购', '已接单', '已入库', '已驳回'].includes(status)) state = 'completed'
        else state = 'pending'
        return {
            key: 'leader-approve',
            label: '管控品团队长审批',
            desc: '管控品需经由团队长审批后，方可进入采购流程。',
            state,
            operator: state !== 'pending' ? '团队长' : '',
            time: state === 'completed' ? fmt(props.request.updated_at) : '',
        }
    })() : null

    const step3 = (() => {
        let state: string
        if (status === '待采购' || status === '待审批') state = 'current'
        else if (['已接单', '已入库'].includes(status)) state = 'completed'
        else if (status === '已驳回') state = 'rejected'
        else state = 'pending'
        return {
            key: 'purchasing',
            label: status === '已驳回' ? '申购已驳回' : '采购部接单与下单',
            desc: status === '已驳回'
                ? '本申购单已被驳回，如有需要请重新提交。'
                : '采购部汇总需求，生成易派客采购订单并发向供应商。',
            state,
            operator: ['completed', 'current'].includes(state) ? '采购部' : '',
            time: state === 'completed' ? fmt(props.request.updated_at) : '',
        }
    })()

    const step4 = (() => {
        let state: string
        if (status === '已接单') state = 'completed'
        else if (status === '已驳回') state = 'rejected'
        else state = 'pending'
        return {
            key: 'confirmed',
            label: '采购员确认接单',
            desc: '采购员已确认接单，向供应商下单，申购审批流程闭环。',
            state,
            operator: state === 'completed' ? '采购部' : '',
            time: state === 'completed' ? fmt(props.request.updated_at) : '',
        }
    })()

    type Step = typeof step1
    return [step1, step2_controlled, step3, step4].filter((s): s is Step => s !== null)
})

const stateVariant = (state: string): any => {
    if (state === 'completed') return 'success'
    if (state === 'current') return 'info'
    if (state === 'rejected') return 'destructive'
    return 'default'
}
const stateBadgeLabel = (state: string) => {
    if (state === 'completed') return '已完成'
    if (state === 'current') return '进行中'
    if (state === 'rejected') return '已驳回'
    return '等待中'
}
</script>

<template>
  <Dialog :open="open" size="lg" @close="$emit('close')">
      <template #header>
          <div class="flex flex-col">
              <span class="text-xl font-bold text-gray-900 tracking-tight">申购单进度流转</span>
              <span class="text-xs text-gray-500 font-normal mt-0.5">审批与到货状态追踪</span>
          </div>
      </template>
      <div v-if="request" class="p-6 space-y-6">
          <!-- Request Info Card -->
          <div class="bg-gray-50/50 rounded-2xl p-5 border border-gray-100 shadow-inner">
              <div class="flex justify-between items-start mb-4 border-b border-gray-100 pb-4">
                  <div>
                      <h4 class="text-base font-bold text-gray-900 mb-1">单号 #{{ request.id }}</h4>
                      <div class="flex items-center gap-2 text-xs text-gray-500">
                          <span>试剂: <span class="font-bold text-gray-700">{{ request.reagent_catalog?.name }}</span></span>
                          <span class="text-gray-300">|</span>
                          <span>数量: <span class="font-bold text-gray-700">{{ formatNumber(request.quantity, 0) }} 瓶</span></span>
                      </div>
                  </div>
                  <div class="flex flex-col items-end gap-2">
                       <Badge variant="info" class="h-6 px-3">{{ request.status }}</Badge>
                       <Badge v-if="request.is_controlled" variant="destructive" class="h-5 px-2 text-[10px] font-bold">⚠️ 管控品</Badge>
                  </div>
              </div>
              <div class="grid grid-cols-2 gap-8">
                  <div class="space-y-1.5 flex flex-col">
                      <span class="text-[11px] font-bold text-gray-400 uppercase tracking-wider">需求优先级</span>
                      <span class="text-sm font-semibold" :class="request.request_type === '紧急' ? 'text-red-600' : 'text-gray-700'">{{ request.request_type || '日常采购' }}</span>
                  </div>
                  <div class="space-y-1.5 flex flex-col border-l border-gray-100 pl-8">
                      <span class="text-[11px] font-bold text-gray-400 uppercase tracking-wider">要求交期</span>
                      <span class="text-sm font-semibold text-gray-700">{{ request.expected_delivery || '尽快到货' }}</span>
                  </div>
              </div>
          </div>

          <!-- 申购审批时间线 -->
          <div>
              <div class="flex items-center gap-2 mb-4 px-1">
                  <ShoppingCart class="w-4 h-4 text-indigo-500" />
                  <span class="text-sm font-bold text-gray-800 tracking-tight">申购审批流程</span>
              </div>

              <div class="relative pl-12 space-y-8 before:absolute before:inset-0 before:left-[19px] before:h-full before:w-0.5 before:bg-gray-100">
                  <div v-for="step in bpmASteps" :key="step.key" class="relative flex items-start">
                      <div class="absolute -left-12 flex items-center justify-center w-10 h-10 rounded-full border-4 border-white shrink-0 shadow-sm z-10 transition-all duration-300"
                           :class="[
                               step.state === 'completed' ? 'bg-emerald-500' : 
                               step.state === 'current' ? 'bg-blue-600 shadow-blue-200' : 
                               step.state === 'rejected' ? 'bg-red-500' : 'bg-gray-200'
                           ]">
                          <CheckCircle2 v-if="step.state === 'completed'" class="w-5 h-5 text-white" />
                          <XCircle v-else-if="step.state === 'rejected'" class="w-5 h-5 text-white" />
                          <Clock v-else-if="step.state === 'current'" class="w-5 h-5 text-white animate-pulse" />
                          <div v-else class="w-2.5 h-2.5 bg-gray-400/50 rounded-full"></div>
                      </div>
                      <div class="flex-1 p-4 rounded-2xl border transition-all duration-300 bg-white"
                           :class="step.state === 'pending' ? 'opacity-40 grayscale' : 'border-gray-100 shadow-sm hover:shadow-md hover:border-blue-100'">
                          <div class="flex items-center justify-between mb-1.5">
                              <div class="font-bold text-gray-900 text-sm tracking-tight">{{ step.label }}</div>
                              <Badge :variant="stateVariant(step.state)" class="h-5 text-[10px] px-2 font-bold">{{ stateBadgeLabel(step.state) }}</Badge>
                          </div>
                          <p class="text-xs text-gray-500 leading-relaxed font-bold opacity-80">{{ step.desc }}</p>
                          <div v-if="step.operator || step.time" class="flex items-center gap-3 mt-3">
                              <span v-if="step.time" class="flex items-center gap-1 text-[10px] text-gray-400 font-mono">
                                  <Clock class="w-3 h-3" /> {{ step.time }}
                              </span>
                              <span v-if="step.operator" class="text-[10px] px-2 py-0.5 bg-indigo-50 text-indigo-700 rounded-full font-bold border border-indigo-100/50">
                                  👤 {{ step.operator }}
                              </span>
                          </div>
                      </div>
                  </div>
              </div>
          </div>

          <!-- 关联实体 -->
          <div class="border-t border-dashed border-gray-200 pt-5">
              <div class="flex items-center justify-between mb-4">
                  <div class="flex items-center gap-2 px-1">
                      <Package class="w-4 h-4 text-teal-500" />
                      <span class="text-sm font-bold text-gray-800">到货实物追踪</span>
                  </div>
                  <Button variant="ghost" size="sm" @click="fetchItems" class="h-7 w-7 p-0 rounded-full">
                      <RefreshCw class="w-3.5 h-3.5 text-gray-400" />
                  </Button>
              </div>

              <div v-if="isLoadingItems" class="p-8 text-center text-gray-400">加载中...</div>

              <div v-else-if="items.length === 0" class="rounded-2xl border border-dashed border-gray-200 bg-gray-50/50 px-6 py-8 text-center">
                  <Package class="w-10 h-10 text-gray-200 mx-auto mb-3" />
                  <p class="text-xs text-gray-400 font-bold uppercase tracking-wider">No associated items</p>
                  <p class="text-xs text-gray-400 mt-2 max-w-[280px] mx-auto leading-relaxed">等待采购单产生实物条码后，系统将自动汇总关联信息。</p>
              </div>

              <div v-else class="space-y-4">
                  <!-- 待入库（只读） -->
                  <div v-if="arrivedItems.length > 0" class="rounded-2xl border border-blue-100 bg-blue-50/30 p-4 space-y-3">
                      <div class="flex items-center justify-between">
                          <span class="text-xs font-bold text-blue-800">待领入库瓶数: {{ arrivedItems.length }}</span>
                          <div class="flex flex-wrap gap-1 justify-end max-w-[50%]">
                              <span v-for="item in arrivedItems" :key="item.uuid"
                                    class="font-mono text-[9px] bg-white border border-blue-100 text-blue-600 px-1.5 py-0.5 rounded font-bold">
                                  #{{ item.uuid.substring(0, 8).toUpperCase() }}
                              </span>
                          </div>
                      </div>
                      <p class="text-xs text-blue-700">
                        入库操作请在「到货台账/扫码页」执行，此窗口仅用于流程追踪。
                      </p>
                  </div>

                  <!-- 在库 -->
                  <div v-if="storedItems.length > 0" class="rounded-2xl border border-green-100 overflow-hidden shadow-sm">
                      <div class="bg-green-50/50 px-4 py-2 text-[10px] font-bold text-green-700 border-b border-green-100 uppercase tracking-widest">目前在库资产</div>
                      <div v-for="item in storedItems" :key="item.uuid"
                           class="flex items-center justify-between px-4 py-3 bg-white border-b border-gray-50 last:border-0 hover:bg-green-50/30 transition-colors">
                          <div class="flex items-center gap-4">
                              <button @click="openLifecycleDialog(item.uuid)" class="font-mono text-[11px] font-bold text-blue-600 hover:text-blue-800 hover:underline flex items-center gap-1.5" title="详情档案">
                                <FileText class="w-3.5 h-3.5"/> #{{ item.uuid.substring(0,8).toUpperCase() }}
                              </button>
                              <span class="flex items-center gap-1 text-xs text-gray-500 font-bold">
                                  <MapPin class="w-3.5 h-3.5 text-gray-300" /> {{ item.location }}
                              </span>
                          </div>
                          <Badge variant="success" class="h-5 text-[10px] font-bold">{{ formatAmount(item.remaining_volume, item.reagent_catalog?.unit, 'ml') }}</Badge>
                      </div>
                  </div>
              </div>
          </div>
      </div>
      <template #footer>
          <div class="flex justify-between items-center w-full">
              <span class="text-[10px] text-gray-400 font-bold uppercase tracking-widest opacity-60">Ready for full lifecycle tracking</span>
              <Button variant="secondary" @click="$emit('close')">确定</Button>
          </div>
      </template>
  </Dialog>

  <!-- 全生命周期悬浮窗 -->
  <ItemLifecycleDialog 
    :is-open="lifecycleDialog.isOpen" 
    :item-uuid="lifecycleDialog.itemUuid"
    @close="lifecycleDialog.isOpen = false"
    @refresh-needed="fetchItems"
  />
</template>
