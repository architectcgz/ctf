<template>
  <section class="journal-shell journal-hero workspace-shell min-h-full">
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
      />
    </div>

    <div v-else-if="challenge" class="detail-content">
      <div class="workspace-tabbar top-tabs" role="tablist" aria-label="题目页面主切换">
        <button
          v-for="tab in workspaceTabs"
          :id="`challenge-workspace-tab-${tab.id}`"
          :key="tab.id"
          type="button"
          role="tab"
          class="workspace-tab top-tab"
          :class="{ 'workspace-tab--active': activeWorkspaceTab === tab.id }"
          :aria-selected="activeWorkspaceTab === tab.id"
          :aria-controls="`challenge-workspace-panel-${tab.id}`"
          :tabindex="activeWorkspaceTab === tab.id ? 0 : -1"
          @click="activeWorkspaceTab = tab.id"
          @keydown="handleWorkspaceTabKeydown($event, tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>

      <div
        class="detail-grid detail-grid--workspace workspace-grid"
        :class="{ 'workspace-grid--single': activeWorkspaceTab !== 'question' }"
      >
        <main class="detail-main content-pane">
          <section
            v-if="activeWorkspaceTab === 'question'"
            id="challenge-workspace-panel-question"
            class="workspace-panel panel panel--question"
            role="tabpanel"
            aria-labelledby="challenge-workspace-tab-question"
          >
            <div class="question-hero">
              <div>
                <div class="overline">Question</div>
                <h1 class="question-title">
                  {{ challenge.title }}
                </h1>
                <div class="meta-strip">
                  <span
                    class="meta-pill meta-pill--brand"
                    :style="buildMetaPillStyle(getCategoryColor(challenge.category))"
                  >
                    {{ getCategoryLabel(challenge.category) }}
                  </span>
                  <span
                    class="meta-pill"
                    :style="buildMetaPillStyle(getDifficultyColor(challenge.difficulty))"
                  >
                    {{ getDifficultyLabel(challenge.difficulty) }}
                  </span>
                  <span v-if="challenge?.is_solved" class="meta-pill"> 已解出 </span>
                  <span v-if="challenge.attachment_url" class="meta-pill"> 附件可下载 </span>
                  <span v-for="tag in challenge.tags" :key="tag" class="meta-pill">
                    {{ tag }}
                  </span>
                </div>
              </div>

              <aside class="score-rail">
                <div class="score-label">分值</div>
                <div class="score-value">{{ challenge.points }} <small>pts</small></div>
                <div class="score-note">
                  {{ challenge.attachment_url ? '当前题目包含附件。' : '当前题目无附件。' }}
                </div>
              </aside>
            </div>

            <section class="section">
              <div class="section-head">
                <div>
                  <div class="overline">Statement</div>
                  <h2 class="section-title">题目描述</h2>
                </div>
                <button
                  v-if="challenge.attachment_url"
                  type="button"
                  class="subtle-action"
                  @click="downloadAttachment"
                >
                  下载附件
                </button>
              </div>
              <!-- eslint-disable-next-line vue/no-v-html -->
              <div
                class="prose challenge-prose description max-w-none"
                v-html="sanitizedDescription"
              />
            </section>

            <section v-if="challenge.hints.length > 0" class="section">
              <div class="section-head">
                <div>
                  <div class="overline">Hints</div>
                  <h2 class="section-title">提示</h2>
                </div>
                <div class="section-hint">共 {{ challenge.hints.length }} 条</div>
              </div>
              <div class="hint-list">
                <div v-for="hint in challenge.hints" :key="hint.id" class="hint-line">
                  <div>
                    <div class="hint-label">
                      提示 {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}
                    </div>
                    <div
                      v-if="isHintExpanded(hint.level)"
                      :id="`challenge-hint-panel-${hint.id}`"
                      class="hint-copy"
                    >
                      {{ hint.content || '暂无提示内容' }}
                    </div>
                  </div>
                  <button
                    type="button"
                    class="subtle-action hint-toggle"
                    :aria-expanded="isHintExpanded(hint.level)"
                    :aria-controls="`challenge-hint-panel-${hint.id}`"
                    @click="toggleHint(hint.level)"
                  >
                    {{ isHintExpanded(hint.level) ? '收起提示' : '展开提示' }}
                  </button>
                </div>
              </div>
            </section>
          </section>

          <section
            v-else-if="activeWorkspaceTab === 'solution'"
            id="challenge-workspace-panel-solution"
            class="workspace-panel panel"
            role="tabpanel"
            aria-labelledby="challenge-workspace-tab-solution"
          >
            <section class="section section--flat">
              <div class="section-head">
                <div>
                  <div class="overline">Solutions</div>
                  <h2 class="section-title">题解区</h2>
                </div>
                <div class="section-hint">
                  推荐 {{ recommendedSolutions.length }} · 社区 {{ communitySolutions.length }}
                </div>
              </div>

              <div class="space-y-5">
                <div v-if="!challenge?.is_solved" class="inline-note inline-note--warning">
                  解出题目后可查看推荐题解与社区题解。
                </div>

                <template v-else>
                  <div class="solution-layout">
                    <div class="solution-nav">
                      <div class="solution-tabbar sub-tabs" role="tablist" aria-label="题解分类">
                        <button
                          id="challenge-solutions-tab-recommended"
                          type="button"
                          role="tab"
                          class="solution-tab sub-tab"
                          :class="{ 'solution-tab--active': activeSolutionTab === 'recommended' }"
                          :aria-selected="activeSolutionTab === 'recommended'"
                          aria-controls="challenge-solutions-panel-recommended"
                          :tabindex="activeSolutionTab === 'recommended' ? 0 : -1"
                          @click="activeSolutionTab = 'recommended'"
                          @keydown="handleSolutionTabKeydown($event, 'recommended')"
                        >
                          推荐题解
                        </button>
                        <button
                          id="challenge-solutions-tab-community"
                          type="button"
                          role="tab"
                          class="solution-tab sub-tab"
                          :class="{ 'solution-tab--active': activeSolutionTab === 'community' }"
                          :aria-selected="activeSolutionTab === 'community'"
                          aria-controls="challenge-solutions-panel-community"
                          :tabindex="activeSolutionTab === 'community' ? 0 : -1"
                          @click="activeSolutionTab = 'community'"
                          @keydown="handleSolutionTabKeydown($event, 'community')"
                        >
                          社区题解
                        </button>
                      </div>

                      <div v-if="displayedSolutionCards.length === 0" class="inline-note">
                        {{
                          activeSolutionTab === 'recommended'
                            ? '还没有推荐题解。'
                            : '还没有公开的社区题解。'
                        }}
                      </div>

                      <button
                        v-for="item in displayedSolutionCards"
                        :key="item.id"
                        type="button"
                        class="solution-list-item solution-item"
                        :class="{
                          'solution-list-item--active active': item.id === activeSolution?.id,
                        }"
                        @click="selectedSolutionId = item.id"
                      >
                        <strong>{{ item.title }}</strong>
                        <span>{{ item.authorName }} · {{ formatWriteupTime(item.updatedAt) }}</span>
                      </button>
                    </div>

                    <article
                      :id="`challenge-solutions-panel-${activeSolutionTab}`"
                      class="solution-preview"
                      role="tabpanel"
                      :aria-labelledby="`challenge-solutions-tab-${activeSolutionTab}`"
                    >
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

                      <div v-else class="inline-note">当前分组还没有可展示的题解。</div>
                    </article>
                  </div>
                </template>
              </div>
            </section>
          </section>

          <section
            v-else-if="activeWorkspaceTab === 'records'"
            id="challenge-workspace-panel-records"
            class="workspace-panel panel"
            role="tabpanel"
            aria-labelledby="challenge-workspace-tab-records"
          >
            <section class="section section--flat">
              <div class="section-head">
                <div>
                  <div class="overline">Submissions</div>
                  <h2 class="section-title">提交记录</h2>
                </div>
                <div class="section-hint">最近提交</div>
              </div>

              <div v-if="submissionRecords.length === 0" class="inline-note">
                还没有提交记录。你在右侧提交 Flag 后，新的提交结果会出现在这里。
              </div>

              <div v-else class="submission-records record-list">
                <div
                  v-for="item in submissionRecords"
                  :key="item.id"
                  class="submission-record-item record-item"
                >
                  <div class="submission-record-time record-time">
                    {{ formatSubmissionTime(item.submittedAt) }}
                  </div>
                  <div class="submission-record-answer record-answer">
                    {{ item.answer }}
                  </div>
                  <div
                    class="submission-record-status status-chip"
                    :class="`submission-record-status--${item.status}`"
                  >
                    {{ submissionStatusText(item.status) }}
                  </div>
                </div>
              </div>
            </section>
          </section>

          <section
            v-else
            id="challenge-workspace-panel-writeup"
            class="workspace-panel panel"
            role="tabpanel"
            aria-labelledby="challenge-workspace-tab-writeup"
          >
            <section class="section section--flat">
              <div class="section-head">
                <div>
                  <div class="overline">My Writeup</div>
                  <h2 class="section-title">我的复盘</h2>
                </div>
                <div class="section-hint">
                  解题过程复盘 · {{ challenge?.is_solved ? '可发布到社区' : '仅可保存草稿' }}
                </div>
              </div>

              <div
                v-if="myWriteup?.visibility_status === 'hidden'"
                class="inline-note inline-note--warning"
              >
                当前题解已被教师或管理员隐藏，仅你自己可见。
              </div>

              <div class="meta-strip meta-strip--compact">
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
                  v-else-if="
                    myWriteup?.submission_status === 'published' ||
                    myWriteup?.submission_status === 'submitted'
                  "
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

              <form class="writeup-form" @submit.prevent>
                <div class="field">
                  <label for="challenge-writeup-title">标题</label>
                  <input
                    id="challenge-writeup-title"
                    v-model="writeupTitle"
                    type="text"
                    maxlength="256"
                    placeholder="例如：从回显异常到拿到 flag 的完整链路"
                    class="challenge-input"
                  />
                </div>

                <div class="field">
                  <label for="challenge-writeup-content">正文</label>
                  <textarea
                    id="challenge-writeup-content"
                    v-model="writeupContent"
                    rows="10"
                    placeholder="建议按『题目理解 → 利用过程 → 核心 payload / 证据 → 踩坑点』组织。"
                    class="challenge-input writeup-textarea"
                  />
                </div>

                <div class="writeup-foot">
                  <div class="writeup-footnote">
                    {{
                      submissionLoading
                        ? '正在同步你的题解...'
                        : myWriteup?.updated_at
                          ? `最近更新：${formatWriteupTime(myWriteup.updated_at)}`
                          : '还没有提交记录，可以先保存草稿。'
                    }}
                  </div>
                  <div class="writeup-actions">
                    <button
                      type="button"
                      :disabled="submissionLoading || submissionSaving !== null"
                      class="ghost-action"
                      @click="saveWriteup('draft')"
                    >
                      {{ submissionSaving === 'draft' ? '保存中...' : '保存草稿' }}
                    </button>
                    <button
                      type="button"
                      :disabled="
                        submissionLoading || submissionSaving !== null || !challenge?.is_solved
                      "
                      class="primary-action disabled:cursor-not-allowed disabled:opacity-50"
                      @click="saveWriteup('published')"
                    >
                      {{ submissionSaving === 'published' ? '发布中...' : '发布题解' }}
                    </button>
                  </div>
                </div>
              </form>
            </section>
          </section>
        </main>

        <aside v-if="activeWorkspaceTab === 'question'" class="detail-aside tool-pane">
          <div class="tool-pane-inner">
            <section class="tool-group">
              <div>
                <div class="overline">Primary Action</div>
                <h2 class="tool-title">Flag 提交</h2>
                <p class="tool-copy">输入当前题目的 Flag 并提交验证。</p>
              </div>
              <span
                v-if="challenge?.is_solved"
                class="writeup-status-pill writeup-status-pill--success"
              >
                已通过
              </span>
              <div class="flag-field">
                <label for="challenge-flag-input" class="flag-label"> Flag </label>
                <div class="flag-row">
                  <input
                    id="challenge-flag-input"
                    v-model="flagInput"
                    type="text"
                    aria-label="Flag"
                    :placeholder="submitPlaceholder"
                    :disabled="challenge?.is_solved"
                    class="challenge-input flag-input disabled:cursor-not-allowed disabled:opacity-50"
                    :class="submitInputClass"
                    @keyup.enter="submitFlagHandler"
                  />
                  <button
                    type="button"
                    :disabled="challenge?.is_solved || submitting"
                    class="primary-action disabled:cursor-not-allowed disabled:opacity-50"
                    @click="submitFlagHandler"
                  >
                    {{ submitting ? '提交中...' : '提交' }}
                  </button>
                </div>
              </div>
              <div
                v-if="submitResult"
                class="status-inline"
                :class="`status-inline--${submitResult.variant}`"
              >
                <span class="status-dot" />
                {{ submitResult.message }}
              </div>
            </section>

            <ChallengeInstanceCard
              v-if="needTarget"
              :instance="instance"
              :loading="instanceLoading"
              :creating="instanceCreating"
              :opening="instanceOpening"
              :extending="instanceExtending"
              :destroying="instanceDestroying"
              :challenge-solved="Boolean(challenge?.is_solved)"
              @start="startInstance"
              @open="openInstance"
              @extend="extendChallengeInstance"
              @destroy="destroyChallengeInstance"
            />
            <section
              v-else
              class="tool-group detail-aside-empty text-sm text-[var(--color-success)]"
            >
              该题目不需要靶机，可直接分析题面并提交 Flag。
            </section>
          </div>
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

