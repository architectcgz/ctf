<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-col space-y-6 rounded-[30px] border p-6 md:p-8"
  >
    <div
      v-if="loading"
      class="flex items-center justify-center py-12"
    >
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
      />
    </div>

    <div
      v-else-if="challenge"
      class="space-y-6"
    >
      <div class="challenge-panel p-6 md:p-8">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-3">
            <div class="journal-eyebrow">
              Challenge Detail
            </div>
            <h1 class="text-3xl font-bold text-[var(--journal-ink)]">
              {{ challenge.title }}
            </h1>
            <div class="flex flex-wrap gap-2">
              <span
                class="rounded-full px-3 py-1 text-sm font-medium"
                :style="{
                  backgroundColor: getCategoryColor(challenge.category) + '22',
                  color: getCategoryColor(challenge.category),
                }"
              >
                {{ getCategoryLabel(challenge.category) }}
              </span>
              <span
                class="rounded-full px-3 py-1 text-sm font-medium"
                :style="{
                  backgroundColor: getDifficultyColor(challenge.difficulty) + '22',
                  color: getDifficultyColor(challenge.difficulty),
                }"
              >
                {{ getDifficultyLabel(challenge.difficulty) }}
              </span>
              <span
                v-for="tag in challenge.tags"
                :key="tag"
                class="rounded-full border border-[var(--journal-border)] bg-white/60 px-3 py-1 text-sm text-[var(--journal-ink)]"
              >
                {{ tag }}
              </span>
            </div>
          </div>

          <div class="challenge-score-card px-4 py-3 text-left lg:min-w-[148px] lg:text-right">
            <div class="text-[11px] uppercase tracking-[0.22em] text-[var(--journal-muted)]">
              Score
            </div>
            <div class="mt-1 font-mono text-2xl font-bold text-[var(--journal-ink)]">
              {{ challenge.points
              }}<span class="ml-1 text-lg text-[var(--journal-muted)]">pts</span>
            </div>
            <span
              v-if="challenge.is_solved"
              class="mt-3 inline-flex rounded-full bg-[var(--color-success)]/18 px-3 py-1 text-sm font-medium text-[var(--color-success)]"
            >
              已完成 ✓
            </span>
          </div>
        </div>
      </div>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
        <main class="space-y-6">
          <div class="challenge-panel p-6">
            <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
              <h2 class="text-lg font-semibold text-[var(--journal-ink)]">
                挑战描述
              </h2>
            </div>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <div
              class="prose challenge-prose max-w-none"
              v-html="sanitizedDescription"
            />
            <button
              v-if="challenge.attachment_url"
              type="button"
              class="challenge-btn-outline mt-4"
              @click="downloadAttachment"
            >
              下载附件
            </button>
          </div>

          <div class="challenge-panel p-6">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-[var(--journal-ink)]">
                  Flag 提交
                </h2>
                <div class="mt-1 text-sm text-[var(--journal-muted)]">
                  保持在当前题目页即可提交答案。
                </div>
              </div>
              <span
                v-if="challenge.is_solved"
                class="rounded-full bg-[var(--color-success)]/18 px-3 py-1 text-xs font-medium text-[var(--color-success)]"
              >
                已通过
              </span>
            </div>
            <div class="mt-4 space-y-4">
              <div class="flex flex-col gap-3 sm:flex-row">
                <input
                  v-model="flagInput"
                  type="text"
                  :placeholder="submitPlaceholder"
                  :disabled="challenge.is_solved"
                  class="challenge-input flex-1 rounded-xl border px-4 py-3 font-mono transition-colors focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
                  :class="submitInputClass"
                  @keyup.enter="submitFlagHandler"
                >
                <button
                  type="button"
                  :disabled="challenge.is_solved || submitting"
                  class="challenge-btn-primary rounded-xl px-6 py-3 text-sm font-medium text-white transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                  @click="submitFlagHandler"
                >
                  {{ submitting ? '提交中...' : '提交' }}
                </button>
              </div>
              <div
                v-if="submitResult"
                :class="submitResult.className"
                class="text-sm"
              >
                {{ submitResult.message }}
              </div>
            </div>
          </div>

          <div
            v-if="challenge.hints.length > 0"
            class="challenge-panel p-6"
          >
            <h2 class="mb-4 text-lg font-semibold text-[var(--journal-ink)]">
              提示系统
            </h2>
            <div class="hint-list">
              <div
                v-for="hint in challenge.hints"
                :key="hint.id"
                class="hint-item"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <div class="text-sm font-medium text-[var(--journal-ink)]">
                      Level {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}
                    </div>
                    <div
                      v-if="hint.cost_points"
                      class="mt-1 text-xs text-[var(--journal-muted)]"
                    >
                      解锁消耗：{{ hint.cost_points }} 分
                    </div>
                  </div>
                  <button
                    v-if="!hint.is_unlocked"
                    type="button"
                    :disabled="unlockingLevel === hint.level"
                    class="challenge-btn-primary rounded-lg px-4 py-2 text-xs font-medium text-white transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                    @click="unlockHintHandler(hint.level)"
                  >
                    {{ unlockingLevel === hint.level ? '解锁中...' : '解锁提示' }}
                  </button>
                  <span
                    v-else
                    class="rounded bg-[var(--color-success)]/20 px-3 py-1 text-xs font-medium text-[var(--color-success)]"
                  >
                    已解锁
                  </span>
                </div>
                <div
                  v-if="hint.is_unlocked"
                  class="mt-3 text-sm leading-6 text-[var(--journal-muted)]"
                >
                  {{ hint.content }}
                </div>
                <div
                  v-else
                  class="mt-3 text-sm text-[var(--journal-muted)]"
                >
                  解锁后显示提示内容
                </div>
              </div>
            </div>
          </div>

          <div class="challenge-panel writeup-workbench p-6">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div class="space-y-3">
                <div class="journal-eyebrow">
                  Solutions Hub
                </div>
                <div>
                  <h2 class="text-lg font-semibold text-[var(--journal-ink)]">
                    题解
                  </h2>
                  <p class="mt-2 text-sm leading-7 text-[var(--journal-muted)]">
                    像 LeetCode 一样在同一块区域里查看推荐题解、社区题解，并维护你自己的复盘。
                  </p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <span class="writeup-status-pill writeup-status-pill--primary">
                    推荐 {{ recommendedSolutions.length }} 篇
                  </span>
                  <span class="writeup-status-pill writeup-status-pill--muted">
                    社区 {{ communitySolutions.length }} 篇
                  </span>
                </div>
              </div>

              <button
                type="button"
                class="challenge-btn-outline self-start"
                @click="toggleSolutionsPanel"
              >
                {{ solutionsExpanded ? '收起题解区' : '展开题解区' }}
              </button>
            </div>

            <div
              v-if="solutionsExpanded"
              class="mt-6 space-y-8"
            >
              <section class="space-y-5">
                <div
                  v-if="!challenge.is_solved"
                  class="rounded-2xl border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 px-5 py-4 text-sm leading-7 text-[var(--color-warning)]"
                >
                  <div>解出题目后可查看推荐题解与社区题解。</div>
                  <div class="mt-1 text-[var(--journal-muted)]">
                    你现在仍然可以先写自己的草稿，解题后再发布到社区。
                  </div>
                </div>

                <template v-else>
                  <div class="solution-tabbar">
                    <button
                      type="button"
                      class="solution-tab"
                      :class="{ 'solution-tab--active': activeTab === 'recommended' }"
                      @click="activeTab = 'recommended'"
                    >
                      推荐题解
                    </button>
                    <button
                      type="button"
                      class="solution-tab"
                      :class="{ 'solution-tab--active': activeTab === 'community' }"
                      @click="activeTab = 'community'"
                    >
                      社区题解
                    </button>
                  </div>

                  <div class="solution-board">
                    <div class="solution-list">
                      <div
                        v-if="displayedSolutionCards.length === 0"
                        class="rounded-2xl border border-dashed border-[var(--journal-border)] px-4 py-6 text-sm text-[var(--journal-muted)]"
                      >
                        {{ activeTab === 'recommended' ? '还没有推荐题解。' : '还没有公开的社区题解。' }}
                      </div>

                      <button
                        v-for="item in displayedSolutionCards"
                        :key="item.id"
                        type="button"
                        class="solution-list-item"
                        :class="{ 'solution-list-item--active': item.id === activeSolution?.id }"
                        @click="selectedSolutionId = item.id"
                      >
                        <div class="flex items-start justify-between gap-3">
                          <div class="min-w-0">
                            <div class="truncate text-sm font-semibold text-[var(--journal-ink)]">
                              {{ item.title }}
                            </div>
                            <div class="mt-1 text-xs text-[var(--journal-muted)]">
                              {{ item.authorName }} · {{ item.sourceLabel }}
                            </div>
                          </div>
                          <span
                            v-if="item.badge"
                            class="writeup-status-pill"
                            :class="item.badgeClass"
                          >
                            {{ item.badge }}
                          </span>
                        </div>
                        <p
                          v-if="item.preview"
                          class="mt-3 line-clamp-3 text-sm leading-6 text-[var(--journal-muted)]"
                        >
                          {{ item.preview }}
                        </p>
                      </button>
                    </div>

                    <div class="solution-preview">
                      <template v-if="activeSolution">
                        <div class="flex flex-wrap items-start justify-between gap-3">
                          <div>
                            <h3 class="text-lg font-semibold text-[var(--journal-ink)]">
                              {{ activeSolution.title }}
                            </h3>
                            <div class="mt-2 text-sm text-[var(--journal-muted)]">
                              {{ activeSolution.authorName }} · {{ activeSolution.sourceLabel }}
                            </div>
                          </div>
                          <div class="flex flex-wrap gap-2">
                            <span
                              v-if="activeSolution.badge"
                              class="writeup-status-pill"
                              :class="activeSolution.badgeClass"
                            >
                              {{ activeSolution.badge }}
                            </span>
                            <span class="writeup-status-pill writeup-status-pill--muted">
                              {{ formatWriteupTime(activeSolution.updatedAt) }}
                            </span>
                          </div>
                        </div>
                        <!-- eslint-disable-next-line vue/no-v-html -->
                        <div
                          class="prose challenge-prose solution-preview__content mt-6 max-w-none"
                          v-html="sanitizedActiveSolutionContent"
                        />
                      </template>

                      <div
                        v-else
                        class="rounded-2xl border border-dashed border-[var(--journal-border)] px-4 py-10 text-sm text-[var(--journal-muted)]"
                      >
                        当前分组还没有可展示的题解。
                      </div>
                    </div>
                  </div>
                </template>
              </section>

              <section class="space-y-5">
                <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                  <div class="max-w-2xl space-y-3">
                    <div class="journal-eyebrow">
                      My Solution
                    </div>
                    <div>
                      <h3 class="text-base font-semibold text-[var(--journal-ink)]">
                        我的题解
                      </h3>
                      <p class="mt-2 text-sm leading-7 text-[var(--journal-muted)]">
                        一人一篇，持续更新。草稿仅自己可见，发布后默认进入社区题解区，教师或管理员可隐藏或设为推荐题解。
                      </p>
                    </div>
                    <div class="flex flex-wrap gap-2">
                      <span class="writeup-status-pill writeup-status-pill--primary">
                        {{ submissionStatusLabel(myWriteup?.submission_status) }}
                      </span>
                      <span
                        v-if="myWriteup?.visibility_status === 'hidden'"
                        class="writeup-status-pill writeup-status-pill--warning"
                      >
                        已隐藏
                      </span>
                      <span
                        v-else-if="myWriteup?.submission_status === 'published' || myWriteup?.submission_status === 'submitted'"
                        class="writeup-status-pill writeup-status-pill--success"
                      >
                        社区可见
                      </span>
                      <span
                        v-if="myWriteup?.is_recommended"
                        class="writeup-status-pill writeup-status-pill--primary"
                      >
                        推荐题解
                      </span>
                    </div>
                  </div>

                  <div class="writeup-side-note">
                    <div
                      class="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--journal-accent)]"
                    >
                      Publish Rule
                    </div>
                    <div class="mt-3 text-sm leading-6 text-[var(--journal-ink)]">
                      {{
                        challenge.is_solved
                          ? '你已解题，可以直接发布到社区。'
                          : '未解题前只能保存草稿，推荐题解与社区题解也仍保持锁定。'
                      }}
                    </div>
                  </div>
                </div>

                <div
                  v-if="myWriteup?.visibility_status === 'hidden'"
                  class="rounded-lg border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 px-4 py-3 text-sm text-[var(--color-warning)]"
                >
                  当前题解已被教师或管理员隐藏，仅你自己可见。
                </div>

                <div class="mt-6 grid gap-4 xl:grid-cols-[0.95fr_1.05fr]">
                  <label class="writeup-field">
                    <span class="writeup-field-label">标题</span>
                    <input
                      v-model="writeupTitle"
                      type="text"
                      maxlength="256"
                      placeholder="例如：从回显异常到拿到 flag 的完整链路"
                      class="challenge-input w-full rounded-2xl border px-4 py-3 text-sm transition-colors focus:outline-none"
                    >
                  </label>

                  <div class="writeup-meta-grid">
                    <div class="writeup-meta-card">
                      <div class="writeup-meta-label">
                        当前状态
                      </div>
                      <div class="writeup-meta-value">
                        {{ submissionStatusLabel(myWriteup?.submission_status) }}
                      </div>
                    </div>
                    <div class="writeup-meta-card">
                      <div class="writeup-meta-label">
                        社区可见性
                      </div>
                      <div class="writeup-meta-value">
                        {{ visibilityStatusLabel(myWriteup?.visibility_status) }}
                      </div>
                    </div>
                  </div>
                </div>

                <label class="writeup-field mt-4 block">
                  <span class="writeup-field-label">正文</span>
                  <textarea
                    v-model="writeupContent"
                    rows="10"
                    placeholder="建议按『题目理解 → 利用过程 → 核心 payload / 证据 → 踩坑点』组织。"
                    class="challenge-input writeup-textarea w-full rounded-[24px] border px-4 py-4 text-sm leading-7 transition-colors focus:outline-none"
                  />
                </label>

                <div class="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div class="text-sm text-[var(--journal-muted)]">
                    {{
                      submissionLoading
                        ? '正在同步你的题解...'
                        : myWriteup?.updated_at
                          ? `最近更新：${formatWriteupTime(myWriteup.updated_at)}`
                          : '还没有提交记录，可以先保存草稿。'
                    }}
                  </div>
                  <div class="flex flex-wrap gap-3">
                    <button
                      type="button"
                      :disabled="submissionLoading || submissionSaving !== null"
                      class="challenge-btn-outline"
                      @click="saveWriteup('draft')"
                    >
                      {{ submissionSaving === 'draft' ? '保存中...' : '保存草稿' }}
                    </button>
                    <button
                      type="button"
                      :disabled="submissionLoading || submissionSaving !== null || !challenge.is_solved"
                      class="challenge-btn-primary rounded-xl px-5 py-3 text-sm font-medium text-white transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                      @click="saveWriteup('published')"
                    >
                      {{ submissionSaving === 'published' ? '发布中...' : '发布题解' }}
                    </button>
                  </div>
                </div>
              </section>
            </div>
          </div>
        </main>

        <aside class="space-y-6 lg:sticky lg:top-6 lg:self-start">
          <ChallengeInstanceCard
            v-if="needTarget"
            :instance="instance"
            :loading="instanceLoading"
            :creating="instanceCreating"
            :opening="instanceOpening"
            :extending="instanceExtending"
            :destroying="instanceDestroying"
            :challenge-solved="challenge.is_solved"
            @start="startInstance"
            @open="openInstance"
            @extend="extendChallengeInstance"
            @destroy="destroyChallengeInstance"
          />
          <section
            v-else
            class="rounded-2xl border border-[var(--color-success)]/30 bg-[var(--color-success)]/10 p-5 text-sm text-[var(--color-success)]"
          >
            该题目不需要靶机，可直接分析题面并提交 Flag。
          </section>
        </aside>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { marked } from 'marked'
