<script setup lang="ts">
import { ref, watch } from 'vue'
import { Sparkles, Loader2, CheckCircle, History, Star, ArrowRight, Package } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import axios from 'axios'
import Switch from '@/components/ui/switch/index.vue'

const emit = defineEmits(['request-submitted'])

const prompt = ref('')
const isAnalyzing = ref(false)
const isSubmitting = ref(false)
const parsedResult = ref<any>(null)

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

// Form fields for Manual Entry and AI correction
const formFields = ref({
    name: '',
    cas_number: '',
    quantity: 1,
    unit: '500ml',
    is_controlled: false,
    request_type: '日常', // New field
    expected_delivery: '', // New field
    project_name: '', // New field
    project_id: '', // New field
    remarks: ''
})

const commonReagents = [
    { name: '无水乙醇', cas: '64-17-5', unit: '500ml', controlled: false },
    { name: '乙腈 (色谱级)', cas: '75-05-8', unit: '4L', controlled: false },
    { name: '甲醇 (色谱级)', cas: '67-56-1', unit: '4L', controlled: true },
    { name: '二氯甲烷', cas: '75-09-2', unit: '500ml', controlled: false },
    { name: '硫酸 (98%)', cas: '7664-93-9', unit: '500ml', controlled: true },
]

const recentRequests = ref<any[]>([])

const fetchRecentRequests = async () => {
    try {
        const res = await axios.get('/api/reagents/requests')
        // 确保按照创建时间降序（最新在前）
        const sortedData = [...res.data].sort((a: any, b: any) => {
            const timeA = new Date(a.created_at).getTime() || 0;
            const timeB = new Date(b.created_at).getTime() || 0;
            return timeB - timeA;
        })
        // Take last 5 unique items based on names
        const unique = new Map()
        sortedData.forEach((r: any) => {
            if (!unique.has(r.reagent_catalog?.name) && unique.size < 5) {
                unique.set(r.reagent_catalog?.name, {
                    request_id: r.id,
                    created_at: r.created_at,
                    name: r.reagent_catalog.name,
                    cas: r.reagent_catalog.cas_number,
                    unit: r.reagent_catalog.unit,
                    controlled: r.reagent_catalog.is_controlled
                })
            }
        })
        recentRequests.value = Array.from(unique.values())
    } catch (error) {
        console.error("Failed to fetch recent requests", error)
    }
}

const quickFill = (reagent: any) => {
    formFields.value = {
        name: reagent.name,
        cas_number: reagent.cas,
        quantity: 1,
        unit: reagent.unit || '500ml',
        is_controlled: reagent.controlled || false,
        request_type: '日常',
        expected_delivery: '',
        project_name: '',
        project_id: '',
        remarks: '快速复购'
    }
}

fetchRecentRequests()

// --- 库存状态查询 ---
const stockInfo = ref<any>(null)
const isCheckingStock = ref(false)
let stockCheckTimer: any = null

const checkStock = async () => {
    const cas = formFields.value.cas_number?.trim()
    const name = formFields.value.name?.trim()
    if (!cas && !name) { stockInfo.value = null; return }

    isCheckingStock.value = true
    try {
        const params: any = {}
        if (cas) params.cas_number = cas
        else params.name = name
        const res = await axios.get('/api/reagents/stock-check', { params })
        stockInfo.value = res.data
        if (stockInfo.value?.catalog) {
            formFields.value.is_controlled = stockInfo.value.catalog.is_controlled || false
        }
    } catch {
        stockInfo.value = null
    } finally {
        isCheckingStock.value = false
    }
}

// 防抖监听表单字段变化
watch(() => formFields.value.cas_number, () => {
    clearTimeout(stockCheckTimer)
    stockCheckTimer = setTimeout(checkStock, 600)
})
watch(() => formFields.value.name, () => {
    if (!formFields.value.cas_number) {
        clearTimeout(stockCheckTimer)
        stockCheckTimer = setTimeout(checkStock, 600)
    }
})

