<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'

type ActionVariant = 'primary' | 'secondary' | 'outline' | 'destructive'

interface FlowAction {
  key: string
  label: string
  variant?: ActionVariant
  disabled?: boolean
  loading?: boolean
}

withDefaults(defineProps<{
  title?: string
  description?: string
  actions: FlowAction[]
  compact?: boolean
  emptyText?: string
  showHeader?: boolean
}>(), {
  title: '可执行动作',
  description: '',
  compact: true,
  emptyText: '当前状态下无可执行动作',
  showHeader: true,
})

const emit = defineEmits<{
  (e: 'action', key: string): void
}>()

const clickAction = (action: FlowAction) => {
  if (action.disabled || action.loading) return
  emit('action', action.key)
}
</script>

<template>
  <Card :class="compact ? 'border-slate-200 bg-slate-50/60' : 'border-slate-200 bg-white'">
    <div :class="compact ? 'p-3' : 'p-4'">
      <div v-if="showHeader" class="mb-2">
        <div class="text-[11px] font-semibold uppercase tracking-[0.08em] text-slate-500">{{ title }}</div>
        <div v-if="description" class="mt-0.5 text-[11px] text-slate-500">{{ description }}</div>
      </div>

      <div v-if="actions.length === 0" class="rounded-lg border border-dashed border-slate-200 bg-white px-3 py-2 text-xs text-slate-400">
        {{ emptyText }}
      </div>

      <div v-else-if="compact" class="flex flex-wrap gap-2">
        <Button
          v-for="action in actions"
          :key="action.key"
          size="sm"
          :variant="action.variant || 'outline'"
          :disabled="action.disabled || action.loading"
          @click="clickAction(action)"
        >
          <Loader2 v-if="action.loading" class="mr-1.5 h-3.5 w-3.5 animate-spin" />
          {{ action.label }}
        </Button>
      </div>

      <div v-else class="space-y-2">
        <Button
          v-for="action in actions"
          :key="action.key"
          size="sm"
          :variant="action.variant || 'outline'"
          class="w-full justify-center"
          :disabled="action.disabled || action.loading"
          @click="clickAction(action)"
        >
          <Loader2 v-if="action.loading" class="mr-1.5 h-3.5 w-3.5 animate-spin" />
          {{ action.label }}
        </Button>
      </div>
    </div>
  </Card>
</template>
