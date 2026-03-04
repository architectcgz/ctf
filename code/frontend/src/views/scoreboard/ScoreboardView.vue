<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-[#e6edf3]">排行榜</h1>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[#30363d]">
      <table class="w-full">
        <thead class="bg-[#161b22]">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">排名</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">队伍/用户</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">得分</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">解题数</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(item, index) in mockData"
            :key="index"
            class="border-b border-[#21262d] transition-colors duration-100 hover:bg-[#1c2128]"
            :class="getRowClass(index + 1)"
          >
            <td class="px-4 py-3 font-mono font-bold text-[#e6edf3]">{{ index + 1 }}</td>
            <td class="px-4 py-3 text-sm text-[#e6edf3]">{{ item.name }}</td>
            <td class="px-4 py-3 font-mono text-sm text-[#e6edf3]">{{ item.score }}</td>
            <td class="px-4 py-3 text-sm text-[#e6edf3]">{{ item.solved }}</td>
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
