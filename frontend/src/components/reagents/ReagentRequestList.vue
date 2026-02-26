<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Loader2, Search, Package } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import axios from 'axios'
import RequestProgressDialog from './RequestProgressDialog.vue'

const requests = ref<any[]>([])
const isLoading = ref(true)
const isProgressOpen = ref(false)
const selectedRequest = ref<any>(null)
const searchQuery = ref('')
const statusFilter = ref('全部')

const statusOptions = ['全部', '待审批', '待采购', '已接单', '已驳回']

const props = defineProps({
    role: { type: String, default: 'researcher' }
})

const fetchRequests = async () => {
    isLoading.value = true
    try {
        const res = await axios.get('/api/reagents/requests')
        requests.value = res.data
    } catch (error) {
        console.error("Failed to fetch requests", error)
    } finally {
        isLoading.value = false
    }
    // 采购角色批量查询库存状态
    if (props.role === 'procurement') fetchStockForRequests()
}

// 库存状态缓存（按 catalog_id 缓存避免重复查询）
const stockMap = ref<Record<number, any>>({})
const fetchStockForRequests = async () => {
    const seen = new Set<number>()
    for (const req of requests.value) {
        const catId = req.reagent_catalog_id
        const cas = req.reagent_catalog?.cas_number
        if (!cas || seen.has(catId)) continue
        seen.add(catId)
        try {
            const res = await axios.get('/api/reagents/stock-check', { params: { cas_number: cas } })
            stockMap.value[catId] = res.data
        } catch { /* ignore */ }
    }
}

const filteredRequests = computed(() => {
    let result = requests.value
    if (statusFilter.value !== '全部') {
        result = result.filter(r => r.status === statusFilter.value)
    }
    if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase()
        result = result.filter(r =>
            String(r.id).includes(q) ||
            r.reagent_catalog?.name?.toLowerCase().includes(q) ||
            r.reagent_catalog?.cas_number?.toLowerCase().includes(q) ||
            r.requestor?.real_name?.toLowerCase().includes(q)
        )
    }
    // 最新提交（最大 ID）排前面
    return [...result].sort((a, b) => b.id - a.id)
})

const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const toast = (msg: string, type: 'success' | 'error' = 'success') => {
    toastMessage.value = msg
    toastType.value = type
    showToast.value = true
    setTimeout(() => { showToast.value = false }, 3000)
}

// --- BPM-A: 采购审批弹窗状态 ---
const orderDialogOpen = ref(false)
const orderDialogReqId = ref<number | null>(null)
const orderDialogReqName = ref('')
const orderReference = ref('')
const orderSubmitting = ref(false)

const openOrderDialog = (req: any) => {
    orderDialogReqId.value = req.id
    orderDialogReqName.value = req.reagent_catalog?.name || '未知试剂'
    orderReference.value = ''
    orderDialogOpen.value = true
}

const confirmApproveAndOrder = async () => {
    if (!orderDialogReqId.value) return
    orderSubmitting.value = true
    try {
        await axios.post(`/api/reagents/requests/${orderDialogReqId.value}/approve`, {
            order_reference: orderReference.value
        })
        toast('审批通过！系统已记录采购动作。')
        orderDialogOpen.value = false
        fetchRequests()
    } catch (error) {
        toast('审批失败，请重试。', 'error')
    } finally {
        orderSubmitting.value = false
    }
}
const leaderApprovingId = ref<number | null>(null)
const leaderApprove = async (id: number, approved: boolean) => {
    leaderApprovingId.value = id
    try {
        await axios.post(`/api/reagents/requests/${id}/leader-approve`, {
            approved,
            reject_msg: ''
        })
        toast(approved ? '审批通过，已转交采购' : '已驳回申请')
        fetchRequests()
    } catch (error) {
        toast('审批操作失败', 'error')
    } finally {
        leaderApprovingId.value = null
    }
}

onMounted(() => { fetchRequests() })

const viewProgress = (req: any) => {
    selectedRequest.value = req
    isProgressOpen.value = true
}

defineExpose({ fetchRequests })

const getStatusColor = (status: string) => {
    const map: Record<string, string> = {
        '待审批': 'bg-orange-100 text-orange-800',
        '待采购': 'bg-blue-100 text-blue-800',
        '已接单': 'bg-green-100 text-green-800',
        '已驳回': 'bg-red-100 text-red-800',
    }
    return map[status] || 'bg-gray-100 text-gray-800'
}
</script>

