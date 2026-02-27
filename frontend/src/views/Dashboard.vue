<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, watch } from 'vue'
import axios from 'axios'
import { RouterLink } from 'vue-router'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import { useSessionStore } from '@/stores/session'
import { ArrowRight, ClipboardCheck, FileSpreadsheet, PackageCheck, ShoppingCart, TestTube } from 'lucide-vue-next'

const ReagentOpsDashboard = defineAsyncComponent(() => import('@/components/reagents/ReagentOpsDashboard.vue'))

const sessionStore = useSessionStore()
const loading = ref(false)
const loadError = ref('')

const stats = ref({
  total_items: 0,
  in_storage_items: 0,
  pending_requests: 0,
  low_stock_alerts: 0,
  pending_checkin_items: 0,
  pending_receive_lines: 0,
  controlled_in_storage: 0,
  near_expiry_30d: 0,
  active_catalogs: 0,
  consumed_volume_7d: 0,
  recent_logs: [] as any[],
  alerts: [] as any[],
  category_distribution: [] as { category: string; count: number }[],
  recent_usage_trend: [] as { date: string; count: number }[],
})

const roleLabel = computed(() => sessionStore.currentRoleLabel)

const quickActions = computed(() => {
  if (sessionStore.currentRole === 'researcher') {
    return [
      { title: '提交申购', desc: '发起试剂申购并跟踪审批状态', to: { path: '/reagents', query: { tab: 'requests' } }, tab: '申购申请', icon: ShoppingCart },
      { title: '到货确认', desc: '查看到货明细并完成入库定位', to: { path: '/reagents', query: { tab: 'arrival-confirm' } }, tab: '到货确认', icon: PackageCheck },
      { title: '领用申请', desc: '在管控台账中快捷发起领用流程', to: { path: '/reagents', query: { tab: 'usage' } }, tab: '领用台账', icon: ClipboardCheck },
    ]
  }
  if (sessionStore.currentRole === 'procurement') {
    return [
      { title: '采购导入', desc: '导入采购明细并完成申购匹配', to: { path: '/reagents', query: { tab: 'batch-import' } }, tab: '采购导入', icon: FileSpreadsheet },
      { title: '到货台账', desc: '统一处理到货确认、赋码与入库', to: { path: '/reagents', query: { tab: 'receiving' } }, tab: '到货台账', icon: PackageCheck },
      { title: '库存台账', desc: '查看全局库存并处理异常状态', to: { path: '/reagents', query: { tab: 'unified-inventory' } }, tab: '库存台账', icon: TestTube },
    ]
  }
  return [
    { title: '领用审批', desc: '处理管控领用审批与通知闭环', to: { path: '/reagents', query: { tab: 'usage' } }, tab: '领用台账', icon: ClipboardCheck },
    { title: '申购跟踪', desc: '统一查看申购与到货流转进度', to: { path: '/reagents', query: { tab: 'history' } }, tab: '申购台账', icon: ShoppingCart },
    { title: '基础数据', desc: '维护品目与基础主数据规范', to: '/master-data', tab: '基础数据', icon: TestTube },
  ]
})

const fetchStats = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const response = await axios.get('/api/reagents/stats', {
      params: { role: sessionStore.currentRole },
    })
    stats.value = response.data
  } catch (error) {
    console.error('Failed to fetch overview stats', error)
    loadError.value = '运营概览数据加载失败，请稍后重试。'
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
watch(() => sessionStore.currentRole, fetchStats)
watch(() => sessionStore.currentUserId, fetchStats)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 rounded-2xl border border-gray-200 bg-white p-6 shadow-sm md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-gray-900">系统概览</h1>
        <p class="mt-1 text-sm text-gray-500">当前为 {{ roleLabel }} 视角，关键待办与运营指标已按角色聚焦。</p>
      </div>
      <Badge variant="info">
        当前角色：{{ roleLabel }}
      </Badge>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <RouterLink
        v-for="action in quickActions"
        :key="action.title"
        :to="action.to"
        class="group block rounded-2xl border border-gray-200 bg-white p-5 shadow-sm transition-all hover:-translate-y-0.5 hover:border-blue-300 hover:shadow-md"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="space-y-1">
            <h3 class="text-base font-semibold text-gray-900">{{ action.title }}</h3>
            <p class="text-sm text-gray-500">{{ action.desc }}</p>
            <p class="pt-1 text-xs font-medium text-blue-600">入口：{{ action.tab }}</p>
          </div>
          <component :is="action.icon" class="h-5 w-5 shrink-0 text-blue-500" />
        </div>
        <div class="mt-4 flex items-center text-xs font-semibold text-blue-600">
          进入
          <ArrowRight class="ml-1 h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
        </div>
      </RouterLink>
    </div>

    <Card class="overflow-hidden border-gray-200 shadow-sm">
      <div class="border-b border-gray-100 px-5 py-4">
        <h2 class="text-lg font-semibold text-gray-900">试剂运营概览</h2>
        <p class="mt-1 text-sm text-gray-500">统一观察库存、告警、流转与近期操作，支持跨角色协同。</p>
      </div>
      <div class="p-4">
        <div v-if="loading" class="rounded-xl border border-dashed border-gray-200 bg-gray-50 px-4 py-10 text-center text-sm text-gray-500">
          正在加载运营数据...
        </div>
        <div v-else-if="loadError" class="rounded-xl border border-red-100 bg-red-50 px-4 py-10 text-center text-sm text-red-600">
          {{ loadError }}
        </div>
        <ReagentOpsDashboard v-else :stats="stats" :current-role="sessionStore.currentRole" />
      </div>
    </Card>
  </div>
</template>
