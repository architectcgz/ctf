<template>
  <div class="workspace-shell journal-eyebrow-text">
    <nav class="workspace-tabbar top-tabs" role="tablist" aria-label="学员分析标签页">
      <button
        v-for="(tab, index) in workspaceTabs"
        :id="tab.buttonId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        class="workspace-tab top-tab"
        :class="{ active: props.activeWorkspaceTab === tab.key }"
        type="button"
        role="tab"
        :tabindex="props.activeWorkspaceTab === tab.key ? 0 : -1"
        :aria-selected="props.activeWorkspaceTab === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        @click="emit('selectWorkspaceTab', tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="content-pane">
      <div v-if="error" class="workspace-alert" role="alert" aria-live="polite">
        <div class="workspace-alert-title">
          学员分析加载失败
        </div>
        <div class="workspace-alert-copy">
          {{ error }}
        </div>
        <div class="workspace-alert-actions">
          <button
            type="button"
            class="quick-action quick-action--compact"
            @click="emit('retry')"
          >
            <span>重试加载</span>
            <span>→</span>
          </button>
        </div>
      </div>

      <section
        :id="activeWorkspaceTab.panelId"
        class="tab-panel section"
        :class="{
          active: true,
          'student-analysis-overview-panel': activeWorkspaceTab.key === 'overview',
        }"
        role="tabpanel"
        :aria-labelledby="activeWorkspaceTab.buttonId"
        aria-hidden="false"
      >
        <template v-if="activeWorkspaceTab.key === 'overview'">
          <StudentAnalysisOverviewHeroPanel
            :selected-student="selectedStudent"
            :progress="progress"
            :solved-rate="solvedRate"
            :weak-dimensions="weakDimensions"
            @open-class-students="emit('openClassStudents')"
            @open-report-export="emit('openReportExport')"
            @open-review-archive="emit('openReviewArchive')"
            @export-review-archive="emit('exportReviewArchive')"
          />
        </template>

        <StudentInsightPanel
          :active-section="props.activeWorkspaceTab"
          :student="selectedStudent"
          :progress="progress"
          :profile="skillProfile"
          :recommendations="recommendations"
          :timeline="timeline"
          :evidence="evidence"
          :attack-sessions="attackSessions"
          :review-challenge-options="reviewChallengeOptions"
          :review-workspace-loading="reviewWorkspaceLoading"
          :review-workspace-query="reviewWorkspaceQuery"
          :writeup-submissions="writeupSubmissions"
          :writeup-page="writeupPage"
          :writeup-total="writeupTotal"
          :writeup-total-pages="writeupTotalPages"
          :writeup-pagination-loading="writeupPaginationLoading"
          :manual-review-submissions="manualReviewSubmissions"
          :active-manual-review="activeManualReview"
          :manual-review-loading="manualReviewLoading"
          :manual-review-saving="manualReviewSaving"
          :loading="loadingDetails"
          empty-text="请先选择一名学生。"
          @open-challenge="emit('openChallenge', $event)"
          @open-manual-review="emit('openManualReview', $event)"
          @moderate-writeup="emit('moderateWriteup', $event)"
          @review-manual-review="emit('reviewManualReview', $event)"
          @change-writeup-page="emit('changeWriteupPage', $event)"
          @update-review-workspace-filters="emit('updateReviewWorkspaceFilters', $event)"
        />
      </section>
    </main>
  </div>
</template>

<style scoped>
.workspace-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-eyebrow-spacing: 0.15em;
  --journal-eyebrow-color: var(--journal-accent-strong);
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --header-control-border: var(--teacher-control-border);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --workspace-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --workspace-shadow-shell: 0 24px 84px
    color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-xl: 28px;
  --workspace-radius-lg: 18px;
}

.content-pane {
  display: grid;
  gap: var(--space-section-gap-compact, var(--space-4));
}

.workspace-alert {
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
  padding: var(--space-4) var(--space-4-5);
}

.workspace-alert-title {
  font-size: var(--font-size-0-92);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-84);
  line-height: 1.65;
  color: var(--journal-muted);
}

.workspace-alert-actions {
  margin-top: var(--space-3);
}

.quick-action {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2-5);
  min-height: 52px;
  padding: var(--space-3) var(--space-1);
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  border-radius: 0;
  background: transparent;
  color: var(--journal-ink);
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.quick-action:hover,
.quick-action:focus-visible {
  border-bottom-color: color-mix(in srgb, var(--workspace-brand) 34%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 6%, transparent);
  color: var(--workspace-brand-ink);
  outline: none;
}

.quick-action--compact {
  min-height: 42px;
}