type WorkspaceTab = 'question' | 'solution' | 'records' | 'writeup'
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

interface SubmissionRecordItem {
  id: string
  answer: string
  status: 'correct' | 'incorrect' | 'pending_review' | 'error'
  message: string
  submittedAt?: string
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
const expandedHintLevels = ref<number[]>([])
const activeWorkspaceTab = ref<WorkspaceTab>('question')
const submitResult = ref<{
  variant: 'success' | 'error' | 'pending'
  className: string
  message: string
} | null>(null)
const activeSolutionTab = ref<SolutionTab>('recommended')
const selectedSolutionId = ref<string | null>(null)
const submissionRecords = ref<SubmissionRecordItem[]>([])
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

const workspaceTabs: Array<{ id: WorkspaceTab; label: string }> = [
  { id: 'question', label: '题目' },
  { id: 'solution', label: '题解' },
  { id: 'records', label: '提交记录' },
  { id: 'writeup', label: '我的复盘' },
]
const solutionTabs: SolutionTab[] = ['recommended', 'community']

function renderRichContent(source?: string): string {
  if (!source) return ''
  const html = marked.parse(source, {
    gfm: true,
    breaks: true,
  })
  return sanitizeHtml(typeof html === 'string' ? html : source)
}

function buildMetaPillStyle(color: string): Record<string, string> {
  return {
    borderColor: `color-mix(in srgb, ${color} 18%, transparent)`,
    backgroundColor: `color-mix(in srgb, ${color} 12%, transparent)`,
    color,
  }
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
  activeSolutionTab.value === 'recommended'
    ? recommendedSolutionCards.value
    : communitySolutionCards.value
)

const activeSolution = computed(() => {
  if (displayedSolutionCards.value.length === 0) return null
  return (
    displayedSolutionCards.value.find((item) => item.id === selectedSolutionId.value) ??
    displayedSolutionCards.value[0]
  )
})

const sanitizedActiveSolutionContent = computed(() =>
  renderRichContent(activeSolution.value?.content)
)

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
      return 'border-[var(--journal-accent)]'
  }
})