import { useRoute, useRouter } from 'vue-router'

import {
  downloadAttachment as downloadChallengeAttachment,
  getChallengeDetail,
  getCommunityChallengeSolutions,
  getMyChallengeWriteupSubmission,
  getRecommendedChallengeSolutions,
  submitFlag,
  unlockHint,
  upsertChallengeWriteupSubmission,
} from '@/api/challenge'
import type {
  ChallengeCategory,
  ChallengeDetailData,
  ChallengeDifficulty,
  CommunityChallengeSolutionData,
  RecommendedChallengeSolutionData,
  SubmissionWriteupData,
  SubmissionWriteupStatus,
  SubmissionWriteupVisibilityStatus,
} from '@/api/contracts'
import ChallengeInstanceCard from '@/components/challenge/ChallengeInstanceCard.vue'
import { useChallengeInstance } from '@/composables/useChallengeInstance'
import { useSanitize } from '@/composables/useSanitize'
import { useToast } from '@/composables/useToast'

type SolutionTab = 'recommended' | 'community'
type EditableWriteupStatus = 'draft' | 'published'

interface SolutionCard {
  id: string
  title: string
  content: string
  preview: string
  authorName: string
  sourceLabel: string
  badge: string
  badgeClass: string
  updatedAt?: string
}

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { sanitizeHtml } = useSanitize()

