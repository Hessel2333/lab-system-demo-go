<script setup lang="ts">
import { computed } from 'vue'
import { AlertCircle, CheckCircle2, Clock3, Lock, XCircle } from 'lucide-vue-next'
import Badge from '@/components/ui/Badge.vue'

type PillVariant = 'default' | 'primary' | 'secondary' | 'destructive' | 'outline' | 'success' | 'warning' | 'info'

interface StatusConfig {
  label: string
  variant: PillVariant
  icon: any
  className?: string
}

const props = withDefaults(defineProps<{ status?: string }>(), {
  status: '-',
})

const config = computed<StatusConfig>(() => {
  const s = props.status || '-'
  const map: Record<string, StatusConfig> = {
    '待审批': { label: '待审批', variant: 'warning', icon: Clock3 },
    '待双签': { label: '待双签', variant: 'info', icon: Lock, className: 'bg-violet-100 text-violet-700 border-violet-200' },
    '待采购': { label: '待采购', variant: 'warning', icon: Clock3 },
    '已通过': { label: '已通过', variant: 'success', icon: CheckCircle2 },
    '已接单': { label: '已接单', variant: 'primary', icon: CheckCircle2 },
    '已完成': { label: '已完成', variant: 'success', icon: CheckCircle2 },
    '已驳回': { label: '已驳回', variant: 'destructive', icon: XCircle },
  }
  return map[s] || { label: s, variant: 'secondary', icon: AlertCircle }
})
</script>

<template>
  <Badge :variant="config.variant" :class="`inline-flex items-center gap-1.5 rounded-xl border px-3.5 py-1.5 text-[10px] font-bold uppercase tracking-wider shadow-sm transition-all hover:scale-105 active:scale-95 ${config.className || ''}`">
    <component :is="config.icon" class="h-3 w-3" />
    <span>{{ config.label }}</span>
  </Badge>
</template>
