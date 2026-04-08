<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import StudentTimelinePage from '@/components/dashboard/student/StudentTimelinePage.vue'
import SkillRadar from '@/components/common/SkillRadar.vue'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherEvidenceData,
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { toRadarScores } from '@/utils/skillProfile'

type StudentInsightSection = 'all' | 'overview' | 'recommendations' | 'writeups' | 'manual-review' | 'evidence'

const props = defineProps<{
  student: TeacherStudentItem | null
  progress: MyProgressData | null
  profile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: TeacherEvidenceData | null
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
  moderateWriteup: [payload: { submissionId: string; action: 'recommend' | 'unrecommend' | 'hide' | 'restore' }]
  reviewManualReview: [payload: { submissionId: string; reviewStatus: 'approved' | 'rejected'; reviewComment?: string }]
  changeWriteupPage: [page: number]
}>()

const radarScores = computed(() => toRadarScores(props.profile))
const publishedWriteupSubmissions = computed(() =>
  props.writeupSubmissions.filter(
    (item) => item.submission_status === 'published' || item.submission_status === 'submitted'
  )
)
const publishedRecommendedWriteupCount = computed(() =>
  publishedWriteupSubmissions.value.filter((item) => item.is_recommended).length
)
const publishedChallengeCount = computed(
  () => new Set(publishedWriteupSubmissions.value.map((item) => String(item.challenge_id))).size
)
const approvedManualReviewCount = computed(() =>
  props.manualReviewSubmissions.filter((item) => item.review_status === 'approved').length
)
const manualReviewComment = ref('')

watch(
  () => props.activeManualReview,
  (value) => {
    manualReviewComment.value = value?.review_comment ?? ''
  },
  { immediate: true },
)

function openChallenge(challengeId: string): void {
  emit('openChallenge', challengeId)
}

function openManualReview(submissionId: string): void {
  emit('openManualReview', submissionId)
}

function changeWriteupPage(page: number): void {
  emit('changeWriteupPage', page)
}

function visibilityStatusLabel(status: TeacherSubmissionWriteupItemData['visibility_status']): string {
  return status === 'hidden' ? '已隐藏' : '已公开'
}

function visibilityStatusClass(status: TeacherSubmissionWriteupItemData['visibility_status']): string {
  return status === 'hidden' ? 'writeup-chip writeup-chip--warning' : 'writeup-chip writeup-chip--success'
}

function manualReviewStatusLabel(status: TeacherManualReviewSubmissionItemData['review_status']): string {
  switch (status) {
    case 'approved':
      return '已通过'
    case 'rejected':
      return '已驳回'
    default:
      return '待审核'
  }
}

function manualReviewStatusClass(status: TeacherManualReviewSubmissionItemData['review_status']): string {
  switch (status) {
    case 'approved':
      return 'writeup-chip writeup-chip--success'
    case 'rejected':
      return 'writeup-chip writeup-chip--warning'
    default:
      return 'writeup-chip writeup-chip--muted'
  }
}

function submitManualReview(reviewStatus: 'approved' | 'rejected'): void {
  if (!props.activeManualReview) return
  emit('reviewManualReview', {
    submissionId: props.activeManualReview.id,
    reviewStatus,
    reviewComment: manualReviewComment.value.trim() || undefined,
  })
}

