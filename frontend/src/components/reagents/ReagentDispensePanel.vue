<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { CheckCircle2, XCircle, Clock, Lock, Loader2, ShieldCheck } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import axios from 'axios'

const props = defineProps({
    role: { type: String, default: 'researcher' }
})

const requests = ref<any[]>([])
const isLoading = ref(true)

const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const toast = (msg: string, type: 'success' | 'error' = 'success') => {
    toastMessage.value = msg
    toastType.value = type
    showToast.value = true
    setTimeout(() => { showToast.value = false }, 3000)
}

const fetchRequests = async () => {
    isLoading.value = true
    try {
        const res = await axios.get('/api/reagents/dispense-requests', {
            params: { role: props.role }
        })
        requests.value = res.data || []
    } catch {
        requests.value = []
    } finally {
        isLoading.value = false
    }
}

onMounted(fetchRequests)

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
        toast(confirmed ? '✅ 已确认开锁' : '已驳回取用')
        fetchRequests()
    } catch {
        toast('操作失败', 'error')
    } finally {
        keyConfirming.value = null
    }
}

// --- 研发申请领用 (Inline Dialog) ---
const newDialogOpen = ref(false)
const newReagentItemId = ref('')
const newAmount = ref(0)
const newPurpose = ref('')
const newSubmitting = ref(false)