function clearSolutions(): void {
  recommendedSolutions.value = []
  communitySolutions.value = []
  activeSolutionTab.value = 'recommended'
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

function isHintExpanded(level: number): boolean {
  return expandedHintLevels.value.includes(level)
}

function toggleHint(level: number): void {
  if (isHintExpanded(level)) {
    expandedHintLevels.value = expandedHintLevels.value.filter((item) => item !== level)
    return
  }
  expandedHintLevels.value = [...expandedHintLevels.value, level]
}

function focusTab(id: string): void {
  requestAnimationFrame(() => {
    document.getElementById(id)?.focus()
  })
}

function handleWorkspaceTabKeydown(event: KeyboardEvent, currentTab: WorkspaceTab): void {
  const currentIndex = workspaceTabs.findIndex((item) => item.id === currentTab)
  if (currentIndex < 0) return

  if (event.key === 'ArrowRight') {
    const nextTab = workspaceTabs[(currentIndex + 1) % workspaceTabs.length]
    activeWorkspaceTab.value = nextTab.id
    focusTab(`challenge-workspace-tab-${nextTab.id}`)
  } else if (event.key === 'ArrowLeft') {
    const nextTab = workspaceTabs[(currentIndex - 1 + workspaceTabs.length) % workspaceTabs.length]
    activeWorkspaceTab.value = nextTab.id
    focusTab(`challenge-workspace-tab-${nextTab.id}`)
  } else if (event.key === 'Home') {
    activeWorkspaceTab.value = workspaceTabs[0].id
    focusTab(`challenge-workspace-tab-${workspaceTabs[0].id}`)
  } else if (event.key === 'End') {
    const lastTab = workspaceTabs[workspaceTabs.length - 1]
    activeWorkspaceTab.value = lastTab.id
    focusTab(`challenge-workspace-tab-${lastTab.id}`)
  } else {
    return
  }

  event.preventDefault()
}

function handleSolutionTabKeydown(event: KeyboardEvent, currentTab: SolutionTab): void {
  const currentIndex = solutionTabs.findIndex((item) => item === currentTab)
  if (currentIndex < 0) return

  if (event.key === 'ArrowRight') {
    activeSolutionTab.value = solutionTabs[(currentIndex + 1) % solutionTabs.length]
  } else if (event.key === 'ArrowLeft') {
    activeSolutionTab.value =
      solutionTabs[(currentIndex - 1 + solutionTabs.length) % solutionTabs.length]
  } else if (event.key === 'Home') {
    activeSolutionTab.value = solutionTabs[0]
  } else if (event.key === 'End') {
    activeSolutionTab.value = solutionTabs[solutionTabs.length - 1]
  } else {
    return
  }

  focusTab(`challenge-solutions-tab-${activeSolutionTab.value}`)
  event.preventDefault()
}

async function submitFlagHandler(): Promise<void> {
  if (!challenge.value || !flagInput.value.trim()) return

  const answer = flagInput.value.trim()
  submitting.value = true
  submitResult.value = null
  try {
    const result = await submitFlag(challenge.value.id, answer)
    submissionRecords.value = [
      {
        id: `${result.submitted_at}-${submissionRecords.value.length}`,
        answer,
        status: result.status,
        message: result.message,
        submittedAt: result.submitted_at,
      },
      ...submissionRecords.value,
    ]
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
    submissionRecords.value = [
      {
        id: `error-${Date.now()}`,
        answer,
        status: 'error',
        message: '提交失败，请重试',
        submittedAt: new Date().toISOString(),
      },
      ...submissionRecords.value,
    ]
    submitResult.value = {
      variant: 'error',
      className: 'text-[var(--color-danger)]',
      message: '提交失败，请重试',
    }
  } finally {
    submitting.value = false
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

function submissionStatusLabel(status?: SubmissionWriteupStatus): string {
  if (status === 'draft') return '草稿'
  if (status === 'published' || status === 'submitted') return '已发布'
  return '未开始'
}

function submissionStatusText(status: SubmissionRecordItem['status']): string {
  if (status === 'correct') return '正确'
  if (status === 'incorrect') return '错误答案'
  if (status === 'pending_review') return '待审核'
  if (status === 'error') return '提交失败'
  return '未知'
}

function visibilityStatusLabel(status?: SubmissionWriteupVisibilityStatus): string {
  if (status === 'hidden') return '已隐藏'
  if (
    myWriteup.value?.submission_status === 'published' ||
    myWriteup.value?.submission_status === 'submitted'
  ) {
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

function formatSubmissionTime(value?: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
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
    web: 'var(--challenge-tone-web)',
    pwn: 'var(--challenge-tone-pwn)',
    reverse: 'var(--challenge-tone-reverse)',
    crypto: 'var(--challenge-tone-crypto)',
    misc: 'var(--challenge-tone-misc)',
    forensics: 'var(--challenge-tone-forensics)',
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
    beginner: 'var(--challenge-tone-beginner)',
    easy: 'var(--challenge-tone-easy)',
    medium: 'var(--challenge-tone-medium)',
    hard: 'var(--challenge-tone-hard)',
    insane: 'var(--challenge-tone-insane)',
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
    if (
      activeSolutionTab.value === 'recommended' &&
      recommended.length === 0 &&
      community.length > 0
    ) {
      activeSolutionTab.value = 'community'
    } else if (
      activeSolutionTab.value === 'community' &&
      community.length === 0 &&
      recommended.length > 0
    ) {
      activeSolutionTab.value = 'recommended'
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
    expandedHintLevels.value = []
    activeWorkspaceTab.value = 'question'
    submitResult.value = null
    submissionRecords.value = []
    void Promise.all([loadChallenge(), loadMyWriteupSubmission()])
  },
  { immediate: true }
)
</script>

<style scoped>
@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 500;
  font-display: swap;
  src: url('/fonts/ibm-plex-sans-500.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 600;
  font-display: swap;
  src: url('/fonts/ibm-plex-sans-600.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 700;
  font-display: swap;
  src: url('/fonts/ibm-plex-sans-700.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 500;
  font-display: swap;
  src: url('/fonts/ibm-plex-mono-500.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 600;
  font-display: swap;
  src: url('/fonts/ibm-plex-mono-600.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 700;
  font-display: swap;
  src: url('/fonts/ibm-plex-mono-700.woff2') format('woff2');
}

.journal-shell {
  color-scheme: light;
  --bg-page: oklch(97.8% 0.006 247);
  --bg-shell: color-mix(in srgb, white 88%, oklch(95.5% 0.011 245));
  --bg-panel: color-mix(in srgb, white 92%, oklch(94.9% 0.014 245));
  --bg-muted: color-mix(in srgb, white 80%, oklch(95.1% 0.01 245));
  --line-soft: color-mix(in srgb, oklch(38% 0.014 252) 12%, transparent);
  --line-strong: color-mix(in srgb, oklch(38% 0.014 252) 20%, transparent);
  --text-main: oklch(24% 0.014 252);
  --text-subtle: oklch(49% 0.016 252);
  --text-faint: oklch(61% 0.012 252);
  --brand: oklch(52% 0.12 254);
  --brand-soft: color-mix(in srgb, var(--brand) 8%, transparent);
  --brand-soft-strong: color-mix(in srgb, var(--brand) 14%, transparent);
  --brand-ink: color-mix(in srgb, var(--brand) 82%, var(--text-main));
  --success: oklch(56% 0.13 154);
  --warning: oklch(68% 0.14 82);
  --danger: oklch(58% 0.16 28);
  --shadow-shell: 0 24px 84px rgba(13, 23, 39, 0.06);
  --radius-xl: 28px;
  --radius-lg: 18px;
  --font-sans:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
  --font-mono: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
  --journal-ink: var(--text-main);
  --journal-muted: var(--text-subtle);
  --journal-faint: var(--text-faint);
  --journal-accent: var(--brand);
  --journal-accent-strong: var(--brand-ink);
  --journal-accent-soft: var(--brand-soft);
  --journal-line-soft: var(--line-soft);
  --journal-line-strong: var(--line-strong);
  --journal-border: var(--line-soft);
  --journal-surface: var(--bg-shell);
  --journal-surface-panel: var(--bg-panel);
  --journal-surface-muted: var(--bg-muted);
  --journal-shadow: var(--shadow-shell);
  --journal-success-ink: color-mix(in srgb, var(--color-success) 80%, var(--journal-ink));
  --journal-success-soft: color-mix(in srgb, var(--color-success) 12%, transparent);
  --journal-warning-ink: color-mix(in srgb, var(--color-warning) 88%, var(--journal-ink));
  --journal-warning-soft: color-mix(in srgb, var(--color-warning) 12%, transparent);
  --journal-danger-ink: color-mix(in srgb, var(--color-danger) 82%, var(--journal-ink));
  --journal-danger-soft: color-mix(in srgb, var(--color-danger) 10%, transparent);
  --challenge-tone-web: color-mix(in srgb, var(--color-primary) 82%, var(--journal-ink));
  --challenge-tone-pwn: color-mix(in srgb, var(--color-danger) 78%, var(--journal-ink));
  --challenge-tone-reverse: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --challenge-tone-crypto: color-mix(in srgb, var(--color-warning) 82%, var(--journal-ink));
  --challenge-tone-misc: color-mix(in srgb, var(--color-success) 76%, var(--journal-ink));
  --challenge-tone-forensics: color-mix(in srgb, var(--color-primary) 62%, var(--journal-ink));
  --challenge-tone-beginner: var(--challenge-tone-misc);
  --challenge-tone-easy: var(--challenge-tone-web);
  --challenge-tone-medium: var(--challenge-tone-crypto);
  --challenge-tone-hard: var(--challenge-tone-pwn);
  --challenge-tone-insane: var(--challenge-tone-reverse);
  --page-top-tabs-gap: 28px;
  --page-top-tabs-margin: 10px 0 0;
  --page-top-tabs-padding: 0 28px;
  --page-top-tabs-border: var(--line-soft);
  --page-top-tab-min-height: 52px;
  --page-top-tab-padding: 10px 0 13px;
  --page-top-tab-font-size: 15px;
  --page-top-tab-font-weight: 600;
  --page-top-tab-color: var(--text-faint);
  --page-top-tab-active-color: var(--brand-ink);
  --page-top-tab-active-border: var(--brand);
}

.workspace-shell {
  --workspace-shell-border: var(--journal-line-soft);
  --workspace-shell-page: var(--bg-page);
  --workspace-shell-bg: var(--bg-shell);
  --workspace-brand: var(--brand);
  --workspace-brand-ink: var(--brand-ink);
  --workspace-brand-soft: var(--brand-soft);
  --workspace-faint: var(--text-faint);
  --workspace-shadow-shell: var(--journal-shadow);
  --workspace-radius-xl: var(--radius-xl);
  --workspace-font-sans: var(--font-sans);
  min-height: max(100%, calc(100vh - 5rem));
  flex: 1 1 auto;
  color: var(--text-main);
}

.workspace-shell,
.workspace-shell button,
.workspace-shell input,
.workspace-shell textarea {
  font-family: var(--font-sans);
}

.workspace-shell code,
.workspace-shell pre,
.workspace-shell .flag-input,
.workspace-shell .record-time,
.workspace-shell .record-answer,
.workspace-shell .submission-record-time,
.workspace-shell .submission-record-answer,
.workspace-shell .score-value {
  font-family: var(--font-mono) !important;
}

.detail-content {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
}

.overline {
  display: inline-block;
  border: 0 !important;
  box-shadow: none !important;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.18em;
  line-height: 1;
  text-decoration: none !important;
  text-decoration-line: none !important;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--brand) 68%, var(--text-faint));
}

.top-note,
.section-hint,
.tool-copy,
.writeup-footnote {
  font-size: 13px;
  line-height: 1.75;
  color: var(--text-faint);
}

.top-tabs::-webkit-scrollbar,
.sub-tabs::-webkit-scrollbar {
  display: none;
}

.workspace-tab--active,
.top-tab.workspace-tab--active {
  border-bottom-color: var(--brand);
  color: var(--brand-ink);
}

.detail-grid,
.workspace-grid {
  display: grid;
  flex: 1 1 auto;
  min-height: 0;
  grid-template-columns: minmax(0, 1.34fr) minmax(320px, 0.66fr);
  align-items: stretch;
}

.workspace-grid--single {
  grid-template-columns: minmax(0, 1fr);
}

.detail-main,
.detail-aside,
.content-pane,
.tool-pane {
  min-width: 0;
}

.detail-main,
.content-pane {
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 28px;
}

.tool-pane {
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 28px;
  border-left: 1px solid var(--line-soft);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--bg-panel) 95%, white),
    color-mix(in srgb, var(--bg-shell) 92%, white)
  );
}

.tool-pane-inner {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 100%;
  position: sticky;
  top: 28px;
}

.workspace-panel,
.panel {
  display: block;
  min-height: 100%;
  animation: rise 280ms cubic-bezier(0.22, 1, 0.36, 1);
}

.question-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 200px;
  gap: 26px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--line-soft);
}

.question-title {
  margin: 12px 0 0;
  font-size: clamp(30px, 4vw, 46px);
  line-height: 1.04;
  letter-spacing: -0.03em;
  color: var(--text-main);
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 16px;
}

.meta-strip--compact {
  margin-top: 0;
  margin-bottom: 1rem;
}

.meta-pill,
.writeup-status-pill,
.status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 14px;
  border: 1px solid var(--line-soft);
  border-radius: 999px;
  background: color-mix(in srgb, var(--bg-panel) 72%, transparent);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-subtle);
}