function isSectionVisible(section: Exclude<StudentInsightSection, 'all'>): boolean {
  return !props.activeSection || props.activeSection === 'all' || props.activeSection === section
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
        <div
          v-if="isSectionVisible('overview')"
          class="insight-overview-layout grid gap-6 lg:grid-cols-[1.15fr_0.85fr]"
        >
          <SectionCard>
            <div
              class="insight-rate-row insight-rate-panel rounded-2xl px-5 py-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <p class="insight-rate-panel__label text-xs uppercase tracking-[0.2em] text-[var(--color-text-secondary)]">
                Solved Rate
              </p>
              <p class="insight-rate-panel__value mt-2 text-3xl font-semibold text-[var(--color-primary)]">
                {{
                  progress?.total_challenges
                    ? Math.round(
                        ((progress.solved_challenges ?? 0) / progress.total_challenges) * 100
                      )
                    : 0
                }}%
              </p>
            </div>
          </SectionCard>

          <SectionCard title="能力画像" subtitle="以雷达图观察当前能力维度分布。">
            <div class="mt-4">
              <SkillRadar :scores="radarScores" />
            </div>
          </SectionCard>
        </div>

        <SectionCard
          v-if="isSectionVisible('recommendations')"
          title="推荐训练任务"
          subtitle="根据当前能力薄弱维度筛出的优先训练题目。"
        >
          <AppEmpty
            v-if="recommendations.length === 0"
            title="暂无推荐题目"
            description="当前画像还没有生成新的推荐训练任务。"
            icon="BookOpen"
          />

          <div v-else class="mt-5 grid gap-3 lg:grid-cols-2">
            <AppCard
              v-for="item in recommendations"
              :key="item.challenge_id"
              as="button"
              variant="action"
              accent="primary"
              interactive
              class="text-left"
              @click="openChallenge(item.challenge_id)"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h5 class="font-semibold text-[var(--color-text-primary)]">{{ item.title }}</h5>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">{{ item.reason }}</p>
                </div>
                <span
                  class="rounded-full px-2.5 py-1 text-xs font-medium"
                  :class="difficultyClass(item.difficulty)"
                >
                  {{ difficultyLabel(item.difficulty) }}
                </span>
              </div>
              <div
                class="mt-3 inline-flex items-center gap-1 text-sm font-medium text-[var(--color-primary)]"
              >
                查看题目
                <ArrowRight class="h-4 w-4" />
              </div>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard
          v-if="isSectionVisible('writeups')"
          title="发布的题解"
          subtitle="按发布时间查看当前学员已发布题解。"
        >
          <AppEmpty
            v-if="publishedWriteupSubmissions.length === 0"
            title="暂无已发布题解"
            description="当前学员还没有发布到题解区的内容。"
            icon="FileText"
          />

          <template v-else>
            <div class="writeup-kpi-grid">
              <article class="insight-kpi-card writeup-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">已发布题解</div>
                <div class="insight-kpi-value">{{ publishedWriteupSubmissions.length }}</div>
                <div class="insight-kpi-hint">当前学员已发布的题解数量</div>
              </article>
              <article class="insight-kpi-card writeup-kpi-card insight-kpi-card--success">
                <div class="insight-kpi-label">对应题目</div>
                <div class="insight-kpi-value">{{ publishedChallengeCount }}</div>
                <div class="insight-kpi-hint">覆盖到的题目总数</div>
              </article>
              <article class="insight-kpi-card writeup-kpi-card insight-kpi-card--warning">
                <div class="insight-kpi-label">推荐中</div>
                <div class="insight-kpi-value">{{ publishedRecommendedWriteupCount }}</div>
                <div class="insight-kpi-hint">发布题解中被标记为推荐的数量</div>
              </article>
            </div>

            <section class="writeup-directory mt-5">
              <header class="writeup-directory-head">
                <span>题目</span>
                <span>题解</span>
                <span>状态</span>
                <span>发布时间</span>
                <span>操作</span>
              </header>

              <article
                v-for="item in publishedWriteupSubmissions"
                :key="item.id"
                class="writeup-directory-row"
              >
                <div class="writeup-directory-cell">
                  <div class="writeup-directory-challenge">{{ item.challenge_title }}</div>
                </div>

                <div class="writeup-directory-cell">
                  <div class="writeup-directory-title">{{ item.title }}</div>
                  <div class="writeup-directory-preview">{{ item.content_preview || '暂无摘要' }}</div>
                </div>

                <div class="writeup-directory-cell">
                  <div class="writeup-directory-status">
                    <span class="writeup-chip writeup-chip--muted">已发布</span>
                    <span :class="visibilityStatusClass(item.visibility_status)">
                      {{ visibilityStatusLabel(item.visibility_status) }}
                    </span>
                    <span
                      v-if="item.is_recommended"
                      class="writeup-chip writeup-chip--primary"
                    >
                      推荐题解
                    </span>
                  </div>
                </div>

                <div class="writeup-directory-cell writeup-directory-time">
                  {{ new Date(item.published_at || item.updated_at).toLocaleString('zh-CN') }}
                </div>

                <div class="writeup-directory-cell writeup-directory-action">
                  <button
                    type="button"
                    class="writeup-open-link inline-flex items-center gap-1 font-medium"
                    @click="openChallenge(item.challenge_id)"
                  >
                    查看题目
                    <ArrowRight class="h-4 w-4" />
                  </button>
                </div>
              </article>
            </section>

            <div class="writeup-pagination mt-4">
              <PagePaginationControls
                :page="writeupPage"
                :total-pages="writeupTotalPages"
                :total="writeupTotal"
                total-label="发布题解总数"
                :disabled="writeupPaginationLoading"
                show-jump
                @change-page="changeWriteupPage"
              />
            </div>
          </template>
        </SectionCard>

        <SectionCard
          v-if="isSectionVisible('manual-review')"
          title="审核题解"
          subtitle="查看该学员待教师评阅的题解内容。"
        >
          <AppEmpty
            v-if="manualReviewSubmissions.length === 0"
            title="暂无题解审核提交"
            description="当前学员还没有需要教师处理的题解审核内容。"
            icon="ClipboardCheck"
          />

          <template v-else>
            <div class="grid gap-3 md:grid-cols-3">
              <article class="insight-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">待处理</div>
                <div class="insight-kpi-value">{{ manualReviewSubmissions.length }}</div>
                <div class="insight-kpi-hint">当前分析页展示的题解审核提交数</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--warning">
                <div class="insight-kpi-label">待审核</div>
                <div class="insight-kpi-value">{{ manualReviewSubmissions.filter((item) => item.review_status === 'pending').length }}</div>
                <div class="insight-kpi-hint">尚未给出审核结果的提交</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--success">
                <div class="insight-kpi-label">已通过</div>
                <div class="insight-kpi-value">{{ approvedManualReviewCount }}</div>
                <div class="insight-kpi-hint">已经通过审核的题解提交</div>
              </article>
            </div>

            <div class="mt-5 grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
              <div class="grid gap-3">
                <AppCard
                  v-for="item in manualReviewSubmissions"
                  :key="item.id"
                  variant="panel"
                  accent="neutral"
                >
                  <div class="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <div class="text-sm font-semibold text-[var(--color-text-primary)]">
                        {{ item.challenge_title }}
                      </div>
                      <div class="mt-1 text-sm text-[var(--color-text-secondary)]">
                        {{ item.answer_preview || '暂无答案摘要' }}
                      </div>
                    </div>
                    <span :class="manualReviewStatusClass(item.review_status)">
                      {{ manualReviewStatusLabel(item.review_status) }}
                    </span>
                  </div>

                  <div class="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs text-[var(--color-text-secondary)]">
                    <span>提交于 {{ new Date(item.submitted_at).toLocaleString('zh-CN') }}</span>
                    <button
                      type="button"
                      class="inline-flex items-center gap-1 font-medium text-[var(--color-primary)]"
                      @click="openManualReview(item.id)"
                    >
                      {{ activeManualReview?.id === item.id ? '刷新详情' : '查看审核' }}
                      <ArrowRight class="h-4 w-4" />
                    </button>
                  </div>
                </AppCard>
              </div>

              <AppCard variant="panel" accent="neutral">
                <div v-if="manualReviewLoading" class="space-y-3">
                  <div class="insight-skeleton-line h-5 w-32 animate-pulse rounded" />
                  <div class="insight-skeleton-block h-24 animate-pulse rounded-2xl" />
                  <div class="insight-skeleton-block h-24 animate-pulse rounded-2xl" />
                </div>

                <AppEmpty
                  v-else-if="!activeManualReview"
                  title="选择一条题解审核提交"
                  description="点击左侧卡片查看完整内容并进行审核。"
                  icon="ClipboardList"
                />

                <template v-else>
                  <div class="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <div class="journal-eyebrow">Writeup Review</div>
                      <h4 class="mt-2 text-lg font-semibold text-[var(--color-text-primary)]">
                        {{ activeManualReview.challenge_title }}
                      </h4>
                      <div class="mt-2 text-sm text-[var(--color-text-secondary)]">
                        {{ activeManualReview.student_name || activeManualReview.student_username }}
                      </div>
                    </div>
                    <span :class="manualReviewStatusClass(activeManualReview.review_status)">
                      {{ manualReviewStatusLabel(activeManualReview.review_status) }}
                    </span>
                  </div>

                  <div class="insight-answer-panel mt-5 rounded-2xl px-4 py-4">
                    <div class="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--color-text-secondary)]">
                      题解内容
                    </div>
                    <p class="mt-3 whitespace-pre-wrap text-sm leading-7 text-[var(--color-text-primary)]">
                      {{ activeManualReview.answer }}
                    </p>
                  </div>

                  <label class="mt-5 block">
                    <span class="text-sm font-medium text-[var(--color-text-primary)]">审核意见</span>
                    <textarea
                      v-model="manualReviewComment"
                      rows="5"
                      class="challenge-input insight-manual-input mt-3 w-full rounded-2xl border px-4 py-3 text-sm leading-7 transition-colors focus:outline-none"
                      placeholder="记录你的判定依据、补充建议或要求学员修改的点。"
                    />
                  </label>

                  <div class="mt-5 flex flex-wrap items-center justify-between gap-3">
                    <div class="text-xs text-[var(--color-text-secondary)]">
                      最近更新：{{ new Date(activeManualReview.updated_at).toLocaleString('zh-CN') }}
                    </div>
                    <div class="flex flex-wrap gap-3">
                      <button
                        type="button"
                        class="challenge-btn-outline insight-outline-action"
                        :disabled="manualReviewSaving || activeManualReview.review_status !== 'pending'"
                        @click="submitManualReview('rejected')"
                      >
                        {{ manualReviewSaving ? '提交中...' : '驳回并说明' }}
                      </button>
                      <button
                        type="button"
                        class="challenge-btn-primary rounded-xl px-5 py-3 text-sm font-medium text-white transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                        :disabled="manualReviewSaving || activeManualReview.review_status !== 'pending'"
                        @click="submitManualReview('approved')"
                      >
                        {{ manualReviewSaving ? '提交中...' : '审核通过' }}
                      </button>
                    </div>
                  </div>
                </template>
              </AppCard>
            </div>
          </template>
        </SectionCard>

        <SectionCard
          v-if="isSectionVisible('evidence')"
          title="攻防证据链"
          subtitle="教师按关键动作查看该学员的利用过程。"
        >
          <AppEmpty
            v-if="!evidence || evidence.events.length === 0"
            title="暂无证据链数据"
            description="当前学员还没有可用于复盘的攻击过程记录。"
            icon="NotebookText"
          />

          <template v-else>
            <div class="grid gap-3 md:grid-cols-4">
              <article class="insight-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">总事件数</div>
                <div class="insight-kpi-value">{{ evidence.summary.total_events }}</div>
                <div class="insight-kpi-hint">纳入教师复盘的动作总数</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--warning">
                <div class="insight-kpi-label">利用请求</div>
                <div class="insight-kpi-value">{{ evidence.summary.proxy_request_count }}</div>
                <div class="insight-kpi-hint">经平台代理的利用请求次数</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--success">
                <div class="insight-kpi-label">提交次数</div>
                <div class="insight-kpi-value">{{ evidence.summary.submit_count }}</div>
                <div class="insight-kpi-hint">当前题目的提交动作统计</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">成功次数</div>
                <div class="insight-kpi-value">{{ evidence.summary.success_count }}</div>
                <div class="insight-kpi-hint">提交命中或利用成功的次数</div>
              </article>
            </div>

            <div class="mt-5 space-y-3">
              <AppCard
                v-for="(event, index) in evidence.events"
                :key="`${event.type}-${event.challenge_id}-${event.timestamp}-${index}`"
                variant="panel"
                accent="neutral"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <div class="text-sm font-semibold text-[var(--color-text-primary)]">{{ event.title }}</div>
                    <div class="mt-1 text-sm text-[var(--color-text-secondary)]">{{ event.detail }}</div>
                    <div class="mt-2 flex flex-wrap gap-2 text-xs text-[var(--color-text-secondary)]">
                      <span
                        class="insight-meta-pill rounded-full border px-2.5 py-1"
                      >
                        {{ String(event.meta?.event_stage || 'trace') }}
                      </span>
                      <span
                        v-if="typeof event.meta?.method === 'string'"
                        class="insight-meta-pill rounded-full border px-2.5 py-1"
                      >
                        {{ String(event.meta?.method) }}
                      </span>
                    </div>
                  </div>
                  <div class="text-right text-xs text-[var(--color-text-secondary)]">
                    <div>{{ new Date(event.timestamp).toLocaleDateString('zh-CN') }}</div>
                    <div class="mt-1">{{ new Date(event.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) }}</div>
                  </div>
                </div>
              </AppCard>
            </div>
          </template>
        </SectionCard>

        <StudentTimelinePage v-if="isSectionVisible('evidence')" :timeline="timeline" />
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
  --color-primary-soft: color-mix(in srgb, var(--journal-accent) 8%, transparent);
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
  align-items: start;
}

