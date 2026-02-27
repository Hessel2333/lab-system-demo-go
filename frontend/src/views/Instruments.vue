<script setup lang="ts">
import { ref } from 'vue'

// Sub-components
import InstrumentDashboard from '@/components/instruments/InstrumentDashboard.vue'
import InstrumentInventory from '@/components/instruments/InstrumentInventory.vue'
import InstrumentProcurement from '@/components/instruments/InstrumentProcurement.vue'
import InstrumentMaintenance from '@/components/instruments/InstrumentMaintenance.vue'
import InstrumentReservations from '@/components/instruments/InstrumentReservations.vue'

const activeTab = ref('dashboard')

const tabs = [
  { id: 'dashboard', label: '工作台' },
  { id: 'inventory', label: '仪器台账' },
  { id: 'procurement', label: '采购管理' },
  { id: 'maintenance', label: '维护保养' },
  { id: 'reservations', label: '预约使用' },
]
</script>

<template>
  <div class="reagent-scope space-y-6">
    <div class="mb-2 border-b border-gray-100 pb-4">
      <h1 class="text-2xl font-bold tracking-tight text-gray-900">仪器管理系统</h1>
      <p class="mt-1 text-sm text-gray-500">统一管理仪器台账、采购、维护和预约流程</p>
    </div>

    <div class="apple-segmented flex w-full gap-1.5">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="[
          'apple-segmented-btn w-full py-2.5 text-sm',
          activeTab === tab.id
            ? 'apple-segmented-btn-active'
            : 'apple-segmented-btn-idle'
        ]"
      >
        {{ tab.label }}
      </button>
    </div>

    <div class="min-h-[420px]">
      <Transition name="fade" mode="out-in">
        <InstrumentDashboard v-if="activeTab === 'dashboard'" />
        <InstrumentInventory v-else-if="activeTab === 'inventory'" />
        <InstrumentProcurement v-else-if="activeTab === 'procurement'" />
        <InstrumentMaintenance v-else-if="activeTab === 'maintenance'" />
        <InstrumentReservations v-else-if="activeTab === 'reservations'" />
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
