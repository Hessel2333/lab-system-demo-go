<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { ChevronDown } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

defineOptions({
  inheritAttrs: false
})

type SelectValue = string | number | boolean | null | undefined

interface SelectOption {
  label: string
  value: string | number | boolean
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue?: SelectValue
  modelModifiers?: {
    number?: boolean
    trim?: boolean
  }
  options?: SelectOption[]
  placeholder?: string
  placeholderValue?: string | number
}>(), {
  modelModifiers: () => ({})
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: SelectValue): void
}>()

const attrs = useAttrs()

const selectClass = computed(() =>
  cn(
    'flex h-10 w-full appearance-none rounded-lg border border-input bg-background px-3 py-2 pr-9 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
    attrs.class as any
  )
)

const passthroughAttrs = computed(() => {
  const { class: _class, ...rest } = attrs
  return rest
})

const onChange = (event: Event) => {
  const rawValue = (event.target as HTMLSelectElement).value
  let nextValue: SelectValue = rawValue

  if (props.modelModifiers.number && rawValue !== '') {
    const parsed = Number(rawValue)
    nextValue = Number.isNaN(parsed) ? rawValue : parsed
  }

  if (!props.modelModifiers.number && (props.modelValue === true || props.modelValue === false)) {
    if (rawValue === 'true') nextValue = true
    if (rawValue === 'false') nextValue = false
  }

  if (props.modelModifiers.trim && typeof nextValue === 'string') {
    nextValue = nextValue.trim()
  }

  emit('update:modelValue', nextValue)
}
</script>

<template>
  <div class="relative">
    <select
      :value="modelValue as any"
      :class="selectClass"
      v-bind="passthroughAttrs"
      @change="onChange"
    >
      <option v-if="placeholder !== undefined" :value="placeholderValue ?? ''">
        {{ placeholder }}
      </option>
      <option
        v-for="option in options || []"
        :key="`${option.value}`"
        :value="option.value"
        :disabled="option.disabled"
      >
        {{ option.label }}
      </option>
      <slot />
    </select>
    <ChevronDown class="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 opacity-50" />
  </div>
</template>
