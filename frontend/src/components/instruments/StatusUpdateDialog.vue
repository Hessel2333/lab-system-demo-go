<script setup lang="ts">
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Select from '@/components/ui/Select.vue'
import Textarea from '@/components/ui/Textarea.vue'
import { ref } from 'vue'
import { updateInstrumentStatus } from '@/api/instruments'

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
             <Select v-model="form.status">
                 <option value="active">空闲 / 可用</option>
                 <option value="in_use">使用中</option>
                 <option value="maintenance">维护中</option>
                 <option value="fault">故障 / 维修</option>
                 <option value="retired">已报废</option>
             </Select>
        </div>
        <div>
            <label class="text-sm font-medium mb-1.5 block text-gray-700">操作员</label>
            <Input v-model="form.operator" placeholder="请输入操作人姓名" />
        </div>
        <div>
            <label class="text-sm font-medium mb-1.5 block text-gray-700">备注原因</label>
            <Textarea
                v-model="form.reason"
                class="min-h-[100px] resize-none"
                placeholder="请输入状态变更原因..."
            />
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
