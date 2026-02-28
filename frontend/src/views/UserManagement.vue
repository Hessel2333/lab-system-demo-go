<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  fetchDepartments,
  fetchUsers,
  createUser,
  updateUser,
  deleteUser,
  type Department,
  type User,
} from '@/api/organization'
import DepartmentTreeItem from '@/components/organization/DepartmentTreeItem.vue'
import UserDialog from '@/components/organization/UserDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import UserPermissionDialog from '@/components/organization/UserPermissionDialog.vue'
import TableSection from '@/components/ui/TableSection.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import LedgerTable from '@/components/reagents/LedgerTable.vue'
import { Loader2, Search } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const departments = ref<Department[]>([])
const users = ref<User[]>([])
const allUsers = ref<User[]>([])
const loading = ref(false)
const selectedDept = ref<Department | null>(null)
const searchQuery = ref('')
const roleFilter = ref('全部')

const showDialog = ref(false)
const editUser = ref<User | null>(null)
const showConfirm = ref(false)
const userToDelete = ref<number | null>(null)
const showPermissionDialog = ref(false)
const permissionUser = ref<User | null>(null)

const userColumns = [
  { key: 'name', label: '姓名' },
  { key: 'username', label: '工号/用户名' },
  { key: 'role', label: '组织角色' },
  { key: 'flow_role', label: '试剂流程角色' },
  { key: 'status', label: '状态' },
  { key: 'actions', label: '操作', align: 'right' as const },
]

const roleMeta: Record<string, { label: string; className: string }> = {
  admin: { label: '系统管理员', className: 'bg-slate-800 text-white' },
  team_leader: { label: '团队长', className: 'bg-purple-100 text-purple-700' },
  director: { label: '负责人', className: 'bg-red-100 text-red-700' },
  procurement: { label: '采购人员', className: 'bg-amber-100 text-amber-700' },
  member: { label: '成员', className: 'bg-gray-100 text-gray-700' },
  researcher: { label: '研发人员', className: 'bg-blue-100 text-blue-700' },
  procurement_specialist: { label: '采购人员', className: 'bg-amber-100 text-amber-700' },
  measurement_specialist: { label: '计量专员', className: 'bg-emerald-100 text-emerald-700' },
  safety_specialist: { label: '安全专员', className: 'bg-orange-100 text-orange-700' },
}

const roleName = (role: string) => roleMeta[role]?.label || role || '-'
const roleBadgeClass = (role: string) => roleMeta[role]?.className || 'bg-gray-100 text-gray-700'

const getFlowRoles = (user: User) => {
  const tags: Array<{ label: string; className: string }> = []
  if (user.role === 'team_leader') {
    tags.push({ label: '团队长审批', className: 'bg-indigo-100 text-indigo-700' })
  }
  if (user.role === 'procurement' || user.role === 'procurement_specialist') {
    tags.push({ label: '采购执行', className: 'bg-amber-100 text-amber-700' })
  }
  if (user.is_dispense_key_holder_a) {
    tags.push({ label: '双签A', className: 'bg-emerald-100 text-emerald-700' })
  }
  if (user.is_dispense_key_holder_b) {
    tags.push({ label: '双签B', className: 'bg-orange-100 text-orange-700' })
  }
  if (tags.length === 0) {
    tags.push({ label: '无特殊流程角色', className: 'bg-slate-100 text-slate-500' })
  }
  return tags
}

const loadDepartments = async () => {
  try {
    const data = await fetchDepartments()
    departments.value = data
    if (!selectedDept.value && data.length > 0) {
      selectedDept.value = data[0] || null
    }
  } catch (e) {
    console.error('Failed to load departments', e)
    toast.error('组织架构加载失败')
  }
}

const loadUsers = async () => {
  if (!selectedDept.value) return
  loading.value = true
  try {
    users.value = await fetchUsers(selectedDept.value.ID, false)
  } catch (e) {
    console.error('Failed to load users', e)
    toast.error('人员台账加载失败')
  } finally {
    loading.value = false
  }
}

const loadAllUsers = async () => {
  try {
    allUsers.value = await fetchUsers()
  } catch (e) {
    console.error('Failed to load all users', e)
  }
}

const refreshData = async () => {
  await Promise.all([loadUsers(), loadAllUsers()])
}

const selectDepartment = (dept: Department) => {
  selectedDept.value = dept
  loadUsers()
}

watch(showPermissionDialog, (open) => {
  if (!open) {
    loadAllUsers()
    loadUsers()
  }
})

const handleAddUser = () => {
  if (!selectedDept.value) return
  editUser.value = null
  showDialog.value = true
}

const handleEditUser = (user: User) => {
  editUser.value = user
  showDialog.value = true
}

const handleDeleteUser = (id: number) => {
  userToDelete.value = id
  showConfirm.value = true
}

const handlePermission = (user: User) => {
  permissionUser.value = user
  showPermissionDialog.value = true
}

const confirmDelete = async () => {
  if (userToDelete.value === null) return
  try {
    await deleteUser(userToDelete.value)
    toast.success('成员已删除')
    await refreshData()
  } catch (e) {
    toast.error('删除失败')
  } finally {
    showConfirm.value = false
    userToDelete.value = null
  }
}

const handleDialogSubmit = async (data: any) => {
  try {
    if (data.id) {
      await updateUser(data.id, data)
      toast.success('成员信息已更新')
    } else {
      await createUser(data)
      toast.success('成员已创建')
    }
    showDialog.value = false
    await refreshData()
  } catch (e) {
    console.error(e)
    toast.error('操作失败')
  }
}

