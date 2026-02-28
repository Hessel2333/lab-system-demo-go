<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import axios from 'axios'
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

const items = ref<any[]>([])
const isLoadingItems = ref(false)

const arrivedItems = computed(() => items.value.filter((i) => i.status === '已到货'))
const storedItems = computed(() => items.value.filter((i) => i.status === '在库'))

const formatTime = (t?: string | null) => {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const latestItemTime = computed(() => {
  const timestamps = items.value
    .map((i) => i.updated_at || i.created_at)
    .filter(Boolean)
    .map((t) => new Date(t).getTime())
    .filter((n) => Number.isFinite(n))
  if (timestamps.length === 0) return ''
  return formatTime(new Date(Math.max(...timestamps)).toISOString())
})

const fetchItems = async () => {
  if (!props.request?.id) {
    items.value = []
    return
  }
  isLoadingItems.value = true
  try {
    const res = await axios.get(`/api/reagents/items?request_id=${props.request.id}`)
    items.value = res.data || []
  } catch {
    items.value = []
  } finally {
    isLoadingItems.value = false
  }
}

watch(
  () => [props.open, props.request?.id],
  ([open]) => {
    if (open) fetchItems()
  },
  { immediate: true }
)

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

  if (isLoadingItems.value) {
    notes.push({ type: 'info', text: '正在同步到货与库存实体数据...' })
  } else if (items.value.length === 0) {
    notes.push({ type: 'info', text: '当前尚未关联到货实物，后续将随采购与到货流程自动生成。' })
  } else {
    notes.push({
      type: 'success',
      text: `已关联实物 ${items.value.length} 瓶（已到货 ${arrivedItems.value.length}，在库 ${storedItems.value.length}）。`,
    })
    if (arrivedItems.value.length > 0) {
      const preview = arrivedItems.value
        .slice(0, 5)
        .map((item) => `#${String(item.uuid || '').substring(0, 8).toUpperCase()}`)
        .join('、')
      notes.push({ type: 'info', text: `待入库条码：${preview}${arrivedItems.value.length > 5 ? ' 等' : ''}` })
    }
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
    state: ['待采购', '待审批'].includes(status)
      ? 'current'
      : ['已接单', '已入库'].includes(status)
      ? 'completed'
      : status === '已驳回'
      ? 'rejected'
      : 'pending',
    description: status === '已驳回' ? '申购流程已驳回结束。' : '采购人员汇总并向供应商下单。',
    operator: ['待采购', '已接单', '已入库'].includes(status) ? '采购人员' : undefined,
    time: ['已接单', '已入库'].includes(status) ? formatTime(req.updated_at) : undefined,
  }

  const step4: FlowStepItem = {
    key: 'arrival-checkin',
    label: '到货与入库',
    state: storedItems.value.length > 0
      ? 'completed'
      : (arrivedItems.value.length > 0 || status === '已接单')
      ? 'current'
      : status === '已驳回'
      ? 'rejected'
      : 'pending',
    description: storedItems.value.length > 0 ? '实物已入库。' : '等待到货确认并完成入库。',
    operator: storedItems.value.length > 0 ? '研发人员' : undefined,
    time: storedItems.value.length > 0 ? latestItemTime.value : undefined,
  }

  return [step1, ...(step2 ? [step2] : []), step3, step4]
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
