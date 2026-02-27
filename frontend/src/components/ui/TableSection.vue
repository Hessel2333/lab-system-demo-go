<script setup lang="ts">
import Card from '@/components/ui/Card.vue'

defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <Card class="apple-table-section">
    <div class="apple-table-section-header">
      <div class="min-w-0">
        <h3 class="apple-table-section-title">{{ title }}</h3>
        <p v-if="description" class="apple-table-section-description">{{ description }}</p>
      </div>
      <div v-if="$slots.actions" class="shrink-0">
        <slot name="actions" />
      </div>
    </div>

    <div v-if="$slots.toolbar" class="apple-table-section-toolbar">
      <slot name="toolbar" />
    </div>

    <div class="apple-table-section-body">
      <slot />
    </div>
  </Card>
</template>
