<script setup lang="ts">
import { ref } from 'vue'
import { PackageCheck, Trash2, QrCode } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import axios from 'axios'

interface Cabinet { id: number; name: string; cabinet_type: string; location: string }

const scannedUUID = ref('')
const currentItem = ref<any>(null)
const isLoading = ref(false)
const locationInput = ref('')
const cabinetInput = ref<number>(0) // 选中的柜 ID
const cabinets = ref<Cabinet[]>([]) // 从 API 加载

const quickLocations = ['E309', 'E307', 'F103', 'F309']

// Toast
const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const toast = (msg: string, type: 'success' | 'error' = 'success') => {
    toastMessage.value = msg
    toastType.value = type
    showToast.value = true
    setTimeout(() => { showToast.value = false }, 3000)
}

const loadCabinets = async (isControlled: boolean) => {
  try {
    const cabinetType = isControlled ? '易制毒制爆试剂柜' : '普通试剂柜'
    const res = await axios.get(`/api/reagents/cabinets?type=${encodeURIComponent(cabinetType)}`)
    cabinets.value = res.data ?? []
  } catch { cabinets.value = [] }
}

const handleScanMock = () => {
  if (!scannedUUID.value) {
    toast("请输入用于扫描的 UUID。", 'error')
    return
  }
  fetchItem(scannedUUID.value)
}

const getStatusLabel = (status: string) => status

const fetchItem = async (uuid: string) => {
  isLoading.value = true
  currentItem.value = null
  cabinets.value = []
  try {
    const response = await axios.get(`/api/reagents/items/${uuid}`)
    currentItem.value = response.data
    if (currentItem.value.location) locationInput.value = currentItem.value.location
    if (currentItem.value.cabinet_id) cabinetInput.value = currentItem.value.cabinet_id

    // 根据品目管控属性加载对应柜列表
    if (currentItem.value.status === '已到货') {
      await loadCabinets(currentItem.value.reagent_catalog?.is_controlled ?? false)
    }
  } catch (error) {
    console.error(error)
    toast("未找到该试剂或无效的二维码。", 'error')
  } finally {
    isLoading.value = false
  }
}

