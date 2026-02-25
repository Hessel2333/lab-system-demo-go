
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchDepartments, fetchUsers, createUser, updateUser, deleteUser, type Department, type User } from '@/api/organization'
import DepartmentTreeItem from '@/components/organization/DepartmentTreeItem.vue'
import UserDialog from '@/components/organization/UserDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import UserPermissionDialog from '@/components/organization/UserPermissionDialog.vue'

const departments = ref<Department[]>([])
const users = ref<User[]>([])
const loading = ref(false)
const selectedDept = ref<Department | null>(null)

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
        case 'team_leader': return 'bg-purple-100 text-purple-700'
        case 'director': return 'bg-red-100 text-red-700'
        case 'member': return 'bg-gray-100 text-gray-700'
        default: return 'bg-blue-100 text-blue-700' // Specialists
    }
}

const roleName = (role: string) => {
    const map: Record<string, string> = {
        'team_leader': '团队长',
        'member': '成员',
        'director': '负责人',
        'procurement_specialist': '采购专员',
        'measurement_specialist': '计量专员',
        'safety_specialist': '安全专员'
    }
    return map[role] || role
}
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
      <div class="flex-1 bg-white border border-gray-200 rounded-xl shadow-sm flex flex-col min-w-0">
          <!-- Header -->
          <div class="p-6 border-b border-gray-100 flex justify-between items-center">
              <div>
                  <h1 class="text-xl font-bold text-gray-900 flex items-center gap-2">
                       {{ selectedDept?.name || '所有人员' }}
                       <span v-if="selectedDept" class="px-2 py-0.5 rounded-full bg-gray-100 text-xs font-normal text-gray-500">{{ selectedDept.type }}</span>
                  </h1>
                  <p class="text-sm text-gray-500 mt-1">
                      共 {{ users.length }} 位成员
                  </p>
              </div>
              <button 
                @click="handleAddUser"
                :disabled="!selectedDept"
                class="bg-black text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
              >
                  添加成员
              </button>
          </div>

          <!-- Table -->
          <div class="flex-1 overflow-auto p-0">
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
                      <tr v-for="user in users" :key="user.ID" class="hover:bg-gray-50/50 transition-colors group">
                          <td class="px-6 py-3 font-medium text-gray-900 flex items-center gap-3">
                              <div class="w-8 h-8 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-xs font-bold text-gray-600">
                                  {{ user.real_name[0] }}
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
                          <td class="px-6 py-3 text-right flex justify-end gap-3">
                              <button @click="handlePermission(user)" class="text-indigo-600 hover:text-indigo-800 font-medium text-xs opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1">
                                  权限
                              </button>
                              <button @click="handleEditUser(user)" class="text-blue-600 hover:text-blue-800 font-medium text-xs opacity-0 group-hover:opacity-100 transition-opacity">
                                  编辑
                              </button>
                              <button @click="handleDeleteUser(user.ID)" class="text-red-400 hover:text-red-600 font-medium text-xs opacity-0 group-hover:opacity-100 transition-opacity">
                                  删除
                              </button>
                          </td>
                      </tr>
                      <tr v-if="users.length === 0 && !loading">
                          <td colspan="5" class="px-6 py-12 text-center text-gray-400">
                              暂无人员数据
                          </td>
                      </tr>
                  </tbody>
              </table>
          </div>
      </div>
      
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
