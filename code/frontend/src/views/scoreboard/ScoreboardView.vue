<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">排行榜</h1>
      <button
        type="button"
        class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)]"
        @click="loadScoreboard"
      >
        刷新
      </button>
    </div>

    <div v-if="!contestId" class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-8 text-center text-[var(--color-text-muted)]">
      请通过 `?contestId=竞赛ID` 指定要查看的竞赛排行榜。
    </div>

    <div v-else class="space-y-4">
      <div v-if="contestTitle" class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-4">
        <div class="text-sm text-[var(--color-text-secondary)]">当前竞赛</div>
        <div class="mt-1 text-lg font-semibold text-[var(--color-text-primary)]">{{ contestTitle }}</div>
        <div v-if="frozen" class="mt-2 inline-flex rounded bg-[#f59e0b]/20 px-2 py-1 text-xs font-medium text-[#f59e0b]">排行榜已冻结</div>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
      </div>

      <div v-else-if="rows.length === 0" class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-8 text-center text-[var(--color-text-muted)]">
        暂无排行榜数据
      </div>

      <div v-else class="overflow-hidden rounded-lg border border-[var(--color-border-default)]">
        <table class="w-full">
          <thead class="bg-[var(--color-bg-surface)]">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">排名</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">队伍</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">得分</th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">解题数</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in rows"
              :key="item.team_id"
              class="border-b border-[var(--color-border-subtle)] transition-colors duration-100 hover:bg-[var(--color-bg-elevated)]"
              :class="getRowClass(item.rank)"
            >
              <td class="px-4 py-3 font-mono font-bold text-[var(--color-text-primary)]">{{ item.rank }}</td>
              <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ item.team_name }}</td>
              <td class="px-4 py-3 font-mono text-sm text-[var(--color-text-primary)]">{{ item.score }}</td>
              <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ item.solved_count }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { getScoreboard } from '@/api/contest'
import type { ScoreboardRow } from '@/api/contracts'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const toast = useToast()
const loading = ref(false)
const rows = ref<ScoreboardRow[]>([])
const contestTitle = ref('')
const frozen = ref(false)

const contestId = computed(() => {
  const queryId = route.query.contestId ?? route.query.contest_id
  return typeof queryId === 'string' ? queryId : ''
})

async function loadScoreboard() {
  if (!contestId.value) {
    rows.value = []
    return
  }

  loading.value = true
  try {
    const data = await getScoreboard(contestId.value, { page: 1, page_size: 100 })
    rows.value = data.scoreboard.list
    contestTitle.value = data.contest.title
    frozen.value = data.frozen
  } catch (error) {
    toast.error('加载排行榜失败')
  } finally {
    loading.value = false
  }
}

function getRowClass(rank: number): string {
  if (rank === 1) return 'bg-amber-500/5 border-l-2 border-amber-400'
  if (rank === 2) return 'bg-slate-300/5 border-l-2 border-slate-400'
  if (rank === 3) return 'bg-orange-400/5 border-l-2 border-orange-400'
  return ''
}

onMounted(() => {
  loadScoreboard()
})
</script>
