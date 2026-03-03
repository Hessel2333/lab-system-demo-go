<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import axios from 'axios'
import { toast } from 'vue-sonner'
import {
  Loader2, Search, AlertTriangle, FileText
} from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Dialog from '@/components/ui/Dialog.vue'
import TableSection from '@/components/ui/TableSection.vue'
import Input from '@/components/ui/Input.vue'
import LedgerTable from './LedgerTable.vue'
import ItemLifecycleDialog from '@/components/reagents/ItemLifecycleDialog.vue'
import { formatAmount, formatNumber, normalizeUnit } from '@/lib/quantity'
import { getInventoryDisplayStatus, getInventoryStatusVariant, isInStorageStatus } from '@/lib/reagent-status'
import { fetchUsers } from '@/api/organization'
import { useSessionStore } from '@/stores/session'

// ——— 类型定义 ———
interface ReagentItem {
  uuid: string
  reagent_catalog_id: number
  location: string
  cabinet_id: number
  cabinet?: {
    id: number
    name: string
    cabinet_type: string
    location: string
    department_id?: number
  }
  status: string
  capacity: number
  remaining_volume: number
  batch_number: string
  expiry_date: string
  reagent_catalog?: {
    cas_number: string
    name: string
    formula: string
    category: string
    is_controlled: boolean
    unit: string
  }
}

// ——— 视图状态 ———
const viewMode = ref<'table' | 'team'>('team')
const sessionStore = useSessionStore()

// ——— 表格视图状态 ———
const allItems = ref<ReagentItem[]>([])
const isLoadingTable = ref(false)
const searchQuery = ref('')
const statusFilter = ref('全部')
const cabinetFilter = ref('全部')
const currentPage = ref(1)
const pageSize = 15
const statusOptions = ['全部', '已入库', '已到货', '已耗尽']
const cabinetOptions = ['全部', '普通柜', '管控柜']
const tableColumns = [
  { key: 'reagent', label: '试剂名称' },
  { key: 'barcode', label: '系统条码' },
  { key: 'remaining', label: '剩余量' },
  { key: 'cabinet', label: '试剂柜' },
  { key: 'location', label: '库位' },
  { key: 'batch', label: '批次来源' },
  { key: 'status', label: '状态' },
  { key: 'expiry', label: '有效期' },
  { key: 'actions', label: '操作', align: 'right' as const },
]
// 辅助：获取柜子显示名称
const getCabinetName = (item: ReagentItem) => item.cabinet?.name ?? (item.cabinet_id > 0 ? `柜#${item.cabinet_id}` : null)
const isControlledCabinet = (item: ReagentItem) => item.cabinet?.cabinet_type === '易制毒制爆试剂柜'
const getItemLocation = (item: ReagentItem) => item.cabinet?.location || item.location || '—'
const isControlledItem = (item: ReagentItem) => !!item.reagent_catalog?.is_controlled

// ——— 团队范围状态 ———
const currentUserDeptId = ref<number | null>(null)
const resolvingDept = ref(false)

// ——— 消耗弹窗 ———
const consumeDialog = ref({
  isOpen: false,
  item: null as ReagentItem | null,
  volume: 1,
  remarks: ''
})

// --- 档案弹窗 ---
const lifecycleDialog = ref({
  isOpen: false,
  itemUuid: null as string | null
})
const openLifecycleDialog = (uuid: string) => {
  lifecycleDialog.value = { isOpen: true, itemUuid: uuid }
}

const emit = defineEmits(['switch-to-arrival'])

// ——— 数据获取 ———
const pendingArrivals = ref<ReagentItem[]>([])
const fetchPendingArrivals = async () => {
  try {
    const res = await axios.get('/api/reagents/items?status=已到货')
    pendingArrivals.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch pending arrivals', e)
  }
}

const fetchTable = async () => {
  isLoadingTable.value = true
  try {
    const res = await axios.get('/api/reagents/items')
    allItems.value = res.data ?? []
  } catch { /* ignore */ } finally {
    isLoadingTable.value = false
  }
}

