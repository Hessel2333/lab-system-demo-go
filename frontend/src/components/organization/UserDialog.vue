<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { X } from 'lucide-vue-next'
import type { User, Department } from '@/api/organization'

const props = defineProps<{
  modelValue: boolean
  editUser?: User | null
  department?: Department | null
  departments?: Department[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', data: any): void
}>()

const formData = ref({
    real_name: '',
    username: '',
    role: 'member',
    department_id: 0
})

const roles = [
    { value: 'admin', label: '系统管理员' },
    { value: 'director', label: '部门负责人' },
    { value: 'team_leader', label: '团队长' },
    { value: 'member', label: '团队成员' },
    { value: 'researcher', label: '研发人员' },
    { value: 'procurement', label: '采购人员' },
    { value: 'measurement_specialist', label: '计量专员' },
    { value: 'safety_specialist', label: '安全专员' },
]

interface DepartmentOption {
  id: number
  label: string
}

const buildDepartmentOptions = (nodes: Department[] = [], depth = 0, acc: DepartmentOption[] = []) => {
  for (const node of nodes) {
    const prefix = depth > 0 ? `${'  '.repeat(depth)}└ ` : ''
    acc.push({ id: node.ID, label: `${prefix}${node.name}` })
    if (node.children?.length) {
      buildDepartmentOptions(node.children, depth + 1, acc)
    }
  }
  return acc
}

const departmentOptions = computed(() => buildDepartmentOptions(props.departments || []))

// Sync form with props
watch(() => props.modelValue, (val) => {
    if (val) {
        if (props.editUser) {
            formData.value = {
                real_name: props.editUser.real_name,
                username: props.editUser.username,
                role: props.editUser.role,
                department_id: props.editUser.department_id
            }
        } else {
            // New user defaults
            formData.value = {
                real_name: '',
                username: '',
                role: 'member',
                department_id: props.department?.ID || 0
            }
        }
    }
})

const handleSubmit = () => {
    emit('submit', { 
        ...formData.value,
        id: props.editUser?.ID // Include ID if editing
    })
}

const isEdit = computed(() => !!props.editUser)
</script>

<template>
  <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
    <div class="fixed inset-0 bg-gray-900/30 backdrop-blur-sm transition-opacity" @click="emit('update:modelValue', false)"></div>

    <div class="relative w-full max-w-md transform overflow-hidden rounded-xl bg-white p-6 text-left shadow-xl transition-all sm:my-8 border border-gray-100">
      <div class="flex items-center justify-between mb-5">
          <h3 class="text-lg font-semibold leading-6 text-gray-900">
              {{ isEdit ? '编辑成员' : '添加成员' }} 
              <span v-if="department" class="text-sm font-normal text-gray-500 ml-2">- {{ department.name }}</span>
          </h3>
          <button @click="emit('update:modelValue', false)" class="text-gray-400 hover:text-gray-500 transition-colors">
              <X class="h-5 w-5" />
          </button>
      </div>
      
      <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">真实姓名</label>
              <input v-model="formData.real_name" required type="text" class="block w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border" placeholder="例如：张三">
          </div>
          
           <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">工号/用户名</label>
              <input v-model="formData.username" required type="text" class="block w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border" placeholder="例如：zhangsan">
          </div>
          
          <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">角色</label>
              <select v-model="formData.role" class="block w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border">
                  <option v-for="role in roles" :key="role.value" :value="role.value">
                      {{ role.label }}
                  </option>
              </select>
          </div>

          <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">所属组织</label>
              <select v-model.number="formData.department_id" class="block w-full rounded-lg border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm px-3 py-2 border">
                  <option :value="0" disabled>请选择组织节点</option>
                  <option v-for="dept in departmentOptions" :key="dept.id" :value="dept.id">
                      {{ dept.label }}
                  </option>
              </select>
          </div>

          <div class="mt-6 flex justify-end gap-3 pt-2">
              <button type="button" class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500" @click="emit('update:modelValue', false)">
                  取消
              </button>
              <button type="submit" class="px-4 py-2 text-sm font-medium text-white bg-black border border-transparent rounded-lg hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900 shadow-sm">
                  {{ isEdit ? '保存修改' : '确认添加' }}
              </button>
          </div>
      </form>
    </div>
  </div>
</template>