.insight-rate-row {
  margin-top: 0;
}

.insight-rate-panel {
  border: 0;
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  background: transparent;
  box-shadow: none;
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 0.1rem 0.7rem;
}

.insight-rate-panel__label,
.insight-rate-panel__value {
  margin: 0;
}

.insight-progress-item {
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  border-radius: 0;
  background: transparent;
}

.insight-progress-track {
  background: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.insight-preview,
.insight-answer-panel {
  border: 0;
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 28%, transparent);
  border-radius: 0;
  background: color-mix(in srgb, var(--journal-surface-subtle) 48%, transparent);
}

.insight-meta-pill {
  border-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.insight-manual-input {
  border-color: color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-ink);
}

.insight-manual-input::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 84%, transparent);
}

.insight-manual-input:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 34%, transparent);
}

.insight-outline-action {
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 90%, transparent);
  color: var(--journal-ink);
}

.insight-outline-action:hover,
.insight-outline-action:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 30%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.insight-outline-action:disabled {
  cursor: not-allowed;
  opacity: 0.56;
}

.writeup-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.35rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 600;
}

.writeup-chip--primary {
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-accent-strong);
}

.writeup-chip--success {
  background: color-mix(in srgb, var(--color-success) 14%, transparent);
  color: color-mix(in srgb, var(--color-success) 82%, var(--journal-ink));
}

