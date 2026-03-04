<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[#e6edf3]">竞赛管理</h1>
      <button class="rounded-lg bg-[#0891b2] px-4 py-2 text-sm font-medium text-white transition-colors duration-150 hover:bg-[#06b6d4]">
        创建竞赛
      </button>
    </div>

    <div class="flex gap-3">
      <input
        type="text"
        placeholder="搜索竞赛..."
        class="rounded-lg border border-[#30363d] bg-[#0f1117] px-3 py-2 text-sm text-[#e6edf3] placeholder:text-[#6e7681] focus:border-[#0891b2] focus:outline-none focus:ring-1 focus:ring-[#0891b2]/50"
      />
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[#30363d]">
      <table class="w-full">
        <thead class="bg-[#161b22]">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">标题</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">状态</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">开始时间</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">参赛人数</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="contest in mockContests"
            :key="contest.id"
            class="border-b border-[#21262d] transition-colors duration-100 hover:bg-[#1c2128]"
          >
            <td class="px-4 py-3 text-sm text-[#e6edf3]">{{ contest.title }}</td>
            <td class="px-4 py-3 text-sm">
              <span class="rounded px-2 py-0.5 text-xs font-medium" :class="getStatusClass(contest.status)">
                {{ contest.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-[#8b949e]">{{ contest.startTime }}</td>
            <td class="px-4 py-3 text-sm text-[#8b949e]">{{ contest.participants }}</td>
            <td class="px-4 py-3 text-sm">
              <button class="text-[#0891b2] hover:text-[#06b6d4]">编辑</button>
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
  if (status === '进行中') return 'bg-[#0891b2]/10 text-[#06b6d4]'
  return 'bg-[#30363d] text-[#8b949e]'
}
</script>
