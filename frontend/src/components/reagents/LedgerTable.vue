<script setup lang="ts">
type ColumnAlign = 'left' | 'center' | 'right'

interface LedgerColumn {
  key: string
  label: string
  align?: ColumnAlign
  class?: string
}

const props = defineProps<{
  columns: LedgerColumn[]
}>()

const getAlignClass = (align?: ColumnAlign) => {
  if (align === 'right') return 'text-right'
  if (align === 'center') return 'text-center'
  return 'text-left'
}
</script>

<template>
  <div class="apple-table-wrap">
    <table class="w-full text-sm text-left">
      <thead class="text-xs text-gray-700 uppercase bg-gray-50 border-b">
        <tr>
          <th
            v-for="col in props.columns"
            :key="col.key"
            :class="['px-6 py-3', getAlignClass(col.align), col.class]"
          >
            {{ col.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <slot />
      </tbody>
    </table>
  </div>
</template>
