<script setup lang="ts">
import { X } from 'lucide-vue-next'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

defineOptions({
  inheritAttrs: false
})

const dialogVariants = cva(
  'z-[101] relative mx-4 flex max-h-[90vh] w-full flex-col overflow-hidden border border-slate-200 bg-white shadow-2xl sm:rounded-2xl',
  {
    variants: {
      size: {
        sm: 'max-w-md',
        md: 'max-w-lg',
        lg: 'max-w-2xl',
        xl: 'max-w-4xl',
        '2xl': 'max-w-6xl',
        full: 'max-w-[95vw]',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  }
)

type DialogProps = VariantProps<typeof dialogVariants>

interface Props {
  open: boolean
  title?: string
  size?: DialogProps['size']
  class?: string
}

const props = defineProps<Props>()

defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div 
          class="fixed inset-0 bg-slate-950/40 backdrop-blur-sm" 
          @click="$emit('close')"
        ></div>
        
        <!-- Modal Content -->
        <Transition
          appear
          enter-active-class="transition duration-300 ease-out-back"
          enter-from-class="opacity-0 scale-95 translate-y-4"
          enter-to-class="opacity-100 scale-100 translate-y-0"
        >
          <div 
            :class="cn(dialogVariants({ size: props.size }), props.class)"
            v-bind="$attrs"
          >
            <!-- Header -->
            <div v-if="title" class="sticky top-0 z-10 flex shrink-0 items-center justify-between border-b border-slate-100 bg-white/90 px-6 py-4 backdrop-blur-md">
              <h3 class="text-lg font-semibold tracking-tight text-slate-900">{{ title }}</h3>
              <button 
                @click="$emit('close')" 
                class="rounded-full p-2 text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600 active:scale-90"
              >
                <X class="h-5 w-5" />
                <span class="sr-only">Close</span>
              </button>
            </div>
            
            <div class="overflow-y-auto flex-1 custom-scrollbar">
              <slot />
            </div>
            
            <div v-if="$slots.footer" class="flex shrink-0 justify-end gap-3 border-t border-slate-100 bg-slate-50/70 px-6 py-4">
               <slot name="footer" />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.ease-out-back {
  transition-timing-function: cubic-bezier(0.34, 1.56, 0.64, 1);
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #e5e7eb;
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #d1d5db;
}
</style>
