<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">竞赛管理</h1>
      <button class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors duration-150 hover:bg-[#06b6d4]">
        创建竞赛
      </button>
    </div>

    <div class="flex gap-3">
      <input
        type="text"
        placeholder="搜索竞赛..."
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none focus:ring-1 focus:ring-[#0891b2]/50"
      />
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[var(--color-border-default)]">
      <table class="w-full">
        <thead class="bg-[var(--color-bg-surface)]">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">标题</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">状态</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">开始时间</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">参赛人数</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="contest in mockContests"
            :key="contest.id"
            class="border-b border-[var(--color-border-subtle)] transition-colors duration-100 hover:bg-[var(--color-bg-elevated)]"
          >
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ contest.title }}</td>
            <td class="px-4 py-3 text-sm">
              <span class="rounded px-2 py-0.5 text-xs font-medium" :class="getStatusClass(contest.status)">
                {{ contest.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-secondary)]">{{ contest.startTime }}</td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-secondary)]">{{ contest.participants }}</td>
            <td class="px-4 py-3 text-sm">
              <button class="text-[var(--color-primary)] hover:text-[#06b6d4]">编辑</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const loading = ref(false)
const mockContests = ref([
  { id: '1', title: '2026 春季校园 CTF', status: '进行中', startTime: '2026-03-15 09:00', participants: 32 }
])

function getStatusClass(status: string): string {
  if (status === '进行中') return 'bg-[var(--color-primary)]/10 text-[#06b6d4]'
  return 'bg-[#30363d] text-[var(--color-text-secondary)]'
}
</script>
