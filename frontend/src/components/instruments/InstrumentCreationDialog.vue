<script setup lang="ts">
import { ref, reactive } from 'vue'
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { createInstrument } from '@/api/instruments'
import { ChevronDown } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success'): void
}>()

const loading = ref(false)
const form = reactive({
    name: '',
    model: '',
    brand: '',
    location: '',
    status: 'active'
})

const submit = async () => {
    if(!form.name || !form.model) {
        alert('请填写必要信息')
        return
    }
    
    loading.value = true
    try {
        await createInstrument({
            ...form,
            purchase_date: new Date().toISOString(),
            admin: 'Admin User', // Default admin
            run_time: 0,
            health: 100,
            reservations_count: 0,
            lifecycle_stage: 'active'
        })
        emit('success')
        emit('close')
        // Reset form
        form.name = ''
        form.model = ''
        form.brand = ''
        form.location = ''
    } catch(e) {
        console.error(e)
        alert('创建失败')
    } finally {
        loading.value = false
    }
}
</script>

<template>
  <Dialog :open="open" @close="$emit('close')" title="新增仪器入库" maxWidth="max-w-xl">
    <div class="space-y-6 py-6 px-6">
        <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
                <label class="text-sm font-medium text-gray-700">仪器名称 <span class="text-red-500">*</span></label>
                <Input v-model="form.name" placeholder="例如: 高分辨质谱仪" />
            </div>
            <div class="space-y-2">
                <label class="text-sm font-medium text-gray-700">型号 <span class="text-red-500">*</span></label>
                <Input v-model="form.model" placeholder="例如: Q Exactive" />
            </div>
        </div>

        <div class="grid grid-cols-2 gap-6">
            <div class="space-y-2">
                <label class="text-sm font-medium text-gray-700">品牌/厂商</label>
                <Input v-model="form.brand" placeholder="例如: Thermo Fisher" />
            </div>
            <div class="space-y-2">
                <label class="text-sm font-medium text-gray-700">存放位置</label>
                <Input v-model="form.location" placeholder="例如: 实验室 302" />
            </div>
        </div>
        
        <div class="space-y-2">
             <label class="text-sm font-medium text-gray-700">初始状态</label>
             <div class="relative">
                 <select v-model="form.status" class="flex h-10 w-full appearance-none rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50">
                     <option value="active">正常运行 (Active)</option>
                     <option value="arrival">待安装 (Arrival)</option>
                     <option value="maintenance">维护中 (Maintenance)</option>
                 </select>
                 <ChevronDown class="absolute right-3 top-3 h-4 w-4 opacity-50 pointer-events-none" />
             </div>
        </div>
    </div>
    
    <template #footer>
        <div class="flex justify-end gap-3 pt-4">
            <Button variant="ghost" @click="$emit('close')">取消</Button>
            <Button @click="submit" :disabled="loading" class="bg-blue-600 hover:bg-blue-700 text-white shadow-sm shadow-blue-200">
                {{ loading ? '提交中...' : '确认入库' }}
            </Button>
        </div>
    </template>
  </Dialog>
</template>
