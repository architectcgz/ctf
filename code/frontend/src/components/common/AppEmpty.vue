<template>
  <div class="flex flex-col items-center justify-center rounded-lg border border-border bg-surface px-6 py-10 text-center">
    <component :is="iconComp" class="h-12 w-12 text-text-muted" />
    <div class="mt-3 text-sm font-semibold">{{ title }}</div>
    <div v-if="description" class="mt-1 text-sm text-text-secondary">{{ description }}</div>
    <div v-if="$slots.action" class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { Flag, Inbox, UsersRound } from 'lucide-vue-next'
import { computed } from 'vue'

type IconComp = Component
const iconRegistry: Record<string, Component> = {
  Inbox,
  Flag,
  UsersRound,
}

const props = withDefaults(
  defineProps<{
    title?: string
    description?: string
    icon?: string
  }>(),
  {
    title: '暂无数据',
    description: '',
    icon: 'Inbox',
  }
)

const iconComp = computed<IconComp>(() => iconRegistry[props.icon] || Inbox)
</script>
