<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { Loader2, Send, XCircle } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import TableSection from '@/components/ui/TableSection.vue'
import Input from '@/components/ui/Input.vue'
import LedgerTable from './LedgerTable.vue'
import FlowStatusPill from '@/components/workflow/FlowStatusPill.vue'
import FlowDetailDialog from '@/components/workflow/FlowDetailDialog.vue'
import axios from 'axios'
import { formatAmount } from '@/lib/quantity'

const props = defineProps({
  role: { type: String, default: 'researcher' },
  userId: { type: Number, default: 0 }
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
type ActionVariant = 'primary' | 'secondary' | 'outline' | 'destructive'
type FlowStepState = 'completed' | 'current' | 'pending' | 'rejected' | 'hidden'

interface FlowActionItem {
  key: string
  label: string
  variant?: ActionVariant
  disabled?: boolean
  loading?: boolean
}

interface FlowStepItem {
  key: string
  label: string
  state: FlowStepState
  description?: string
  operator?: string
  time?: string
}

const activeResearchTab = ref<ResearchTab>('catalog')
const requests = ref<any[]>([])
const isLoading = ref(true)
const controlledItems = ref<ControlledItem[]>([])
const loadingControlledItems = ref(false)
const searchKeyword = ref('')
const latestRequestId = ref<number | null>(null)
const flowDetailOpen = ref(false)
const flowDetailRequestId = ref<number | null>(null)

const controlledColumns = [
  { key: 'reagent', label: '试剂' },
  { key: 'barcode', label: '条码' },
  { key: 'location', label: '当前位置' },
  { key: 'volume', label: '余量' },
  { key: 'actions', label: '操作', align: 'right' as const },
]

const requestColumns = [
  { key: 'reagent', label: '试剂信息' },
  { key: 'requester', label: '申请人' },
  { key: 'amount', label: '申请量' },
  { key: 'status', label: '状态' },
  { key: 'meta', label: '申请信息' },
  { key: 'actions', label: '操作', align: 'right' as const },
]

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
    const res = await axios.get('/api/reagents/dispense-requests', { params: { role: props.role, user_id: props.userId } })
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
    })
    toast(approved ? '已批准领用申请' : '已驳回领用申请')
    approveDialogOpen.value = false
    await fetchRequests()
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
    await fetchRequests()
  } catch {
    toast('操作失败', 'error')
  } finally {
    keyConfirming.value = null
  }
}

const canCurrentUserKeyConfirm = (req: any) => {
  const uid = Number(props.userId || 0)
  if (!uid || req.status !== '待双签') return false
  return req.key_holder_a_id === uid || req.key_holder_b_id === uid
}

const formatTime = (t: string) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

const getBpmSteps = (req: any): FlowStepItem[] => {
  const status = req.status
  const controlled = !!req.reagent_item?.reagent_catalog?.is_controlled
  const latestDualTime = req.key_holder_b_confirmed_at || req.key_holder_a_confirmed_at || ''
  const leaderOperator = req.leader?.real_name || (req.leader_id ? `用户#${req.leader_id}` : '')
  const dualOperator = [
    req.key_holder_a_confirmed_at ? (req.key_holder_a?.real_name || '钥匙A') : '',
    req.key_holder_b_confirmed_at ? (req.key_holder_b?.real_name || '钥匙B') : '',
  ].filter(Boolean).join(' / ')

  const stepState = (key: 'submit' | 'leader' | 'dual' | 'done'): FlowStepState => {
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
    {
      key: 'submit',
      label: '提交申请',
      state: stepState('submit'),
      description: '研发提交试剂领用申请',
      operator: req.requester?.real_name || '-',
      time: formatTime(req.created_at),
    },
    {
      key: 'leader',
      label: '团队长审批',
      state: stepState('leader'),
      description: '团队长审批领用必要性与合规性',
      operator: leaderOperator || undefined,
      time: req.leader_approved_at ? formatTime(req.leader_approved_at) : undefined,
    },
    ...(controlled
      ? [{
          key: 'dual',
          label: '双签确认',
          state: stepState('dual') as FlowStepState,
          description: 'A/B 钥匙持有人分别确认',
          operator: dualOperator || undefined,
          time: latestDualTime ? formatTime(latestDualTime) : undefined,
        }]
      : []),
    {
      key: 'done',
      label: '完成领用',
      state: stepState('done'),
      description: '流程结束并记录领用流水',
      operator: stepState('done') === 'completed' ? '系统' : undefined,
      time: stepState('done') === 'completed' ? formatTime(req.updated_at) : undefined,
    },
  ]
}