const challengeId = computed(() => String(route.params.id ?? ''))
const challenge = ref<ChallengeDetailData | null>(null)
const loading = ref(false)
const submitting = ref(false)
const recommendedSolutions = ref<RecommendedChallengeSolutionData[]>([])
const communitySolutions = ref<CommunityChallengeSolutionData[]>([])
const myWriteup = ref<SubmissionWriteupData | null>(null)
const submissionLoading = ref(false)
const submissionSaving = ref<EditableWriteupStatus | null>(null)
const writeupTitle = ref('')
const writeupContent = ref('')
const flagInput = ref('')
const unlockingLevel = ref<number | null>(null)
const submitResult = ref<{
  variant: 'success' | 'error' | 'pending'
  className: string
  message: string
} | null>(null)
const solutionsExpanded = ref(true)
const activeTab = ref<SolutionTab>('recommended')
const selectedSolutionId = ref<string | null>(null)

const {
  instance,
  loading: instanceLoading,
  creating: instanceCreating,
  opening: instanceOpening,
  extending: instanceExtending,
  destroying: instanceDestroying,
  start: startInstance,
  open: openInstance,
  extend: extendChallengeInstance,
  destroy: destroyChallengeInstance,
} = useChallengeInstance(challengeId)

function renderRichContent(source?: string): string {
  if (!source) return ''
  const html = marked.parse(source, {
    gfm: true,
    breaks: true,
  })
  return sanitizeHtml(typeof html === 'string' ? html : source)
}

