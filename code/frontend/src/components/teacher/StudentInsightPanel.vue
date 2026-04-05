<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
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
import { getWeakDimensions, toRadarScores } from '@/utils/skillProfile'

const props = defineProps<{
  student: TeacherStudentItem | null
  progress: MyProgressData | null
  profile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: TeacherEvidenceData | null
  writeupSubmissions: TeacherSubmissionWriteupItemData[]
  manualReviewSubmissions: TeacherManualReviewSubmissionItemData[]
  activeManualReview: TeacherManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
  loading: boolean
  emptyText?: string
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
  openManualReview: [submissionId: string]
  moderateWriteup: [payload: { submissionId: string; action: 'recommend' | 'unrecommend' | 'hide' | 'restore' }]
  reviewManualReview: [payload: { submissionId: string; reviewStatus: 'approved' | 'rejected'; reviewComment?: string }]
}>()

const radarScores = computed(() => toRadarScores(props.profile))
const weakDimensions = computed(() => getWeakDimensions(props.profile))
const visibleWriteupCount = computed(() =>
  props.writeupSubmissions.filter((item) => item.visibility_status === 'visible').length
)
const recommendedWriteupCount = computed(() =>
  props.writeupSubmissions.filter((item) => item.is_recommended).length
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

function submissionStatusLabel(status: TeacherSubmissionWriteupItemData['submission_status']): string {
  return status === 'published' || status === 'submitted' ? '已发布' : '草稿'
}

function visibilityStatusLabel(status: TeacherSubmissionWriteupItemData['visibility_status']): string {
  return status === 'hidden' ? '已隐藏' : '已公开'
}

function visibilityStatusClass(status: TeacherSubmissionWriteupItemData['visibility_status']): string {
  return status === 'hidden' ? 'writeup-chip writeup-chip--warning' : 'writeup-chip writeup-chip--success'
}

function emitWriteupModeration(
  submissionId: string,
  action: 'recommend' | 'unrecommend' | 'hide' | 'restore'
): void {
  emit('moderateWriteup', { submissionId, action })
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
          <div class="h-6 w-36 animate-pulse rounded bg-[var(--color-bg-base)]" />
          <div class="mt-6 space-y-3">
            <div class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]" />
            <div class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]" />
          </div>
        </AppCard>
        <AppCard variant="panel" accent="neutral">
          <div class="h-[280px] animate-pulse rounded-2xl bg-[var(--color-bg-base)]" />
        </AppCard>
      </div>

      <template v-else-if="student">
        <div class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
          <SectionCard title="当前学员" subtitle="聚合进度、难度完成情况和薄弱维度。">
            <AppCard
              variant="hero"
              accent="primary"
              eyebrow="Student Snapshot"
              :title="student.name || student.username"
              subtitle="查看当前学员的关键指标和推荐方向。"
            >
              <template #header>
                <span
                  class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
                  style="
                    border-color: color-mix(
                      in srgb,
                      var(--color-primary) 18%,
                      var(--color-border-default)
                    );
                    background-color: var(--color-primary-soft);
                    color: var(--color-primary);
                  "
                >
                  @{{ student.username }}
                </span>
              </template>

              <div class="insight-kpi-grid grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                <article class="insight-kpi-card insight-kpi-card--primary">
                  <div class="insight-kpi-label">总题量</div>
                  <div class="insight-kpi-value">{{ progress?.total_challenges ?? 0 }}</div>
                  <div class="insight-kpi-hint">该学员当前纳入统计的挑战总数</div>
                </article>
                <article class="insight-kpi-card insight-kpi-card--success">
                  <div class="insight-kpi-label">已完成</div>
                  <div class="insight-kpi-value">{{ progress?.solved_challenges ?? 0 }}</div>
                  <div class="insight-kpi-hint">已成功完成的挑战数量</div>
                </article>
                <article class="insight-kpi-card insight-kpi-card--warning">
                  <div class="insight-kpi-label">薄弱维度</div>
                  <div class="insight-kpi-value">
                    {{ weakDimensions.length > 0 ? weakDimensions.join('、') : '暂无' }}
                  </div>
                  <div class="insight-kpi-hint">基于能力画像提炼的风险点</div>
                </article>
                <article class="insight-kpi-card insight-kpi-card--primary">
                  <div class="insight-kpi-label">推荐题目</div>
                  <div class="insight-kpi-value">{{ recommendations.length }}</div>
                  <div class="insight-kpi-hint">可立即布置的补强任务数量</div>
                </article>
              </div>
            </AppCard>

            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div class="rounded-2xl bg-[var(--color-bg-base)] px-5 py-4 text-center">
                <p class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-secondary)]">
                  Solved Rate
                </p>
                <p class="mt-2 text-3xl font-semibold text-[var(--color-primary)]">
                  {{
                    progress?.total_challenges
                      ? Math.round(
                          ((progress.solved_challenges ?? 0) / progress.total_challenges) * 100
                        )
                      : 0
                  }}%
                </p>
              </div>
            </div>

            <div class="mt-6 grid gap-4 xl:grid-cols-2">
              <AppCard
                variant="panel"
                accent="primary"
                eyebrow="分类进度"
                subtitle="按知识方向查看当前完成覆盖率。"
              >
                <div class="mt-4 space-y-3">
                  <div
                    v-for="(value, key) in progress?.by_category || {}"
                    :key="key"
                    class="rounded-lg bg-[var(--color-bg-surface)] px-3 py-3"
                  >
                    <div class="flex items-center justify-between text-sm">
                      <span class="font-medium text-[var(--color-text-primary)]">{{ key }}</span>
                      <span class="text-[var(--color-text-secondary)]"
                        >{{ value.solved }} / {{ value.total }}</span
                      >
                    </div>
                    <div
                      class="mt-2 h-2 overflow-hidden rounded-full bg-[var(--color-border-default)]"
                    >
                      <div
                        class="h-full rounded-full bg-[var(--color-primary)]"
                        :style="{
                          width: `${value.total ? Math.round((value.solved / value.total) * 100) : 0}%`,
                        }"
                      />
                    </div>
                  </div>
                </div>
              </AppCard>

              <AppCard
                variant="panel"
                accent="warning"
                eyebrow="难度进度"
                subtitle="按题目难度查看学员当前突破情况。"
              >
                <div class="mt-4 space-y-3">
                  <div
                    v-for="(value, key) in progress?.by_difficulty || {}"
                    :key="key"
                    class="flex items-center justify-between rounded-lg bg-[var(--color-bg-surface)] px-3 py-3 text-sm"
                  >
                    <span class="font-medium text-[var(--color-text-primary)]">{{
                      difficultyLabel(key)
                    }}</span>
                    <span class="text-[var(--color-text-secondary)]"
                      >{{ value.solved }} / {{ value.total }}</span
                    >
                  </div>
                </div>
              </AppCard>
            </div>
          </SectionCard>

          <SectionCard title="能力画像" subtitle="以雷达图观察当前能力维度分布。">
            <div class="mt-4">
              <SkillRadar :scores="radarScores" />
            </div>
          </SectionCard>
        </div>

        <SectionCard title="推荐训练任务" subtitle="根据当前能力薄弱维度筛出的优先训练题目。">
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
                打开挑战
                <ArrowRight class="h-4 w-4" />
              </div>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard title="社区题解状态" subtitle="查看当前学员最近的社区题解发布、隐藏与推荐状态。">
          <AppEmpty
            v-if="writeupSubmissions.length === 0"
            title="暂无社区题解"
            description="当前学员还没有保存或发布题解。"
            icon="FileText"
          />

          <template v-else>
            <div class="grid gap-3 md:grid-cols-3">
              <article class="insight-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">最近题解</div>
                <div class="insight-kpi-value">{{ writeupSubmissions.length }}</div>
                <div class="insight-kpi-hint">当前分析页展示的题解数量</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--success">
                <div class="insight-kpi-label">公开中</div>
                <div class="insight-kpi-value">{{ visibleWriteupCount }}</div>
                <div class="insight-kpi-hint">当前仍对已解题学生可见的题解</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--warning">
                <div class="insight-kpi-label">推荐中</div>
                <div class="insight-kpi-value">{{ recommendedWriteupCount }}</div>
                <div class="insight-kpi-hint">会进入学生端推荐题解区的内容</div>
              </article>
            </div>

            <div class="mt-5 grid gap-3">
              <AppCard
                v-for="item in writeupSubmissions"
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
                      {{ item.title }}
                    </div>
                  </div>
                  <div class="flex flex-wrap gap-2">
                    <span class="writeup-chip writeup-chip--muted">
                      {{ submissionStatusLabel(item.submission_status) }}
                    </span>
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

                <div class="mt-4 rounded-2xl bg-[var(--color-bg-base)] px-4 py-3 text-sm leading-7 text-[var(--color-text-secondary)]">
                  {{ item.content_preview || '暂无摘要' }}
                </div>

                <div class="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs text-[var(--color-text-secondary)]">
                  <span>最近更新：{{ new Date(item.updated_at).toLocaleString('zh-CN') }}</span>
                  <div class="flex flex-wrap items-center gap-3">
                    <button
                      type="button"
                      class="inline-flex items-center gap-1 font-medium text-[var(--color-primary)]"
                      @click="emitWriteupModeration(item.id, item.is_recommended ? 'unrecommend' : 'recommend')"
                    >
                      {{ item.is_recommended ? '取消推荐' : '设为推荐' }}
                    </button>
                    <button
                      type="button"
                      class="inline-flex items-center gap-1 font-medium"
                      :class="item.visibility_status === 'hidden' ? 'text-[var(--color-success)]' : 'text-[var(--color-warning)]'"
                      @click="emitWriteupModeration(item.id, item.visibility_status === 'hidden' ? 'restore' : 'hide')"
                    >
                      {{ item.visibility_status === 'hidden' ? '恢复可见' : '隐藏题解' }}
                    </button>
                    <button
                      type="button"
                      class="inline-flex items-center gap-1 font-medium text-[var(--color-primary)]"
                      @click="openChallenge(item.challenge_id)"
                    >
                      打开挑战
                      <ArrowRight class="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </AppCard>
            </div>
          </template>
        </SectionCard>

        <SectionCard title="人工审核题" subtitle="查看该学员待教师评阅的非标准答案题目。">
          <AppEmpty
            v-if="manualReviewSubmissions.length === 0"
            title="暂无人工审核提交"
            description="当前学员还没有需要教师处理的人工审核题。"
            icon="ClipboardCheck"
          />

          <template v-else>
            <div class="grid gap-3 md:grid-cols-3">
              <article class="insight-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">待处理</div>
                <div class="insight-kpi-value">{{ manualReviewSubmissions.length }}</div>
                <div class="insight-kpi-hint">当前分析页展示的人工审核提交数</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--warning">
                <div class="insight-kpi-label">待审核</div>
                <div class="insight-kpi-value">{{ manualReviewSubmissions.filter((item) => item.review_status === 'pending').length }}</div>
                <div class="insight-kpi-hint">尚未给出审核结果的提交</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--success">
                <div class="insight-kpi-label">已通过</div>
                <div class="insight-kpi-value">{{ approvedManualReviewCount }}</div>
                <div class="insight-kpi-hint">已经转为得分的人工审核提交</div>
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
                  <div class="h-5 w-32 animate-pulse rounded bg-[var(--color-bg-base)]" />
                  <div class="h-24 animate-pulse rounded-2xl bg-[var(--color-bg-base)]" />
                  <div class="h-24 animate-pulse rounded-2xl bg-[var(--color-bg-base)]" />
                </div>

                <AppEmpty
                  v-else-if="!activeManualReview"
                  title="选择一条人工审核提交"
                  description="点击左侧卡片查看完整答案并进行审核。"
                  icon="ClipboardList"
                />

                <template v-else>
                  <div class="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <div class="journal-eyebrow">Manual Review</div>
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

                  <div class="mt-5 rounded-2xl bg-[var(--color-bg-base)] px-4 py-4">
                    <div class="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--color-text-secondary)]">
                      提交答案
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
                      class="challenge-input mt-3 w-full rounded-2xl border px-4 py-3 text-sm leading-7 transition-colors focus:outline-none"
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
                        class="challenge-btn-outline"
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

        <SectionCard title="攻防证据链" subtitle="教师按关键动作查看该学员的利用过程。">
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
                        class="rounded-full border border-[var(--color-border-default)] px-2.5 py-1"
                      >
                        {{ String(event.meta?.event_stage || 'trace') }}
                      </span>
                      <span
                        v-if="typeof event.meta?.method === 'string'"
                        class="rounded-full border border-[var(--color-border-default)] px-2.5 py-1"
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

        <StudentTimelinePage :timeline="timeline" />
      </template>
    </template>
  </div>