const resolveCurrentUserDept = async () => {
  resolvingDept.value = true
  try {
    const users = await fetchUsers()
    const me = users.find((user) => user.ID === sessionStore.currentUserId)
    currentUserDeptId.value = me?.department_id ?? null
  } catch {
    currentUserDeptId.value = null
  } finally {
    resolvingDept.value = false
  }
}

const fetchAll = () => { 
  fetchTable()
  fetchPendingArrivals()
}

watch(
  () => sessionStore.currentUserId,
  () => {
    resolveCurrentUserDept()
  },
  { immediate: true }
)

onMounted(fetchAll)

// ——— 表格视图计算 ———
const getItemDepartmentId = (item: ReagentItem) => {
  const raw = item.cabinet?.department_id
  if (typeof raw !== 'number' || Number.isNaN(raw)) return null
  return raw
}

const isInTeamScope = (item: ReagentItem) => {
  if (viewMode.value !== 'team') return true
  const deptId = getItemDepartmentId(item)
  if (deptId === 0) return true
  if (currentUserDeptId.value === null) return true
  return deptId === currentUserDeptId.value
}

const scopedItems = computed(() => allItems.value.filter((item) => isInTeamScope(item)))

const filteredItems = computed(() => {
  let r = scopedItems.value
  if (statusFilter.value !== '全部') r = r.filter(i => getInventoryDisplayStatus(i.status) === statusFilter.value)
  if (cabinetFilter.value === '普通柜') r = r.filter(i => i.cabinet?.cabinet_type === '普通试剂柜')
  else if (cabinetFilter.value === '管控柜') r = r.filter(i => i.cabinet?.cabinet_type === '易制毒制爆试剂柜')
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    r = r.filter(i =>
      i.uuid.toLowerCase().includes(q) ||
      i.reagent_catalog?.name?.toLowerCase().includes(q) ||
      i.reagent_catalog?.cas_number?.toLowerCase().includes(q) ||
      getCabinetName(i)?.toLowerCase().includes(q) ||
      getItemLocation(i).toLowerCase().includes(q)
    )
  }
  return r
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredItems.value.length / pageSize)))
const paginatedItems = computed(() => filteredItems.value.slice((currentPage.value - 1) * pageSize, currentPage.value * pageSize))

const setStatusFilter = (s: string) => { statusFilter.value = s; currentPage.value = 1 }
const setCabinetFilter = (c: string) => { cabinetFilter.value = c; currentPage.value = 1 }

watch(viewMode, () => {
  currentPage.value = 1
})

// ——— 操作：领用 ———
const openConsumeDialog = (item: ReagentItem) => {
  if (!isControlledItem(item)) {
    toast.error('普通试剂不支持逐次消耗，请直接执行“用尽”')
    return
  }
  consumeDialog.value = { isOpen: true, item, volume: 1, remarks: '' }
}
const submitConsume = async () => {
  if (!consumeDialog.value.item) return
  if (!isControlledItem(consumeDialog.value.item)) {
    toast.error('普通试剂不支持逐次消耗，请直接执行“用尽”')
    consumeDialog.value.isOpen = false
    return
  }
  if (consumeDialog.value.volume <= 0) { toast.error('领用量必须大于0'); return }
  try {
    await axios.put(`/api/reagents/items/${consumeDialog.value.item.uuid}/consume`, {
      consume_volume: Number(consumeDialog.value.volume),
      remarks: consumeDialog.value.remarks
    })
    toast.success('已记录领用消耗')
    consumeDialog.value.isOpen = false
    fetchAll()
  } catch (e: any) {
    toast.error('领用失败: ' + (e.response?.data?.error || '操作失败'))
  }
}

