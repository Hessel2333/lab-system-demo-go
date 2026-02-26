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
const batchStatus = ref('')
const importedCreatedCount = ref(0)
const requests = ref<any[]>([])
const users = ref<any[]>([])

// 为了 UI 区分标签页
const activeTab = ref<'待处理' | '已匹配' | '已忽略' | '全部'>('待处理')

const showToast = ref(false)
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const toast = (msg: string, type: 'success' | 'error' = 'success') => {
    toastMessage.value = msg
    toastType.value = type
    showToast.value = true
    setTimeout(() => { showToast.value = false }, 3000)
}

// 加载待分发的申购单（替代原来的字典匹配）
const fetchRequests = async () => {
    try {
        const res = await axios.get('/api/reagents/requests')
        requests.value = res.data.filter((r: any) => ['待采购', '已接单'].includes(r.status))
    } catch { /* ignore */ }
}

// 加载可指派的人员列表
const fetchUsers = async () => {
    try {
        const res = await axios.get('/api/users')
        users.value = res.data
    } catch { /* ignore */ }
}

// 统计
const matchStats = computed(() => {
    const total = batchItems.value.length
    const matched = batchItems.value.filter(i => i.match_status === '自动匹配' || i.match_status === '手动匹配').length
    const ignored = batchItems.value.filter(i => i.match_status === '已忽略').length
    const unmatched = total - matched - ignored
    return { total, matched, unmatched, ignored }
})