onMounted(async () => {
  await loadDepartments()
  await refreshData()
})

const roleOptions = computed(() => {
  const unique = new Set(users.value.map((user) => roleName(user.role)))
  return ['全部', ...Array.from(unique)]
})

const filteredUsers = computed(() => {
  let result = users.value
  if (roleFilter.value !== '全部') {
    result = result.filter((user) => roleName(user.role) === roleFilter.value)
  }
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    result = result.filter((user) =>
      user.real_name?.toLowerCase().includes(q) ||
      user.username?.toLowerCase().includes(q)
    )
  }
  return result
})

const keyHolderA = computed(() => allUsers.value.find((u) => !!u.is_dispense_key_holder_a))
const keyHolderB = computed(() => allUsers.value.find((u) => !!u.is_dispense_key_holder_b))

const sectionTitle = computed(() => {
  if (!selectedDept.value) return '人员台账'
  return `${selectedDept.value.name} · 人员台账`
})

const sectionDescription = computed(() => {
  return `共 ${users.value.length} 人`
})
</script>

<template>
  <div class="h-full flex gap-6">
    <div class="w-64 flex-shrink-0 bg-white border border-gray-200 rounded-xl shadow-sm flex flex-col p-4">
      <h2 class="text-sm font-semibold text-gray-900 mb-4 px-2 tracking-tight">组织架构</h2>
      <ul class="space-y-1 overflow-y-auto flex-1">
        <DepartmentTreeItem
          v-for="root in departments"
          :key="root.ID"
          :model="root"
          :selected-id="selectedDept?.ID"
          @select="selectDepartment"
        />
      </ul>
    </div>

    <TableSection class="min-w-0 flex-1" :title="sectionTitle" :description="sectionDescription">
      <template #actions>
        <Button variant="primary" size="sm" :disabled="!selectedDept" @click="handleAddUser">
          添加成员
        </Button>
      </template>

      <template #toolbar>
        <div class="flex w-full flex-col gap-3">
          <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
            <div class="relative w-full sm:w-80">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
              <Input v-model="searchQuery" class="pl-9" placeholder="搜索姓名或用户名..." />
            </div>
            <div class="text-xs text-gray-500 sm:ml-auto">当前显示所选组织直属成员</div>
          </div>
          <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
            <div class="flex items-center gap-2 text-xs text-gray-500">
              <span class="inline-flex items-center rounded-md bg-slate-100 px-2 py-1">双签A：{{ keyHolderA?.real_name || '-' }}</span>
              <span class="inline-flex items-center rounded-md bg-slate-100 px-2 py-1">双签B：{{ keyHolderB?.real_name || '-' }}</span>
            </div>
            <div class="apple-segmented w-fit sm:ml-auto">
              <button
                v-for="option in roleOptions"
                :key="option"
                @click="roleFilter = option"
                :class="['apple-segmented-btn', roleFilter === option ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']"
              >
                {{ option }}
              </button>
            </div>
          </div>
        </div>
      </template>

      <div v-if="loading" class="flex justify-center py-10">
        <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
      </div>
      <div v-else-if="filteredUsers.length === 0" class="apple-table-empty">
        暂无符合条件的人员数据。
      </div>
      <LedgerTable v-else :columns="userColumns">
        <tr v-for="user in filteredUsers" :key="user.ID" class="hover:bg-gray-50/50 transition-colors">
          <td class="px-6 py-3 font-medium text-gray-900">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-xs font-bold text-gray-600">
                {{ user.real_name?.[0] || 'U' }}
              </div>
              <span>{{ user.real_name }}</span>
            </div>
          </td>
          <td class="px-6 py-3 text-gray-500 font-mono text-xs">{{ user.username }}</td>
          <td class="px-6 py-3">
            <span class="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium" :class="roleBadgeClass(user.role)">
              {{ roleName(user.role) }}
            </span>
          </td>
          <td class="px-6 py-3">
            <div class="flex flex-wrap items-center gap-1.5">
              <span
                v-for="tag in getFlowRoles(user)"
                :key="tag.label"
                class="inline-flex items-center rounded px-2 py-0.5 text-[11px] font-semibold"
                :class="tag.className"
              >
                {{ tag.label }}
              </span>
            </div>
          </td>
          <td class="px-6 py-3">
            <span class="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-green-50 text-green-700">
              <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
              在职
            </span>
          </td>
          <td class="px-6 py-3 text-right">
            <div class="flex justify-end gap-2">
              <Button size="sm" variant="outline" class="h-7 px-2 text-[11px]" @click="handlePermission(user)">
                权限
              </Button>
              <Button size="sm" variant="outline" class="h-7 px-2 text-[11px]" @click="handleEditUser(user)">
                编辑
              </Button>
              <Button size="sm" variant="destructive" class="h-7 px-2 text-[11px]" @click="handleDeleteUser(user.ID)">
                删除
              </Button>
            </div>
          </td>
        </tr>
      </LedgerTable>
    </TableSection>

    <UserDialog
      v-model="showDialog"
      :edit-user="editUser"
      :department="selectedDept"
      :departments="departments"
      @submit="handleDialogSubmit"
    />

    <ConfirmDialog
      v-model="showConfirm"
      title="确认删除成员"
      message="删除后该成员将无法登录系统，是否继续？"
      confirm-text="彻底删除"
      @confirm="confirmDelete"
    />

    <UserPermissionDialog
      v-model="showPermissionDialog"
      :user="permissionUser"
    />
  </div>
</template>