const getFlowActions = (req: any): FlowActionItem[] => {
  const actions: FlowActionItem[] = []
  const isProcessing = keyConfirming.value === req.id

  if (props.role === 'leader' && req.status === '待审批') {
    actions.push({ key: 'leader-approve', label: '审批', variant: 'primary' })
  }
  if (canCurrentUserKeyConfirm(req)) {
    actions.push({ key: 'key-confirm', label: '确认', variant: 'primary', disabled: isProcessing, loading: isProcessing })
    actions.push({ key: 'key-reject', label: '驳回', variant: 'destructive', disabled: isProcessing })
  }

  return actions
}

const getLedgerActions = (req: any): FlowActionItem[] => {
  const actions: FlowActionItem[] = []
  const isProcessing = keyConfirming.value === req.id
  if (props.role === 'leader' && req.status === '待审批') {
    actions.push({ key: 'leader-approve', label: '审批', variant: 'primary' })
  }
  if (canCurrentUserKeyConfirm(req)) {
    actions.push({
      key: 'key-confirm',
      label: '确认',
      variant: 'primary',
      disabled: isProcessing,
      loading: isProcessing,
    })
  }
  return actions
}

const hasLedgerActions = (req: any) => getLedgerActions(req).length > 0
const currentFlowDetailRequest = computed(() => requests.value.find((req) => req.id === flowDetailRequestId.value) || null)

const openFlowDetail = (req: any) => {
  flowDetailRequestId.value = req.id
  flowDetailOpen.value = true
}

const getFlowMeta = (req: any) => {
  const controlled = !!req.reagent_item?.reagent_catalog?.is_controlled
  const meta = [
    { label: '申请单号', value: `#${req.id}` },
    { label: '申请人', value: req.requester?.real_name || '-' },
    { label: '申请量', value: String(req.amount ?? '-') },
    { label: '申请时间', value: formatTime(req.created_at) },
    { label: '用途', value: req.purpose || '-' },
  ]
  if (req.leader?.real_name || req.leader_id) {
    meta.push({ label: '审批人', value: req.leader?.real_name || `用户#${req.leader_id}` })
  }
  if (controlled) {
    meta.push({ label: '钥匙A', value: req.key_holder_a?.real_name || (req.key_holder_a_id ? `用户#${req.key_holder_a_id}` : '-') })
    meta.push({ label: '钥匙B', value: req.key_holder_b?.real_name || (req.key_holder_b_id ? `用户#${req.key_holder_b_id}` : '-') })
  }
  return meta
}

const getFlowNotes = (req: any) => {
  const notes: Array<{ type?: 'info' | 'warning' | 'error' | 'success'; text: string }> = []
  if (req.status === '待审批') {
    notes.push({ type: 'info', text: '当前待团队长审批。' })
  }
  if (req.status === '待双签') {
    notes.push({ type: 'info', text: `钥匙A（${req.key_holder_a?.real_name || '-'}）：${req.key_holder_a_confirmed_at ? '已确认' : '等待确认'}` })
    notes.push({ type: 'info', text: `钥匙B（${req.key_holder_b?.real_name || '-'}）：${req.key_holder_b_confirmed_at ? '已确认' : '等待确认'}` })
    if (req.expires_at) notes.push({ type: 'warning', text: `双签截止时间：${formatTime(req.expires_at)}` })
  }
  if (req.status === '已完成') {
    notes.push({ type: 'success', text: '流程已完成，领用流水已归档。' })
  }
  if (req.status === '已驳回' && (req.leader_reject_msg || req.key_holder_reject_msg)) {
    notes.push({ type: 'error', text: `驳回原因：${req.leader_reject_msg || req.key_holder_reject_msg}` })
  }
  return notes
}

const handleFlowAction = (req: any, actionKey: string) => {
  if (actionKey === 'leader-approve') {
    flowDetailOpen.value = false
    openApproveDialog(req)
    return
  }
  if (actionKey === 'key-confirm') {
    submitKeyHolderConfirm(req.id, true)
    return
  }
  if (actionKey === 'key-reject') {
    submitKeyHolderConfirm(req.id, false, '钥匙持有人驳回')
  }
}