// 按状态过滤列表
const filteredItems = computed(() => {
    if (activeTab.value === '全部') return batchItems.value
    if (activeTab.value === '待处理') return batchItems.value.filter(i => i.match_status === '未匹配')
    if (activeTab.value === '已匹配') return batchItems.value.filter(i => i.match_status === '自动匹配' || i.match_status === '手动匹配')
    if (activeTab.value === '已忽略') return batchItems.value.filter(i => i.match_status === '已忽略')
    return batchItems.value
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

const normalizeCell = (cell: any) => {
    if (cell == null) return ''
    if (typeof cell === 'string' || typeof cell === 'number' || typeof cell === 'boolean') {
        return String(cell).trim()
    }
    if (typeof cell === 'object') {
        if (typeof cell.w === 'string') return cell.w.trim()
        if (cell.v != null) return String(cell.v).trim()
    }
    return String(cell).trim()
}

const buildSourceRowHash = (row: any[]) => {
    const normalized = row.map((cell) => normalizeCell(cell).toLowerCase())
    return normalized.join('|')
}

const processFile = async (file: File) => {
    isUploading.value = true
    try {
        let items: any[] = []
        let autoPeriod = ''
        let orderNumber = ''
        const isManualSeed = file.size === 0

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
                const name = normalizeCell(row[23])
                if (!name || name === '商品名称') continue // 跳过空行或表头本身
                
                // 订单编号 (从“订单编号”第 1 列)
                if (!orderNumber && row[0]) {
                    orderNumber = normalizeCell(row[0])
                }

                // 自动尝试提取周期 (从“下单时间”第 11 列)
                if (!autoPeriod && row[10]) {
                    const dateStr = normalizeCell(row[10])
                    const match = dateStr.match(/^(\d{4}-\d{2})/)
                    if (match && match[1]) autoPeriod = match[1]
                }
                
                items.push({
                    row_hash: buildSourceRowHash(row),
                    reagent_name: name,
                    cas_number: '', // 易派客导出通常没有分离的 CAS 号
                    quantity: parseFloat(normalizeCell(row[27])) || 1,
                    unit: normalizeCell(row[28]) || '瓶',
                    material_category: normalizeCell(row[22]),
                    product_category: normalizeCell(row[24])
                })
            }

            if (items.length === 0) {
                toast('未能在 Excel 中解析到有效试剂明细数据', 'error')
                isUploading.value = false
                return
            }
            
            if (autoPeriod) {
                period.value = autoPeriod
            }

            // 先做重复识别，不改变现有导入流程和界面结构
            const precheck = await axios.post('/api/reagents/procurement-batches', {
                period: period.value,
                order_number: orderNumber,
                items: items,
                dry_run: true
            })
            const preSkipped = Number(precheck.data?.skipped_count ?? 0)
            if (preSkipped > 0) {
                const hintNames = Array.isArray(precheck.data?.duplicate_name_hints)
                    ? precheck.data.duplicate_name_hints.slice(0, 5).join('、')
                    : ''
                toast(hintNames
                    ? `检测到重复项目 ${preSkipped} 条（如：${hintNames}），导入时将自动跳过`
                    : `检测到重复项目 ${preSkipped} 条，导入时将自动跳过`
                )
            }
        }

        // 发送真实解析到的 items 到后端创建批次
        const res = await axios.post('/api/reagents/procurement-batches', {
            period: period.value,
            order_number: orderNumber,
            items: items
        })

        batchId.value = res.data.id || res.data.batch?.id || null
        batchItems.value = res.data.items || res.data.batch?.items || []
        batchStatus.value = res.data.batch?.status || '待确认'

        await fetchRequests()
        await fetchUsers()
        const createdCount = Number(res.data.created_count ?? batchItems.value.length ?? 0)
        importedCreatedCount.value = createdCount
        // 只要本次有新增条目，就应允许继续确认，不受历史批次状态干扰
        if (createdCount > 0) {
            batchStatus.value = '待确认'
        }
        const skippedCount = Number(res.data.skipped_count ?? 0)
        const responseMessage = String(res.data.message || '')
        const noNewRows = createdCount === 0 && (skippedCount > 0 || responseMessage.includes('无新增条目'))
        step.value = 'match'
        activeTab.value = '待处理'
        if (noNewRows) {
            if (batchStatus.value !== '待确认' && !isManualSeed) {
                toast(`检测到重复导入：无新增条目（重复 ${skippedCount} 条），已读取文件并展示历史批次`)
            } else {
                toast(`检测到重复导入：无新增条目（重复 ${skippedCount} 条）`)
            }
        } else if (skippedCount > 0) {
            toast(`导入完成：新增 ${createdCount} 条，跳过重复 ${skippedCount} 条`)
        } else if (isManualSeed) {
            toast('空批次已创建，请在下方手动追加明细')
        } else {
            toast(`成功解析并导入 ${createdCount} 条数据，请核对`)
        }

    } catch (error: any) {
        toast(error.response?.data?.error || '文件处理失败，请重试', 'error')
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

// --- Step 2: 手动匹配与忽略 ---
const updateItemMatch = async (item: any, selectedVal: string) => {
    if (!selectedVal) return
    const isUser = selectedVal.startsWith('usr_')
    const id = Number(selectedVal.split('_')[1])
    const payload = isUser ? { matched_user_id: id, cas_number: item.cas_number } : { matched_request_id: id, cas_number: item.cas_number }

    try {
        await axios.put(`/api/reagents/procurement-batches/${batchId.value}/items/${item.id}`, payload)
        if (isUser) {
            item.matched_user_id = id
            item.matched_request_id = null
        } else {
            item.matched_request_id = id
            item.matched_user_id = null
        }
        item.match_status = '手动匹配'
        toast('认领并关联成功')
    } catch {
        toast('认领失败', 'error')
    }
}

const ignoreItem = async (item: any) => {
    try {
        await axios.put(`/api/reagents/procurement-batches/${batchId.value}/items/${item.id}`, {
            match_status: '已忽略'
        })
        item.match_status = '已忽略'
        item.matched_catalog_id = null
        item.matched_request_id = null
    } catch {
        toast('忽略失败', 'error')
    }
}

const ignoreAllUnmatched = async () => {
    const unmatchedItems = batchItems.value.filter(i => i.match_status === '未匹配')
    if (unmatchedItems.length === 0) return
    
    try {
        await Promise.all(unmatchedItems.map(item => 
            axios.put(`/api/reagents/procurement-batches/${batchId.value}/items/${item.id}`, {
                match_status: '已忽略'
            })
        ))
        for (const item of unmatchedItems) {
            item.match_status = '已忽略'
            item.matched_catalog_id = null
            item.matched_request_id = null
        }
        toast(`已一键将 ${unmatchedItems.length} 项耗材标记为放行忽略`)
    } catch {
        toast('一键忽略部分失败', 'error')
    }
}

// --- Step 3: 确认并赋码 ---
const confirmBatch = async () => {
    if (!batchId.value) return
    if (batchStatus.value !== '待确认') {
        toast('该批次已确认，可直接到「到货台账」执行点验与入库')
        step.value = 'done'
        return
    }
    isConfirming.value = true
    try {
        const res = await axios.post(`/api/reagents/procurement-batches/${batchId.value}/confirm`)
        batchStatus.value = '已确认'
        toast(`到货确认完成！共生成 ${res.data.items_created} 件库存条目`)
        step.value = 'done'
    } catch (error: any) {
        toast(error.response?.data?.error || '确认失败，请重试', 'error')
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
    batchStatus.value = ''
    importedCreatedCount.value = 0
}

// 当前年月
const currentPeriod = new Date().toISOString().slice(0, 7) // e.g. "2026-02"

const createEmptyBatch = () => {
    if (!period.value) {
        toast('创建空批次必须手动指定所属周期', 'error')
        return
    }
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
        <div class="flex items-center gap-3">
          <label class="text-sm font-medium text-gray-700">所属周期</label>
          <input
            v-model="period"
            type="month"
            :max="currentPeriod"
            class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
          <span class="text-xs text-blue-600 bg-blue-50 px-2 py-1 rounded-md border border-blue-100">✨ 上传 Excel 会自动提取</span>
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
            @click="createEmptyBatch"
            class="text-xs text-blue-600 hover:text-blue-700 underline"
          >
            没有 Excel？手动创建空批次并逐条录入
          </button>
        </div>
      </div>

      <!-- Step 2: 明细确认与匹配 -->
      <div v-if="step === 'match'" class="space-y-4">
        <!-- 统计条与Tabs -->
        <div class="flex flex-col gap-2 mb-2">
          <div class="flex items-center gap-4 bg-gray-50 rounded-lg px-4 py-2.5 border">
            <span class="text-xs text-gray-600">批次 #{{ batchId }} · 周期 {{ period }}</span>
            <span class="text-xs font-medium text-emerald-600">✅ 已匹配 {{ matchStats.matched }}</span>
            <span v-if="matchStats.unmatched > 0" class="text-xs font-medium text-red-600">⚠️ 待处理 {{ matchStats.unmatched }}</span>
            <span v-if="matchStats.ignored > 0" class="text-xs font-medium text-gray-500">👻 已忽略 {{ matchStats.ignored }}</span>
            <span class="text-xs text-gray-500 ml-auto">共 {{ matchStats.total }} 项</span>
          </div>
          
          <div class="flex items-center justify-between border-b">
            <div class="flex gap-4">
              <button 
                v-for="tab in ['待处理', '已匹配', '已忽略', '全部']" 
                :key="tab"
                @click="activeTab = tab as any"
                :class="[ 'px-2 py-2 text-sm font-medium border-b-2 transition-colors', activeTab === tab ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300' ]"
              >
                {{ tab }}
                <span v-if="tab === '待处理'" class="ml-1 text-[10px] bg-red-100 text-red-600 px-1.5 py-0.5 rounded-full">{{ matchStats.unmatched }}</span>
              </button>
            </div>
            
            <Button 
                v-if="activeTab === '待处理' && matchStats.unmatched > 0"
                size="sm" variant="outline" class="h-7 text-xs text-gray-600 font-medium border-red-200 hover:bg-red-50 hover:text-red-600 transition"
                @click="ignoreAllUnmatched"
            >
              🧹一键忽略全部待处理杂项
            </Button>
          </div>
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
          <Button size="sm" variant="primary" class="text-xs shrink-0" @click="addManualItem">
            + 追加
          </Button>
        </div>

        <!-- 明细表格 -->
        <div class="apple-table-wrap max-h-96 overflow-y-auto">
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
              <tr v-for="item in filteredItems" :key="item.id" :class="[item.match_status === '未匹配' ? 'bg-red-50/50' : '', item.match_status === '已忽略' ? 'opacity-60 bg-gray-50' : '']">
                <td class="px-3 py-2 text-gray-900">{{ item.reagent_name }}</td>
                <td class="px-3 py-2 font-mono text-gray-500">{{ item.cas_number || '-' }}</td>
                <td class="px-3 py-2 text-center">{{ item.quantity }} {{ item.unit }}</td>
                <td class="px-3 py-2 text-center">
                  <span :class="[
                    'inline-flex items-center gap-0.5 text-[10px] px-2 py-0.5 rounded-full font-medium',
                    item.match_status === '自动匹配' ? 'bg-green-100 text-green-700' :
                    item.match_status === '手动匹配' ? 'bg-blue-100 text-blue-700' :
                    item.match_status === '已忽略' ? 'bg-gray-200 text-gray-600' :
                    'bg-red-100 text-red-700'
                  ]">
                    <CheckCircle2 v-if="item.match_status === '自动匹配' || item.match_status === '手动匹配'" class="w-3 h-3" />
                    <AlertCircle v-else class="w-3 h-3" />
                    {{ item.match_status }}
                  </span>
                </td>
                <td class="px-3 py-2 min-w-56">
                  <div v-if="item.match_status !== '已忽略'" class="flex flex-col gap-1.5">
                    <select
                      @change="updateItemMatch(item, ($event.target as HTMLSelectElement).value)"
                      class="w-full text-xs border border-gray-300 rounded px-1.5 py-1.5 bg-white focus:ring-1 focus:ring-blue-500"
                      :value="item.matched_request_id ? `req_${item.matched_request_id}` : (item.matched_user_id ? `usr_${item.matched_user_id}` : '')"
                    >
                      <option value="">分配给申购需求 / 指派个人...</option>
                      <optgroup label="最佳推荐 (现有申购单)">
                        <option v-for="req in requests" :key="`req_${req.id}`" :value="`req_${req.id}`">
                          {{ req.requestor?.real_name || '未知' }} - {{ req.reagent_catalog?.name }} (需{{ req.quantity }}瓶)
                        </option>
                      </optgroup>
                      <optgroup label="直接指派 (无申购单直接补录入库)">
                        <option v-for="user in users" :key="`usr_${user.id}`" :value="`usr_${user.id}`">
                          指派给：{{ user.real_name }} ({{ user.role === 'admin' ? '系统管理员' : user.role === 'leader' ? '课题组长' : '研发人员' }})
                        </option>
                      </optgroup>
                    </select>
                    
                    <div class="flex items-center justify-between">
                      <span v-if="item.matched_catalog_id" class="text-[10px] text-emerald-600 font-medium">✨ 系统归属品目 #{{ item.matched_catalog_id }}</span>
                      <span v-else class="text-[10px] text-gray-400">系统尚无分类档案</span>
                      <button v-if="item.match_status === '未匹配'" @click="ignoreItem(item)" class="text-[10px] text-gray-400 hover:text-red-500 transition underline decoration-dashed">直接忽略 (非试剂)</button>
                    </div>
                  </div>
                  <span v-else class="text-[10px] text-gray-400 border border-gray-200 px-1 py-0.5 rounded bg-gray-100 inline-block">🚫 无视通过 (不入库)</span>
                </td>
              </tr>
              <tr v-if="filteredItems.length === 0">
                <td colspan="5" class="px-3 py-8 text-center text-gray-400">
                  当前视图暂无明细记录
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 操作按钮 -->
        <div class="flex justify-end gap-2 pt-2">
          <Button size="sm" variant="secondary" @click="resetAll">
            取消
          </Button>
          <Button
            size="sm"
            class="bg-emerald-600 hover:bg-emerald-700 text-white"
            :disabled="isConfirming || matchStats.matched === 0"
            @click="confirmBatch"
          >
            <Loader2 v-if="isConfirming" class="w-3.5 h-3.5 animate-spin mr-1" />
            <span v-if="batchStatus !== '待确认'">已确认，去到货台账</span>
            <span v-else-if="matchStats.unmatched > 0">忽略其余杂项，仅赋码这 {{ matchStats.matched }} 项</span>
            <span v-else>确认入库 ({{ matchStats.matched }}项)</span>
          </Button>
        </div>
      </div>

      <!-- Step 3: 完成 -->
      <div v-if="step === 'done'" class="text-center py-10 space-y-4">
        <CheckCircle2 class="w-16 h-16 mx-auto text-green-500" />
        <h3 class="text-lg font-semibold text-gray-900">批次导入完成！</h3>
        <p class="text-sm text-gray-600">试剂已进入「暂存区」，等待研发人员或采购员扫码入库。</p>
        <Button size="sm" variant="primary" @click="resetAll">
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
      <div v-if="showToast" class="apple-toast-wrap">
        <div :class="[
          'apple-toast',
          toastType === 'success' ? 'apple-toast-success' : 'apple-toast-error'
        ]">
          <span>{{ toastMessage }}</span>
        </div>
      </div>
    </Transition>
  </Card>
</template>
