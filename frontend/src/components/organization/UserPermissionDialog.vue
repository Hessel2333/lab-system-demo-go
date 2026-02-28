<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { X, Shield, Lock, UserCog, Building2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import type { User } from '@/api/organization'
import {
  fetchUserPermissions,
  updateUserPermission,
  fetchUserReagentPermissions,
  updateUserReagentPermissions,
  fetchUsers,
  type InstrumentPermission,
} from '@/api/organization'
import Button from '@/components/ui/Button.vue'

const props = defineProps<{
  modelValue: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

type PermissionTab = 'instruments' | 'reagent'

const activeTab = ref<PermissionTab>('instruments')
const permissions = ref<InstrumentPermission[]>([])
const loading = ref(false)
const savingInstrumentId = ref<number | null>(null)

const reagentPermLoading = ref(false)
const savingReagentPerm = ref(false)
const reagentPerm = ref({
  is_dispense_key_holder_a: false,
  is_dispense_key_holder_b: false,
})

const keyHolderAName = ref('-')
const keyHolderBName = ref('-')

const roleName = (role?: string) => {
  const map: Record<string, string> = {
    admin: '系统管理员',
    team_leader: '团队长',
    member: '成员',
    director: '负责人',
    procurement: '采购人员',
    procurement_specialist: '采购人员',
    measurement_specialist: '计量专员',
    safety_specialist: '安全专员',
    researcher: '研发人员',
  }
  return map[String(role || '')] || (role || '-')
}

const sortedPermissions = computed(() => [...permissions.value].sort((a, b) => a.instrument_name.localeCompare(b.instrument_name, 'zh-CN')))

const roleFlowCapabilities = computed(() => {
  const role = String(props.user?.role || '')
  if (!role) return ['无流程权限信息']
  if (role === 'admin') {
    return ['全流程管理（配置/审批/执行）', '可维护用户与权限策略']
  }
  if (role === 'team_leader' || role === 'director') {
    return ['申购审批', '管控领用审批', '团队流程监督']
  }
  if (role === 'procurement' || role === 'procurement_specialist') {
    return ['采购导入与匹配', '到货确认与赋码', '采购链路执行']
  }
  if (role === 'researcher' || role === 'member') {
    return ['提交申购', '到货入库', '发起领用申请']
  }
  if (role === 'measurement_specialist' || role === 'safety_specialist') {
    return ['专项流程支持（按模块授权）']
  }
  return ['按组织角色默认授权']
})

const loadPermissions = async () => {
  if (!props.user) return
  loading.value = true
  try {
    permissions.value = await fetchUserPermissions(props.user.ID)
  } catch (e) {
    console.error(e)
    toast.error('加载仪器权限失败')
  } finally {
    loading.value = false
  }
}

const loadReagentPermissions = async () => {
  if (!props.user) return
  reagentPermLoading.value = true
  try {
    const data = await fetchUserReagentPermissions(props.user.ID)
    reagentPerm.value = {
      is_dispense_key_holder_a: !!data.is_dispense_key_holder_a,
      is_dispense_key_holder_b: !!data.is_dispense_key_holder_b,
    }
  } catch (e) {
    console.error(e)
    toast.error('加载试剂双签权限失败')
  } finally {
    reagentPermLoading.value = false
  }
}

const loadGlobalKeyHolders = async () => {
  try {
    const allUsers = await fetchUsers()
    const holderA = allUsers.find((u) => !!u.is_dispense_key_holder_a)
    const holderB = allUsers.find((u) => !!u.is_dispense_key_holder_b)
    keyHolderAName.value = holderA?.real_name || '-'
    keyHolderBName.value = holderB?.real_name || '-'
  } catch (e) {
    keyHolderAName.value = '-'
    keyHolderBName.value = '-'
  }
}

watch(
  () => props.modelValue,
  async (val) => {
    if (!val || !props.user) return
    activeTab.value = 'instruments'
    await Promise.all([loadPermissions(), loadReagentPermissions(), loadGlobalKeyHolders()])
  }
)

const togglePermission = async (perm: InstrumentPermission) => {
  if (!props.user) return
  const newStatus = !perm.has_permission
  const prevStatus = perm.has_permission
  perm.has_permission = newStatus
  savingInstrumentId.value = perm.instrument_id
  try {
    await updateUserPermission(props.user.ID, perm.instrument_id, newStatus)
  } catch (e) {
    perm.has_permission = prevStatus
    toast.error('更新仪器权限失败')
    console.error(e)
  } finally {
    savingInstrumentId.value = null
  }
}

const saveReagentPermissions = async () => {
  if (!props.user) return
  if (reagentPerm.value.is_dispense_key_holder_a && reagentPerm.value.is_dispense_key_holder_b) {
    toast.error('同一用户不能同时担任A/B双签持有人')
    return
  }

  savingReagentPerm.value = true
  try {
    await updateUserReagentPermissions(props.user.ID, reagentPerm.value)
    await loadGlobalKeyHolders()
    toast.success('已保存试剂双签权限')
  } catch (e) {
    console.error(e)
    toast.error('保存失败')
  } finally {
    savingReagentPerm.value = false
  }
}
</script>

<template>
  <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6" style="z-index: 100;">
    <div class="fixed inset-0 bg-gray-900/30 backdrop-blur-sm transition-opacity" @click="emit('update:modelValue', false)"></div>

    <div class="relative w-full max-w-3xl transform overflow-hidden rounded-xl bg-white flex flex-col max-h-[86vh] shadow-2xl transition-all sm:my-8 border border-gray-100">
      <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between bg-gray-50/50">
        <div class="flex items-center gap-3 min-w-0">
          <div class="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 shrink-0">
            <UserCog class="w-5 h-5" />
          </div>
          <div class="min-w-0">
            <h3 class="text-lg font-bold text-gray-900">权限编辑</h3>
            <p class="text-sm text-gray-500 truncate" v-if="user">
              {{ user.real_name }} · {{ roleName(user.role) }} · {{ user.department?.name || '未分配组织' }}
            </p>
          </div>
        </div>
        <button @click="emit('update:modelValue', false)" class="text-gray-400 hover:text-gray-500 transition-colors p-1 rounded-full hover:bg-gray-100">
          <X class="h-5 w-5" />
        </button>
      </div>

      <div class="flex border-b border-gray-100 px-6 gap-6">
        <button
          @click="activeTab = 'instruments'"
          class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2"
          :class="activeTab === 'instruments' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
        >
          <Shield class="w-4 h-4" /> 仪器预约权限
        </button>
        <button
          @click="activeTab = 'reagent'"
          class="py-3 text-sm font-medium border-b-2 transition-colors flex items-center gap-2"
          :class="activeTab === 'reagent' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'"
        >
          <Lock class="w-4 h-4" /> 试剂双签权限
        </button>
      </div>

      <div class="flex-1 overflow-y-auto min-h-[360px]">
        <div v-if="activeTab === 'instruments'" class="p-6 space-y-4">
          <div class="rounded-lg border border-gray-100 bg-gray-50 px-4 py-3 text-xs text-gray-600">
            仪器权限用于控制预约资格，不影响试剂流程审批权限。
          </div>

          <div v-if="loading" class="flex items-center justify-center h-40 text-gray-400 text-sm">加载中...</div>
          <div v-else-if="sortedPermissions.length === 0" class="rounded-lg border border-dashed border-gray-200 px-4 py-10 text-center text-sm text-gray-400">
            暂无仪器权限数据
          </div>
          <div v-else class="rounded-xl border border-gray-100 divide-y divide-gray-100 overflow-hidden">
            <div v-for="perm in sortedPermissions" :key="perm.instrument_id" class="flex items-center justify-between px-4 py-3 hover:bg-gray-50 transition-colors">
              <div class="min-w-0">
                <h4 class="text-sm font-medium text-gray-900 truncate">{{ perm.instrument_name }}</h4>
                <p class="text-xs text-gray-500 mt-0.5">仪器 ID: {{ perm.instrument_id }}</p>
              </div>
              <button
                @click="togglePermission(perm)"
                :disabled="savingInstrumentId === perm.instrument_id"
                class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out disabled:opacity-60"
                :class="perm.has_permission ? 'bg-green-500' : 'bg-gray-200'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                  :class="perm.has_permission ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>
          </div>
        </div>

        <div v-else class="p-6 space-y-4">
          <div class="rounded-lg border border-slate-200 bg-slate-50 px-4 py-3">
            <div class="text-xs font-semibold text-slate-700 mb-2">角色驱动权限（只读）</div>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="capability in roleFlowCapabilities"
                :key="capability"
                class="inline-flex items-center rounded-md bg-white border border-slate-200 px-2 py-1 text-[11px] text-slate-600"
              >
                {{ capability }}
              </span>
            </div>
            <p class="mt-2 text-[11px] text-slate-500">以上权限由“组织角色”自动决定；如需调整请编辑成员角色。</p>
          </div>

          <div class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-700">
            这里仅配置双签持有人。A/B 为系统全局唯一角色，保存后会自动替换原持有人。
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div class="rounded-lg border border-gray-100 bg-white px-4 py-3">
              <div class="text-xs text-gray-500">当前持有人 A</div>
              <div class="mt-1 text-sm font-semibold text-gray-900 flex items-center gap-2"><Building2 class="w-4 h-4 text-gray-400" />{{ keyHolderAName }}</div>
            </div>
            <div class="rounded-lg border border-gray-100 bg-white px-4 py-3">
              <div class="text-xs text-gray-500">当前持有人 B</div>
              <div class="mt-1 text-sm font-semibold text-gray-900 flex items-center gap-2"><Building2 class="w-4 h-4 text-gray-400" />{{ keyHolderBName }}</div>
            </div>
          </div>

          <div v-if="reagentPermLoading" class="py-10 text-sm text-gray-400 text-center">加载中...</div>
          <div v-else class="space-y-3">
            <div class="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3">
              <div>
                <div class="text-sm font-medium text-gray-900">钥匙持有人 A</div>
                <div class="text-xs text-gray-500">建议用于团队长审批链路</div>
              </div>
              <button
                @click="reagentPerm.is_dispense_key_holder_a = !reagentPerm.is_dispense_key_holder_a; if (reagentPerm.is_dispense_key_holder_a) reagentPerm.is_dispense_key_holder_b = false"
                class="relative inline-flex h-6 w-11 rounded-full border-2 border-transparent transition-colors"
                :class="reagentPerm.is_dispense_key_holder_a ? 'bg-green-500' : 'bg-gray-200'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow transition"
                  :class="reagentPerm.is_dispense_key_holder_a ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>

            <div class="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3">
              <div>
                <div class="text-sm font-medium text-gray-900">钥匙持有人 B</div>
                <div class="text-xs text-gray-500">建议用于采购侧双签链路</div>
              </div>
              <button
                @click="reagentPerm.is_dispense_key_holder_b = !reagentPerm.is_dispense_key_holder_b; if (reagentPerm.is_dispense_key_holder_b) reagentPerm.is_dispense_key_holder_a = false"
                class="relative inline-flex h-6 w-11 rounded-full border-2 border-transparent transition-colors"
                :class="reagentPerm.is_dispense_key_holder_b ? 'bg-green-500' : 'bg-gray-200'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow transition"
                  :class="reagentPerm.is_dispense_key_holder_b ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>

            <div class="pt-2">
              <Button variant="primary" :disabled="savingReagentPerm" @click="saveReagentPermissions">
                {{ savingReagentPerm ? '保存中...' : '保存试剂双签权限' }}
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div class="px-6 py-4 bg-gray-50 border-t border-gray-100 flex justify-end">
        <Button variant="outline" @click="emit('update:modelValue', false)">关闭</Button>
      </div>
    </div>
  </div>
</template>
