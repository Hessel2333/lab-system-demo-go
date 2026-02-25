<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { Loader2, Search, Pencil, Trash2, Plus, X, FlaskConical, Archive } from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import ReagentCabinetManager from '@/components/reagents/ReagentCabinetManager.vue'

const API = '/api/reagents/catalogs'

const activeSubTab = ref<'catalogs' | 'cabinets'>('catalogs')

interface Catalog {
  id: number
  cas_number: string
  name: string
  alias: string
  formula: string
  category: string
  is_controlled: boolean
  description: string
  storage: string
  alert_threshold: number
  unit: string
  chemical_labels: string
  aliases: string
  storage_condition: string
  physical_state: string
}

const ALL_LABELS = ['普通化学品', '危险化学品', '易制毒化学品', '易制爆化学品', '剧毒化学品', '限制化学品']
const PHYSICAL_STATES = ['液体', '固体', '气体']

const isLoading = ref(false)
const catalogList = ref<Catalog[]>([])
const searchQuery = ref('')
const labelFilter = ref('')

// Edit dialog
const isDialogOpen = ref(false)
const isEditing = ref(false)
const isSaving = ref(false)
const editForm = ref<Partial<Catalog>>({})
const editLabels = ref<string[]>([])

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

const fetchCatalogs = async () => {
    isLoading.value = true
    try {
        const params: any = {}
        if (searchQuery.value) params.search = searchQuery.value
        if (labelFilter.value) params.label = labelFilter.value
        const res = await axios.get(API, { params })
        catalogList.value = res.data ?? []
    } catch (e) { console.error(e) }
    finally { isLoading.value = false }
}

onMounted(fetchCatalogs)

const parseLabels = (jsonStr: string): string[] => {
    try { return JSON.parse(jsonStr || '[]') } catch { return [] }
}

const openCreate = () => {
    isEditing.value = false
    editForm.value = { chemical_labels: '[]', physical_state: '液体', alert_threshold: 5, unit: '500ml' }
    editLabels.value = []
    isDialogOpen.value = true
}

const openEdit = (cat: Catalog) => {
    isEditing.value = true
    editForm.value = { ...cat }
    editLabels.value = parseLabels(cat.chemical_labels)
    isDialogOpen.value = true
}

const toggleLabel = (label: string) => {
    const idx = editLabels.value.indexOf(label)
    if (idx >= 0) editLabels.value.splice(idx, 1)
    else editLabels.value.push(label)
}

const saveForm = async () => {
    if (!editForm.value.name || !editForm.value.cas_number) {
        toast('请至少填写试剂名称和 CAS 号', 'error'); return
    }
    isSaving.value = true
    editForm.value.chemical_labels = JSON.stringify(editLabels.value)
    try {
        if (isEditing.value && editForm.value.id) {
            await axios.put(`${API}/${editForm.value.id}`, editForm.value)
            toast('品目已更新')
        } else {
            await axios.post(API, editForm.value)
            toast('品目已创建')
        }
        isDialogOpen.value = false
        fetchCatalogs()
    } catch (e: any) {
        toast(e.response?.data?.error || '保存失败', 'error')
    } finally { isSaving.value = false }
}

const deleteCatalog = async (id: number, name: string) => {
    if (!confirm(`确定要删除品目「${name}」吗？此操作不可撤销。`)) return
    try {
        await axios.delete(`${API}/${id}`)
        toast('品目已删除')
        fetchCatalogs()
    } catch (e: any) {
        toast(e.response?.data?.error || '删除失败', 'error')
    }
}

const labelColorMap: Record<string, string> = {
    '普通化学品': 'bg-gray-100 text-gray-600',
    '危险化学品': 'bg-orange-100 text-orange-700',
    '易制毒化学品': 'bg-purple-100 text-purple-700',
    '易制爆化学品': 'bg-red-100 text-red-700',
    '剧毒化学品': 'bg-red-200 text-red-900',
    '限制化学品': 'bg-yellow-100 text-yellow-700',
}
const getLabelColor = (label: string) => labelColorMap[label] || 'bg-gray-100 text-gray-600'

let searchTimer: any = null
const onSearchInput = () => {
    clearTimeout(searchTimer)
    searchTimer = setTimeout(fetchCatalogs, 300)
}
</script>

