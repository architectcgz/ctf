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
    pending: 'text-[var(--color-warning)]',
    creating: 'text-[var(--color-warning)]',
    running: 'text-[var(--color-success)]',
    expired: 'text-[var(--color-text-muted)]',
    destroying: 'text-[var(--color-warning)]',
    destroyed: 'text-[var(--color-text-muted)]',
    failed: 'text-[var(--color-danger)]',
    crashed: 'text-[var(--color-danger)]',
  }
  return classes[props.instance.status]
})

const remainingLabel = computed(() => {
  if (!props.instance) return '--:--:--'
  if (isExpired.value) return '已过期'
  return formatted.value
})

function formatEta(seconds?: number) {
  if (typeof seconds !== 'number' || seconds <= 0) return '预计时间计算中'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  if (minutes <= 0) return `${secs} 秒`
  return `${minutes} 分 ${secs} 秒`
}

const canOpen = computed(() => props.instance?.status === 'running')
const canExtend = computed(() => canOpen.value && (props.instance?.remaining_extends ?? 0) > 0)
const isWaiting = computed(
  () => props.instance?.status === 'pending' || props.instance?.status === 'creating'
)
const isFailed = computed(
  () => props.instance?.status === 'failed' || props.instance?.status === 'crashed'
)
const createdAtLabel = computed(() => {
  if (!props.instance?.created_at) return ''
  return formatTime(props.instance.created_at)
})

const remainingExtendsLabel = computed(() => {
  if (!props.instance) return '0 次'
  return `${props.instance.remaining_extends} 次`
})

const queueLabel = computed(() => {
  if (!props.instance || !isWaiting.value) return ''
  if (typeof props.instance.queue_position === 'number' && props.instance.queue_position > 0) {
    return `当前排队：第 ${props.instance.queue_position} 位`
  }
  return '当前排队：排队信息同步中'
})

const etaLabel = computed(() => {
  if (!props.instance || !isWaiting.value) return ''
  return `预计等待：${formatEta(props.instance.eta_seconds)}`
})

const progressLabel = computed(() => {
  if (!props.instance || !isWaiting.value || typeof props.instance.progress !== 'number') return ''
  const normalized = Math.max(0, Math.min(100, Math.round(props.instance.progress)))
  return `创建进度：${normalized}%`
})

const accessLabel = computed(() => {
  if (!props.instance) return ''
  if (canOpen.value) {
    return props.instance.access_url || '通过右侧按钮打开代理访问'
  }
  if (isWaiting.value) {
    return '实例仍在排队/创建中，完成后可打开目标'
  }
  if (isFailed.value) {
    return '实例不可访问，请销毁后重新启动'
  }
  return props.instance.access_url || '--'
})

const openButtonLabel = computed(() => {
  if (props.opening) return '正在打开...'
  if (isWaiting.value) return '等待实例就绪'
  if (isFailed.value) return '实例不可用'
  return '打开目标'
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
        <div class="mt-2 text-2xl font-semibold" :class="isUrgent ? 'text-[var(--color-warning)]' : 'text-[var(--color-text-primary)]'">
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
            {{ accessLabel }}
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

        <div
          v-if="isWaiting"
          class="rounded-xl border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 px-4 py-3 text-xs leading-6 text-[var(--color-warning)]"
        >
          <div>实例正在排队创建，系统会自动刷新状态。</div>
          <div>{{ queueLabel }}</div>
          <div>{{ etaLabel }}</div>
          <div v-if="progressLabel">{{ progressLabel }}</div>
        </div>
        <div
          v-else-if="isFailed"
          class="rounded-xl border border-[var(--color-danger)]/30 bg-[var(--color-danger)]/10 px-4 py-3 text-xs leading-6 text-[var(--color-danger)]"
        >
          <div>实例创建失败或运行异常，当前目标不可访问。</div>
          <div>建议先销毁当前实例，再重新启动。</div>
        </div>
      </div>

      <div class="space-y-3">
        <button
          type="button"
          class="w-full rounded-xl bg-[var(--color-primary)] px-4 py-3 text-sm font-semibold text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="!canOpen || opening"
          @click="emit('open')"
        >
          {{ openButtonLabel }}
        </button>
        <div class="grid grid-cols-2 gap-3">
          <button
            type="button"
            class="rounded-xl border border-[var(--color-primary)]/40 bg-[var(--color-primary)]/10 px-4 py-3 text-sm font-medium text-[var(--color-primary-hover)] transition-colors hover:bg-[var(--color-primary)]/20 disabled:cursor-not-allowed disabled:border-[var(--color-border-default)] disabled:bg-[var(--color-bg-base)] disabled:text-[var(--color-text-muted)]"
            :disabled="!canExtend || extending"
            @click="emit('extend')"
          >
            {{ extending ? '延时中...' : '延时' }}
          </button>
          <button
            type="button"
            class="rounded-xl border border-[var(--color-danger)]/25 bg-[var(--color-danger)]/10 px-4 py-3 text-sm font-medium text-[var(--color-danger)] transition-colors hover:bg-[var(--color-danger)]/20 disabled:cursor-not-allowed disabled:opacity-50"
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
      <div v-else class="rounded-xl border border-[var(--color-success)]/25 bg-[var(--color-success)]/10 px-4 py-3 text-sm text-[var(--color-success)]">
        当前题目已完成，如仍需验证环境可前往实例列表查看历史实例。
      </div>
    </div>
  </section>
</template>
