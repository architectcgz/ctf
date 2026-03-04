<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[#e6edf3]">我的实例</h1>
      <span class="text-sm text-[#8b949e]">运行中: {{ runningCount }}/{{ maxInstances }}</span>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="instance in mockInstances"
        :key="instance.id"
        class="rounded-lg border border-[#30363d] bg-[#161b22] p-5"
      >
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-[#e6edf3]">{{ instance.title }}</h3>
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
            <span class="text-[#8b949e]">{{ getStatusLabel(instance.status) }}</span>
          </div>

          <div v-if="instance.status === 'running'" class="space-y-2 text-sm">
            <div class="flex items-center justify-between">
              <span class="text-[#8b949e]">地址:</span>
              <span class="font-mono text-[#e6edf3]">{{ instance.address }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-[#8b949e]">剩余:</span>
              <span class="font-mono" :class="instance.remaining < 300 ? 'text-[#f59e0b]' : 'text-[#e6edf3]'">
                {{ formatTime(instance.remaining) }}
              </span>
            </div>
          </div>

          <div class="flex gap-3">
            <button
              v-if="instance.status === 'running'"
              class="rounded-lg border border-[#30363d] bg-[#21262d] px-4 py-2 text-sm font-medium text-[#e6edf3] transition-colors duration-150 hover:bg-[#30363d]"
            >
              延时 +30min
            </button>
            <button
              class="rounded-lg border border-[#ef4444]/20 bg-[#ef4444]/10 px-4 py-2 text-sm font-medium text-[#f87171] transition-colors duration-150 hover:bg-[#ef4444]/20"
            >
              {{ instance.status === 'expired' ? '重新启动' : '销毁' }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="mockInstances.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
        <div class="text-[#6e7681]">暂无运行中的实例</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const loading = ref(false)
const maxInstances = 3
const mockInstances = ref([
  { id: '1', title: 'SQL 注入基础', category: 'Web', difficulty: '简单', status: 'running', address: '10.10.1.42:8080', remaining: 4425 },
  { id: '2', title: '栈溢出入门', category: 'Pwn', difficulty: '中等', status: 'running', address: '10.10.1.43:9999', remaining: 272 },
  { id: '3', title: 'RSA 基础', category: 'Crypto', difficulty: '简单', status: 'expired', address: '', remaining: 0 }
])

const runningCount = computed(() => mockInstances.value.filter(i => i.status === 'running').length)

function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    running: '运行中',
    expired: '已过期',
    pending: '排队中'
  }
  return labels[status] || status
}

function getStatusClass(status: string): string {
  if (status === 'running') return 'text-[#22c55e]'
  if (status === 'expired') return 'text-[#6e7681]'
  return 'text-[#0891b2]'
}

function formatTime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}
</script>
