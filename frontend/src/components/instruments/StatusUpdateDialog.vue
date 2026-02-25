<script setup lang="ts">
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { ref } from 'vue'
import { updateInstrumentStatus } from '@/api/instruments'
import { ChevronDown } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  instrumentId?: number
  currentStatus: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'update'): void
}>()

const form = ref({
  status: '',
  operator: '',
  reason: ''
})

const submit = async () => {
    if (!props.instrumentId) return
    try {
        await updateInstrumentStatus(props.instrumentId, form.value.status)
        emit('update')
    } catch (e) {
        console.error(e)
    }
}
</script>

<template>
  <Dialog :open="open" @close="$emit('close')" title="更新设备状态" maxWidth="max-w-md">
    <div class="space-y-5 py-6 px-6">
        <div>
            <label class="text-sm font-medium mb-1.5 block text-gray-700">当前状态</label>
            <div class="px-3 py-2.5 bg-gray-50 border border-gray-200 rounded-lg text-sm text-gray-600 font-medium">{{ currentStatus }}</div>
        </div>
        <div>
             <label class="text-sm font-medium mb-1.5 block text-gray-700">新状态</label>
             <div class="relative">
                 <select v-model="form.status" class="flex h-10 w-full appearance-none rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                     <option value="active">空闲 / 可用</option>
                     <option value="in_use">使用中</option>
                     <option value="maintenance">维护中</option>
                     <option value="fault">故障 / 维修</option>
                     <option value="retired">已报废</option>
                 </select>
                 <ChevronDown class="absolute right-3 top-3 h-4 w-4 opacity-50 pointer-events-none" />
             </div>
        </div>
        <div>
            <label class="text-sm font-medium mb-1.5 block text-gray-700">操作员</label>
            <Input v-model="form.operator" placeholder="请输入操作人姓名" />
        </div>
        <div>
            <label class="text-sm font-medium mb-1.5 block text-gray-700">备注原因</label>
            <textarea 
                v-model="form.reason"
                class="flex min-h-[100px] w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 resize-none"
                placeholder="请输入状态变更原因..."
            ></textarea>
        </div>
    </div>
    <template #footer>
        <div class="flex justify-end gap-3 pt-2">
            <Button variant="ghost" @click="$emit('close')">取消</Button>
            <Button @click="submit">确认更新</Button>
        </div>
    </template>
  </Dialog>
</template>