</template>

<style scoped>
.student-insight-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #2563eb;
  --journal-accent-strong: #1d4ed8;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --color-primary-soft: color-mix(in srgb, var(--journal-accent) 8%, transparent);
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

:deep(.section-card) {
  padding: 1.1rem 1.1rem 1.05rem;
  border: 1px solid var(--teacher-card-border);
  border-radius: 16px;
  background: var(--journal-surface-subtle);
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

:deep(.section-card__header) {
  margin-bottom: 1rem;
  border-bottom: 1px dashed var(--teacher-divider);
  padding-bottom: 0.75rem;
}

:deep(.section-card__body) {
  padding-left: 0;
}

.insight-kpi-grid {
  align-items: stretch;
}

.insight-kpi-card {
  border: 1px solid var(--teacher-card-border);
  border-radius: 16px;
  background: var(--journal-surface);
  padding: 0.9rem 0.95rem;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.insight-kpi-card--primary {
  border-top: 3px solid color-mix(in srgb, var(--journal-accent) 36%, transparent);
}

.insight-kpi-card--success {
  border-top: 3px solid color-mix(in srgb, var(--color-success) 34%, transparent);
}

.insight-kpi-card--warning {
  border-top: 3px solid color-mix(in srgb, var(--color-warning) 34%, transparent);
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
