<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Loader2, Search } from 'lucide-vue-next'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import axios from 'axios'
import { formatAmount, formatNumber, normalizeUnit } from '@/lib/quantity'

const items = ref<any[]>([])
const isLoading = ref(true)
const searchQuery = ref('')
const statusFilter = ref('全部')
const currentPage = ref(1)
const pageSize = 15

const statusOptions = ['全部', '在库', '已到货', '已耗尽']

const fetchItems = async () => {
    isLoading.value = true
    try {
        const res = await axios.get('/api/reagents/items')
        items.value = res.data
    } catch (error) {
        console.error("Failed to fetch inventory items", error)
    } finally {
        isLoading.value = false
    }
}

const filteredItems = computed(() => {
    let result = items.value
    // Status filter
    if (statusFilter.value !== '全部') {
        result = result.filter(item => {
            const label = getStatusLabel(item.status)
            return label === statusFilter.value
        })
    }
    // Search filter
    if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase()
        result = result.filter(item => 
            item.uuid.toLowerCase().includes(q) ||
            item.reagent_catalog?.name?.toLowerCase().includes(q) ||
            item.reagent_catalog?.cas_number?.toLowerCase().includes(q) ||
            item.location?.toLowerCase().includes(q)
        )
    }
    return result
})

// Pagination
const totalPages = computed(() => Math.max(1, Math.ceil(filteredItems.value.length / pageSize)))
const paginatedItems = computed(() => {
    const start = (currentPage.value - 1) * pageSize
    return filteredItems.value.slice(start, start + pageSize)
})

// Reset page when filters change
const setStatusFilter = (s: string) => {
    statusFilter.value = s
    currentPage.value = 1
}

const getStatusLabel = (status: string) => {
    const map: Record<string, string> = {
        'InStorage': '在库',
        'Arrived': '已到货',
        'Used': '已耗尽',
        'Expired': '已过期',
        'Disposed': '已处理'
    }
    return map[status] || status;
}

const getStatusColor = (status: string) => {
    const label = getStatusLabel(status)
    const map: Record<string, string> = {
        '在库': 'bg-green-100 text-green-800',
        '已到货': 'bg-blue-100 text-blue-800',
        '已耗尽': 'bg-gray-100 text-gray-500',
        '已过期': 'bg-red-100 text-red-800',
    }
    return map[label] || 'bg-gray-100 text-gray-800'
}

const getExpiryInfo = (dateStr: string) => {
    if (!dateStr || dateStr.startsWith('0001-01-01')) return { text: '未知', class: 'text-gray-400' }
    
    // 如果年份极小或者极大，也认为是无效的
    const expiry = new Date(dateStr)
    const year = expiry.getFullYear()
    if (isNaN(year) || year < 1900 || year > 2100) return { text: '未知', class: 'text-gray-400' }
    
    const now = new Date()
    const diffDays = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    if (diffDays < 0) return { text: `已过期 ${Math.abs(diffDays)} 天`, class: 'text-red-600 font-medium' }
    if (diffDays <= 30) return { text: `${diffDays} 天后过期`, class: 'text-orange-600 font-medium' }
    if (diffDays <= 90) return { text: `${diffDays} 天后过期`, class: 'text-yellow-600' }
    return { text: expiry.toLocaleDateString('zh-CN'), class: 'text-gray-500' }
}

const getRemainingPercent = (item: any) => {
    if (!item.capacity || item.capacity === 0) return -1
    return Math.round((item.remaining_volume / item.capacity) * 100)
}
const getRemainingColor = (pct: number) => {
    if (pct <= 15) return 'bg-red-500'
    if (pct <= 40) return 'bg-amber-400'
    return 'bg-emerald-500'
}

onMounted(() => {
    fetchItems()
})
</script>

