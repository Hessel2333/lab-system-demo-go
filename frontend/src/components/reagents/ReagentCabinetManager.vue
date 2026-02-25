<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { Plus, Pencil, Trash2, Loader2, FlaskConical, ShieldAlert } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

interface Cabinet {
  id: number
  name: string
  cabinet_type: string
  department_id: number
  location: string
  notes: string
}

interface Dept { id: number; name: string }

const cabinets = ref<Cabinet[]>([])
const depts = ref<Dept[]>([])
const isLoading = ref(false)

const showDialog = ref(false)
const isEditing = ref(false)
const form = ref<Partial<Cabinet> & { department_id: number }>({
  name: '', cabinet_type: '普通试剂柜', department_id: 0, location: '', notes: ''
})

const cabinetTypes = ['普通试剂柜', '易制毒制爆试剂柜']

const deptName = (id: number) => {
  if (id === 0) return '公共'
  return depts.value.find(d => d.id === id)?.name ?? `ID:${id}`
}

const groupedCabinets = computed(() => {
  const res: Record<string, Cabinet[]> = {}
  for (const c of cabinets.value) {
    const key = c.cabinet_type
    if (!res[key]) res[key] = []
    res[key].push(c)
  }
  return res
})

const fetch = async () => {
  isLoading.value = true
  try {
    const [cabRes, deptRes] = await Promise.all([
      axios.get('/api/reagents/cabinets'),
      axios.get('/api/departments')
    ])
    cabinets.value = cabRes.data ?? []
    depts.value = deptRes.data ?? []
  } finally {
    isLoading.value = false
  }
}

const openCreate = () => {
  form.value = { name: '', cabinet_type: '普通试剂柜', department_id: 0, location: '', notes: '' }
  isEditing.value = false
  showDialog.value = true
}

const openEdit = (c: Cabinet) => {
  form.value = { ...c }
  isEditing.value = true
  showDialog.value = true
}

const save = async () => {
  if (!form.value.name) { toast.error('柜名不能为空'); return }
  if (!form.value.cabinet_type) { toast.error('请选择柜类型'); return }
  try {
    if (isEditing.value && form.value.id) {
      await axios.put(`/api/reagents/cabinets/${form.value.id}`, form.value)
      toast.success('柜点位已更新')
    } else {
      await axios.post('/api/reagents/cabinets', form.value)
      toast.success('已新建试剂柜点位')
    }
    showDialog.value = false
    fetch()
  } catch (e: any) {
    toast.error('保存失败: ' + (e.response?.data?.error || e.message))
  }
}

const del = async (c: Cabinet) => {
  if (!confirm(`确认删除「${c.name}」？已分配该柜有在库试剂的情况下会阻止删除。`)) return
  try {
    await axios.delete(`/api/reagents/cabinets/${c.id}`)
    toast.success('已删除')
    fetch()
  } catch (e: any) {
    toast.error(e.response?.data?.error || '删除失败')
  }
}

onMounted(fetch)
</script>

