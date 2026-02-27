<script setup lang="ts">
import { computed } from 'vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import { formatNumber } from '@/lib/quantity'

import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, LineChart, BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([
  CanvasRenderer,
  PieChart,
  LineChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
])

type AppRole = 'researcher' | 'procurement' | 'leader'
type Priority = 'high' | 'medium' | 'low'

interface InsightCard {
  key: string
  title: string
  value: number
  hint: string
  valueClass: string
}

interface TodoItem {
  key: string
  title: string
  count: number
  hint: string
  priority: Priority
}

const props = defineProps<{
  stats: any
  currentRole: AppRole
}>()

const roleFocus = computed<Record<string, any>>(() => props.stats?.role_focus || {})

const toNumber = (value: any): number => {
  const n = Number(value)
  return Number.isFinite(n) ? n : 0
}

const normalizeChartValues = (values: number[]) => {
  if (values.some((v) => v > 0)) return values
  return values.map(() => 0)
}

const rolePrimaryCards = computed<InsightCard[]>(() => {
  const focus = roleFocus.value
  if (props.currentRole === 'researcher') {
    return [
      {
        key: 'r-open-request',
        title: '我的待处理申购',
        value: toNumber(focus.my_open_requests),
        hint: '待审批 / 待采购 / 已接单',
        valueClass: 'text-blue-600',
      },
      {
        key: 'r-pending-checkin',
        title: '我的待入库',
        value: toNumber(focus.my_pending_checkin),
        hint: '到货后待选择实验室与试剂柜',
        valueClass: 'text-amber-600',
      },
      {
        key: 'r-pending-dispense',
        title: '我的领用待办',
        value: toNumber(focus.my_pending_dispense),
        hint: '待审批或待双签确认',
        valueClass: 'text-violet-600',
      },
    ]
  }
  if (props.currentRole === 'procurement') {
    return [
      {
        key: 'p-pending-procurement',
        title: '待采购申购单',
        value: toNumber(focus.pending_procurement_requests),
        hint: '需要接单并形成订单',
        valueClass: 'text-orange-600',
      },
      {
        key: 'p-receiving-todo',
        title: '收货待处理行',
        value: toNumber(focus.receiving_todo_lines),
        hint: '可进入到货台账点验赋码',
        valueClass: 'text-blue-600',
      },
      {
        key: 'p-unmatched-import',
        title: '导入待匹配行',
        value: toNumber(focus.unmatched_import_lines),
        hint: '待绑定申购需求或指派研发',
        valueClass: 'text-rose-600',
      },
    ]
  }
  return [
    {
      key: 'l-pending-request-approval',
      title: '待审批申购',
      value: toNumber(focus.pending_leader_approvals),
      hint: '管控试剂申购待决策',
      valueClass: 'text-orange-600',
    },
    {
      key: 'l-pending-dispense-approval',
      title: '待审批领用',
      value: toNumber(focus.pending_dispense_approvals),
      hint: '领用流程待团队长审批',
      valueClass: 'text-blue-600',
    },
    {
      key: 'l-dual-sign-progress',
      title: '双签进行中',
      value: toNumber(focus.dual_sign_in_progress),
      hint: '已进入 A/B 持有人确认',
      valueClass: 'text-violet-600',
    },
  ]
})

const roleTodos = computed<TodoItem[]>(() => {
  const focus = roleFocus.value
  const makePriority = (value: number): Priority => {
    if (value >= 10) return 'high'
    if (value >= 3) return 'medium'
    return 'low'
  }

  if (props.currentRole === 'researcher') {
    const items = [
      {
        key: 'todo-r-checkin',
        title: '到货入库',
        count: toNumber(focus.my_pending_checkin),
        hint: '优先完成到货试剂入库定位',
      },
      {
        key: 'todo-r-dispense',
        title: '领用流程跟进',
        count: toNumber(focus.my_pending_dispense),
        hint: '补充用途说明或等待双签确认',
      },
      {
        key: 'todo-r-expiry',
        title: '临期处理',
        count: toNumber(focus.my_near_expiry_30d),
        hint: '30天内到期，建议优先使用',
      },
    ]
    return items.map((item) => ({ ...item, priority: makePriority(item.count) }))
  }

  if (props.currentRole === 'procurement') {
    const items = [
      {
        key: 'todo-p-procurement',
        title: '采购接单',
        count: toNumber(focus.pending_procurement_requests),
        hint: '待采购申请需尽快完成接单',
      },
      {
        key: 'todo-p-receiving',
        title: '到货点验',
        count: toNumber(focus.receiving_todo_lines),
        hint: '待收货明细建议当日完成赋码',
      },
      {
        key: 'todo-p-import',
        title: '导入匹配',
        count: toNumber(focus.unmatched_import_lines),
        hint: '未匹配项目会阻塞批次确认',
      },
    ]
    return items.map((item) => ({ ...item, priority: makePriority(item.count) }))
  }

  const items = [
    {
      key: 'todo-l-request',
      title: '申购审批',
      count: toNumber(focus.pending_leader_approvals),
      hint: '涉及管控试剂，建议优先审批',
    },
    {
      key: 'todo-l-dispense',
      title: '领用审批',
      count: toNumber(focus.pending_dispense_approvals),
      hint: '审批结果会影响实验进度',
    },
    {
      key: 'todo-l-risk',
      title: '库存风险',
      count: toNumber(focus.low_stock_alerts),
      hint: '低库存品目建议触发补货动作',
    },
  ]
  return items.map((item) => ({ ...item, priority: makePriority(item.count) }))
})