function buildPreview(source?: string): string {
  if (!source) return ''
  return source
    .replace(/<[^>]+>/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .slice(0, 120)
}

const sanitizedDescription = computed(() => renderRichContent(challenge.value?.description))
const needTarget = computed(() => challenge.value?.need_target ?? true)

const recommendedSolutionCards = computed<SolutionCard[]>(() =>
  recommendedSolutions.value.map((item) => ({
    id: item.id,
    title: item.title,
    content: item.content,
    preview: buildPreview(item.content),
    authorName: item.author_name,
    sourceLabel: item.source_type === 'official' ? '官方题解' : '社区推荐',
    badge: '推荐题解',
    badgeClass: 'writeup-status-pill--primary',
    updatedAt: item.updated_at,
  }))
)

const communitySolutionCards = computed<SolutionCard[]>(() =>
  communitySolutions.value.map((item) => ({
    id: item.id,
    title: item.title,
    content: item.content,
    preview: item.content_preview || buildPreview(item.content),
    authorName: item.author_name,
    sourceLabel: '社区题解',
    badge: item.is_recommended ? '推荐' : '',
    badgeClass: item.is_recommended ? 'writeup-status-pill--primary' : 'writeup-status-pill--muted',
    updatedAt: item.updated_at,
  }))
)

const displayedSolutionCards = computed(() =>
  activeTab.value === 'recommended' ? recommendedSolutionCards.value : communitySolutionCards.value
)

const activeSolution = computed(() => {
  if (displayedSolutionCards.value.length === 0) return null
  return (
    displayedSolutionCards.value.find((item) => item.id === selectedSolutionId.value) ??
    displayedSolutionCards.value[0]
  )
})

const sanitizedActiveSolutionContent = computed(() => renderRichContent(activeSolution.value?.content))

const submitPlaceholder = computed(() => {
  if (challenge.value?.is_solved) return '该题已通过'

  switch (submitResult.value?.variant) {
    case 'success':
      return '答案已通过'
    case 'pending':
      return '已提交，等待教师审核'
    case 'error':
      return '答案不正确，请继续尝试'
    default:
      return 'flag{...}'
  }
})

const submitInputClass = computed(() => {
  switch (submitResult.value?.variant) {
    case 'success':
      return 'border-[var(--color-success)] bg-[var(--color-success)]/5'
    case 'pending':
      return 'border-[var(--color-warning)] bg-[var(--color-warning)]/8'
    case 'error':
      return 'border-[var(--color-danger)] bg-[var(--color-danger)]/5'
    default:
      return 'border-[#0891b2]'
  }
})

function clearSolutions(): void {
  recommendedSolutions.value = []
  communitySolutions.value = []
  activeTab.value = 'recommended'
  selectedSolutionId.value = null
}

function hydrateSubmissionForm(item: SubmissionWriteupData | null): void {
  writeupTitle.value = item?.title ?? ''
  writeupContent.value = item?.content ?? ''
}

async function loadSolutions(id: string): Promise<void> {
  try {
    const [recommended, communityPage] = await Promise.all([
      getRecommendedChallengeSolutions(id),
      getCommunityChallengeSolutions(id),
    ])
    recommendedSolutions.value = recommended
    communitySolutions.value = communityPage.list
  } catch {
    clearSolutions()
    toast.error('加载题解失败')
  }
}

async function loadChallenge(): Promise<void> {
  const id = challengeId.value
  loading.value = true

  try {
    const detail = await getChallengeDetail(id)
    challenge.value = detail

    if (detail.is_solved) {
      await loadSolutions(id)
    } else {
      clearSolutions()
    }
  } catch {
    toast.error('加载挑战详情失败')
    void router.push('/challenges')
  } finally {
    loading.value = false
  }
}

async function loadMyWriteupSubmission(): Promise<void> {
  if (!challengeId.value) return

  submissionLoading.value = true
  try {
    myWriteup.value = await getMyChallengeWriteupSubmission(challengeId.value)
    hydrateSubmissionForm(myWriteup.value)
  } catch {
    toast.error('加载个人题解失败')
  } finally {
    submissionLoading.value = false
  }
}

async function submitFlagHandler(): Promise<void> {
  if (!challenge.value || !flagInput.value.trim()) return

  submitting.value = true
  submitResult.value = null
  try {
    const result = await submitFlag(challenge.value.id, flagInput.value.trim())
    switch (result.status) {
      case 'correct':
        submitResult.value = {
          variant: 'success',
          className: 'text-[var(--color-success)]',
          message: result.message,
        }
        toast.success('Flag 正确！')
        challenge.value.is_solved = true
        await loadSolutions(challenge.value.id)
        break
      case 'pending_review':
        submitResult.value = {
          variant: 'pending',
          className: 'text-[var(--color-warning)]',
          message: result.message,
        }
        toast.info('答案已提交，等待教师审核')
        break
      default:
        submitResult.value = {
          variant: 'error',
          className: 'text-[var(--color-danger)]',
          message: result.message,
        }
        break
    }
  } catch {
    submitResult.value = {
      variant: 'error',
      className: 'text-[var(--color-danger)]',
      message: '提交失败，请重试',
    }
  } finally {
    submitting.value = false
  }
}

async function unlockHintHandler(level: number): Promise<void> {
  if (!challenge.value) return

  unlockingLevel.value = level
  try {
    const result = await unlockHint(challenge.value.id, level)
    challenge.value.hints = challenge.value.hints.map((hint) =>
      hint.level === level ? result.hint : hint
    )
    toast.success('提示已解锁')
  } catch {
    toast.error('解锁提示失败')
  } finally {
    unlockingLevel.value = null
  }
}

async function downloadAttachment(): Promise<void> {
  if (!challenge.value?.attachment_url) return

  const attachmentURL = challenge.value.attachment_url
  try {
    const parsed = new URL(attachmentURL, window.location.origin)
    if (parsed.origin !== window.location.origin) {
      window.open(attachmentURL, '_blank', 'noopener')
      return
    }
  } catch {
    // keep axios fallback for relative urls
  }

  try {
    const { blob, filename } = await downloadChallengeAttachment(attachmentURL)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(url)
  } catch {
    toast.error('下载附件失败')
  }
}

async function saveWriteup(status: EditableWriteupStatus): Promise<void> {
  if (!challenge.value) return
  if (!writeupTitle.value.trim() || !writeupContent.value.trim()) {
    toast.error('请先补全题解标题和正文')
    return
  }
  if (status === 'published' && !challenge.value.is_solved) {
    toast.error('解题后才能发布到社区')
    return
  }

  submissionSaving.value = status
  try {
    const saved = await upsertChallengeWriteupSubmission(challenge.value.id, {
      title: writeupTitle.value.trim(),
      content: writeupContent.value.trim(),
      submission_status: status,
    })
    myWriteup.value = saved
    hydrateSubmissionForm(saved)
    toast.success(status === 'published' ? '题解已发布到社区' : '草稿已保存')
  } catch {
    toast.error(status === 'published' ? '发布题解失败' : '保存草稿失败')
  } finally {
    submissionSaving.value = null
  }
}

function toggleSolutionsPanel(): void {
  solutionsExpanded.value = !solutionsExpanded.value
}

function submissionStatusLabel(status?: SubmissionWriteupStatus): string {
  if (status === 'draft') return '草稿'
  if (status === 'published' || status === 'submitted') return '已发布'
  return '未开始'
}

function visibilityStatusLabel(status?: SubmissionWriteupVisibilityStatus): string {
  if (status === 'hidden') return '已隐藏'
  if (myWriteup.value?.submission_status === 'published' || myWriteup.value?.submission_status === 'submitted') {
    return '已公开'
  }
  return '未发布'
}

function formatWriteupTime(value?: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getCategoryLabel(category: ChallengeCategory): string {
  const labels: Record<ChallengeCategory, string> = {
    web: 'Web',
    pwn: 'Pwn',
    reverse: '逆向',
    crypto: '密码',
    misc: '杂项',
    forensics: '取证',
  }
  return labels[category]
}

function getCategoryColor(category: ChallengeCategory): string {
  const colors: Record<ChallengeCategory, string> = {
    web: '#3b82f6',
    pwn: '#ef4444',
    reverse: '#8b5cf6',
    crypto: '#f59e0b',
    misc: '#10b981',
    forensics: '#06b6d4',
  }
  return colors[category]
}

function getDifficultyLabel(difficulty: ChallengeDifficulty): string {
  const labels: Record<ChallengeDifficulty, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    insane: '地狱',
  }
  return labels[difficulty]
}

function getDifficultyColor(difficulty: ChallengeDifficulty): string {
  const colors: Record<ChallengeDifficulty, string> = {
    beginner: '#10b981',
    easy: '#3b82f6',
    medium: '#f59e0b',
    hard: '#ef4444',
    insane: '#7c3aed',
  }
  return colors[difficulty]
}

watch(
  displayedSolutionCards,
  (items) => {
    if (!items.some((item) => item.id === selectedSolutionId.value)) {
      selectedSolutionId.value = items[0]?.id ?? null
    }
  },
  { immediate: true }
)

watch(
  [recommendedSolutionCards, communitySolutionCards],
  ([recommended, community]) => {
    if (activeTab.value === 'recommended' && recommended.length === 0 && community.length > 0) {
      activeTab.value = 'community'
    } else if (activeTab.value === 'community' && community.length === 0 && recommended.length > 0) {
      activeTab.value = 'recommended'
    }
  },
  { immediate: true }
)

watch(
  challengeId,
  () => {
    challenge.value = null
    myWriteup.value = null
    clearSolutions()
    writeupTitle.value = ''
    writeupContent.value = ''
    flagInput.value = ''
    submitResult.value = null
    solutionsExpanded.value = true
    void Promise.all([loadChallenge(), loadMyWriteupSubmission()])
  },
  { immediate: true }
)
</script>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

.journal-hero {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base))
    );
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.05);
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.22);
  background: rgba(99, 102, 241, 0.07);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.challenge-score-card {
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.challenge-panel {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.challenge-btn-outline {
  border-radius: 10px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.5rem 1rem;
  font-size: 0.85rem;
  color: var(--journal-ink);
  transition: all 0.15s;
}

.challenge-btn-outline:hover {
  border-color: var(--journal-accent);
  color: var(--journal-accent);
}

.challenge-btn-primary {
  background: var(--journal-accent);
}

.challenge-btn-primary:hover:not(:disabled) {
  background: #4338ca;
}

.challenge-input {
  background: var(--journal-surface);
  color: var(--journal-ink);
}

.challenge-input::placeholder {
  color: var(--journal-muted);
}

.challenge-prose {
  color: var(--journal-muted);
}

.writeup-workbench {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.1), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 94%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 96%, var(--color-bg-base))
    );
}

