<script setup lang="ts">
import Card from '@/components/ui/Card.vue'
import { ClipboardList, Wrench } from 'lucide-vue-next'
</script>

<template>
  <div class="space-y-6">
    <!-- Quick Stats -->
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card class="p-6 flex items-center space-x-4">
            <div class="p-3 bg-blue-100 text-blue-600 rounded-full">
                <ClipboardList class="w-6 h-6" />
            </div>
            <div>
                <p class="text-sm font-medium text-muted-foreground">待验收设备</p>
                <h3 class="text-2xl font-bold">2</h3>
            </div>
        </Card>
        <Card class="p-6 flex items-center space-x-4">
            <div class="p-3 bg-orange-100 text-orange-600 rounded-full">
                <Wrench class="w-6 h-6" />
            </div>
            <div>
                <p class="text-sm font-medium text-muted-foreground">待维护任务</p>
                <h3 class="text-2xl font-bold">5</h3>
            </div>
        </Card>
        <!-- More cards -->
    </div>
    
    <div class="grid gap-6 md:grid-cols-2">
        <Card class="p-6">
            <h3 class="text-lg font-semibold mb-4">待办事项</h3>
            <div class="text-sm text-gray-500">暂无待办事项</div>
        </Card>
        <Card class="p-6">
             <h3 class="text-lg font-semibold mb-4">最近预警</h3>
             <div class="text-sm text-gray-500">系统运行正常</div>
        </Card>
    </div>
  </div>
</template>