const flowChartOption = computed(() => {
  if (props.currentRole === 'researcher') {
    const categories = ['待处理申购', '待入库', '领用待办', '临期处理']
    const values = normalizeChartValues([
      toNumber(roleFocus.value.my_open_requests),
      toNumber(roleFocus.value.my_pending_checkin),
      toNumber(roleFocus.value.my_pending_dispense),
      toNumber(roleFocus.value.my_near_expiry_30d),
    ])
    return {
      title: { text: '我的流程负载', left: 'center', textStyle: { fontSize: 14, fontWeight: 500 } },
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: '3%', right: '4%', bottom: '6%', top: '18%', containLabel: true },
      xAxis: { type: 'value', minInterval: 1 },
      yAxis: { type: 'category', data: categories },
      series: [{ type: 'bar', data: values, barWidth: 18, itemStyle: { color: '#3b82f6', borderRadius: [0, 6, 6, 0] } }],
    }
  }

  if (props.currentRole === 'procurement') {
    const categories = ['待采购', '已接单待到货', '收货待处理', '导入待匹配']
    const values = normalizeChartValues([
      toNumber(roleFocus.value.pending_procurement_requests),
      toNumber(roleFocus.value.ordered_waiting_arrival),
      toNumber(roleFocus.value.receiving_todo_lines),
      toNumber(roleFocus.value.unmatched_import_lines),
    ])
    return {
      title: { text: '采购执行负载', left: 'center', textStyle: { fontSize: 14, fontWeight: 500 } },
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      grid: { left: '3%', right: '4%', bottom: '6%', top: '18%', containLabel: true },
      xAxis: { type: 'value', minInterval: 1 },
      yAxis: { type: 'category', data: categories },
      series: [{ type: 'bar', data: values, barWidth: 18, itemStyle: { color: '#0ea5e9', borderRadius: [0, 6, 6, 0] } }],
    }
  }

  const categories = ['申购待审批', '领用待审批', '双签进行中', '近7天驳回']
  const values = normalizeChartValues([
    toNumber(roleFocus.value.pending_leader_approvals),
    toNumber(roleFocus.value.pending_dispense_approvals),
    toNumber(roleFocus.value.dual_sign_in_progress),
    toNumber(roleFocus.value.rejected_dispense_7d),
  ])
  return {
    title: { text: '审批治理负载', left: 'center', textStyle: { fontSize: 14, fontWeight: 500 } },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '4%', bottom: '6%', top: '18%', containLabel: true },
    xAxis: { type: 'value', minInterval: 1 },
    yAxis: { type: 'category', data: categories },
    series: [{ type: 'bar', data: values, barWidth: 18, itemStyle: { color: '#f97316', borderRadius: [0, 6, 6, 0] } }],
  }
})

const secondaryChartOption = computed(() => {
  if (props.currentRole === 'researcher') {
    const last7Days: string[] = []
    const counts: number[] = []
    for (let i = 6; i >= 0; i--) {
      const d = new Date()
      d.setDate(d.getDate() - i)
      const dateStr = d.toISOString().split('T')[0] || ''
      last7Days.push(dateStr)
      const stat = (props.stats?.recent_usage_trend || []).find((t: any) => t.date === dateStr)
      counts.push(stat ? stat.count : 0)
    }
    return {
      title: { text: '近7天消耗趋势', left: 'center', textStyle: { fontSize: 14, fontWeight: 500 } },
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '4%', bottom: '6%', top: '18%', containLabel: true },
      xAxis: { type: 'category', boundaryGap: false, data: last7Days },
      yAxis: { type: 'value', minInterval: 1 },
      series: [{
        name: '空瓶核销',
        type: 'line',
        smooth: true,
        data: counts,
        itemStyle: { color: '#8b5cf6' },
        areaStyle: { color: 'rgba(139, 92, 246, 0.16)' },
      }],
    }
  }

  return {
    title: { text: '试剂分类分布', left: 'center', textStyle: { fontSize: 14, fontWeight: 500 } },
    tooltip: { trigger: 'item' },
    legend: { orient: 'horizontal', bottom: 'bottom' },
    series: [
      {
        type: 'pie',
        radius: ['42%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
        label: { show: false },
        labelLine: { show: false },
        data: (props.stats?.category_distribution || []).map((s: any) => ({ value: s.count, name: s.category })),
      },
    ],
  }
})