// ——— 操作：核销 ———
const markAsEmpty = async (item: ReagentItem) => {
  if (!confirm(`确认将该瓶 [${item.reagent_catalog?.name}] 标记为已耗尽并核销吗？`)) return
  try {
    await axios.post(`/api/reagents/items/${item.uuid}/deplete`, { remarks: '库存台账执行耗尽核销' })
    toast.success('空瓶已核销回收')
    fetchAll()
  } catch (e: any) {
    toast.error('核销失败: ' + (e.response?.data?.error || '操作失败'))
  }
}

// ——— 辅助函数 ———
const remainingPct = (item: ReagentItem) => {
  if (!item.capacity || item.capacity === 0) return 0
  return Math.round((item.remaining_volume / item.capacity) * 100)
}
const getExpiryInfo = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith('0001-01-01')) return { text: '未知', class: 'text-gray-400' }
  const expiry = new Date(dateStr)
  const year = expiry.getFullYear()
  if (isNaN(year) || year < 1900 || year > 2100) return { text: '未知', class: 'text-gray-400' }
  const diffDays = Math.ceil((expiry.getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  if (diffDays < 0) return { text: `已过期 ${Math.abs(diffDays)} 天`, class: 'text-red-600 font-medium' }
  if (diffDays <= 30) return { text: `${diffDays} 天后到期`, class: 'text-orange-600 font-medium' }
  if (diffDays <= 90) return { text: `${diffDays} 天后到期`, class: 'text-yellow-600' }
  return { text: expiry.toLocaleDateString('zh-CN'), class: 'text-gray-500' }
}
const getStatusLabel = (status: string) => getInventoryDisplayStatus(status)
const getStatusVariant = (status: string): any => getInventoryStatusVariant(status)
const getRemainingColor = (pct: number) => {
  if (pct <= 15) return 'bg-red-500'
  if (pct <= 40) return 'bg-amber-400'
  return 'bg-emerald-500'
}
</script>

<template>
  <div class="space-y-4">
    <!-- 到货提醒 Banner -->
    <div v-if="pendingArrivals.length > 0" 
         class="bg-amber-50 border border-amber-200 rounded-xl p-4 flex items-center justify-between shadow-sm animate-in fade-in slide-in-from-top-2 duration-500">
        <div class="flex items-center gap-3">
            <div class="p-2 bg-amber-100 text-amber-600 rounded-full">
                <AlertTriangle class="h-5 w-5" />
            </div>
            <div>
                <p class="text-amber-900 font-bold tracking-tight">您有 {{ pendingArrivals.length }} 件试剂待入库</p>
                <p class="text-amber-700 text-xs mt-0.5">实物已送至暂存区，请尽快领取并办理入库以进入库存台账</p>
            </div>
        </div>
        <Button @click="emit('switch-to-arrival')" variant="outline" size="sm" class="bg-white border-amber-200 hover:bg-amber-100 text-amber-700 font-bold transition-all">
            立即办理入库 →
        </Button>
    </div>

    <TableSection title="库存台账" description="统一表格模板：团队台账默认显示“当前团队 + 公共库”，全库台账显示全部库存。">
      <template #actions>
        <Button @click="fetchAll" variant="outline" size="sm">刷新台账</Button>
      </template>

      <template #toolbar>
        <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 w-full">
          <div class="relative w-full sm:w-80">
            <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
            <Input v-model="searchQuery" class="pl-9" placeholder="搜索名称/CAS/库位..." />
          </div>
          <div v-if="viewMode === 'team'" class="text-xs text-slate-500 sm:ml-auto">
            <span v-if="resolvingDept">正在识别当前团队...</span>
            <span v-else-if="currentUserDeptId !== null">当前团队 + 公共库，共 {{ scopedItems.length }} 条</span>
            <span v-else>未识别到团队，当前显示全库数据（可继续搜索筛选）</span>
          </div>

          <div class="apple-segmented shrink-0" :class="viewMode === 'table' ? 'sm:ml-auto' : ''">
            <button @click="viewMode='team'" :class="['apple-segmented-btn', viewMode==='team' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">
              团队台账
            </button>
            <button @click="viewMode='table'" :class="['apple-segmented-btn', viewMode==='table' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">
              全库台账
            </button>
          </div>
        </div>
      </template>

    <!-- ================================ 统一表格视图 ================================ -->
    <div>
      <!-- 筛选条 -->
      <div class="mb-3 flex w-full flex-wrap gap-2">
        <div class="apple-segmented">
          <button v-for="s in statusOptions" :key="s" @click="setStatusFilter(s)"
            :class="['apple-segmented-btn', statusFilter === s ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">
            {{ s }}
          </button>
        </div>
        <div class="apple-segmented sm:ml-auto">
          <button v-for="c in cabinetOptions" :key="c" @click="setCabinetFilter(c)"
            :class="['apple-segmented-btn', cabinetFilter === c ? 'apple-segmented-btn-active text-amber-700' : 'apple-segmented-btn-idle']">
            {{ c === '全部' ? '全部柜' : c }}
          </button>
        </div>
      </div>

      <div v-if="isLoadingTable" class="flex justify-center p-8">
        <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
      </div>
      <div v-else-if="filteredItems.length === 0" class="apple-table-empty">暂无匹配的试剂库存记录。</div>
      <LedgerTable v-else :columns="tableColumns">
        <tr v-for="item in paginatedItems" :key="item.uuid" class="border-b border-gray-100 hover:bg-gray-50 group">
          <td class="px-6 py-4 font-medium text-gray-900">
            <button @click="openLifecycleDialog(item.uuid)" class="hover:text-blue-600 hover:underline text-left" title="查看全生命周期档案">
              {{ item.reagent_catalog?.name || '未知' }}
            </button>
            <span class="block text-xs text-gray-400 font-normal">CAS: {{ item.reagent_catalog?.cas_number }}</span>
          </td>
          <td class="px-6 py-4 font-mono text-xs text-blue-600">
            <button @click="openLifecycleDialog(item.uuid)" class="hover:underline flex items-center gap-1" title="查看生命周期档案">
              <FileText class="w-3 h-3" /> <span :title="item.uuid">{{ item.uuid.substring(0,8) }}…</span>
            </button>
          </td>
          <td class="px-6 py-4">
            <div class="min-w-[90px]">
              <div class="mb-1 flex items-center justify-between text-[10px] text-gray-500">
                <span>{{ formatAmount(item.remaining_volume, item.reagent_catalog?.unit, 'ml') }}</span>
                <span>/ {{ formatNumber(item.capacity) }}</span>
              </div>
              <div class="h-1.5 w-full overflow-hidden rounded-full bg-gray-100">
                <div :class="['h-full rounded-full transition-all', getRemainingColor(remainingPct(item))]" :style="{ width: remainingPct(item) + '%' }"></div>
              </div>
            </div>
          </td>
          <td class="px-6 py-4">
            <span v-if="item.cabinet_id > 0" :class="['text-xs px-2 py-0.5 border rounded font-medium', isControlledCabinet(item) ? 'bg-red-50 text-red-700 border-red-200' : 'bg-blue-50 text-blue-600 border-blue-200']">
              {{ getCabinetName(item) }}
            </span>
            <span v-else class="text-xs text-gray-300">—</span>
          </td>
          <td class="px-6 py-4 text-xs text-gray-600">{{ getItemLocation(item) }}</td>
          <td class="px-6 py-4 text-xs text-gray-700">{{ item.batch_number || '系统批次' }}</td>
          <td class="px-6 py-4">
            <Badge :variant="getStatusVariant(item.status)">{{ getStatusLabel(item.status) }}</Badge>
          </td>
          <td class="px-6 py-4">
            <span :class="['text-xs', getExpiryInfo(item.expiry_date).class]">{{ getExpiryInfo(item.expiry_date).text }}</span>
          </td>
          <td class="px-6 py-4 text-right">
            <div class="flex items-center justify-end gap-1.5">
              <Button @click="openLifecycleDialog(item.uuid)" variant="outline" size="sm" class="h-7 px-2 text-[11px]">
                流转单
              </Button>
              <Button
                v-if="isInStorageStatus(item.status) && isControlledItem(item)"
                @click="openConsumeDialog(item)"
                variant="outline"
                size="sm"
                class="h-7 px-2 text-[11px] border-blue-100 text-blue-600 hover:bg-blue-50"
              >
                使用
              </Button>
              <Button
                v-else-if="isInStorageStatus(item.status)"
                @click="markAsEmpty(item)"
                variant="destructive"
                size="sm"
                class="h-7 px-2 text-[11px]"
              >
                用尽
              </Button>
              <span v-else class="text-xs text-slate-400">仅查看</span>
            </div>
          </td>
        </tr>
      </LedgerTable>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="flex items-center justify-between pt-3">
        <div class="text-xs text-gray-500">共 {{ filteredItems.length }} 条，第 {{ currentPage }} / {{ totalPages }} 页</div>
        <div class="flex gap-1">
          <button @click="currentPage = Math.max(1, currentPage - 1)" :disabled="currentPage === 1" class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors">上一页</button>
          <button @click="currentPage = Math.min(totalPages, currentPage + 1)" :disabled="currentPage === totalPages" class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors">下一页</button>
        </div>
      </div>
    </div>
    </TableSection>

    <!-- ================================ 领用弹窗 (Standardized) ================================ -->
    <Dialog :open="consumeDialog.isOpen" size="sm" @close="consumeDialog.isOpen = false">
      <div class="p-6 space-y-6">
        <div>
          <h3 class="text-xl font-bold text-gray-900 tracking-tight">登记管控试剂消耗</h3>
          <p class="text-sm text-gray-500 mt-1" v-if="consumeDialog.item">
            条码: <span class="font-mono text-blue-600 font-medium">#{{ consumeDialog.item.uuid.substring(0,8).toUpperCase() }}</span> — {{ consumeDialog.item.reagent_catalog?.name }}
          </p>
        </div>
        <div class="space-y-4">
          <div class="space-y-2">
            <label class="text-sm font-bold text-gray-700">本次耗用量</label>
            <div class="relative">
              <input type="number" v-model="consumeDialog.volume" min="1" :max="consumeDialog.item?.remaining_volume"
                class="w-full h-11 px-4 bg-gray-50 border border-gray-200 rounded-xl pr-16 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 outline-none transition-all" />
              <span class="absolute right-4 top-3 text-sm font-medium text-gray-400">{{ normalizeUnit(consumeDialog.item?.reagent_catalog?.unit, 'ml') }}</span>
            </div>
            <div class="flex justify-between items-center px-1">
              <span class="text-[11px] text-gray-400">当前余量: {{ formatAmount(consumeDialog.item?.remaining_volume, consumeDialog.item?.reagent_catalog?.unit, 'ml') }}</span>
              <button @click="consumeDialog.volume = consumeDialog.item?.remaining_volume || 0" class="text-[11px] text-blue-600 hover:underline">全部用尽</button>
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-sm font-bold text-gray-700">消耗原因/用途 (选填)</label>
            <textarea v-model="consumeDialog.remarks" rows="2" placeholder="例如：日常实验配置溶液..."
              class="w-full p-4 bg-gray-50 border border-gray-200 rounded-xl text-sm resize-none focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 outline-none transition-all"></textarea>
          </div>
        </div>
      </div>
      <template #footer>
          <Button @click="consumeDialog.isOpen = false" variant="secondary">取消</Button>
          <Button @click="submitConsume" variant="primary">确认扣减</Button>
      </template>
    </Dialog>

    <!-- ================================ 生命周期弹窗 ================================ -->
    <ItemLifecycleDialog 
      :is-open="lifecycleDialog.isOpen" 
      :item-uuid="lifecycleDialog.itemUuid"
      @close="lifecycleDialog.isOpen = false"
      @refresh-needed="fetchAll"
    />

  </div>
</template>
