<script setup lang="ts">
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import MaintenanceDialog from './MaintenanceDialog.vue'
import { ref } from 'vue'

const tasks = [
    { id: 1, instrument: '离心机 (CF-5000)', type: 'maintenance', due: '2024-03-25', status: 'pending', desc: '季度深度保养' },
    { id: 2, instrument: '液相色谱 (HPLC-01)', type: 'repair', due: '2024-03-24', status: 'in_progress', desc: '基线漂移故障排查' },
]

const dialogOpen = ref(false)
const dialogMode = ref<'maintenance' | 'repair'>('maintenance')

const openMaintenance = () => {
    dialogMode.value = 'maintenance'
    dialogOpen.value = true
}

const openRepair = () => {
    dialogMode.value = 'repair'
    dialogOpen.value = true
}
</script>

<template>
  <div class="space-y-6">
      <div class="flex justify-between items-center">
          <h2 class="text-lg font-semibold">维护保养计划</h2>
          <div class="space-x-2">
               <Button variant="outline" @click="openRepair">新增报修</Button>
               <Button @click="openMaintenance">执行保养</Button>
          </div>
      </div>

      <!-- Task List -->
      <div class="grid gap-4">
          <Card v-for="task in tasks" :key="task.id" class="p-4 flex items-center justify-between">
              <div class="flex items-center gap-4">
                  <div class="w-10 h-10 rounded-full flex items-center justify-center" 
                       :class="task.type === 'maintenance' ? 'bg-blue-100 text-blue-600' : 'bg-red-100 text-red-600'">
                       <span class="font-bold text-xs">{{ task.type === 'maintenance' ? '保' : '修' }}</span>
                  </div>
                  <div>
                      <div class="font-medium">{{ task.instrument }}</div>
                      <div class="text-sm text-gray-500">{{ task.desc }} · 截止: {{ task.due }}</div>
                  </div>
              </div>
              <div class="flex items-center gap-3">
                  <Badge>{{ task.status === 'pending' ? '待执行' : '进行中' }}</Badge>
                  <Button size="sm" variant="outline">处理</Button>
              </div>
          </Card>
      </div>
      
      <MaintenanceDialog :open="dialogOpen" :mode="dialogMode" @close="dialogOpen = false" @submit="dialogOpen = false" />
  </div>
</template>