<template>
  <Card>
    <div class="p-6 space-y-4">
      <!-- Toolbar: Search + Status Filters -->
      <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
          <div class="relative w-72">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
              <Input v-model="searchQuery" class="pl-9" placeholder="搜索条形码、名称或存放位置..." />
          </div>
          <div class="apple-segmented">
              <button
                v-for="s in statusOptions"
                :key="s"
                @click="setStatusFilter(s)"
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
      
      <div v-if="isLoading" class="flex justify-center p-8">
        <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
      </div>
      <div v-else-if="filteredItems.length === 0" class="text-center text-gray-500 py-8">
        暂无匹配的试剂库存记录。
      </div>
      <div v-else class="apple-table-wrap">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3">试剂名称</th>
              <th scope="col" class="px-6 py-3">系统条形码</th>
              <th scope="col" class="px-6 py-3">剩余量</th>
              <th scope="col" class="px-6 py-3">申购人</th>
              <th scope="col" class="px-6 py-3">状态</th>
              <th scope="col" class="px-6 py-3">存放位置</th>
              <th scope="col" class="px-6 py-3">有效期</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in paginatedItems" :key="item.uuid" class="bg-white border-b hover:bg-gray-50">
              <td class="px-6 py-4 font-medium text-gray-900">
                  {{ item.reagent_catalog?.name || '未知' }}
                   <span class="block text-xs text-gray-500 font-normal">CAS: {{ item.reagent_catalog?.cas_number }}</span>
              </td>
              <td class="px-6 py-4 font-mono text-xs text-gray-500">
                  <span :title="item.uuid" class="cursor-help">{{ item.uuid.substring(0, 8) }}...</span>
              </td>
              <td class="px-6 py-4">
                  <div v-if="getRemainingPercent(item) >= 0" class="min-w-[90px]">
                    <div class="flex items-center justify-between text-[10px] text-gray-500 mb-1">
                      <span>{{ formatAmount(item.remaining_volume, item.reagent_catalog?.unit, 'ml') }}</span>
                      <span>/ {{ formatNumber(item.capacity) }}</span>
                    </div>
                    <div class="w-full h-1.5 bg-gray-100 rounded-full overflow-hidden">
                      <div :class="['h-full rounded-full transition-all', getRemainingColor(getRemainingPercent(item))]" :style="{ width: getRemainingPercent(item) + '%' }"></div>
                    </div>
                  </div>
                  <span v-else class="text-xs text-gray-400">{{ normalizeUnit(item.reagent_catalog?.unit, '--') }}</span>
              </td>
              <td class="px-6 py-4 text-gray-700">
                  <span v-if="item.reagent_request?.requestor?.real_name" class="inline-flex items-center gap-1">
                      {{ item.reagent_request.requestor.real_name }}
                  </span>
                  <span v-else class="text-gray-400">System</span>
              </td>
              <td class="px-6 py-4">
                  <span :class="['px-2 py-1 rounded-full text-xs font-medium', getStatusColor(item.status)]">
                      {{ getStatusLabel(item.status) }}
                  </span>
              </td>
              <td class="px-6 py-4">{{ item.location || '--' }}</td>
              <td class="px-6 py-4">
                  <span :class="getExpiryInfo(item.expiry_date).class">
                      {{ getExpiryInfo(item.expiry_date).text }}
                  </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-between pt-2">
          <div class="text-xs text-gray-500">
              共 {{ filteredItems.length }} 条，第 {{ currentPage }} / {{ totalPages }} 页
          </div>
          <div class="flex gap-1">
              <button
                @click="currentPage = Math.max(1, currentPage - 1)"
                :disabled="currentPage === 1"
                class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors"
              >
                上一页
              </button>
              <button
                @click="currentPage = Math.min(totalPages, currentPage + 1)"
                :disabled="currentPage === totalPages"
                class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-200 disabled:opacity-40 hover:bg-gray-50 transition-colors"
              >
                下一页
              </button>
          </div>
      </div>
    </div>
  </Card>
</template>
