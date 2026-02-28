<script setup lang="ts">
import { computed } from 'vue'
import FlowDetailDialog from '@/components/workflow/FlowDetailDialog.vue'
import { formatNumber } from '@/lib/quantity'

type FlowStepState = 'completed' | 'current' | 'pending' | 'rejected' | 'hidden'

interface FlowStepItem {
  key: string
  label: string
  state: FlowStepState
  description?: string
  operator?: string
  time?: string
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

const props = defineProps<{ open: boolean; request: any }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'refresh'): void
}>()

const formatTime = (t?: string | null) => {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const flowMeta = computed<FlowMetaItem[]>(() => {
  const req = props.request
  if (!req) return []
  return [
    { label: '申请单号', value: `#${req.id}` },
    { label: '申请人', value: req.requestor?.real_name || '-' },
    { label: '试剂', value: req.reagent_catalog?.name || '-' },
    { label: '数量', value: `${formatNumber(req.quantity, 0)} 瓶` },
    { label: '优先级', value: req.request_type || '日常' },
    { label: '要求交期', value: req.expected_delivery || '尽快到货' },
  ]
})

const flowDocuments = computed<FlowDocumentItem[]>(() => {
  const req = props.request
  if (!req) return []
  const docs: FlowDocumentItem[] = []
  const attachment = String(req.order_attachment || '').trim()
  if (attachment) {
    const ext = attachment.split('.').pop()?.toUpperCase() || 'DOC'
    docs.push({
      key: 'order-attachment',
      name: '采购凭证附件',
      type: ext,
    })
  }
  if (req.order_reference) {
    docs.push({
      key: 'order-reference',
      name: `采购单号：${req.order_reference}`,
      type: 'PO',
    })
  }
  return docs
})

const flowNotes = computed<FlowNoteItem[]>(() => {
  const req = props.request
  if (!req) return []

  const notes: FlowNoteItem[] = []
  if (req.reagent_catalog?.is_controlled || req.is_controlled) {
    notes.push({ type: 'warning', text: '该申请涉及管控品，需团队长审批后才可进入采购执行。' })
  }
  notes.push({ type: 'info', text: '该流转单仅展示申购 BPM 流程；到货确认与入库请在“到货台账”独立跟踪。' })
  if (req.status === '已接单') {
    notes.push({ type: 'success', text: '采购执行已完成，后续物理台账由到货/库存流程独立管理。' })
  }

  if (req.status === '已驳回' && req.remarks) {
    notes.push({ type: 'error', text: `驳回备注：${req.remarks}` })
  }

  return notes
})

const bpmASteps = computed<FlowStepItem[]>(() => {
  const req = props.request
  if (!req) return []

  const status = req.status
  const isControlled = !!(req.reagent_catalog?.is_controlled || req.is_controlled)

  const step1: FlowStepItem = {
    key: 'submitted',
    label: '提交申购',
    state: 'completed',
    description: `${req.requestor?.real_name || '申请人'} 已提交申购需求。`,
    operator: req.requestor?.real_name || '申请人',
    time: formatTime(req.created_at),
  }

  const step2: FlowStepItem | null = isControlled
    ? {
        key: 'leader-approve',
        label: '团队长审批',
        state: status === '待审批' ? 'current' : ['待采购', '已接单', '已入库'].includes(status) ? 'completed' : status === '已驳回' ? 'rejected' : 'pending',
        description: '团队长审批后进入采购执行。',
        operator: ['待采购', '已接单', '已入库'].includes(status) ? '团队长' : undefined,
        time: ['待采购', '已接单', '已入库'].includes(status) ? formatTime(req.updated_at) : undefined,
      }
    : null

  const step3: FlowStepItem = {
    key: 'procurement',
    label: '采购执行',
    state: ['待采购'].includes(status)
      ? 'current'
      : ['已接单'].includes(status)
      ? 'completed'
      : status === '已驳回'
      ? 'rejected'
      : 'pending',
    description: status === '已驳回' ? '申购流程已驳回结束。' : '采购人员完成下单并回填采购凭证。',
    operator: ['待采购', '已接单'].includes(status) ? '采购人员' : undefined,
    time: ['已接单'].includes(status) ? formatTime(req.updated_at) : undefined,
  }

  return [step1, ...(step2 ? [step2] : []), step3]
})
</script>

<template>
  <FlowDetailDialog
    :open="open"
    title="申购单进度流转"
    :subtitle="request?.reagent_catalog?.name || '申购流程详情'"
    :status="request?.status || '-'"
    :meta="flowMeta"
    :steps="bpmASteps"
    :notes="flowNotes"
    :documents="flowDocuments"
    @close="emit('close')"
  />
</template>
