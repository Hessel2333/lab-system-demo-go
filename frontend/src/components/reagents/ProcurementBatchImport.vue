<script setup lang="ts">
import { ref, computed } from 'vue'
import { Upload, CheckCircle2, AlertCircle, Loader2 } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Card from '@/components/ui/Card.vue'
import axios from 'axios'
import * as XLSX from 'xlsx'

// --- 状态管理 ---
const step = ref<'upload' | 'match' | 'done'>('upload')
const period = ref('')
const isUploading = ref(false)
const isConfirming = ref(false)

const batchId = ref<number | null>(null)
const batchItems = ref<any[]>([])
const catalogs = ref<any[]>([])


const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const toast = (msg: string, type: 'success' | 'error' = 'success') => {
    toastMessage.value = msg
    toastType.value = type
    showToast.value = true
    setTimeout(() => { showToast.value = false }, 3000)
}

// 加载品目字典（用于手动匹配下拉）
const fetchCatalogs = async () => {
    try {
        const res = await axios.get('/api/reagents/catalogs')
        catalogs.value = res.data
    } catch { /* ignore */ }
}

// 统计
const matchStats = computed(() => {
    const total = batchItems.value.length
    const matched = batchItems.value.filter(i => i.match_status === '自动匹配' || i.match_status === '手动匹配').length
    const unmatched = total - matched
    return { total, matched, unmatched }
})

// --- Step 1: 文件上传与解析 ---
const fileInput = ref<HTMLInputElement | null>(null)
const dragActive = ref(false)

const onDrop = (e: DragEvent) => {
    dragActive.value = false
    const file = e.dataTransfer?.files[0]
    if (file) processFile(file)
}

const onFileSelect = (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (file) processFile(file)
}

const processFile = async (file: File) => {
    if (!period.value) {
        toast('请先选择所属周期', 'error')
        return
    }

    isUploading.value = true
    try {
        let items: any[] = []

        if (file.size > 0) {
            //真实解析 Excel
            const buffer = await file.arrayBuffer()
            const workbook = XLSX.read(buffer, { type: 'array' })
            const firstSheetName = workbook.SheetNames[0]
            if (!firstSheetName) {
                throw new Error("Excel 文件中没有任何工作表")
            }
            const worksheet = workbook.Sheets[firstSheetName]
            if (!worksheet) {
                throw new Error("无法读取第一个工作表")
            }
            // header: 1 表示返回二维数组
            const rawData = XLSX.utils.sheet_to_json<any[]>(worksheet, { header: 1 })
            
            // 易派客通常第二行(index 1)或第一行是实际表头，这里我们从第3行(index 2)开始往下扫
            // "商品名称" 在第24列(index 23), "采购数量" 在第28列(index 27), "基本计量单位" 在第29列(index 28)
            for (let i = 2; i < rawData.length; i++) {
                const row = rawData[i]
                if (!row || row.length < 28) continue
                const name = row[23]
                if (!name || name === '商品名称') continue // 跳过空行或表头本身
                
                items.push({
                    reagent_name: name.toString().trim(),
                    cas_number: '', // 易派客导出通常没有分离的 CAS 号
                    quantity: parseFloat(row[27]) || 1,
                    unit: (row[28] || '瓶').toString().trim()
                })
            }

            if (items.length === 0) {
                toast('未能在 Excel 中解析到有效试剂明细数据', 'error')
                isUploading.value = false
                return
            }
        }

        // 发送真实解析到的 items 到后端创建批次
        const res = await axios.post('/api/reagents/procurement-batches', {
            period: period.value,
            items: items
        })

        batchId.value = res.data.id
        batchItems.value = res.data.items || []

        await fetchCatalogs()
        step.value = 'match'
        toast(`成功解析 ${items.length} 条数据，请核对`)

    } catch (error) {
        toast('文件处理失败，请重试', 'error')
    } finally {
        isUploading.value = false
    }
}

// 手动添加明细行（MVP 补充方式）
const newItemName = ref('')
const newItemCas = ref('')
const newItemQty = ref(1)
const newItemUnit = ref('瓶')

const addManualItem = async () => {
    if (!newItemName.value || !batchId.value) return
    try {
        // 直接通过后端 API 追加明细项
        await axios.post('/api/reagents/procurement-batches', {
            period: period.value,
            items: [{
                reagent_name: newItemName.value,
                cas_number: newItemCas.value,
                quantity: newItemQty.value,
                unit: newItemUnit.value,
            }]
        })
        // 刷新批次明细
        await refreshBatchItems()
        newItemName.value = ''
        newItemCas.value = ''
        newItemQty.value = 1
        toast('已添加明细行')
    } catch {
        toast('添加失败', 'error')
    }
}

const refreshBatchItems = async () => {
    if (!batchId.value) return
    try {
        const res = await axios.get(`/api/reagents/procurement-batches/${batchId.value}/items`)
        batchItems.value = res.data
    } catch { /* ignore */ }
}

