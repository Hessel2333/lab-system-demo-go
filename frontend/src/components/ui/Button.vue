<script setup lang="ts">
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap rounded-xl font-semibold transition-all duration-200 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/15 disabled:pointer-events-none disabled:opacity-50 active:scale-[0.98]',
  {
    variants: {
      variant: {
        default: 'bg-blue-600 text-white shadow-sm shadow-blue-200 hover:bg-blue-700',
        primary: 'bg-blue-600 text-white shadow-sm shadow-blue-200 hover:bg-blue-700',
        secondary: 'bg-slate-100 text-slate-700 hover:bg-slate-200',
        destructive: 'bg-red-500 text-white shadow-sm shadow-red-200 hover:bg-red-600',
        outline: 'border border-slate-200 bg-white text-slate-700 hover:bg-slate-50',
        ghost: 'text-slate-600 hover:bg-slate-100 hover:text-slate-900',
        link: 'text-blue-600 underline-offset-4 hover:underline'
      },
      size: {
        default: 'h-10 px-4 py-2 text-sm',
        md: 'h-10 px-4 py-2 text-sm',
        sm: 'h-8 px-3 text-xs rounded-lg',
        lg: 'h-11 px-8 text-sm',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
)

type ButtonVariants = VariantProps<typeof buttonVariants>

interface Props {
  variant?: ButtonVariants['variant']
  size?: ButtonVariants['size']
  class?: string
}

const props = defineProps<Props>()
</script>

<template>
  <button :class="cn(buttonVariants({ variant, size }), props.class)">
    <slot />
  </button>
</template>
