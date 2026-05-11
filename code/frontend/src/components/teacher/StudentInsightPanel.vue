<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import SkillRadar from '@/components/common/SkillRadar.vue'
import StudentTimelinePage from '@/components/dashboard/student/StudentTimelinePage.vue'
import { ChallengeCategoryDifficultyPills } from '@/entities/challenge'
import type { TeacherAttackSessionQuery } from '@/api/teacher'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherAttackSessionResponseData,
  TeacherEvidenceData,
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import { toRadarScores } from '@/utils/skillProfile'
import StudentInsightAttackSessionsSection from '@/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue'
import StudentInsightManualReviewSection from '@/components/teacher/student-insight/StudentInsightManualReviewSection.vue'
import StudentInsightWriteupsSection from '@/components/teacher/student-insight/StudentInsightWriteupsSection.vue'
import type { StudentInsightSection } from '@/components/teacher/student-insight/studentInsightShared'

const props = defineProps<{
  student: TeacherStudentItem | null
  progress: MyProgressData | null
  profile: SkillProfileData | null
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
  loading: boolean
  emptyText?: string
  activeSection?: StudentInsightSection
}>()

const emit = defineEmits<{
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

const radarScores = computed(() => toRadarScores(props.profile))
const rankedProfileDimensions = computed(() =>
  [...(props.profile?.dimensions ?? [])].sort((left, right) => right.value - left.value)
)
const showManualReviewSection = computed(() => props.activeSection === 'manual-review')

function isSectionVisible(section: Exclude<StudentInsightSection, 'all'>): boolean {
  return !props.activeSection || props.activeSection === 'all' || props.activeSection === section
}

function openChallenge(challengeId: string): void {
  emit('openChallenge', challengeId)
}
</script>

<template>
  <div class="student-insight-shell teacher-surface space-y-6">
    <AppEmpty
      v-if="!student && !loading"
      title="尚未选择学员"
      :description="emptyText || '请先选择学员。'"
      icon="GraduationCap"
    />

    <template v-else>
      <div v-if="loading" class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <AppCard variant="panel" accent="neutral">
          <div class="insight-skeleton-line h-6 w-36 animate-pulse rounded" />
          <div class="mt-6 space-y-3">
            <div class="insight-skeleton-block h-16 animate-pulse rounded-xl" />
            <div class="insight-skeleton-block h-16 animate-pulse rounded-xl" />
          </div>
        </AppCard>
        <AppCard variant="panel" accent="neutral">
          <div class="insight-skeleton-block h-[280px] animate-pulse rounded-2xl" />
        </AppCard>
      </div>

      <template v-else-if="student">
        <div v-if="isSectionVisible('overview')" class="insight-overview-layout">
          <SectionCard title="能力雷达" subtitle="左侧雷达图展示当前能力维度分布。">
            <div class="mt-4">
              <SkillRadar :scores="radarScores" />
            </div>
          </SectionCard>

          <SectionCard title="能力比例" subtitle="右侧条状图展示各维度当前分值。">
            <div v-if="rankedProfileDimensions.length > 0" class="insight-dimension-bars mt-4">
              <article
                v-for="item in rankedProfileDimensions"
                :key="item.key"
                class="insight-dimension-item"
              >
                <div class="insight-dimension-item__head">
                  <strong>{{ item.name }}</strong>
                  <span>{{ item.value }}%</span>
                </div>
                <div class="insight-dimension-item__track">
                  <div class="insight-dimension-item__fill" :style="{ width: `${item.value}%` }" />
                </div>
              </article>
            </div>
            <div v-else class="insight-dimension-empty mt-4">暂无画像维度数据</div>
          </SectionCard>
        </div>

        <SectionCard
          v-if="isSectionVisible('recommendations')"
          class="insight-tab-section-card"
          title="推荐训练任务"
          subtitle="根据当前能力薄弱维度筛出的优先训练题目。"
        >
          <AppEmpty
            v-if="recommendations.length === 0"
            title="暂无推荐题目"
            description="当前画像还没有生成新的推荐训练任务。"
            icon="BookOpen"
          />

          <div v-else class="insight-recommendation-list workspace-directory-list">
            <button
              v-for="item in recommendations"
              :key="item.challenge_id"
              type="button"
              class="insight-recommendation-row workspace-directory-grid-row"
              @click="openChallenge(item.challenge_id)"
            >
              <div class="workspace-directory-cell insight-recommendation-main">
                <h5 class="workspace-directory-row-title">
                  {{ item.title }}
                </h5>
                <p class="workspace-directory-row-subtitle">
                  {{ item.summary }}
                </p>
                <p
                  v-if="item.evidence"
                  class="insight-recommendation-evidence"
                >
                  {{ item.evidence }}
                </p>
              </div>
              <div class="insight-recommendation-pills">
                <ChallengeCategoryDifficultyPills
                  :category="item.category"
                  :difficulty="item.difficulty"
                />
              </div>
              <span class="workspace-directory-row-btn insight-recommendation-action">
                <span>查看题目</span>
                <ArrowRight class="h-4 w-4" />
              </span>
            </button>
          </div>
        </SectionCard>

        <StudentInsightWriteupsSection
          v-if="isSectionVisible('writeups')"
          :writeup-submissions="writeupSubmissions"
          :writeup-page="writeupPage"
          :writeup-total="writeupTotal"
          :writeup-total-pages="writeupTotalPages"
          :writeup-pagination-loading="writeupPaginationLoading"
          :manual-review-submissions="manualReviewSubmissions"
          :active-manual-review="activeManualReview"
          :manual-review-loading="manualReviewLoading"
          :manual-review-saving="manualReviewSaving"
          @open-challenge="emit('openChallenge', $event)"
          @open-manual-review="emit('openManualReview', $event)"
          @moderate-writeup="emit('moderateWriteup', $event)"
          @review-manual-review="emit('reviewManualReview', $event)"
          @change-writeup-page="emit('changeWriteupPage', $event)"
        />

        <StudentInsightManualReviewSection
          v-if="showManualReviewSection"
          :manual-review-submissions="manualReviewSubmissions"
          :active-manual-review="activeManualReview"
          :manual-review-loading="manualReviewLoading"
          :manual-review-saving="manualReviewSaving"
          @open-manual-review="emit('openManualReview', $event)"
          @review-manual-review="emit('reviewManualReview', $event)"
        />

        <StudentInsightAttackSessionsSection
          v-if="isSectionVisible('evidence')"
          :attack-sessions="attackSessions"
          :evidence="evidence"
          :review-challenge-options="reviewChallengeOptions"
          :review-workspace-loading="reviewWorkspaceLoading"
          :review-workspace-query="reviewWorkspaceQuery"
          @update-review-workspace-filters="emit('updateReviewWorkspaceFilters', $event)"
        />

        <StudentTimelinePage v-if="isSectionVisible('timeline')" :timeline="timeline" />
      </template>
    </template>
  </div>
</template>

<style scoped>
.student-insight-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
}

