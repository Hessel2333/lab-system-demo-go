<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import axios from 'axios'
import Card from '@/components/ui/Card.vue'
import { LayoutDashboard, ShoppingCart, List, ScanLine, Layers, BookOpen, FileSpreadsheet, ClipboardCheck, PackageCheck } from 'lucide-vue-next'

import ReagentScanner from '@/components/reagents/ReagentScanner.vue'
import ReagentRequestWizard from '@/components/reagents/ReagentRequestWizard.vue'
import ReagentRequestList from '@/components/reagents/ReagentRequestList.vue'
import ReagentUnifiedInventory from '@/components/reagents/ReagentUnifiedInventory.vue'
import ReagentCatalogManager from '@/components/reagents/ReagentCatalogManager.vue'
import ProcurementBatchImport from '@/components/reagents/ProcurementBatchImport.vue'
import ProcurementReceiving from '@/components/reagents/ProcurementReceiving.vue'
import ReagentDispensePanel from '@/components/reagents/ReagentDispensePanel.vue'

// ECharts
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, LineChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer,
  PieChart,
  LineChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

const activeTab = ref('requests')
const requestList = ref<InstanceType<typeof ReagentRequestList> | null>(null)

// Mock RBAC Role Switcher
const currentRole = ref<'researcher' | 'procurement' | 'leader'>('researcher')

// Stats Data
const stats = ref({
    total_items: 0,
    in_storage_items: 0,
    pending_requests: 0,
    low_stock_alerts: 0,
    recent_logs: [] as any[],
    alerts: [] as any[],
    category_distribution: [] as { category: string, count: number }[],
    recent_usage_trend: [] as { date: string, count: number }[]
})

const fetchStats = async () => {
    try {
        const response = await axios.get('/api/reagents/stats')
        stats.value = response.data
    } catch (error) {
        console.error('Failed to fetch reagent stats', error)
    }
}

onMounted(() => {
    fetchStats()
})

watch(activeTab, (newVal) => {
    if (newVal === 'dashboard') {
        fetchStats()
    }
})

watch(currentRole, () => {
    activeTab.value = tabs.value[0]?.id || 'requests'
    if (activeTab.value === 'dashboard') fetchStats()
})

const handleRequestSubmitted = () => {
  activeTab.value = 'history'
  if (requestList.value) {
      if (typeof (requestList.value as any).fetchRequests === 'function') {
          (requestList.value as any).fetchRequests()
      }
  }
}

const getActionLabel = (action: string) => {
    // The backend now returns Chinese directly, so we can just return it.
    return action
}

const getActionColor = (action: string) => {
     const map: Record<string, string> = {
        '入库登记': 'text-green-600 bg-green-50',
        '扫码入库': 'text-green-600 bg-green-50',
        '空瓶核销': 'text-orange-600 bg-orange-50',
        '库位移动': 'text-blue-600 bg-blue-50',
        '状态变更': 'text-gray-600 bg-gray-50',
        '变更信息': 'text-gray-600 bg-gray-50'
    }
    return map[action] || 'text-gray-600 bg-gray-50'
}

const tabs = computed(() => {
  switch (currentRole.value) {
      case 'researcher':
          return [
              { id: 'requests', label: '试剂申购', icon: ShoppingCart },
              { id: 'history', label: '我的申购', icon: List },
              { id: 'dispense', label: '我的领用', icon: ClipboardCheck },
              { id: 'unified-inventory', label: '库存台账', icon: Layers },
              { id: 'scanner', label: '扫码领用', icon: ScanLine },
          ]
      case 'procurement':
          return [
              { id: 'dashboard', label: '宏观大盘', icon: LayoutDashboard },
              { id: 'history', label: '采购工作池', icon: List },
              { id: 'batch-import', label: '易派客明细分发', icon: FileSpreadsheet },
              { id: 'receiving', label: '暂存区实物点收', icon: PackageCheck },
              { id: 'unified-inventory', label: '库存台账', icon: Layers },
              { id: 'catalog', label: '品目管理', icon: BookOpen },
          ]
      case 'leader':
          return [
              { id: 'dashboard', label: '概览仪表盘(报表)', icon: LayoutDashboard },
              { id: 'dispense-approve', label: '领用审批', icon: ClipboardCheck },
              { id: 'unified-inventory', label: '库存台账', icon: Layers },
          ]
      default:
          return []
  }
})

