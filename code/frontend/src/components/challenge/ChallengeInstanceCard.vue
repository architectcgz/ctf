<script setup lang="ts">
import { computed } from 'vue'

import type { InstanceData, InstanceSharing, InstanceStatus } from '@/api/contracts'
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
  instanceSharing?: InstanceSharing
}>()

const emit = defineEmits<{
  start: []
  open: []
  extend: []
  destroy: []
}>()

const { formatted, isExpired, isUrgent } = useCountdown(() => props.instance?.expires_at)

const effectiveStatus = computed<InstanceStatus | null>(() => {
  if (!props.instance) return null
  if (props.instance.status === 'running' && isExpired.value) return 'expired'
  return props.instance.status
})

const statusLabel = computed(() => {
  if (!effectiveStatus.value) return '未创建'

  const labels: Record<InstanceStatus, string> = {
    pending: '等待中',
    creating: '创建中',
    running: '运行中',
    expired: '已自动回收',
    destroying: '销毁中',
    destroyed: '已销毁',
    failed: '启动失败',
    crashed: '运行异常',
  }
  return labels[effectiveStatus.value]
})

const statusClass = computed(() => {
  if (!effectiveStatus.value) return 'text-[var(--color-text-muted)]'

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
  return classes[effectiveStatus.value]
})

const remainingLabel = computed(() => {
  if (!props.instance) return '--:--:--'
  if (effectiveStatus.value === 'expired') return '已自动回收'
  return formatted.value
})

function formatEta(seconds?: number) {
  if (typeof seconds !== 'number' || seconds <= 0) return '预计时间计算中'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  if (minutes <= 0) return `${secs} 秒`
  return `${minutes} 分 ${secs} 秒`
}

const canOpen = computed(() => effectiveStatus.value === 'running')
const isWaiting = computed(
  () => effectiveStatus.value === 'pending' || effectiveStatus.value === 'creating'
)
const isFailed = computed(
  () => effectiveStatus.value === 'failed' || effectiveStatus.value === 'crashed'
)
const isReclaimingState = computed(
  () => effectiveStatus.value === 'expired' || effectiveStatus.value === 'destroyed'
)
const isRestartable = computed(() => isFailed.value || isReclaimingState.value)
const createdAtLabel = computed(() => {
  if (!props.instance?.created_at) return ''
  return formatTime(props.instance.created_at)
})

const remainingExtendsLabel = computed(() => {
  if (isSharedInstance.value) return '系统托管'
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
  if (effectiveStatus.value === 'expired') {
    return '实例已自动回收，请重新启动'
  }
  if (isFailed.value) {
    return '实例不可访问，请重新启动'
  }
  if (effectiveStatus.value === 'destroyed') {
    return '实例已销毁，请重新启动'
  }
  return props.instance.access_url || '--'
})

const openButtonLabel = computed(() => {
  if (props.opening) return '正在打开...'
  if (isWaiting.value) return '等待实例就绪'
  if (isFailed.value) return '实例不可用'
  return '打开目标'
})

const isSharedInstance = computed(() => props.instance?.share_scope === 'shared')

const sharedStrategyLabel = computed(() => {
  if (props.instanceSharing === 'shared' || isSharedInstance.value) {
    return '共享实例'
  }
  if (props.instanceSharing === 'per_team') {
    return '队伍共享'
  }
  return ''
})

const canExtend = computed(
  () => !isSharedInstance.value && canOpen.value && (props.instance?.remaining_extends ?? 0) > 0
)

const restartButtonLabel = computed(() => {
  if (props.creating) return '正在创建实例...'
  if (props.instanceSharing === 'shared' || isSharedInstance.value) {
    return '重新进入共享靶机'
  }
  return '重启实例'
})

const startButtonLabel = computed(() => {
  if (props.creating) return '正在创建实例...'
  if (props.challengeSolved) {
    return props.instanceSharing === 'shared' ? '重新进入共享靶机' : '重启实例'
  }
  return props.instanceSharing === 'shared' ? '进入共享靶机' : '启动靶机'
})
</script>

