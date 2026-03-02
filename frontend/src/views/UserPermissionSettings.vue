<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RefreshCw, ShieldCheck, UserRoundCheck } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import {
  fetchUsers,
  updateUserReagentPermissions,
  type User,
} from '@/api/organization'
import TableSection from '@/components/ui/TableSection.vue'
import Button from '@/components/ui/Button.vue'
import LedgerTable from '@/components/reagents/LedgerTable.vue'

const loading = ref(false)
const saving = ref(false)
const users = ref<User[]>([])
const selectedAId = ref<number | null>(null)
const selectedBId = ref<number | null>(null)

const roleMap: Record<string, string> = {
  admin: '系统管理员',
  team_leader: '团队长',
  director: '负责人',
  procurement: '采购人员',
  procurement_specialist: '采购人员',
  member: '成员',
  researcher: '研发人员',
  measurement_specialist: '计量专员',
  safety_specialist: '安全专员',
}

const roleName = (role?: string) => roleMap[String(role || '')] || role || '-'

const tableColumns = [
  { key: 'name', label: '成员' },
  { key: 'role', label: '组织角色' },
  { key: 'department', label: '所属组织' },
  { key: 'dual_sign', label: '双签状态' },
]

const sortedUsers = computed(() => {
  return [...users.value].sort((a, b) => {
    const byDept = String(a.department?.name || '').localeCompare(String(b.department?.name || ''), 'zh-CN')
    if (byDept !== 0) return byDept
    return String(a.real_name || '').localeCompare(String(b.real_name || ''), 'zh-CN')
  })
})

const holderA = computed(() => users.value.find((user) => user.ID === selectedAId.value) || null)
const holderB = computed(() => users.value.find((user) => user.ID === selectedBId.value) || null)

const loadUsersData = async () => {
  loading.value = true
  try {
    const data = await fetchUsers()
    users.value = data
    const currentA = data.find((user) => !!user.is_dispense_key_holder_a)
    const currentB = data.find((user) => !!user.is_dispense_key_holder_b)
    selectedAId.value = currentA?.ID ?? null
    selectedBId.value = currentB?.ID ?? null
  } catch (error) {
    console.error(error)
    toast.error('加载成员列表失败')
  } finally {
    loading.value = false
  }
}

const saveDualSignHolders = async () => {
  if (!selectedAId.value || !selectedBId.value) {
    toast.error('请先设置A角和B角持有人')
    return
  }
  if (selectedAId.value === selectedBId.value) {
    toast.error('A角和B角不能是同一人')
    return
  }

  const updateTasks: Array<Promise<unknown>> = []
  for (const user of users.value) {
    const nextA = user.ID === selectedAId.value
    const nextB = user.ID === selectedBId.value
    const changed =
      !!user.is_dispense_key_holder_a !== nextA ||
      !!user.is_dispense_key_holder_b !== nextB
    if (!changed) continue
    updateTasks.push(
      updateUserReagentPermissions(user.ID, {
        is_dispense_key_holder_a: nextA,
        is_dispense_key_holder_b: nextB,
      })
    )
  }

  if (updateTasks.length === 0) {
    toast.info('未检测到变更')
    return
  }

  saving.value = true
  try {
    await Promise.all(updateTasks)
    toast.success('权限策略已更新')
    await loadUsersData()
  } catch (error) {
    console.error(error)
    toast.error('保存失败，请稍后重试')
  } finally {
    saving.value = false
  }
}

onMounted(loadUsersData)
</script>