.insight-skeleton-line,
.insight-skeleton-block {
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-border) 78%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
}

.insight-overview-layout {
  display: grid;
  gap: var(--space-6);
  grid-template-columns: minmax(0, 1.08fr) minmax(0, 0.92fr);
  align-items: start;
  padding-top: var(--space-3);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
}

.insight-overview-layout :deep(.section-card) {
  border-top: 0;
}

.insight-dimension-bars {
  display: grid;
  gap: var(--space-3-5);
}

.insight-dimension-item {
  display: grid;
  gap: var(--space-2);
}

.insight-dimension-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  color: var(--journal-ink);
}

.insight-dimension-item__head span {
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.insight-dimension-item__track {
  height: 10px;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-border) 36%, transparent);
}

.insight-dimension-item__fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink)),
    color-mix(in srgb, var(--journal-accent) 48%, white) 58%,
    color-mix(in srgb, var(--color-warning) 84%, var(--journal-accent))
  );
}

.insight-dimension-empty {
  font-size: var(--font-size-0-84);
  color: var(--journal-muted);
}

.insight-recommendation-pills {
  display: inline-flex;
  flex: 0 0 auto;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.insight-recommendation-list {
  --workspace-directory-grid-columns: minmax(0, 1fr) auto auto;
  --workspace-directory-shell-padding: var(--space-2) var(--space-4);
  margin-top: var(--space-5);
}

.insight-recommendation-row {
  gap: var(--space-4);
}

.insight-recommendation-main {
  display: grid;
  gap: var(--space-1);
  align-content: center;
}

.insight-recommendation-evidence {
  margin: var(--space-1) 0 0;
  color: color-mix(in srgb, var(--journal-muted) 86%, transparent);
  font-size: var(--font-size-12);
  line-height: 1.65;
}

.insight-recommendation-action {
  justify-self: end;
  white-space: nowrap;
}

:deep(.section-card) {
  padding: var(--space-4) var(--space-1) var(--space-3);
  border: 0;
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

:deep(.section-card__header) {
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom: 1px dashed color-mix(in srgb, var(--teacher-divider) 86%, transparent);
}

:deep(.section-card__body) {
  padding-left: 0;
}

.insight-tab-section-card :deep(.section-card__header) {
  border-bottom: 0;
}

.insight-tab-section-card.section-card {
  border-top: 0;
}

@media (max-width: 767px) {
  .insight-overview-layout {
    grid-template-columns: 1fr;
  }

  .insight-recommendation-list {
    --workspace-directory-grid-columns: minmax(0, 1fr);
  }

  .insight-recommendation-row {
    gap: var(--space-3);
  }

  .insight-recommendation-pills,
  .insight-recommendation-action {
    justify-self: start;
  }
}
</style>