const updateStatus = async (newStatus: string) => {
  if (!currentItem.value) return
  try {
    const payload: any = { status: newStatus }
    if (newStatus === '在库') {
       if (!locationInput.value) {
         toast("请指定一个入库存放位置。", 'error')
         return
       }
       payload.location = locationInput.value
       if (cabinetInput.value > 0) payload.cabinet_id = cabinetInput.value
    }
    await axios.put(`/api/reagents/items/${currentItem.value.uuid}/status`, payload)
    toast(`该试剂已被标记为 ${getStatusLabel(newStatus)}。`)
    fetchItem(currentItem.value.uuid)
  } catch (error) {
    toast("无法更新试剂状态。", 'error')
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Scanner Simulation Area -->
    <Card class="border-dashed border-2 bg-gray-50/50">
      <div class="p-6 flex flex-col items-center gap-4">
        <div class="h-16 w-16 rounded-full bg-blue-100 flex items-center justify-center animate-pulse">
          <QrCode class="h-8 w-8 text-blue-600" />
        </div>
        <div class="text-center space-y-1">
          <h3 class="font-medium text-lg">手持端扫码录入</h3>
          <p class="text-sm text-gray-500">此框用于外部扫码枪读取试剂二维码输入</p>
        </div>
        <div class="flex gap-2 w-full max-w-sm">
          <Input v-model="scannedUUID" placeholder="例如: 550e8400-e29b..." />
          <Button @click="handleScanMock" :disabled="isLoading" class="whitespace-nowrap min-w-[80px]">
            查询
          </Button>
        </div>
      </div>
    </Card>

    <!-- Scanned Result Card -->
    <Card v-if="currentItem" class="overflow-hidden border-l-4" 
        :class="{
            'border-l-blue-500': currentItem.status === '已到货',
            'border-l-green-500': currentItem.status === '在库',
            'border-l-gray-500': currentItem.status === '已耗尽'
        }">
      <div class="p-6 space-y-6">
          <div>
              <h3 class="text-lg font-semibold">{{ currentItem.reagent_catalog?.name }}</h3>
              <p class="text-sm text-gray-500 font-mono">UUID: {{ currentItem.uuid }}</p>
          </div>

          <div class="grid grid-cols-2 gap-4">
              <div>
                  <span class="text-xs text-gray-400">CAS 编号</span>
                  <p class="font-mono text-sm">{{ currentItem.reagent_catalog?.cas_number || '--' }}</p>
              </div>
              <div>
                  <span class="text-xs text-gray-400">当前状态</span>
                  <p><Badge :variant="currentItem.status === '在库' ? 'default' : 'secondary'">{{ getStatusLabel(currentItem.status) }}</Badge></p>
              </div>
              <div>
                  <span class="text-xs text-gray-400">当前位置</span>
                  <p class="text-sm">{{ currentItem.location || '未分配' }}</p>
              </div>
               <div>
                  <span class="text-xs text-gray-400">管制品</span>
                  <p class="text-sm">{{ currentItem.reagent_catalog?.is_controlled ? '🔒 是' : '否' }}</p>
              </div>
          </div>

          <!-- Actions based on status -->
          <div class="pt-4 border-t">
              <div v-if="currentItem.status === '已到货'" class="space-y-3">
                  <p class="text-sm font-medium text-blue-800 bg-blue-50 p-3 rounded-lg">
                    💡 此试剂已从供应商到货，位于分拣区。请选择存放的试剂柜和实验室位置并确认入库。
                  </p>
                  <!-- 试剂柜选择 -->
                  <div>
                    <p class="text-xs text-gray-500 mb-1.5 font-medium">选择试剂柜
                      <span class="text-blue-500 ml-1">（{{ currentItem.reagent_catalog?.is_controlled ? '⚠️ 易制毒制爆品专柜' : '普通品' }}）</span>
                    </p>
                    <div v-if="cabinets.length === 0" class="text-xs text-gray-400 py-2">正在加载柜点位列表...</div>
                    <div v-else class="flex flex-wrap gap-2">
                      <Button
                        v-for="cab in cabinets"
                        :key="cab.id"
                        size="sm"
                        :variant="cabinetInput === cab.id ? 'default' : 'outline'"
                        @click="() => { cabinetInput = cab.id; locationInput = cab.location; }"
                      >
                        🗄️ {{ cab.name }}
                        <span class="ml-1 text-[10px] opacity-60">{{ cab.location }}</span>
                      </Button>
                    </div>
                  </div>
                  <!-- 位置选择 -->
                  <div>
                    <p class="text-xs text-gray-500 mb-1.5 font-medium flex items-center gap-1.5">
                      选择库位（房间号）
                      <span v-if="cabinetInput > 0" class="text-[10px] bg-blue-50 text-blue-500 px-1.5 py-0.5 rounded border border-blue-100 animate-in fade-in slide-in-from-left-1">已随柜联动同步</span>
                    </p>
                    <div class="flex flex-wrap gap-2">
                      <Button 
                          v-for="loc in quickLocations" 
                          :key="loc"
                          size="sm"
                          :variant="locationInput === loc ? 'default' : 'outline'"
                          @click="locationInput = loc"
                      >
                          📍 {{ loc }}
                      </Button>
                    </div>
                  </div>
                  <Input v-model="locationInput" placeholder="或手动输入库位..." class="mt-1" />
                  <Button class="w-full mt-2" @click="updateStatus('在库')">
                      <PackageCheck class="mr-2 h-4 w-4" />
                      确认入库：{{ cabinets.find(c => c.id === cabinetInput)?.name || '未选柜' }} · {{ locationInput || '...' }}
                  </Button>
              </div>

              <!-- Status: InStorage -> Consume/Dispose -->
              <div v-else-if="currentItem.status === '在库'" class="space-y-3">
                   <p class="text-sm font-medium text-green-800 bg-green-50 p-3 rounded-lg">
                    ✅ 此试剂目前在库，存放于 {{ currentItem.location }}。
                   </p>
                   <Button variant="destructive" class="w-full" @click="updateStatus('已耗尽')">
                      <Trash2 class="mr-2 h-4 w-4" />
                      确认空瓶核销
                   </Button>
              </div>

              <!-- Status: Used -->
              <div v-else-if="currentItem.status === '已耗尽'" class="text-center py-4">
                  <p class="text-sm text-gray-500">此试剂已完成生命周期 (空瓶核销)。</p>
              </div>
          </div>
      </div>
    </Card>

    <!-- Toast Notification -->
    <Transition
      enter-active-class="transition ease-out duration-300"
      enter-from-class="translate-y-4 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-4 opacity-0"
    >
      <div v-if="showToast" class="fixed bottom-6 right-6 z-50 max-w-sm">
        <div :class="[
          'px-4 py-3 rounded-lg shadow-lg border text-sm font-medium flex items-center gap-2',
          toastType === 'success' ? 'bg-green-50 text-green-800 border-green-200' : 'bg-red-50 text-red-800 border-red-200'
        ]">
          <span>{{ toastMessage }}</span>
        </div>
      </div>
    </Transition>
  </div>
</template>