const handleFlowDetailAction = (actionKey: string) => {
  if (!currentFlowDetailRequest.value) return
  handleFlowAction(currentFlowDetailRequest.value, actionKey)
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="role === 'researcher'" class="grid grid-cols-1 xl:grid-cols-12 gap-4 items-start">
      <TableSection class="xl:col-span-8" title="管控领用台账" description="申请台账负责选品，流程台账负责跟踪 BPM 状态">
        <template #actions>
          <Button variant="outline" size="sm" @click="() => { loadControlledItems(); fetchRequests() }">刷新列表</Button>
        </template>

        <template #toolbar>
          <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
            <div v-if="activeResearchTab === 'catalog'" class="relative w-full sm:w-72">
              <Input v-model="searchKeyword" placeholder="搜索名称/条码/库位" />
            </div>
            <div v-else class="text-xs text-gray-500">查看领用申请流转、审批和双签状态</div>
            <div class="apple-segmented w-fit sm:ml-auto">
              <button @click="activeResearchTab='catalog'" :class="['apple-segmented-btn', activeResearchTab==='catalog' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">申请台账</button>
              <button @click="activeResearchTab='bpm'" :class="['apple-segmented-btn', activeResearchTab==='bpm' ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']">流程台账</button>
            </div>
          </div>
        </template>

        <template v-if="activeResearchTab === 'catalog'">
          <div v-if="loadingControlledItems" class="flex justify-center py-10"><Loader2 class="w-6 h-6 text-gray-400 animate-spin" /></div>
          <div v-else-if="filteredControlledItems.length === 0" class="apple-table-empty">暂无可申请领用的在库管控试剂。</div>
          <LedgerTable v-else :columns="controlledColumns">
                <tr
                  v-for="item in filteredControlledItems"
                  :key="item.uuid"
                  class="border-b border-gray-100 hover:bg-gray-50 cursor-pointer"
                  :class="quickApply.reagentItemId === item.uuid ? 'bg-blue-50/60' : ''"
                  @click="selectForQuickApply(item)"
                >
                  <td class="px-6 py-4 font-medium text-gray-900">{{ item.reagent_catalog?.name || '未知试剂' }}</td>
                  <td class="px-6 py-4 font-mono text-xs text-blue-600">#{{ item.uuid.substring(0, 8).toUpperCase() }}</td>
                  <td class="px-6 py-4 text-xs text-gray-600">{{ item.location || '未分配' }}</td>
                  <td class="px-6 py-4 text-xs text-gray-700">{{ formatAmount(item.remaining_volume, item.reagent_catalog?.unit, 'ml') }}</td>
                  <td class="px-6 py-4 text-right">
                    <Button size="sm" variant="outline" @click.stop="selectForQuickApply(item)">快捷申请</Button>
                  </td>
                </tr>
          </LedgerTable>
        </template>

        <template v-else>
          <div v-if="isLoading" class="flex justify-center py-10"><Loader2 class="w-6 h-6 text-gray-400 animate-spin" /></div>
          <div v-else-if="requests.length === 0" class="apple-table-empty">暂无领用申请记录。</div>
          <LedgerTable v-else :columns="requestColumns">
            <tr
              v-for="req in requests"
              :key="req.id"
              class="border-b border-gray-100 hover:bg-gray-50"
              :class="latestRequestId === req.id ? 'bg-blue-50/60' : ''"
            >
              <td class="px-6 py-4">
                <div class="font-medium text-sm text-gray-900">{{ req.reagent_item?.reagent_catalog?.name || '未知试剂' }}</div>
                <div class="text-xs text-gray-500 mt-1">#{{ req.id }} · {{ formatTime(req.created_at) }}</div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-700">{{ req.requester?.real_name || '-' }}</td>
              <td class="px-6 py-4 text-sm text-gray-700">{{ req.amount }}</td>
              <td class="px-6 py-4"><FlowStatusPill :status="req.status" /></td>
              <td class="px-6 py-4 text-xs text-gray-600">
                <div>时间：{{ formatTime(req.created_at) }}</div>
                <div class="mt-1 truncate max-w-[220px]">用途：{{ req.purpose || '-' }}</div>
                <div v-if="req.status === '已驳回' && (req.leader_reject_msg || req.key_holder_reject_msg)" class="mt-1 text-red-600">
                  驳回：{{ req.leader_reject_msg || req.key_holder_reject_msg }}
                </div>
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <Button size="sm" variant="outline" class="h-7 px-3 text-[11px]" @click="openFlowDetail(req)">流转单</Button>
                  <Button
                    v-for="action in getLedgerActions(req)"
                    :key="action.key"
                    size="sm"
                    :variant="action.variant || 'outline'"
                    class="h-7 px-3 text-[11px]"
                    :disabled="action.disabled"
                    @click="handleFlowAction(req, action.key)"
                  >
                    <Loader2 v-if="action.loading" class="w-3 h-3 animate-spin mr-1" />
                    {{ action.label }}
                  </Button>
                  <span v-if="!hasLedgerActions(req)" class="text-xs text-slate-400">仅查看</span>
                </div>
              </td>
            </tr>
          </LedgerTable>
        </template>
      </TableSection>

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

    <TableSection v-else title="领用流程台账" description="显示每条领用申请的 BPM 流转状态">
      <template #actions>
        <Button variant="outline" size="sm" @click="fetchRequests">刷新列表</Button>
      </template>

      <div v-if="isLoading" class="flex justify-center py-10"><Loader2 class="w-6 h-6 text-gray-400 animate-spin" /></div>
      <div v-else-if="requests.length === 0" class="apple-table-empty">暂无领用申请记录。</div>
      <LedgerTable v-else :columns="requestColumns">
        <tr v-for="req in requests" :key="req.id" class="border-b border-gray-100 hover:bg-gray-50">
          <td class="px-6 py-4">
            <div class="font-medium text-sm text-gray-900">{{ req.reagent_item?.reagent_catalog?.name || '未知试剂' }}</div>
            <div class="text-xs text-gray-500 mt-1">#{{ req.id }} · {{ formatTime(req.created_at) }}</div>
          </td>
          <td class="px-6 py-4 text-sm text-gray-700">{{ req.requester?.real_name || '-' }}</td>
          <td class="px-6 py-4 text-sm text-gray-700">{{ req.amount }}</td>
          <td class="px-6 py-4"><FlowStatusPill :status="req.status" /></td>
          <td class="px-6 py-4 text-xs text-gray-600">
            <div>时间：{{ formatTime(req.created_at) }}</div>
            <div class="mt-1 truncate max-w-[220px]">用途：{{ req.purpose || '-' }}</div>
            <div v-if="req.status === '已驳回' && (req.leader_reject_msg || req.key_holder_reject_msg)" class="mt-1 text-red-600">
              驳回：{{ req.leader_reject_msg || req.key_holder_reject_msg }}
            </div>
          </td>
          <td class="px-6 py-4 text-right">
            <div class="flex items-center justify-end gap-2">
              <Button size="sm" variant="outline" class="h-7 px-3 text-[11px]" @click="openFlowDetail(req)">流转单</Button>
              <Button
                v-for="action in getLedgerActions(req)"
                :key="action.key"
                size="sm"
                :variant="action.variant || 'outline'"
                class="h-7 px-3 text-[11px]"
                :disabled="action.disabled"
                @click="handleFlowAction(req, action.key)"
              >
                <Loader2 v-if="action.loading" class="w-3 h-3 animate-spin mr-1" />
                {{ action.label }}
              </Button>
              <span v-if="!hasLedgerActions(req)" class="text-xs text-slate-400">仅查看</span>
            </div>
          </td>
        </tr>
      </LedgerTable>
    </TableSection>

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
            <div v-if="approveTarget?.reagent_item?.reagent_catalog?.is_controlled" class="bg-purple-50 border border-purple-200 rounded-lg px-4 py-3 space-y-2">
              <p class="text-xs font-semibold text-purple-800">管控品将自动分配固定双签持有人（由用户权限管理配置）</p>
              <p class="text-[11px] text-purple-700">如需调整持有人，请在「用户与组织 -> 权限策略」中修改。</p>
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

    <Transition enter-active-class="transition ease-out duration-300" enter-from-class="translate-y-4 opacity-0" enter-to-class="translate-y-0 opacity-100" leave-active-class="transition ease-in duration-200" leave-from-class="translate-y-0 opacity-100" leave-to-class="translate-y-4 opacity-0">
      <div v-if="showToast" class="apple-toast-wrap">
        <div :class="['apple-toast', toastType === 'success' ? 'apple-toast-success' : 'apple-toast-error']">
          <span>{{ toastMessage }}</span>
        </div>
      </div>
    </Transition>

    <FlowDetailDialog
      :open="flowDetailOpen"
      title="领用流程流转单"
      :subtitle="currentFlowDetailRequest ? (currentFlowDetailRequest.reagent_item?.reagent_catalog?.name || '未知试剂') : ''"
      :status="currentFlowDetailRequest?.status || '-'"
      :meta="currentFlowDetailRequest ? getFlowMeta(currentFlowDetailRequest) : []"
      :steps="currentFlowDetailRequest ? getBpmSteps(currentFlowDetailRequest) : []"
      :actions="currentFlowDetailRequest ? getFlowActions(currentFlowDetailRequest) : []"
      :notes="currentFlowDetailRequest ? getFlowNotes(currentFlowDetailRequest) : []"
      @close="flowDetailOpen = false"
      @action="handleFlowDetailAction"
    />
  </div>
</template>