const submitNewDispense = async () => {
    if (!newReagentItemId.value || newAmount.value <= 0) {
        toast('请填写完整信息', 'error')
        return
    }
    newSubmitting.value = true
    try {
        await axios.post('/api/reagents/dispense-requests', {
            reagent_item_id: newReagentItemId.value,
            amount: newAmount.value,
            purpose: newPurpose.value,
        })
        toast('领用申请已提交，等待团队长审批')
        newDialogOpen.value = false
        newReagentItemId.value = ''
        newAmount.value = 0
        newPurpose.value = ''
        fetchRequests()
    } catch {
        toast('提交失败', 'error')
    } finally {
        newSubmitting.value = false
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
</script>

<template>
  <Card>
    <div class="p-6 space-y-4">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">
            {{ role === 'leader' ? '领用审批' : role === 'key_holder' ? '双签确认' : '我的领用' }}
          </h2>
          <p class="text-xs text-gray-500 mt-0.5">
            {{ role === 'leader' ? '审批研发人员的试剂领用申请，管控品需指定钥匙持有人' :
               role === 'key_holder' ? '确认或驳回管控品双人双锁取用请求' :
               '查看我提交的领用申请和审批进度' }}
          </p>
        </div>
        <Button
          v-if="role === 'researcher'"
          size="sm"
          class="bg-blue-600 hover:bg-blue-700 text-white"
          @click="newDialogOpen = true"
        >
          + 新建领用申请
        </Button>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="flex justify-center py-10">
        <Loader2 class="w-6 h-6 text-gray-400 animate-spin" />
      </div>

      <!-- Empty -->
      <div v-else-if="requests.length === 0" class="text-center py-10 text-gray-400 text-sm">
        暂无领用申请记录
      </div>

      <!-- List -->
      <div v-else class="space-y-3">
        <div
          v-for="req in requests" :key="req.id"
          class="border rounded-xl p-4 hover:shadow-sm transition-shadow"
        >
          <div class="flex items-start justify-between gap-4">
            <!-- Left: Info -->
            <div class="flex-1 min-w-0 space-y-1.5">
              <div class="flex items-center gap-2">
                <span class="font-medium text-sm text-gray-900">
                  {{ req.reagent_item?.reagent_catalog?.name || '未知试剂' }}
                </span>
                <span :class="['inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full font-medium border', getStatusConfig(req.status).color]">
                  <component :is="getStatusConfig(req.status).icon" class="w-3 h-3" />
                  {{ getStatusConfig(req.status).label }}
                </span>
                <span v-if="req.reagent_item?.reagent_catalog?.is_controlled" class="text-[10px] px-1.5 py-0.5 rounded bg-red-50 text-red-600 border border-red-100 font-medium">
                  🔒 管控品
                </span>
              </div>
              <div class="flex items-center gap-3 text-xs text-gray-500">
                <span>申请人: {{ req.requester?.real_name || '-' }}</span>
                <span>用量: {{ req.amount }}</span>
                <span v-if="req.purpose">用途: {{ req.purpose }}</span>
                <span>{{ formatTime(req.created_at) }}</span>
              </div>

              <!-- 双签进度（管控品） -->
              <div v-if="req.status === '待双签'" class="flex items-center gap-3 mt-2 bg-purple-50 rounded-lg px-3 py-2 border border-purple-100">
                <Lock class="w-4 h-4 text-purple-500 shrink-0" />
                <div class="flex items-center gap-4 text-xs">
                  <span :class="req.key_holder_a_confirmed_at ? 'text-green-600' : 'text-gray-500'">
                    钥匙A {{ req.key_holder_a?.real_name || '-' }}:
                    {{ req.key_holder_a_confirmed_at ? '✅ 已确认' : '⏳ 等待中' }}
                  </span>
                  <span :class="req.key_holder_b_confirmed_at ? 'text-green-600' : 'text-gray-500'">
                    钥匙B {{ req.key_holder_b?.real_name || '-' }}:
                    {{ req.key_holder_b_confirmed_at ? '✅ 已确认' : '⏳ 等待中' }}
                  </span>
                  <span v-if="req.expires_at" class="text-orange-500">
                    ⏰ 截止 {{ formatTime(req.expires_at) }}
                  </span>
                </div>
              </div>

              <!-- 驳回原因 -->
              <div v-if="req.status === '已驳回' && (req.leader_reject_msg || req.key_holder_reject_msg)" class="text-xs text-red-600 bg-red-50 rounded px-2 py-1 mt-1">
                驳回原因: {{ req.leader_reject_msg || req.key_holder_reject_msg }}
              </div>
            </div>

            <!-- Right: Actions -->
            <div class="flex items-center gap-2 shrink-0">
              <!-- 团队长审批 -->
              <template v-if="role === 'leader' && req.status === '待审批'">
                <Button size="sm" class="bg-green-600 hover:bg-green-700 text-white text-xs" @click="openApproveDialog(req)">
                  审批
                </Button>
              </template>

              <!-- 钥匙持有人确认 -->
              <template v-if="role === 'key_holder' && req.status === '待双签'">
                <Button
                  size="sm"
                  class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs"
                  :disabled="keyConfirming === req.id"
                  @click="submitKeyHolderConfirm(req.id, true)"
                >
                  <Loader2 v-if="keyConfirming === req.id" class="w-3 h-3 animate-spin mr-1" />
                  ✅ 确认开锁
                </Button>
                <Button
                  size="sm"
                  class="bg-red-100 hover:bg-red-200 text-red-700 text-xs"
                  @click="submitKeyHolderConfirm(req.id, false, '钥匙持有人驳回')"
                >
                  ❌ 驳回
                </Button>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 团队长审批弹窗 -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="approveDialogOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
        <div class="bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 overflow-hidden">
          <div class="bg-gradient-to-r from-green-600 to-emerald-600 px-6 py-4">
            <h3 class="text-white font-semibold text-base">审批领用申请</h3>
            <p class="text-green-100 text-xs mt-0.5">{{ approveTarget?.reagent_item?.reagent_catalog?.name }} · 申请人: {{ approveTarget?.requester?.real_name }}</p>
          </div>
          <div class="px-6 py-5 space-y-4">
            <div class="grid grid-cols-2 gap-3 text-sm">
              <div><span class="text-gray-500">用量:</span> {{ approveTarget?.amount }}</div>
              <div><span class="text-gray-500">用途:</span> {{ approveTarget?.purpose || '-' }}</div>
            </div>

            <!-- 管控品：指定钥匙持有人 -->
            <div v-if="approveTarget?.reagent_item?.reagent_catalog?.is_controlled" class="bg-purple-50 border border-purple-200 rounded-lg px-4 py-3 space-y-3">
              <p class="text-xs font-semibold text-purple-800">🔒 管控品 — 需指定双人双锁持有人</p>
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

            <!-- 驳回原因 -->
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">驳回原因 <span class="text-gray-400">(驳回时必填)</span></label>
              <textarea
                v-model="approveRejectMsg"
                rows="2"
                placeholder="如需驳回，请填写原因..."
                class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500 outline-none"
              ></textarea>
            </div>
          </div>
          <div class="px-6 py-3 bg-gray-50 flex justify-end gap-2 border-t">
            <Button size="sm" class="bg-gray-200 hover:bg-gray-300 text-gray-700" @click="approveDialogOpen = false">取消</Button>
            <Button size="sm" class="bg-red-100 hover:bg-red-200 text-red-700" :disabled="approving || !approveRejectMsg" @click="submitLeaderApproval(false)">
              驳回
            </Button>
            <Button size="sm" class="bg-green-600 hover:bg-green-700 text-white" :disabled="approving" @click="submitLeaderApproval(true)">
              <Loader2 v-if="approving" class="w-3.5 h-3.5 animate-spin mr-1" />
              批准
            </Button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 研发新建领用弹窗 -->
    <Transition
      enter-active-class="transition ease-out duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition ease-in duration-150"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="newDialogOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
        <div class="bg-white rounded-xl shadow-2xl w-full max-w-md mx-4 overflow-hidden">
          <div class="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-4">
            <h3 class="text-white font-semibold text-base">新建领用申请</h3>
          </div>
          <div class="px-6 py-5 space-y-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">试剂 UUID</label>
              <input v-model="newReagentItemId" type="text" placeholder="扫码或粘贴试剂瓶身 UUID" class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">领取量</label>
              <input v-model.number="newAmount" type="number" min="0" step="0.1" placeholder="ml / g" class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1">用途 / 关联实验 <span class="text-gray-400">(选填)</span></label>
              <input v-model="newPurpose" type="text" placeholder="如：拉伸测试实验" class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none" />
            </div>
          </div>
          <div class="px-6 py-3 bg-gray-50 flex justify-end gap-2 border-t">
            <Button size="sm" class="bg-gray-200 hover:bg-gray-300 text-gray-700" @click="newDialogOpen = false">取消</Button>
            <Button size="sm" class="bg-blue-600 hover:bg-blue-700 text-white" :disabled="newSubmitting" @click="submitNewDispense">
              <Loader2 v-if="newSubmitting" class="w-3.5 h-3.5 animate-spin mr-1" />
              提交申请
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