.meta-pill--brand {
  border-color: color-mix(in srgb, var(--brand) 20%, transparent);
  background: var(--brand-soft);
  color: var(--brand-ink);
}

.score-rail {
  padding-left: 22px;
  border-left: 1px solid var(--line-soft);
}

.score-label {
  font-size: 11px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.score-value {
  margin-top: 8px;
  color: var(--text-main);
  font: 700 34px/1 var(--font-mono);
}

.score-value small {
  font-size: 16px;
  color: var(--text-faint);
}

.score-note {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--line-soft);
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-subtle);
}

.section {
  padding-top: 24px;
  border-top: 1px solid var(--line-soft);
}

.section--flat,
.section:first-child {
  padding-top: 0;
  border-top: 0;
}

.panel--question > .section:first-of-type {
  border-top: 0;
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.section-title {
  margin: 10px 0 0;
  font-size: 20px;
  line-height: 1.2;
  color: var(--text-main);
}

.description,
.solution-preview,
.tool-copy {
  color: var(--journal-muted);
}

.description {
  font-size: 15px;
  line-height: 1.92;
  color: var(--text-subtle);
}

.challenge-prose :deep(p),
.challenge-prose :deep(ul),
.challenge-prose :deep(ol) {
  margin-bottom: 1rem;
}

.challenge-prose :deep(pre) {
  overflow: auto;
  margin: 20px 0;
  padding: 18px 20px;
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--bg-panel) 72%, white);
  color: var(--text-main);
  font: 13px/1.7 var(--font-mono);
}

