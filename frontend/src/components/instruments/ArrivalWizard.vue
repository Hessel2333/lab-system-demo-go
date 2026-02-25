<script setup lang="ts">
import Dialog from '@/components/ui/Dialog.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { ref } from 'vue'
import { CheckCircle, Upload, Box } from 'lucide-vue-next'

defineProps<{
  open: boolean
}>()

defineEmits<{
  (e: 'close'): void
  (e: 'complete', data: any): void
}>()

const step = ref(1)
const form = ref({
    packaging: 'intact',
    photo: null,
    checklist: {
        host: false,
        powerCord: false,
        manual: false,
        warranty: false
    },
    assetNo: 'ASSET-' + Math.floor(Math.random() * 10000),
    sn: '',
    location: ''
})

const nextStep = () => {
    if (step.value < 3) step.value++
}

const prevStep = () => {
    if (step.value > 1) step.value--
}


const steps = [
    { num: 1, title: '外观检查' },
    { num: 2, title: '配件核对' },
    { num: 3, title: '资产登记' }
]
</script>

<template>
  <Dialog :open="open" @close="$emit('close')" title="到货验收向导" class="max-w-xl p-6">
      <!-- Stepper -->
      <div class="mb-8 flex items-center justify-between px-4">
          <div v-for="s in steps" :key="s.num" class="flex flex-col items-center relative z-10">
              <div class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold border-2 transition-colors"
                   :class="step >= s.num ? 'bg-blue-600 border-blue-600 text-white' : 'bg-white border-gray-300 text-gray-400'">
                  <CheckCircle v-if="step > s.num" class="w-5 h-5" />
                  <span v-else>{{ s.num }}</span>
              </div>
              <span class="text-xs mt-2 font-medium" :class="step >= s.num ? 'text-blue-600' : 'text-gray-400'">{{ s.title }}</span>
          </div>
          <!-- Progress Line (Visual hack) -->
          <div class="absolute top-16 left-0 w-full h-0.5 bg-gray-200 -z-0 hidden md:block" style="top: 85px; width: 80%; left: 10%;"></div> 
      </div>

      <div class="py-4">
          <!-- Step 1 -->
          <div v-if="step === 1" class="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
              <h3 class="font-semibold text-lg flex items-center gap-2"><Box class="w-5 h-5 text-blue-600" /> 外观包装检查</h3>
              <div class="grid grid-cols-2 gap-4">
                  <div class="border rounded-xl p-4 cursor-pointer hover:border-blue-500 hover:bg-blue-50 transition-all text-center"
                       :class="form.packaging === 'intact' ? 'border-blue-500 bg-blue-50 ring-1 ring-blue-500' : ''"
                       @click="form.packaging = 'intact'">
                      <CheckCircle class="w-8 h-8 mx-auto mb-2 text-green-500" />
                      <div class="font-medium">完好无损</div>
                  </div>
                  <div class="border rounded-xl p-4 cursor-pointer hover:border-orange-500 hover:bg-orange-50 transition-all text-center"
                       :class="form.packaging === 'damaged' ? 'border-orange-500 bg-orange-50 ring-1 ring-orange-500' : ''"
                       @click="form.packaging = 'damaged'">
                      <Upload class="w-8 h-8 mx-auto mb-2 text-orange-500" /> <!-- Warning icon replacement -->
                      <div class="font-medium">存在破损</div>
                  </div>
              </div>
              <div>
                  <label class="block text-sm font-medium mb-2">上传包装照片</label>
                  <div class="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center hover:border-blue-400 cursor-pointer">
                      <Upload class="w-6 h-6 mx-auto text-gray-400 mb-2" />
                      <p class="text-sm text-gray-500">点击上传或拖拽文件</p>
                  </div>
              </div>
          </div>

          <!-- Step 2 -->
          <div v-if="step === 2" class="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
               <h3 class="font-semibold text-lg">配件清单核对</h3>
               <div class="space-y-2 border rounded-xl p-4">
                   <label class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg cursor-pointer">
                       <input type="checkbox" v-model="form.checklist.host" class="w-5 h-5 rounded border-gray-300 text-blue-600" />
                       <span class="font-medium">主机设备</span>
                   </label>
                   <label class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg cursor-pointer">
                       <input type="checkbox" v-model="form.checklist.powerCord" class="w-5 h-5 rounded border-gray-300 text-blue-600" />
                       <span class="font-medium">电源线/数据线</span>
                   </label>
                   <label class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg cursor-pointer">
                       <input type="checkbox" v-model="form.checklist.manual" class="w-5 h-5 rounded border-gray-300 text-blue-600" />
                       <span class="font-medium">说明书/保修卡</span>
                   </label>
               </div>
          </div>

          <!-- Step 3 -->
          <div v-if="step === 3" class="space-y-4 animate-in fade-in slide-in-from-right-4 duration-300">
              <h3 class="font-semibold text-lg">资产登记</h3>
              <div class="grid gap-4">
                  <div>
                      <label class="text-sm font-medium mb-1 block">资产编号 (自动生成)</label>
                      <Input v-model="form.assetNo" disabled class="bg-gray-100" />
                  </div>
                  <div>
                      <label class="text-sm font-medium mb-1 block">序列号 (S/N)</label>
                      <Input v-model="form.sn" placeholder="扫描或输入设备SN码" />
                  </div>
                  <div>
                      <label class="text-sm font-medium mb-1 block">存放位置</label>
                      <Input v-model="form.location" placeholder="例：分析室 302" />
                  </div>
              </div>
          </div>
      </div>
      
      <template #footer>
          <div class="flex justify-between w-full">
              <Button v-if="step > 1" variant="ghost" @click="prevStep">上一步</Button>
              <div v-else></div>
              
              <Button v-if="step < 3" @click="nextStep">下一步</Button>
              <Button v-else @click="$emit('complete', form)">完成验收</Button>
          </div>
      </template>
  </Dialog>
</template>
