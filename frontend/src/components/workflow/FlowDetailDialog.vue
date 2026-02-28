<script setup lang="ts">
import { computed } from 'vue'
import Dialog from '@/components/ui/Dialog.vue'
import FlowStatusPill from '@/components/workflow/FlowStatusPill.vue'
import FlowTimeline from '@/components/workflow/FlowTimeline.vue'
import FlowActionPanel from '@/components/workflow/FlowActionPanel.vue'
import { Clock3 } from 'lucide-vue-next'

type FlowStepState = 'completed' | 'current' | 'pending' | 'rejected' | 'hidden'
type ActionVariant = 'primary' | 'secondary' | 'outline' | 'destructive'

interface FlowStepItem {
  key: string
  label: string
  state: FlowStepState
  description?: string
  operator?: string
  time?: string
}

interface FlowActionItem {
  key: string
  label: string
  variant?: ActionVariant
  disabled?: boolean
  loading?: boolean
}

interface FlowMetaItem {
  label: string
  value: string
}

interface FlowNoteItem {
  type?: 'info' | 'warning' | 'error' | 'success'
  text: string
}

interface FlowDocumentItem {
  key?: string
  name: string
  type?: string
  size?: string
}

const props = withDefaults(defineProps<{
  open: boolean
  title?: string
  subtitle?: string
  status?: string
  meta?: FlowMetaItem[]
  steps?: FlowStepItem[]
  actions?: FlowActionItem[]
  notes?: FlowNoteItem[]
  documents?: FlowDocumentItem[]
}>(), {
  title: '流程流转单',
  subtitle: '',
  status: '-',
  meta: () => [],
  steps: () => [],
  actions: () => [],
  notes: () => [],
  documents: () => [],
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'action', key: string): void
  (e: 'document-click', doc: FlowDocumentItem): void
}>()

const hasActions = computed(() => (props.actions || []).length > 0)
const flowTicketCode = computed(() => {
  const candidate = (props.meta || []).find((item) => item.label.includes('单号') || item.label.includes('编号'))
  return candidate?.value || '-'
})
const documentBadgeClass = (type?: string) => {
  const t = (type || '').toUpperCase()
  if (t === 'PDF') return 'bg-red-50 text-red-500 border-red-100'
  if (t === 'IMG' || t === 'PNG' || t === 'JPG') return 'bg-blue-50 text-blue-500 border-blue-100'
  return 'bg-slate-100 text-slate-500 border-slate-200'
}
</script>

<template>
  <Dialog :open="open" size="xl" :title="title" @close="emit('close')">
    <!-- 极限适配底座：横向保留呼吸感，纵向极限压缩 -->
    <div class="bg-slate-50/80 px-6 py-4 min-h-[550px]">
      <div class="grid gap-4 lg:grid-cols-12">
        <!-- 左侧核心区 -->
        <div class="lg:col-span-8 flex flex-col gap-4">
          <!-- 头部卡片 -->
          <div class="apple-card apple-card-hover p-5 border-blue-50/50">
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-baseline gap-4">
                <h2 class="text-2xl font-bold tracking-tight text-slate-900 leading-none">
                  {{ subtitle || '流程流转详情' }}
                </h2>
                <span class="text-[11px] font-bold text-slate-400 bg-slate-100/80 px-1.5 py-0.5 rounded">{{ flowTicketCode }}</span>
              </div>
              <FlowStatusPill :status="status" />
            </div>

            <!-- 数据条 -->
            <div v-if="meta.length > 0" class="flex flex-wrap items-center gap-x-10 gap-y-3 border-t border-slate-50 pt-4">
              <div v-for="item in meta" :key="item.label" class="flex items-center gap-2">
                <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400 whitespace-nowrap">{{ item.label }}:</span>
                <span class="text-sm font-bold text-slate-700">{{ item.value || '-' }}</span>
              </div>
            </div>
          </div>

          <!-- 时间轴内容 (悬浮层级) -->
          <div class="apple-card apple-card-hover p-5 flex-grow shadow-md border-slate-200/40">
            <div class="mb-4 flex items-center justify-between border-b border-slate-50 pb-3">
              <h3 class="flex items-center gap-2 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                <Clock3 class="h-3.5 w-3.5" />
                执行节点记录
              </h3>
            </div>
            <FlowTimeline :steps="steps" mode="full" />
          </div>
        </div>

        <!-- 右侧辅助增强区 (嵌入层级) -->
        <div class="lg:col-span-4 flex flex-col gap-4 text-nowrap">
          <!-- 关联附件 -->
          <div class="apple-card p-5 bg-slate-50/40 border-slate-200/30 shadow-none">
            <h4 class="mb-3 text-[10px] font-bold uppercase tracking-widest text-slate-400">关联文档</h4>
            <div v-if="documents.length > 0" class="space-y-2">
              <div
                v-for="doc in documents"
                :key="doc.key || doc.name"
                class="group flex items-center justify-between p-2 rounded-lg border border-slate-200 bg-white/80 hover:bg-white hover:border-blue-200 hover:shadow-sm transition-all cursor-pointer"
                @click="emit('document-click', doc)"
              >
                <div class="flex items-center gap-2.5 overflow-hidden">
                  <div :class="['h-7 w-7 rounded-md flex items-center justify-center shrink-0 border', documentBadgeClass(doc.type)]">
                    <span class="text-[9px] font-bold">{{ (doc.type || 'DOC').toUpperCase() }}</span>
                  </div>
                  <span class="text-xs font-bold text-slate-700 truncate">{{ doc.name }}</span>
                </div>
                <span v-if="doc.size" class="text-[9px] text-slate-400 shrink-0 ml-2">{{ doc.size }}</span>
              </div>
            </div>
            <div v-else class="rounded-lg border border-dashed border-slate-200 bg-white px-3 py-3 text-[11px] text-slate-400">暂无关联文档</div>
          </div>

          <!-- 审批操作 -->
          <div v-if="hasActions" class="apple-card p-5 bg-slate-50/40 border-slate-200/30 shadow-none">
            <h4 class="mb-2 text-[10px] font-bold uppercase tracking-widest text-slate-400">流转操作</h4>
            <FlowActionPanel
              :compact="true"
              :show-header="false"
              :actions="actions"
              @action="(actionKey) => emit('action', actionKey)"
              class="border-0 shadow-none p-0 bg-transparent"
            />
          </div>

          <!-- 业务备注 -->
          <div v-if="notes.length > 0" class="apple-card p-5 bg-slate-50/60 border-slate-200/20 shadow-none">
            <h4 class="mb-2 text-[10px] font-bold uppercase tracking-widest text-slate-400">附注</h4>
            <div class="space-y-2">
              <div
                v-for="(note, idx) in notes"
                :key="idx"
                class="text-[11px] leading-tight text-slate-500 flex items-start gap-2"
              >
                <div class="mt-1.5 h-0.5 w-0.5 shrink-0 rounded-full bg-slate-400/60"></div>
                {{ note.text }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Dialog>
</template>
