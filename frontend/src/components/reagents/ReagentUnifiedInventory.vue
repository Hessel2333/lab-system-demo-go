<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { toast } from 'vue-sonner'
import {
  Loader2, MapPin, FlaskConical, Users, Search,
  ChevronDown, ChevronRight, AlertTriangle,
  MinusCircle, Trash2, LayoutList, LayoutGrid
} from 'lucide-vue-next'

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
  reagent_request?: {
    requestor?: { real_name: string }
  }
}

interface TeamGroup {
  department_id: number
  department_name: string
  items: ReagentItem[]
  total_count: number
}

// ——— 视图状态 ———
const viewMode = ref<'table' | 'team'>('team')

// ——— 表格视图状态 ———
const allItems = ref<ReagentItem[]>([])
const isLoadingTable = ref(false)
const searchQuery = ref('')
const statusFilter = ref('全部')
const cabinetFilter = ref('全部')
const currentPage = ref(1)
const pageSize = 15
const statusOptions = ['全部', '在库', '已到货', '已耗尽']
const cabinetOptions = ['全部', '普通柜', '管控柜']
// 辅助：获取柜子显示名称
const getCabinetName = (item: ReagentItem) => item.cabinet?.name ?? (item.cabinet_id > 0 ? `柜#${item.cabinet_id}` : null)
const isControlledCabinet = (item: ReagentItem) => item.cabinet?.cabinet_type === '易制毒制爆试剂柜'
const getItemLocation = (item: ReagentItem) => item.cabinet?.location || item.location || '—'

// ——— 团队视图状态 ———
const allGroups = ref<TeamGroup[]>([])
const isLoadingTeam = ref(false)
const selectedDeptId = ref<number | null>(null)
const expandedCategories = ref<Set<string>>(new Set())

// ——— 消耗弹窗 ———
const consumeDialog = ref({
  isOpen: false,
  item: null as ReagentItem | null,
  volume: 1,
  remarks: ''
})

// ——— 数据获取 ———
const fetchTable = async () => {
  isLoadingTable.value = true
  try {
    const res = await axios.get('/api/reagents/items')
    allItems.value = res.data ?? []
  } catch { /* ignore */ } finally {
    isLoadingTable.value = false
  }
}

const fetchTeam = async () => {
  isLoadingTeam.value = true
  try {
    const res = await axios.get('/api/reagents/team-inventory')
    allGroups.value = (res.data ?? []).filter((g: TeamGroup) => g.department_id >= 0)
    if (allGroups.value.length > 0 && selectedDeptId.value === null) {
      selectedDeptId.value = allGroups.value[0]?.department_id ?? null
    }
  } catch { /* ignore */ } finally {
    isLoadingTeam.value = false
  }
}

const fetchAll = () => { fetchTable(); fetchTeam() }

onMounted(fetchAll)

// ——— 表格视图计算 ———
const filteredItems = computed(() => {
  let r = allItems.value
  if (statusFilter.value !== '全部') r = r.filter(i => i.status === statusFilter.value || getStatusLabel(i.status) === statusFilter.value)
  if (cabinetFilter.value === '普通柜') r = r.filter(i => i.cabinet?.cabinet_type === '普通试剂柜')
  else if (cabinetFilter.value === '管控柜') r = r.filter(i => i.cabinet?.cabinet_type === '易制毒制爆试剂柜')
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    r = r.filter(i =>
      i.uuid.toLowerCase().includes(q) ||
      i.reagent_catalog?.name?.toLowerCase().includes(q) ||
      i.reagent_catalog?.cas_number?.toLowerCase().includes(q) ||
      i.location?.toLowerCase().includes(q)
    )
  }
  return r
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredItems.value.length / pageSize)))
const paginatedItems = computed(() => filteredItems.value.slice((currentPage.value - 1) * pageSize, currentPage.value * pageSize))

