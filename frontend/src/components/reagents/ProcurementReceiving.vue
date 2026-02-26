<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import Input from '@/components/ui/Input.vue'
import Dialog from '@/components/ui/Dialog.vue'
import { CheckCircle, Clock, Search, Package, MapPin, User } from 'lucide-vue-next'
import LedgerTable from './LedgerTable.vue'
import { toast } from 'vue-sonner'
import { useActionFeedback } from '@/lib/feedback'

const pendingItems = ref<any[]>([])
const loading = ref(false)
const receiveInputs = ref<Record<number, number>>({})
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
            String(i.reagent_request_id).toLowerCase().includes(q)
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
        // Initialize default input to the remaining quantity
        pendingItems.value.forEach(item => {
             receiveInputs.value[item.id] = item.quantity - item.received_quantity
        })
    } catch (e: any) {
        toast.error(e.response?.data?.error || "获取待收货明细失败")
    } finally {
        loading.value = false
    }
}

const handleReceive = async (item: any) => {
    const qty = receiveInputs.value[item.id]
    if (!qty || qty <= 0 || qty > (item.quantity - item.received_quantity)) {
        toast.error("输入的收货数量无效")
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
        successMessage: `成功收货 ${qty} 瓶，已生成二维码。`,
        errorMessage: '收货提交失败'
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

const getStatusVariant = (status: string) => {
   if (status === '待收货') return 'secondary'
   if (status === '部分收货') return 'warning'
   return 'default'
}

// === 新增：未入库追踪逻辑 ===
const activeTab = ref<'receiving' | 'tracking'>('receiving')
const pendingStockItems = ref<any[]>([])
const loadingTracking = ref(false)

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
    toast.success(`已向 ${item.reagent_request?.requestor?.real_name || '申请人'} 发送超期入库提醒！`)
}

const receivingColumns = [
  { key: 'reagent', label: '试剂信息' },
  { key: 'supplier', label: '采购供应商' },
  { key: 'progress', label: '申量及进度' },
  { key: 'status', label: '验收状态' },
  { key: 'actions', label: '执行点收操作', align: 'right' as const },
]

const trackingColumns = [
  { key: 'reagent', label: '试剂信息' },
  { key: 'request', label: '来源申请' },
  { key: 'location', label: '存放位置' },
  { key: 'status', label: '台账状态' },
  { key: 'time', label: '到货时间' },
  { key: 'actions', label: '跟进操作', align: 'right' as const },
]

onMounted(() => {
    fetchPendingReceives()
    fetchPendingStockItems()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex justify-between items-center">
      <div>
        <h2 class="text-lg font-semibold tracking-tight text-gray-900">到货台账</h2>
        <p class="text-[13px] text-gray-500 mt-0.5">采购视角：执行到货确认、跟踪待入库条目，并推动台账闭环</p>
      </div>
      <Button @click="fetchPendingReceives" variant="outline" size="sm">刷新列表</Button>
    </div>

    <Card>
      <div class="p-6 space-y-4">
        <!-- Tabs & Toolbar -->
        <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
          <div class="relative w-72">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
              <Input v-model="searchQuery" class="pl-9" placeholder="搜索试剂名称、批次号、申请单号..." />
          </div>
          <div class="apple-segmented w-fit">
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

        <!-- Tab 1: 待点检验收清单 -->
        <div v-show="activeTab === 'receiving'">
      <div v-if="loading" class="text-center py-10 text-gray-400">正在加载待收货在途清单...</div>
      <div v-else-if="filteredPendingItems.length === 0" class="text-center py-10 text-gray-400">目前没有待入库或待确认的试剂。</div>
      
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
                  <span class="text-gray-500 font-normal">总量:</span> {{ item.quantity }}{{item.unit}} 
                  <span class="text-gray-300 mx-1">|</span> 
                  <span class="text-gray-500 font-normal">待收:</span> <span class="text-blue-600">{{ item.quantity - item.received_quantity }}</span>
                </div>
                <div class="text-xs text-gray-500 mt-1 flex items-center gap-1">
                  <Clock class="w-3 h-3 text-emerald-500" /> 已点收: <span class="text-emerald-600">{{ item.received_quantity }}</span>
                </div>
              </td>

              <!-- 状态 -->
              <td class="px-6 py-4">
                <Badge :variant="getStatusVariant(item.receive_status)">{{ item.receive_status }}</Badge>
              </td>

              <!-- 点收动作 -->
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-3 shrink-0">
                  <div class="flex items-center gap-2">
                    <label class="text-[11px] font-medium text-gray-500 whitespace-nowrap">本次赋码点收</label>
                    <Input type="number" 
                        v-model="receiveInputs[item.id]" 
                        class="w-16 text-center h-8 text-xs focus:ring-1 focus:ring-blue-500" 
                        :min="1" 
                        :max="item.quantity - item.received_quantity" 
                    />
                  </div>
                  <Button @click="handleReceive(item)" :disabled="isPending(`receive-${item.id}`)" variant="primary" size="sm" class="h-8 shadow-sm text-xs px-3 whitespace-nowrap">
                      <CheckCircle class="w-3.5 h-3.5 mr-1" />确认
                  </Button>
                </div>
              </td>
            </tr>
      </LedgerTable>
    </div>

    <!-- Tab 2: 等待入库跟进 -->
    <div v-show="activeTab === 'tracking'">
      <div v-if="loadingTracking" class="text-center py-10 text-gray-400">正在加载待入库台账...</div>
      <div v-else-if="filteredStockItems.length === 0" class="text-center py-10 text-gray-400">
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
                  <span class="font-medium text-gray-900">申购单 #{{ item.reagent_request_id || '--' }}</span>
                  <div class="flex items-center gap-1.5 text-xs text-gray-500 mt-0.5">
                    <User class="w-3.5 h-3.5" />
                    <span>{{ item.reagent_request?.requestor?.real_name || '无法追溯' }} ({{ item.reagent_request?.requestor?.department?.name || '未知组别' }})</span>
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
                <Badge v-else variant="info" class="px-1.5 py-0.5">等待研发领走</Badge>
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
                <Button v-if="isOverdue(item.created_at)" 
                        @click="handleRemind(item)" 
                        variant="destructive" size="sm" class="flex items-center shadow-sm text-xs h-8 ml-auto whitespace-nowrap">
                    <Clock class="w-3.5 h-3.5 mr-1" />
                    一键催办入库
                </Button>
                <div v-else class="text-xs font-medium text-blue-600 flex items-center justify-end gap-1.5">
                  <Clock class="w-3.5 h-3.5" />
                  <span class="whitespace-nowrap">研发未收走</span>
                </div>
              </td>
            </tr>
      </LedgerTable>
    </div>
      </div>
    </Card>

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
  </div>
</template>
