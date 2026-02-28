<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Input from '@/components/ui/Input.vue'
import Dialog from '@/components/ui/Dialog.vue'
import TableSection from '@/components/ui/TableSection.vue'
import FlowDetailDialog from '@/components/workflow/FlowDetailDialog.vue'
import { CheckCircle, Clock, Search, Package, MapPin, User } from 'lucide-vue-next'
import LedgerTable from './LedgerTable.vue'
import { toast } from 'vue-sonner'
import { useActionFeedback } from '@/lib/feedback'
import { getProcurementReceiveDisplayStatus, getProcurementReceiveStatusVariant } from '@/lib/reagent-status'

const pendingItems = ref<any[]>([])
const loading = ref(false)
const searchQuery = ref('')
const labelDialogOpen = ref(false)
const labelPrintUUIDs = ref<string[]>([])
const { isPending, runAction } = useActionFeedback()

const filteredPendingItems = computed(() => {
    let result = pendingItems.value
    if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase()
        result = result.filter(i => 
            i.reagent_name?.toLowerCase().includes(q) ||
            i.cas_number?.toLowerCase().includes(q) ||
            i.batch?.order_number?.toLowerCase().includes(q)
        )
    }
    return result
})

const filteredStockItems = computed(() => {
    let result = pendingStockItems.value
    if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase()
        result = result.filter(i => 
            i.reagent_catalog?.name?.toLowerCase().includes(q) ||
            String(i.uuid).toLowerCase().includes(q) ||
            String(i.batch_number || '').toLowerCase().includes(q)
        )
    }
    return result
})
const overdueCount = computed(() => pendingStockItems.value.filter((item: any) => isOverdue(item.created_at)).length)

const fetchPendingReceives = async () => {
    loading.value = true
    try {
        const res = await axios.get('/api/reagents/pending-receives')
        pendingItems.value = res.data
    } catch (e: any) {
        toast.error(e.response?.data?.error || "获取待收货明细失败")
    } finally {
        loading.value = false
    }
}

const handleReceive = async (item: any) => {
    const qty = Math.max((item.quantity || 0) - (item.received_quantity || 0), 0)
    if (qty <= 0) {
        toast.error("该条目无需重复确认到货")
        return
    }

    await runAction(
      `receive-${item.id}`,
      async () => {
        const res = await axios.post(`/api/reagents/pending-receives/${item.id}/receive`, { quantity: qty })
        const uuids: string[] = res.data?.created_uuids || []
        if (uuids.length > 0) {
          labelPrintUUIDs.value = uuids
          labelDialogOpen.value = true
        }
        await fetchPendingReceives()
      },
      {
        successMessage: `到货确认成功 ${qty} 瓶，已生成二维码。`,
        errorMessage: '到货确认失败'
      }
    ).catch(() => {})
}

const printLabels = async () => {
    if (labelPrintUUIDs.value.length === 0) return
    await runAction(
      'print-labels',
      async () => {
        const res = await axios.post('/api/reagents/items/print-labels', { uuids: labelPrintUUIDs.value })
        labelDialogOpen.value = false
        return res
      },
      {
        successMessage: `已生成并记录 ${labelPrintUUIDs.value.length} 张标签`,
        errorMessage: '标签打印失败'
      }
    ).catch(() => {})
}

// === 新增：未入库追踪逻辑 ===
const activeTab = ref<'receiving' | 'tracking'>('receiving')
const pendingStockItems = ref<any[]>([])
const loadingTracking = ref(false)
const flowDialogOpen = ref(false)
const flowDialogContext = ref<{ type: 'receiving' | 'tracking'; item: any } | null>(null)

const fetchPendingStockItems = async () => {
    loadingTracking.value = true
    try {
        // 请求“已赋码，等待研发入库”的物资（status=已到货）
        const res = await axios.get('/api/reagents/items?status=已到货')
        pendingStockItems.value = res.data
    } catch (e: any) {
        toast.error(e.response?.data?.error || "获取追踪清单失败")
    } finally {
        loadingTracking.value = false
    }
}

const isOverdue = (createdAt: string) => {
    if (!createdAt) return false
    const diffHours = (Date.now() - new Date(createdAt).getTime()) / (1000 * 60 * 60)
    return diffHours > 24
}

const handleRemind = (item: any) => {
    // 模拟的催办动作
    toast.success(`已向研发值班组发送“条码 #${String(item.uuid || '').substring(0, 8).toUpperCase()}”超期入库提醒！`)
}

const openFlowDialog = (type: 'receiving' | 'tracking', item: any) => {
    flowDialogContext.value = { type, item }
    flowDialogOpen.value = true
}

