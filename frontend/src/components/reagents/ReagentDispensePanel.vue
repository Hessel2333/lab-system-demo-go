<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { CheckCircle2, XCircle, Clock, Lock, Loader2, ShieldCheck, Send } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import axios from 'axios'
import { formatAmount } from '@/lib/quantity'

const props = defineProps({
  role: { type: String, default: 'researcher' }
})

interface ControlledItem {
  uuid: string
  location: string
  remaining_volume: number
  reagent_catalog?: {
    name: string
    unit: string
    is_controlled: boolean
  }
}

type ResearchTab = 'catalog' | 'bpm'

const activeResearchTab = ref<ResearchTab>('catalog')
const requests = ref<any[]>([])
const isLoading = ref(true)
const controlledItems = ref<ControlledItem[]>([])
const loadingControlledItems = ref(false)
const searchKeyword = ref('')
const latestRequestId = ref<number | null>(null)

const quickApply = ref({
  reagentItemId: '',
  amount: 0,
  purpose: '',
})
const quickSubmitting = ref(false)

const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const toast = (msg: string, type: 'success' | 'error' = 'success') => {
  toastMessage.value = msg
  toastType.value = type
  showToast.value = true
  setTimeout(() => { showToast.value = false }, 2800)
}

const fetchRequests = async () => {
  isLoading.value = true
  try {
    const res = await axios.get('/api/reagents/dispense-requests', { params: { role: props.role } })
    requests.value = res.data || []
  } catch {
    requests.value = []
  } finally {
    isLoading.value = false
  }
}

const loadControlledItems = async () => {
  if (props.role !== 'researcher') return
  loadingControlledItems.value = true
  try {
    const res = await axios.get('/api/reagents/items', { params: { status: '在库' } })
    const items = (res.data || []) as ControlledItem[]
    controlledItems.value = items.filter(i => i.reagent_catalog?.is_controlled)
  } catch {
    controlledItems.value = []
  } finally {
    loadingControlledItems.value = false
  }
}

onMounted(() => {
  fetchRequests()
  loadControlledItems()
})

const filteredControlledItems = computed(() => {
  const q = searchKeyword.value.trim().toLowerCase()
  if (!q) return controlledItems.value
  return controlledItems.value.filter(item => {
    const name = item.reagent_catalog?.name?.toLowerCase() || ''
    const uuid = item.uuid.toLowerCase()
    const location = item.location?.toLowerCase() || ''
    return name.includes(q) || uuid.includes(q) || location.includes(q)
  })
})

const selectedItem = computed(() => controlledItems.value.find(i => i.uuid === quickApply.value.reagentItemId))

const selectForQuickApply = (item: ControlledItem) => {
  quickApply.value.reagentItemId = item.uuid
  quickApply.value.amount = item.remaining_volume > 0 ? Math.min(1, item.remaining_volume) : 0
  quickApply.value.purpose = ''
  toast(`已选择：${item.reagent_catalog?.name || item.uuid.substring(0, 8)}`)
}

const submitQuickApply = async () => {
  const item = selectedItem.value
  if (!item) {
    toast('请先从左侧台账选择一条管控试剂', 'error')
    return
  }
  if (quickApply.value.amount <= 0) {
    toast('领取量必须大于 0', 'error')
    return
  }
  if (item.remaining_volume > 0 && quickApply.value.amount > item.remaining_volume) {
    toast('领取量超过当前余量', 'error')
    return
  }

  quickSubmitting.value = true
  try {
    const res = await axios.post('/api/reagents/dispense-requests', {
      reagent_item_id: quickApply.value.reagentItemId,
      amount: quickApply.value.amount,
      purpose: quickApply.value.purpose,
    })
    latestRequestId.value = res.data?.id || null
    toast('领用申请已提交，已切换到流程台账')
    quickApply.value = { reagentItemId: '', amount: 0, purpose: '' }
    activeResearchTab.value = 'bpm'
    await Promise.all([fetchRequests(), loadControlledItems()])
  } catch (e: any) {
    toast('提交失败: ' + (e.response?.data?.error || '服务器错误'), 'error')
  } finally {
    quickSubmitting.value = false
  }
}

// --- 团队长审批 ---
const approveDialogOpen = ref(false)
const approveTarget = ref<any>(null)
const approveRejectMsg = ref('')
const approveKeyHolderAId = ref<number | undefined>()
const approveKeyHolderBId = ref<number | undefined>()
const approving = ref(false)

