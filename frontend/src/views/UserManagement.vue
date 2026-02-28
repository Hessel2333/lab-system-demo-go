
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { fetchDepartments, fetchUsers, createUser, updateUser, deleteUser, type Department, type User } from '@/api/organization'
import DepartmentTreeItem from '@/components/organization/DepartmentTreeItem.vue'
import UserDialog from '@/components/organization/UserDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import UserPermissionDialog from '@/components/organization/UserPermissionDialog.vue'
import TableSection from '@/components/ui/TableSection.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { Loader2, Search } from 'lucide-vue-next'

const departments = ref<Department[]>([])
const users = ref<User[]>([])
const loading = ref(false)
const selectedDept = ref<Department | null>(null)
const searchQuery = ref('')
const roleFilter = ref('全部')

// Dialog State
const showDialog = ref(false)
const editUser = ref<User | null>(null)

// Confirm Dialog State
const showConfirm = ref(false)
const userToDelete = ref<number | null>(null)

// Permission Dialog State
const showPermissionDialog = ref(false)
const permissionUser = ref<User | null>(null)

// Load sidebar tree
const loadDepartments = async () => {
    try {
        const data = await fetchDepartments()
        departments.value = data
        // Select root by default if available
        if (data.length > 0 && data[0]) {
            selectDepartment(data[0])
        }
    } catch (e) {
        console.error("Failed to load departments", e)
    }
}

// Load users for selected department
const loadUsers = async (deptId?: number) => {
    loading.value = true
    try {
        users.value = await fetchUsers(deptId)
    } catch (e) {
        console.error("Failed to load users", e)
    } finally {
        loading.value = false
    }
}

const selectDepartment = (dept: Department) => {
    selectedDept.value = dept
    loadUsers(dept.ID)
}

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
        if (selectedDept.value) loadUsers(selectedDept.value.ID)
    } catch (e) {
        alert('删除失败')
    } finally {
        showConfirm.value = false
        userToDelete.value = null
    }
}

const handleDialogSubmit = async (data: any) => {
    try {
        if (data.id) {
            await updateUser(data.id, data)
        } else {
            await createUser(data)
        }
        showDialog.value = false
        if (selectedDept.value) loadUsers(selectedDept.value.ID)
    } catch (e) {
        console.error(e)
        alert('操作失败')
    }
}

onMounted(() => {
    loadDepartments()
})

const roleBadgeClass = (role: string) => {
    switch(role) {
        case 'admin': return 'bg-slate-800 text-white'
        case 'team_leader': return 'bg-purple-100 text-purple-700'
        case 'director': return 'bg-red-100 text-red-700'
        case 'procurement': return 'bg-amber-100 text-amber-700'
        case 'member': return 'bg-gray-100 text-gray-700'
        case 'researcher': return 'bg-blue-100 text-blue-700'
        default: return 'bg-gray-100 text-gray-700' 
    }
}

const roleName = (role: string) => {
    const map: Record<string, string> = {
        'admin': '系统管理员',
        'team_leader': '团队长',
        'member': '成员',
        'director': '负责人',
        'procurement': '采购人员',
        'procurement_specialist': '采购人员',
        'measurement_specialist': '计量专员',
        'safety_specialist': '安全专员',
        'researcher': '研发人员'
    }
    return map[role] || role
}

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

const sectionTitle = computed(() => {
    if (!selectedDept.value) return '人员台账'
    return `${selectedDept.value.name} · 人员台账`
})

const sectionDescription = computed(() => `共 ${users.value.length} 位成员`)
</script>

<template>
  <div class="h-full flex gap-6">
      
      <!-- Sidebar: Organization Tree -->
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


      <!-- Main Content: User List -->
      <TableSection class="min-w-0 flex-1" :title="sectionTitle" :description="sectionDescription">
        <template #actions>
          <Button
            variant="primary"
            size="sm"
            :disabled="!selectedDept"
            @click="handleAddUser"
          >
            添加成员
          </Button>
        </template>

        <template #toolbar>
          <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
            <div class="relative w-full sm:w-80">
              <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
              <Input v-model="searchQuery" class="pl-9" placeholder="搜索姓名或用户名..." />
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
        </template>

        <div v-if="loading" class="flex justify-center py-10">
          <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
        </div>
        <div v-else-if="filteredUsers.length === 0" class="apple-table-empty">
          暂无符合条件的人员数据。
        </div>
        <div v-else class="apple-table-wrap">
          <table class="w-full text-left text-sm">
            <thead class="bg-gray-50 text-gray-500 font-medium border-b border-gray-100">
              <tr>
                <th class="px-6 py-3">姓名</th>
                <th class="px-6 py-3">工号/用户名</th>
                <th class="px-6 py-3">角色</th>
                <th class="px-6 py-3">状态</th>
                <th class="px-6 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100">
              <tr v-for="user in filteredUsers" :key="user.ID" class="hover:bg-gray-50/50 transition-colors group">
                <td class="px-6 py-3 font-medium text-gray-900 flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-xs font-bold text-gray-600">
                    {{ user.real_name?.[0] || 'U' }}
                  </div>
                  {{ user.real_name }}
                </td>
                <td class="px-6 py-3 text-gray-500 font-mono text-xs">{{ user.username }}</td>
                <td class="px-6 py-3">
                  <span class="inline-flex items-center px-2 py-1 rounded-md text-xs font-medium" :class="roleBadgeClass(user.role)">
                    {{ roleName(user.role) }}
                  </span>
                </td>
                <td class="px-6 py-3">
                  <span class="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-green-50 text-green-700">
                    <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
                    在职
                  </span>
                </td>
                <td class="px-6 py-3 text-right">
                  <div class="flex justify-end gap-3">
                    <button @click="handlePermission(user)" class="text-indigo-600 hover:text-indigo-800 font-medium text-xs opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1">
                      权限
                    </button>
                    <button @click="handleEditUser(user)" class="text-blue-600 hover:text-blue-800 font-medium text-xs opacity-0 group-hover:opacity-100 transition-opacity">
                      编辑
                    </button>
                    <button @click="handleDeleteUser(user.ID)" class="text-red-400 hover:text-red-600 font-medium text-xs opacity-0 group-hover:opacity-100 transition-opacity">
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </TableSection>
      
      <UserDialog 
         v-model="showDialog"
         :edit-user="editUser"
         :department="selectedDept"
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
