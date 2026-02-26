<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { Loader2, MapPin, FlaskConical, Users, ChevronDown, ChevronRight, AlertTriangle, MinusCircle, Trash2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { formatAmount, formatRatio, normalizeUnit } from '@/lib/quantity'

interface ReagentItem {
  uuid: string
  reagent_catalog_id: number
  location: string
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

const allGroups = ref<TeamGroup[]>([])
const selectedDeptId = ref<number | null>(null)
const isLoading = ref(true)
const expandedCategories = ref<Set<string>>(new Set())

const fetchInventory = async () => {
  isLoading.value = true
  try {
    const res = await axios.get('/api/reagents/team-inventory')
    // 原来是 g.department_id > 0，这里改为 >= 0，从而显示 "公共库" (ID 0)
    allGroups.value = (res.data ?? []).filter(
      (g: TeamGroup) => g.department_id >= 0
    )
    // 默认选第一个团队
    if (allGroups.value.length > 0 && selectedDeptId.value === null) {
      selectedDeptId.value = allGroups.value[0]?.department_id ?? null
    }
  } catch (e) {
    console.error('Failed to fetch team inventory', e)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchInventory()
})

const selectedGroup = computed<TeamGroup | null>(() => {
  // Fix 2: Explicitly check for null, because 0 is falsy
  if (selectedDeptId.value === null) return null
  return allGroups.value.find(g => g.department_id === selectedDeptId.value) ?? null
})

const totalInStock = computed(() => {
  return selectedGroup.value?.items?.length || 0
})

const groupedByCategory = computed(() => {
  if (!selectedGroup.value?.items) return {}
  return selectedGroup.value.items.reduce((acc, item) => {
    const catName = item.reagent_catalog?.name || '未知品类'
    if (!acc[catName]) acc[catName] = []
    acc[catName].push(item)
    return acc
  }, {} as Record<string, ReagentItem[]>)
})

const categoryCount = computed(() => Object.keys(groupedByCategory.value).length)

const toggleCategory = (key: string) => {
  if (expandedCategories.value.has(key)) {
    expandedCategories.value.delete(key)
  } else {
    expandedCategories.value.add(key)
  }
  // Trigger reactivity
  expandedCategories.value = new Set(expandedCategories.value)
}

const isExpanded = (key: string) => expandedCategories.value.has(key)

/** 剩余量百分比 */
const remainingPct = (item: ReagentItem) => {
  if (!item.capacity || item.capacity === 0) return 0
  return Math.round((item.remaining_volume / item.capacity) * 100)
}

// 领用消耗弹窗状态
const consumeDialog = ref({
  isOpen: false,
  item: null as ReagentItem | null,
  volume: 1,
  remarks: ''
})

const openConsumeDialog = (item: ReagentItem) => {
  consumeDialog.value = {
    isOpen: true,
    item,
    volume: 1,
    remarks: ''
  }
}

const submitConsume = async () => {
  if (!consumeDialog.value.item) return
  if (consumeDialog.value.volume <= 0) {
    toast.error('领用量必须大于0')
    return
  }
  
  try {
    const vol = Number(consumeDialog.value.volume)
    await axios.put(`/api/reagents/items/${consumeDialog.value.item.uuid}/consume`, {
      consume_volume: vol,
      remarks: consumeDialog.value.remarks
    })
    toast.success('已记录领用消耗')
    consumeDialog.value.isOpen = false
    fetchInventory() // 刷新库存
  } catch (error: any) {
    const errMsg = error.response?.data?.error || '操作失败'
    toast.error('领用失败: ' + errMsg)
  }
}

const markAsEmpty = async (item: ReagentItem) => {
  if (!confirm(`确认将该瓶 [${item.reagent_catalog?.name}] 标记为已耗尽并核销吗？`)) return
  try {
    await axios.put(`/api/reagents/items/${item.uuid}/status`, {
      status: '已耗尽',
      location: item.location
    })
    toast.success('空瓶已核销回收')
    fetchInventory()
  } catch (error: any) {
    const errMsg = error.response?.data?.error || '操作失败'
    toast.error('核销失败: ' + errMsg)
  }
}

/** 到期预警：距今 ≤ 90 天 */
const isNearExpiry = (item: ReagentItem) => {
  if (!item.expiry_date) return false
  const diff = new Date(item.expiry_date).getTime() - Date.now()
  return diff > 0 && diff <= 90 * 24 * 3600 * 1000
}

const formatDate = (d: string) => {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('zh-CN')
}
</script>

<template>
  <div class="space-y-4">
    <!-- 加载中 -->
    <div v-if="isLoading" class="flex justify-center items-center py-20">
      <Loader2 class="w-8 h-8 animate-spin text-gray-400" />
    </div>

    <template v-else>
      <!-- 无数据提示 -->
      <div v-if="allGroups.length === 0" class="text-center py-16 text-gray-500">
        <FlaskConical class="w-12 h-12 mx-auto mb-4 text-gray-300" />
        <p class="text-sm">暂无在库试剂数据</p>
        <p class="text-xs text-gray-400 mt-1">请先通过采购流程将试剂入库，或联系管理员初始化演示数据</p>
      </div>

      <div v-else class="flex gap-5">

        <!-- 左侧：团队选择器 -->
        <div class="w-44 shrink-0 space-y-1.5">
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wider px-2 mb-2">团队列表</p>
          <button
            v-for="group in allGroups"
            :key="group.department_id"
            @click="selectedDeptId = group.department_id"
            :class="[
              'w-full text-left px-3 py-2.5 rounded-lg text-sm font-medium transition-all flex items-center justify-between',
              selectedDeptId === group.department_id
                ? 'bg-blue-600 text-white shadow-sm'
                : 'text-gray-700 hover:bg-gray-100'
            ]"
          >
            <span class="flex items-center gap-2">
              <Users class="w-3.5 h-3.5 shrink-0" />
              {{ group.department_name }}
            </span>
            <span :class="['text-xs px-1.5 py-0.5 rounded-full font-normal',
              selectedDeptId === group.department_id ? 'bg-white/20' : 'bg-gray-200 text-gray-500']">
              {{ group.total_count }}
            </span>
          </button>
        </div>

        <!-- 右侧：库存台账 -->
        <div class="flex-1 min-w-0 space-y-4">

          <!-- 统计头 -->
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-base font-bold text-gray-900">{{ selectedGroup?.department_name }} · 在库台账</h3>
              <p class="text-xs text-gray-500 mt-0.5">共 <span class="font-semibold text-gray-800">{{ totalInStock }}</span> 瓶，涉及 <span class="font-semibold text-gray-800">{{ categoryCount }}</span> 个品类</p>
            </div>
            <button @click="fetchInventory" class="text-xs text-blue-600 hover:underline">刷新</button>
          </div>

          <!-- 按品类折叠的试剂列表 -->
          <div v-for="(items, catalogName) in groupedByCategory" :key="String(catalogName)" class="rounded-xl border border-gray-200 overflow-hidden">

            <!-- 品类折叠头 -->
            <button
              class="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 transition-colors text-left"
              @click="toggleCategory(String(catalogName))"
            >
              <div class="flex items-center gap-3">
                <component :is="isExpanded(String(catalogName)) ? ChevronDown : ChevronRight" class="w-4 h-4 text-gray-400 shrink-0" />
                <span class="font-semibold text-gray-800 text-sm">{{ catalogName }}</span>
                <span class="text-xs text-gray-500 font-normal">
                  {{ items[0]?.reagent_catalog?.formula ?? '' }}
                </span>
                <span v-if="items[0]?.reagent_catalog?.is_controlled"
                      class="text-[10px] px-1.5 py-0.5 rounded bg-red-100 text-red-600 font-medium">管控品</span>
                <span class="text-xs px-1.5 py-0.5 bg-blue-100 text-blue-700 rounded font-medium">{{ items[0]?.reagent_catalog?.category }}</span>
              </div>
              <div class="flex items-center gap-3 text-xs text-gray-500 shrink-0">
                <span>{{ items.length }} 瓶在库</span>
                <span class="flex items-center gap-0.5"><MapPin class="w-3 h-3" />{{ [...new Set(items.map(i => i.location))].join(' / ') }}</span>
              </div>
            </button>

            <!-- 展开内容：每瓶明细 -->
            <div v-if="isExpanded(String(catalogName))" class="divide-y divide-gray-100">
              <div
                v-for="item in items"
                :key="item.uuid"
                class="group flex items-center px-4 py-2.5 bg-white hover:bg-gray-50 gap-4 text-sm"
              >
                <!-- 条码 ID -->
                <span class="font-mono text-[11px] text-gray-400 w-20 shrink-0">
                  #{{ item.uuid.substring(0, 8).toUpperCase() }}
                </span>

                <!-- 库位 -->
                <span class="flex items-center gap-1 text-xs text-gray-600 w-16 shrink-0">
                  <MapPin class="w-3 h-3 text-gray-400" /> {{ item.location }}
                </span>

                <!-- 剩余量进度条 -->
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <div class="w-24 h-1.5 bg-gray-200 rounded-full overflow-hidden shrink-0">
                    <div
                      class="h-full rounded-full transition-all"
                      :class="remainingPct(item) > 50 ? 'bg-green-500' : remainingPct(item) > 20 ? 'bg-yellow-500' : 'bg-red-500'"
                      :style="`width: ${remainingPct(item)}%`"
                    ></div>
                  </div>
                  <span class="text-xs text-gray-500 whitespace-nowrap">
                    {{ formatRatio(item.remaining_volume, item.capacity, item.reagent_catalog?.unit, 'ml') }}
                  </span>
                </div>

                <!-- 批次号 -->
                <span class="text-xs text-gray-400 font-mono shrink-0 hidden xl:block">{{ item.batch_number }}</span>

                <!-- 到期日 + 预警 -->
                <span class="flex items-center gap-1 text-xs shrink-0 w-24"
                      :class="isNearExpiry(item) ? 'text-orange-600' : 'text-gray-400'">
                  <AlertTriangle v-if="isNearExpiry(item)" class="w-3 h-3" />
                  {{ formatDate(item.expiry_date) }}
                </span>

                <!-- 快捷操作区 (仅自己团队或授权时可用) -->
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
    </template>
    
    <!-- 领用耗用弹窗 -->
    <div v-if="consumeDialog.isOpen" class="apple-modal-backdrop" @click.self="consumeDialog.isOpen=false">
      <div class="apple-modal-panel max-w-sm overflow-hidden p-6 space-y-6">
        <div>
          <h3 class="text-lg font-bold text-gray-900">登记试剂消耗</h3>
          <p class="text-sm text-gray-500 mt-1" v-if="consumeDialog.item">
            #{{ consumeDialog.item.uuid.substring(0,8).toUpperCase() }} - {{ consumeDialog.item.reagent_catalog?.name }}
          </p>
        </div>
        
        <div class="space-y-4">
          <div class="space-y-2">
            <label class="text-sm font-semibold text-gray-700">本次耗用量</label>
            <div class="relative">
              <input type="number" v-model="consumeDialog.volume" min="1" :max="consumeDialog.item?.remaining_volume" class="w-full h-10 px-3 bg-gray-50 border border-gray-200 rounded-lg pr-12 focus:border-blue-500 focus:ring-1 focus:ring-blue-500" />
              <span class="absolute right-3 top-2.5 text-sm text-gray-400">{{ normalizeUnit(consumeDialog.item?.reagent_catalog?.unit, 'ml') }}</span>
            </div>
            <p class="text-[11px] text-gray-400 text-right mt-1">当前余量: {{ formatAmount(consumeDialog.item?.remaining_volume, consumeDialog.item?.reagent_catalog?.unit, 'ml') }}</p>
          </div>

          <div class="space-y-2">
            <label class="text-sm font-semibold text-gray-700">消耗原因/用途 (选填)</label>
            <textarea v-model="consumeDialog.remarks" rows="2" placeholder="例如：日常实验配置溶液..." class="w-full p-3 bg-gray-50 border border-gray-200 rounded-lg text-sm resize-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"></textarea>
          </div>
        </div>

        <div class="flex gap-3 pt-2">
          <button @click="consumeDialog.isOpen=false" class="flex-1 py-2 text-sm font-medium text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors">取消</button>
          <button @click="submitConsume" class="flex-1 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 shadow-sm shadow-blue-200 rounded-xl transition-colors">确认扣减</button>
        </div>
      </div>
    </div>
  </div>
</template>
