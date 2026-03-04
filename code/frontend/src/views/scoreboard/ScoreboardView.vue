<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">排行榜</h1>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[var(--color-border-default)]">
      <table class="w-full">
        <thead class="bg-[var(--color-bg-surface)]">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">排名</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">队伍/用户</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">得分</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">解题数</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in mockData"
            :key="index"
            class="border-b border-[var(--color-border-subtle)] transition-colors duration-100 hover:bg-[var(--color-bg-elevated)]"
            :class="getRowClass(index + 1)"
          >
            <td class="px-4 py-3 font-mono font-bold text-[var(--color-text-primary)]">{{ index + 1 }}</td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ item.name }}</td>
            <td class="px-4 py-3 font-mono text-sm text-[var(--color-text-primary)]">{{ item.score }}</td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ item.solved }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const loading = ref(false)
const mockData = ref([
  { name: 'Binary Wizards', score: 2450, solved: 8 },
  { name: 'Null Pointers', score: 2100, solved: 7 },
  { name: 'Stack Overflow', score: 1850, solved: 7 }
])

function getRowClass(rank: number): string {
  if (rank === 1) return 'bg-amber-500/5 border-l-2 border-amber-400'
  if (rank === 2) return 'bg-slate-300/5 border-l-2 border-slate-400'
  if (rank === 3) return 'bg-orange-400/5 border-l-2 border-orange-400'
  return ''
}
</script>