<template>
  <div class="space-y-5">
    <!-- 顶栏 -->
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-base font-bold text-gray-900">试剂柜管理</h3>
        <p class="text-xs text-gray-500 mt-0.5">管理各团队的试剂柜点位，区分普通柜和管控品专柜</p>
      </div>
      <button @click="openCreate"
        class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors shadow-sm shadow-blue-200">
        <Plus class="w-4 h-4" /> 新建柜点位
      </button>
    </div>

    <div v-if="isLoading" class="flex justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-gray-400" />
    </div>

    <!-- 分类展示 -->
    <div v-else class="space-y-6">
      <div v-for="(items, type) in groupedCabinets" :key="String(type)" class="space-y-3">
        <div class="flex items-center gap-2">
          <component :is="type === '易制毒制爆试剂柜' ? ShieldAlert : FlaskConical"
            :class="['w-4 h-4', type === '易制毒制爆试剂柜' ? 'text-red-500' : 'text-blue-500']" />
          <h4 class="text-sm font-semibold text-gray-700">{{ type }}</h4>
          <span class="text-xs text-gray-400 bg-gray-100 px-2 py-0.5 rounded-full">{{ items.length }} 个</span>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <div v-for="cab in items" :key="cab.id"
            :class="['group relative border rounded-xl p-4 hover:shadow-sm transition-all', type === '易制毒制爆试剂柜' ? 'border-red-200 bg-red-50/40' : 'border-blue-100 bg-blue-50/30']">
            <div class="flex items-start justify-between">
              <div class="min-w-0 flex-1">
                <p class="font-semibold text-sm text-gray-900 truncate">🗄️ {{ cab.name }}</p>
                <p class="text-xs text-gray-500 mt-0.5">{{ deptName(cab.department_id) }} · {{ cab.location || '位置未设置' }}</p>
                <p v-if="cab.notes" class="text-xs text-gray-400 mt-1 truncate">{{ cab.notes }}</p>
              </div>
              <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2 shrink-0">
                <button @click="openEdit(cab)" class="p-1.5 rounded-lg hover:bg-gray-100 text-gray-500 hover:text-blue-600 transition-colors">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click="del(cab)" class="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500 transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="Object.keys(groupedCabinets).length === 0" class="text-center py-12 text-gray-400">
        <FlaskConical class="w-10 h-10 mx-auto mb-3 text-gray-300" />
        <p class="text-sm">暂无试剂柜点位，点击右上角新建</p>
      </div>
    </div>

    <!-- 新建/编辑弹窗 -->
    <div v-if="showDialog" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      @click.self="showDialog = false">
      <div class="bg-white rounded-2xl w-full max-w-md shadow-xl p-6 space-y-5">
        <h3 class="text-lg font-bold text-gray-900">{{ isEditing ? '编辑' : '新建' }}试剂柜点位</h3>
        <div class="space-y-4">
          <div class="space-y-1.5">
            <label class="text-sm font-semibold text-gray-700">柜名 <span class="text-red-500">*</span></label>
            <input v-model="form.name" placeholder="例如：分析团队-普通试剂柜B"
              class="w-full h-10 px-3 bg-gray-50 border border-gray-200 rounded-lg text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none" />
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-semibold text-gray-700">柜类型 <span class="text-red-500">*</span></label>
            <div class="flex gap-2">
              <button v-for="t in cabinetTypes" :key="t" @click="form.cabinet_type = t"
                :class="['flex-1 py-2 text-sm font-medium rounded-lg border transition-colors', form.cabinet_type === t ? (t === '易制毒制爆试剂柜' ? 'bg-red-600 text-white border-red-600' : 'bg-blue-600 text-white border-blue-600') : 'border-gray-200 text-gray-600 hover:bg-gray-50']">
                {{ t === '易制毒制爆试剂柜' ? '⚠️ ' : '🗄️ ' }}{{ t }}
              </button>
            </div>
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-semibold text-gray-700">归属团队</label>
            <select v-model="form.department_id"
              class="w-full h-10 px-3 bg-gray-50 border border-gray-200 rounded-lg text-sm focus:border-blue-500 outline-none">
              <option :value="0">公共（全体可用）</option>
              <option v-for="d in depts" :key="d.id" :value="d.id">{{ d.name }}</option>
            </select>
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-semibold text-gray-700">所在房间/位置</label>
            <input v-model="form.location" placeholder="例如：E309 / 东楼402"
              class="w-full h-10 px-3 bg-gray-50 border border-gray-200 rounded-lg text-sm focus:border-blue-500 outline-none" />
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-semibold text-gray-700">备注（选填）</label>
            <textarea v-model="form.notes" rows="2" placeholder="如：双人双锁，需申请使用"
              class="w-full p-3 bg-gray-50 border border-gray-200 rounded-lg text-sm resize-none focus:border-blue-500 outline-none"></textarea>
          </div>
        </div>
        <div class="flex gap-3">
          <button @click="showDialog = false"
            class="flex-1 py-2 text-sm font-medium text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors">取消</button>
          <button @click="save"
            class="flex-1 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-xl transition-colors shadow-sm shadow-blue-200">
            {{ isEditing ? '保存更改' : '确认新建' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
