<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Dialog from '@/components/ui/Dialog.vue'
import { Plus } from 'lucide-vue-next'

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
</script>

<template>
    <div class="h-full flex flex-col bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
        <!-- Header -->
        <div class="p-6 border-b border-gray-100 flex justify-between items-center bg-white">
            <div>
                <h1 class="text-xl font-bold text-gray-900">供应商管理</h1>
                <p class="text-sm text-gray-500 mt-1">管理仪器、试剂及耗材的各类供应商库</p>
            </div>
            <button @click="openAddDialog" class="bg-black text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-800 transition-colors shadow-sm flex items-center gap-2">
                <Plus class="w-4 h-4" /> 新增供应商
            </button>
        </div>

        <!-- Table -->
        <div class="flex-1 overflow-auto">
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
                    <tr v-for="sup in suppliers" :key="sup.ID" class="hover:bg-gray-50/50 transition-colors">
                        <td class="px-6 py-4 font-medium text-gray-900">{{ sup.name }}</td>
                        <td class="px-6 py-4">
                            <span class="px-2 py-1 rounded text-xs bg-gray-100 text-gray-600 border border-gray-200">
                                {{ { instrument: '仪器设备', reagent: '试剂耗材', active: '综合' }[sup.type] || sup.type }}
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
                            <span class="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium"
                                :class="sup.status === 'active' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'">
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
    </div>

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
