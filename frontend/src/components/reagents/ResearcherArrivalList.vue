<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import axios from 'axios'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import { CheckCircle, Clock, MapPin, Package, FileText, Search, Loader2 } from 'lucide-vue-next'
import ItemLifecycleDialog from '@/components/reagents/ItemLifecycleDialog.vue'
import Input from '@/components/ui/Input.vue'
import LedgerTable from './LedgerTable.vue'
import Dialog from '@/components/ui/Dialog.vue'
import TableSection from '@/components/ui/TableSection.vue'
import FlowDetailDialog from '@/components/workflow/FlowDetailDialog.vue'
import { formatRatio } from '@/lib/quantity'
import { toast } from 'vue-sonner'
import { useActionFeedback } from '@/lib/feedback'
import { getInventoryDisplayStatus, getInventoryStatusVariant, isArrivedStatus, isInStorageStatus, isUsedStatus } from '@/lib/reagent-status'

const items = ref<any[]>([])
const loading = ref(false)
const processing = ref<string | null>(null)
const cabinets = ref<any[]>([])
const searchQuery = ref('')
const checkInDialogOpen = ref(false)
const checkInTarget = ref<any>(null)
const checkInCabinetId = ref<number>(0)
const flowDialogOpen = ref(false)
const flowDialogItem = ref<any>(null)
const { isPending, runAction } = useActionFeedback()

const props = defineProps<{
  userId?: number
}>()

// 模拟当前用户 ID，如果外部没有传入，默认为 1 (admin，用于测试)
const currentUserId = computed(() => props.userId || 1) 

// 状态过滤 Tab
const statusFilter = ref('已到货')
const statusOptions = ['全部', '已到货', '已入库']

// --- 档案弹窗 ---
const lifecycleDialog = ref({
  isOpen: false,
  itemUuid: null as string | null
})
const openLifecycleDialog = (uuid: string) => {
  lifecycleDialog.value = { isOpen: true, itemUuid: uuid }
}

const fetchArrivals = async () => {
    loading.value = true
    try {
        // 获取“已赋码”试剂实体，按状态过滤出到货/在库记录
        const res = await axios.get('/api/reagents/items')
        items.value = res.data
    } catch (e) {
        console.error("Failed to fetch arrivals", e)
        toast.error('获取到货台账失败')
    } finally {
        loading.value = false
    }
}

watch(currentUserId, () => {
    fetchArrivals()
})

const filteredItems = computed(() => {
    const q = searchQuery.value.trim().toLowerCase()
    const source = statusFilter.value === '全部'
      ? items.value.filter((i: any) => !isUsedStatus(i.status))
      : items.value.filter((i: any) => getInventoryDisplayStatus(i.status) === statusFilter.value)

    if (!q) return source

    return source.filter((i: any) =>
      i.reagent_catalog?.name?.toLowerCase()?.includes(q) ||
      i.reagent_catalog?.cas_number?.toLowerCase()?.includes(q) ||
      String(i.batch_number || '').toLowerCase().includes(q) ||
      String(i.uuid || '').toLowerCase().includes(q)
    )
})

const getLedgerStatus = (item: any) => {
    if (isArrivedStatus(item.status)) return '待我入库'
    if (isInStorageStatus(item.status)) return '已入库'
    if (isUsedStatus(item.status)) return '已耗尽'
    return getInventoryDisplayStatus(item.status)
}

const getStatusVariant = (item: any) => {
    return getInventoryStatusVariant(item.status)
}

const formatArriveTime = (t: string) => {
    if (!t) return '--'
    return new Date(t).toLocaleString('zh-CN')
}

const arrivalColumns = [
  { key: 'reagent', label: '试剂信息' },
  { key: 'batch', label: '批次记录' },
  { key: 'location', label: '存放位置' },
  { key: 'status', label: '台账状态' },
  { key: 'time', label: '到货时间' },
  { key: 'actions', label: '操作' },
]

const fetchCabinets = async () => {
    try {
        const res = await axios.get('/api/reagents/cabinets')
        cabinets.value = res.data
    } catch (e) {
        console.error("Failed to fetch cabinets", e)
        toast.error('获取试剂柜列表失败')
    }
}

