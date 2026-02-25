<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  variant?: 'default' | 'outline' | 'ghost' | 'destructive' | 'secondary'
  size?: 'sm' | 'md' | 'lg' | 'icon'
}>()

const variantClasses = {
  default: 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm',
  outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
  ghost: 'hover:bg-accent hover:text-accent-foreground',
  destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-sm',
  secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80 shadow-sm',
}

const sizeClasses = {
  sm: 'h-8 px-3 text-xs',
  md: 'h-10 px-4 py-2',
  lg: 'h-12 px-8 text-lg',
  icon: 'h-10 w-10 p-0 items-center justify-center',
}

const classes = computed(() => {
  return [
    'inline-flex items-center justify-center rounded-lg font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
    variantClasses[props.variant || 'default'],
    sizeClasses[props.size || 'md']
  ].join(' ')
})
</script>

<template>
  <button :class="classes">
    <slot />
  </button>
</template>
