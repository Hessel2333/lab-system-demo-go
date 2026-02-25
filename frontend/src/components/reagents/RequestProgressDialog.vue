<script setup lang="ts">
import Dialog from '@/components/ui/Dialog.vue'
import { CheckCircle2, Clock, Truck, Package, PackageCheck, MapPin, Trash2, RefreshCw } from 'lucide-vue-next'
import { ref, watch, computed } from 'vue'
import axios from 'axios'
import Input from '@/components/ui/Input.vue'

const props = defineProps<{ open: boolean, request: any }>()
const emit = defineEmits(['close', 'refresh'])

const items = ref<any[]>([])
const isLoadingItems = ref(false)
const batchLocation = ref('')
const quickLocations = ['E309', 'E307', 'F103', 'F309', 'B201']

const arrivedItems = computed(() => items.value.filter(i => i.status === '已到货'))
const storedItems = computed(() => items.value.filter(i => i.status === '在库'))
const consumedItems = computed(() => items.value.filter(i => i.status === '已耗尽'))

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
    } catch (e) { console.error('Failed to load items', e) }
    finally { isLoadingItems.value = false }
}

watch(() => [props.open, props.request?.id], ([open]) => {
    if (open) { batchLocation.value = ''; fetchItems() }
}, { immediate: true })

const storeAllItems = async () => {
    if (!batchLocation.value) { toast('请先选择或输入库位', 'error'); return }
    try {
        await Promise.all(arrivedItems.value.map(item =>
            axios.put(`/api/reagents/items/${item.uuid}/status`, { status: '在库', location: batchLocation.value })
        ))
        toast(`${arrivedItems.value.length} 瓶已全部入库至 ${batchLocation.value}`)
        batchLocation.value = ''; fetchItems(); emit('refresh')
    } catch (e) { toast('入库操作失败', 'error') }
}

const consumeItem = async (item: any) => {
    try {
        await axios.put(`/api/reagents/items/${item.uuid}/status`, { status: '已耗尽' })
        toast('已标记为空瓶核销'); fetchItems(); emit('refresh')
    } catch (e) { toast('操作失败', 'error') }
}

const getStepStatus = (stepName: string) => {
    if (!props.request) return 'pending'
    const statusOrder = ['待提交', '待处理', '采购中', '已到货', '已入库']
    const stepOrderMap: Record<string, number> = { 'submitted': 1, 'approved': 2, 'arrived': 3, 'instorage': 4 }
    const currentIdx = statusOrder.indexOf(props.request.status)
    const stepIdx = stepOrderMap[stepName] || 0
    if (currentIdx >= stepIdx) return 'completed'
    if (currentIdx + 1 === stepIdx) return 'current'
    return 'pending'
}

const getTimelineLabel = (stepStatus: string, defaultLabel: string) => {
    if (stepStatus === 'completed') return '已完成'
    if (stepStatus === 'current') return '进行中'
    return defaultLabel
}

const getStepTime = (stepName: string) => {
    if (!props.request) return null
    if (stepName === 'submitted') return props.request.created_at
    // 对于已完成的步骤使用 updated_at 作为近似时间
    return getStepStatus(stepName) === 'completed' ? props.request.updated_at : null
}

const getStepOperator = (stepName: string) => {
    if (!props.request) return ''
    const status = getStepStatus(stepName)
    if (status === 'pending') return ''
    switch (stepName) {
        case 'submitted': return props.request.requestor?.real_name || '申购人'
        case 'approved': return '采购部'
        case 'arrived': return '供应商 → 仓库'
        case 'instorage': return '库管员'
        default: return ''
    }
}

