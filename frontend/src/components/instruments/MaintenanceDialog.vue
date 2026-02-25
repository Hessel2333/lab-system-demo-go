<script setup lang="ts">
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import { ref } from 'vue'
import { CheckSquare, AlertTriangle } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  mode: 'maintenance' | 'repair' // 维护保养 vs 故障报修
}>()

defineEmits<{
  (e: 'close'): void
  (e: 'submit', data: any): void
}>()

const checklist = ref([
    { id: 1, label: '清洁进样针', checked: false },
    { id: 2, label: '检查管路气密性', checked: false },
    { id: 3, label: '更换密封垫', checked: false },
    { id: 4, label: '校准基线', checked: false },
])

const repairForm = ref({
    issue: '',
    urgency: 'normal'
})
</script>

<template>
  <Dialog :open="open" @close="$emit('close')" :title="mode === 'maintenance' ? '执行定期保养' : '故障报修申请'" class="p-6">
      <div v-if="mode === 'maintenance'" class="space-y-4">
          <div class="bg-blue-50 p-3 rounded-lg flex items-center gap-2 text-sm text-blue-700 mb-4">
              <CheckSquare class="w-4 h-4" />
              请按照SOP-2024-001执行以下保养步骤
          </div>
          <div class="space-y-2 border rounded-xl p-4">
              <label v-for="item in checklist" :key="item.id" class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg cursor-pointer">
                  <input type="checkbox" v-model="item.checked" class="w-5 h-5 rounded border-gray-300 text-blue-600" />
                  <span class="font-medium">{{ item.label }}</span>
              </label>
          </div>
          <div>
              <label class="block text-sm font-medium mb-2">保养照片</label>
              <div class="border-2 border-dashed border-gray-300 rounded-xl p-4 text-center hover:border-blue-400 cursor-pointer text-sm text-gray-500">
                  点击上传保养后照片
              </div>
          </div>
      </div>

      <div v-else class="space-y-4">
           <div class="bg-orange-50 p-3 rounded-lg flex items-center gap-2 text-sm text-orange-700 mb-4">
              <AlertTriangle class="w-4 h-4" />
              提交后将通知设备管理员，并将设备状态更为“维修中”
          </div>
          <div>
              <label class="text-sm font-medium mb-1 block">故障现象描述</label>
              <textarea 
                  v-model="repairForm.issue"
                  class="flex min-h-[100px] w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background disabled:cursor-not-allowed disabled:opacity-50"
                  placeholder="请详细描述故障现象、错误代码..."
              ></textarea>
          </div>
          <div>
               <label class="text-sm font-medium mb-1 block">紧急程度</label>
               <select v-model="repairForm.urgency" class="w-full h-10 rounded-lg border border-input px-3">
                   <option value="normal">一般 - 不影响核心功能</option>
                   <option value="urgent">紧急 - 设备完全无法使用</option>
                   <option value="critical">危急 - 存在安全隐患</option>
               </select>
          </div>
      </div>

      <template #footer>
          <Button variant="ghost" @click="$emit('close')">取消</Button>
          <Button @click="$emit('submit', mode === 'maintenance' ? checklist : repairForm)">提交记录</Button>
      </template>
  </Dialog>
</template>