const setStatusFilter = (s: string) => { statusFilter.value = s; currentPage.value = 1 }
const setCabinetFilter = (c: string) => { cabinetFilter.value = c; currentPage.value = 1 }

// ——— 团队视图计算 ———
const selectedGroup = computed<TeamGroup | null>(() =>
  selectedDeptId.value === null ? null : allGroups.value.find(g => g.department_id === selectedDeptId.value) ?? null
)
const totalInStock = computed(() => selectedGroup.value?.items?.length || 0)
const groupedByCategory = computed(() => {
  if (!selectedGroup.value?.items) return {}
  return selectedGroup.value.items.reduce((acc, item) => {
    const cat = item.reagent_catalog?.name || '未知品类'
    if (!acc[cat]) acc[cat] = []
    acc[cat].push(item)
    return acc
  }, {} as Record<string, ReagentItem[]>)
})
const categoryCount = computed(() => Object.keys(groupedByCategory.value).length)

const toggleCategory = (key: string) => {
  if (expandedCategories.value.has(key)) expandedCategories.value.delete(key)
  else expandedCategories.value.add(key)
  expandedCategories.value = new Set(expandedCategories.value)
}
const isExpanded = (key: string) => expandedCategories.value.has(key)

// ——— 操作：领用 ———
const openConsumeDialog = (item: ReagentItem) => {
  consumeDialog.value = { isOpen: true, item, volume: 1, remarks: '' }
}
const submitConsume = async () => {
  if (!consumeDialog.value.item) return
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
    await axios.put(`/api/reagents/items/${item.uuid}/status`, { status: '已耗尽', location: item.location })
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
const getStatusLabel = (status: string) => {
  const map: Record<string, string> = { InStorage: '在库', Arrived: '已到货', Used: '已耗尽' }
  return map[status] || status
}
const getStatusColor = (status: string) => {
  const map: Record<string, string> = {
    '在库': 'bg-green-100 text-green-800',
    '已到货': 'bg-blue-100 text-blue-800',
    '已耗尽': 'bg-gray-100 text-gray-500',
  }
  return map[getStatusLabel(status)] || 'bg-gray-100 text-gray-800'
}
const getRemainingColor = (pct: number) => {
  if (pct <= 15) return 'bg-red-500'
  if (pct <= 40) return 'bg-amber-400'
  return 'bg-emerald-500'
}
</script>

<template>
  <div class="space-y-4">
    <!-- 顶部工具栏 -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <!-- 左侧：搜索框（仅表格视图显示） -->
      <div v-if="viewMode === 'table'" class="relative w-72">
        <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
        <input v-model="searchQuery" class="w-full h-9 pl-9 pr-3 text-sm bg-white border border-gray-200 rounded-lg focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none" placeholder="搜索名称/CAS/库位..." />
      </div>
      <div v-else class="text-sm text-gray-500">按团队分组展示，共 {{ allGroups.length }} 个团队</div>

      <!-- 右侧：视图切换 -->
      <div class="flex items-center gap-1 bg-gray-100 p-1 rounded-lg shrink-0">
        <button @click="viewMode='team'" :class="['flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all', viewMode==='team' ? 'bg-white text-blue-700 shadow-sm' : 'text-gray-500 hover:text-gray-700']">
          <LayoutGrid class="w-3.5 h-3.5" /> 团队视图
        </button>
        <button @click="viewMode='table'" :class="['flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all', viewMode==='table' ? 'bg-white text-blue-700 shadow-sm' : 'text-gray-500 hover:text-gray-700']">
          <LayoutList class="w-3.5 h-3.5" /> 全库台账
        </button>
      </div>
    </div>

    <!-- ================================ 团队视图 ================================ -->
    <div v-show="viewMode === 'team'">
      <div v-if="isLoadingTeam" class="flex justify-center py-16">
        <Loader2 class="w-8 h-8 animate-spin text-gray-400" />
      </div>
      <div v-else-if="allGroups.length === 0" class="text-center py-16 text-gray-500">
        <FlaskConical class="w-12 h-12 mx-auto mb-4 text-gray-300" />
        <p class="text-sm">暂无在库试剂数据</p>
      </div>
      <div v-else class="flex gap-5">
        <!-- 左侧：团队选择 -->
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

        <!-- 右侧：品类折叠列表 -->
        <div class="flex-1 min-w-0 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-base font-bold text-gray-900">{{ selectedGroup?.department_name }} · 在库台账</h3>
              <p class="text-xs text-gray-500 mt-0.5">共 <span class="font-semibold">{{ totalInStock }}</span> 瓶，涉及 <span class="font-semibold">{{ categoryCount }}</span> 个品类</p>
            </div>
            <button @click="fetchTeam" class="text-xs text-blue-600 hover:underline">刷新</button>
          </div>

          <div v-for="(items, catalogName) in groupedByCategory" :key="String(catalogName)" class="rounded-xl border border-gray-200 overflow-hidden">
            <button class="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 transition-colors text-left" @click="toggleCategory(String(catalogName))">
              <div class="flex items-center gap-3">
                <component :is="isExpanded(String(catalogName)) ? ChevronDown : ChevronRight" class="w-4 h-4 text-gray-400 shrink-0" />
                <span class="font-semibold text-gray-800 text-sm">{{ catalogName }}</span>
                <span class="text-xs text-gray-400">{{ items[0]?.reagent_catalog?.formula ?? '' }}</span>
                <span v-if="items[0]?.reagent_catalog?.is_controlled" class="text-[10px] px-1.5 py-0.5 rounded bg-red-100 text-red-600 font-medium">管控品</span>
                <span class="text-xs px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded font-medium">{{ items[0]?.reagent_catalog?.category }}</span>
              </div>
              <div class="flex items-center gap-3 text-xs text-gray-500 shrink-0">
                <span>{{ items.length }} 瓶在库</span>
                <span class="flex items-center gap-0.5"><MapPin class="w-3 h-3" />{{ [...new Set(items.map(i => getItemLocation(i)))].join(' / ') }}</span>
              </div>
            </button>

            <div v-if="isExpanded(String(catalogName))" class="divide-y divide-gray-100">
              <div v-for="item in items" :key="item.uuid" class="group flex items-center px-4 py-2.5 bg-white hover:bg-gray-50 gap-3 text-sm">
                <span class="font-mono text-[11px] text-gray-400 w-20 shrink-0">#{{ item.uuid.substring(0,8).toUpperCase() }}</span>
                <!-- 试剂柜标签 -->
                <span v-if="item.cabinet_id > 0" :class="['text-[10px] px-1.5 py-0.5 border rounded shrink-0 font-medium', isControlledCabinet(item) ? 'bg-red-50 text-red-700 border-red-200' : 'bg-blue-50 text-blue-600 border-blue-200']">
                  {{ getCabinetName(item) }}
                </span>
                <span class="flex items-center gap-1 text-xs text-gray-600 w-16 shrink-0"><MapPin class="w-3 h-3 text-gray-400" /> {{ getItemLocation(item) }}</span>
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <div class="w-24 h-1.5 bg-gray-200 rounded-full overflow-hidden shrink-0">
                    <div class="h-full rounded-full transition-all" :class="remainingPct(item) > 50 ? 'bg-green-500' : remainingPct(item) > 20 ? 'bg-yellow-500' : 'bg-red-500'" :style="`width: ${remainingPct(item)}%`"></div>
                  </div>
                  <span class="text-xs text-gray-500 whitespace-nowrap">{{ item.remaining_volume }}/{{ item.capacity }} {{ item.reagent_catalog?.unit }}</span>
                </div>
                <span class="text-xs text-gray-400 font-mono shrink-0 hidden xl:block">{{ item.batch_number }}</span>
                <span class="flex items-center gap-1 text-xs shrink-0 w-24" :class="isNearExpiry(item) ? 'text-orange-600' : 'text-gray-400'">
                  <AlertTriangle v-if="isNearExpiry(item)" class="w-3 h-3" />
                  {{ formatDate(item.expiry_date) }}
                </span>
                <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity ml-auto pl-4 border-l border-gray-100 shrink-0">
                  <button @click.stop="openConsumeDialog(item)" class="text-[11px] px-2 py-1 bg-blue-50 text-blue-600 rounded flex items-center gap-1 hover:bg-blue-100 transition-colors">
                    <MinusCircle class="w-3 h-3" /> 使用
                  </button>
                  <button @click.stop="markAsEmpty(item)" class="text-[11px] px-2 py-1 bg-gray-50 text-gray-400 rounded flex items-center gap-1 hover:bg-red-50 hover:text-red-500 transition-colors">
                    <Trash2 class="w-3 h-3" /> 用尽
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ================================ 表格视图 ================================ -->
    <div v-show="viewMode === 'table'">
      <!-- 筛选条 -->
      <div class="flex flex-wrap gap-2 mb-3">
        <div class="flex gap-1 rounded-lg bg-gray-100 p-1">
          <button v-for="s in statusOptions" :key="s" @click="setStatusFilter(s)"
            :class="['px-3 py-1.5 text-xs font-medium rounded-md transition-all', statusFilter === s ? 'bg-white text-blue-700 shadow-sm' : 'text-gray-500 hover:text-gray-700']">
            {{ s }}
          </button>
        </div>
        <div class="flex gap-1 rounded-lg bg-gray-100 p-1">
          <button v-for="c in cabinetOptions" :key="c" @click="setCabinetFilter(c)"
            :class="['px-3 py-1.5 text-xs font-medium rounded-md transition-all', cabinetFilter === c ? 'bg-white text-amber-700 shadow-sm' : 'text-gray-500 hover:text-gray-700']">
            {{ c === '全部' ? '全部柜' : c === '管控柜' ? '⚠️ 管控柜' : c }}
          </button>
        </div>
      </div>

      <div v-if="isLoadingTable" class="flex justify-center p-8">
        <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
      </div>
      <div v-else-if="filteredItems.length === 0" class="text-center text-gray-500 py-8">暂无匹配的试剂库存记录。</div>
      <div v-else class="overflow-x-auto rounded-lg border border-gray-200">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-gray-700 uppercase bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-4 py-3">试剂名称</th>
              <th class="px-4 py-3">系统条码</th>
              <th class="px-4 py-3">剩余量</th>
              <th class="px-4 py-3">试剂柜</th>
              <th class="px-4 py-3">库位</th>
              <th class="px-4 py-3">申购人</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3">有效期</th>
              <th class="px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in paginatedItems" :key="item.uuid" class="border-b border-gray-100 hover:bg-gray-50 group">
              <td class="px-4 py-3 font-medium text-gray-900">
                {{ item.reagent_catalog?.name || '未知' }}
                <span class="block text-xs text-gray-400 font-normal">CAS: {{ item.reagent_catalog?.cas_number }}</span>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-gray-400">
                <span :title="item.uuid">{{ item.uuid.substring(0,8) }}…</span>
              </td>
              <td class="px-4 py-3">
                <div class="min-w-[90px]">
                  <div class="flex items-center justify-between text-[10px] text-gray-500 mb-1">
                    <span>{{ item.remaining_volume }}{{ item.reagent_catalog?.unit?.replace(/[0-9]/g, '') || 'ml' }}</span>
                    <span>/ {{ item.capacity }}</span>
                  </div>
                  <div class="w-full h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div :class="['h-full rounded-full transition-all', getRemainingColor(remainingPct(item))]" :style="{ width: remainingPct(item) + '%' }"></div>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3">
                <span v-if="item.cabinet_id > 0" :class="['text-xs px-2 py-0.5 border rounded font-medium', isControlledCabinet(item) ? 'bg-red-50 text-red-700 border-red-200' : 'bg-blue-50 text-blue-600 border-blue-200']">
                  {{ getCabinetName(item) }}
                </span>
                <span v-else class="text-xs text-gray-300">—</span>
              </td>
              <td class="px-4 py-3 text-xs text-gray-600">{{ getItemLocation(item) }}</td>
              <td class="px-4 py-3 text-gray-700 text-xs">
                {{ item.reagent_request?.requestor?.real_name || 'System' }}
              </td>
              <td class="px-4 py-3">
                <span :class="['px-2 py-0.5 rounded-full text-xs font-medium', getStatusColor(item.status)]">{{ getStatusLabel(item.status) }}</span>
              </td>
              <td class="px-4 py-3">
                <span :class="getExpiryInfo(item.expiry_date).class">{{ getExpiryInfo(item.expiry_date).text }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity" v-if="item.status === '在库'">
                  <button @click="openConsumeDialog(item)" class="text-[11px] px-2 py-1 bg-blue-50 text-blue-600 rounded flex items-center gap-1 hover:bg-blue-100">
                    <MinusCircle class="w-3 h-3" /> 使用
                  </button>
                  <button @click="markAsEmpty(item)" class="text-[11px] px-2 py-1 bg-gray-50 text-gray-400 rounded flex items-center gap-1 hover:bg-red-50 hover:text-red-500">
                    <Trash2 class="w-3 h-3" /> 用尽
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="flex items-center justify-between pt-3">
        <div class="text-xs text-gray-500">共 {{ filteredItems.length }} 条，第 {{ currentPage }} / {{ totalPages }} 页</div>
        <div class="flex gap-1">
          <button @click="currentPage = Math.max(1, currentPage - 1)" :disabled="currentPage === 1" class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors">上一页</button>
          <button @click="currentPage = Math.min(totalPages, currentPage + 1)" :disabled="currentPage === totalPages" class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors">下一页</button>
        </div>
      </div>
    </div>

    <!-- ================================ 领用弹窗 ================================ -->
    <div v-if="consumeDialog.isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm" @click.self="consumeDialog.isOpen=false">
      <div class="bg-white rounded-2xl w-full max-w-sm shadow-xl p-6 space-y-5">
        <div>
          <h3 class="text-lg font-bold text-gray-900">登记试剂消耗</h3>
          <p class="text-sm text-gray-500 mt-1" v-if="consumeDialog.item">
            #{{ consumeDialog.item.uuid.substring(0,8).toUpperCase() }} — {{ consumeDialog.item.reagent_catalog?.name }}
          </p>
        </div>
        <div class="space-y-4">
          <div class="space-y-1.5">
            <label class="text-sm font-semibold text-gray-700">本次耗用量</label>
            <div class="relative">
              <input type="number" v-model="consumeDialog.volume" min="1" :max="consumeDialog.item?.remaining_volume"
                class="w-full h-10 px-3 bg-gray-50 border border-gray-200 rounded-lg pr-14 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none" />
              <span class="absolute right-3 top-2.5 text-sm text-gray-400">{{ consumeDialog.item?.reagent_catalog?.unit?.replace(/[0-9]/g, '') || 'ml' }}</span>
            </div>
            <p class="text-[11px] text-gray-400 text-right">当前余量: {{ consumeDialog.item?.remaining_volume }}</p>
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-semibold text-gray-700">消耗原因/用途 (选填)</label>
            <textarea v-model="consumeDialog.remarks" rows="2" placeholder="例如：日常实验配置溶液…"
              class="w-full p-3 bg-gray-50 border border-gray-200 rounded-lg text-sm resize-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none"></textarea>
          </div>
        </div>
        <div class="flex gap-3">
          <button @click="consumeDialog.isOpen=false" class="flex-1 py-2 text-sm font-medium text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors">取消</button>
          <button @click="submitConsume" class="flex-1 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 shadow-sm shadow-blue-200 rounded-xl transition-colors">确认扣减</button>
        </div>
      </div>
    </div>
  </div>
</template>