// Step 1: AI Parse
const analyzeRequest = async () => {
    if (!prompt.value.trim()) return
    
    isAnalyzing.value = true
    parsedResult.value = null
    
    try {
        const response = await axios.post('/api/reagents/ai/parse', {
            message: prompt.value
        })
        parsedResult.value = response.data
        
        // Populate the editable form fields with AI parsed data
        if (parsedResult.value && parsedResult.value.parsed_catalog) {
             formFields.value = {
                name: parsedResult.value.parsed_catalog.name || '',
                cas_number: parsedResult.value.parsed_catalog.cas_number || '',
                quantity: parsedResult.value.quantity || 1,
                unit: parsedResult.value.parsed_catalog.unit || '500ml',
                is_controlled: parsedResult.value.parsed_catalog.is_controlled || false,
                request_type: parsedResult.value.request_type || '日常',
                expected_delivery: parsedResult.value.expected_delivery || '',
                project_name: parsedResult.value.project_name || '',
                project_id: parsedResult.value.project_id || '',
                remarks: prompt.value // Keep original prompt as part of remarks
            }
        }
    } catch (error: any) {
        console.error(error)
        const errMsg = error.response?.data?.error || "解析失败: 无法理解该请求，请重试或修改描述。"
        toast("AI 解析出错: " + errMsg, 'error')
    } finally {
        isAnalyzing.value = false
    }
}

// Step 2: Confirm & Submit
const submitRequest = async () => {
    // Validate minimal fields
    if (!formFields.value.name || !formFields.value.quantity) {
        toast("请填写完整的试剂名称和数量", 'error')
        return
    }

    isSubmitting.value = true

    try {
        // --- Find-or-Create Catalog ---
        let catalogId: number | null = null

        // Try to find existing catalog by name or CAS number
        const allCatalogs = await axios.get('/api/reagents/catalogs')
        const casNumber = formFields.value.cas_number?.trim()
        const name = formFields.value.name?.trim()

        const existing = allCatalogs.data.find((c: any) => {
            if (casNumber && c.cas_number && casNumber === c.cas_number) return true
            if (name && c.name && name === c.name) return true
            return false
        })

        if (existing) {
            // Reuse existing catalog
            catalogId = existing.id
        } else {
            // Create a new catalog entry
            const catalogPayload = {
                name: formFields.value.name,
                cas_number: formFields.value.cas_number,
                unit: formFields.value.unit,
                is_controlled: formFields.value.is_controlled,
                category: "通用试剂",
                storage: formFields.value.is_controlled ? "防爆柜" : "常温柜",
                alert_threshold: 5
            }
            const catRes = await axios.post('/api/reagents/catalogs', catalogPayload)
            catalogId = catRes.data.id
        }

        if (!catalogId) {
            toast("无法确定试剂品目，请检查输入。", 'error')
            return
        }

        // --- Submit Request ---
        const payload = {
            reagent_catalog_id: catalogId,
            quantity: Number(formFields.value.quantity),
            requestor_id: 1,
            request_type: formFields.value.request_type,
            expected_delivery: formFields.value.expected_delivery,
            project_name: formFields.value.project_name,
            project_id: formFields.value.project_id,
            remarks: formFields.value.remarks || prompt.value
        }

        await axios.post('/api/reagents/requests', payload)

        toast("申购成功：您的试剂申购单已提交并等待审批。")
        emit('request-submitted')
        resetForm()
        fetchRecentRequests() // 刷新复购历史，确保新提交立即出现在顶部

    } catch (error: any) {
        console.error('[ReagentRequestWizard] submitRequest failed:', error.response?.data || error.message)
        const errMsg = error.response?.data?.error || '请求失败，请检查网络或联系管理员。'
        toast(`提交失败：${errMsg}`, 'error')
    } finally {
        isSubmitting.value = false
    }
}

const resetForm = () => {
    prompt.value = ''
    parsedResult.value = null
    formFields.value = {
        name: '',
        cas_number: '',
        quantity: 1,
        unit: '500ml',
        is_controlled: false,
        request_type: '日常',
        expected_delivery: '',
        project_name: '',
        project_id: '',
        remarks: ''
    }
}
</script>

