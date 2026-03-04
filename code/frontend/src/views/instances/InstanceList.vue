<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">我的实例</h1>
      <span class="text-sm text-[var(--color-text-secondary)]">运行中: {{ runningCount }}/{{ maxInstances }}</span>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="instance in instances"
        :key="instance.id"
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-5"
      >
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">{{ instance.challenge_title }}</h3>
            <div class="flex gap-2">
              <span class="rounded bg-[#06b6d4]/10 px-2 py-0.5 text-xs font-medium text-[#06b6d4]">
                {{ instance.category }}
              </span>
              <span class="rounded bg-[#34d399]/10 px-2 py-0.5 text-xs font-medium text-[#34d399]">
                {{ instance.difficulty }}
              </span>
            </div>
          </div>

          <div class="flex items-center gap-2 text-sm">
            <span :class="getStatusClass(instance.status)">●</span>
            <span class="text-[var(--color-text-secondary)]">{{ getStatusLabel(instance.status) }}</span>
          </div>

          <div v-if="instance.status === 'running'" class="space-y-2 text-sm">
            <div class="flex items-center justify-between">
              <span class="text-[var(--color-text-secondary)]">地址:</span>
              <div class="flex items-center gap-2">
                <span class="font-mono text-[var(--color-text-primary)]">{{ instance.access_url || (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '') }}</span>
                <button
                  @click="copyAddress(instance.access_url || (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : ''))"
                  class="rounded px-2 py-1 text-xs text-[var(--color-primary)] hover:bg-[var(--color-primary)]/10"
                >
                  复制
                </button>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-[var(--color-text-secondary)]">剩余:</span>
              <span class="font-mono" :class="instance.remaining < WARNING_THRESHOLD_SECONDS ? 'text-[#f59e0b] font-semibold' : 'text-[var(--color-text-primary)]'">
                {{ formatTime(instance.remaining) }}
              </span>
            </div>
          </div>

          <div class="flex gap-3">
            <button
              v-if="instance.status === 'running'"
              @click="extendTime(instance.id)"
              :disabled="instance.remaining_extends <= 0"
              class="rounded-lg border border-[var(--color-border-default)] bg-[#21262d] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition-colors duration-150 hover:bg-[#30363d] disabled:opacity-50 disabled:cursor-not-allowed"
            >
              延时 +{{ EXTEND_DURATION_SECONDS / 60 }}min ({{ instance.remaining_extends }})
            </button>
            <button
              @click="confirmDestroy(instance.id)"
              class="rounded-lg border border-[#ef4444]/20 bg-[#ef4444]/10 px-4 py-2 text-sm font-medium text-[#f87171] transition-colors duration-150 hover:bg-[#ef4444]/20"
            >
              销毁
            </button>
          </div>
        </div>
      </div>

      <div v-if="instances.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
        <div class="text-[var(--color-text-muted)] mb-4">暂无运行中的实例</div>
        <router-link to="/student/challenges" class="text-[var(--color-primary)] hover:underline">
          前往靶场列表创建实例
        </router-link>
      </div>
    </div>

    <!-- 超时提醒弹窗 -->
    <div
      v-if="showWarning"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click.self="showWarning = false"
    >
      <div class="w-full max-w-md rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-xl">
        <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">实例即将过期</h3>
        <p class="mt-2 text-sm text-[var(--color-text-secondary)]">
          实例 "{{ warningInstance?.challenge_title }}" 剩余时间不足 5 分钟，是否延长？
        </p>
        <div class="mt-6 flex justify-end gap-3">
          <button
            @click="showWarning = false"
            class="rounded-lg px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-hover)]"
          >
            取消
          </button>
          <button
            @click="extendFromWarning"
            class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--color-primary-hover)]"
          >
            延长 {{ EXTEND_DURATION_SECONDS / 60 }} 分钟
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getMyInstances, destroyInstance as apiDestroyInstance, extendInstance } from '@/api/instance'
import type { InstanceListItem, InstanceStatus } from '@/api/contracts'

