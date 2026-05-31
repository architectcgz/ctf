<template>
  <section class="instance-panel">
    <header class="instance-panel__header">
      <span class="instance-panel__title">我的实例</span>
      <button
        type="button"
        class="ui-btn ui-btn--secondary ui-btn--sm"
        @click="emit('refresh')"
      >
        刷新
      </button>
    </header>

    <div
      v-if="loading"
      class="instance-panel__empty"
    >
      加载中...
    </div>

    <div
      v-else-if="instances.length === 0"
      class="instance-panel__empty"
    >
      暂无运行中的实例
    </div>

    <div
      v-else
      class="instance-panel__list"
    >
      <article
        v-for="instance in instances"
        :key="instance.id"
        class="instance-card"
      >
        <div class="instance-card__body">
          <div class="instance-card__head">
            <div>
              <h3 class="instance-card__title">
                {{ instance.challenge_title }}
              </h3>
              <span
                class="instance-chip"
                :class="getStatusChipClass(instance.status)"
              >
                {{ getStatusLabel(instance.status) }}
              </span>
            </div>
            <div class="instance-card__countdown">
              <div
                :class="getTimeColor(instance.expires_at)"
                class="instance-card__countdown-value"
              >
                {{ formatCountdown(instance.expires_at) }}
              </div>
              <div class="instance-card__countdown-label">
                剩余时间
              </div>
            </div>
          </div>

          <div
            v-if="instance.access_url"
            class="instance-card__access"
          >
            <div class="instance-card__access-row">
              <span class="instance-card__access-label">访问地址：</span>
              <button
                type="button"
                class="instance-card__access-link"
                @click="emit('openTarget', instance.id)"
              >
                {{ formatInstanceAccessDisplay(instance) }}
              </button>
            </div>
          </div>

          <div class="instance-card__actions">
            <button
              v-if="instance.share_scope !== 'shared' && instance.remaining_extends > 0"
              type="button"
              class="ui-btn ui-btn--primary ui-btn--sm"
              @click="emit('extend', instance.id)"
            >
              延时 (剩余 {{ instance.remaining_extends }} 次)
            </button>
            <button
              v-if="instance.share_scope !== 'shared'"
              type="button"
              class="ui-btn ui-btn--danger ui-btn--sm"
              @click="emit('destroy', instance.id)"
            >
              销毁
            </button>
            <div
              v-if="instance.share_scope === 'shared'"
              class="instance-card__managed-note"
            >
              系统托管
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

import {
  formatInstanceAccessDisplay,
  getInstanceRemainingSeconds,
  getInstanceRemainingTone,
  getInstanceStatusLabel,
  getInstanceStatusTone,
} from '@/entities/instance'
import type { InstancePanelItem, InstancePanelStatus } from '@/shared/model/common/instancePanel.types'

const props = withDefaults(
  defineProps<{
    instances: InstancePanelItem[]
    loading?: boolean
  }>(),
  {
    loading: false,
  }
)

const emit = defineEmits<{
  refresh: []
  openTarget: [id: string]
  extend: [id: string]
  destroy: [id: string]
  expiringSoon: [instance: InstancePanelItem]
}>()

const instances = computed(() => props.instances)
const now = ref(Date.now())
const warnedInstances = new Set<string>()
let timer: number | null = null

function formatCountdown(expiresAt: string): string {
  const diff = getInstanceRemainingSeconds(expiresAt, now.value) * 1000

  if (diff <= 0) return '已过期'

  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)
  const seconds = Math.floor((diff % 60000) / 1000)

  if (hours > 0)
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

function getTimeColor(expiresAt: string): string {
  const tone = getInstanceRemainingTone(getInstanceRemainingSeconds(expiresAt, now.value), {
    warningSeconds: 600,
    dangerSeconds: 300,
  })

  if (tone === 'danger') return 'instance-countdown--danger'
  if (tone === 'warning') return 'instance-countdown--warning'
  if (tone === 'success') return 'instance-countdown--success'
  return 'instance-countdown--muted'
}

function checkExpiringSoon() {
  instances.value.forEach((instance) => {
    const expires = new Date(instance.expires_at).getTime()
    const diff = expires - now.value

    if (diff > 0 && diff < 300000 && !warnedInstances.has(instance.id)) {
      warnedInstances.add(instance.id)
      emit('expiringSoon', instance)
    }
  })
}

function getStatusLabel(status: InstancePanelStatus): string {
  return getInstanceStatusLabel(status)
}

function getStatusChipClass(status: InstancePanelStatus): string {
  const tone = getInstanceStatusTone(status)
  if (tone === 'warning') return 'instance-chip--warning'
  if (tone === 'success') return 'instance-chip--success'
  if (tone === 'danger') return 'instance-chip--danger'
  return 'instance-chip--muted'
}

onMounted(() => {
  checkExpiringSoon()
  timer = window.setInterval(() => {
    now.value = Date.now()
    checkExpiringSoon()
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

watch(
  () => props.instances,
  (nextInstances) => {
    const nextIds = new Set(nextInstances.map((instance) => instance.id))
    Array.from(warnedInstances).forEach((id) => {
      if (!nextIds.has(id)) {
        warnedInstances.delete(id)
      }
    })
    checkExpiringSoon()
  },
  { deep: true }
)
</script>

<style scoped>
.instance-panel {
  display: grid;
  gap: 1rem;
  padding: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
}

.instance-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.instance-panel__title {
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--color-text-primary);
}

.instance-panel__empty {
  padding: 2rem 0;
  text-align: center;
  color: var(--color-text-muted);
}

.instance-panel__list {
  display: grid;
  gap: 1rem;
}

.instance-card {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-surface));
}

.instance-card__body {
  display: grid;
  gap: 0.75rem;
  padding: 1rem;
}

.instance-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.instance-card__title {
  font-size: var(--font-size-1-05);
  font-weight: 700;
  color: var(--color-text-primary);
}

.instance-chip {
  display: inline-flex;
  align-items: center;
  margin-top: 0.45rem;
  min-height: 1.6rem;
  padding: 0 0.55rem;
  border-radius: 999px;
  font-size: var(--font-size-0-75);
  font-weight: 700;
}

.instance-chip--muted {
  background: color-mix(in srgb, var(--color-text-muted) 12%, transparent);
  color: var(--color-text-muted);
}

.instance-chip--warning {
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: var(--color-warning);
}

.instance-chip--success {
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  color: var(--color-success);
}

.instance-chip--danger {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
}

.instance-card__countdown {
  text-align: right;
}

.instance-card__countdown-value {
  font-size: var(--font-size-1-05);
  font-weight: 700;
}

.instance-countdown--muted {
  color: var(--color-text-muted);
}

.instance-countdown--danger {
  color: var(--color-danger);
}

.instance-countdown--warning {
  color: var(--color-warning);
}

.instance-countdown--success {
  color: var(--color-success);
}

.instance-card__countdown-label,
.instance-card__managed-note {
  font-size: var(--font-size-0-75);
  color: var(--color-text-muted);
}

.instance-card__access {
  display: grid;
  gap: 0.5rem;
}

.instance-card__access-row {
  font-size: var(--font-size-0-88);
}

.instance-card__access-label {
  color: var(--color-text-secondary);
}

.instance-card__access-link {
  color: var(--color-primary);
  text-decoration: underline;
  text-underline-offset: 0.14em;
}

.instance-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

@media (max-width: 640px) {
  .instance-card__head {
    flex-direction: column;
  }

  .instance-card__countdown {
    text-align: left;
  }
}
</style>
