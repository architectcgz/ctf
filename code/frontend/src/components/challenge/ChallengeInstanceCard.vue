<script setup lang="ts">
import { computed } from 'vue'

import type { InstanceData, InstanceStatus } from '@/api/contracts'
import { useCountdown } from '@/composables/useCountdown'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  instance: InstanceData | null
  loading: boolean
  creating: boolean
  opening: boolean
  extending: boolean
  destroying: boolean
  challengeSolved: boolean
}>()

const emit = defineEmits<{
  start: []
  open: []
  extend: []
  destroy: []
}>()

const { formatted, isExpired, isUrgent } = useCountdown(() => props.instance?.expires_at)

const statusLabel = computed(() => {
  if (!props.instance) return '未创建'

  const labels: Record<InstanceStatus, string> = {
    pending: '等待中',
    creating: '创建中',
    running: '运行中',
    expired: '已过期',
    destroying: '销毁中',
    destroyed: '已销毁',
    failed: '失败',
    crashed: '崩溃',
  }
  return labels[props.instance.status]
})

const statusClass = computed(() => {
  if (!props.instance) return 'text-[var(--color-text-muted)]'

  const classes: Record<InstanceStatus, string> = {
    pending: 'text-amber-400',
    creating: 'text-amber-400',
    running: 'text-emerald-400',
    expired: 'text-[var(--color-text-muted)]',
    destroying: 'text-amber-400',
    destroyed: 'text-[var(--color-text-muted)]',
    failed: 'text-rose-400',
    crashed: 'text-rose-400',
  }
  return classes[props.instance.status]
})

const remainingLabel = computed(() => {
  if (!props.instance) return '--:--:--'
  if (isExpired.value) return '已过期'
  return formatted.value
})

const canOpen = computed(() => props.instance?.status === 'running')
const canExtend = computed(() => canOpen.value && (props.instance?.remaining_extends ?? 0) > 0)
const createdAtLabel = computed(() => {
  if (!props.instance?.created_at) return ''
  return formatTime(props.instance.created_at)
})

const remainingExtendsLabel = computed(() => {
  if (!props.instance) return '0 次'
  return `${props.instance.remaining_extends} 次`
})
</script>

<template>
  <section class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-5 shadow-[0_18px_48px_rgba(15,23,42,0.18)]">
    <div class="flex items-start justify-between gap-3">
      <div>
        <div class="text-xs uppercase tracking-[0.22em] text-[var(--color-text-muted)]">Instance</div>
        <h2 class="mt-2 text-xl font-semibold text-[var(--color-text-primary)]">靶机实例</h2>
      </div>
      <span
        class="rounded-full border border-white/10 px-3 py-1 text-xs font-medium"
        :class="statusClass"
      >
        {{ statusLabel }}
      </span>
    </div>

    <div v-if="loading && !instance" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-5 text-sm text-[var(--color-text-secondary)]">
      正在同步当前题目的实例状态...
    </div>

    <div v-else-if="instance" class="mt-5 space-y-5">
      <div class="rounded-xl bg-[var(--color-bg-base)] p-4">
        <div class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-muted)]">Remaining</div>
        <div class="mt-2 text-2xl font-semibold" :class="isUrgent ? 'text-amber-300' : 'text-[var(--color-text-primary)]'">
          {{ remainingLabel }}
        </div>
        <div class="mt-2 text-xs text-[var(--color-text-secondary)]">
          创建于 {{ createdAtLabel }}
        </div>
      </div>

      <div class="space-y-3 text-sm">
        <div class="rounded-xl border border-[var(--color-border-default)] px-4 py-3">
          <div class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-muted)]">Access</div>
          <div class="mt-2 break-all font-mono text-[var(--color-text-primary)]">
            {{ instance.access_url || '通过右侧按钮打开代理访问' }}
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div class="rounded-xl border border-[var(--color-border-default)] px-4 py-3">
            <div class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-muted)]">剩余延时</div>
            <div class="mt-2 text-lg font-semibold text-[var(--color-text-primary)]">{{ remainingExtendsLabel }}</div>
          </div>
          <div class="rounded-xl border border-[var(--color-border-default)] px-4 py-3">
            <div class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-muted)]">实例状态</div>
            <div class="mt-2 text-lg font-semibold" :class="statusClass">{{ statusLabel }}</div>
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <button
          type="button"
          class="w-full rounded-xl bg-[var(--color-primary)] px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="!canOpen || opening"
          @click="emit('open')"
        >
          {{ opening ? '正在打开...' : '打开目标' }}
        </button>
        <div class="grid grid-cols-2 gap-3">
          <button
            type="button"
            class="rounded-xl border border-sky-400/40 bg-sky-500/12 px-4 py-3 text-sm font-medium text-sky-950 transition-colors hover:bg-sky-500/20 disabled:cursor-not-allowed disabled:border-[var(--color-border-default)] disabled:bg-[var(--color-bg-base)] disabled:text-[var(--color-text-muted)]"
            :disabled="!canExtend || extending"
            @click="emit('extend')"
          >
            {{ extending ? '延时中...' : '延时' }}
          </button>
          <button
            type="button"
            class="rounded-xl border border-rose-400/25 bg-rose-500/10 px-4 py-3 text-sm font-medium text-rose-300 transition-colors hover:bg-rose-500/20 disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="destroying"
            @click="emit('destroy')"
          >
            {{ destroying ? '销毁中...' : '销毁' }}
          </button>
        </div>
      </div>
    </div>

    <div v-else class="mt-5 space-y-4">
      <div class="rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-5 text-sm leading-6 text-[var(--color-text-secondary)]">
        <div>实例会在当前题目页右侧保持可见，便于一边读题一边打开目标、延时或销毁。</div>
        <div class="mt-2">默认有效期 2 小时。</div>
      </div>
      <button
        v-if="!challengeSolved"
        type="button"
        class="w-full rounded-xl bg-[var(--color-primary)] px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:cursor-not-allowed disabled:opacity-50"
        :disabled="creating"
        @click="emit('start')"
      >
        {{ creating ? '正在创建实例...' : '启动靶机' }}
      </button>
      <div v-else class="rounded-xl border border-emerald-500/25 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-300">
        当前题目已完成，如仍需验证环境可前往实例列表查看历史实例。
      </div>
    </div>
  </section>
</template>
