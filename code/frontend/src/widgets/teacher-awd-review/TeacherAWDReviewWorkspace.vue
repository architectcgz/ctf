<script setup lang="ts">
import { computed } from 'vue'
import {
  Trophy,
} from 'lucide-vue-next'

import type {
  TeacherAWDReviewArchiveData,
  TeacherAWDReviewAttackItemData,
  TeacherAWDReviewRoundItemData,
  TeacherAWDReviewSelectedRoundData,
  TeacherAWDReviewServiceItemData,
  TeacherAWDReviewTeamItemData,
  TeacherAWDReviewTrafficItemData,
} from '@/api/contracts'
import TeacherAWDReviewAnalysisSection from '@/components/teacher/awd-review/TeacherAWDReviewAnalysisSection.vue'
import TeacherAWDReviewEvidenceGrid from '@/components/teacher/awd-review/TeacherAWDReviewEvidenceGrid.vue'
import TeacherAWDReviewRoundSelector from '@/components/teacher/awd-review/TeacherAWDReviewRoundSelector.vue'
import TeacherAWDReviewTeamDrawer from '@/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue'
import TeacherAWDReviewStatusChip from './TeacherAWDReviewStatusChip.vue'
import TeacherAWDReviewWorkspaceActions from './TeacherAWDReviewWorkspaceActions.vue'
import TeacherAWDReviewSummaryPanel from './TeacherAWDReviewSummaryPanel.vue'
import TeacherAWDReviewSurfaceShell from './TeacherAWDReviewSurfaceShell.vue'
import TeacherAWDReviewWorkspaceHeader from './TeacherAWDReviewWorkspaceHeader.vue'
import TeacherAWDReviewWorkspaceState from './TeacherAWDReviewWorkspaceState.vue'
import {
  buildTeacherAwdReviewSummaryItems,
  TEACHER_AWD_REVIEW_WORKSPACE_COPY,
} from './model/presentation'

type ExportKind = 'archive' | 'report'

interface SummaryStats {
  roundCount: number
  teamCount: number
  serviceCount: number
  attackCount: number
  trafficCount: number
}

const props = defineProps<{
  polling: boolean
  loading: boolean
  error: string | null
  review: TeacherAWDReviewArchiveData | null
  exporting: ExportKind | null
  activeContestTitle: string
  activeSummaryTitle: string
  summaryStats: SummaryStats
  timelineRounds: TeacherAWDReviewRoundItemData[]
  selectedRoundNumber?: number
  selectedRound?: TeacherAWDReviewSelectedRoundData
  selectedTeam: TeacherAWDReviewTeamItemData | null
  selectedTeamServices: TeacherAWDReviewServiceItemData[]
  selectedTeamAttacks: TeacherAWDReviewAttackItemData[]
  selectedTeamTraffic: TeacherAWDReviewTrafficItemData[]
  canExportReport: boolean
  contestStatusLabel: (status: string) => string
  formatServiceRef: (serviceId?: string) => string
}>()

const emit = defineEmits<{
  openIndex: []
  exportArchive: []
  exportReport: []
  loadReview: []
  setRound: [roundNumber?: number]
  openTeam: [team: TeacherAWDReviewTeamItemData]
  closeTeam: []
}>()

const summaryItems = computed(() =>
  buildTeacherAwdReviewSummaryItems(props.summaryStats, props.polling)
)
</script>

<template>
  <TeacherAWDReviewSurfaceShell section-class="teacher-review-workspace">
    <div class="teacher-page">
      <TeacherAWDReviewWorkspaceHeader
        :overline="TEACHER_AWD_REVIEW_WORKSPACE_COPY.overline"
        :title="TEACHER_AWD_REVIEW_WORKSPACE_COPY.title"
        header-class="awd-review-detail-header"
        overline-class="awd-review-detail-overline"
      >
        <template #description>
          <span class="awd-review-detail-contest-title">{{ activeContestTitle }}</span>
          <span> · </span>
          {{ TEACHER_AWD_REVIEW_WORKSPACE_COPY.descriptionSuffix }}
        </template>

        <template #actions>
          <TeacherAWDReviewWorkspaceActions
            :loading="loading"
            :has-review="Boolean(review)"
            :exporting="exporting"
            :can-export-report="canExportReport"
            @open-index="emit('openIndex')"
            @export-archive="emit('exportArchive')"
            @export-report="emit('exportReport')"
          />
        </template>
      </TeacherAWDReviewWorkspaceHeader>

      <TeacherAWDReviewSummaryPanel
        :title="activeSummaryTitle"
        :items="summaryItems"
        summary-class="awd-review-summary"
      >
        <template #title-prefix>
          <Trophy class="h-4 w-4" />
        </template>
        <template #title-suffix>
          <TeacherAWDReviewStatusChip
            v-if="review"
            :status="review.contest.status || ''"
            :label="contestStatusLabel(review.contest.status || '')"
          />
        </template>
      </TeacherAWDReviewSummaryPanel>

      <TeacherAWDReviewRoundSelector
        :rounds="timelineRounds"
        :selected-round-number="selectedRoundNumber"
        @set-round="emit('setRound', $event)"
      />

      <TeacherAWDReviewWorkspaceState
        :loading="loading"
        :error="error"
        :has-review="Boolean(review)"
        @load-review="emit('loadReview')"
      >
        <template v-if="review">
          <TeacherAWDReviewAnalysisSection
            :active-summary-title="activeSummaryTitle"
            :rounds="review.rounds"
            :selected-round="selectedRound"
            :team-count="summaryStats.teamCount"
            @set-round="emit('setRound', $event)"
            @open-team="emit('openTeam', $event)"
          />

          <TeacherAWDReviewEvidenceGrid
            v-if="selectedRound"
            :selected-round="selectedRound"
            :format-service-ref="formatServiceRef"
          />
        </template>
      </TeacherAWDReviewWorkspaceState>

      <TeacherAWDReviewTeamDrawer
        :visible="Boolean(selectedTeam)"
        :team="selectedTeam"
        :services="selectedTeamServices"
        :attacks="selectedTeamAttacks"
        :traffic="selectedTeamTraffic"
        @close="emit('closeTeam')"
      />
    </div>
  </TeacherAWDReviewSurfaceShell>
</template>

<style scoped>
.teacher-review-workspace {
  --awd-review-surface: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base));
  --awd-review-surface-subtle: color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 92%, var(--color-bg-base));
  --awd-review-line: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 74%, transparent);
  --awd-review-line-strong: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 84%, transparent);
  --awd-review-text: var(--journal-ink, var(--color-text-primary));
  --awd-review-muted: var(--journal-muted, var(--color-text-secondary));
  --awd-review-faint: color-mix(in srgb, var(--color-text-muted) 88%, var(--awd-review-muted));
  --awd-review-primary: var(--journal-accent, var(--color-primary));
  --awd-review-primary-strong: var(--journal-accent-strong, var(--color-primary-hover));
  --awd-review-success: var(--color-success);
  --awd-review-warning: var(--color-warning);
  --awd-review-danger: var(--color-danger);
  --awd-review-blue: var(--color-brand-swatch-blue);
  gap: var(--space-6);
}

.awd-review-detail-header {
  gap: var(--space-4);
}

.awd-review-summary {
  --metric-panel-border: color-mix(in srgb, var(--awd-review-primary) 14%, var(--awd-review-line));
}

.awd-review-status-text {
  color: var(--awd-review-muted);
}
</style>
