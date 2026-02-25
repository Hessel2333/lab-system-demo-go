<script setup lang="ts">
import { Building2, Shield, Users as UsersIcon } from 'lucide-vue-next'
import type { Department } from '@/api/organization'

defineProps<{
  model: Department
  selectedId?: number
}>()

const emit = defineEmits<{
  (e: 'select', dept: Department): void
}>()
</script>

<template>
  <li>
    <div 
        class="flex items-center gap-2 px-3 py-2 rounded-md cursor-pointer text-sm transition-colors"
        :class="selectedId === model.ID ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-600 hover:bg-gray-50'"
        @click="emit('select', model)"
    >
        <span v-if="model.type === 'institute'" class="text-blue-600"><Building2 :size="16"/></span>
        <span v-else-if="model.type === 'team'" class="text-indigo-500"><UsersIcon :size="16"/></span>
        <span v-else class="text-gray-400"><Shield :size="16"/></span>
        
        {{ model.name }}
    </div>
    <ul v-if="model.children && model.children.length" class="pl-4 mt-1 border-l border-gray-100 ml-3">
        <DepartmentTreeItem 
            v-for="child in model.children" 
            :key="child.ID" 
            :model="child" 
            :selected-id="selectedId"
            @select="emit('select', $event)"
        />
    </ul>
  </li>
</template>

<script lang="ts">
export default {
    name: 'DepartmentTreeItem'
}
</script>