<template>
  <div class="w-full">
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      
      <!-- Left Column: Main Form Area (Current view) -->
      <div class="lg:col-span-2 space-y-6">
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden p-6 space-y-6">
                <!-- Input Area -->
                <div class="relative">
                    <textarea 
                        v-model="prompt" 
                        placeholder="例如：我需要申购 5 瓶丙酮用于清洗实验，每瓶 500ml。CAS 号为 67-64-1" 
                        class="min-h-[120px] w-full p-4 rounded-xl border border-gray-200 bg-gray-50 text-base shadow-inner focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-all resize-none"
                        @keydown.enter.ctrl.prevent="analyzeRequest"
                    />
                    <div class="absolute bottom-4 right-4">
                        <Button 
                            class="rounded-xl px-6 h-10 shadow-sm"
                            :disabled="!prompt || isAnalyzing || isSubmitting"
                            @click="analyzeRequest"
                        >
                            <Loader2 v-if="isAnalyzing" class="h-4 w-4 animate-spin mr-2" />
                            <Sparkles v-else class="h-4 w-4 mr-2" />
                            {{ isAnalyzing ? '正在解析...' : '智能识别' }}
                        </Button>
                    </div>
                </div>

                <!-- Analysis Result (Editable) -->
                <transition name="fade">
                    <div v-if="parsedResult" class="space-y-4 pt-4 border-t">
                        <div class="flex items-center justify-between">
                             <div class="flex items-center gap-2 text-blue-700 font-medium">
                                <CheckCircle class="h-5 w-5 text-green-500" />
                                <span>识别成功，请核对并完善以下信息：</span>
                            </div>
                        </div>
                    </div>
                </transition>
            </div>
            
            <div class="flex items-center gap-4 py-2">
                <div class="flex-1 h-px bg-gray-200"></div>
                <span class="text-sm text-gray-400">或者直接填写申购表单</span>
                <div class="flex-1 h-px bg-gray-200"></div>
            </div>

            <!-- Manual Form (Always Visible) -->
            <div class="bg-white rounded-2xl border border-gray-200 shadow-sm p-8 space-y-8">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-6">
                        <!-- Left Column -->
                        <div class="space-y-6">
                            <div class="space-y-2.5">
                                <Label class="text-sm font-semibold text-gray-700">试剂名称 <span class="text-red-500">*</span></Label>
                                <Input v-model="formFields.name" placeholder="请输入标准化学名称" class="h-11 bg-white border-gray-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all" />
                            </div>
                            <div class="space-y-2.5">
                                <Label class="text-sm font-semibold text-gray-700">CAS 编号</Label>
                                <Input v-model="formFields.cas_number" placeholder="例如: 64-17-5" class="h-11 bg-white border-gray-200 font-mono focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all" />
                            </div>
                        </div>
                        
                        <!-- Right Column -->
                        <div class="space-y-6">
                            <div class="flex space-x-6">
                                <div class="space-y-2.5 flex-1">
                                    <Label class="text-sm font-semibold text-gray-700">申购数量 <span class="text-red-500">*</span></Label>
                                    <Input v-model="formFields.quantity" type="number" min="1" class="h-11 bg-white border-gray-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all" />
                                </div>
                                <div class="space-y-2.5 flex-1">
                                    <Label class="text-sm font-semibold text-gray-700">包装规格 / 单位</Label>
                                    <Input v-model="formFields.unit" placeholder="例如: 500ml, 1kg" class="h-11 bg-white border-gray-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all" />
                                </div>
                            </div>
                            
                            <div class="p-4 bg-orange-50/50 border border-orange-200 rounded-xl flex items-center justify-between group hover:bg-orange-50 transition-colors">
                                <div class="flex-1">
                                    <h4 class="text-sm font-bold text-orange-900 mb-0.5">易制毒/易制爆管制</h4>
                                    <p class="text-[11px] text-orange-600/80 leading-relaxed font-medium">
                                        需要双人双锁核验 
                                        <span v-if="stockInfo?.catalog" class="ml-1 text-orange-500 font-bold">(系统自动判定)</span>
                                    </p>
                                </div>
                                <Switch v-model:checked="formFields.is_controlled" class="ml-4" :disabled="!!stockInfo?.catalog" />
                            </div>
                        </div>
                    </div>

                    <!-- Extended Details Row -->
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-6 pt-2 border-t border-gray-100">
                        <div class="space-y-6">
                            <div class="space-y-2.5">
                                <Label class="text-sm font-semibold text-gray-700">需求类型 <span class="text-red-500">*</span></Label>
                                <div class="flex gap-2">
                                    <button
                                        v-for="type in ['日常', '储备', '紧急']"
                                        :key="type"
                                        @click="formFields.request_type = type"
                                        class="flex-1 py-2.5 rounded-lg text-sm font-medium transition-all border"
                                        :class="formFields.request_type === type 
                                            ? (type === '紧急' ? 'bg-red-50 text-red-700 border-red-200 shadow-sm' : 'bg-blue-50 text-blue-700 border-blue-200 shadow-sm')
                                            : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50'"
                                    >
                                        {{ type }}
                                    </button>
                                </div>
                            </div>
                            <div class="space-y-2.5">
                                <Label class="text-sm font-semibold text-gray-700">要求交期</Label>
                                <Input type="date" v-model="formFields.expected_delivery" class="h-11 bg-white border-gray-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all" />
                            </div>
                        </div>

                        <div class="space-y-6">
                            <div class="space-y-2.5">
                                <Label class="text-sm font-semibold text-gray-700">所属项目</Label>
                                <Input v-model="formFields.project_name" placeholder="例如: 某某研发项目" class="h-11 bg-white border-gray-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all" />
                            </div>
                            <div class="space-y-2.5">
                                <Label class="text-sm font-semibold text-gray-700">项目编号</Label>
                                <Input v-model="formFields.project_id" placeholder="例如: PROJ-2026-X1" class="h-11 bg-white border-gray-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all font-mono text-sm" />
                            </div>
                        </div>
                    </div>

                    <!-- Bottom Fields (Full Width) -->
                    <div class="space-y-2.5 pt-2 border-t border-gray-100">
                        <Label class="text-sm font-semibold text-gray-700">申购用途详细说明</Label>
                        <Input v-model="formFields.remarks" placeholder="如有其他特殊要求请在此说明..." class="h-11 bg-white border-gray-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 rounded-lg transition-all" />
                    </div>
                    
                    <!-- Submission Actions -->
                    <div class="flex justify-end gap-4 pt-4">
                        <Button variant="outline" @click="resetForm" class="h-11 px-8 rounded-xl font-medium text-gray-600 hover:bg-gray-50 border-gray-200 transition-all">清空/取消</Button>
                        <Button @click="submitRequest" variant="primary" :disabled="isSubmitting" class="h-11 px-10 shadow-lg shadow-blue-500/20 rounded-xl font-bold transition-all transform hover:-translate-y-0.5 active:translate-y-0">
                            <Loader2 v-if="isSubmitting" class="mr-2 h-5 w-5 animate-spin" />
                            <CheckCircle v-else class="mr-2 h-5 w-5" />
                            确认提交申购
                        </Button>
                    </div>
                </div>
      </div>

      <!-- Right Column: Quick Fill Sections -->
      <div class="lg:col-span-1 space-y-6">

          <!-- Stock Status Panel -->
          <Transition
            enter-active-class="transition ease-out duration-300"
            enter-from-class="opacity-0 -translate-y-2"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition ease-in duration-200"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
          >
          <div v-if="stockInfo || isCheckingStock" class="rounded-2xl border shadow-sm overflow-hidden"
               :class="stockInfo?.in_stock === 0 ? 'bg-red-50/50 border-red-200' : 'bg-emerald-50/50 border-emerald-200'">
            <div class="px-5 py-3 border-b flex items-center gap-2"
                 :class="stockInfo?.in_stock === 0 ? 'border-red-100' : 'border-emerald-100'">
              <Package class="w-4 h-4" :class="stockInfo?.in_stock === 0 ? 'text-red-500' : 'text-emerald-600'" />
              <span class="text-sm font-bold" :class="stockInfo?.in_stock === 0 ? 'text-red-700' : 'text-emerald-700'">库存状态</span>
              <Loader2 v-if="isCheckingStock" class="w-3.5 h-3.5 animate-spin text-gray-400 ml-auto" />
            </div>
            <div v-if="stockInfo" class="px-5 py-4 space-y-3">
              <div class="text-xs font-semibold text-gray-700">{{ stockInfo.catalog.name }}
                <span class="text-gray-400 font-mono ml-1">{{ stockInfo.catalog.cas_number }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2 text-center">
                <div class="bg-white rounded-lg py-2 px-1 border border-gray-100">
                  <div class="text-lg font-bold" :class="stockInfo.in_stock === 0 ? 'text-red-600' : 'text-emerald-600'">{{ stockInfo.in_stock }}</div>
                  <div class="text-[10px] text-gray-500">在库</div>
                </div>
                <div class="bg-white rounded-lg py-2 px-1 border border-gray-100">
                  <div class="text-lg font-bold text-blue-600">{{ stockInfo.pending_arrival }}</div>
                  <div class="text-[10px] text-gray-500">待到货</div>
                </div>
                <div class="bg-white rounded-lg py-2 px-1 border border-gray-100">
                  <div class="text-lg font-bold text-amber-600">{{ stockInfo.pending_requests }}</div>
                  <div class="text-[10px] text-gray-500">待审单</div>
                </div>
              </div>
              <div class="text-xs leading-relaxed px-2 py-2 rounded-lg"
                   :class="stockInfo.in_stock === 0 ? 'bg-red-100/60 text-red-700' : 'bg-emerald-100/60 text-emerald-700'">
                {{ stockInfo.advice }}
              </div>
            </div>
          </div>
          </Transition>

          <!-- Common Reagents -->
          <div class="bg-white rounded-2xl shadow-sm border border-gray-200 p-5 space-y-4">
              <div class="flex items-center gap-2 text-sm font-bold text-gray-500 uppercase tracking-wider pb-2 border-b border-gray-100">
                  <Star class="h-4 w-4 text-amber-400 fill-amber-400" />
                  <span>快捷选取表</span>
              </div>
              <!-- Fixed height scrollable container -->
              <div class="grid grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-1 custom-scrollbar">
                  <button 
                      v-for="r in commonReagents" 
                      :key="r.cas"
                      @click="quickFill(r)"
                      class="flex flex-col p-3 bg-gray-50/50 hover:bg-white border border-transparent hover:border-blue-200 hover:shadow-sm rounded-xl transition-all duration-300 w-full text-left group"
                  >
                      <div class="flex-1 min-w-0 mb-2">
                        <span class="text-xs font-bold text-gray-800 line-clamp-2 leading-snug">{{ r.name }}</span>
                      </div>
                      <div class="flex items-center justify-between mt-auto">
                        <span class="text-[9px] px-1.5 py-0.5 bg-gray-100 text-gray-500 rounded-md whitespace-nowrap">{{ r.unit }}</span>
                        <ArrowRight class="h-3 w-3 text-gray-300 group-hover:text-blue-500 transition-colors flex-shrink-0" />
                      </div>
                  </button>
              </div>
          </div>

          <!-- Recent History -->
          <div v-if="recentRequests.length > 0" class="bg-blue-50/30 rounded-2xl border border-blue-100 p-5 space-y-4">
              <div class="flex items-center gap-2 text-sm font-bold text-blue-600 uppercase tracking-wider pb-2 border-b border-blue-100/50">
                  <History class="h-4 w-4" />
                  <span>复购记录 (快捷回填)</span>
              </div>
              <div class="grid grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-1 custom-scrollbar">
                  <button 
                      v-for="r in recentRequests" 
                      :key="r.cas"
                      @click="quickFill(r)"
                      class="flex flex-col p-3 bg-white border border-blue-100/50 hover:border-blue-300 hover:shadow-sm rounded-xl transition-all duration-300 w-full text-left group"
                  >
                      <div class="mb-1">
                          <div class="flex items-center justify-between mb-0.5">
                              <span class="text-[9px] text-blue-500 font-bold">单号 #{{ r.request_id }}</span>
                              <span class="text-[9px] text-gray-400 font-medium">{{ new Date(r.created_at).toLocaleDateString('zh-CN', {month: 'numeric', day: 'numeric'}) }}</span>
                          </div>
                          <span class="text-xs font-bold text-gray-800 line-clamp-1 block">{{ r.name }}</span>
                      </div>
                      <div class="flex justify-end mt-auto">
                        <ArrowRight class="h-3 w-3 text-blue-200 group-hover:text-blue-500 transition-colors flex-shrink-0" />
                      </div>
                  </button>
              </div>
          </div>
      </div>
    </div>

    <!-- Toast Notification -->
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
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
