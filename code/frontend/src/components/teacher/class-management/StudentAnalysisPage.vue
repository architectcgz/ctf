<script setup lang="ts">
import { computed } from 'vue'
import type { TeacherAttackSessionQuery } from '@/api/teacher'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherAttackSessionResponseData,
  TeacherEvidenceData,
  TeacherClassItem,
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import StudentAnalysisOverviewHeroPanel from '@/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue'
import StudentInsightPanel from '@/components/teacher/StudentInsightPanel.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

const props = defineProps<{
  classes: TeacherClassItem[]
  students: TeacherStudentItem[]
  selectedClassName: string
  selectedStudentId: string
  selectedStudent: TeacherStudentItem | null
  loadingClasses: boolean
  loadingStudents: boolean
  loadingDetails: boolean
  error: string | null
  progress: MyProgressData | null
  skillProfile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: TeacherEvidenceData | null
  attackSessions: TeacherAttackSessionResponseData | null
  reviewChallengeOptions: Array<{ value: string; label: string }>
  reviewWorkspaceLoading: boolean
  reviewWorkspaceQuery: TeacherAttackSessionQuery
  writeupSubmissions: TeacherSubmissionWriteupItemData[]
  writeupPage: number
  writeupTotal: number
  writeupTotalPages: number
  writeupPaginationLoading: boolean
  manualReviewSubmissions: TeacherManualReviewSubmissionItemData[]
  activeManualReview: TeacherManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
  solvedRate: number
  weakDimensions: string[]
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openClassStudents: []
  openReportExport: []
  openReviewArchive: []
  exportReviewArchive: []
  selectClass: [className: string]
  selectStudent: [studentId: string]
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
  updateReviewWorkspaceFilters: [payload: Partial<TeacherAttackSessionQuery>]
}>()

type WorkspaceTab =
  | 'overview'
  | 'recommendations'
  | 'writeups'
  | 'evidence'
  | 'timeline'

interface WorkspaceTabItem {
  key: WorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

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

const workspaceTabOrder = workspaceTabs.map((tab) => tab.key) as WorkspaceTab[]
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } = useUrlSyncedTabs<WorkspaceTab>({
  orderedTabs: workspaceTabOrder,
  defaultTab: 'overview',
})
const activeWorkspaceTab = computed(
  () => workspaceTabs.find((tab) => tab.key === activeTab.value) ?? workspaceTabs[0]
)
</script>

<template>
  <div class="workspace-shell journal-eyebrow-text">
    <nav class="workspace-tabbar top-tabs"
      role="tablist"
      aria-label="学员分析标签页"
    >
      <button
        v-for="(tab, index) in workspaceTabs"
        :id="tab.buttonId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        class="workspace-tab top-tab"
        :class="{ active: activeTab === tab.key }"
        type="button"
        role="tab"
        :tabindex="activeTab === tab.key ? 0 : -1"
        :aria-selected="activeTab === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        @click="selectTab(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="content-pane">
        <div
          v-if="error"
          class="workspace-alert"
          role="alert"
          aria-live="polite"
        >
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
            :active-section="activeTab"
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

.context-rail {
  min-width: 0;
  padding: var(--space-workspace-content-padding, var(--space-7))
    var(--space-workspace-content-padding, var(--space-7))
    var(--space-workspace-content-padding, var(--space-7)) 0;
  border-left: 1px solid color-mix(in srgb, var(--teacher-divider) 80%, transparent);
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

.rail-stack {
  display: grid;
  gap: var(--space-4);
}

.class-switch-list {
  display: grid;
  gap: var(--space-2-5);
}

.class-switch-list--scroll,
.student-directory--scroll {
  max-height: min(34vh, 21rem);
  overflow: auto;
  padding-right: 0.25rem;
}

.class-switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  min-height: 2.75rem;
  padding: var(--space-3) var(--space-1);
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  border-radius: 0;
  background: transparent;
  font-size: var(--font-size-0-86);
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.class-switch:hover,
.class-switch:focus-visible,
.class-switch.active {
  border-bottom-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  color: var(--journal-accent-strong);
  outline: none;
}

.student-directory {
  display: grid;
}

.student-directory-head,
.student-directory-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-3);
  align-items: center;
}

.student-directory-head {
  padding: 0 var(--space-1-5) var(--space-2);
  border-bottom: 1px solid var(--teacher-divider);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.student-directory-row {
  padding: var(--space-3-5) var(--space-1);
  border-bottom: 1px solid var(--teacher-divider);
  background: transparent;
  text-align: left;
  transition: background 160ms ease;
}

.student-directory-row:hover,
.student-directory-row:focus-visible,
.student-directory-row.active {
  background: color-mix(in srgb, var(--journal-accent) 6%, transparent);
  outline: none;
}

.student-directory-main {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.student-directory-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.9rem;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 16%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
  flex-shrink: 0;
}

.student-directory-copy {
  min-width: 0;
}

.student-directory-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-0-92);
  font-weight: 600;
  color: var(--journal-ink);
}

.student-directory-meta {
  margin-top: var(--space-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-0-80);
  color: var(--journal-muted);
}

.student-directory-state {
  font-size: var(--font-size-0-78);
  font-weight: 600;
  color: var(--workspace-brand-ink);
}

.student-directory-skeleton {
  display: grid;
  gap: var(--space-3);
}

.student-directory-skeleton__item {
  height: 3.5rem;
  border-radius: 0.95rem;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-border) 80%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  animation: pulse 1.35s ease-in-out infinite;
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

.quick-action--primary {
  border-bottom-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
}

.context-block {
  position: relative;
}

.context-block + .context-block {
  margin-top: var(--space-2-5);
}

.context-block--actions {
  margin-top: auto;
}

.quick-action__main {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.quick-action__main span:last-child {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.quick-action__main strong {
  font-size: var(--font-size-0-90);
  font-weight: 600;
  color: inherit;
}

.quick-action__main small {
  font-size: var(--font-size-0-78);
  line-height: 1.55;
  color: var(--journal-muted);
}

.quick-action__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.2rem;
  height: 2.2rem;
  border-radius: 0.85rem;
  flex-shrink: 0;
}

.quick-action__icon--neutral {
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.quick-action__icon--warning {
  border: 1px solid color-mix(in srgb, var(--color-warning) 18%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: color-mix(in srgb, var(--color-warning) 82%, var(--journal-ink));
}

.quick-action__icon--primary {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
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

@keyframes pulse {
  0%,
  100% {
    opacity: 0.58;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 1279px) {
  .context-rail {
    padding: 0 var(--space-workspace-content-padding, var(--space-7))
      var(--space-workspace-content-padding, var(--space-7));
    border-left: 0;
    border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 80%, transparent);
  }
}

@media (max-width: 767px) {
  .top-tabs,
  .content-pane,
  .context-rail {
    padding-left: var(--space-5);
    padding-right: var(--space-5);
  }

  .top-tabs {
    gap: var(--space-4-5);
  }
}
</style>
