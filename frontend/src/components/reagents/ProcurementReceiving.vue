<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import Input from '@/components/ui/Input.vue'
import { CheckCircle, Clock } from 'lucide-vue-next'

const pendingItems = ref<any[]>([])
const loading = ref(false)
const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const receiveInputs = ref<Record<number, number>>({})

const toast = (msg: string, type: 'success' | 'error' = 'success') => {
    toastMessage.value = msg
    toastType.value = type
    showToast.value = true
    setTimeout(() => { showToast.value = false }, 3000)
}

const fetchPendingReceives = async () => {
    loading.value = true
    try {
        const res = await axios.get('/api/reagents/pending-receives')
        pendingItems.value = res.data
        // Initialize default input to the remaining quantity
        pendingItems.value.forEach(item => {
             receiveInputs.value[item.id] = item.quantity - item.received_quantity
        })
    } catch (e) {
        toast("获取待收货明细失败", "error")
    } finally {
        loading.value = false
    }
}

const handleReceive = async (item: any) => {
    const qty = receiveInputs.value[item.id]
    if (!qty || qty <= 0 || qty > (item.quantity - item.received_quantity)) {
        toast("输入的收货数量无效", "error")
        return
    }

    try {
        await axios.post(`/api/reagents/pending-receives/${item.id}/receive`, { quantity: qty })
        toast("成功收货！后台已生成对应物资二维码。")
        fetchPendingReceives()
    } catch (e: any) {
        toast(e.response?.data?.error || "收货提交失败", "error")
    }
}

const getStatusVariant = (status: string) => {
   if (status === '待收货') return 'secondary'
   if (status === '部分收货') return 'warning'
   return 'default'
}

onMounted(() => {
    fetchPendingReceives()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex justify-between items-center">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-gray-900">到货点验与赋码台</h2>
        <p class="text-sm text-gray-500 mt-1">负责对外部 Excel 引入的处于"发货在途"状态的物资进行物理点检验收收库</p>
      </div>
      <Button @click="fetchPendingReceives" variant="outline" size="sm">刷新列表</Button>
    </div>

    <!-- Alert List -->
    <div v-if="loading" class="text-center py-10 text-gray-400">正在加载待收货在途清单...</div>
    <div v-else-if="pendingItems.length === 0" class="text-center py-10 text-gray-400">目前没有待入库/在此暂存区点验的试剂。</div>
    
    <div v-else class="space-y-4">
      <Card v-for="item in pendingItems" :key="item.id" class="p-4 flex items-center justify-between shadow-sm">
         <!-- Info -->
         <div class="flex items-start gap-4 flex-1">
             <div class="h-10 w-10 flex-shrink-0 bg-blue-50 text-blue-600 rounded-full flex items-center justify-center">
                 <Clock class="h-5 w-5" />
             </div>
             <div>
                 <h3 class="font-medium text-gray-900">{{ item.reagent_name }}</h3>
                 <div class="flex gap-3 text-sm text-gray-500 mt-1.5 font-mono">
                     <span>CAS: {{ item.cas_number || '--' }}</span>
                     <span>批次凭证: {{ item.batch?.order_number || '无' }}</span>
                     <span>供应商: {{ item.supplier }}</span>
                 </div>
                 <div class="mt-2">
                     <Badge :variant="getStatusVariant(item.receive_status)">{{ item.receive_status }}</Badge>
                     <span class="text-xs text-gray-500 ml-2">总申量: {{ item.quantity }}{{item.unit}} | 已收: {{ item.received_quantity }}</span>
                 </div>
             </div>
         </div>

         <!-- Action -->
         <div class="flex items-center gap-3">
             <div class="flex flex-col items-end gap-1">
               <label class="text-xs text-gray-500">本次点收数量</label>
               <Input type="number" 
                   v-model="receiveInputs[item.id]" 
                   class="w-24 text-center h-9" 
                   :min="1" 
                   :max="item.quantity - item.received_quantity" 
               />
             </div>
             <Button @click="handleReceive(item)" variant="default" class="mt-5">
                 <CheckCircle class="w-4 h-4 mr-1.5" />
                 确认点收
             </Button>
         </div>
      </Card>
    </div>

    <!-- Toast Notification -->
    <div v-if="showToast" 
         class="fixed bottom-4 right-4 px-6 py-3 rounded-lg shadow-lg text-white font-medium transition-all duration-300 z-50 flex items-center gap-2"
         :class="toastType === 'success' ? 'bg-gray-800' : 'bg-red-600'">
        <CheckCircle v-if="toastType === 'success'" class="w-5 h-5" />
        {{ toastMessage }}
    </div>
  </div>
</template>