const openApproveDialog = (req: any) => {
  approveTarget.value = req
  approveRejectMsg.value = ''
  approveDialogOpen.value = true
}

const submitLeaderApproval = async (approved: boolean) => {
  if (!approveTarget.value) return
  approving.value = true
  try {
    await axios.post(`/api/reagents/dispense-requests/${approveTarget.value.id}/leader-approve`, {
      approved,
      reject_msg: approveRejectMsg.value,
      key_holder_a_id: approveKeyHolderAId.value,
      key_holder_b_id: approveKeyHolderBId.value,
    })
    toast(approved ? '已批准领用申请' : '已驳回领用申请')
    approveDialogOpen.value = false
    fetchRequests()
  } catch {
    toast('操作失败', 'error')
  } finally {
    approving.value = false
  }
}

// --- 钥匙持有人确认 ---
const keyConfirming = ref<number | null>(null)
const submitKeyHolderConfirm = async (reqId: number, confirmed: boolean, rejectMsg = '') => {
  keyConfirming.value = reqId
  try {
    await axios.post(`/api/reagents/dispense-requests/${reqId}/key-holder-confirm`, {
      confirmed,
      reject_msg: rejectMsg,
    })
    toast(confirmed ? '已确认开锁' : '已驳回取用')
    fetchRequests()
  } catch {
    toast('操作失败', 'error')
  } finally {
    keyConfirming.value = null
  }
}

const getStatusConfig = (status: string) => {
  const map: Record<string, { color: string, label: string, icon: any }> = {
    '待审批': { color: 'bg-orange-100 text-orange-800 border-orange-200', label: '待审批', icon: Clock },
    '已通过': { color: 'bg-green-100 text-green-800 border-green-200', label: '已通过', icon: CheckCircle2 },
    '已驳回': { color: 'bg-red-100 text-red-800 border-red-200', label: '已驳回', icon: XCircle },
    '待双签': { color: 'bg-purple-100 text-purple-800 border-purple-200', label: '待双签', icon: Lock },
    '已完成': { color: 'bg-emerald-100 text-emerald-800 border-emerald-200', label: '已完成', icon: ShieldCheck },
  }
  return map[status] || { color: 'bg-gray-100 text-gray-800 border-gray-200', label: status, icon: Clock }
}

const formatTime = (t: string) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const getBpmSteps = (req: any) => {
  const status = req.status
  const controlled = !!req.reagent_item?.reagent_catalog?.is_controlled

  const stepState = (key: 'submit' | 'leader' | 'dual' | 'done') => {
    if (key === 'submit') return 'completed'
    if (key === 'leader') {
      if (status === '待审批') return 'current'
      if (['待双签', '已完成', '已通过'].includes(status)) return 'completed'
      if (status === '已驳回') return 'rejected'
      return 'pending'
    }
    if (key === 'dual') {
      if (!controlled) return 'hidden'
      if (status === '待双签') return 'current'
      if (status === '已完成') return 'completed'
      if (status === '已驳回') return 'rejected'
      return 'pending'
    }
    if (key === 'done') {
      if (status === '已完成') return 'completed'
      if (status === '已驳回') return 'rejected'
      return 'pending'
    }
    return 'pending'
  }

  return [
    { key: 'submit', label: '提交申请', state: stepState('submit') },
    { key: 'leader', label: '团队长审批', state: stepState('leader') },
    ...(controlled ? [{ key: 'dual', label: '双签确认', state: stepState('dual') }] : []),
    { key: 'done', label: '完成领用', state: stepState('done') },
  ]
}