.writeup-side-note {
  max-width: 18rem;
  border: 1px solid rgba(99, 102, 241, 0.14);
  border-radius: 20px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  padding: 1rem 1.1rem;
}

.writeup-field {
  display: block;
}

.writeup-field-label {
  display: block;
  margin-bottom: 0.65rem;
  font-size: 0.76rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-meta-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.writeup-meta-card {
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  padding: 0.95rem 1rem;
}

.writeup-meta-label {
  font-size: 0.74rem;
  text-transform: uppercase;
  letter-spacing: 0.18em;
  color: var(--journal-muted);
}

.writeup-meta-value {
  margin-top: 0.5rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.writeup-textarea {
  min-height: 16rem;
}

.writeup-status-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.4rem 0.8rem;
  font-size: 0.74rem;
  font-weight: 600;
  letter-spacing: 0.04em;
}

.writeup-status-pill--primary {
  background: rgba(79, 70, 229, 0.12);
  color: #4338ca;
}

.writeup-status-pill--success {
  background: rgba(16, 185, 129, 0.14);
  color: #047857;
}

.writeup-status-pill--warning {
  background: rgba(245, 158, 11, 0.16);
  color: #b45309;
}

.writeup-status-pill--muted {
  background: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 34%, transparent);
  color: #475569;
}

