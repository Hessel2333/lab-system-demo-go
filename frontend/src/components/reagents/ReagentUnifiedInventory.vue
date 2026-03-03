<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import axios from 'axios'
import { toast } from 'vue-sonner'
import {
  Loader2, MapPin, FlaskConical, Users, Search,
  ChevronDown, ChevronRight, AlertTriangle,
  MinusCircle, Trash2, FileText
} from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Dialog from '@/components/ui/Dialog.vue'
import TableSection from '@/components/ui/TableSection.vue'
import Input from '@/components/ui/Input.vue'
import LedgerTable from './LedgerTable.vue'
import ItemLifecycleDialog from '@/components/reagents/ItemLifecycleDialog.vue'
import { formatAmount, formatNumber, formatRatio, normalizeUnit } from '@/lib/quantity'
import { getInventoryDisplayStatus, getInventoryStatusVariant, isInStorageStatus } from '@/lib/reagent-status'
import { fetchUsers } from '@/api/organization'
import { useSessionStore } from '@/stores/session'

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

interface TeamGroup {
  department_id: number
  department_name: string
  items: ReagentItem[]
  total_count: number
}

const sessionStore = useSessionStore()
const viewMode = ref<'table' | 'team'>('table')

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
  { key: 'reagent', label: '试剂名称', class: 'w-[18%]' },
  { key: 'barcode', label: '系统条码', class: 'w-[14%]' },
  { key: 'remaining', label: '剩余量', class: 'w-[14%]' },
  { key: 'cabinet', label: '试剂柜', class: 'w-[14%]' },
  { key: 'location', label: '库位', class: 'w-[8%]' },
  { key: 'batch', label: '批次来源', class: 'w-[12%]' },
  { key: 'status', label: '状态', class: 'w-[8%]' },
  { key: 'expiry', label: '有效期', class: 'w-[8%]' },
  { key: 'actions', label: '操作', align: 'right' as const, class: 'w-[14%] whitespace-nowrap' },
]

const getCabinetName = (item: ReagentItem) => item.cabinet?.name ?? (item.cabinet_id > 0 ? `柜#${item.cabinet_id}` : null)
const isControlledCabinet = (item: ReagentItem) => item.cabinet?.cabinet_type === '易制毒制爆试剂柜'
const getItemLocation = (item: ReagentItem) => item.cabinet?.location || item.location || '—'
const isControlledItem = (item: ReagentItem) => !!item.reagent_catalog?.is_controlled

const splitCabinetLabel = (label?: string | null) => {
  const raw = String(label || '').trim()
  if (!raw) return ['—']
  const parts = raw.split(/[-/·]/).map((s) => s.trim()).filter(Boolean)
  if (parts.length <= 1) return [raw]
  return [parts[0]!, parts.slice(1).join('·')]
}

const splitBatchLabel = (batch?: string | null) => {
  const raw = String(batch || '系统批次').trim()
  if (!raw) return ['系统批次']
  const parts = raw.split('-').map((s) => s.trim()).filter(Boolean)
  if (parts.length <= 1) return [raw]
  return [parts[0]!, parts.slice(1).join('-')]
}

const allGroups = ref<TeamGroup[]>([])
const isLoadingTeam = ref(false)
const selectedDeptId = ref<number | null>(null)
const expandedCategories = ref<Set<string>>(new Set())

const currentUserDeptId = ref<number | null>(null)
const currentUserOrgRole = ref('')
const resolvingDept = ref(false)

const canViewGlobalInventory = computed(() => {
  return sessionStore.currentRole === 'procurement' || currentUserOrgRole.value === 'director'
})

const canViewTeamDistribution = computed(() => canViewGlobalInventory.value)

const inventoryDescription = computed(() => {
  if (canViewTeamDistribution.value) {
    return '支持“团队分布”与“全库台账”双视角：团队分布用于观察跨团队试剂分布，全库台账用于明细检索。'
  }
  return '仅展示本团队与公共库库存，统一台账明细支持搜索、筛选与流转操作。'
})

const consumeDialog = ref({
  isOpen: false,
  item: null as ReagentItem | null,
  volume: 1,
  remarks: ''
})

