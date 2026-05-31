<template>
  <ScoreboardDetailWorkspace
    :contest="contest"
    :page="page"
    :rows="rows"
    :scoreboard="scoreboard"
    :total="total"
    :total-pages="totalPages"
    :loading="loading"
    :refreshing="refreshing"
    :supports-realtime="supportsRealtime"
    :accent-style="accentStyle"
    :empty-title="emptyTitle"
    :empty-description="emptyDescription"
    :top-score="topScore"
    :solved-count="solvedCount"
    :get-status-label="getStatusLabel"
    :format-date-time="formatDateTime"
    :format-contest-window="formatContestWindow"
    :get-row-class="getRowClass"
    :get-rank-pill-class="getRankPillClass"
    :get-status-copy="getStatusCopy"
    :change-page="changePage"
    :load-scoreboard="loadScoreboard"
  />
</template>

<script setup lang="ts">
import { toRef } from 'vue'

import { useScoreboardDetailPage } from '@/features/scoreboard'
import { ScoreboardDetailWorkspace } from '@/widgets/scoreboard-detail-workspace'

const props = defineProps<{
  contestId: string
}>()

const {
  contest,
  page,
  rows,
  scoreboard,
  total,
  totalPages,
  loading,
  refreshing,
  supportsRealtime,
  accentStyle,
  emptyTitle,
  emptyDescription,
  topScore,
  solvedCount,
  getStatusLabel,
  formatDateTime,
  formatContestWindow,
  getRowClass,
  getRankPillClass,
  getStatusCopy,
  changePage,
  loadScoreboard,
} = useScoreboardDetailPage(toRef(props, 'contestId'))
</script>