.challenge-prose :deep(h1),
.challenge-prose :deep(h2),
.challenge-prose :deep(h3),
.challenge-prose :deep(strong),
.challenge-prose :deep(code) {
  color: var(--journal-ink);
}

.hint-list {
  display: flex;
  flex-direction: column;
}

.hint-line {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 14px;
  padding: 16px 0;
  border-top: 1px dashed var(--line-soft);
}

.hint-line:first-of-type {
  padding-top: 0;
  border-top: 0;
}

.hint-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main);
}

.hint-copy {
  margin-top: 7px;
  font-size: 14px;
  line-height: 1.8;
  color: var(--text-subtle);
}

.solution-layout {
  display: grid;
  grid-template-columns: minmax(240px, 0.54fr) minmax(0, 1fr);
  gap: 24px;
}

.solution-nav {
  padding-right: 20px;
  border-right: 1px solid var(--line-soft);
}

.sub-tabs,
.solution-tabbar {
  display: flex;
  gap: 18px;
  padding-bottom: 10px;
  overflow-x: auto;
  border-bottom: 1px solid var(--line-soft);
  scrollbar-width: none;
}

.sub-tab,
.solution-tab {
  font: 600 14px/1.2 var(--font-sans);
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  padding: 0 0 8px;
  font-weight: 600;
  color: var(--text-faint);
}

