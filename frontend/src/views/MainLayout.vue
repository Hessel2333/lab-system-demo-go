<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import {
  BarChart3,
  Brain,
  Settings,
  FlaskConical,
  TestTube,
  Database,
  Atom,
  Menu,
  Box,
  Users,
  ShoppingBag,
  SlidersHorizontal
} from 'lucide-vue-next'
import { useSessionStore, type AppRole } from '@/stores/session'

const isSidebarOpen = ref(false)
const route = useRoute()
const sessionStore = useSessionStore()

const roleOptions: Array<{ value: AppRole; label: string }> = [
  { value: 'researcher', label: '研发' },
  { value: 'procurement', label: '采购' },
  { value: 'leader', label: '负责人' },
]

const initials = computed(() => sessionStore.currentUserName.slice(0, 2))

interface NavigationItem {
  name: string
  href: string
  icon: any
  color: string
  upcoming?: boolean
}

const stableNavigation: NavigationItem[] = [
  { name: '系统概览', href: '/dashboard', icon: BarChart3, color: 'text-blue-600' },
  { name: '试剂管理', href: '/reagents', icon: TestTube, color: 'text-emerald-600' },
  { name: '仪器管理', href: '/instruments', icon: Settings, color: 'text-blue-500' },
  { name: '供应商管理', href: '/suppliers', icon: ShoppingBag, color: 'text-purple-600' },
  { name: '用户与组织', href: '/users', icon: Users, color: 'text-indigo-600' },
  { name: '基础数据', href: '/master-data', icon: SlidersHorizontal, color: 'text-slate-600' },
]

const upcomingNavigation: NavigationItem[] = [
  { name: 'AI智能中心', href: '/ai-center', icon: Brain, color: 'text-cyan-600', upcoming: true },
  { name: '实验管理', href: '/experiments', icon: FlaskConical, color: 'text-orange-600' },
  { name: '耗材管理', href: '/consumables', icon: Box, color: 'text-purple-600', upcoming: true },
  { name: '基因库分析', href: '/analysis', icon: Database, color: 'text-pink-600', upcoming: true },
  { name: '聚合物数据库', href: '/polymer', icon: Atom, color: 'text-indigo-600', upcoming: true },
]

const toggleSidebar = () => {
  isSidebarOpen.value = !isSidebarOpen.value
}

const isActive = (href: string) => route.path.startsWith(href)
</script>

<template>
  <div class="flex h-screen bg-gray-50/50">
    <!-- Mobile sidebar overlay -->
    <div
      v-if="isSidebarOpen"
      class="fixed inset-0 z-50 bg-gray-900/80 backdrop-blur-sm lg:hidden"
      @click="isSidebarOpen = false"
    />

    <!-- Sidebar -->
    <div
      class="fixed inset-y-0 left-0 z-50 w-72 bg-white/80 backdrop-blur-xl border-r border-gray-200/50 transition-transform duration-300 lg:translate-x-0 lg:static lg:inset-auto lg:flex lg:w-72 lg:flex-col"
      :class="isSidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <!-- Logo -->
      <div class="flex h-16 shrink-0 items-center px-6 gap-3 border-b border-gray-100/50">
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-md">
            <FlaskConical class="w-5 h-5 text-white" />
        </div>
        <span class="text-lg font-bold bg-clip-text text-transparent bg-gradient-to-r from-gray-900 to-gray-600">
          Lab System
        </span>
      </div>

      <!-- Nav -->
      <nav class="flex flex-1 flex-col gap-y-1 px-4 py-6 overflow-y-auto">
        <div class="px-3 pb-1 text-[11px] font-semibold uppercase tracking-widest text-gray-400">核心模块</div>
        <RouterLink
          v-for="item in stableNavigation"
          :key="item.name"
          :to="item.href"
          class="group flex gap-x-3 rounded-xl p-3 text-sm font-semibold leading-6 transition-all duration-200"
          :class="[
            isActive(item.href)
              ? 'bg-gray-100/80 text-gray-900 shadow-sm'
              : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900',
          ]"
          @click="isSidebarOpen = false"
        >
          <component :is="item.icon" class="h-5 w-5 shrink-0 transition-colors" :class="item.color" aria-hidden="true" />
          {{ item.name }}
        </RouterLink>

        <div class="my-3 border-t border-gray-100" />

        <div class="px-3 pb-1 text-[11px] font-semibold uppercase tracking-widest text-gray-400">建设中模块</div>
        <RouterLink
          v-for="item in upcomingNavigation"
          :key="item.name"
          :to="item.href"
          class="group flex items-center gap-x-3 rounded-xl p-3 text-sm font-semibold leading-6 transition-all duration-200"
          :class="[
            isActive(item.href)
              ? 'bg-gray-100/80 text-gray-900 shadow-sm'
              : 'text-gray-500 hover:bg-gray-50 hover:text-gray-700',
          ]"
          @click="isSidebarOpen = false"
        >
          <component :is="item.icon" class="h-5 w-5 shrink-0 transition-colors opacity-80" :class="item.color" aria-hidden="true" />
          <span>{{ item.name }}</span>
          <span class="ml-auto rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-semibold text-slate-500">
            建设中
          </span>
        </RouterLink>
      </nav>
      
      <!-- Footer User Profile -->
      <div class="border-t border-gray-100/50 p-4">
         <div class="flex items-center gap-3 p-2 rounded-xl bg-gray-50/80 border border-gray-100">
            <div class="w-8 h-8 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center border border-gray-200">
                <span class="text-xs font-bold text-gray-600">{{ initials }}</span>
            </div>
            <div class="text-sm">
                <p class="font-medium text-gray-900">{{ sessionStore.currentUserName }}</p>
                <p class="text-xs text-gray-500">{{ sessionStore.currentUserTitle }}</p>
            </div>
         </div>
         <div class="mt-3 apple-segmented w-full justify-between">
            <button
              v-for="role in roleOptions"
              :key="role.value"
              @click="sessionStore.setRole(role.value)"
              :class="[
                'apple-segmented-btn flex-1 text-center',
                sessionStore.currentRole === role.value ? 'apple-segmented-btn-active' : 'apple-segmented-btn-idle'
              ]"
            >
              {{ role.label }}
            </button>
         </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="flex flex-1 flex-col overflow-hidden">
        <!-- Mobile Header -->
        <div class="sticky top-0 z-40 flex h-16 items-center gap-x-4 border-b border-gray-200 bg-white/80 backdrop-blur-md px-4 shadow-sm lg:hidden">
            <button type="button" class="-m-2.5 p-2.5 text-gray-700 lg:hidden" @click="toggleSidebar">
                <span class="sr-only">Open sidebar</span>
                <Menu class="h-6 w-6" aria-hidden="true" />
            </button>
            <div class="flex-1 text-sm font-semibold leading-6 text-gray-900">实验室管理系统</div>
        </div>

        <main class="flex-1 overflow-y-auto bg-gray-50/50 p-4 lg:p-8">
            <RouterView v-slot="{ Component }">
                 <transition name="fade" mode="out-in">
                    <component :is="Component" />
                 </transition>
            </RouterView>
        </main>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