const checkInCabinets = computed(() => {
    if (!checkInTarget.value) return cabinets.value
    const isControlled = !!checkInTarget.value.reagent_catalog?.is_controlled
    return cabinets.value.filter((cab: any) =>
      isControlled ? cab.cabinet_type === '易制毒制爆试剂柜' : cab.cabinet_type !== '易制毒制爆试剂柜'
    )
})

const selectedCabinet = computed(() => checkInCabinets.value.find((c: any) => c.id === checkInCabinetId.value))

const openCheckInDialog = (item: any) => {
    checkInTarget.value = item
    const defaultCab = cabinets.value.find((cab: any) =>
      item.reagent_catalog?.is_controlled ? cab.cabinet_type === '易制毒制爆试剂柜' : cab.cabinet_type !== '易制毒制爆试剂柜'
    )
    checkInCabinetId.value = defaultCab?.id || 0
    checkInDialogOpen.value = true
}

const submitCheckIn = async () => {
    if (!checkInTarget.value || !checkInCabinetId.value) return
    processing.value = checkInTarget.value.uuid
    await runAction(
      `checkin-${checkInTarget.value.uuid}`,
      async () => {
        await axios.post(`/api/reagents/items/${checkInTarget.value.uuid}/check-in`, {
          lab_room: selectedCabinet.value?.location || '',
          cabinet_id: checkInCabinetId.value
        })
        checkInDialogOpen.value = false
        await fetchArrivals()
      },
      {
        successMessage: '入库完成',
        errorMessage: '入库失败，请重试'
      }
    ).catch(() => {})
    processing.value = null
}

const openFlowDialog = (item: any) => {
    flowDialogItem.value = item
    flowDialogOpen.value = true
}

const flowTitle = computed(() => '到货入库流转单')
const flowSubtitle = computed(() => flowDialogItem.value?.reagent_catalog?.name || '到货台账')
const flowStatus = computed(() => flowDialogItem.value ? getLedgerStatus(flowDialogItem.value) : '-')
const flowMeta = computed(() => {
    const item = flowDialogItem.value
    if (!item) return []
    return [
      { label: '条码', value: `#${String(item.uuid || '').substring(0, 8).toUpperCase()}` },
      { label: '批次号', value: item.batch_number || '-' },
      { label: 'CAS', value: item.reagent_catalog?.cas_number || '-' },
      { label: '当前位置', value: item.location || '暂存区' },
      { label: '到货时间', value: formatArriveTime(item.created_at) },
    ]
})
const flowSteps = computed(() => {
    const item = flowDialogItem.value
    if (!item) return []
    const arrived = isArrivedStatus(item.status)
    const inStorage = isInStorageStatus(item.status)
    const used = isUsedStatus(item.status)
    return [
      {
        key: 'receive',
        label: '采购到货确认',
        state: 'completed' as const,
        description: '采购侧已完成到货确认并生成条码',
        operator: '采购人员',
        time: formatArriveTime(item.created_at),
      },
      {
        key: 'staging',
        label: '赋码暂存',
        state: 'completed' as const,
        description: '条码已进入暂存区，等待研发入库',
        operator: '系统',
      },
      {
        key: 'checkin',
        label: '研发扫码入库',
        state: inStorage || used ? 'completed' as const : (arrived ? 'current' as const : 'pending' as const),
        description: inStorage || used ? '已完成实验室与试剂柜入库' : '请在当前页面确认入库',
        operator: inStorage || used ? '研发人员' : undefined,
      },
      {
        key: 'manage',
        label: '库存管理',
        state: used ? 'completed' as const : (inStorage ? 'current' as const : 'pending' as const),
        description: used ? '已耗尽归档' : '进入库存台账进行后续管理',
      },
    ]
})
const flowNotes = computed(() => {
    const item = flowDialogItem.value
    if (!item) return []
    if (isArrivedStatus(item.status)) {
      return [{ type: 'info' as const, text: '当前仍在暂存区，请选择实验室与试剂柜完成入库。' }]
    }
    if (isInStorageStatus(item.status)) {
      return [{ type: 'success' as const, text: '该条码已入库，可在库存台账执行使用或耗尽。' }]
    }
    return [{ type: 'info' as const, text: '该条码已进入生命周期后续阶段。' }]
})
const flowActions = computed(() => {
    const item = flowDialogItem.value
    if (!item || !isArrivedStatus(item.status)) return []
    return [{ key: 'checkin', label: '确认入库', variant: 'primary' as const }]
})
const handleFlowAction = (actionKey: string) => {
    if (actionKey !== 'checkin' || !flowDialogItem.value) return
    flowDialogOpen.value = false
    openCheckInDialog(flowDialogItem.value)
}