.solution-tab--active,
.sub-tab.solution-tab--active {
  border-bottom-color: var(--journal-accent);
  color: var(--journal-accent-strong);
}

.solution-item,
.solution-list-item {
  width: 100%;
  text-align: left;
  padding: 15px 0 16px 14px;
  border: 0;
  border-left: 2px solid transparent;
  border-bottom: 1px solid var(--line-soft);
  background: transparent;
}

.solution-item strong,
.solution-list-item strong {
  display: block;
  font-size: 14px;
  color: var(--text-main);
}

.solution-item span,
.solution-list-item span {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-faint);
}

.solution-item.active,
.solution-list-item--active,
.solution-list-item:hover {
  border-left-color: color-mix(in srgb, var(--journal-accent) 40%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
}

.solution-preview {
  min-height: 22rem;
  font-size: 14px;
  line-height: 1.9;
  color: var(--text-subtle);
}

.solution-preview__content {
  min-height: 15rem;
}

.solution-preview__content :deep(h1),
.solution-preview__content :deep(h2),
.solution-preview__content :deep(h3) {
  margin-top: 1.2rem;
}

.inline-note {
  padding-left: 1rem;
  border-left: 2px solid var(--line-soft);
  font-size: 0.9rem;
  line-height: 1.8;
  color: var(--text-subtle);
}

.inline-note--warning {
  border-left-color: color-mix(in srgb, var(--color-warning) 34%, transparent);
  color: var(--journal-warning-ink);
}

.record-list,
.submission-records {
  display: grid;
  gap: 0;
  border-top: 1px solid var(--line-soft);
}

.record-item,
.submission-record-item {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr) auto;
  gap: 18px;
  align-items: center;
  padding: 18px 0;
  border-bottom: 1px solid var(--line-soft);
}

