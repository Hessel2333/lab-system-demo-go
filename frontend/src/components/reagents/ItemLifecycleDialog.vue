<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import axios from 'axios'
import { toast } from 'vue-sonner'
import {
  AlertTriangle,
  MapPin,
  CalendarClock,
  QrCode,
  TestTube2,
  History
} from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import Dialog from '@/components/ui/Dialog.vue'
import { formatNumber, normalizeUnit } from '@/lib/quantity'
import { getInventoryDisplayStatus, getInventoryStatusVariant, isArrivedStatus, isInStorageStatus, isUsedStatus } from '@/lib/reagent-status'

const props = defineProps<{
  isOpen: boolean
  itemUuid: string | null
  userRole?: 'researcher' | 'procurement' | 'leader'
}>()

const emit = defineEmits(['close', 'refresh-needed'])

const loading = ref(false)
const itemData = ref<any>(null)
const actionProcessing = ref(false)
const consumeVolume = ref(1)
const consumeRemarks = ref('')
const consumeUnit = computed(() => normalizeUnit(itemData.value?.reagent_catalog?.unit, 'ml'))

const isArrivalStageLog = (log: any) => {
  const action = String(log?.action || '')
  const remarks = String(log?.remarks || '')
  const arrivalKeywords = ['点收', '点验', '到货确认', '收货确认', '批次确认', '二维码打印', '赋码', '已到货', '暂存区']
  return arrivalKeywords.some((k) => action.includes(k) || remarks.includes(k))
}

const displayLogs = computed(() => {
  const logs = Array.isArray(itemData.value?.logs) ? itemData.value.logs : []
  // 全周期流转单不展示采购到货确认阶段痕迹，只保留入库后的使用周期
  return logs.filter((log: any) => !isArrivalStageLog(log))
})

watch(() => props.isOpen, (newVal) => {
  if (newVal && props.itemUuid) {
    fetchItemDetails()
  } else {
    itemData.value = null
  }
})

const fetchItemDetails = async () => {
  loading.value = true
  try {
    const res = await axios.get(`/api/reagents/items/${props.itemUuid}`)
    itemData.value = res.data
    const remaining = Number(itemData.value?.remaining_volume || 0)
    consumeVolume.value = remaining > 0 ? Math.min(1, remaining) : 0
    consumeRemarks.value = ''
  } catch (e: any) {
    console.error(e)
    toast.error('无法提取物资档案')
    emit('close')
  } finally {
    loading.value = false
  }
}

const getStatusVariant = (status: string): any => getInventoryStatusVariant(status)

const getLogDotColor = (action: string) => {
  const map: Record<string, string> = {
    '扫码入库': 'bg-emerald-500',
    '入库登记': 'bg-emerald-500',
    '确认领回': 'bg-blue-500',
    '扫码领用': 'bg-indigo-500',
    '领用消耗': 'bg-indigo-500',
    '空瓶核销': 'bg-orange-500',
  }
  return map[action] || 'bg-gray-400'
}

const handleAction = async (actionType: 'consume' | 'dispose') => {
  if (!itemData.value) return
  actionProcessing.value = true

  try {
    if (actionType === 'consume') {
      if (!itemData.value.reagent_catalog?.is_controlled) {
        toast.error('普通试剂不支持逐次消耗，请直接执行“用尽”')
        return
      }
      const remaining = Number(itemData.value.remaining_volume || 0)
      if (consumeVolume.value <= 0) {
        toast.error('消耗量必须大于 0')
        return
      }
      if (consumeVolume.value > remaining) {
        toast.error('消耗量不能超过当前余量')
        return
      }
      await axios.put(`/api/reagents/items/${itemData.value.uuid}/consume`, {
        consume_volume: Number(consumeVolume.value),
        remarks: consumeRemarks.value || `流转单登记消耗 ${consumeVolume.value}${consumeUnit.value}`,
      })
      toast.success('已登记试剂消耗')
    } else {
      if (!confirm('确认该试剂已经彻底使用完毕并需空瓶核销吗？')) return
      await axios.post(`/api/reagents/items/${itemData.value.uuid}/deplete`, {
        remarks: '流转单执行耗尽核销',
      })
      toast.success('已空瓶核销该试剂')
    }

    await fetchItemDetails()
    emit('refresh-needed')
  } catch (e: any) {
    toast.error('操作失败: ' + (e.response?.data?.error || '服务器错误'))
  } finally {
    actionProcessing.value = false
  }
}