const activityTitle = computed(() => (props.currentRole === 'researcher' ? '我的近期操作' : '关键风险提示'))
const activityList = computed(() => {
  if (props.currentRole === 'researcher') {
    return (props.stats?.recent_logs || []).slice(0, 4)
  }
  return (props.stats?.alerts || []).slice(0, 5)
})

const getPriorityVariant = (priority: Priority) => {
  if (priority === 'high') return 'destructive'
  if (priority === 'medium') return 'warning'
  return 'success'
}

const getLogVariant = (action: string): any => {
  const map: Record<string, string> = {
    入库登记: 'success',
    扫码入库: 'success',
    空瓶核销: 'warning',
    库位移动: 'info',
    状态变更: 'default',
    变更信息: 'default',
  }
  return map[action] || 'default'
}
</script>

<template>
  <div class="space-y-4">
    <div class="grid gap-4 md:grid-cols-3">
      <Card v-for="card in rolePrimaryCards" :key="card.key" class="transition-colors hover:border-blue-200">
        <div class="p-5">
          <div class="text-[11px] font-bold uppercase tracking-widest text-gray-400">{{ card.title }}</div>
          <div class="mt-2 text-3xl font-bold tracking-tight" :class="card.valueClass">{{ card.value }}</div>
          <div class="mt-1 text-[11px] text-gray-500">{{ card.hint }}</div>
        </div>
      </Card>
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <Card class="h-80 overflow-hidden shadow-sm">
        <div class="h-full p-4">
          <v-chart class="h-full w-full" :option="flowChartOption" autoresize />
        </div>
      </Card>
      <Card class="h-80 overflow-hidden shadow-sm">
        <div class="h-full p-4">
          <v-chart class="h-full w-full" :option="secondaryChartOption" autoresize />
        </div>
      </Card>
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <Card class="overflow-hidden shadow-sm">
        <div class="p-5">
          <h3 class="mb-3 text-base font-semibold text-gray-900">关键待办</h3>
          <div class="space-y-2.5">
            <div
              v-for="todo in roleTodos"
              :key="todo.key"
              class="flex items-center justify-between rounded-xl border border-gray-100 px-3 py-2.5"
            >
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold text-gray-900">{{ todo.title }}</div>
                <div class="truncate text-xs text-gray-500">{{ todo.hint }}</div>
              </div>
              <div class="ml-3 flex items-center gap-2">
                <span class="text-sm font-bold text-gray-800">{{ todo.count }}</span>
                <Badge :variant="getPriorityVariant(todo.priority)">
                  {{ todo.priority === 'high' ? '高' : todo.priority === 'medium' ? '中' : '低' }}
                </Badge>
              </div>
            </div>
          </div>
        </div>
      </Card>

      <Card class="overflow-hidden shadow-sm">
        <div class="p-5">
          <h3 class="mb-3 text-base font-semibold text-gray-900">{{ activityTitle }}</h3>
          <div v-if="activityList.length === 0" class="rounded-xl border border-dashed border-gray-200 bg-gray-50 py-12 text-center text-sm text-gray-500">
            当前无需要关注的数据。
          </div>

          <div v-else-if="currentRole === 'researcher'" class="space-y-2">
            <div v-for="log in activityList" :key="log.id" class="rounded-xl border border-gray-100 px-3 py-2.5">
              <div class="flex items-center gap-2">
                <Badge :variant="getLogVariant(log.action)">
                  {{ log.action }}
                </Badge>
                <span class="text-sm font-semibold text-gray-900">{{ log.reagent_item?.reagent_catalog?.name || '未知品目' }}</span>
              </div>
              <div class="mt-1.5 flex gap-3 text-[11px] text-gray-500">
                <span>数量: {{ formatNumber(log.quantity) }}</span>
                <span>{{ new Date(log.created_at).toLocaleDateString('zh-CN') }}</span>
              </div>
            </div>
          </div>

          <div v-else class="space-y-2">
            <div v-for="alert in activityList" :key="alert.ID" class="rounded-xl border border-red-100 bg-red-50/40 px-3 py-2.5">
              <div class="text-sm font-semibold text-gray-900">{{ alert.Name }} 低于阈值</div>
              <div class="mt-1.5 flex gap-3 text-[11px] text-gray-600">
                <span>剩余: <span class="font-semibold text-red-600">{{ alert.Count }}</span></span>
                <span>阈值: {{ alert.AlertThreshold }}</span>
              </div>
            </div>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>