.record-time,
.submission-record-time {
  color: var(--text-faint);
  font: 500 13px/1.6 var(--font-mono);
}

.record-answer,
.submission-record-answer {
  min-width: 0;
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--text-subtle);
  word-break: break-all;
}

.submission-record-status--correct {
  background: var(--journal-success-soft);
  color: var(--journal-success-ink);
}

.submission-record-status--incorrect,
.submission-record-status--error {
  background: var(--journal-danger-soft);
  color: var(--journal-danger-ink);
}

.submission-record-status--pending_review {
  background: var(--journal-warning-soft);
  color: var(--journal-warning-ink);
}

.writeup-form {
  display: grid;
  gap: 16px;
}

.field label,
.flag-label {
  display: block;
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.challenge-input,
.field input,
.field textarea,
.flag-input {
  width: 100%;
  border: 1px solid var(--line-strong);
  border-radius: 14px;
  background: white;
  color: var(--text-main);
}

.field input,
.flag-input {
  min-height: 50px;
  padding: 0 14px;
}

.field textarea {
  min-height: 260px;
  padding: 14px;
  resize: vertical;
}

.field input,
.field textarea {
  font: 500 14px/1.6 var(--font-sans);
}

.flag-input {
  font: 500 15px/1 var(--font-mono);
}

.challenge-input::placeholder {
  color: var(--journal-faint);
}

.writeup-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.writeup-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.flag-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
}

.tool-group + .tool-group,
:deep(.instance-shell) {
  margin-top: 1.6rem;
}

