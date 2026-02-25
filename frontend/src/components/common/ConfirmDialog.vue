<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next'

defineProps<{
  modelValue: boolean
  title?: string
  message?: string
  confirmText?: string
  cancelText?: string
  type?: 'danger' | 'warning' | 'info'
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
}>()
</script>

<template>
  <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
    <div class="fixed inset-0 bg-gray-900/30 backdrop-blur-sm transition-opacity" @click="emit('update:modelValue', false)"></div>

    <div class="relative w-full max-w-sm transform overflow-hidden rounded-xl bg-white p-6 text-left shadow-xl transition-all sm:my-8 border border-gray-100">
      <div class="flex flex-col items-center text-center">
          <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-red-100 mb-4">
              <AlertTriangle class="h-6 w-6 text-red-600" />
          </div>
          <h3 class="text-lg font-semibold leading-6 text-gray-900 mb-2">
              {{ title || '确认操作' }}
          </h3>
          <p class="text-sm text-gray-500 mb-6">
              {{ message || '您确定要执行此操作吗？' }}
          </p>
          
          <div class="grid grid-cols-2 gap-3 w-full">
              <button 
                type="button" 
                class="w-full justify-center rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                @click="emit('update:modelValue', false)"
              >
                  {{ cancelText || '取消' }}
              </button>
              <button 
                type="button" 
                class="w-full justify-center rounded-lg border border-transparent bg-red-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
                @click="emit('confirm')"
              >
                  {{ confirmText || '确认删除' }}
              </button>
          </div>
      </div>
    </div>
  </div>
</template>