// 常量配置
const MAX_INSTANCES = 3
const WARNING_THRESHOLD_SECONDS = 300
const EXTEND_DURATION_SECONDS = 1800

// 本地 ViewModel 类型
interface InstanceViewModel extends InstanceListItem {
  remaining: number
}

const loading = ref(false)
const maxInstances = MAX_INSTANCES
const instances = ref<InstanceViewModel[]>([])

const showWarning = ref(false)
const warningInstance = ref<InstanceViewModel | null>(null)
const warnedInstances = new Set<string>()

let timer: number | null = null

const runningCount = computed(() => instances.value.filter(i => i.status === 'running').length)

function getStatusLabel(status: InstanceStatus): string {
  const labels: Record<InstanceStatus, string> = {
    pending: '等待中',
    creating: '创建中',
    running: '运行中',
    expired: '已过期',
    destroying: '销毁中',
    destroyed: '已销毁',
    failed: '失败',
    crashed: '崩溃'
  }
  return labels[status] || status
}

function getStatusClass(status: InstanceStatus): string {
  const classes: Record<InstanceStatus, string> = {
    pending: 'text-[#f59e0b]',
    creating: 'text-[#f59e0b]',
    running: 'text-[#22c55e]',
    expired: 'text-[var(--color-text-muted)]',
    destroying: 'text-[#f59e0b]',
    destroyed: 'text-[var(--color-text-muted)]',
    failed: 'text-[#ef4444]',
    crashed: 'text-[#ef4444]'
  }
  return classes[status] || 'text-[var(--color-text-muted)]'
}

function formatTime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

async function copyAddress(address: string) {
  try {
    await navigator.clipboard.writeText(address)
    // TODO: 显示成功提示
  } catch (error) {
    console.error('复制失败:', error)
    // TODO: 显示错误提示
  }
}

async function extendTime(id: string) {
  try {
    const result = await extendInstance(id)
    const instance = instances.value.find(i => i.id === id)
    if (instance) {
      instance.remaining = calculateRemaining(result.expires_at)
      instance.remaining_extends = result.remaining_extends
      warnedInstances.delete(id)
    }
  } catch (error) {
    console.error('延时失败:', error)
    // TODO: 显示错误提示
  }
}

async function confirmDestroy(id: string) {
  if (!confirm('确定要销毁该实例吗？此操作不可恢复。')) {
    return
  }
  try {
    await apiDestroyInstance(id)
    instances.value = instances.value.filter(i => i.id !== id)
    warnedInstances.delete(id)
  } catch (error) {
    console.error('销毁失败:', error)
    // TODO: 显示错误提示
  }
}

async function extendFromWarning() {
  if (warningInstance.value) {
    await extendTime(warningInstance.value.id)
  }
  showWarning.value = false
}

function handleEscKey(event: KeyboardEvent) {
  if (event.key === 'Escape' && showWarning.value) {
    showWarning.value = false
  }
}

function calculateRemaining(expiresAt: string): number {
  return Math.max(0, Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000))
}

function updateCountdown() {
  const now = Date.now()
  instances.value.forEach(instance => {
    if (instance.status === 'running') {
      instance.remaining = Math.max(0, Math.floor((new Date(instance.expires_at).getTime() - now) / 1000))

      if (instance.remaining < WARNING_THRESHOLD_SECONDS && !warnedInstances.has(instance.id)) {
        warnedInstances.add(instance.id)
        warningInstance.value = instance
        showWarning.value = true
      }
    }
  })
}

onMounted(async () => {
  loading.value = true
  try {
    const data = await getMyInstances()
    instances.value = data.map(item => ({
      ...item,
      remaining: calculateRemaining(item.expires_at)
    }))
  } catch (error) {
    console.error('加载实例失败:', error)
  } finally {
    loading.value = false
  }
  timer = window.setInterval(updateCountdown, 1000)
  window.addEventListener('keydown', handleEscKey)
})

onUnmounted(() => {
  if (timer !== null) {
    clearInterval(timer)
    timer = null
  }
  window.removeEventListener('keydown', handleEscKey)
})
</script>

