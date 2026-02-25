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
  <div class="min-h-full flex flex-col p-6 space-y-6">
    <!-- Header -->
    <div class="rounded-2xl bg-gradient-to-r from-blue-600 to-cyan-600 p-6 text-white shadow-lg relative overflow-hidden shrink-0">
      <!-- Decorative circles -->
      <div class="absolute top-0 right-0 -mr-20 -mt-20 w-64 h-64 rounded-full bg-white/10 blur-3xl pointer-events-none"></div>
      
      <div class="flex items-center justify-between relative z-10">
        <div>
          <h1 class="text-3xl font-bold tracking-tight">仪器管理系统</h1>
          <p class="mt-2 text-blue-100 opacity-90 text-sm font-medium">全生命周期资产管理 · 采购 · 维护 · 预约</p>
        </div>
        <div class="flex gap-3">
            <!-- Action buttons moved to specific tabs -->
        </div>
      </div>
    </div>

    <!-- Tabs Navigation -->
    <div class="border-b border-gray-200 shrink-0">
      <nav class="-mb-px flex space-x-8" aria-label="Tabs">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="[
            activeTab === tab.id
              ? 'border-blue-500 text-blue-600'
              : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
            'group inline-flex items-center border-b-2 py-4 px-1 text-sm font-medium transition-colors duration-200'
          ]"
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- Tab Content -->
    <div class="flex-1">
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
