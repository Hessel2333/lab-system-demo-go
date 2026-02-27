<script setup lang="ts">
const props = defineProps<{
  moduleName: string
}>()
</script>

<template>
  <div class="reagent-scope space-y-6">
    <div class="mb-2 border-b border-gray-100 pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-gray-900">{{ props.moduleName }}</h1>
      <p class="mt-1 text-sm text-gray-500">该模块正在建设中，后续将按统一台账与流程规范上线。</p>
    </div>

    <div class="rounded-2xl border border-slate-200 bg-white p-10 text-center shadow-sm">
      <h2 class="text-lg font-semibold text-slate-900">功能建设中</h2>
      <p class="mt-2 text-sm text-slate-500">当前仅开放信息架构入口，详细功能会在后续迭代逐步补齐。</p>
    </div>
  </div>
</template>