<template>
  <Card>
    <div class="p-6 space-y-4">
      <!-- Toolbar -->
      <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
          <div class="relative w-72">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
              <Input v-model="searchQuery" class="pl-9" placeholder="搜索单号、试剂名称、CAS号..." />
          </div>
          <div class="flex gap-1 rounded-lg bg-gray-100 p-1">
              <button
                v-for="s in statusOptions"
                :key="s"
                @click="statusFilter = s"
                :class="[
                  'px-3 py-1.5 text-xs font-medium rounded-md transition-all',
                  statusFilter === s
                    ? 'bg-white text-blue-700 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700'
                ]"
              >
                {{ s }}
              </button>
          </div>
      </div>

      <div v-if="isLoading" class="flex justify-center p-8">
        <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
      </div>
      <div v-else-if="filteredRequests.length === 0" class="text-center text-gray-500 py-8">
        暂无匹配的申购记录。
      </div>
      <div v-else class="overflow-x-auto rounded-lg border">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" class="px-6 py-3">单号</th>
              <th scope="col" class="px-6 py-3">申购人</th>
              <th scope="col" class="px-6 py-3">试剂名称</th>
              <th scope="col" class="px-6 py-3">数量</th>
              <th scope="col" class="px-6 py-3">状态</th>
              <th scope="col" class="px-6 py-3">申请日期</th>
              <th scope="col" class="px-6 py-3">操作</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="req in filteredRequests" :key="req.id">
              <!-- 主行 -->
              <tr class="bg-white hover:bg-gray-50"
                  :class="[
                    (req.status === '待采购' && role === 'procurement' && stockMap[req.reagent_catalog_id]) ? 'border-b-0' : 'border-b'
                  ]">
                <td class="px-6 py-4 font-mono text-gray-500">#{{ req.id }}</td>
                <td class="px-6 py-4 text-gray-700">
                    <span v-if="req.requestor?.real_name">{{ req.requestor.real_name }}</span>
                    <span v-else class="text-gray-400">System</span>
                </td>
                <td class="px-6 py-4 font-medium text-gray-900">
                    {{ req.reagent_catalog?.name || '未知' }}
                    <span class="block text-xs text-gray-500 font-normal">{{ req.reagent_catalog?.cas_number }}</span>
                    <!-- 采购角色库存状态微标签 (仅在非待采购状态显示，避免与下方的 AI 建议面板重复) -->
                    <div v-if="role === 'procurement' && stockMap[req.reagent_catalog_id] && req.status !== '待采购'" class="flex items-center gap-2 mt-1">
                      <span class="inline-flex items-center gap-0.5 text-[10px] px-1.5 py-0.5 rounded-full font-medium"
                            :class="stockMap[req.reagent_catalog_id].in_stock === 0 ? 'bg-red-100 text-red-700' : 'bg-emerald-100 text-emerald-700'">
                        <Package class="w-2.5 h-2.5" />
                        在库 {{ stockMap[req.reagent_catalog_id].in_stock }}
                      </span>
                      <span v-if="stockMap[req.reagent_catalog_id].pending_arrival > 0"
                            class="text-[10px] px-1.5 py-0.5 rounded-full bg-blue-100 text-blue-700 font-medium">
                        待到 {{ stockMap[req.reagent_catalog_id].pending_arrival }}
                      </span>
                    </div>
                </td>
                <td class="px-6 py-4">{{ req.quantity }} 瓶</td>
                <td class="px-6 py-4">
                    <span :class="['px-2 py-1 rounded-full text-xs font-medium', getStatusColor(req.status)]">
                        {{ req.status }}
                    </span>
                </td>
                <td class="px-6 py-4 text-gray-500">
                    {{ new Date(req.created_at).toLocaleDateString('zh-CN') }}
                </td>
                <td class="px-6 py-4">
                    <div class="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          @click="viewProgress(req)"
                          class="text-blue-600 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 border-blue-200"
                        >
                            查看进度
                        </Button>

                        <!-- Leader 审批 -->
                        <template v-if="req.status === '待审批' && role === 'leader'">
                            <Button size="sm" class="bg-indigo-600 hover:bg-indigo-700 text-white" 
                                :disabled="leaderApprovingId === req.id"
                                @click="leaderApprove(req.id, true)">
                                同意采购
                            </Button>
                            <Button size="sm" variant="outline" class="text-red-600 border-red-200 hover:bg-red-50"
                                :disabled="leaderApprovingId === req.id"
                                @click="leaderApprove(req.id, false)">
                                驳回
                            </Button>
                        </template>

                        <!-- 采购审批决策辅助区块 -->
                        <Button
                          v-else-if="req.status === '待采购' && role === 'procurement'"
                          size="sm"
                          class="bg-blue-600 hover:bg-blue-700 text-white shadow-sm"
                          @click="openOrderDialog(req)"
                        >
                            标记为已接单
                        </Button>

                        <span v-else-if="req.status === '待审批'" class="text-xs text-orange-500">
                            等待团队长审批
                        </span>
                        <span v-else-if="req.status === '待采购'" class="text-xs text-blue-500">
                            等待统一下单
                        </span>
                        <span v-else-if="req.status === '已接单'" class="text-xs text-green-600">
                            BPM-A 已闭环
                        </span>
                        <span v-else-if="req.status === '已驳回'" class="text-xs text-red-600">
                            已驳回
                        </span>
                    </div>
                </td>
              </tr>

              <!-- 展开行：AI 审批决策面板（采购员 + 待采购 + 有建议数据时展示） -->
              <tr v-if="req.status === '待采购' && role === 'procurement' && stockMap[req.reagent_catalog_id]" class="border-b-2 border-indigo-100 bg-indigo-50/40">
                <td colspan="7" class="px-6 py-2">
                    <div class="flex items-center gap-4 whitespace-nowrap overflow-hidden">
                        <div class="flex items-center gap-1.5 shrink-0 text-indigo-600 font-semibold text-xs text-nowrap">
                            <span>🤖 AI 建议:</span>
                        </div>

                        <div class="flex items-center gap-2 shrink-0">
                            <!-- 库存信息 Chips (已提至左侧) -->
                            <span v-if="stockMap[req.reagent_catalog_id].in_stock !== undefined"
                                  :class="['inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full font-medium border', stockMap[req.reagent_catalog_id].in_stock === 0 ? 'bg-red-100 text-red-700 border-red-200' : 'bg-emerald-100 text-emerald-700 border-emerald-200']">
                                在库 {{ stockMap[req.reagent_catalog_id].in_stock }}
                            </span>
                            <span v-if="stockMap[req.reagent_catalog_id].pending_arrival > 0"
                                  class="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full font-medium border bg-blue-100 text-blue-700 border-blue-200">
                                在途 {{ stockMap[req.reagent_catalog_id].pending_arrival }}
                            </span>
                            <span v-if="stockMap[req.reagent_catalog_id].last_consumed_at"
                                  class="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full font-medium border bg-gray-100 text-gray-600 border-gray-200">
                                消耗: {{ stockMap[req.reagent_catalog_id].last_consumed_at }}
                            </span>
                        </div>
                        
                        <!-- AI 建议文本 (随后的建议) -->
                        <div v-if="stockMap[req.reagent_catalog_id].advice"
                             class="flex-1 min-w-0 text-[11px] text-indigo-900 bg-indigo-100/60 border border-indigo-200 rounded px-2 py-0.5 truncate"
                             :title="stockMap[req.reagent_catalog_id].advice">
                            {{ stockMap[req.reagent_catalog_id].advice }}
                        </div>
                    </div>
                </td>
              </tr>

            </template>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Progress Timeline Dialog -->
    <RequestProgressDialog
        :open="isProgressOpen"
        :request="selectedRequest"
        @close="isProgressOpen = false"
        @refresh="fetchRequests"
    />

    <!-- BPM-A: 审批下单确认弹窗 -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="orderDialogOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
        <div class="bg-white rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden">
          <!-- Header -->
          <div class="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-4">
            <h3 class="text-white font-semibold text-base">标记已下单采购</h3>
            <p class="text-blue-100 text-xs mt-0.5">{{ orderDialogReqName }}</p>
          </div>
          <!-- Body -->
          <div class="px-6 py-5 space-y-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1.5">外部平台订单号 <span class="text-gray-400">(选填)</span></label>
              <input
                v-model="orderReference"
                type="text"
                placeholder="如：易派客订单号 EP20260201..."
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition"
              />
              <p class="text-[10px] text-gray-400 mt-1">可在外部平台下单后回填，也可以稍后补录</p>
            </div>
            <div class="bg-amber-50 border border-amber-200 rounded-lg px-3 py-2">
              <p class="text-xs text-amber-800">⚡ 确认后，此申购单状态将变更为「已接单」，BPM-A 需求闭环。</p>
            </div>
          </div>
          <!-- Footer -->
          <div class="px-6 py-3 bg-gray-50 flex justify-end gap-2 border-t">
            <Button size="sm" class="bg-gray-200 hover:bg-gray-300 text-gray-700" @click="orderDialogOpen = false">
              取消
            </Button>
            <Button
              size="sm"
              class="bg-blue-600 hover:bg-blue-700 text-white"
              :disabled="orderSubmitting"
              @click="confirmApproveAndOrder"
            >
              <Loader2 v-if="orderSubmitting" class="w-3.5 h-3.5 animate-spin mr-1" />
              确认标记已接单
            </Button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Toast -->
    <Transition
      enter-active-class="transition ease-out duration-300"
      enter-from-class="translate-y-4 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-4 opacity-0"
    >
      <div v-if="showToast" class="fixed bottom-6 right-6 z-50 max-w-sm">
        <div :class="[
          'px-4 py-3 rounded-lg shadow-lg border text-sm font-medium flex items-center gap-2',
          toastType === 'success' ? 'bg-green-50 text-green-800 border-green-200' : 'bg-red-50 text-red-800 border-red-200'
        ]">
          <span>{{ toastMessage }}</span>
        </div>
      </div>
    </Transition>
  </Card>
</template>