// ECharts Computed Options
const categoryOption = computed(() => {
    return {
        title: { text: '试剂分类分布', left: 'center', textStyle: { fontSize: 15, fontWeight: 'normal' } },
        tooltip: { trigger: 'item' },
        legend: { orient: 'horizontal', bottom: 'bottom' },
        series: [
            {
                type: 'pie',
                radius: ['40%', '70%'],
                avoidLabelOverlap: false,
                itemStyle: { borderRadius: 5, borderColor: '#fff', borderWidth: 2 },
                label: { show: false, position: 'center' },
                emphasis: { label: { show: true, fontSize: '18', fontWeight: 'bold' } },
                labelLine: { show: false },
                data: stats.value.category_distribution.map(s => ({ value: s.count, name: s.category }))
            }
        ]
    }
})

const trendOption = computed(() => {
    // Fill last 7 days dynamically
    const last7Days = []
    const counts = []
    for (let i = 6; i >= 0; i--) {
        const d = new Date()
        d.setDate(d.getDate() - i)
        const dateStr = d.toISOString().split('T')[0]
        last7Days.push(dateStr)
        
        // Find if we have data for this date
        const stat = stats.value.recent_usage_trend?.find(t => t.date === dateStr)
        counts.push(stat ? stat.count : 0)
    }

    return {
        title: { text: '近7天物理消耗趋势 (单瓶)', left: 'center', textStyle: { fontSize: 15, fontWeight: 'normal' } },
        tooltip: { trigger: 'axis' },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', boundaryGap: false, data: last7Days },
        yAxis: { type: 'value', minInterval: 1 },
        series: [
            {
                name: '空瓶核销数',
                type: 'line',
                data: counts,
                smooth: true,
                areaStyle: {
                    color: {
                        type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                        colorStops: [{ offset: 0, color: 'rgba(59, 130, 246, 0.5)' }, { offset: 1, color: 'rgba(59, 130, 246, 0.0)' }]
                    }
                },
                itemStyle: { color: '#3b82f6' }
            }
        ]
    }
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-2 border-b pb-4">
       <div>
         <h1 class="text-2xl font-bold tracking-tight text-gray-900">试剂管理模块</h1>
         <p class="text-sm text-gray-500">管理实验室试剂全生命周期、AI申购及安全合规库存。</p>
       </div>
       <div class="flex items-center gap-3 bg-blue-50/50 p-2.5 rounded-lg border border-blue-100 shadow-sm">
           <span class="text-sm font-medium text-blue-800">当前视角模拟:</span>
           <select v-model="currentRole" class="text-sm border-gray-300 rounded-md shadow-sm focus:border-blue-500 focus:ring-blue-500 pl-3 pr-8 py-1.5 font-medium">
               <option value="researcher">👨‍🔬 研发人员</option>
               <option value="procurement">🛒 采购人员</option>
               <option value="leader">📊 领导(决策)</option>
           </select>
       </div>
    </div>

    <!-- Custom Tabs -->
    <div class="flex space-x-1 rounded-xl bg-gray-100/80 p-1">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="[
          'flex items-center gap-2 w-full justify-center rounded-lg py-2.5 text-sm font-medium leading-5 ring-white ring-opacity-60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2 transition-all',
          activeTab === tab.id
            ? 'bg-white text-blue-700 shadow'
            : 'text-gray-600 hover:bg-white/[0.12] hover:text-gray-800'
        ]"
      >
        <component :is="tab.icon" class="h-4 w-4" />
        {{ tab.label }}
      </button>
    </div>

    <!-- Content: Dashboard -->
    <div v-if="activeTab === 'dashboard'" class="space-y-4">
         <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
             <Card>
                <div class="p-6">
                   <div class="text-sm font-medium text-gray-500">总库存件数</div>
                   <div class="text-2xl font-bold mt-1">{{ stats.total_items }}</div>
                   <div class="text-xs mt-1" :class="stats.recent_usage_trend?.length > 0 ? 'text-emerald-600' : 'text-gray-400'">
                     <span v-if="stats.recent_usage_trend?.length > 0">📈 近7日入库 {{ stats.recent_usage_trend.reduce((s: number, d: any) => s + d.count, 0) }} 件</span>
                     <span v-else>不含已耗尽试剂</span>
                   </div>
                </div>
            </Card>
             <Card>
                <div class="p-6">
                  <div class="text-sm font-medium text-gray-500">正常在库</div>
                  <div class="text-2xl font-bold text-blue-600 mt-1">{{ stats.in_storage_items }}</div>
                  <div class="text-xs text-blue-500 mt-1">占总库存 {{ stats.total_items > 0 ? Math.round(stats.in_storage_items / stats.total_items * 100) : 0 }}%</div>
                </div>
            </Card>
             <Card>
                <div class="p-6">
                  <div class="text-sm font-medium text-gray-500">待处理申购</div>
                  <div class="text-2xl font-bold text-orange-600 mt-1">{{ stats.pending_requests }}</div>
                  <div class="text-xs mt-1" :class="stats.pending_requests > 0 ? 'text-orange-500' : 'text-emerald-500'">
                    {{ stats.pending_requests > 0 ? '⚠️ 需要采购审批' : '✅ 暂无待办' }}
                  </div>
                </div>
            </Card>
             <Card>
                <div class="p-6">
                  <div class="text-sm font-medium" :class="stats.low_stock_alerts > 0 ? 'text-red-600' : 'text-gray-500'">库存预警</div>
                  <div class="text-2xl font-bold mt-1">{{ stats.low_stock_alerts }}</div>
                  <div class="text-xs mt-1" :class="stats.low_stock_alerts > 0 ? 'text-red-500' : 'text-emerald-500'">
                    {{ stats.low_stock_alerts > 0 ? '🔴 ' + stats.low_stock_alerts + ' 个品类低于阈值' : '✅ 库存充足' }}
                  </div>
                </div>
            </Card>
         </div>
         
         <!-- Charts Row (Only visible for Leader or if data exists) -->
         <div v-if="currentRole === 'leader' || currentRole === 'procurement'" class="grid gap-4 md:grid-cols-2">
             <Card class="col-span-1">
                 <div class="p-6 h-80">
                     <v-chart class="h-full w-full" :option="categoryOption" autoresize />
                 </div>
             </Card>
             <Card class="col-span-1">
                 <div class="p-6 h-80">
                     <v-chart class="h-full w-full" :option="trendOption" autoresize />
                 </div>
             </Card>
         </div>

         <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
             <Card class="col-span-4">
                 <div class="p-6">
                     <h3 class="font-semibold text-lg mb-4">近期操作记录</h3>
                     <div v-if="stats.recent_logs.length === 0" class="text-sm text-gray-500 text-center py-8">
                         暂无近期活动记录。
                     </div>
                     <div v-else class="space-y-4">
                         <div v-for="log in stats.recent_logs" :key="log.id" class="flex items-start justify-between border-b pb-3 last:border-0">
                             <div>
                                 <div class="flex items-center gap-2 mb-1">
                                     <span :class="['px-2 py-0.5 rounded text-xs font-medium', getActionColor(log.action)]">
                                         {{ getActionLabel(log.action) }}
                                     </span>
                                     <span class="font-medium text-sm">{{ log.reagent_item?.reagent_catalog?.name || '未知试剂' }}</span>
                                 </div>
                                 <div class="text-xs text-gray-500 mt-1">
                                     操作人: {{ log.user?.real_name || 'System' }} | 数量变动: {{ log.quantity }} | 备注: {{ log.remarks || '无' }}
                                 </div>
                             </div>
                             <div class="text-xs text-gray-400 whitespace-nowrap">
                                 {{ new Date(log.created_at).toLocaleString('zh-CN') }}
                             </div>
                         </div>
                     </div>
                 </div>
             </Card>
             <Card class="col-span-3">
                 <div class="p-5">
                     <h3 class="font-semibold text-lg mb-3 flex items-center justify-between">
                         <span class="flex items-center gap-2">系统消息 <span v-if="stats.alerts.length > 0" class="bg-red-100 text-red-600 text-xs px-2 py-0.5 rounded-full">{{ stats.alerts.length }}</span></span>
                     </h3>
                      <div class="space-y-3 max-h-[280px] overflow-y-auto pr-2" style="scrollbar-width: thin; scrollbar-color: #cbd5e1 transparent;">
                          <div v-if="stats.alerts.length === 0" class="text-sm text-gray-400 text-center py-10 bg-gray-50 rounded-lg">
                              ✅ 暂无需要处理的系统告警。
                          </div>
                          <div v-for="alert in stats.alerts" :key="alert.ID" class="flex gap-3 items-start border-l-4 border-l-red-500 bg-red-50/50 p-3 rounded-r-md border border-y-red-100 border-r-red-100">
                              <span class="text-lg leading-none mt-0.5" aria-hidden="true">⚠️</span>
                              <div class="flex-1">
                                <div class="text-[13px] font-medium text-red-900 leading-tight">
                                  试剂 <span class="font-bold">"{{ alert.Name }}"</span> 的库存低于预警线
                                </div>
                                <div class="text-xs text-red-700/80 mt-1 font-mono">
                                  剩余: <span class="font-bold text-red-600">{{ alert.Count }}</span> | 阈值: {{ alert.AlertThreshold }}
                                </div>
                              </div>
                          </div>
                      </div>
                 </div>
             </Card>
         </div>
    </div>

    <!-- Content: Requests -->
    <div v-if="activeTab === 'requests'" class="max-w-4xl mx-auto py-2">
         <h3 class="font-medium text-lg mb-6 text-center text-gray-700">请选择申购方式</h3>
         <ReagentRequestWizard @request-submitted="handleRequestSubmitted" />
    </div>

    <!-- Content: History -->
    <div v-if="activeTab === 'history'" class="space-y-4">
         <div class="flex items-center justify-between mb-2">
             <h3 class="font-medium text-lg">{{ currentRole === 'researcher' ? '我的申购记录' : '全局采购任务池' }}</h3>
             <button v-if="currentRole === 'researcher'" @click="activeTab = 'requests'" class="text-sm bg-blue-50 text-blue-600 px-4 py-2 rounded-md hover:bg-blue-100 transition-colors">
                 + 提交新申购
             </button>
         </div>
         <ReagentRequestList ref="requestList" :role="currentRole" />
    </div>

    <!-- Content: Unified Inventory -->
    <div v-if="activeTab === 'unified-inventory'" class="space-y-4">
        <ReagentUnifiedInventory />
    </div>

    <!-- Content: Catalog Management -->
    <div v-if="activeTab === 'catalog'" class="space-y-4">
        <div class="flex items-center justify-between mb-2">
            <div>
                <h3 class="font-medium text-lg">品目管理</h3>
                <p class="text-sm text-gray-500">维护试剂品目数据库，管理化学品分类标签、别称和储存条件</p>
            </div>
        </div>
        <ReagentCatalogManager />
    </div>

    <!-- Content: BPM-B Batch Import -->
    <div v-if="activeTab === 'batch-import'" class="space-y-4">
        <ProcurementBatchImport />
    </div>

    <!-- Content: BPM-B Batch Receiving -->
    <div v-if="activeTab === 'receiving'" class="space-y-4">
        <ProcurementReceiving />
    </div>

    <!-- Content: Scanner -->
    <div v-if="activeTab === 'scanner'" class="max-w-md mx-auto">
        <ReagentScanner />
    </div>

    <!-- Content: Dispense (Researcher) -->
    <div v-if="activeTab === 'dispense'" class="space-y-4">
        <ReagentDispensePanel role="researcher" />
    </div>

    <!-- Content: Dispense Approve (Leader) -->
    <div v-if="activeTab === 'dispense-approve'" class="space-y-4">
        <ReagentDispensePanel role="leader" />
    </div>

  </div>
</template>