.writeup-chip--warning {
  background: color-mix(in srgb, var(--color-warning) 16%, transparent);
  color: color-mix(in srgb, var(--color-warning) 82%, var(--journal-ink));
}

.writeup-chip--muted {
  background: color-mix(in srgb, var(--journal-border) 36%, transparent);
  color: var(--journal-muted);
}

.writeup-directory {
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 84%, transparent);
}

.writeup-kpi-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.85rem;
}

.writeup-kpi-card {
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  border-radius: 16px;
  background: linear-gradient(
    160deg,
    color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
  );
  box-shadow: 0 12px 22px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent);
  padding: 0.85rem 0.95rem 0.8rem;
}

.writeup-directory-head,
.writeup-directory-row {
  display: grid;
  grid-template-columns:
    minmax(0, 1.2fr)
    minmax(0, 2fr)
    minmax(0, 1.2fr)
    minmax(0, 1.35fr)
    minmax(108px, 0.9fr);
  gap: 0.9rem;
  align-items: start;
}

.writeup-directory-head {
  padding: 0.7rem 0.15rem 0.62rem;
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-directory-row {
  padding: 0.9rem 0.15rem;
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 84%, transparent);
}

.writeup-directory-cell {
  min-width: 0;
}

.writeup-directory-challenge {
  font-size: 0.86rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.writeup-directory-title {
  font-size: 0.86rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.writeup-directory-preview {
  margin-top: 0.35rem;
  line-height: 1.6;
  font-size: 0.8rem;
  color: var(--journal-muted);
}

.writeup-directory-status {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.writeup-directory-time {
  font-size: 0.8rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.writeup-directory-action {
  display: flex;
  justify-content: flex-end;
}

.writeup-pagination {
  padding-top: 0.35rem;
}

.writeup-open-link {
  min-height: 34px;
  padding: 0 0.75rem;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 28%, var(--teacher-divider));
  color: var(--journal-accent-strong);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    color 160ms ease;
}

.writeup-open-link:hover,
.writeup-open-link:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 46%, var(--teacher-divider));
  background: color-mix(in srgb, var(--journal-accent) 16%, transparent);
  color: var(--journal-accent);
  outline: none;
}

@media (max-width: 1023px) {
  .writeup-kpi-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .writeup-directory-head {
    display: none;
  }

  .writeup-directory-row {
    grid-template-columns: 1fr;
    gap: 0.7rem;
  }

  .writeup-directory-action {
    justify-content: flex-start;
  }
}

@media (max-width: 767px) {
  .writeup-kpi-grid {
    grid-template-columns: 1fr;
  }
}

:deep(.section-card) {
  padding: 0.95rem 0.2rem 0.7rem;
  border: 0;
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

:deep(.section-card__header) {
  margin-bottom: 1rem;
  border-bottom: 1px dashed color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  padding-bottom: 0.75rem;
}

:deep(.section-card__body) {
  padding-left: 0;
}

.insight-kpi-grid {
  align-items: stretch;
}

.insight-kpi-card {
  border: 0;
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  border-radius: 0;
  background: transparent;
  padding: 0.75rem 0.1rem 0.65rem;
  box-shadow: none;
}

.insight-kpi-card--primary {
  border-top: 2px solid color-mix(in srgb, var(--journal-accent) 36%, transparent);
}

.insight-kpi-card--success {
  border-top: 2px solid color-mix(in srgb, var(--color-success) 34%, transparent);
}

.insight-kpi-card--warning {
  border-top: 2px solid color-mix(in srgb, var(--color-warning) 34%, transparent);
}

.insight-kpi-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.insight-kpi-value {
  margin-top: 0.45rem;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.5;
  color: var(--journal-ink);
}

.insight-kpi-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}
</style>