const lifecycleDialog = ref({
  isOpen: false,
  itemUuid: null as string | null
})
const openLifecycleDialog = (uuid: string) => {
  lifecycleDialog.value = { isOpen: true, itemUuid: uuid }
}

const emit = defineEmits(['switch-to-arrival'])

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
  } catch {
    // ignore
  } finally {
    isLoadingTable.value = false
  }
}

const fetchTeam = async () => {
  if (!canViewTeamDistribution.value) {
    allGroups.value = []
    selectedDeptId.value = null
    return
  }
  isLoadingTeam.value = true
  try {
    const res = await axios.get('/api/reagents/team-inventory')
    const groups = ((res.data ?? []) as TeamGroup[])
      .filter((g) => typeof g.department_id === 'number')
      .sort((a, b) => {
        if (a.department_id === 0) return 1
        if (b.department_id === 0) return -1
        return String(a.department_name || '').localeCompare(String(b.department_name || ''), 'zh-CN')
      })
    allGroups.value = groups
    if (groups.length > 0) {
      const exists = groups.some((g) => g.department_id === selectedDeptId.value)
      if (!exists) selectedDeptId.value = groups[0]?.department_id ?? null
    } else {
      selectedDeptId.value = null
    }
  } catch {
    // ignore
  } finally {
    isLoadingTeam.value = false
  }
}

const resolveCurrentUserDept = async () => {
  resolvingDept.value = true
  try {
    const users = await fetchUsers()
    const me = users.find((user) => user.ID === sessionStore.currentUserId)
    currentUserDeptId.value = me?.department_id ?? null
    currentUserOrgRole.value = String(me?.role || '')
  } catch {
    currentUserDeptId.value = null
    currentUserOrgRole.value = ''
  } finally {
    resolvingDept.value = false
  }
}

const fetchAll = async () => {
  await Promise.all([fetchTable(), fetchPendingArrivals(), fetchTeam()])
}

watch(
  () => sessionStore.currentUserId,
  async () => {
    await resolveCurrentUserDept()
    await fetchAll()
    if (!canViewTeamDistribution.value) {
      viewMode.value = 'table'
    }
  },
  { immediate: true }
)

watch(canViewTeamDistribution, (canView) => {
  if (!canView && viewMode.value === 'team') {
    viewMode.value = 'table'
  }
})

watch(viewMode, () => {
  currentPage.value = 1
})

const getItemDepartmentId = (item: ReagentItem) => {
  const raw = item.cabinet?.department_id
  if (typeof raw !== 'number' || Number.isNaN(raw)) return null
  return raw
}

const shouldApplyTeamScope = computed(() => {
  return !canViewGlobalInventory.value || viewMode.value === 'team'
})

const isInTeamScope = (item: ReagentItem) => {
  if (!shouldApplyTeamScope.value) return true
  const deptId = getItemDepartmentId(item)
  if (deptId === 0) return true
  if (currentUserDeptId.value === null) return false
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

const selectedGroup = computed<TeamGroup | null>(() =>
  selectedDeptId.value === null ? null : allGroups.value.find(g => g.department_id === selectedDeptId.value) ?? null
)
const totalInStock = computed(() => selectedGroup.value?.items?.length || 0)
const groupedByReagentName = computed(() => {
  if (!selectedGroup.value?.items) return {}
  return selectedGroup.value.items.reduce((acc, item) => {
    const key = item.reagent_catalog?.name || '未知试剂'
    if (!acc[key]) acc[key] = []
    acc[key].push(item)
    return acc
  }, {} as Record<string, ReagentItem[]>)
})
const reagentGroupCount = computed(() => Object.keys(groupedByReagentName.value).length)

const toggleCategory = (key: string) => {
  if (expandedCategories.value.has(key)) expandedCategories.value.delete(key)
  else expandedCategories.value.add(key)
  expandedCategories.value = new Set(expandedCategories.value)
}
const isExpanded = (key: string) => expandedCategories.value.has(key)

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
  if (consumeDialog.value.volume <= 0) {
    toast.error('领用量必须大于0')
    return
  }
  try {
    await axios.put(`/api/reagents/items/${consumeDialog.value.item.uuid}/consume`, {
      consume_volume: Number(consumeDialog.value.volume),
      remarks: consumeDialog.value.remarks,
    })
    toast.success('已记录领用消耗')
    consumeDialog.value.isOpen = false
    await fetchAll()
  } catch (e: any) {
    toast.error('领用失败: ' + (e.response?.data?.error || '操作失败'))
  }
}

