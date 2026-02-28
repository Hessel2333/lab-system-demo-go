<script setup lang="ts">
import { computed } from 'vue'
import { Check, Circle, Clock3, X } from 'lucide-vue-next'

type StepState = 'completed' | 'current' | 'pending' | 'rejected' | 'hidden'

interface FlowStep {
  key: string
  label: string
  state: StepState
  description?: string
  operator?: string
  time?: string
}

const props = withDefaults(defineProps<{
  steps: FlowStep[]
  mode?: 'compact' | 'full'
}>(), {
  mode: 'compact',
})

const visibleSteps = computed(() => (props.steps || []).filter((step) => step.state !== 'hidden'))

const stateStyles: Record<Exclude<StepState, 'hidden'>, { icon: any, iconBg: string, iconColor: string, badge: string, label: string }> = {
  completed: {
    icon: Check,
    iconBg: 'bg-emerald-50 text-emerald-600 border-emerald-100',
    iconColor: 'text-emerald-600',
    badge: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    label: '已完成'
  },
  current: {
    icon: Clock3,
    iconBg: 'bg-blue-50 text-blue-600 border-blue-100',
    iconColor: 'text-blue-600',
    badge: 'border-blue-200 bg-blue-50 text-blue-700',
    label: '进行中'
  },
  rejected: {
    icon: X,
    iconBg: 'bg-red-50 text-red-600 border-red-100',
    iconColor: 'text-red-600',
    badge: 'border-red-200 bg-red-50 text-red-700',
    label: '已驳回'
  },
  pending: {
    icon: Circle,
    iconBg: 'bg-slate-50 text-slate-400 border-slate-100',
    iconColor: 'text-slate-400',
    badge: 'border-slate-200 bg-slate-50 text-slate-500',
    label: '待处理'
  }
}

// 获取连接线背景
const getLineBackground = (index: number) => {
  const currentStep = visibleSteps.value[index]
  const nextStep = visibleSteps.value[index + 1]
  if (!currentStep || !nextStep) return 'bg-slate-100'
  
  if (currentStep.state === 'completed' && nextStep.state === 'completed') return 'bg-emerald-200'
  if (currentStep.state === 'completed' && nextStep.state === 'current') return 'bg-blue-200'
  
  return 'bg-slate-100'
}
</script>

<template>
  <div v-if="mode === 'compact'" class="flex flex-wrap gap-2">
    <div
      v-for="step in visibleSteps"
      :key="step.key"
      :class="['inline-flex items-center rounded-full border px-2.5 py-1 text-[11px] font-medium transition-all', (step.state !== 'hidden') ? stateStyles[step.state].badge : '']"
    >
      <component :is="stateStyles[step.state as Exclude<StepState, 'hidden'>].icon" class="mr-1 h-3 w-3" />
      {{ step.label }}
    </div>
  </div>

  <ol v-else class="relative space-y-4">
    <li v-for="(step, idx) in visibleSteps" :key="step.key" class="group relative flex gap-6 pl-0">
      <!-- 动态连接线 (几何修正：top 从 32px 移至 34px 以对齐下移的圆圈) -->
      <div 
        v-if="idx < visibleSteps.length - 1"
        :class="['absolute left-4 top-[34px] bottom-0 w-[1.5px] transition-colors duration-500', getLineBackground(idx)]"
      ></div>

      <!-- 极大简化节点图标 (像素微调 mt-[1.5px] 以对齐右侧标题) -->
      <div
        :class="[
          'relative z-10 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border transition-all duration-300 mt-[1.5px]',
          step.state === 'current' ? 'ring-4 ring-blue-500/10 scale-105 border-blue-400 animate-pulse-soft shadow-[0_0_15px_rgba(59,130,246,0.2)]' : '',
          step.state !== 'hidden' ? stateStyles[step.state].iconBg : ''
        ]"
      >
        <component :is="stateStyles[step.state as Exclude<StepState, 'hidden'>].icon" class="h-3.5 w-3.5" />
      </div>

      <!-- 节点卡片 (精工对齐与 pb-4 黄金比例) -->
      <div 
        :class="[
          'flex-grow pb-4 pt-1 px-5 rounded-xl border transition-all duration-300 group-hover:translate-x-1',
          step.state === 'current' 
            ? 'bg-blue-50/30 border-blue-200/60 shadow-[0_4px_12px_rgba(59,130,246,0.08)] ring-1 ring-blue-400/5' 
            : 'bg-white/30 border-slate-100/50 hover:bg-white hover:shadow-sm'
        ]"
      >
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-3 py-0.5">
            <span :class="['text-sm font-bold transition-colors', step.state === 'current' ? 'text-blue-700' : 'text-slate-800 uppercase']">
              {{ step.label }}
            </span>
            <span v-if="step.state !== 'hidden'" :class="['rounded-full border px-2 py-0.5 text-[10px] font-bold', stateStyles[step.state].badge]">
              {{ stateStyles[step.state].label }}
            </span>
          </div>

          <!-- 右对齐元数据 -->
          <div class="flex items-center gap-4 text-xs whitespace-nowrap">
            <div v-if="step.operator" class="flex items-center gap-2 font-bold text-slate-800">
              <span class="text-[10px] text-slate-400 uppercase tracking-tighter">操作人:</span>
              <span>{{ step.operator }}</span>
            </div>
            <span v-if="step.time" class="font-bold text-slate-500 opacity-90">{{ step.time }}</span>
          </div>
        </div>
        
        <div v-if="step.description" class="mt-1 text-[13px] leading-relaxed text-slate-500 pr-32">
          {{ step.description }}
        </div>
      </div>
    </li>
  </ol>
</template>