<template>
  <div class="space-y-4">
    <!-- 子 Tab 切换 -->
    <div class="flex gap-1 bg-gray-100 p-1 rounded-lg w-fit">
      <button @click="activeSubTab='catalogs'"
        :class="['flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all', activeSubTab==='catalogs' ? 'bg-white text-blue-700 shadow-sm' : 'text-gray-500 hover:text-gray-700']">
        <FlaskConical class="w-3.5 h-3.5" /> 品目数据库
      </button>
      <button @click="activeSubTab='cabinets'"
        :class="['flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md transition-all', activeSubTab==='cabinets' ? 'bg-white text-blue-700 shadow-sm' : 'text-gray-500 hover:text-gray-700']">
        <Archive class="w-3.5 h-3.5" /> 试剂柜管理
      </button>
    </div>

    <!-- 子页：品目数据库 -->
    <div v-show="activeSubTab === 'catalogs'">
    <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
        <div class="flex gap-3 items-center flex-wrap">
            <div class="relative w-64">
                <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
                <Input v-model="searchQuery" @input="onSearchInput" class="pl-9" placeholder="搜索名称 / CAS / 别称..." />
            </div>
            <select v-model="labelFilter" @change="fetchCatalogs"
                    class="text-sm border-gray-300 rounded-md px-3 py-2 focus:border-blue-500 focus:ring-blue-500">
                <option value="">全部分类</option>
                <option v-for="l in ALL_LABELS" :key="l" :value="l">{{ l }}</option>
            </select>
        </div>
        <Button @click="openCreate" class="bg-blue-600 hover:bg-blue-700 text-white shadow-sm flex items-center gap-1.5">
            <Plus class="w-4 h-4" /> 新增品目
        </Button>
    </div>

    <!-- Loading -->
    <div v-if="isLoading" class="flex justify-center py-12">
        <Loader2 class="w-8 h-8 animate-spin text-gray-400" />
    </div>

    <!-- Table -->
    <div v-else-if="catalogList.length > 0" class="overflow-x-auto rounded-lg border">
      <table class="w-full text-sm text-left">
        <thead class="text-xs text-gray-700 uppercase bg-gray-50">
          <tr>
            <th class="px-4 py-3">名称</th>
            <th class="px-4 py-3">CAS 号</th>
            <th class="px-4 py-3">分类标签</th>
            <th class="px-4 py-3">别称</th>
            <th class="px-4 py-3">物态</th>
            <th class="px-4 py-3">储存条件</th>
            <th class="px-4 py-3">规格</th>
            <th class="px-4 py-3 text-right">预警线</th>
            <th class="px-4 py-3 text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="cat in catalogList" :key="cat.id" class="bg-white border-b hover:bg-gray-50">
            <td class="px-4 py-3">
              <div class="font-medium text-gray-900">{{ cat.name }}</div>
              <div class="text-[11px] text-gray-400 font-mono">{{ cat.formula }}</div>
            </td>
            <td class="px-4 py-3 font-mono text-gray-600 text-xs">{{ cat.cas_number }}</td>
            <td class="px-4 py-3">
              <div class="flex flex-wrap gap-1">
                <span v-for="label in parseLabels(cat.chemical_labels)" :key="label"
                      :class="['text-[10px] px-1.5 py-0.5 rounded-full font-medium whitespace-nowrap', getLabelColor(label)]">
                  {{ label }}
                </span>
                <span v-if="!cat.chemical_labels || cat.chemical_labels === '[]'" class="text-[10px] text-gray-400">未设置</span>
              </div>
            </td>
            <td class="px-4 py-3 text-xs text-gray-500 max-w-[200px] truncate" :title="cat.aliases">
              {{ cat.aliases || '-' }}
            </td>
            <td class="px-4 py-3 text-xs text-gray-600">{{ cat.physical_state || '-' }}</td>
            <td class="px-4 py-3 text-xs text-gray-500 max-w-[180px] truncate" :title="cat.storage_condition">
              {{ cat.storage_condition || '-' }}
            </td>
            <td class="px-4 py-3 text-xs text-gray-600">{{ cat.unit }}</td>
            <td class="px-4 py-3 text-right text-xs">
              <span class="inline-block whitespace-nowrap font-bold text-orange-600 bg-orange-50 px-2 py-0.5 rounded border border-orange-200 shadow-sm">{{ cat.alert_threshold }} 件</span>
            </td>
            <td class="px-4 py-3 text-right">
              <div class="flex items-center justify-end gap-1">
                <button @click="openEdit(cat)" class="p-1.5 rounded-md hover:bg-blue-50 text-gray-400 hover:text-blue-600 transition">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click="deleteCatalog(cat.id, cat.name)" class="p-1.5 rounded-md hover:bg-red-50 text-gray-400 hover:text-red-600 transition">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty -->
    <div v-else class="text-center py-16 text-gray-500">
      <FlaskConical class="w-12 h-12 mx-auto mb-4 text-gray-300" />
      <p class="text-sm">暂无品目数据</p>
    </div>

    <!-- Edit/Create Dialog (Overlay) -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
      <div v-if="isDialogOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" @click.self="isDialogOpen = false">
        <div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl mx-4 max-h-[85vh] overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b">
            <h3 class="text-lg font-bold text-gray-900">{{ isEditing ? '编辑品目' : '新增品目' }}</h3>
            <button @click="isDialogOpen = false" class="p-1 rounded-md hover:bg-gray-100"><X class="w-5 h-5 text-gray-400" /></button>
          </div>
          <div class="px-6 py-5 space-y-5">

            <!-- Row 1: Name + CAS -->
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-1.5">
                <Label class="text-sm font-semibold">试剂名称 <span class="text-red-500">*</span></Label>
                <Input v-model="editForm.name" placeholder="如：无水乙醇" class="h-10" />
              </div>
              <div class="space-y-1.5">
                <Label class="text-sm font-semibold">CAS 号 <span class="text-red-500">*</span></Label>
                <Input v-model="editForm.cas_number" placeholder="如：64-17-5" class="h-10 font-mono" />
              </div>
            </div>

            <!-- Row 2: Formula + Unit + PhysicalState -->
            <div class="grid grid-cols-3 gap-4">
              <div class="space-y-1.5">
                <Label class="text-sm font-semibold">分子式</Label>
                <Input v-model="editForm.formula" placeholder="如：C2H6O" class="h-10 font-mono" />
              </div>
              <div class="space-y-1.5">
                <Label class="text-sm font-semibold">规格</Label>
                <Input v-model="editForm.unit" placeholder="如：500ml" class="h-10" />
              </div>
              <div class="space-y-1.5">
                <Label class="text-sm font-semibold">物态</Label>
                <div class="flex gap-2 mt-1">
                  <button v-for="s in PHYSICAL_STATES" :key="s" @click="editForm.physical_state = s"
                    class="flex-1 py-2 rounded-lg text-xs font-medium transition-all border"
                    :class="editForm.physical_state === s ? 'bg-blue-50 text-blue-700 border-blue-200 shadow-sm' : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50'">
                    {{ s }}
                  </button>
                </div>
              </div>
            </div>

            <!-- Row 3: Aliases -->
            <div class="space-y-1.5">
              <Label class="text-sm font-semibold">别称（逗号分隔）</Label>
              <Input v-model="editForm.aliases" placeholder="如：乙醇,酒精,Ethanol,EtOH" class="h-10" />
            </div>

            <!-- Row 4: Chemical Labels -->
            <div class="space-y-2">
              <Label class="text-sm font-semibold">化学品分类标签</Label>
              <div class="flex flex-wrap gap-2">
                <button v-for="label in ALL_LABELS" :key="label" @click="toggleLabel(label)"
                  class="text-xs px-3 py-1.5 rounded-full font-medium transition-all border"
                  :class="editLabels.includes(label)
                    ? getLabelColor(label) + ' border-transparent ring-1 ring-offset-1 ring-blue-400'
                    : 'bg-white text-gray-500 border-gray-200 hover:bg-gray-50'">
                  {{ label }}
                </button>
              </div>
            </div>

            <!-- Row 5: Storage Condition + Position + Alert -->
            <div class="grid grid-cols-3 gap-4">
              <div class="space-y-1.5">
                <Label class="text-sm font-semibold">储存条件</Label>
                <Input v-model="editForm.storage_condition" placeholder="如：阴凉干燥处" class="h-10" />
              </div>
              <div class="space-y-1.5">
                <Label class="text-sm font-semibold">默认库位</Label>
                <Input v-model="editForm.storage" placeholder="如：E309" class="h-10" />
              </div>
              <div class="space-y-1.5">
                 <Label class="text-sm font-semibold">预警线 (件)</Label>
                 <Input v-model.number="editForm.alert_threshold" type="number" min="0" placeholder="如：5" class="h-10" />
              </div>
            </div>

          </div>
          <div class="flex justify-end gap-3 px-6 py-4 border-t bg-gray-50 rounded-b-2xl">
            <Button variant="outline" @click="isDialogOpen = false" class="px-6">取消</Button>
            <Button @click="saveForm" :disabled="isSaving" class="px-8 bg-blue-600 hover:bg-blue-700 text-white">
              <Loader2 v-if="isSaving" class="w-4 h-4 animate-spin mr-1" />
              {{ isEditing ? '保存修改' : '创建品目' }}
            </Button>
          </div>
        </div>
      </div>
      </Transition>
    </Teleport>

    <!-- Toast -->
    <Transition
      enter-active-class="transition ease-out duration-300"
      enter-from-class="translate-y-4 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-4 opacity-0"
    >
      <div v-if="showToast" class="fixed bottom-6 right-6 z-[60] max-w-sm">
        <div :class="['px-4 py-3 rounded-lg shadow-lg border text-sm font-medium',
          toastType === 'success' ? 'bg-green-50 text-green-800 border-green-200' : 'bg-red-50 text-red-800 border-red-200']">
          {{ toastMessage }}
        </div>
      </div>
    </Transition>

    </div>
    <!-- 子页：试剂柜管理 -->
    <div v-show="activeSubTab === 'cabinets'">
      <ReagentCabinetManager />
    </div>
  </div>
</template>