const markAsEmpty = async (item: ReagentItem) => {
  if (!confirm(`确认将该瓶 [${item.reagent_catalog?.name}] 标记为已耗尽并核销吗？`)) return
  try {
    await axios.post(`/api/reagents/items/${item.uuid}/deplete`, { remarks: '库存台账执行耗尽核销' })
    toast.success('空瓶已核销回收')
    await fetchAll()
  } catch (e: any) {
    toast.error('核销失败: ' + (e.response?.data?.error || '操作失败'))
  }
}

const remainingPct = (item: ReagentItem) => {
  if (!item.capacity || item.capacity === 0) return 0
  return Math.round((item.remaining_volume / item.capacity) * 100)
}

const isNearExpiry = (item: ReagentItem) => {
  if (!item.expiry_date) return false
  const diff = new Date(item.expiry_date).getTime() - Date.now()
  return diff > 0 && diff <= 90 * 24 * 3600 * 1000
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

const formatDate = (d: string) => {
  if (!d || d.startsWith('0001')) return '-'
  const dt = new Date(d)
  if (isNaN(dt.getFullYear()) || dt.getFullYear() < 1900) return '-'
  return dt.toLocaleDateString('zh-CN')
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
    <div
      v-if="pendingArrivals.length > 0"
      class="bg-amber-50 border border-amber-200 rounded-xl p-4 flex items-center justify-between shadow-sm animate-in fade-in slide-in-from-top-2 duration-500"
    >
      <div class="flex items-center gap-3">
        <div class="p-2 bg-amber-100 text-amber-600 rounded-full">
          <AlertTriangle class="h-5 w-5" />
        </div>
        <div>
          <p class="text-amber-900 font-bold tracking-tight">您有 {{ pendingArrivals.length }} 件试剂待入库</p>
          <p class="text-amber-700 text-xs mt-0.5">实物已送至暂存区，请尽快领取并办理入库以进入库存台账</p>
        </div>
      </div>
      <Button
        @click="emit('switch-to-arrival')"
        variant="outline"
        size="sm"
        class="bg-white border-amber-200 hover:bg-amber-100 text-amber-700 font-bold transition-all"
      >
        立即办理入库 →
      </Button>
    </div>

    <TableSection title="库存台账" :description="inventoryDescription">
      <template #actions>
        <Button @click="fetchAll" variant="outline" size="sm">刷新台账</Button>
      </template>

      <template #toolbar>
        <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 w-full">
          <div class="relative w-full sm:w-80">
            <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
            <Input v-model="searchQuery" class="pl-9" placeholder="搜索名称/CAS/库位..." />
          </div>

          <div v-if="!canViewTeamDistribution" class="text-xs text-slate-500 sm:ml-auto">
            <span v-if="resolvingDept">正在识别当前团队...</span>
            <span v-else-if="currentUserDeptId !== null">仅显示本团队 + 公共库</span>
            <span v-else>未识别到团队，当前只显示公共库</span>
          </div>

          <div v-if="canViewTeamDistribution" class="apple-segmented shrink-0 sm:ml-auto">
            <button @click="viewMode='table'" :class="['apple-segmented-btn', viewMode==='table' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">
              全库台账
            </button>
            <button @click="viewMode='team'" :class="['apple-segmented-btn', viewMode==='team' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">
              团队分布
            </button>
          </div>
        </div>
      </template>

      <div v-if="canViewTeamDistribution && viewMode === 'team'">
        <div v-if="isLoadingTeam" class="flex justify-center py-16">
          <Loader2 class="w-8 h-8 animate-spin text-gray-400" />
        </div>
        <div v-else-if="allGroups.length === 0" class="apple-table-empty text-gray-500">
          <FlaskConical class="w-12 h-12 mx-auto mb-4 text-gray-300" />
          <p class="text-sm">暂无在库试剂数据</p>
        </div>
        <div v-else class="flex gap-5">
          <div class="w-44 shrink-0 space-y-1.5">
            <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider px-2 mb-2">团队列表</p>
            <button
              v-for="group in allGroups"
              :key="group.department_id"
              @click="selectedDeptId = group.department_id"
              :class="['w-full text-left px-3 py-2.5 rounded-lg text-sm font-medium transition-all flex items-center justify-between', selectedDeptId === group.department_id ? 'bg-blue-600 text-white shadow-sm' : 'text-gray-700 hover:bg-gray-100']"
            >
              <span class="flex items-center gap-2">
                <Users class="w-3.5 h-3.5 shrink-0" />{{ group.department_name }}
              </span>
              <span :class="['text-xs px-1.5 py-0.5 rounded-full font-normal', selectedDeptId === group.department_id ? 'bg-white/20' : 'bg-gray-200 text-gray-500']">{{ group.total_count }}</span>
            </button>
          </div>

          <div class="flex-1 min-w-0 space-y-4">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-base font-bold text-gray-900">{{ selectedGroup?.department_name }} · 团队试剂分布</h3>
                <p class="text-xs text-gray-500 mt-0.5">共 <span class="font-semibold">{{ totalInStock }}</span> 瓶，涉及 <span class="font-semibold">{{ reagentGroupCount }}</span> 种试剂</p>
              </div>
              <button @click="fetchTeam" class="text-xs text-blue-600 hover:underline">刷新</button>
            </div>

            <div v-for="(items, reagentName) in groupedByReagentName" :key="String(reagentName)" class="rounded-xl border border-gray-200 overflow-hidden">
              <button class="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 transition-colors text-left" @click="toggleCategory(String(reagentName))">
                <div class="flex items-center gap-3">
                  <component :is="isExpanded(String(reagentName)) ? ChevronDown : ChevronRight" class="w-4 h-4 text-gray-400 shrink-0" />
                  <span class="font-bold text-gray-900 text-sm tracking-tight">{{ reagentName }}</span>
                  <span class="text-xs text-gray-400 font-mono">{{ items[0]?.reagent_catalog?.formula ?? '' }}</span>
                  <Badge v-if="items[0]?.reagent_catalog?.is_controlled" variant="destructive" class="px-1.5 h-4.5 text-[10px]">管控品</Badge>
                  <Badge variant="info" class="px-1.5 h-4.5 text-[10px]">{{ items[0]?.reagent_catalog?.category }}</Badge>
                </div>
                <div class="flex items-center gap-3 text-xs text-gray-500 shrink-0">
                  <span>{{ items.length }} 瓶在库</span>
                  <span class="flex items-center gap-0.5"><MapPin class="w-3 h-3" />{{ [...new Set(items.map(i => getItemLocation(i)))].join(' / ') }}</span>
                </div>
              </button>

              <div v-if="isExpanded(String(reagentName))" class="divide-y divide-gray-100">
                <div v-for="item in items" :key="item.uuid" class="group flex items-center px-4 py-2.5 bg-white hover:bg-gray-50 gap-3 text-sm">
                  <button @click="openLifecycleDialog(item.uuid)" class="font-mono text-[11px] text-blue-600 hover:text-blue-800 hover:underline flex items-center gap-1 w-24 shrink-0" title="查看生命周期档案">
                    <FileText class="w-3 h-3"/> {{ item.uuid.substring(0,8).toUpperCase() }}
                  </button>
                  <span v-if="item.cabinet_id > 0" :class="['text-[10px] px-1.5 py-0.5 border rounded shrink-0 font-medium', isControlledCabinet(item) ? 'bg-red-50 text-red-700 border-red-200' : 'bg-blue-50 text-blue-600 border-blue-200']">
                    {{ getCabinetName(item) }}
                  </span>
                  <span class="flex items-center gap-1 text-xs text-gray-600 w-20 shrink-0"><MapPin class="w-3 h-3 text-gray-400" /> {{ getItemLocation(item) }}</span>
                  <div class="flex items-center gap-2 flex-1 min-w-0">
                    <div class="w-24 h-1.5 bg-gray-200 rounded-full overflow-hidden shrink-0">
                      <div class="h-full rounded-full transition-all" :class="remainingPct(item) > 50 ? 'bg-green-500' : remainingPct(item) > 20 ? 'bg-yellow-500' : 'bg-red-500'" :style="`width: ${remainingPct(item)}%`"></div>
                    </div>
                    <span class="text-xs text-gray-500 whitespace-nowrap">{{ formatRatio(item.remaining_volume, item.capacity, item.reagent_catalog?.unit, 'ml') }}</span>
                  </div>
                  <span class="text-xs text-gray-400 font-mono shrink-0 hidden xl:block">{{ item.batch_number }}</span>
                  <span class="flex items-center gap-1 text-xs shrink-0 w-24" :class="isNearExpiry(item) ? 'text-orange-600' : 'text-gray-400'">
                    <AlertTriangle v-if="isNearExpiry(item)" class="w-3 h-3" />
                    {{ formatDate(item.expiry_date) }}
                  </span>
                  <div class="flex items-center gap-2 ml-auto pl-4 border-l border-gray-100 shrink-0">
                    <Button @click.stop="openLifecycleDialog(item.uuid)" variant="outline" size="sm" class="h-7 px-2 text-[11px]">
                      流转单
                    </Button>
                    <Button
                      v-if="isInStorageStatus(item.status) && isControlledItem(item)"
                      @click.stop="openConsumeDialog(item)"
                      variant="outline"
                      size="sm"
                      class="h-7 px-2 text-[11px] border-blue-100 text-blue-600 hover:bg-blue-50"
                    >
                      <MinusCircle class="w-3.5 h-3.5 mr-1" /> 使用
                    </Button>
                    <Button
                      v-else-if="isInStorageStatus(item.status)"
                      @click.stop="markAsEmpty(item)"
                      variant="destructive"
                      size="sm"
                      class="h-7 px-2 text-[11px]"
                    >
                      <Trash2 class="w-3.5 h-3.5 mr-1" /> 用尽
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else>
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
              <span v-if="item.cabinet_id > 0" :class="['inline-flex max-w-[170px] items-center px-2 py-0.5 border rounded font-medium text-xs', isControlledCabinet(item) ? 'bg-red-50 text-red-700 border-red-200' : 'bg-blue-50 text-blue-600 border-blue-200']" :title="getCabinetName(item) || ''">
                <span class="inline-flex flex-col leading-[1.2]">
                  <span
                    v-for="(line, idx) in splitCabinetLabel(getCabinetName(item))"
                    :key="`${item.uuid}-cab-${idx}`"
                    class="max-w-[154px] truncate whitespace-nowrap"
                  >
                    {{ line }}
                  </span>
                </span>
              </span>
              <span v-else class="text-xs text-gray-300">—</span>
            </td>
            <td class="px-6 py-4 text-xs text-gray-600 whitespace-nowrap">{{ getItemLocation(item) }}</td>
            <td class="px-6 py-4 text-xs text-gray-700">
              <span class="inline-flex max-w-[150px] flex-col leading-[1.2] align-middle" :title="item.batch_number || '系统批次'">
                <span
                  v-for="(line, idx) in splitBatchLabel(item.batch_number || '系统批次')"
                  :key="`${item.uuid}-batch-${idx}`"
                  class="max-w-[150px] truncate whitespace-nowrap"
                >
                  {{ line }}
                </span>
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <Badge :variant="getStatusVariant(item.status)" class="min-w-[56px] justify-center">{{ getStatusLabel(item.status) }}</Badge>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
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

        <div v-if="totalPages > 1" class="flex items-center justify-between pt-3">
          <div class="text-xs text-gray-500">共 {{ filteredItems.length }} 条，第 {{ currentPage }} / {{ totalPages }} 页</div>
          <div class="flex gap-1">
            <button @click="currentPage = Math.max(1, currentPage - 1)" :disabled="currentPage === 1" class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors">上一页</button>
            <button @click="currentPage = Math.min(totalPages, currentPage + 1)" :disabled="currentPage === totalPages" class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors">下一页</button>
          </div>
        </div>
      </div>
    </TableSection>

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

    <ItemLifecycleDialog
      :is-open="lifecycleDialog.isOpen"
      :item-uuid="lifecycleDialog.itemUuid"
      @close="lifecycleDialog.isOpen = false"
      @refresh-needed="fetchAll"
    />
  </div>
</template>