const stepClass = (state: string) => {
  if (state === 'completed') return 'bg-emerald-100 border-emerald-200 text-emerald-700'
  if (state === 'current') return 'bg-blue-100 border-blue-200 text-blue-700'
  if (state === 'rejected') return 'bg-red-100 border-red-200 text-red-700'
  return 'bg-gray-100 border-gray-200 text-gray-500'
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="role === 'researcher'" class="grid grid-cols-1 xl:grid-cols-12 gap-4 items-start">
      <Card class="xl:col-span-8">
        <div class="p-6 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900">管控领用台账</h2>
              <p class="text-xs text-gray-500 mt-0.5">申请台账负责选品，流程台账负责跟踪 BPM 状态</p>
            </div>
            <button @click="() => { loadControlledItems(); fetchRequests() }" class="text-xs text-blue-600 hover:underline">刷新</button>
          </div>

          <div class="apple-segmented w-fit">
            <button @click="activeResearchTab='catalog'" :class="['apple-segmented-btn', activeResearchTab==='catalog' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">申请台账</button>
            <button @click="activeResearchTab='bpm'" :class="['apple-segmented-btn', activeResearchTab==='bpm' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">流程台账</button>
          </div>

          <template v-if="activeResearchTab === 'catalog'">
            <div class="relative w-72">
              <input v-model="searchKeyword" class="w-full h-9 rounded-lg border border-gray-200 px-3 text-sm" placeholder="搜索名称/条码/库位" />
            </div>

            <div v-if="loadingControlledItems" class="text-sm text-gray-400 py-6">加载管控试剂台账中...</div>
            <div v-else-if="filteredControlledItems.length === 0" class="text-sm text-gray-400 py-6">暂无可申请领用的在库管控试剂。</div>
            <div v-else class="apple-table-wrap">
              <table class="w-full text-sm text-left">
                <thead>
                  <tr>
                    <th class="px-4 py-3">试剂</th>
                    <th class="px-4 py-3">条码</th>
                    <th class="px-4 py-3">当前位置</th>
                    <th class="px-4 py-3">余量</th>
                    <th class="px-4 py-3 text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in filteredControlledItems"
                    :key="item.uuid"
                    class="border-b border-gray-100 hover:bg-gray-50 cursor-pointer"
                    :class="quickApply.reagentItemId === item.uuid ? 'bg-blue-50/60' : ''"
                    @click="selectForQuickApply(item)"
                  >
                    <td class="px-4 py-3 font-medium text-gray-900">{{ item.reagent_catalog?.name || '未知试剂' }}</td>
                    <td class="px-4 py-3 font-mono text-xs text-blue-600">#{{ item.uuid.substring(0, 8).toUpperCase() }}</td>
                    <td class="px-4 py-3 text-xs text-gray-600">{{ item.location || '未分配' }}</td>
                    <td class="px-4 py-3 text-xs text-gray-700">{{ formatAmount(item.remaining_volume, item.reagent_catalog?.unit, 'ml') }}</td>
                    <td class="px-4 py-3 text-right">
                      <Button size="sm" variant="outline" @click.stop="selectForQuickApply(item)">快捷申请</Button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </template>

          <template v-else>
            <div v-if="isLoading" class="flex justify-center py-10"><Loader2 class="w-6 h-6 text-gray-400 animate-spin" /></div>
            <div v-else-if="requests.length === 0" class="text-center py-10 text-gray-400 text-sm">暂无领用申请记录</div>
            <div v-else class="space-y-3">
              <div
                v-for="req in requests"
                :key="req.id"
                class="border rounded-xl p-4 hover:shadow-sm transition-shadow"
                :class="latestRequestId === req.id ? 'ring-2 ring-blue-200 border-blue-300' : ''"
              >
                <div class="space-y-2">
                  <div class="flex items-center gap-2">
                    <span class="font-medium text-sm text-gray-900">{{ req.reagent_item?.reagent_catalog?.name || '未知试剂' }}</span>
                    <span :class="['inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full font-medium border', getStatusConfig(req.status).color]">
                      <component :is="getStatusConfig(req.status).icon" class="w-3 h-3" />
                      {{ getStatusConfig(req.status).label }}
                    </span>
                  </div>
                  <div class="flex items-center gap-3 text-xs text-gray-500">
                    <span>申请人: {{ req.requester?.real_name || '-' }}</span>
                    <span>用量: {{ req.amount }}</span>
                    <span v-if="req.purpose">用途: {{ req.purpose }}</span>
                    <span>{{ formatTime(req.created_at) }}</span>
                  </div>
                  <div class="flex flex-wrap gap-2 pt-1">
                    <span v-for="step in getBpmSteps(req)" :key="step.key" :class="['inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-medium', stepClass(step.state)]">{{ step.label }}</span>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </Card>

      <Card class="xl:col-span-4 xl:sticky xl:top-4">
        <div class="p-6 space-y-4">
          <div>
            <h3 class="text-base font-semibold text-gray-900">固定申请面板</h3>
            <p class="text-xs text-gray-500 mt-0.5">右侧始终可见，左侧选中条目后即可提交</p>
          </div>

          <div>
            <label class="block text-xs text-gray-500 mb-1">已选试剂</label>
            <div class="h-10 rounded-lg border border-gray-200 bg-white px-3 flex items-center text-sm text-gray-700">
              <span v-if="selectedItem">{{ selectedItem.reagent_catalog?.name }} (#{{ selectedItem.uuid.substring(0, 8).toUpperCase() }})</span>
              <span v-else class="text-gray-400">请先在左侧点击“快捷申请”</span>
            </div>
          </div>

          <div>
            <label class="block text-xs text-gray-500 mb-1">领取量</label>
            <input v-model.number="quickApply.amount" type="number" min="0" step="0.1" class="w-full h-10 rounded-lg border border-gray-200 px-3 text-sm bg-white" placeholder="输入数值" />
            <p class="text-[11px] text-gray-500 mt-1">当前余量：{{ formatAmount(selectedItem?.remaining_volume, selectedItem?.reagent_catalog?.unit, 'ml') }}</p>
          </div>

          <div>
            <label class="block text-xs text-gray-500 mb-1">用途（选填）</label>
            <textarea v-model="quickApply.purpose" rows="3" class="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm bg-white" placeholder="如：样品前处理"></textarea>
          </div>

          <Button class="w-full" size="sm" variant="primary" :disabled="quickSubmitting || !selectedItem" @click="submitQuickApply">
            <Send class="w-3.5 h-3.5 mr-1" /> 提交申请
          </Button>
        </div>
      </Card>
    </div>

    <Card v-else>
      <div class="p-6 space-y-4">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">领用流程台账</h2>
          <p class="text-xs text-gray-500 mt-0.5">显示每条领用申请的 BPM 流转状态</p>
        </div>

        <div v-if="isLoading" class="flex justify-center py-10"><Loader2 class="w-6 h-6 text-gray-400 animate-spin" /></div>
        <div v-else-if="requests.length === 0" class="text-center py-10 text-gray-400 text-sm">暂无领用申请记录</div>

        <div v-else class="space-y-3">
          <div v-for="req in requests" :key="req.id" class="border rounded-xl p-4 hover:shadow-sm transition-shadow">
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1 min-w-0 space-y-2">
                <div class="flex items-center gap-2">
                  <span class="font-medium text-sm text-gray-900">{{ req.reagent_item?.reagent_catalog?.name || '未知试剂' }}</span>
                  <span :class="['inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full font-medium border', getStatusConfig(req.status).color]">
                    <component :is="getStatusConfig(req.status).icon" class="w-3 h-3" />
                    {{ getStatusConfig(req.status).label }}
                  </span>
                </div>
                <div class="flex items-center gap-3 text-xs text-gray-500">
                  <span>申请人: {{ req.requester?.real_name || '-' }}</span>
                  <span>用量: {{ req.amount }}</span>
                  <span v-if="req.purpose">用途: {{ req.purpose }}</span>
                  <span>{{ formatTime(req.created_at) }}</span>
                </div>
                <div class="flex flex-wrap gap-2 pt-1">
                  <span v-for="step in getBpmSteps(req)" :key="step.key" :class="['inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-medium', stepClass(step.state)]">{{ step.label }}</span>
                </div>

                <div v-if="req.status === '待双签'" class="flex items-center gap-4 text-xs bg-purple-50 rounded-lg px-3 py-2 border border-purple-100">
                  <span :class="req.key_holder_a_confirmed_at ? 'text-green-600' : 'text-gray-500'">钥匙A {{ req.key_holder_a?.real_name || '-' }}: {{ req.key_holder_a_confirmed_at ? '已确认' : '等待中' }}</span>
                  <span :class="req.key_holder_b_confirmed_at ? 'text-green-600' : 'text-gray-500'">钥匙B {{ req.key_holder_b?.real_name || '-' }}: {{ req.key_holder_b_confirmed_at ? '已确认' : '等待中' }}</span>
                  <span v-if="req.expires_at" class="text-orange-500">截止 {{ formatTime(req.expires_at) }}</span>
                </div>

                <div v-if="req.status === '已驳回' && (req.leader_reject_msg || req.key_holder_reject_msg)" class="text-xs text-red-600 bg-red-50 rounded px-2 py-1">驳回原因: {{ req.leader_reject_msg || req.key_holder_reject_msg }}</div>
              </div>

              <div class="flex items-center gap-2 shrink-0">
                <template v-if="role === 'leader' && req.status === '待审批'">
                  <Button size="sm" variant="primary" class="text-xs" @click="openApproveDialog(req)">审批</Button>
                </template>
                <template v-if="role === 'key_holder' && req.status === '待双签'">
                  <Button size="sm" class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs" :disabled="keyConfirming === req.id" @click="submitKeyHolderConfirm(req.id, true)">
                    <Loader2 v-if="keyConfirming === req.id" class="w-3 h-3 animate-spin mr-1" />确认
                  </Button>
                  <Button size="sm" class="bg-red-100 hover:bg-red-200 text-red-700 text-xs" @click="submitKeyHolderConfirm(req.id, false, '钥匙持有人驳回')">驳回</Button>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Transition enter-active-class="transition ease-out duration-200" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition ease-in duration-150" leave-from-class="opacity-100" leave-to-class="opacity-0">
        <div v-if="approveDialogOpen" class="apple-modal-backdrop">
          <div class="apple-modal-panel max-w-lg overflow-hidden">
            <div class="flex items-center justify-between px-6 py-4 border-b">
              <div>
                <h3 class="text-lg font-semibold text-gray-900 leading-tight">审批领用申请</h3>
                <p class="text-gray-500 text-xs mt-1">{{ approveTarget?.reagent_item?.reagent_catalog?.name }} · 申请人: {{ approveTarget?.requester?.real_name }}</p>
              </div>
              <button @click="approveDialogOpen = false" class="p-1.5 rounded-md hover:bg-gray-100 transition-colors"><XCircle class="w-5 h-5 text-gray-400" /></button>
            </div>
            <div class="px-6 py-5 space-y-4">
              <div class="grid grid-cols-2 gap-3 text-sm">
                <div><span class="text-gray-500">用量:</span> {{ approveTarget?.amount }}</div>
                <div><span class="text-gray-500">用途:</span> {{ approveTarget?.purpose || '-' }}</div>
              </div>
              <div v-if="approveTarget?.reagent_item?.reagent_catalog?.is_controlled" class="bg-purple-50 border border-purple-200 rounded-lg px-4 py-3 space-y-3">
                <p class="text-xs font-semibold text-purple-800">管控品需指定双人双锁持有人</p>
                <div class="grid grid-cols-2 gap-3">
                  <div>
                    <label class="block text-[10px] text-gray-500 mb-0.5">钥匙 A 持有人 ID</label>
                    <input v-model.number="approveKeyHolderAId" type="number" placeholder="用户 ID" class="w-full px-2 py-1 text-xs border rounded" />
                  </div>
                  <div>
                    <label class="block text-[10px] text-gray-500 mb-0.5">钥匙 B 持有人 ID</label>
                    <input v-model.number="approveKeyHolderBId" type="number" placeholder="用户 ID" class="w-full px-2 py-1 text-xs border rounded" />
                  </div>
                </div>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1">驳回原因 <span class="text-gray-400">(驳回时必填)</span></label>
                <textarea v-model="approveRejectMsg" rows="2" placeholder="如需驳回，请填写原因..." class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 outline-none"></textarea>
              </div>
            </div>
            <div class="px-6 py-3 bg-gray-50 flex justify-end gap-2 border-t">
              <Button size="sm" variant="secondary" @click="approveDialogOpen = false">取消</Button>
              <Button size="sm" variant="destructive" :disabled="approving || !approveRejectMsg" @click="submitLeaderApproval(false)">驳回</Button>
              <Button size="sm" variant="primary" :disabled="approving" @click="submitLeaderApproval(true)">
                <Loader2 v-if="approving" class="w-3.5 h-3.5 animate-spin mr-1" />批准
              </Button>
            </div>
          </div>
        </div>
      </Transition>
    </Card>

    <Transition enter-active-class="transition ease-out duration-300" enter-from-class="translate-y-4 opacity-0" enter-to-class="translate-y-0 opacity-100" leave-active-class="transition ease-in duration-200" leave-from-class="translate-y-0 opacity-100" leave-to-class="translate-y-4 opacity-0">
      <div v-if="showToast" class="apple-toast-wrap">
        <div :class="['apple-toast', toastType === 'success' ? 'apple-toast-success' : 'apple-toast-error']">
          <span>{{ toastMessage }}</span>
        </div>
      </div>
    </Transition>
  </div>
</template>
