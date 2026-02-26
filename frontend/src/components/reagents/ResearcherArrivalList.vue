<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import axios from 'axios'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import { CheckCircle, Clock, MapPin, Package, FileText, Search, Loader2 } from 'lucide-vue-next'
import ItemLifecycleDialog from '@/components/reagents/ItemLifecycleDialog.vue'
import Input from '@/components/ui/Input.vue'
import LedgerTable from './LedgerTable.vue'
import Dialog from '@/components/ui/Dialog.vue'
import { formatRatio } from '@/lib/quantity'
import { toast } from 'vue-sonner'
import { useActionFeedback } from '@/lib/feedback'

const items = ref<any[]>([])
const loading = ref(false)
const processing = ref<string | null>(null)
const cabinets = ref<any[]>([])
const searchQuery = ref('')
const checkInDialogOpen = ref(false)
const checkInTarget = ref<any>(null)
const checkInCabinetId = ref<number>(0)
const { isPending, runAction } = useActionFeedback()

const props = defineProps<{
  userId?: number
}>()

// 模拟当前用户 ID，如果外部没有传入，默认为 1 (admin，用于测试)
const currentUserId = computed(() => props.userId || 1) 

// 状态过滤 Tab
const statusFilter = ref('已到货')
const statusOptions = ['全部', '已到货', '在库']

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
        // 请求：获取当前申购人全生命周期(所有已赋码)的试剂实体
        const res = await axios.get(`/api/reagents/items?requestor_id=${currentUserId.value}`)
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
      ? items.value.filter((i: any) => i.status !== '已耗尽')
      : items.value.filter((i: any) => i.status === statusFilter.value)

    if (!q) return source

    return source.filter((i: any) =>
      i.reagent_catalog?.name?.toLowerCase()?.includes(q) ||
      i.reagent_catalog?.cas_number?.toLowerCase()?.includes(q) ||
      String(i.batch_number || '').toLowerCase().includes(q) ||
      String(i.uuid || '').toLowerCase().includes(q) ||
      String(i.reagent_request_id || '').toLowerCase().includes(q)
    )
})

const getLedgerStatus = (item: any) => {
    if (item.status === '已到货') return '待我入库'
    if (item.status === '在库') return '已入库'
    if (item.status === '已耗尽') return '已耗尽'
    return item.status || '未知'
}

const getStatusVariant = (item: any) => {
    if (item.status === '已到货') return 'warning'
    if (item.status === '在库') return 'success'
    if (item.status === '已耗尽') return 'outline'
    return 'default'
}

const formatArriveTime = (t: string) => {
    if (!t) return '--'
    return new Date(t).toLocaleString('zh-CN')
}

const arrivalColumns = [
  { key: 'reagent', label: '试剂信息' },
  { key: 'request', label: '来源申请' },
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

onMounted(() => {
    fetchArrivals()
    fetchCabinets()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex justify-between items-center">
      <div>
        <h2 class="text-xl font-bold text-gray-900">到货台账</h2>
        <p class="text-sm text-gray-500 mt-1">研发视角：查看属于我的到货条目并完成入库确认</p>
      </div>
      <Button @click="fetchArrivals" variant="outline" size="sm">刷新列表</Button>
    </div>

    <Card>
      <div class="p-6 space-y-4">
        <!-- Toolbar -->
        <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
          <div class="relative w-72">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
              <Input v-model="searchQuery" class="pl-9" placeholder="搜索试剂名称、批次号、申请单号..." />
          </div>
          <div class="apple-segmented w-fit">
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

        <!-- Content -->
        <div v-if="loading" class="flex justify-center py-12">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>

        <div v-else-if="filteredItems.length === 0" class="bg-gray-50 rounded-xl border-2 border-dashed py-16 text-center">
          <Package class="h-12 w-12 text-gray-300 mx-auto mb-3" />
          <h3 class="text-lg font-medium text-gray-900">暂无相关试剂</h3>
          <p class="text-gray-500 mt-1">没有找到符合当前条件“{{ statusFilter }}”的试剂记录</p>
        </div>

        <LedgerTable v-else :columns="arrivalColumns">
              <tr v-for="item in filteredItems" :key="item.uuid" 
                  class="bg-white border-b hover:bg-gray-50 group transition-colors"
                  :class="item.status === '已耗尽' ? 'opacity-70 grayscale' : ''">
                <!-- 物资条码与类型 -->
                <td class="px-6 py-4">
                  <div class="flex items-center gap-3">
                    <div class="p-2 rounded-lg bg-blue-50 text-blue-600 hidden sm:block"
                         :class="item.status === '已耗尽' ? 'bg-gray-100 text-gray-500' : ''">
                      <Package class="h-5 w-5" />
                    </div>
                    <div>
                      <button @click="openLifecycleDialog(item.uuid)" class="font-bold text-gray-900 hover:text-blue-600 hover:underline text-left">
                        {{ item.reagent_catalog?.name }}
                      </button>
                      <div class="flex items-center gap-2 mt-1">
                        <span class="text-xs text-gray-500 font-mono">{{ item.cas_number || '--' }}</span>
                        <button @click="openLifecycleDialog(item.uuid)" class="text-[10px] text-blue-500 hover:underline flex items-center gap-0.5 font-mono bg-blue-50 px-1 py-0.5 rounded"
                                :class="item.status==='已耗尽'? 'text-gray-500 bg-gray-100' : ''">
                          <FileText class="w-3 h-3" /> #{{ String(item.uuid).substring(0,8).toUpperCase() }}
                        </button>
                      </div>
                    </div>
                  </div>
                </td>
                
                <!-- 批次记录 -->
                <td class="px-6 py-4 text-gray-600">
                  <div class="font-medium text-gray-900">申购单 #{{ item.reagent_request_id || '--' }}</div>
                  <div class="text-xs text-gray-500 mt-1">{{ item.batch_number || '无批次号' }}</div>
                </td>

                <!-- 位置信息 -->
                <td class="px-6 py-4">
                  <div class="flex items-center gap-1.5 text-gray-700">
                    <MapPin class="h-4 w-4 text-gray-400" />
                    <span v-if="item.status === '已到货'"><span class="text-amber-600 font-medium">{{ item.location }}</span> (暂存区)</span>
                    <span v-else><span class="text-emerald-600 font-medium">{{ item.location || '已不在柜' }}</span></span>
                  </div>
                </td>

                <!-- 台账状态 -->
                <td class="px-6 py-4">
                  <Badge :variant="getStatusVariant(item)">{{ getLedgerStatus(item) }}</Badge>
                  <div v-if="item.status === '在库'" class="mt-1 text-[11px] text-emerald-600">
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
                  <Button v-if="item.status === '已到货'"
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
                  <Button v-else variant="outline" size="sm" class="h-8 text-gray-500 whitespace-nowrap" @click="openLifecycleDialog(item.uuid)">
                    查看台账流水
                  </Button>
                </td>
              </tr>
        </LedgerTable>
      </div>
    </Card>

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
  </div>
</template>