const formatDateTime = (dateStr: string | null) => {
    if (!dateStr) return ''
    return new Date(dateStr).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <Dialog :open="open" @close="$emit('close')" title="申购单流转进度" class="max-w-2xl px-2">
      <div v-if="request" class="py-4 space-y-8">

          <!-- Request Info Summary -->
          <div class="bg-gray-50 rounded-xl p-4 border border-gray-100 mb-6">
              <div class="flex justify-between items-start mb-3 border-b border-gray-100 pb-3">
                  <div>
                      <h4 class="text-sm font-bold text-gray-900 mb-1">申购单号: #{{ request.id }}</h4>
                      <p class="text-xs text-gray-500">试剂名称: <span class="font-medium text-gray-700">{{ request.reagent_catalog?.name }}</span></p>
                      <p class="text-xs text-gray-500 mt-1">申购数量: <span class="font-medium text-gray-700">{{ request.quantity }} {{ request.reagent_catalog?.unit || '瓶' }}</span></p>
                      <p class="text-xs text-gray-500 mt-1">申购人: <span class="font-medium text-gray-700">{{ request.requestor?.real_name || 'System' }}</span></p>
                  </div>
                  <div class="text-right">
                      <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          当前状态: {{ request.status }}
                      </span>
                  </div>
              </div>

              <!-- Extended Fields Summary -->
              <div class="grid grid-cols-2 gap-4 text-xs mt-3">
                  <div class="space-y-1.5 flex flex-col">
                      <span class="text-gray-400">需求类型</span>
                      <span class="font-medium" :class="request.request_type === '紧急' ? 'text-red-600' : 'text-gray-900'">
                          {{ request.request_type || '日常' }}
                      </span>
                  </div>
                  <div class="space-y-1.5 flex flex-col">
                      <span class="text-gray-400">要求交期</span>
                      <span class="font-medium text-gray-900">{{ request.expected_delivery || '未指定' }}</span>
                  </div>
                  <div class="space-y-1.5 flex flex-col col-span-2 mt-1" v-if="request.project_name">
                      <span class="text-gray-400">所属项目</span>
                      <div class="flex items-center gap-2">
                          <span class="font-medium text-gray-900">{{ request.project_name }}</span>
                          <span v-if="request.project_id" class="px-2 py-[2px] bg-gray-200 text-gray-600 rounded text-[10px] font-mono">{{ request.project_id }}</span>
                      </div>
                  </div>
              </div>
          </div>

          <!-- Vertical Timeline -->
          <div class="relative pl-12 space-y-12 before:absolute before:inset-0 before:left-5 before:h-full before:w-0.5 before:bg-gradient-to-b before:from-transparent before:via-gray-200 before:to-transparent">

              <!-- 1. Submitted -->
              <div class="relative flex items-start group">
                  <div class="absolute -left-12 flex items-center justify-center w-10 h-10 rounded-full border-4 border-white shrink-0 shadow-sm z-10"
                       :class="getStepStatus('submitted') === 'completed' ? 'bg-green-500' : 'bg-blue-500'">
                      <CheckCircle2 v-if="getStepStatus('submitted') === 'completed'" class="w-5 h-5 text-white" />
                      <Clock v-else class="w-5 h-5 text-white" />
                  </div>
                  <div class="flex-1 p-4 rounded-xl border border-gray-200 bg-white shadow-sm hover:shadow-md transition-shadow">
                      <div class="flex items-center justify-between space-x-2 mb-1">
                          <div class="font-bold text-gray-900 text-sm">提交申购申请</div>
                          <span class="text-xs font-medium text-green-500 bg-green-50 px-2 py-0.5 rounded">已完成</span>
                      </div>
                      <div class="text-xs text-gray-500 leading-relaxed">{{ request.requestor?.real_name || '研发人员' }} 提交了申购请求。</div>
                      <div v-if="getStepTime('submitted')" class="text-[10px] text-gray-400 mt-2 flex items-center gap-1">
                          <Clock class="w-3 h-3" /> {{ formatDateTime(getStepTime('submitted')) }}
                      </div>
                  </div>
              </div>

              <!-- 2. Approved & Purchasing -->
              <div class="relative flex items-start group">
                  <div class="absolute -left-12 flex items-center justify-center w-10 h-10 rounded-full border-4 border-white bg-gray-200 shrink-0 shadow-sm z-10"
                       :class="getStepStatus('approved') === 'completed' ? 'bg-green-500' : (getStepStatus('approved') === 'current' ? 'bg-blue-500 ring-4 ring-blue-50' : 'bg-gray-100')">
                      <CheckCircle2 v-if="getStepStatus('approved') === 'completed'" class="w-5 h-5 text-white" />
                      <Clock v-else-if="getStepStatus('approved') === 'current'" class="w-5 h-5 text-white" />
                      <div v-else class="w-3 h-3 bg-gray-300 rounded-full"></div>
                  </div>
                  <div class="flex-1 p-4 rounded-xl border border-gray-200 bg-white shadow-sm hover:shadow-md transition-shadow" :class="getStepStatus('approved') === 'pending' ? 'opacity-60 grayscale-[0.5]' : ''">
                      <div class="flex items-center justify-between space-x-2 mb-1">
                          <div class="font-bold text-gray-900 text-sm">采购审批</div>
                          <span class="text-xs font-medium px-2 py-0.5 rounded" :class="getStepStatus('approved') === 'completed' ? 'text-green-500 bg-green-50' : (getStepStatus('approved') === 'current' ? 'text-blue-500 bg-blue-50' : 'text-gray-400 bg-gray-50')">
                            {{ getTimelineLabel(getStepStatus('approved'), '待审批') }}
                          </span>
                      </div>
                      <div class="text-xs text-gray-500 leading-relaxed">采购人员审批通过，生成采购订单并发向供应商。</div>
                      <div v-if="getStepStatus('approved') !== 'pending'" class="text-[10px] text-gray-400 mt-2 flex items-center gap-2">
                          <span v-if="getStepTime('approved')" class="flex items-center gap-1"><Clock class="w-3 h-3" /> {{ formatDateTime(getStepTime('approved')) }}</span>
                          <span v-if="getStepOperator('approved')" class="px-1.5 py-0.5 bg-blue-50 text-blue-600 rounded font-medium">👤 {{ getStepOperator('approved') }}</span>
                      </div>
                  </div>
              </div>

              <!-- 3. Logistics & Arrival -->
              <div class="relative flex items-start group">
                  <div class="absolute -left-12 flex items-center justify-center w-10 h-10 rounded-full border-4 border-white bg-gray-200 shrink-0 shadow-sm z-10"
                       :class="getStepStatus('arrived') === 'completed' ? 'bg-green-500' : (getStepStatus('arrived') === 'current' ? 'bg-blue-500 ring-4 ring-blue-50' : 'bg-gray-100')">
                      <CheckCircle2 v-if="getStepStatus('arrived') === 'completed'" class="w-5 h-5 text-white" />
                      <Truck v-else-if="getStepStatus('arrived') === 'current'" class="w-5 h-5 text-white" />
                      <div v-else class="w-3 h-3 bg-gray-300 rounded-full"></div>
                  </div>
                  <div class="flex-1 p-4 rounded-xl border border-gray-200 bg-white shadow-sm hover:shadow-md transition-shadow" :class="getStepStatus('arrived') === 'pending' ? 'opacity-60 grayscale-[0.5]' : ''">
                      <div class="flex items-center justify-between space-x-2 mb-1">
                          <div class="font-bold text-gray-900 text-sm">物流运输与到货</div>
                          <span class="text-xs font-medium px-2 py-0.5 rounded" :class="getStepStatus('arrived') === 'completed' ? 'text-green-500 bg-green-50' : (getStepStatus('arrived') === 'current' ? 'text-blue-500 bg-blue-50' : 'text-gray-400 bg-gray-50')">
                            {{ getTimelineLabel(getStepStatus('arrived'), '等待收货') }}
                          </span>
                      </div>
                      <div class="text-xs text-gray-500 leading-relaxed">供应商发货，单据流转至【分拣区】，由采购人员确认到货并生成条码。</div>
                      <div v-if="getStepStatus('arrived') !== 'pending'" class="text-[10px] text-gray-400 mt-2 flex items-center gap-2">
                          <span v-if="getStepTime('arrived')" class="flex items-center gap-1"><Clock class="w-3 h-3" /> {{ formatDateTime(getStepTime('arrived')) }}</span>
                          <span v-if="getStepOperator('arrived')" class="px-1.5 py-0.5 bg-blue-50 text-blue-600 rounded font-medium">📦 {{ getStepOperator('arrived') }}</span>
                      </div>
                  </div>
              </div>

              <!-- 4. Storage -->
              <div class="relative flex items-start group">
                  <div class="absolute -left-12 flex items-center justify-center w-10 h-10 rounded-full border-4 border-white shrink-0 shadow-sm z-10"
                       :class="getStepStatus('instorage') === 'completed' ? 'bg-green-500' : (getStepStatus('instorage') === 'current' ? 'bg-blue-500 ring-4 ring-blue-50' : 'bg-gray-100 border-gray-200')">
                      <PackageCheck v-if="getStepStatus('instorage') === 'completed'" class="w-5 h-5 text-white" />
                      <Package v-else-if="getStepStatus('instorage') === 'current'" class="w-5 h-5 text-white" />
                      <div v-else class="w-3 h-3 bg-gray-300 rounded-full"></div>
                  </div>
                  <div class="flex-1 rounded-xl border border-gray-200 bg-white shadow-sm overflow-hidden"
                       :class="getStepStatus('instorage') === 'pending' ? 'opacity-60 grayscale-[0.5]' : ''">

                      <!-- Header -->
                      <div class="flex items-center justify-between p-4 border-b border-gray-100">
                          <div class="font-bold text-gray-900 text-sm">扫码入库与库存管理</div>
                          <div class="flex items-center gap-2">
                              <button @click="fetchItems" class="text-gray-400 hover:text-blue-500 transition-colors p-1 rounded">
                                  <RefreshCw class="w-3.5 h-3.5" />
                              </button>
                              <span class="text-xs font-medium px-2 py-0.5 rounded"
                                  :class="getStepStatus('instorage') === 'completed' ? 'text-green-500 bg-green-50' : (getStepStatus('instorage') === 'current' ? 'text-blue-500 bg-blue-50' : 'text-gray-400 bg-gray-50')">
                                {{ getStepStatus('instorage') === 'completed' ? '全部入库' : (getStepStatus('instorage') === 'current' ? `待入库 ${arrivedItems.length}/${items.length} 瓶` : '待操作') }}
                              </span>
                          </div>
                      </div>

                      <!-- Pending state -->
                      <div v-if="getStepStatus('instorage') === 'pending'" class="p-4">
                          <p class="text-xs text-gray-500">等待采购人员确认到货并生成条码后，将在此处显示操作界面。</p>
                      </div>

                      <!-- Active state -->
                      <div v-else class="p-4 space-y-3">
                          <div v-if="isLoadingItems" class="text-xs text-center text-gray-400 py-4">加载中...</div>

                          <!-- ① 待入库：统一批量操作区（N 瓶共用一个库位） -->
                          <div v-if="arrivedItems.length > 0" class="rounded-lg border border-blue-200 bg-blue-50 p-3 space-y-2">
                              <div class="flex items-center justify-between">
                                  <span class="text-xs font-semibold text-blue-800">待入库 · {{ arrivedItems.length }} 瓶</span>
                                  <div class="flex flex-wrap gap-1">
                                      <span v-for="item in arrivedItems" :key="item.uuid"
                                            class="font-mono text-[10px] bg-white border border-blue-200 text-gray-500 px-1.5 py-0.5 rounded">
                                          #{{ item.uuid.substring(0, 8).toUpperCase() }}
                                      </span>
                                  </div>
                              </div>
                              <div class="flex flex-wrap gap-1.5">
                                  <button v-for="loc in quickLocations" :key="loc"
                                          @click="batchLocation = loc"
                                          class="text-xs px-2.5 py-1 rounded-md border transition-all font-medium"
                                          :class="batchLocation === loc ? 'bg-blue-600 text-white border-blue-600' : 'bg-white text-gray-600 border-gray-300 hover:border-blue-400'">
                                      📍 {{ loc }}
                                  </button>
                              </div>
                              <div class="flex gap-2">
                                  <Input v-model="batchLocation" placeholder="或手动输入库位..." class="h-8 text-xs flex-1" />
                                  <button @click="storeAllItems"
                                          :disabled="!batchLocation"
                                          class="px-3 py-1.5 rounded-lg text-xs font-semibold whitespace-nowrap transition-all disabled:opacity-40 disabled:cursor-not-allowed"
                                          :class="batchLocation ? 'bg-blue-600 text-white hover:bg-blue-700' : 'bg-gray-200 text-gray-500'">
                                      ✅ 全部确认入库
                                  </button>
                              </div>
                          </div>

                          <!-- ② 在库：紧凑列表（每瓶一行） -->
                          <div v-if="storedItems.length > 0" class="rounded-lg border border-green-200 overflow-hidden">
                              <div class="bg-green-50 px-3 py-1.5 text-xs font-semibold text-green-800 border-b border-green-200">在库 · {{ storedItems.length }} 瓶</div>
                              <div v-for="item in storedItems" :key="item.uuid"
                                   class="flex items-center justify-between px-3 py-2 bg-white border-b border-gray-100 last:border-0 hover:bg-gray-50">
                                  <div class="flex items-center gap-2">
                                      <span class="font-mono text-[10px] text-gray-400">#{{ item.uuid.substring(0, 8).toUpperCase() }}</span>
                                      <span class="flex items-center gap-0.5 text-xs text-gray-500">
                                          <MapPin class="w-3 h-3" /> {{ item.location }}
                                      </span>
                                      <span class="text-xs text-gray-400">剩余 {{ item.remaining_volume }}{{ item.reagent_catalog?.unit }}</span>
                                  </div>
                                  <button @click="consumeItem(item)"
                                          class="flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-md bg-red-50 text-red-500 border border-red-200 hover:bg-red-100 transition-all">
                                      <Trash2 class="w-2.5 h-2.5" /> 核销
                                  </button>
                              </div>
                          </div>

                          <!-- ③ 已耗尽：折叠摘要行 -->
                          <div v-if="consumedItems.length > 0" class="rounded-lg border border-gray-200 px-3 py-2 flex items-center gap-2">
                              <span class="text-[10px] text-gray-400 shrink-0">已耗尽 {{ consumedItems.length }} 瓶：</span>
                              <div class="flex flex-wrap gap-1">
                                  <span v-for="item in consumedItems" :key="item.uuid"
                                        class="font-mono text-[10px] bg-gray-100 text-gray-400 px-1 py-0.5 rounded line-through">
                                      #{{ item.uuid.substring(0, 8).toUpperCase() }}
                                  </span>
                              </div>
                          </div>

                          <!-- Empty state -->
                          <div v-if="!isLoadingItems && items.length === 0" class="text-xs text-center text-gray-400 py-4">
                              暂无关联试剂条码
                          </div>
                      </div>

                      <!-- Completion timestamp -->
                      <div v-if="getStepStatus('instorage') === 'completed' && request.updated_at"
                           class="px-4 pb-3 text-[10px] text-gray-400 flex items-center gap-1">
                          <Clock class="w-3 h-3" /> 完成时间: {{ formatDateTime(request.updated_at) }}
                      </div>
                  </div>
              </div>

          </div>
      </div>

      <!-- Toast Notification -->
      <Transition
        enter-active-class="transition ease-out duration-300"
        enter-from-class="translate-y-4 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition ease-in duration-200"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-4 opacity-0"
      >
        <div v-if="showToast" class="fixed bottom-6 right-6 z-[9999] max-w-sm">
          <div :class="[
            'px-4 py-3 rounded-lg shadow-lg border text-sm font-medium flex items-center gap-2',
            toastType === 'success' ? 'bg-green-50 text-green-800 border-green-200' : 'bg-red-50 text-red-800 border-red-200'
          ]">
            <span>{{ toastMessage }}</span>
          </div>
        </div>
      </Transition>
  </Dialog>
</template>