<template>
  <section class="instance-shell tool-group">
    <div class="instance-kicker">
      Instance
    </div>
    <h2 class="instance-title">
      靶机实例
    </h2>

    <div
      v-if="loading && !instance"
      class="instance-note"
    >
      正在同步当前题目的实例状态...
    </div>

    <div v-else-if="instance">
      <div class="instance-hero">
        <div class="instance-meta-label">
          剩余时间
        </div>
        <div
          class="instance-time"
          :class="isUrgent ? 'instance-time--urgent' : ''"
        >
          {{ remainingLabel }}
        </div>
        <div class="instance-created">
          创建于 {{ createdAtLabel }}
        </div>
        <div
          v-if="sharedStrategyLabel"
          class="instance-created"
        >
          {{ sharedStrategyLabel }}
        </div>
      </div>

      <div class="instance-grid">
        <div class="instance-stat">
          <span>状态</span>
          <strong :class="statusClass">{{ statusLabel }}</strong>
        </div>
        <div class="instance-stat">
          <span>{{ isSharedInstance ? '实例管理' : '剩余延时' }}</span>
          <strong>{{ remainingExtendsLabel }}</strong>
        </div>
        <div class="instance-stat instance-stat--full">
          <span>访问地址</span>
          <strong class="instance-access">{{ accessLabel }}</strong>
        </div>
      </div>

      <div
        v-if="isWaiting"
        class="instance-callout instance-callout--warning"
      >
        <div>实例正在排队创建，系统会自动刷新状态。</div>
        <div>{{ queueLabel }}</div>
        <div>{{ etaLabel }}</div>
        <div v-if="progressLabel">
          {{ progressLabel }}
        </div>
      </div>
      <div
        v-else-if="isReclaimingState"
        class="instance-callout instance-callout--success"
      >
        <div>
          {{
            effectiveStatus === 'expired'
              ? '实例已到期，系统已自动回收当前环境。'
              : '实例已结束，可直接重新启动。'
          }}
        </div>
        <div>如需继续验证，可直接重启实例。</div>
      </div>
      <div
        v-else-if="isFailed"
        class="instance-callout instance-callout--danger"
      >
        <div>
          {{
            props.instance?.status === 'failed'
              ? '实例启动失败，当前目标不可访问。'
              : '实例运行异常，当前目标不可访问。'
          }}
        </div>
        <div>可直接重启实例，系统会为你申请新的环境。</div>
      </div>

      <div
        class="tool-actions"
        :class="{ 'tool-actions--single': isRestartable }"
      >
        <button
          v-if="isRestartable"
          type="button"
          class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="creating"
          @click="emit('start')"
        >
          {{ restartButtonLabel }}
        </button>
        <template v-else>
          <button
            type="button"
            class="ui-btn ui-btn--secondary disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="!canOpen || opening"
            @click="emit('open')"
          >
            {{ openButtonLabel }}
          </button>
          <button
            v-if="!isSharedInstance"
            type="button"
            class="ui-btn ui-btn--secondary disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="!canExtend || extending"
            @click="emit('extend')"
          >
            {{ extending ? '延时中...' : '延时' }}
          </button>
          <button
            v-if="!isSharedInstance"
            type="button"
            class="ui-btn ui-btn--danger disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="destroying"
            @click="emit('destroy')"
          >
            {{ destroying ? '销毁中...' : '销毁' }}
          </button>
          <div
            v-if="isSharedInstance"
            class="instance-note instance-note--managed"
          >
            共享实例由系统统一保活与回收。
          </div>
        </template>
      </div>
    </div>

    <div v-else>
      <div class="instance-note">
        <div>
          {{
            props.instanceSharing === 'shared'
              ? '该题使用共享实例，再次启动会进入同一环境并自动刷新有效期。'
              : '实例会在当前题目页右侧保持可见，便于一边读题一边打开目标、延时或重启。'
          }}
        </div>
        <div>
          {{
            props.challengeSolved
              ? '题目已解出后仍可继续起环境验证，重复正确提交不会重复计分。'
              : '默认有效期 2 小时。'
          }}
        </div>
      </div>
      <button
        type="button"
        class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
        :disabled="creating"
        @click="emit('start')"
      >
        {{ startButtonLabel }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.instance-shell {
  --line-soft: color-mix(in srgb, oklch(38% 0.014 252) 12%, transparent);
  --line-strong: color-mix(in srgb, oklch(38% 0.014 252) 20%, transparent);
  --text-main: oklch(24% 0.014 252);
  --text-subtle: oklch(49% 0.016 252);
  --text-faint: oklch(61% 0.012 252);
  --brand: oklch(52% 0.12 254);
  --warning: oklch(68% 0.14 82);
  --danger: oklch(58% 0.16 28);
  --font-sans: var(--font-family-sans);
  --font-mono: var(--font-family-mono);
  margin-top: 26px;
  padding-top: 26px;
  border-top: 1px solid var(--line-soft);
  font-family: var(--font-sans);
}

.instance-shell,
.instance-shell button {
  font-family: var(--font-sans);
}

.instance-note--managed {
  margin: 0;
}

.instance-time,
.instance-access {
  font-family: var(--font-mono) !important;
}

.instance-kicker {
  font-size: var(--font-size-11);
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--brand) 68%, var(--text-faint));
}

