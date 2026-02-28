<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import Dialog from '@/components/ui/Dialog.vue'
import TableSection from '@/components/ui/TableSection.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { Plus, Search, Loader2 } from 'lucide-vue-next'

interface Supplier {
    ID: number
    name: string
    type: string
    contact_person: string
    phone: string
    email: string
    rating: number
    response_speed: number
    status: string
}

const suppliers = ref<Supplier[]>([])
const loading = ref(false)
const searchQuery = ref('')
const typeFilter = ref('全部')
const statusFilter = ref('全部')

const isDialogOpen = ref(false)
const currentSupplier = ref<Partial<Supplier>>({})
const isEdit = ref(false)

const fetchSuppliers = async () => {
    loading.value = true
    try {
        const res = await fetch('/api/suppliers')
        suppliers.value = await res.json()
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

onMounted(() => {
    fetchSuppliers()
})

const getRatingColor = (rating: number) => {
    if (rating >= 4.5) return 'text-green-500'
    if (rating >= 4.0) return 'text-blue-500'
    return 'text-yellow-500'
}

const openAddDialog = () => {
    currentSupplier.value = { status: 'active', rating: 5, response_speed: 5, type: 'instrument' }
    isEdit.value = false
    isDialogOpen.value = true
}

const openEditDialog = (sup: Supplier) => {
    currentSupplier.value = { ...sup }
    isEdit.value = true
    isDialogOpen.value = true
}

const handleSave = async () => {
    loading.value = true
    try {
        const url = isEdit.value 
            ? `/api/suppliers/${currentSupplier.value.ID}`
            : '/api/suppliers'
        
        const method = isEdit.value ? 'PUT' : 'POST'
        
        await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(currentSupplier.value)
        })
        
        await fetchSuppliers()
        isDialogOpen.value = false
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

const typeLabelMap: Record<string, string> = {
    instrument: '仪器设备',
    reagent: '试剂耗材',
    general: '综合服务'
}

const filteredSuppliers = computed(() => {
    let result = suppliers.value
    if (typeFilter.value !== '全部') {
        result = result.filter((sup) => (typeLabelMap[sup.type] || sup.type) === typeFilter.value)
    }
    if (statusFilter.value !== '全部') {
        result = result.filter((sup) => (sup.status === 'active' ? '合作中' : '已拉黑') === statusFilter.value)
    }
    const q = searchQuery.value.trim().toLowerCase()
    if (q) {
        result = result.filter((sup) =>
            sup.name?.toLowerCase().includes(q) ||
            sup.contact_person?.toLowerCase().includes(q) ||
            sup.phone?.toLowerCase().includes(q)
        )
    }
    return result
})
</script>

<template>
    <TableSection title="供应商管理" description="管理仪器、试剂及耗材的各类供应商库">
        <template #actions>
            <Button variant="primary" size="sm" @click="openAddDialog">
                <Plus class="w-4 h-4 mr-1" />
                新增供应商
            </Button>
        </template>

        <template #toolbar>
            <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
                <div class="relative w-full sm:w-80">
                    <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
                    <Input v-model="searchQuery" class="pl-9" placeholder="搜索供应商、联系人或电话..." />
                </div>
                <div class="apple-segmented w-fit sm:ml-auto">
                    <button
                        v-for="type in ['全部', '仪器设备', '试剂耗材', '综合服务']"
                        :key="type"
                        @click="typeFilter = type"
                        :class="['apple-segmented-btn', typeFilter === type ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']"
                    >
                        {{ type }}
                    </button>
                </div>
                <div class="apple-segmented w-fit">
                    <button
                        v-for="status in ['全部', '合作中', '已拉黑']"
                        :key="status"
                        @click="statusFilter = status"
                        :class="['apple-segmented-btn', statusFilter === status ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle']"
                    >
                        {{ status }}
                    </button>
                </div>
            </div>
        </template>

        <div v-if="loading" class="flex justify-center py-10">
            <Loader2 class="h-8 w-8 animate-spin text-gray-400" />
        </div>
        <div v-else-if="filteredSuppliers.length === 0" class="apple-table-empty">
            暂无符合条件的供应商记录。
        </div>
        <div v-else class="apple-table-wrap">
            <table class="w-full text-left text-sm">
                <thead class="bg-gray-50 text-gray-500 font-medium border-b border-gray-100 sticky top-0 z-10">
                    <tr>
                        <th class="px-6 py-3">供应商名称</th>
                        <th class="px-6 py-3">类型</th>
                        <th class="px-6 py-3">联系人</th>
                        <th class="px-6 py-3">综合评分</th>
                        <th class="px-6 py-3">当前状态</th>
                        <th class="px-6 py-3 text-right">操作</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                    <tr v-for="sup in filteredSuppliers" :key="sup.ID" class="hover:bg-gray-50/50 transition-colors">
                        <td class="px-6 py-4 font-medium text-gray-900">{{ sup.name }}</td>
                        <td class="px-6 py-4">
                            <span class="px-2 py-1 rounded text-xs bg-gray-100 text-gray-600 border border-gray-200">
                                {{ typeLabelMap[sup.type] || sup.type }}
                            </span>
                        </td>
                        <td class="px-6 py-4 text-gray-600">
                            <div>{{ sup.contact_person }}</div>
                            <div class="text-xs text-gray-400 mt-0.5">{{ sup.phone }}</div>
                        </td>
                        <td class="px-6 py-4">
                            <div class="flex items-center gap-2">
                                <span class="font-bold text-lg" :class="getRatingColor(sup.rating)">{{ sup.rating }}</span>
                                <div class="text-xs text-gray-400">
                                    <div>质量: {{ sup.rating }}</div>
                                    <div>响应: {{ sup.response_speed }}</div>
                                </div>
                            </div>
                        </td>
                        <td class="px-6 py-4">
                            <span
                                class="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium"
                                :class="sup.status === 'active' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'"
                            >
                                <span class="w-1.5 h-1.5 rounded-full" :class="sup.status === 'active' ? 'bg-green-500' : 'bg-red-500'"></span>
                                {{ sup.status === 'active' ? '合作中' : '已拉黑' }}
                            </span>
                        </td>
                        <td class="px-6 py-4 text-right">
                            <button @click="openEditDialog(sup)" class="text-blue-600 hover:text-blue-800 font-medium text-xs">详情/编辑</button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </TableSection>

    <Dialog :open="isDialogOpen" @close="isDialogOpen = false">
        <div class="p-6">
            <h3 class="text-lg font-bold mb-4">{{ isEdit ? '编辑供应商' : '新增供应商' }}</h3>
            
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">供应商名称</label>
                    <input v-model="currentSupplier.name" type="text" class="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border">
                </div>
                
                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
                        <select v-model="currentSupplier.type" class="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border">
                            <option value="instrument">仪器设备</option>
                            <option value="reagent">试剂耗材</option>
                            <option value="general">综合服务</option>
                        </select>
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">状态</label>
                        <select v-model="currentSupplier.status" class="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border">
                            <option value="active">合作中</option>
                            <option value="blacklist">已拉黑</option>
                        </select>
                    </div>
                </div>

                <div class="grid grid-cols-2 gap-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">联系人</label>
                        <input v-model="currentSupplier.contact_person" type="text" class="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">联系电话</label>
                        <input v-model="currentSupplier.phone" type="text" class="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border">
                    </div>
                </div>

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">综合评分 (0-5)</label>
                    <input v-model.number="currentSupplier.rating" type="number" step="0.1" max="5" min="0" class="w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border">
                </div>

                <div class="flex justify-end gap-3 mt-6">
                    <button @click="isDialogOpen = false" class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50">取消</button>
                    <button @click="handleSave" class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700">保存</button>
                </div>
            </div>
        </div>
    </Dialog>
</template>