// --- Step 2: 手动匹配 ---
const updateItemMatch = async (item: any, catalogId: number) => {
    try {
        await axios.put(`/api/reagents/procurement-batches/${batchId.value}/items/${item.id}`, {
            matched_catalog_id: catalogId,
            cas_number: item.cas_number
        })
        item.matched_catalog_id = catalogId
        item.match_status = '手动匹配'
        toast('匹配成功')
    } catch {
        toast('匹配失败', 'error')
    }
}

// --- Step 3: 确认并赋码 ---
const confirmBatch = async () => {
    if (!batchId.value) return
    isConfirming.value = true
    try {
        const res = await axios.post(`/api/reagents/procurement-batches/${batchId.value}/confirm`)
        toast(`到货确认完成！共生成 ${res.data.items_created} 件库存条目`)
        step.value = 'done'
    } catch {
        toast('确认失败，请重试', 'error')
    } finally {
        isConfirming.value = false
    }
}

// 重置（开始新批次）
const resetAll = () => {
    step.value = 'upload'
    period.value = ''
    batchId.value = null
    batchItems.value = []
}

// 当前年月
const currentPeriod = new Date().toISOString().slice(0, 7) // e.g. "2026-02"

const createEmptyBatch = () => {
    processFile(new File([], 'manual'))
}
</script>