</script>

<template>
  <Dialog :open="isOpen" size="xl" @close="emit('close')">
    <template #header>
      <div class="flex flex-col">
        <span class="text-xl font-bold tracking-tight text-slate-900 flex items-center gap-2">
          <History class="h-5 w-5 text-blue-600" />
          试剂个体全生命周期流转单
        </span>
        <span class="text-xs font-mono text-slate-400 mt-0.5 uppercase tracking-wider">UUID: {{ itemUuid }}</span>
      </div>
    </template>

    <div v-if="loading" class="p-12">
      <div class="flex h-48 items-center justify-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    </div>

    <div v-else-if="itemData" class="bg-slate-50/80 px-6 py-4 min-h-[550px]">
      <div class="grid gap-4 lg:grid-cols-12">
        <div class="lg:col-span-8 flex flex-col gap-4">
          <div class="apple-card apple-card-hover p-5 border-blue-50/50">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center gap-3">
                <Badge :variant="getStatusVariant(itemData.status)" class="px-3 h-6 font-bold uppercase tracking-widest">
                  {{ getInventoryDisplayStatus(itemData.status) }}
                </Badge>
                <Badge v-if="itemData.reagent_catalog?.is_controlled" variant="destructive" class="px-2 h-5 text-[10px] font-bold">
                  管控品
                </Badge>
              </div>
              <span class="text-[11px] font-bold text-slate-400 bg-slate-100/80 px-1.5 py-0.5 rounded">
                #{{ itemData.uuid?.substring(0, 8).toUpperCase() }}
              </span>
            </div>

            <h2 class="text-2xl font-bold tracking-tight text-slate-900">{{ itemData.reagent_catalog?.name }}</h2>
            <p class="text-sm font-mono text-slate-500 mt-1">CAS 号: {{ itemData.reagent_catalog?.cas_number || '未知' }}</p>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 mt-5">
              <div class="rounded-lg border border-slate-200 bg-slate-50/60 px-3 py-2">
                <div class="text-[10px] font-bold uppercase tracking-widest text-slate-400">剩余余量</div>
                <div class="mt-1 flex items-end gap-1.5">
                  <span class="text-2xl font-bold font-mono tracking-tight" :class="itemData.remaining_volume > 0 ? 'text-blue-600' : 'text-red-500'">
                    {{ formatNumber(itemData.remaining_volume) }}
                  </span>
                  <span class="text-xs text-slate-400 mb-1">/ {{ formatNumber(itemData.capacity) }} {{ normalizeUnit(itemData.reagent_catalog?.unit, 'ml') }}</span>
                </div>
              </div>
              <div class="rounded-lg border border-slate-200 bg-slate-50/60 px-3 py-2">
                <div class="text-[10px] font-bold uppercase tracking-widest text-slate-400">批次来源</div>
                <div class="mt-1 text-sm font-mono font-bold text-slate-700">{{ itemData.batch_number || '-' }}</div>
              </div>
              <div class="rounded-lg border border-slate-200 bg-slate-50/60 px-3 py-2">
                <div class="text-[10px] font-bold uppercase tracking-widest text-slate-400">当前库位</div>
                <div class="mt-1 text-sm font-bold text-slate-700">{{ itemData.location || '--' }}</div>
              </div>
            </div>
          </div>

          <div class="apple-card apple-card-hover p-5 flex-grow shadow-md border-slate-200/40">
            <div class="mb-4 flex items-center justify-between border-b border-slate-50 pb-3">
              <h3 class="flex items-center gap-2 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                <History class="h-3.5 w-3.5" />
                全生命周期操作流转
              </h3>
            </div>

            <div v-if="displayLogs.length === 0" class="rounded-lg border border-dashed border-slate-200 bg-white px-3 py-3 text-[11px] text-slate-400">
              暂无可展示流转历史
            </div>
            <ol v-else class="relative space-y-4">
              <li v-for="(log, idx) in displayLogs" :key="log.id" class="group relative flex gap-6 pl-0">
                <div
                  v-if="Number(idx) < displayLogs.length - 1"
                  class="absolute left-4 top-[34px] bottom-0 w-[1.5px] bg-slate-100"
                ></div>
                <div class="relative z-10 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-slate-200 bg-white mt-[1.5px]">
                  <div class="w-1.5 h-1.5 rounded-full" :class="getLogDotColor(log.action)"></div>
                </div>
                <div class="flex-grow pb-4 pt-1 px-5 rounded-xl border border-slate-100/50 bg-white/30 hover:bg-white hover:shadow-sm transition-all duration-300 group-hover:translate-x-1">
                  <div class="flex items-center justify-between gap-4">
                    <div class="flex items-center gap-3 py-0.5">
                      <span class="text-sm font-bold text-slate-800">{{ log.action || '系统操作' }}</span>
                    </div>
                    <div class="flex items-center gap-4 text-xs whitespace-nowrap">
                      <div class="flex items-center gap-2 font-bold text-slate-700">
                        <span class="text-[10px] text-slate-400 uppercase tracking-tighter">操作人:</span>
                        <span>{{ log.user?.real_name || 'System' }}</span>
                      </div>
                      <span class="font-bold text-slate-500 opacity-90">{{ log.created_at ? new Date(log.created_at).toLocaleString('zh-CN', { hour12: false }) : '--' }}</span>
                    </div>
                  </div>
                  <div class="mt-1 text-[13px] leading-relaxed text-slate-500">
                    <span v-if="log.quantity">数量变动：{{ formatNumber(log.quantity) }}</span>
                    <span v-else>流程节点记录</span>
                  </div>
                  <div v-if="log.remarks" class="mt-2 rounded-lg border border-dashed border-slate-200 bg-white/80 px-3 py-2 text-[11px] text-slate-500 italic">
                    {{ log.remarks }}
                  </div>
                </div>
              </li>
            </ol>
          </div>

        </div>

        <div class="lg:col-span-4 flex flex-col gap-4 text-nowrap">
          <div class="apple-card p-5 bg-slate-50/40 border-slate-200/30 shadow-none">
            <h4 class="mb-3 text-[10px] font-bold uppercase tracking-widest text-slate-400">身份标识</h4>
            <div class="rounded-xl border border-slate-200 bg-white p-4 flex flex-col items-center">
              <QrCode class="w-12 h-12 text-slate-800" stroke-width="1.5" />
              <span class="text-[10px] text-slate-400 font-mono mt-3 tracking-widest font-bold">{{ itemUuid?.substring(0,8).toUpperCase() }}</span>
              <span class="text-[9px] text-slate-400 mt-1">扫码防伪溯源</span>
            </div>
          </div>

          <div class="apple-card p-5 bg-slate-50/40 border-slate-200/30 shadow-none">
            <h4 class="mb-3 text-[10px] font-bold uppercase tracking-widest text-slate-400">库位与有效期</h4>
            <div class="space-y-3 text-sm">
              <div class="flex items-center justify-between"><span class="text-slate-500">实物位置</span><span class="font-bold text-slate-800">{{ itemData.location || '--' }}</span></div>
              <div class="flex items-center justify-between"><span class="text-slate-500">归属柜位</span><span class="font-bold text-slate-800">{{ itemData.cabinet?.name || '公用区' }}</span></div>
              <div class="flex items-center justify-between"><span class="text-slate-500">存储要求</span><span class="font-bold text-slate-800 text-xs">{{ itemData.reagent_catalog?.storage_condition || '常温避光' }}</span></div>
              <div class="flex items-center justify-between"><span class="text-slate-500">失效日期</span><span :class="itemData.expiry_date && new Date(itemData.expiry_date) < new Date() ? 'text-red-600 font-bold' : 'font-bold text-slate-800'">{{ itemData.expiry_date ? new Date(itemData.expiry_date).toLocaleDateString() : '长期有效' }}</span></div>
            </div>
          </div>

          <div class="apple-card p-5 bg-slate-50/60 border-slate-200/20 shadow-none">
            <h4 class="mb-2 text-[10px] font-bold uppercase tracking-widest text-slate-400">状态摘要</h4>
            <div class="space-y-2 text-[11px] text-slate-500">
              <p class="flex items-center gap-2"><MapPin class="h-3.5 w-3.5 text-cyan-500" /> 当前位置已记录，可扫码复核。</p>
              <p class="flex items-center gap-2"><CalendarClock class="h-3.5 w-3.5 text-amber-500" /> 有效期与存储条件已关联品目档案。</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div v-if="itemData" class="flex flex-wrap gap-3 w-full items-center">
        <Button variant="secondary" @click="emit('close')">关闭详情</Button>

        <div v-if="isUsedStatus(itemData.status)" class="flex-1 flex items-center justify-center text-xs font-bold text-slate-400 uppercase tracking-widest bg-slate-50 rounded-xl">
          已完成全生命周期
        </div>
        <div v-else-if="isArrivedStatus(itemData.status)" class="flex-1 flex items-center justify-center text-xs font-medium text-slate-500 bg-slate-50 rounded-xl px-4 py-2">
          该试剂处于待入库阶段，请在「到货台账」完成扫码入库
        </div>
        <template v-else-if="isInStorageStatus(itemData.status)">
          <div v-if="itemData.reagent_catalog?.is_controlled" class="flex-1 min-w-[260px] rounded-xl border border-slate-200 bg-slate-50 p-2.5">
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2 items-center">
              <div class="sm:col-span-1">
                <label class="block text-[10px] text-slate-500 mb-1">本次消耗量</label>
                <div class="relative">
                  <input
                    v-model.number="consumeVolume"
                    type="number"
                    min="0"
                    :max="itemData.remaining_volume"
                    step="0.1"
                    class="w-full h-9 rounded-lg border border-slate-200 bg-white px-2 pr-9 text-sm"
                  />
                  <span class="absolute right-2 top-2 text-xs text-slate-400">{{ consumeUnit }}</span>
                </div>
              </div>
              <div class="sm:col-span-2">
                <label class="block text-[10px] text-slate-500 mb-1">用途/备注（选填）</label>
                <input
                  v-model="consumeRemarks"
                  type="text"
                  class="w-full h-9 rounded-lg border border-slate-200 bg-white px-2 text-sm"
                  placeholder="如：滴定实验A组"
                />
              </div>
            </div>
          </div>
          <div v-else class="flex-1 flex items-center justify-center text-xs font-medium text-slate-500 bg-slate-50 rounded-xl px-4 py-2">
            普通试剂无需逐次消耗登记，可直接执行“用尽”
          </div>
          <Button class="flex-1" variant="outline" @click="handleAction('dispose')" :disabled="actionProcessing">
            <AlertTriangle class="w-4 h-4 mr-2 text-orange-500" /> 用尽
          </Button>
          <Button
            v-if="itemData.reagent_catalog?.is_controlled"
            class="flex-1 shadow-blue-100 shadow-lg"
            variant="primary"
            @click="handleAction('consume')"
            :disabled="actionProcessing || consumeVolume <= 0"
          >
            <TestTube2 class="w-4 h-4 mr-2" /> 使用
          </Button>
        </template>
        <div v-else class="flex-1 flex items-center justify-center text-xs font-medium text-slate-500 bg-slate-50 rounded-xl px-4 py-2">
          该页面仅展示流转记录，不提供执行操作
        </div>
      </div>
    </template>
  </Dialog>
</template>