<template>
  <div class="h-full flex flex-col gap-6">
    <TableSection
      title="权限策略中心"
      description="全局策略统一在此维护。当前已启用：试剂双签角色。"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="gap-1.5" :disabled="loading" @click="loadUsersData">
          <RefreshCw class="h-3.5 w-3.5" />
          刷新
        </Button>
      </template>

      <div class="space-y-4">
        <div class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-xs text-amber-700">
          双签角色为全局策略，不在单个成员权限弹窗中单独编辑。
        </div>

        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
          <div class="rounded-xl border border-slate-200 bg-white p-4">
            <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-slate-900">
              <ShieldCheck class="h-4 w-4 text-emerald-600" />
              A角持有人
            </div>
            <select
              v-model.number="selectedAId"
              class="h-10 w-full rounded-xl border border-slate-200 px-3 text-sm text-slate-700 focus:border-blue-400 focus:outline-none focus:ring-4 focus:ring-blue-500/10"
            >
              <option :value="null">请选择A角成员</option>
              <option v-for="user in sortedUsers" :key="user.ID" :value="user.ID">
                {{ user.real_name }}（{{ roleName(user.role) }}）
              </option>
            </select>
            <p class="mt-2 text-xs text-slate-500">
              当前：{{ holderA?.real_name || '未配置' }}
            </p>
          </div>

          <div class="rounded-xl border border-slate-200 bg-white p-4">
            <div class="mb-2 flex items-center gap-2 text-sm font-semibold text-slate-900">
              <UserRoundCheck class="h-4 w-4 text-blue-600" />
              B角持有人
            </div>
            <select
              v-model.number="selectedBId"
              class="h-10 w-full rounded-xl border border-slate-200 px-3 text-sm text-slate-700 focus:border-blue-400 focus:outline-none focus:ring-4 focus:ring-blue-500/10"
            >
              <option :value="null">请选择B角成员</option>
              <option v-for="user in sortedUsers" :key="user.ID" :value="user.ID">
                {{ user.real_name }}（{{ roleName(user.role) }}）
              </option>
            </select>
            <p class="mt-2 text-xs text-slate-500">
              当前：{{ holderB?.real_name || '未配置' }}
            </p>
          </div>
        </div>

        <div class="flex justify-end">
          <Button variant="primary" :disabled="saving || loading" @click="saveDualSignHolders">
            {{ saving ? '保存中...' : '保存策略' }}
          </Button>
        </div>
      </div>
    </TableSection>

    <TableSection
      title="策略模块规划"
      description="后续可在本页扩展审批链、通知范围、超时规则等全局策略。"
    >
      <div class="rounded-xl border border-dashed border-slate-200 bg-slate-50 px-4 py-6 text-sm text-slate-500">
        暂未启用更多策略模块。
      </div>
    </TableSection>

    <TableSection
      title="成员双签映射"
      description="仅用于查看当前配置映射，便于确认A/B角是否正确。"
    >
      <div v-if="loading" class="apple-table-empty">加载中...</div>
      <div v-else-if="sortedUsers.length === 0" class="apple-table-empty">暂无成员数据</div>
      <LedgerTable v-else :columns="tableColumns">
        <tr
          v-for="user in sortedUsers"
          :key="user.ID"
          class="hover:bg-gray-50/50 transition-colors"
        >
          <td class="px-6 py-3">
            <div class="font-medium text-slate-900">{{ user.real_name }}</div>
            <div class="text-xs font-mono text-slate-500">{{ user.username }}</div>
          </td>
          <td class="px-6 py-3 text-sm text-slate-700">{{ roleName(user.role) }}</td>
          <td class="px-6 py-3 text-sm text-slate-700">{{ user.department?.name || '-' }}</td>
          <td class="px-6 py-3">
            <div class="flex items-center gap-2">
              <span
                v-if="user.ID === selectedAId"
                class="inline-flex items-center rounded-md bg-emerald-50 border border-emerald-200 px-2 py-1 text-[11px] font-semibold text-emerald-700"
              >
                A角
              </span>
              <span
                v-if="user.ID === selectedBId"
                class="inline-flex items-center rounded-md bg-blue-50 border border-blue-200 px-2 py-1 text-[11px] font-semibold text-blue-700"
              >
                B角
              </span>
              <span
                v-if="user.ID !== selectedAId && user.ID !== selectedBId"
                class="text-xs text-slate-400"
              >
                -
              </span>
            </div>
          </td>
        </tr>
      </LedgerTable>
    </TableSection>
  </div>
</template>
