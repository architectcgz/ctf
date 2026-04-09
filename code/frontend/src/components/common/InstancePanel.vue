<template>
  <ElCard>
    <template #header>
      <div class="flex items-center justify-between">
        <span class="font-semibold">我的实例</span>
        <ElButton
          size="small"
          @click="emit('refresh')"
        >
          刷新
        </ElButton>
      </div>
    </template>

    <div
      v-if="loading"
      class="py-8 text-center text-[var(--color-text-muted)]"
    >
      加载中...
    </div>

    <div
      v-else-if="instances.length === 0"
      class="text-center text-[var(--color-text-muted)] py-8"
    >
      暂无运行中的实例
    </div>

    <div
      v-else
      class="space-y-4"
    >
      <ElCard
        v-for="instance in instances"
        :key="instance.id"
        shadow="never"
        class="border"
      >
        <div class="space-y-3">
          <div class="flex items-start justify-between">
            <div>
              <h3 class="font-semibold text-lg">
                {{ instance.challenge_title }}
              </h3>
              <ElTag
                :type="getStatusColor(instance.status)"
                size="small"
                class="mt-1"
              >
                {{ getStatusLabel(instance.status) }}
              </ElTag>
            </div>
            <div class="text-right">
              <div
                :class="getTimeColor(instance.expires_at)"
                class="text-lg font-semibold"
              >
                {{ formatCountdown(instance.expires_at) }}
              </div>
              <div class="text-xs text-[var(--color-text-muted)]">
                剩余时间
              </div>
            </div>
          </div>

          <div
            v-if="instance.access_url"
            class="space-y-2"
          >
            <div class="text-sm">
              <span class="text-[var(--color-text-secondary)]">访问地址：</span>
              <button
                type="button"
                class="text-primary hover:underline"
                @click="emit('openTarget', instance.id)"
              >
                {{ instance.access_url }}
              </button>
            </div>
          </div>

          <div class="flex gap-2">
            <ElButton
              v-if="instance.share_scope !== 'shared' && instance.remaining_extends > 0"
              size="small"
              type="primary"
              plain
              class="!border-[var(--color-primary)] !bg-[var(--color-primary-soft)] !text-[var(--color-primary-hover)] hover:!bg-[var(--color-primary-soft)]"
              @click="emit('extend', instance.id)"
            >
              延时 (剩余 {{ instance.remaining_extends }} 次)
            </ElButton>
            <ElButton
              v-if="instance.share_scope !== 'shared'"
              size="small"
              type="danger"
              @click="emit('destroy', instance.id)"
            >
              销毁
            </ElButton>
            <div
              v-if="instance.share_scope === 'shared'"
              class="text-xs text-[var(--color-text-muted)]"
            >
              系统托管
            </div>
          </div>
        </div>
      </ElCard>
    </div>
  </ElCard>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { InstanceListItem, InstanceStatus } from '@/api/contracts'

const props = withDefaults(
  defineProps<{
    instances: InstanceListItem[]
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
  expiringSoon: [instance: InstanceListItem]
}>()

const instances = computed(() => props.instances)
const now = ref(Date.now())
const warnedInstances = new Set<string>()
let timer: number | null = null

function formatCountdown(expiresAt: string): string {
  const expires = new Date(expiresAt).getTime()
  const diff = expires - now.value

  if (diff <= 0) return '已过期'

  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)
  const seconds = Math.floor((diff % 60000) / 1000)

  if (hours > 0)
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

function getTimeColor(expiresAt: string): string {
  const expires = new Date(expiresAt).getTime()
  const diff = expires - now.value

  if (diff <= 0) return 'text-[var(--color-text-muted)]'
  if (diff < 300000) return 'text-[var(--color-danger)]'
  if (diff < 600000) return 'text-[var(--color-warning)]'
  return 'text-[var(--color-success)]'
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

function getStatusLabel(status: InstanceStatus): string {
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
  return labels[status]
}

function getStatusColor(status: InstanceStatus): string {
  const colors: Record<InstanceStatus, string> = {
    pending: 'info',
    creating: 'warning',
    running: 'success',
    expired: 'info',
    destroying: 'warning',
    destroyed: 'info',
    failed: 'danger',
    crashed: 'danger',
  }
  return colors[status]
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
