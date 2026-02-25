<script setup lang="ts">
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { ref } from 'vue'
import { Sparkles } from 'lucide-vue-next'

defineProps<{
  open: boolean
}>()

defineEmits<{
  (e: 'close'): void
  (e: 'submit', data: any): void
}>()

const form = ref({
  name: '',
  specs: '',
  budget: '',
  reason: ''
})

const aiRecommendation = ref<any[] | null>(null)

const getAIRecommendation = () => {
    // Mock AI
    setTimeout(() => {
        aiRecommendation.value = [
            { supplier: 'Agilent', price: '¥450,000', delivery: '2 weeks' },
            { supplier: 'Shimadzu', price: '¥420,000', delivery: '3 weeks' }
        ]
    }, 1000)
}
</script>

<template>
  <Dialog :open="open" @close="$emit('close')" title="发起仪器采购申请" class="p-6">
      <div class="space-y-4">
          <div>
              <label class="text-sm font-medium mb-1 block">设备名称</label>
              <div class="flex gap-2">
                  <Input v-model="form.name" placeholder="请输入设备名称，如：离心机" />
                  <Button variant="secondary" size="icon" @click="getAIRecommendation" title="AI推荐">
                      <Sparkles class="w-4 h-4 text-purple-600" />
                  </Button>
              </div>
          </div>
          
          <!-- AI Suggestion Area -->
          <div v-if="aiRecommendation" class="bg-purple-50 p-3 rounded-lg border border-purple-100 animate-in fade-in slide-in-from-top-2">
              <p class="text-xs font-bold text-purple-700 mb-2 flex items-center gap-1"><Sparkles class="w-3 h-3" /> AI 智能推荐供应商</p>
              <div class="space-y-2">
                  <div v-for="(rec, idx) in aiRecommendation" :key="idx" class="flex justify-between text-sm bg-white p-2 rounded border border-purple-100 cursor-pointer hover:border-purple-300 transition-colors">
                      <span class="font-medium">{{ rec.supplier }}</span>
                      <span class="text-gray-500">{{ rec.price }} · {{ rec.delivery }}</span>
                  </div>
              </div>
          </div>

          <div>
              <label class="text-sm font-medium mb-1 block">规格要求</label>
              <textarea 
                  v-model="form.specs"
                  class="flex min-h-[60px] w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  placeholder="请输入详细规格参数..."
              ></textarea>
          </div>
          <div>
              <label class="text-sm font-medium mb-1 block">预算范围</label>
              <Input v-model="form.budget" placeholder="例：50000 - 80000" />
          </div>
          <div>
              <label class="text-sm font-medium mb-1 block">申请理由</label>
              <textarea 
                  v-model="form.reason"
                  class="flex min-h-[60px] w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  placeholder="项目需求..."
              ></textarea>
          </div>
      </div>
      <template #footer>
          <Button variant="ghost" @click="$emit('close')">取消</Button>
          <Button @click="$emit('submit', form)">提交申请</Button>
      </template>
  </Dialog>
</template>
