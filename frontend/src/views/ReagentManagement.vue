<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import axios from 'axios'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Badge from '@/components/ui/Badge.vue'
import { LayoutDashboard, ShoppingCart, List, Layers, BookOpen, FileSpreadsheet, ClipboardCheck, PackageCheck } from 'lucide-vue-next'

import ReagentRequestWizard from '@/components/reagents/ReagentRequestWizard.vue'
import ReagentRequestList from '@/components/reagents/ReagentRequestList.vue'
import ReagentUnifiedInventory from '@/components/reagents/ReagentUnifiedInventory.vue'
import ReagentCatalogManager from '@/components/reagents/ReagentCatalogManager.vue'
import ProcurementBatchImport from '@/components/reagents/ProcurementBatchImport.vue'
import ProcurementReceiving from '@/components/reagents/ProcurementReceiving.vue'
import ReagentDispensePanel from '@/components/reagents/ReagentDispensePanel.vue'
import ResearcherArrivalList from '@/components/reagents/ResearcherArrivalList.vue'
import { formatNumber } from '@/lib/quantity'

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

// Mock UserId Switcher (关联角色方便测试)
// 默认 admin: 1, 采购 manager: 2, leader: 101, 这里根据 currentRole 做个最简单的 mapping
const currentUserId = computed(() => {
    switch(currentRole.value) {
        case 'leader': return 101 // 张明
        case 'procurement': return 2 // 王伟
        case 'researcher': return 1 // 默认测试用 admin 或者是普通研发账号 ID
        default: return 1
    }
})

// Stats Data
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
    return action
}

const getLogVariant = (action: string): any => {
     const map: Record<string, string> = {
        '入库登记': 'success',
        '扫码入库': 'success',
        '空瓶核销': 'warning',
        '库位移动': 'info',
        '状态变更': 'default',
        '变更信息': 'default'
    }
    return map[action] || 'default'
}

