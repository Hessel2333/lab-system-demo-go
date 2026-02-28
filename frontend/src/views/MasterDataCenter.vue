<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue'
import TableSection from '@/components/ui/TableSection.vue'
import Input from '@/components/ui/Input.vue'
import { Search } from 'lucide-vue-next'

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

const activeTabLabel = computed(() => tabs.find((tab) => tab.id === activeTab.value)?.label || '主数据')
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

    <TableSection v-else :title="`${activeTabLabel}台账`" description="主数据台账骨架已就绪，后续将按同一字段规范逐步开放。">
      <template #toolbar>
        <div class="flex w-full flex-col gap-3 sm:flex-row sm:items-center">
          <div class="relative w-full sm:w-80">
            <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-gray-400" />
            <Input disabled class="pl-9" placeholder="搜索功能建设中" />
          </div>
        </div>
      </template>
      <div class="apple-table-empty">该主数据子模块建设中，后续将提供统一台账与新增流程。</div>
    </TableSection>
  </div>
</template>
