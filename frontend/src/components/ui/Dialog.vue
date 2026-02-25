<script setup lang="ts">
import { X } from 'lucide-vue-next'

defineOptions({
  inheritAttrs: false
})

defineProps<{
  open: boolean
  title?: string
  maxWidth?: string
}>()

defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[100] flex items-center justify-center">
      <!-- Backdrop -->
      <div class="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity" @click="$emit('close')"></div>
      
      <!-- Modal Content -->
      <div 
        class="z-[101] w-full max-h-[90vh] flex flex-col bg-background shadow-lg sm:rounded-xl border animate-in fade-in zoom-in-95 duration-200 relative mx-4"
        :class="[$attrs.class, maxWidth || 'max-w-lg']"
        v-bind="$attrs"
      >
        <div v-if="title" class="flex items-center justify-between border-b px-6 py-4 shrink-0">
          <h3 class="text-lg font-semibold">{{ title }}</h3>
          <button @click="$emit('close')" class="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
            <X class="h-4 w-4" />
            <span class="sr-only">Close</span>
          </button>
        </div>
        
        <div class="overflow-y-auto flex-1">
          <slot />
        </div>
        
        <div v-if="$slots.footer" class="border-t bg-muted/50 px-6 py-4 flex justify-end gap-2 shrink-0">
           <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>