:deep(.section-card) {
  padding: var(--space-3-5) var(--space-1) var(--space-3);
  border: 0;
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 90%, transparent);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

:deep(.section-card__header) {
  margin-bottom: var(--space-4);
  border-bottom: 1px dashed color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  padding-bottom: var(--space-3);
}

:deep(.section-card__body) {
  padding-left: 0;
}

@media (max-width: 767px) {
  .top-tabs,
  .content-pane {
    padding-left: var(--space-5);
    padding-right: var(--space-5);
  }

  .top-tabs {
    gap: var(--space-4-5);
  }
}
</style>

<script setup lang="ts">
import { computed } from 'vue'

import type {
  AttackSessionResponseData,
  ManualReviewSubmissionDetailData,
  ManualReviewSubmissionItemData,
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  StudentDirectoryItem,
  StudentEvidenceData,
  TimelineEvent,
  WriteupSubmissionItemData,
} from '@/api/contracts'
import { useTabKeyboardNavigation } from '@/shared/lib/keyboard/useTabKeyboardNavigation'
import type { StudentAnalysisWorkspaceTab } from '../model/useStudentAnalysisPage'
import StudentAnalysisOverviewHeroPanel from './StudentAnalysisOverviewHeroPanel.vue'
import StudentInsightPanel from './StudentInsightPanel.vue'

interface ReviewWorkspaceQuery {
  mode?: 'practice' | 'jeopardy' | 'awd'
  challenge_id?: string
  contest_id?: string
  round_id?: string
  result?: 'success' | 'failed' | 'in_progress' | 'unknown'
  with_events?: boolean
  limit?: number
  offset?: number
}

interface WorkspaceTabItem {
  key: StudentAnalysisWorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

const props = defineProps<{
  selectedStudent: StudentDirectoryItem | null
  loadingDetails: boolean
  error: string | null
  progress: MyProgressData | null
  skillProfile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: StudentEvidenceData | null
  attackSessions: AttackSessionResponseData | null
  reviewChallengeOptions: Array<{ value: string; label: string }>
  reviewWorkspaceLoading: boolean
  reviewWorkspaceQuery: ReviewWorkspaceQuery
  activeWorkspaceTab: StudentAnalysisWorkspaceTab
  writeupSubmissions: WriteupSubmissionItemData[]
  writeupPage: number
  writeupTotal: number
  writeupTotalPages: number
  writeupPaginationLoading: boolean
  manualReviewSubmissions: ManualReviewSubmissionItemData[]
  activeManualReview: ManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
  solvedRate: number
  weakDimensions: string[]
}>()

const emit = defineEmits<{
  retry: []
  openClassStudents: []
  openReportExport: []
  openReviewArchive: []
  exportReviewArchive: []
  selectWorkspaceTab: [tab: StudentAnalysisWorkspaceTab]
  openChallenge: [challengeId: string]
  openManualReview: [submissionId: string]
  moderateWriteup: [
    payload: { submissionId: string; action: 'recommend' | 'unrecommend' | 'hide' | 'restore' },
  ]
  reviewManualReview: [
    payload: {
      submissionId: string
      reviewStatus: 'approved' | 'rejected'
      reviewComment?: string
    },
  ]
  changeWriteupPage: [page: number]
  updateReviewWorkspaceFilters: [payload: Partial<ReviewWorkspaceQuery>]
}>()

const workspaceTabs: WorkspaceTabItem[] = [
  {
    key: 'overview',
    label: '学员画像',
    buttonId: 'student-tab-overview',
    panelId: 'student-overview',
  },
  {
    key: 'recommendations',
    label: '推荐任务',
    buttonId: 'student-tab-recommendations',
    panelId: 'student-recommendations',
  },
  {
    key: 'writeups',
    label: '发布的题解',
    buttonId: 'student-tab-writeups',
    panelId: 'student-writeups',
  },
  {
    key: 'evidence',
    label: '证据链',
    buttonId: 'student-tab-evidence',
    panelId: 'student-evidence',
  },
  {
    key: 'timeline',
    label: '训练记录',
    buttonId: 'student-tab-timeline',
    panelId: 'student-timeline',
  },
]

const activeWorkspaceTab = computed(
  () => workspaceTabs.find((tab) => tab.key === props.activeWorkspaceTab) ?? workspaceTabs[0]
)
const workspaceTabOrder = workspaceTabs.map((tab) => tab.key) as StudentAnalysisWorkspaceTab[]
const { setTabButtonRef, handleTabKeydown } =
  useTabKeyboardNavigation<StudentAnalysisWorkspaceTab>({
    orderedTabs: workspaceTabOrder,
    selectTab: (tab) => emit('selectWorkspaceTab', tab),
  })
</script>
