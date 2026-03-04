<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
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
  SlidersHorizontal,
  Bell,
  ArrowRight
} from 'lucide-vue-next'
import axios from 'axios'
import { useSessionStore, type AppRole } from '@/stores/session'

const isSidebarOpen = ref(false)
const route = useRoute()
const sessionStore = useSessionStore()
const globalNoticeOpen = ref(false)
const statsSnapshot = ref<Record<string, any>>({})
const noticeMenuRef = ref<HTMLElement | null>(null)
let noticeRefreshTimer: ReturnType<typeof setInterval> | null = null

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

interface GlobalNoticeItem {
  key: string
  title: string
  description: string
  count: number
  tone: 'amber' | 'blue' | 'violet' | 'rose'
  to: any
}

const noticeItems = computed<GlobalNoticeItem[]>(() => {
  const stats = statsSnapshot.value || {}
  const focus = stats.role_focus || {}
  const role = sessionStore.currentRole
  const items: GlobalNoticeItem[] = []
  const add = (item: Omit<GlobalNoticeItem, 'count'> & { count: number }) => {
    const n = Number(item.count || 0)
    if (Number.isFinite(n) && n > 0) items.push({ ...item, count: n })
  }

  if (role === 'researcher') {
    add({
      key: 'pending-checkin',
      title: '待入库',
      description: '实物在暂存区，需选择实验室与试剂柜完成入库。',
      count: Number(focus.my_pending_checkin || 0),
      tone: 'amber',
      to: { path: '/reagents', query: { tab: 'arrival-confirm' } },
    })
    add({
      key: 'open-requests',
      title: '待处理申购',
      description: '待审批 / 待采购 / 已接单申购仍在进行中。',
      count: Number(focus.my_open_requests || 0),
      tone: 'blue',
      to: { path: '/reagents', query: { tab: 'history' } },
    })
    add({
      key: 'pending-dispense',
      title: '领用待办',
      description: '管控领用申请待审批或待双签确认。',
      count: Number(focus.my_pending_dispense || 0),
      tone: 'violet',
      to: { path: '/reagents', query: { tab: 'usage' } },
    })
    add({
      key: 'near-expiry',
      title: '临期试剂',
      description: '30 天内到期，建议优先使用或处理。',
      count: Number(focus.my_near_expiry_30d || 0),
      tone: 'rose',
      to: { path: '/reagents', query: { tab: 'unified-inventory' } },
    })
    return items
  }

  if (role === 'procurement') {
    add({
      key: 'receiving-lines',
      title: '到货待处理',
      description: '待确认到货或待研发完成入库闭环。',
      count: Number(focus.receiving_todo_lines || 0) + Number(stats.pending_checkin_items || 0),
      tone: 'amber',
      to: { path: '/reagents', query: { tab: 'receiving' } },
    })
    add({
      key: 'pending-procurement',
      title: '待采购申购',
      description: '待采购申请需尽快完成接单。',
      count: Number(focus.pending_procurement_requests || 0),
      tone: 'blue',
      to: { path: '/reagents', query: { tab: 'history' } },
    })
    add({
      key: 'unmatched-import',
      title: '导入待匹配',
      description: '采购导入仍有未匹配明细。',
      count: Number(focus.unmatched_import_lines || 0),
      tone: 'violet',
      to: { path: '/reagents', query: { tab: 'batch-import' } },
    })
    add({
      key: 'key-holder',
      title: '双签待确认',
      description: '你被配置为双签持有人，存在待确认单据。',
      count: Number(focus.key_holder_todo || 0),
      tone: 'rose',
      to: { path: '/reagents', query: { tab: 'usage' } },
    })
    return items
  }

  add({
    key: 'leader-request',
    title: '申购待审批',
    description: '待团队负责人处理的申购流程。',
    count: Number(focus.pending_leader_approvals || 0),
    tone: 'blue',
    to: { path: '/reagents', query: { tab: 'history' } },
  })
  add({
    key: 'leader-dispense',
    title: '领用待审批',
    description: '管控领用申请待团队负责人审批。',
    count: Number(focus.pending_dispense_approvals || 0),
    tone: 'amber',
    to: { path: '/reagents', query: { tab: 'usage' } },
  })
  add({
    key: 'leader-dual-sign',
    title: '双签进行中',
    description: '已进入 A/B 持有人确认阶段。',
    count: Number(focus.dual_sign_in_progress || 0),
    tone: 'violet',
    to: { path: '/reagents', query: { tab: 'usage' } },
  })
  add({
    key: 'leader-checkin',
    title: '待入库',
    description: '到货条目仍在暂存区，需推进协同处理。',
    count: Number(stats.pending_checkin_items || 0),
    tone: 'amber',
    to: { path: '/reagents', query: { tab: 'unified-inventory' } },
  })
  add({
    key: 'leader-risk',
    title: '低库存风险',
    description: '存在低于阈值的在库品目。',
    count: Number(stats.low_stock_alerts || 0),
    tone: 'rose',
    to: { path: '/dashboard' },
  })
  return items
})

const globalNoticeCount = computed(() => noticeItems.value.reduce((sum, item) => sum + item.count, 0))
const primaryNoticeRoute = computed(() => noticeItems.value[0]?.to || { path: '/dashboard' })

