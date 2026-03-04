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
            <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">{{ instance.title }}</h3>
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
                <span class="font-mono text-[var(--color-text-primary)]">{{ instance.address }}</span>
                <button
                  @click="copyAddress(instance.address)"
                  class="rounded px-2 py-1 text-xs text-[var(--color-primary)] hover:bg-[var(--color-primary)]/10"
                >
                  复制
                </button>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-[var(--color-text-secondary)]">剩余:</span>
              <span class="font-mono" :class="instance.remaining < 300 ? 'text-[#f59e0b] font-semibold' : 'text-[var(--color-text-primary)]'">
                {{ formatTime(instance.remaining) }}
              </span>
            </div>
          </div>

          <div class="flex gap-3">
            <button
              v-if="instance.status === 'running'"
              @click="extendTime(instance.id)"
              class="rounded-lg border border-[var(--color-border-default)] bg-[#21262d] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition-colors duration-150 hover:bg-[#30363d]"
            >
              延时 +30min
            </button>
            <button
              @click="destroyInstance(instance.id)"
              class="rounded-lg border border-[#ef4444]/20 bg-[#ef4444]/10 px-4 py-2 text-sm font-medium text-[#f87171] transition-colors duration-150 hover:bg-[#ef4444]/20"
            >
              销毁
            </button>
          </div>
        </div>
      </div>

      <div v-if="instances.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
        <div class="text-[var(--color-text-muted)]">暂无运行中的实例</div>
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
          实例 "{{ warningInstance?.title }}" 剩余时间不足 5 分钟，是否延长？
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
            延长 30 分钟
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const loading = ref(false)
const maxInstances = 3
const instances = ref([
  { id: '1', title: 'SQL 注入基础', category: 'Web', difficulty: '简单', status: 'running', address: '10.10.1.42:8080', remaining: 4425 },
  { id: '2', title: '栈溢出入门', category: 'Pwn', difficulty: '中等', status: 'running', address: '10.10.1.43:9999', remaining: 272 }
])

const showWarning = ref(false)
const warningInstance = ref<any>(null)
const warnedInstances = new Set<string>()

let timer: number | null = null

const runningCount = computed(() => instances.value.filter(i => i.status === 'running').length)

function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    running: '运行中',
    starting: '启动中',
    stopping: '停止中'
  }
  return labels[status] || status
}

function getStatusClass(status: string): string {
  if (status === 'running') return 'text-[#22c55e]'
  if (status === 'starting') return 'text-[#f59e0b]'
  return 'text-[var(--color-text-muted)]'
}

function formatTime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

function copyAddress(address: string) {
  navigator.clipboard.writeText(address)
}

function extendTime(id: string) {
  const instance = instances.value.find(i => i.id === id)
  if (instance) {
    instance.remaining += 1800
    warnedInstances.delete(id)
  }
}

function destroyInstance(id: string) {
  instances.value = instances.value.filter(i => i.id !== id)
  warnedInstances.delete(id)
}

function extendFromWarning() {
  if (warningInstance.value) {
    extendTime(warningInstance.value.id)
  }
  showWarning.value = false
}

function updateCountdown() {
  instances.value.forEach(instance => {
    if (instance.status === 'running' && instance.remaining > 0) {
      instance.remaining--

      if (instance.remaining < 300 && !warnedInstances.has(instance.id)) {
        warnedInstances.add(instance.id)
        warningInstance.value = instance
        showWarning.value = true
      }
    }
  })
}

onMounted(() => {
  timer = window.setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