const tabs = computed(() => {
  switch (currentRole.value) {
      case 'researcher':
          return [
              { id: 'requests', label: '申购申请', icon: ShoppingCart },
              { id: 'history', label: '申购进度', icon: List },
              { id: 'arrival-confirm', label: '到货确认', icon: PackageCheck },
              { id: 'unified-inventory', label: '库存台账', icon: Layers },
              { id: 'usage', label: '管控领用', icon: ClipboardCheck },
          ]
      case 'procurement':
          return [
              { id: 'dashboard', label: '运营概览', icon: LayoutDashboard },
              { id: 'history', label: '申购台账', icon: List },
              { id: 'batch-import', label: '采购导入', icon: FileSpreadsheet },
              { id: 'receiving', label: '到货台账', icon: PackageCheck },
              { id: 'unified-inventory', label: '库存台账', icon: Layers },
              { id: 'catalog', label: '品目管理', icon: BookOpen },
          ]
      case 'leader':
          return [
              { id: 'dashboard', label: '运营概览', icon: LayoutDashboard },
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
    const last7Days = []
    const counts = []
    for (let i = 6; i >= 0; i--) {
        const d = new Date()
        d.setDate(d.getDate() - i)
        const dateStr = d.toISOString().split('T')[0]
        last7Days.push(dateStr)
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
  <div class="reagent-scope space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-2 border-b border-gray-100 pb-4">
       <div>
         <h1 class="text-2xl font-bold tracking-tight text-gray-900">试剂管理系统</h1>
         <p class="text-sm text-gray-500 mt-1">实验室物资生命周期概览、AI辅助申购与合规盘点看板</p>
       </div>
       <div class="flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 p-2">
           <span class="text-[11px] font-bold text-gray-400 uppercase tracking-widest pl-2">角色切换模拟:</span>
           <select v-model="currentRole" class="h-8 rounded-lg border border-slate-200 bg-white pl-3 pr-8 text-xs font-semibold text-slate-700 shadow-sm">
               <option value="researcher">研发人员</option>
               <option value="procurement">采购人员</option>
               <option value="leader">团队负责人</option>
           </select>
       </div>
    </div>

    <!-- Custom Tabs (Standardized) -->
    <div class="apple-segmented flex w-full gap-1.5">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="[
          'apple-segmented-btn-icon w-full justify-center py-2.5 text-sm',
          activeTab === tab.id
            ? 'apple-segmented-btn-active'
            : 'apple-segmented-btn-idle'
        ]"
      >
        <component :is="tab.icon" :class="['h-4 w-4', activeTab === tab.id ? 'text-blue-600' : 'text-gray-400']" />
        {{ tab.label }}
      </button>
    </div>

    <!-- Content: Dashboard -->
    <div v-if="activeTab === 'dashboard'" class="space-y-4">
         <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
             <Card class="hover:border-blue-200 transition-colors">
                <div class="p-6">
                   <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">总计在册瓶数</div>
                   <div class="text-3xl font-bold mt-2 tracking-tight">{{ stats.total_items }}</div>
                   <div class="text-[10px] mt-2 font-medium" :class="stats.recent_usage_trend?.length > 0 ? 'text-emerald-600' : 'text-gray-400'">
                     <span v-if="stats.recent_usage_trend?.length > 0">📈 近7日累计入库 {{ stats.recent_usage_trend.reduce((s: number, d: any) => s + d.count, 0) }} 瓶</span>
                     <span v-else>包含库房与领用各环节记录</span>
                   </div>
                </div>
            </Card>
             <Card class="hover:border-blue-200 transition-colors">
                <div class="p-6">
                  <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest text-blue-600/60">在库可用状态</div>
                  <div class="text-3xl font-bold text-blue-600 mt-2 tracking-tight">{{ stats.in_storage_items }}</div>
                  <div class="text-[10px] text-blue-500/60 mt-2 font-bold italic">占总数 {{ stats.total_items > 0 ? Math.round(stats.in_storage_items / stats.total_items * 100) : 0 }}% 流转正常</div>
                </div>
            </Card>
             <Card class="hover:border-orange-200 transition-colors">
                <div class="p-6">
                  <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest text-orange-600/60">待办公务流转</div>
                  <div class="text-3xl font-bold text-orange-600 mt-2 tracking-tight">{{ stats.pending_requests }}</div>
                  <div class="text-[10px] mt-2 font-bold" :class="stats.pending_requests > 0 ? 'text-orange-500' : 'text-emerald-500'">
                    {{ stats.pending_requests > 0 ? '⚠️ 需要处理采购/审批' : '✅ 流程全部闭环' }}
                  </div>
                </div>
            </Card>
             <Card class="hover:border-red-200 transition-colors">
                <div class="p-6">
                  <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest" :class="stats.low_stock_alerts > 0 ? 'text-red-600/60' : 'text-gray-400'">库存阈值警报</div>
                  <div class="text-3xl font-bold mt-2 tracking-tight">{{ stats.low_stock_alerts }}</div>
                  <div class="text-[10px] mt-2 font-bold" :class="stats.low_stock_alerts > 0 ? 'text-red-500' : 'text-emerald-500'">
                    {{ stats.low_stock_alerts > 0 ? '🔴 ' + stats.low_stock_alerts + ' 类物资低于设定阈值' : '✅ 目前货源充足' }}
                  </div>
                </div>
            </Card>
         </div>

         <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <Card class="hover:border-amber-200 transition-colors">
              <div class="p-5">
                <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">暂存待入库</div>
                <div class="text-2xl font-bold mt-2 text-amber-600">{{ stats.pending_checkin_items }}</div>
                <div class="text-[11px] text-gray-500 mt-1">已到货但仍在暂存区的试剂瓶数</div>
              </div>
            </Card>
            <Card class="hover:border-sky-200 transition-colors">
              <div class="p-5">
                <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">在途待收货</div>
                <div class="text-2xl font-bold mt-2 text-sky-600">{{ stats.pending_receive_lines }}</div>
                <div class="text-[11px] text-gray-500 mt-1">采购明细中未完全收货的行数</div>
              </div>
            </Card>
            <Card class="hover:border-red-200 transition-colors">
              <div class="p-5">
                <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">管控试剂在库</div>
                <div class="text-2xl font-bold mt-2 text-red-600">{{ stats.controlled_in_storage }}</div>
                <div class="text-[11px] text-gray-500 mt-1">需重点监管的在库瓶数</div>
              </div>
            </Card>
            <Card class="hover:border-orange-200 transition-colors">
              <div class="p-5">
                <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">30天内到期</div>
                <div class="text-2xl font-bold mt-2 text-orange-600">{{ stats.near_expiry_30d }}</div>
                <div class="text-[11px] text-gray-500 mt-1">近一个月到期的在库试剂</div>
              </div>
            </Card>
            <Card class="hover:border-emerald-200 transition-colors">
              <div class="p-5">
                <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">活跃品目数</div>
                <div class="text-2xl font-bold mt-2 text-emerald-600">{{ stats.active_catalogs }}</div>
                <div class="text-[11px] text-gray-500 mt-1">当前仍有库存流转的品目数</div>
              </div>
            </Card>
            <Card class="hover:border-indigo-200 transition-colors">
              <div class="p-5">
                <div class="text-[11px] font-bold text-gray-400 uppercase tracking-widest">近7日消耗量</div>
                <div class="text-2xl font-bold mt-2 text-indigo-600">{{ formatNumber(stats.consumed_volume_7d) }}</div>
                <div class="text-[11px] text-gray-500 mt-1">按日志累计的消耗数值（含耗尽）</div>
              </div>
            </Card>
         </div>
         
         <!-- Charts Row -->
         <div v-if="currentRole === 'leader' || currentRole === 'procurement'" class="grid gap-4 md:grid-cols-2">
             <Card class="col-span-1 shadow-sm h-80">
                 <div class="p-6 h-full">
                     <v-chart class="h-full w-full" :option="categoryOption" autoresize />
                 </div>
             </Card>
             <Card class="col-span-1 shadow-sm h-80">
                 <div class="p-6 h-full">
                     <v-chart class="h-full w-full" :option="trendOption" autoresize />
                 </div>
             </Card>
         </div>

         <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
             <Card class="col-span-4 shadow-sm overflow-hidden">
                 <div class="p-6">
                     <h3 class="font-bold text-lg mb-4 text-gray-900 tracking-tight">近期操作流水</h3>
                     <div v-if="stats.recent_logs.length === 0" class="text-sm text-gray-500 text-center py-12 bg-gray-50/50 rounded-2xl border-2 border-dashed border-gray-200/50">
                         暂无任何实验或入库记录。
                     </div>
                     <div v-else class="space-y-1">
                         <div v-for="log in stats.recent_logs" :key="log.id" class="flex items-start justify-between py-3.5 hover:bg-gray-50/60 transition-all rounded-xl px-3 group border-b border-gray-50 last:border-0 -mx-3">
                             <div class="flex-1">
                                 <div class="flex items-center gap-2.5 mb-1.5">
                                     <Badge :variant="getLogVariant(log.action)">
                                         {{ getActionLabel(log.action) }}
                                     </Badge>
                                     <span class="font-bold text-sm text-gray-900 leading-none group-hover:text-blue-700 transition-colors">{{ log.reagent_item?.reagent_catalog?.name || '未知品目' }}</span>
                                 </div>
                                 <div class="text-[11px] text-gray-500 font-medium flex gap-4">
                                     <span>👤 {{ log.user?.real_name || 'System' }}</span>
                                     <span>数量: <span class="font-bold text-blue-600 font-mono">{{ formatNumber(log.quantity) }}</span></span>
                                     <span v-if="log.remarks" class="italic opacity-60">"{{ log.remarks }}"</span>
                                 </div>
                             </div>
                             <div class="text-[11px] text-gray-400 whitespace-nowrap pt-1 tabular-nums font-mono">
                                 {{ new Date(log.created_at).toLocaleString('zh-CN', { hour12: false }).replace(/\//g, '-') }}
                             </div>
                         </div>
                     </div>
                 </div>
             </Card>
             <Card class="col-span-3 shadow-sm overflow-hidden">
                 <div class="p-6">
                     <h3 class="font-bold text-lg mb-4 flex items-center justify-between">
                         <span class="flex items-center gap-2">系统智能告警 <span v-if="stats.alerts.length > 0" class="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1.5 text-[11px] font-bold leading-none text-white">{{ stats.alerts.length }}</span></span>
                     </h3>
                      <div class="space-y-3 max-h-[280px] overflow-y-auto pr-2" style="scrollbar-width: thin; scrollbar-color: #cbd5e1 transparent;">
                          <div v-if="stats.alerts.length === 0" class="text-sm text-gray-400 text-center py-16 bg-gray-50/50 rounded-2xl border border-dashed border-gray-200">
                               🎉 环境处于安全且稳定的库存状态。
                          </div>
                          <div v-for="alert in stats.alerts" :key="alert.ID" class="flex gap-4 items-start bg-white p-4 rounded-xl border border-gray-100 shadow-sm hover:border-red-200 transition-all group">
                              <div class="h-10 w-10 shrink-0 bg-red-50 rounded-full flex items-center justify-center text-red-500 group-hover:bg-red-500 group-hover:text-white transition-colors duration-300">
                                ⚠️
                              </div>
                              <div class="flex-1">
                                <div class="text-sm font-bold text-gray-900 leading-tight">
                                  "{{ alert.Name }}" 已低于预警水位
                                </div>
                                <div class="text-[11px] text-gray-500 mt-2 font-mono flex items-center gap-3">
                                  <span>剩余: <span class="text-red-600 font-bold">{{ alert.Count }}</span></span>
                                  <span class="text-gray-200">|</span>
                                  <span>警报阈值: {{ alert.AlertThreshold }}</span>
                                </div>
                              </div>
                          </div>
                      </div>
                 </div>
             </Card>
         </div>
    </div>

    <!-- Content: Requests -->
    <div v-if="activeTab === 'requests'" class="w-full py-2">
         <h3 class="font-bold text-lg mb-8 text-center text-gray-800 tracking-tight">请选择试剂申购方式</h3>
         <ReagentRequestWizard @request-submitted="handleRequestSubmitted" />
    </div>

    <!-- Content: History -->
     <div v-if="activeTab === 'history'" class="space-y-4">
          <div class="flex items-center justify-between mb-2">
              <h3 class="font-bold text-lg text-gray-900 tracking-tight">{{ currentRole === 'researcher' ? '我的申购台账' : '申购台账' }}</h3>
              <Button v-if="currentRole === 'researcher'" @click="activeTab = 'requests'" variant="primary" size="sm" class="shadow-sm">
                  + 提交新申购
              </Button>
          </div>
          <ReagentRequestList ref="requestList" :role="currentRole" />
     </div>

    <div v-if="activeTab === 'arrival-confirm'" class="space-y-4">
        <ResearcherArrivalList :userId="currentUserId" />
    </div>

    <div v-if="activeTab === 'unified-inventory'" class="space-y-4">
        <ReagentUnifiedInventory @switch-to-arrival="activeTab = 'arrival-confirm'" />
    </div>

    <!-- Content: Catalog Management -->
    <div v-if="activeTab === 'catalog'" class="space-y-4">
        <div class="flex items-center justify-between mb-2 px-1">
            <div>
                <h3 class="font-bold text-lg text-gray-900 tracking-tight">试剂品目库管理</h3>
                <p class="text-sm text-gray-500 mt-1">标准化化学品分类、危险等级及储存策略维护</p>
            </div>
        </div>
        <ReagentCatalogManager />
    </div>

    <div v-if="activeTab === 'batch-import'" class="space-y-4">
        <ProcurementBatchImport />
    </div>

    <div v-if="activeTab === 'receiving'" class="space-y-4">
        <ProcurementReceiving />
    </div>

    <div v-if="activeTab === 'usage'" class="space-y-4">
        <ReagentDispensePanel role="researcher" />
    </div>

    <div v-if="activeTab === 'dispense-approve'" class="space-y-4">
        <ReagentDispensePanel role="leader" />
    </div>

  </div>
</template>