.solution-tabbar {
  display: inline-flex;
  gap: 0.75rem;
  padding: 0.35rem;
  border: 1px solid var(--journal-border);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.solution-tab {
  border: 0;
  border-radius: 999px;
  padding: 0.7rem 1.1rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--journal-muted);
  background: transparent;
  transition: all 0.15s;
}

.solution-tab--active {
  color: #fff;
  background: var(--journal-accent);
}

.solution-board {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 320px) minmax(0, 1fr);
}

.solution-list {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.solution-list-item {
  width: 100%;
  text-align: left;
  border: 1px solid var(--journal-border);
  border-radius: 20px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  padding: 1rem 1.05rem;
  transition:
    border-color 0.15s,
    transform 0.15s,
    box-shadow 0.15s;
}

.solution-list-item:hover,
.solution-list-item--active {
  border-color: color-mix(in srgb, var(--journal-accent) 38%, var(--journal-border));
  box-shadow: 0 14px 26px rgba(79, 70, 229, 0.08);
  transform: translateY(-1px);
}

.solution-preview {
  min-height: 24rem;
  border: 1px solid var(--journal-border);
  border-radius: 24px;
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 94%, var(--color-bg-base));
  padding: 1.4rem 1.5rem;
}

.solution-preview__content {
  min-height: 16rem;
}

.hint-list {
  border-radius: 20px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.hint-item {
  padding: 1rem 1.1rem;
}

.hint-item + .hint-item {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.challenge-prose :deep(h1),
.challenge-prose :deep(h2),
.challenge-prose :deep(h3),
.challenge-prose :deep(strong),
.challenge-prose :deep(code) {
  color: var(--journal-ink);
}

@media (max-width: 1024px) {
  .solution-board {
    grid-template-columns: minmax(0, 1fr);
  }
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 16%, transparent), transparent 20rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
}

:global([data-theme='dark']) .challenge-score-card,
:global([data-theme='dark']) .challenge-panel,
:global([data-theme='dark']) .hint-list,
:global([data-theme='dark']) .challenge-btn-outline,
:global([data-theme='dark']) .challenge-input,
:global([data-theme='dark']) .solution-list-item,
:global([data-theme='dark']) .solution-preview,
:global([data-theme='dark']) .solution-tabbar {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}
</style>