.instance-title {
  margin: 10px 0 0;
  font-size: var(--font-size-18);
  color: var(--text-main);
}

.instance-hero {
  margin-top: 18px;
  padding-left: 16px;
  border-left: 2px solid color-mix(in srgb, var(--brand) 24%, transparent);
}

.instance-meta-label {
  font-size: var(--font-size-11);
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.instance-time {
  margin-top: 8px;
  color: var(--text-main);
  font: 700 28px/1 var(--font-mono);
}

.instance-time--urgent {
  color: var(--warning);
}

.instance-created {
  margin-top: 8px;
  font-size: var(--font-size-12);
  color: var(--text-faint);
}

.instance-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 18px;
}

.instance-stat {
  padding-bottom: 12px;
  border-bottom: 1px solid var(--line-soft);
}

.instance-stat--full {
  grid-column: 1 / -1;
}

.instance-stat span {
  display: block;
  font-size: var(--font-size-12);
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.instance-stat strong {
  display: block;
  margin-top: 8px;
  font-size: var(--font-size-16);
  color: var(--text-main);
}

.instance-access {
  font-family: var(--font-mono);
  font-size: var(--font-size-14);
  line-height: 1.6;
  word-break: break-all;
}

.instance-note {
  margin-top: 16px;
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--text-subtle);
}

.instance-note div + div {
  margin-top: 8px;
}

.instance-callout {
  margin-top: 16px;
  border-left: 2px solid currentColor;
  padding-left: 12px;
  font-size: var(--font-size-12);
  line-height: 1.7;
}

.instance-callout--warning {
  color: var(--warning);
}

.instance-callout--danger {
  color: var(--danger);
}

.instance-callout--success {
  color: color-mix(in srgb, var(--color-success) 80%, var(--text-main));
}

.tool-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: 18px;
}

.tool-actions--single {
  grid-template-columns: minmax(0, 1fr);
}

.tool-actions .ui-btn,
.instance-shell > .ui-btn {
  width: 100%;
  min-height: 48px;
  border-radius: 14px;
}

@media (max-width: 1024px) {
  .instance-grid,
  .tool-actions {
    grid-template-columns: minmax(0, 1fr);
  }
}

:global([data-theme='dark']) .instance-shell {
  --line-soft: color-mix(in srgb, var(--color-border-default) 78%, transparent);
  --line-strong: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  --text-main: var(--color-text-primary);
  --text-subtle: var(--color-text-secondary);
  --text-faint: color-mix(in srgb, var(--color-text-secondary) 82%, var(--color-bg-base));
  --brand: color-mix(in srgb, var(--color-primary) 88%, var(--color-text-primary));
}

@media (prefers-reduced-motion: reduce) {
  button,
  button::before,
  button::after {
    transition-duration: 0.01ms !important;
  }
}
</style>
