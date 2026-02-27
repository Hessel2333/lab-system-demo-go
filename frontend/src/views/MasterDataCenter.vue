<script setup lang="ts">
import { defineAsyncComponent, ref } from 'vue'
import Card from '@/components/ui/Card.vue'

const ReagentCatalogManager = defineAsyncComponent(() => import('@/components/reagents/ReagentCatalogManager.vue'))

type MasterTab = 'reagent' | 'consumable' | 'polymer' | 'gene' | 'bottle'
const activeTab = ref<MasterTab>('reagent')

const tabs: Array<{ id: MasterTab; label: string }> = [
  { id: 'reagent', label: '试剂品目' },
  { id: 'consumable', label: '耗材品目' },
  { id: 'polymer', label: '聚合物库' },
  { id: 'gene', label: '基因库条目' },
  { id: 'bottle', label: '试剂瓶模板' },
]
</script>

<template>
  <div class="reagent-scope space-y-6">
    <div class="mb-2 border-b border-gray-100 pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-gray-900">基础数据中心</h1>
      <p class="mt-1 text-sm text-gray-500">统一维护跨模块主数据，避免业务页承载底层数据治理职责。</p>
    </div>

    <div class="apple-segmented flex w-full gap-1.5">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="[
          'apple-segmented-btn w-full py-2.5 text-sm',
          activeTab === tab.id ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle'
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <div v-if="activeTab === 'reagent'" class="space-y-4">
      <ReagentCatalogManager />
    </div>

    <Card v-else>
      <div class="p-10 text-center">
        <h2 class="text-base font-semibold text-gray-900">模块建设中</h2>
        <p class="mt-2 text-sm text-gray-500">该主数据子模块将按统一字段规范与权限策略逐步开放。</p>
      </div>
    </Card>
  </div>
</template>