onMounted(() => {
    fetchArrivals()
    fetchCabinets()
})
</script>

<template>
  <div class="space-y-6">
    <TableSection title="到货台账" description="研发视角：查看属于我的到货条目并完成入库确认">
      <template #actions>
        <Button @click="fetchArrivals" variant="outline" size="sm">刷新列表</Button>
      </template>

      <template #toolbar>
        <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
          <div class="relative w-full sm:w-80">
            <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
            <Input v-model="searchQuery" class="pl-9" placeholder="搜索试剂名称、批次号、条码..." />
          </div>
          <div class="apple-segmented w-fit sm:ml-auto">
            <button
              v-for="s in statusOptions"
              :key="s"
              @click="statusFilter = s"
              :class="[
                'apple-segmented-btn',
                statusFilter === s
                  ? 'apple-segmented-btn-active'
                  : 'apple-segmented-btn-idle'
              ]"
            >
              {{ s }}
            </button>
          </div>
        </div>
      </template>

      <div v-if="loading" class="flex justify-center py-12">
        <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
      </div>

      <div v-else-if="filteredItems.length === 0" class="apple-table-empty">
        没有找到符合当前条件“{{ statusFilter }}”的试剂记录。
      </div>

      <LedgerTable v-else :columns="arrivalColumns">
              <tr v-for="item in filteredItems" :key="item.uuid" 
                  class="bg-white border-b hover:bg-gray-50 group transition-colors"
                  :class="isUsedStatus(item.status) ? 'opacity-70 grayscale' : ''">
                <!-- 物资条码与类型 -->
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="p-2 rounded-lg bg-blue-50 text-blue-600 hidden sm:block"
                         :class="isUsedStatus(item.status) ? 'bg-gray-100 text-gray-500' : ''">
                      <Package class="h-5 w-5" />
                    </div>
                    <div>
                      <button @click="openLifecycleDialog(item.uuid)" class="font-bold text-gray-900 hover:text-blue-600 hover:underline text-left">
                        {{ item.reagent_catalog?.name }}
                      </button>
                      <div class="flex items-center gap-2 mt-1">
                        <span class="text-xs text-gray-500 font-mono">{{ item.cas_number || '--' }}</span>
                        <button @click="openLifecycleDialog(item.uuid)" class="text-[10px] text-blue-500 hover:underline flex items-center gap-0.5 font-mono bg-blue-50 px-1 py-0.5 rounded"
                                :class="isUsedStatus(item.status) ? 'text-gray-500 bg-gray-100' : ''">
                          <FileText class="w-3 h-3" /> #{{ String(item.uuid).substring(0,8).toUpperCase() }}
                        </button>
                      </div>
                    </div>
                  </div>
                </td>
                
                <!-- 批次记录 -->
                <td class="px-6 py-4 text-gray-600">
                  <div class="font-medium text-gray-900">批次 {{ item.batch_number || '--' }}</div>
                  <div class="text-xs text-gray-500 mt-1">条码 #{{ String(item.uuid).substring(0,8).toUpperCase() }}</div>
                </td>

                <!-- 位置信息 -->
                <td class="px-6 py-4">
                  <div class="flex items-center gap-1.5 text-gray-700">
                    <MapPin class="h-4 w-4 text-gray-400" />
                    <span v-if="isArrivedStatus(item.status)" class="text-amber-600 font-medium">{{ item.location || '暂存区' }}</span>
                    <span v-else class="text-emerald-600 font-medium">{{ item.location || '已不在柜' }}</span>
                  </div>
                </td>

                <!-- 台账状态 -->
                <td class="px-6 py-4">
                  <Badge :variant="getStatusVariant(item)">{{ getLedgerStatus(item) }}</Badge>
                  <div v-if="isInStorageStatus(item.status)" class="mt-1 text-[11px] text-emerald-600">
                    余量 {{ formatRatio(item.remaining_volume, item.capacity, item.reagent_catalog?.unit, 'ml') }}
                  </div>
                </td>

                <!-- 到货时间 -->
                <td class="px-6 py-4">
                  <div class="flex items-center gap-1 text-xs text-gray-500">
                    <Clock class="h-3.5 w-3.5 text-gray-400" />
                    {{ formatArriveTime(item.created_at) }}
                  </div>
                </td>

                <!-- 操作栏 -->
                <td class="px-6 py-4 text-right">
                  <div class="flex items-center justify-end gap-2">
                    <Button variant="outline" size="sm" class="h-8 whitespace-nowrap" @click="openFlowDialog(item)">
                      流转单
                    </Button>
                    <Button v-if="isArrivedStatus(item.status)"
                      @click="openCheckInDialog(item)"
                      :disabled="processing === item.uuid || isPending(`checkin-${item.uuid}`)"
                      variant="primary"
                      size="sm"
                      class="h-8 shadow-sm whitespace-nowrap"
                    >
                      <Loader2 v-if="processing === item.uuid" class="mr-1.5 h-4 w-4 animate-spin" />
                      <CheckCircle v-else class="h-4 w-4 mr-1.5" />
                      确认入库
                    </Button>
                    <span v-else class="text-xs text-slate-400">仅查看</span>
                  </div>
                </td>
              </tr>
      </LedgerTable>
    </TableSection>

    <!-- 引入全生命周期悬浮窗 -->
    <ItemLifecycleDialog 
      :is-open="lifecycleDialog.isOpen" 
      :item-uuid="lifecycleDialog.itemUuid"
      @close="lifecycleDialog.isOpen = false"
      @refresh-needed="fetchArrivals"
    />

    <Dialog :open="checkInDialogOpen" size="sm" title="选择入库位置" @close="checkInDialogOpen = false">
      <div class="space-y-4 p-6">
        <p class="text-sm text-slate-600">
          {{ checkInTarget?.reagent_catalog?.name }}（#{{ checkInTarget?.uuid?.substring(0, 8)?.toUpperCase() }}）
        </p>
        <div class="space-y-2">
          <label class="text-xs font-medium text-slate-600">实验室（由试剂柜自动确定）</label>
          <Input :model-value="selectedCabinet?.location || '请选择试剂柜'" disabled />
        </div>
        <div class="space-y-2">
          <label class="text-xs font-medium text-slate-600">试剂柜</label>
          <select v-model.number="checkInCabinetId" class="h-10 w-full">
            <option :value="0">请选择试剂柜</option>
            <option v-for="cab in checkInCabinets" :key="cab.id" :value="cab.id">
              {{ cab.name }}（{{ cab.location }}）
            </option>
          </select>
        </div>
      </div>
      <template #footer>
        <Button variant="secondary" @click="checkInDialogOpen = false">取消</Button>
        <Button variant="primary" :disabled="!checkInCabinetId || !!processing || (checkInTarget && isPending(`checkin-${checkInTarget.uuid}`))" @click="submitCheckIn">确认入库</Button>
      </template>
    </Dialog>

    <FlowDetailDialog
      :open="flowDialogOpen"
      :title="flowTitle"
      :subtitle="flowSubtitle"
      :status="flowStatus"
      :meta="flowMeta"
      :steps="flowSteps"
      :actions="flowActions"
      :notes="flowNotes"
      @close="flowDialogOpen = false"
      @action="handleFlowAction"
    />
  </div>
</template>