const toneClassMap: Record<GlobalNoticeItem['tone'], string> = {
  amber: 'border-amber-200 bg-amber-50/60 text-amber-900',
  blue: 'border-blue-200 bg-blue-50/60 text-blue-900',
  violet: 'border-violet-200 bg-violet-50/60 text-violet-900',
  rose: 'border-rose-200 bg-rose-50/60 text-rose-900',
}

const fetchGlobalNotice = async () => {
  try {
    const { data } = await axios.get('/api/reagents/stats', {
      params: { role: sessionStore.currentRole },
    })
    statsSnapshot.value = data || {}
  } catch {
    statsSnapshot.value = {}
  }
}

const toggleGlobalNotice = () => {
  globalNoticeOpen.value = !globalNoticeOpen.value
}

const closeGlobalNotice = () => {
  globalNoticeOpen.value = false
}

const handleOutsideClick = (event: MouseEvent) => {
  if (!globalNoticeOpen.value) return
  const target = event.target as Node | null
  if (!target) return
  if (noticeMenuRef.value && !noticeMenuRef.value.contains(target)) {
    closeGlobalNotice()
  }
}

onMounted(() => {
  fetchGlobalNotice()
  noticeRefreshTimer = setInterval(fetchGlobalNotice, 60000)
  document.addEventListener('click', handleOutsideClick)
})

onUnmounted(() => {
  if (noticeRefreshTimer) clearInterval(noticeRefreshTimer)
  document.removeEventListener('click', handleOutsideClick)
})

watch(
  () => [sessionStore.currentRole, sessionStore.currentUserId] as const,
  () => {
    fetchGlobalNotice()
    closeGlobalNotice()
  },
  { immediate: true }
)

watch(
  () => route.fullPath,
  () => {
    fetchGlobalNotice()
  }
)
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
            <RouterLink :to="primaryNoticeRoute" class="relative inline-flex h-10 w-10 items-center justify-center rounded-xl border border-gray-200 bg-white text-gray-600 shadow-sm">
              <Bell class="h-4.5 w-4.5" />
              <span
                v-if="globalNoticeCount > 0"
                class="absolute -right-1.5 -top-1.5 inline-flex min-w-5 items-center justify-center rounded-full bg-amber-500 px-1.5 text-[10px] font-bold text-white"
              >
                {{ globalNoticeCount > 99 ? '99+' : globalNoticeCount }}
              </span>
            </RouterLink>
        </div>

        <main class="flex-1 overflow-y-auto bg-gray-50/50 p-4 lg:p-8">
            <div class="sticky top-0 z-30 mb-4 hidden items-center justify-end lg:flex">
              <div ref="noticeMenuRef" class="relative">
                <button
                  class="group relative inline-flex h-11 w-11 items-center justify-center rounded-xl border border-gray-200 bg-white text-gray-600 shadow-sm transition-all hover:-translate-y-0.5 hover:border-blue-200 hover:text-blue-600"
                  @click.stop="toggleGlobalNotice"
                >
                  <Bell class="h-4.5 w-4.5" />
                  <span
                    v-if="globalNoticeCount > 0"
                    class="absolute -right-1.5 -top-1.5 inline-flex min-w-5 items-center justify-center rounded-full bg-amber-500 px-1.5 text-[10px] font-bold text-white"
                  >
                    {{ globalNoticeCount > 99 ? '99+' : globalNoticeCount }}
                  </span>
                </button>
                <div
                  v-if="globalNoticeOpen"
                  class="absolute right-0 mt-2 w-[340px] rounded-2xl border border-gray-200 bg-white p-4 shadow-xl"
                >
                  <div class="flex items-start justify-between gap-3">
                    <div>
                      <p class="text-sm font-semibold text-gray-900">全局提醒</p>
                      <p class="mt-1 text-xs text-gray-500">跨页面聚合待办，点击直达处理页面。</p>
                    </div>
                    <span class="inline-flex rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-700">
                      {{ globalNoticeCount }} 条
                    </span>
                  </div>

                  <div v-if="noticeItems.length === 0" class="mt-3 rounded-xl border border-gray-200 bg-gray-50 p-4 text-center text-xs text-gray-500">
                    当前没有需要处理的提醒。
                  </div>

                  <div v-else class="mt-3 space-y-2">
                    <RouterLink
                      v-for="item in noticeItems"
                      :key="item.key"
                      :to="item.to"
                      :class="['block rounded-xl border px-3 py-2.5 transition-all hover:-translate-y-0.5 hover:shadow-sm', toneClassMap[item.tone]]"
                      @click="closeGlobalNotice"
                    >
                      <div class="flex items-center justify-between gap-2">
                        <p class="text-sm font-semibold">{{ item.title }}</p>
                        <span class="inline-flex min-w-6 items-center justify-center rounded-full bg-white/70 px-1.5 py-0.5 text-xs font-bold">
                          {{ item.count }}
                        </span>
                      </div>
                      <p class="mt-1 text-xs leading-5 opacity-90">{{ item.description }}</p>
                      <span class="mt-1.5 inline-flex items-center text-xs font-semibold">
                        去处理
                        <ArrowRight class="ml-1 h-3.5 w-3.5" />
                      </span>
                    </RouterLink>
                  </div>
                </div>
              </div>
            </div>
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