.tool-group + .tool-group {
  margin-top: 26px;
  padding-top: 26px;
  border-top: 1px solid var(--line-soft);
}

.tool-title {
  margin: 10px 0 0;
  font-size: 18px;
  line-height: 1.25;
  color: var(--text-main);
}

.tool-copy {
  margin-top: 8px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-subtle);
}

.primary-action,
.ghost-action,
.subtle-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 48px;
  padding: 0 16px;
  border-radius: 14px;
  font: 600 14px/1 var(--font-sans);
}

.primary-action {
  border: 0;
  background: var(--brand);
  color: white;
}

.ghost-action,
.subtle-action {
  border: 1px solid var(--line-strong);
  background: transparent;
  color: var(--text-main);
}

.status-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  font-size: 14px;
  color: var(--text-subtle);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--warning);
}

.status-inline--success {
  color: var(--journal-success-ink);
}

.status-inline--success .status-dot {
  background: var(--color-success);
}

.status-inline--pending {
  color: var(--journal-warning-ink);
}

.status-inline--pending .status-dot {
  background: var(--color-warning);
}

.status-inline--error {
  color: var(--journal-danger-ink);
}

.status-inline--error .status-dot {
  background: var(--color-danger);
}

.writeup-status-pill--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: var(--journal-accent-soft);
  color: var(--journal-accent-strong);
}

.writeup-status-pill--success {
  border-color: color-mix(in srgb, var(--color-success) 18%, transparent);
  background: var(--journal-success-soft);
  color: var(--journal-success-ink);
}

.writeup-status-pill--warning {
  border-color: color-mix(in srgb, var(--color-warning) 18%, transparent);
  background: var(--journal-warning-soft);
  color: var(--journal-warning-ink);
}

.writeup-status-pill--muted {
  border-color: var(--line-soft);
  background: var(--bg-muted);
  color: var(--text-subtle);
}

.top-tab:hover,
.sub-tab:hover,
.solution-item:hover,
.primary-action:hover,
.ghost-action:hover,
.subtle-action:hover,
.workspace-tab:hover {
  transform: translateY(-1px);
}

.workspace-tab:focus-visible,
.solution-tab:focus-visible,
.solution-list-item:focus-visible,
.primary-action:focus-visible,
.ghost-action:focus-visible,
.subtle-action:focus-visible,
.challenge-input:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--brand) 44%, white);
  outline-offset: 3px;
}

@keyframes rise {
  from {
    opacity: 0;
    transform: translateY(16px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1080px) {
  .detail-grid,
  .workspace-grid {
    flex: initial;
    grid-template-columns: minmax(0, 1fr);
  }

  .tool-pane {
    border-left: 0;
    border-top: 1px solid var(--journal-line-soft);
  }

  .tool-pane-inner {
    min-height: 0;
    position: static;
  }

  .question-hero {
    grid-template-columns: minmax(0, 1fr);
  }

  .score-rail {
    padding-left: 0;
    padding-top: 18px;
    border-left: 0;
    border-top: 1px solid var(--line-soft);
  }
}

@media (max-width: 760px) {
  .workspace-topbar,
  .top-tabs,
  .workspace-tabbar,
  .content-pane,
  .tool-pane {
    padding-left: 1.1rem;
    padding-right: 1.1rem;
  }

  .top-tabs,
  .workspace-tabbar {
    gap: 22px;
  }

  .solution-layout,
  .flag-row,
  .record-item,
  .submission-record-item,
  .writeup-foot {
    grid-template-columns: minmax(0, 1fr);
  }

  .solution-nav {
    padding-right: 0;
    border-right: 0;
  }

  .record-item,
  .submission-record-item {
    gap: 0.6rem;
  }

  .writeup-actions {
    flex-direction: column;
  }
}

@media (max-width: 767px) {
  .workspace-shell {
    min-height: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation: none !important;
    transition-duration: 0.01ms !important;
  }
}

:global([data-theme='dark']) .journal-shell {
  --bg-shell: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --bg-panel: color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base));
  --bg-muted: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --line-soft: color-mix(in srgb, var(--color-border-default) 78%, transparent);
  --line-strong: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  --text-main: var(--color-text-primary);
  --text-subtle: var(--color-text-secondary);
  --text-faint: color-mix(in srgb, var(--color-text-secondary) 82%, var(--color-bg-base));
  --brand: color-mix(in srgb, var(--color-primary) 88%, white);
  --brand-soft: color-mix(in srgb, var(--brand) 12%, transparent);
  --brand-ink: color-mix(in srgb, var(--brand) 84%, var(--text-main));
}

:global([data-theme='dark']) .workspace-shell {
  --workspace-shell-radial-strength: 14%;
  --workspace-shell-radial-size: 24rem;
  --workspace-shell-top-strength: 97%;
}

:global([data-theme='dark']) .challenge-input,
:global([data-theme='dark']) .field input,
:global([data-theme='dark']) .field textarea,
:global([data-theme='dark']) .flag-input {
  background: color-mix(in srgb, var(--bg-panel) 96%, var(--color-bg-base));
}
</style>