const formatFlowTime = (time?: string) => {
    if (!time) return ''
    return new Date(time).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const flowTitle = computed(() => flowDialogContext.value?.type === 'receiving' ? '到货确认流转单' : '待入库跟进流转单')

const flowSubtitle = computed(() => {
    const ctx = flowDialogContext.value
    if (!ctx) return ''
    if (ctx.type === 'receiving') return ctx.item?.reagent_name || '到货确认'
    return ctx.item?.reagent_catalog?.name || '待入库台账'
})

const flowStatus = computed(() => {
    const ctx = flowDialogContext.value
    if (!ctx) return '-'
    if (ctx.type === 'receiving') return getProcurementReceiveDisplayStatus(ctx.item?.receive_status)
    return isOverdue(ctx.item?.created_at) ? '超24H未入库' : '等待研发领走'
})

const flowMeta = computed(() => {
    const ctx = flowDialogContext.value
    if (!ctx) return []
    if (ctx.type === 'receiving') {
        const item = ctx.item
        return [
            { label: '订单号', value: item.batch?.order_number || '-' },
            { label: '供应商', value: item.supplier || '-' },
            { label: '采购数量', value: `${item.quantity || 0}${item.unit || '瓶'}` },
            { label: '导入时间', value: formatFlowTime(item.batch?.created_at || item.created_at) || '-' },
        ]
    }
    const item = ctx.item
    return [
        { label: '条码', value: `#${String(item.uuid || '').substring(0, 8).toUpperCase()}` },
        { label: '批次号', value: item.batch_number || '-' },
        { label: 'CAS', value: item.reagent_catalog?.cas_number || '-' },
        { label: '当前位置', value: item.location || '暂存区' },
        { label: '到货时间', value: formatFlowTime(item.created_at) || '-' },
    ]
})

const flowSteps = computed(() => {
    const ctx = flowDialogContext.value
    if (!ctx) return []

    if (ctx.type === 'receiving') {
        const item = ctx.item
        const isFullyReceived = item.receive_status === '已收货'
        const hasReceived = Number(item.received_quantity || 0) > 0
        return [
            {
                key: 'import',
                label: '采购导入',
                state: 'completed' as const,
                description: `批次 ${item.batch?.order_number || '-'} 已导入系统`,
                operator: '采购人员',
                time: formatFlowTime(item.batch?.created_at || item.created_at),
            },
            {
                key: 'receive',
                label: '到货确认',
                state: isFullyReceived ? 'completed' as const : (hasReceived ? 'current' as const : 'pending' as const),
                description: isFullyReceived ? '采购人员已完成到货确认' : '等待采购人员确认到货',
                operator: '采购人员',
                time: hasReceived ? formatFlowTime(item.updated_at) : undefined,
            },
            {
                key: 'staging',
                label: '条码生成与暂存',
                state: hasReceived ? 'completed' as const : 'pending' as const,
                description: hasReceived ? '已生成条码并转入暂存区等待研发入库' : '到货确认后自动生成条码并进入暂存区',
                operator: hasReceived ? '系统' : undefined,
            },
            {
                key: 'checkin',
                label: '研发扫码入库',
                state: 'pending' as const,
                description: '研发人员在到货台账选择实验室与试剂柜后完成入库',
            },
        ]
    }

    const item = ctx.item
    const overdue = isOverdue(item.created_at)
    return [
        {
            key: 'receive',
            label: '到货确认',
            state: 'completed' as const,
            description: '采购已完成到货确认并生成瓶身条码',
            operator: '采购人员',
            time: formatFlowTime(item.created_at),
        },
        {
            key: 'staging',
            label: '暂存待入库',
            state: 'current' as const,
            description: overdue ? '条目已超 24h 未入库，建议催办' : '等待研发人员领走并扫码入库',
        },
        {
            key: 'checkin',
            label: '研发扫码入库',
            state: 'pending' as const,
            description: '完成后该条目将从待入库台账移除',
        },
    ]
})

const flowNotes = computed(() => {
    const ctx = flowDialogContext.value
    if (!ctx) return []
    if (ctx.type === 'receiving') {
        return [
            { type: 'info' as const, text: '该流转单仅展示到货执行进度与后续入库节点。' },
        ]
    }
    return [
        { type: isOverdue(ctx.item.created_at) ? 'warning' as const : 'info' as const, text: isOverdue(ctx.item.created_at) ? '该条码已超过 24 小时未入库。' : '条码已进入暂存区，等待研发执行扫码入库。' },
    ]
})

const flowActions = computed(() => {
    const ctx = flowDialogContext.value
    if (!ctx) return []
    if (ctx.type === 'tracking' && isOverdue(ctx.item.created_at)) {
        return [{ key: 'remind', label: '发送催办', variant: 'destructive' as const }]
    }
    return []
})

const onFlowAction = (actionKey: string) => {
    if (actionKey !== 'remind') return
    const ctx = flowDialogContext.value
    if (!ctx || ctx.type !== 'tracking') return
    handleRemind(ctx.item)
}

const receivingColumns = [
  { key: 'reagent', label: '试剂信息' },
  { key: 'supplier', label: '采购供应商' },
  { key: 'progress', label: '采购数量' },
  { key: 'status', label: '到货状态' },
  { key: 'actions', label: '操作', align: 'right' as const },
]

const trackingColumns = [
  { key: 'reagent', label: '试剂信息' },
  { key: 'batch', label: '批次来源' },
  { key: 'location', label: '存放位置' },
  { key: 'status', label: '入库状态' },
  { key: 'time', label: '到货时间' },
  { key: 'actions', label: '操作', align: 'right' as const },
]

onMounted(() => {
    fetchPendingReceives()
    fetchPendingStockItems()
})
</script>

<template>
  <div class="space-y-6">
    <TableSection
      title="到货台账"
      description="采购视角：执行到货确认、跟踪待入库条目，并推动台账闭环"
    >
      <template #actions>
        <Button @click="fetchPendingReceives" variant="outline" size="sm">刷新列表</Button>
      </template>

      <template #toolbar>
        <!-- Tabs & Toolbar -->
        <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
          <div class="relative w-full sm:w-80">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
              <Input v-model="searchQuery" class="pl-9" placeholder="搜索试剂名称、批次号、条码..." />
          </div>
          <div class="apple-segmented w-fit sm:ml-auto">
            <button @click="activeTab='receiving'"
              :class="['apple-segmented-btn-icon', activeTab==='receiving' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">
              到货确认
            </button>
            <button @click="activeTab='tracking'"
              :class="['apple-segmented-btn-icon', activeTab==='tracking' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">
              待入库台账
              <span v-if="overdueCount > 0" class="ml-1 rounded-full bg-red-100 px-1.5 py-0.5 text-[10px] font-semibold text-red-700">
                {{ overdueCount }}
              </span>
            </button>
          </div>
        </div>
      </template>

        <!-- Tab 1: 到货确认清单 -->
        <div v-show="activeTab === 'receiving'">
      <div v-if="loading" class="text-center py-10 text-gray-400">正在加载待到货确认清单...</div>
      <div v-else-if="filteredPendingItems.length === 0" class="apple-table-empty">目前没有待入库或待确认的试剂。</div>
      
      <LedgerTable v-else :columns="receivingColumns">
            <tr v-for="item in filteredPendingItems" :key="item.id" class="bg-white border-b hover:bg-gray-50 group">
              <!-- 物资信息 -->
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="p-2 rounded-lg bg-blue-50 text-blue-600 hidden sm:block">
                    <Package class="h-5 w-5" />
                  </div>
                  <div>
                    <h3 class="font-bold text-gray-900 text-left">{{ item.reagent_name }}</h3>
                    <div class="flex items-center gap-2 mt-1">
                      <span class="text-xs text-gray-500 font-mono">{{ item.cas_number || '--' }}</span>
                      <span class="text-[10px] text-gray-500 flex items-center gap-0.5 font-mono bg-gray-100 px-1 py-0.5 rounded">
                        凭证: {{ item.batch?.order_number || '无' }}
                      </span>
                    </div>
                  </div>
                </div>
              </td>

              <!-- 供应商 -->
              <td class="px-6 py-4">
                <div class="flex items-center gap-1.5 text-gray-700">
                  <MapPin class="h-4 w-4 text-gray-400 shrink-0" />
                  <span class="truncate">{{ item.supplier }}</span>
                </div>
              </td>

              <!-- 数量明细 -->
              <td class="px-6 py-4">
                <div class="text-sm font-medium text-gray-900">
                  <span class="text-gray-500 font-normal">采购数量:</span> {{ item.quantity }}{{item.unit}} 
                </div>
              </td>

              <!-- 状态 -->
              <td class="px-6 py-4">
                <Badge :variant="getProcurementReceiveStatusVariant(item.receive_status)">{{ getProcurementReceiveDisplayStatus(item.receive_status) }}</Badge>
              </td>

              <!-- 到货动作 -->
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2 shrink-0">
                  <Button @click="openFlowDialog('receiving', item)" variant="outline" size="sm" class="h-8 text-xs whitespace-nowrap">
                      流转单
                  </Button>
                  <Button @click="handleReceive(item)" :disabled="isPending(`receive-${item.id}`)" variant="primary" size="sm" class="h-8 shadow-sm text-xs px-3 whitespace-nowrap">
                      <CheckCircle class="w-3.5 h-3.5 mr-1" />确认到货
                  </Button>
                </div>
              </td>
            </tr>
      </LedgerTable>
    </div>

    <!-- Tab 2: 等待入库跟进 -->
    <div v-show="activeTab === 'tracking'">
      <div v-if="loadingTracking" class="text-center py-10 text-gray-400">正在加载待入库台账...</div>
      <div v-else-if="filteredStockItems.length === 0" class="apple-table-empty">
        所有已赋码试剂均已入库，暂无需要跟进的条目。
      </div>
      <LedgerTable v-else :columns="trackingColumns">
            <tr v-for="item in filteredStockItems" :key="item.uuid" 
                class="bg-white border-b hover:bg-gray-50 transition-colors"
                :class="isOverdue(item.created_at) ? 'bg-red-50/10' : ''">
              <!-- 实体信息 -->
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="p-2 rounded-lg hidden sm:block"
                       :class="isOverdue(item.created_at) ? 'bg-red-100 text-red-600' : 'bg-blue-50 text-blue-600'">
                    <Package class="h-5 w-5" />
                  </div>
                  <div>
                    <h3 class="font-bold text-gray-900 text-left">
                      {{ item.reagent_catalog?.name }}
                    </h3>
                    <div class="flex items-center gap-2 mt-1">
                      <span class="text-xs text-gray-500 font-mono">{{ item.batch_number || '--' }}</span>
                      <span class="text-[10px] text-gray-500 flex items-center gap-0.5 font-mono bg-gray-100 px-1 py-0.5 rounded">
                        #{{ String(item.uuid).substring(0,8).toUpperCase() }}
                      </span>
                    </div>
                  </div>
                </div>
              </td>

              <!-- 来源单据 -->
              <td class="px-6 py-4">
                <div class="flex flex-col gap-1 text-sm text-gray-700">
                  <span class="font-medium text-gray-900">批次号 {{ item.batch_number || '--' }}</span>
                  <div class="flex items-center gap-1.5 text-xs text-gray-500 mt-0.5">
                    <User class="w-3.5 h-3.5" />
                    <span>条码 #{{ String(item.uuid).substring(0, 8).toUpperCase() }}</span>
                  </div>
                </div>
              </td>

              <!-- 存放位置 -->
              <td class="px-6 py-4">
                <div class="flex items-center gap-1.5 text-sm text-gray-700">
                  <MapPin class="h-4 w-4 text-gray-400 shrink-0" />
                  <span>{{ item.location || '暂存区' }}</span>
                </div>
              </td>

              <!-- 台账状态 -->
              <td class="px-6 py-4">
                <Badge v-if="isOverdue(item.created_at)" variant="destructive" class="px-1.5">超24H未入库</Badge>
                <Badge v-else variant="info" class="px-1.5 py-0.5">待研发入库</Badge>
              </td>

              <!-- 时间 -->
              <td class="px-6 py-4">
                <div class="flex items-center gap-1 text-xs text-gray-500">
                  <Clock class="w-3.5 h-3.5 text-gray-400 shrink-0" />
                  <span class="truncate">{{ new Date(item.created_at).toLocaleString('zh-CN') }}</span>
                </div>
              </td>

              <!-- 催办动作 -->
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <Button @click="openFlowDialog('tracking', item)" variant="outline" size="sm" class="h-8 text-xs whitespace-nowrap">
                    流转单
                  </Button>
                  <Button v-if="isOverdue(item.created_at)" 
                          @click="handleRemind(item)" 
                          variant="destructive" size="sm" class="flex items-center shadow-sm text-xs h-8 whitespace-nowrap">
                      <Clock class="w-3.5 h-3.5 mr-1" />
                      一键催办
                  </Button>
                  <span v-else class="text-xs text-slate-400">仅查看</span>
                </div>
              </td>
            </tr>
      </LedgerTable>
    </div>
    </TableSection>

    <Dialog :open="labelDialogOpen" size="md" title="到货标签打印" @close="labelDialogOpen = false">
      <div class="space-y-4 p-6">
        <p class="text-sm text-slate-600">以下瓶身已完成赋码，请打印并贴标后移交研发入库：</p>
        <div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
          <div class="max-h-40 space-y-1 overflow-auto text-xs font-mono text-slate-700">
            <div v-for="uuid in labelPrintUUIDs" :key="uuid">#{{ uuid }}</div>
          </div>
        </div>
      </div>
      <template #footer>
        <Button variant="secondary" @click="labelDialogOpen = false">稍后打印</Button>
        <Button variant="primary" :disabled="isPending('print-labels')" @click="printLabels">
          {{ isPending('print-labels') ? '打印中...' : '打印标签' }}
        </Button>
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
      @action="onFlowAction"
    />
  </div>
</template>
