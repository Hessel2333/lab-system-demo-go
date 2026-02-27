<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, watch } from 'vue'
import axios from 'axios'
import { useRoute } from 'vue-router'
import { ClipboardCheck, FileSpreadsheet, Layers, List, PackageCheck, ShoppingCart } from 'lucide-vue-next'
import { useSessionStore } from '@/stores/session'

type TabId = 'requests' | 'history' | 'arrival-confirm' | 'unified-inventory' | 'usage' | 'batch-import' | 'receiving'

interface TabItem {
  id: TabId
  label: string
  icon: any
  badge?: number
}

const activeTab = ref<TabId>('requests')
const requestList = ref<any>(null)
const dispenseTodoCount = ref(0)
const sessionStore = useSessionStore()
const route = useRoute()

const ReagentRequestWizard = defineAsyncComponent(() => import('@/components/reagents/ReagentRequestWizard.vue'))
const ReagentRequestList = defineAsyncComponent(() => import('@/components/reagents/ReagentRequestList.vue'))
const ReagentUnifiedInventory = defineAsyncComponent(() => import('@/components/reagents/ReagentUnifiedInventory.vue'))
const ProcurementBatchImport = defineAsyncComponent(() => import('@/components/reagents/ProcurementBatchImport.vue'))
const ProcurementReceiving = defineAsyncComponent(() => import('@/components/reagents/ProcurementReceiving.vue'))
const ReagentDispensePanel = defineAsyncComponent(() => import('@/components/reagents/ReagentDispensePanel.vue'))
const ResearcherArrivalList = defineAsyncComponent(() => import('@/components/reagents/ResearcherArrivalList.vue'))

const tabs = computed<TabItem[]>(() => {
  if (sessionStore.currentRole === 'researcher') {
    return [
      { id: 'requests', label: '申购申请', icon: ShoppingCart },
      { id: 'history', label: '申购进度', icon: List },
      { id: 'arrival-confirm', label: '到货确认', icon: PackageCheck },
      { id: 'unified-inventory', label: '库存台账', icon: Layers },
      { id: 'usage', label: '领用台账', icon: ClipboardCheck, badge: dispenseTodoCount.value },
    ]
  }
  if (sessionStore.currentRole === 'procurement') {
    return [
      { id: 'history', label: '申购台账', icon: List },
      { id: 'batch-import', label: '采购导入', icon: FileSpreadsheet },
      { id: 'receiving', label: '到货台账', icon: PackageCheck },
      { id: 'unified-inventory', label: '库存台账', icon: Layers },
      { id: 'usage', label: '领用台账', icon: ClipboardCheck, badge: dispenseTodoCount.value },
    ]
  }
  return [
    { id: 'history', label: '申购台账', icon: List },
    { id: 'usage', label: '领用台账', icon: ClipboardCheck, badge: dispenseTodoCount.value },
    { id: 'unified-inventory', label: '库存台账', icon: Layers },
  ]
})

const ensureActiveTab = () => {
  const allowed = tabs.value.map((item) => item.id)
  if (!allowed.includes(activeTab.value)) {
    activeTab.value = tabs.value[0]?.id || 'history'
  }
}

const applyRouteTab = () => {
  const raw = Array.isArray(route.query.tab) ? route.query.tab[0] : route.query.tab
  if (typeof raw !== 'string') return
  const allowed = tabs.value.map((item) => item.id)
  if (allowed.includes(raw as TabId)) {
    activeTab.value = raw as TabId
  }
}

const fetchDispenseTodo = async () => {
  try {
    const res = await axios.get('/api/reagents/dispense-notifications', {
      params: { role: sessionStore.currentRole, user_id: sessionStore.currentUserId },
    })
    dispenseTodoCount.value = Number(res.data?.todo_count || 0)
  } catch {
    dispenseTodoCount.value = 0
  }
}

const handleRequestSubmitted = () => {
  activeTab.value = 'history'
  if (requestList.value && typeof requestList.value.fetchRequests === 'function') {
    requestList.value.fetchRequests()
  }
}

onMounted(async () => {
  applyRouteTab()
  ensureActiveTab()
  await fetchDispenseTodo()
})

watch(() => sessionStore.currentRole, async () => {
  applyRouteTab()
  ensureActiveTab()
  await fetchDispenseTodo()
})

watch(() => sessionStore.currentUserId, fetchDispenseTodo)
watch(tabs, ensureActiveTab)
watch(() => route.query.tab, applyRouteTab)
</script>

<template>
  <div class="reagent-scope space-y-6">
    <div class="mb-2 border-b border-gray-100 pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-gray-900">试剂管理系统</h1>
      <p class="mt-1 text-sm text-gray-500">统一处理申购、到货、库存、领用与流转台账</p>
    </div>

    <div class="apple-segmented flex w-full gap-1.5">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="[
          'apple-segmented-btn-icon w-full justify-center py-2.5 text-sm',
          activeTab === tab.id ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle',
        ]"
      >
        <component :is="tab.icon" :class="['h-4 w-4', activeTab === tab.id ? 'text-blue-600' : 'text-gray-400']" />
        {{ tab.label }}
        <span
          v-if="tab.badge && tab.badge > 0"
          class="inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-red-100 px-1 text-[10px] font-semibold text-red-700"
        >
          {{ tab.badge }}
        </span>
      </button>
    </div>

    <div v-if="activeTab === 'requests'" class="w-full py-2">
      <h3 class="mb-8 text-center text-lg font-bold tracking-tight text-gray-800">请选择试剂申购方式</h3>
      <ReagentRequestWizard @request-submitted="handleRequestSubmitted" />
    </div>

    <div v-if="activeTab === 'history'" class="space-y-4">
      <ReagentRequestList
        ref="requestList"
        :role="sessionStore.currentRole"
        @create-request="activeTab = 'requests'"
      />
    </div>

    <div v-if="activeTab === 'arrival-confirm'" class="space-y-4">
      <ResearcherArrivalList :user-id="sessionStore.currentUserId" />
    </div>

    <div v-if="activeTab === 'unified-inventory'" class="space-y-4">
      <ReagentUnifiedInventory @switch-to-arrival="activeTab = 'arrival-confirm'" />
    </div>

    <div v-if="activeTab === 'batch-import'" class="space-y-4">
      <ProcurementBatchImport />
    </div>

    <div v-if="activeTab === 'receiving'" class="space-y-4">
      <ProcurementReceiving />
    </div>

    <div v-if="activeTab === 'usage'" class="space-y-4">
      <ReagentDispensePanel :role="sessionStore.currentRole" :user-id="sessionStore.currentUserId" />
    </div>
  </div>
</template>