<template>
  <Card>
    <div class="p-6 space-y-5">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-lg font-semibold text-gray-900">采购明细导入</h2>
          <p class="text-xs text-gray-500 mt-0.5">上传易派客等外部平台的采购明细，自动匹配申购需求并触发到货赋码</p>
        </div>
        <div class="flex items-center gap-2">
          <span
            :class="[
              'inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded-full font-medium border',
              step === 'upload' ? 'bg-blue-100 text-blue-700 border-blue-200' :
              step === 'match' ? 'bg-amber-100 text-amber-700 border-amber-200' :
              'bg-green-100 text-green-700 border-green-200'
            ]"
          >
            {{ step === 'upload' ? '📤 待上传' : step === 'match' ? '🔗 待确认' : '✅ 已完成' }}
          </span>
        </div>
      </div>

      <!-- Step 1: 上传 -->
      <div v-if="step === 'upload'" class="space-y-4">
        <!-- 周期选择 -->
        <div class="flex items-center gap-3">
          <label class="text-sm font-medium text-gray-700">所属周期</label>
          <input
            v-model="period"
            type="month"
            :max="currentPeriod"
            class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>

        <!-- 拖拽上传区 -->
        <div
          @dragover.prevent="dragActive = true"
          @dragleave="dragActive = false"
          @drop.prevent="onDrop"
          :class="[
            'border-2 border-dashed rounded-xl p-10 text-center transition-all cursor-pointer',
            dragActive ? 'border-blue-400 bg-blue-50' : 'border-gray-300 hover:border-gray-400 hover:bg-gray-50'
          ]"
          @click="fileInput?.click()"
        >
          <Upload v-if="!isUploading" class="w-10 h-10 mx-auto text-gray-400 mb-3" />
          <Loader2 v-else class="w-10 h-10 mx-auto text-blue-500 animate-spin mb-3" />
          <p class="text-sm font-medium text-gray-700">
            {{ isUploading ? '正在解析...' : '拖拽 Excel 文件至此处，或点击选择' }}
          </p>
          <p class="text-xs text-gray-400 mt-1">支持 .xls / .xlsx 格式（易派客导出明细）</p>
          <input ref="fileInput" type="file" accept=".xls,.xlsx" class="hidden" @change="onFileSelect" />
        </div>

        <!-- 或者手动创建批次 -->
        <div class="text-center">
          <button
            v-if="period"
            @click="createEmptyBatch"
            class="text-xs text-blue-600 hover:text-blue-700 underline"
          >
            没有 Excel？手动创建空批次并逐条录入
          </button>
        </div>
      </div>

      <!-- Step 2: 明细确认与匹配 -->
      <div v-if="step === 'match'" class="space-y-4">
        <!-- 统计条 -->
        <div class="flex items-center gap-4 bg-gray-50 rounded-lg px-4 py-2.5">
          <span class="text-xs text-gray-600">
            批次 #{{ batchId }} · 周期 {{ period }}
          </span>
          <span class="text-xs font-medium text-emerald-600">
            ✅ 已匹配 {{ matchStats.matched }}
          </span>
          <span v-if="matchStats.unmatched > 0" class="text-xs font-medium text-red-600">
            ⚠️ 未匹配 {{ matchStats.unmatched }}
          </span>
          <span class="text-xs text-gray-500">共 {{ matchStats.total }} 项</span>
        </div>

        <!-- 手动添加行 -->
        <div class="flex items-end gap-2 bg-blue-50 rounded-lg px-4 py-3 border border-blue-100">
          <div class="flex-1">
            <label class="block text-[10px] text-gray-500 mb-0.5">试剂名称</label>
            <input v-model="newItemName" type="text" placeholder="如：丙酮" class="w-full px-2 py-1 text-xs border rounded" />
          </div>
          <div class="w-28">
            <label class="block text-[10px] text-gray-500 mb-0.5">CAS号</label>
            <input v-model="newItemCas" type="text" placeholder="67-64-1" class="w-full px-2 py-1 text-xs border rounded" />
          </div>
          <div class="w-16">
            <label class="block text-[10px] text-gray-500 mb-0.5">数量</label>
            <input v-model.number="newItemQty" type="number" min="1" class="w-full px-2 py-1 text-xs border rounded" />
          </div>
          <Button size="sm" class="bg-blue-600 hover:bg-blue-700 text-white text-xs shrink-0" @click="addManualItem">
            + 追加
          </Button>
        </div>

        <!-- 明细表格 -->
        <div class="overflow-x-auto rounded-lg border max-h-96 overflow-y-auto">
          <table class="w-full text-xs">
            <thead class="bg-gray-50 sticky top-0">
              <tr>
                <th class="px-3 py-2 text-left text-gray-600 font-medium">商品名称</th>
                <th class="px-3 py-2 text-left text-gray-600 font-medium">CAS 号</th>
                <th class="px-3 py-2 text-center text-gray-600 font-medium">数量</th>
                <th class="px-3 py-2 text-center text-gray-600 font-medium">匹配状态</th>
                <th class="px-3 py-2 text-left text-gray-600 font-medium">匹配品目</th>
              </tr>
            </thead>
            <tbody class="divide-y">
              <tr v-for="item in batchItems" :key="item.id" :class="item.match_status === '未匹配' ? 'bg-red-50/50' : ''">
                <td class="px-3 py-2 text-gray-900">{{ item.reagent_name }}</td>
                <td class="px-3 py-2 font-mono text-gray-500">{{ item.cas_number || '-' }}</td>
                <td class="px-3 py-2 text-center">{{ item.quantity }} {{ item.unit }}</td>
                <td class="px-3 py-2 text-center">
                  <span :class="[
                    'inline-flex items-center gap-0.5 text-[10px] px-2 py-0.5 rounded-full font-medium',
                    item.match_status === '自动匹配' ? 'bg-green-100 text-green-700' :
                    item.match_status === '手动匹配' ? 'bg-blue-100 text-blue-700' :
                    'bg-red-100 text-red-700'
                  ]">
                    <CheckCircle2 v-if="item.match_status !== '未匹配'" class="w-3 h-3" />
                    <AlertCircle v-else class="w-3 h-3" />
                    {{ item.match_status }}
                  </span>
                </td>
                <td class="px-3 py-2">
                  <select
                    v-if="item.match_status === '未匹配'"
                    @change="updateItemMatch(item, Number(($event.target as HTMLSelectElement).value))"
                    class="w-full text-xs border rounded px-1.5 py-1 bg-white"
                  >
                    <option value="">选择品目...</option>
                    <option v-for="cat in catalogs" :key="cat.id" :value="cat.id">
                      {{ cat.name }} ({{ cat.cas_number }})
                    </option>
                  </select>
                  <span v-else class="text-gray-500">
                    ID: {{ item.matched_catalog_id }}
                  </span>
                </td>
              </tr>
              <tr v-if="batchItems.length === 0">
                <td colspan="5" class="px-3 py-8 text-center text-gray-400">
                  暂无明细，请通过上方"追加"按钮手动添加
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 操作按钮 -->
        <div class="flex justify-end gap-2 pt-2">
          <Button size="sm" class="bg-gray-200 hover:bg-gray-300 text-gray-700" @click="resetAll">
            取消
          </Button>
          <Button
            size="sm"
            class="bg-emerald-600 hover:bg-emerald-700 text-white"
            :disabled="isConfirming || matchStats.matched === 0"
            @click="confirmBatch"
          >
            <Loader2 v-if="isConfirming" class="w-3.5 h-3.5 animate-spin mr-1" />
            确认到货并赋码 ({{ matchStats.matched }} 项)
          </Button>
        </div>
      </div>

      <!-- Step 3: 完成 -->
      <div v-if="step === 'done'" class="text-center py-10 space-y-4">
        <CheckCircle2 class="w-16 h-16 mx-auto text-green-500" />
        <h3 class="text-lg font-semibold text-gray-900">批次导入完成！</h3>
        <p class="text-sm text-gray-600">试剂已进入「分拣区」，等待研发人员或采购员扫码入库。</p>
        <Button size="sm" class="bg-blue-600 hover:bg-blue-700 text-white" @click="resetAll">
          导入新批次
        </Button>
      </div>
    </div>

    <!-- Toast -->
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
  </Card>
</template>
