<template>
  <section class="journal-shell journal-shell-user journal-hero workspace-shell min-h-full">
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
      class="detail-content"
    >
      <div
        class="workspace-tabbar top-tabs"
        role="tablist"
        aria-label="题目页面主切换"
      >
        <button
          v-for="(tab, index) in workspaceTabs"
          :id="`challenge-workspace-tab-${tab.id}`"
          :key="tab.id"
          :ref="(element) => setTabButtonRef(tab.id, element as HTMLButtonElement | null)"
          type="button"
          role="tab"
          class="workspace-tab top-tab"
          :class="{ active: activeWorkspaceTab === tab.id }"
          :aria-selected="activeWorkspaceTab === tab.id"
          :aria-controls="`challenge-workspace-panel-${tab.id}`"
          :tabindex="activeWorkspaceTab === tab.id ? 0 : -1"
          @click="selectWorkspaceTab(tab.id)"
          @keydown="handleWorkspaceTabKeydown($event, index)"
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
              <div class="question-hero-main">
                <div class="workspace-overline">
                  Question
                </div>
                <h1 class="question-title workspace-page-title">
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
                  <span
                    v-if="challenge?.is_solved"
                    class="meta-pill"
                  > 已解出 </span>
                  <span
                    v-if="challenge.attachment_url"
                    class="meta-pill"
                  > 附件可下载 </span>
                  <span
                    v-for="tag in challenge.tags"
                    :key="tag"
                    class="meta-pill"
                  >
                    {{ tag }}
                  </span>
                </div>
              </div>

              <aside class="score-rail">
                <div class="score-label">
                  分值
                </div>
                <div class="score-value">
                  {{ challenge.points }} <small>pts</small>
                </div>
                <div class="score-note">
                  {{ challenge.attachment_url ? '当前题目包含附件。' : '当前题目无附件。' }}
                </div>
              </aside>
            </div>

            <section class="section">
              <div class="section-head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    Statement
                  </div>
                  <h2 class="section-title workspace-tab-heading__title">
                    题目描述
                  </h2>
                </div>
                <button
                  v-if="challenge.attachment_url"
                  type="button"
                  class="ui-btn ui-btn--secondary"
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

            <section
              v-if="challenge.hints.length > 0"
              class="section"
            >
              <div class="section-head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    Hints
                  </div>
                  <h2 class="section-title workspace-tab-heading__title">
                    提示
                  </h2>
                </div>
                <div class="section-hint">
                  共 {{ challenge.hints.length }} 条
                </div>
              </div>
              <div class="hint-list">
                <div
                  v-for="hint in challenge.hints"
                  :key="hint.id"
                  class="hint-line"
                >
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
                    class="ui-btn ui-btn--sm ui-btn--ghost hint-toggle"
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
              <div class="section-head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    Solutions
                  </div>
                  <h2 class="section-title workspace-tab-heading__title">
                    题解区
                  </h2>
                </div>
                <div class="section-hint">
                  推荐 {{ recommendedSolutions.length }} · 社区 {{ communitySolutions.length }}
                </div>
              </div>

              <div class="space-y-5">
                <div
                  v-if="!challenge?.is_solved"
                  class="inline-note inline-note--warning"
                >
                  解出题目后可查看推荐题解与社区题解。
                </div>

                <template v-else>
                  <div class="solution-layout">
                    <div class="solution-nav">
                      <div
                        class="solution-tabbar top-tabs challenge-subtabs"
                        role="tablist"
                        aria-label="题解分类"
                      >
                        <button
                          id="challenge-solutions-tab-recommended"
                          :ref="
                            (element) =>
                              setSolutionTabButtonRef(
                                'recommended',
                                element as HTMLButtonElement | null
                              )
                          "
                          type="button"
                          role="tab"
                          class="solution-tab top-tab challenge-subtab"
                          :class="{ active: activeSolutionTab === 'recommended' }"
                          :aria-selected="activeSolutionTab === 'recommended'"
                          aria-controls="challenge-solutions-panel-recommended"
                          :tabindex="activeSolutionTab === 'recommended' ? 0 : -1"
                          @click="selectSolutionTab('recommended')"
                          @keydown="handleSolutionTabKeydown($event, 0)"
                        >
                          推荐题解
                        </button>
                        <button
                          id="challenge-solutions-tab-community"
                          :ref="
                            (element) =>
                              setSolutionTabButtonRef(
                                'community',
                                element as HTMLButtonElement | null
                              )
                          "
                          type="button"
                          role="tab"
                          class="solution-tab top-tab challenge-subtab"
                          :class="{ active: activeSolutionTab === 'community' }"
                          :aria-selected="activeSolutionTab === 'community'"
                          aria-controls="challenge-solutions-panel-community"
                          :tabindex="activeSolutionTab === 'community' ? 0 : -1"
                          @click="selectSolutionTab('community')"
                          @keydown="handleSolutionTabKeydown($event, 1)"
                        >
                          社区题解
                        </button>
                      </div>

                      <div
                        v-if="displayedSolutionCards.length === 0"
                        class="inline-note"
                      >
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

                      <div
                        v-else
                        class="inline-note"
                      >
                        当前分组还没有可展示的题解。
                      </div>
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
              <div class="section-head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    Submissions
                  </div>
                  <h2 class="section-title workspace-tab-heading__title">
                    提交记录
                  </h2>
                </div>
                <div class="section-hint">
                  最近提交
                </div>
              </div>

              <div
                v-if="submissionRecords.length === 0"
                class="inline-note"
              >
                还没有提交记录。你在右侧提交 Flag 后，新的提交结果会出现在这里。
              </div>

              <div
                v-else
                class="submission-records record-list"
              >
                <div
                  v-for="item in paginatedSubmissionRecords"
                  :key="item.id"
                  class="submission-record-item record-item"
                >
                  <div class="submission-record-time record-time">
                    {{ formatSubmissionTime(item.submittedAt) }}
                  </div>
                  <div class="submission-record-answer record-answer">
                    {{ item.answer || submissionRecordMessage(item.status) }}
                  </div>
                  <div
                    class="submission-record-status status-chip"
                    :class="`submission-record-status--${item.status}`"
                  >
                    {{ submissionStatusText(item.status) }}
                  </div>
                </div>
              </div>

              <div
                v-if="submissionRecordTotal > 0"
                class="submission-pagination workspace-directory-pagination"
              >
                <PagePaginationControls
                  :page="submissionRecordPage"
                  :total-pages="submissionRecordTotalPages"
                  :total="submissionRecordTotal"
                  :total-label="`共 ${submissionRecordTotal} 条提交`"
                  @change-page="changeSubmissionRecordPage"
                />
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
              <div class="section-head workspace-tab-heading">
                <div class="workspace-tab-heading__main">
                  <div class="workspace-overline">
                    My Writeup
                  </div>
                  <h2 class="section-title workspace-tab-heading__title">
                    编写题解
                  </h2>
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

              <form
                class="writeup-form"
                @submit.prevent
              >
                <div class="field">
                  <label for="challenge-writeup-title">标题</label>
                  <div class="ui-control-wrap">
                    <input
                      id="challenge-writeup-title"
                      v-model="writeupTitle"
                      type="text"
                      maxlength="256"
                      placeholder="例如：从回显异常到拿到 flag 的完整链路"
                      class="ui-control challenge-input"
                    >
                  </div>
                </div>

                <div class="field">
                  <label for="challenge-writeup-content">正文</label>
                  <div class="ui-control-wrap writeup-textarea-wrap">
                    <textarea
                      id="challenge-writeup-content"
                      v-model="writeupContent"
                      rows="10"
                      placeholder="建议按『题目理解 → 利用过程 → 核心 payload / 证据 → 踩坑点』组织。"
                      class="ui-control challenge-input writeup-textarea"
                    />
                  </div>
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
                      class="ui-btn ui-btn--secondary disabled:cursor-not-allowed disabled:opacity-50"
                      @click="saveWriteup('draft')"
                    >
                      {{ submissionSaving === 'draft' ? '保存中...' : '保存草稿' }}
                    </button>
                    <button
                      type="button"
                      :disabled="
                        submissionLoading || submissionSaving !== null || !challenge?.is_solved
                      "
                      class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
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

        <aside
          v-if="activeWorkspaceTab === 'question'"
          class="detail-aside tool-pane"
        >
          <div class="tool-pane-inner">
            <section class="tool-group">
              <div>
                <div class="workspace-overline">
                  Primary Action
                </div>
                <h2 class="tool-title">
                  {{ submitPanelTitle }}
                </h2>
                <p class="tool-copy">
                  {{ submitPanelCopy }}
                </p>
              </div>
              <span
                v-if="challenge?.is_solved"
                class="writeup-status-pill writeup-status-pill--success"
              >
                已通过
              </span>
              <div class="flag-field">
                <label
                  for="challenge-flag-input"
                  class="flag-label"
                >
                  {{ submitFieldLabel }}
                </label>
                <div class="flag-row">
                  <div
                    class="ui-control-wrap flag-input-wrap"
                    :class="[submitInputClass, { 'is-disabled': submitting }]"
                  >
                    <input
                      id="challenge-flag-input"
                      v-model="flagInput"
                      type="text"
                      aria-label="Flag"
                      :placeholder="submitPlaceholder"
                      class="ui-control challenge-input flag-input disabled:cursor-not-allowed disabled:opacity-50"
                      :disabled="submitting"
                      @keyup.enter="submitFlagHandler"
                    >
                  </div>
                  <button
                    type="button"
                    :disabled="submitting"
                    class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
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
              :instance-sharing="challenge.instance_sharing"
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
import { useRoute, useRouter } from 'vue-router'

import {
  getChallengeDetail,
  getCommunityChallengeSolutions,
  getRecommendedChallengeSolutions,
} from '@/api/challenge'
import type {
  ChallengeDetailData,
  CommunityChallengeSolutionData,
  RecommendedChallengeSolutionData,
} from '@/api/contracts'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import ChallengeInstanceCard from '@/components/challenge/ChallengeInstanceCard.vue'
import { useChallengeDetailInteractions } from '@/composables/useChallengeDetailInteractions'
import {
  useChallengeDetailPresentation,
  type ChallengeSolutionTab,
} from '@/composables/useChallengeDetailPresentation'
import { useChallengeInstance } from '@/composables/useChallengeInstance'
import { useSanitize } from '@/composables/useSanitize'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'
import { useToast } from '@/composables/useToast'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

type WorkspaceTab = 'question' | 'solution' | 'records' | 'writeup'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { sanitizeHtml } = useSanitize()

const challengeId = computed(() => String(route.params.id ?? ''))
const challenge = ref<ChallengeDetailData | null>(null)
const loading = ref(false)
const recommendedSolutions = ref<RecommendedChallengeSolutionData[]>([])
const communitySolutions = ref<CommunityChallengeSolutionData[]>([])
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

const workspaceTabs: Array<{ id: WorkspaceTab; label: string }> = [
  { id: 'question', label: '题目' },
  { id: 'solution', label: '题解' },
  { id: 'records', label: '提交记录' },
  { id: 'writeup', label: '编写题解' },
]
const workspaceTabOrder = workspaceTabs.map((tab) => tab.id) as WorkspaceTab[]
const {
  activeTab: activeWorkspaceTab,
  setTabButtonRef,
  selectTab: selectWorkspaceTab,
  handleTabKeydown: handleWorkspaceTabKeydown,
} = useUrlSyncedTabs<WorkspaceTab>({
  orderedTabs: workspaceTabOrder,
  defaultTab: 'question',
})
const solutionTabOrder: ChallengeSolutionTab[] = ['recommended', 'community']
const submissionRecordPageSize = 10
const submissionRecordPage = ref(1)

const {
  myWriteup,
  submitting,
  submissionLoading,
  submissionSaving,
  writeupTitle,
  writeupContent,
  flagInput,
  submitResult,
  submissionRecords,
  resetChallengeInteractions,
  loadMyWriteupSubmission,
  loadSubmissionRecords,
  isHintExpanded,
  toggleHint,
  submitFlagHandler,
  downloadAttachment,
  saveWriteup,
} = useChallengeDetailInteractions({
  challengeId,
  challenge,
  loadSolutions,
})
const submissionRecordTotal = computed(() => submissionRecords.value.length)
const submissionRecordTotalPages = computed(() =>
  Math.max(1, Math.ceil(submissionRecordTotal.value / submissionRecordPageSize))
)
const paginatedSubmissionRecords = computed(() => {
  const start = (submissionRecordPage.value - 1) * submissionRecordPageSize
  return submissionRecords.value.slice(start, start + submissionRecordPageSize)
})

const needTarget = computed(() => challenge.value?.need_target ?? true)
const {
  activeSolutionTab,
  sanitizedDescription,
  displayedSolutionCards,
  activeSolution,
  sanitizedActiveSolutionContent,
  submitPlaceholder,
  submitPanelTitle,
  submitPanelCopy,
  submitFieldLabel,
  submitInputClass,
  clearSolutions,
  buildMetaPillStyle,
  submissionStatusLabel,
  submissionStatusText,
  submissionRecordMessage,
  visibilityStatusLabel,
  formatWriteupTime,
  formatSubmissionTime,
  getCategoryLabel,
  getCategoryColor,
  getDifficultyLabel,
  getDifficultyColor,
} = useChallengeDetailPresentation({
  challenge,
  recommendedSolutions,
  communitySolutions,
  myWriteup,
  selectedSolutionId,
  submitResult,
  sanitizeHtml,
})

function selectSolutionTab(tab: ChallengeSolutionTab): void {
  activeSolutionTab.value = tab
}

function changeSubmissionRecordPage(page: number): void {
  submissionRecordPage.value = page
}

const { setTabButtonRef: setSolutionTabButtonRef, handleTabKeydown: handleSolutionTabKeydown } =
  useTabKeyboardNavigation<ChallengeSolutionTab>({
    orderedTabs: solutionTabOrder,
    selectTab: selectSolutionTab,
  })

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
    toast.error('加载题目详情失败')
    void router.push('/challenges')
  } finally {
    loading.value = false
  }
}

watch(
  challengeId,
  () => {
    challenge.value = null
    submissionRecordPage.value = 1
    resetChallengeInteractions()
    clearSolutions()
    selectWorkspaceTab('question')
    void Promise.all([loadChallenge(), loadMyWriteupSubmission(), loadSubmissionRecords()])
  },
  { immediate: true }
)

watch(
  submissionRecords,
  () => {
    submissionRecordPage.value = 1
  },
  { deep: true }
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
  --bg-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --bg-shell: var(--journal-surface);
  --bg-panel: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  --bg-muted: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  --line-soft: var(--journal-border);
  --line-strong: color-mix(in srgb, var(--journal-border) 92%, var(--color-border-default));
  --text-main: var(--journal-ink);
  --text-subtle: var(--journal-muted);
  --text-faint: color-mix(in srgb, var(--journal-muted) 84%, var(--color-bg-base));
  --brand: var(--journal-accent);
  --brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --brand-soft-strong: color-mix(in srgb, var(--brand) 14%, transparent);
  --brand-ink: var(--journal-accent-strong);
  --success: var(--color-success);
  --warning: var(--color-warning);
  --danger: var(--color-danger);
  --shadow-shell: var(--journal-shell-hero-shadow, 0 22px 50px var(--color-shadow-soft));
  --radius-xl: 28px;
  --radius-lg: 18px;
  --font-sans: var(--font-family-sans);
  --font-mono: var(--font-family-mono);
  --journal-faint: var(--text-faint);
  --journal-accent-soft: var(--brand-soft);
  --journal-line-soft: var(--line-soft);
  --journal-line-strong: var(--line-strong);
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
  --page-top-tabs-gap: var(--space-7);
  --page-top-tabs-margin: var(--space-2-5) 0 0;
  --page-top-tabs-padding: 0 var(--space-7);
  --page-top-tabs-border: var(--line-soft);
  --page-top-tab-min-height: 52px;
  --page-top-tab-padding: var(--space-2-5) 0 var(--space-3-5);
  --page-top-tab-font-size: var(--font-size-15);
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

.top-note,
.section-hint,
.tool-copy,
.writeup-footnote {
  font-size: var(--font-size-13);
  line-height: 1.75;
  color: var(--text-faint);
}

.challenge-subtabs {
  --page-top-tabs-gap: var(--space-4-5);
  --page-top-tabs-margin: 0;
  --page-top-tabs-padding: 0 0 var(--space-2-5);
  --page-top-tabs-border: var(--line-soft);
  --page-top-tab-min-height: 3rem;
  --page-top-tab-padding: 0 0 var(--space-2);
  --page-top-tab-font-size: var(--font-size-14);
  --page-top-tab-font-weight: 600;
  --page-top-tab-color: var(--text-faint);
  --page-top-tab-active-color: var(--journal-accent-strong);
  --page-top-tab-active-border: var(--journal-accent);
  scrollbar-width: none;
}

.challenge-subtab {
  min-width: fit-content;
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
  padding: var(--space-7);
}

.tool-pane {
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: var(--space-7);
  border-left: 1px solid var(--line-soft);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--bg-panel) 92%, var(--color-bg-base)),
    color-mix(in srgb, var(--bg-shell) 88%, var(--color-bg-base))
  );
}

.tool-pane-inner {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 100%;
  position: sticky;
  top: var(--space-7);
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
  gap: var(--space-6);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--line-soft);
}

.question-hero-main {
  min-width: 0;
}

.question-title {
  margin: var(--space-3) 0 0;
  color: var(--text-main);
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-4);
}

.meta-strip--compact {
  margin-top: 0;
  margin-bottom: var(--space-4);
}

.meta-pill,
.writeup-status-pill,
.status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 var(--space-3-5);
  border: 1px solid var(--line-soft);
  border-radius: 999px;
  background: color-mix(in srgb, var(--bg-panel) 72%, transparent);
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--text-subtle);
}

.meta-pill--brand {
  border-color: color-mix(in srgb, var(--brand) 20%, transparent);
  background: var(--brand-soft);
  color: var(--brand-ink);
}

.score-rail {
  padding-left: var(--space-5-5);
  border-left: 1px solid var(--line-soft);
}

.score-label {
  font-size: var(--font-size-11);
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.score-value {
  margin-top: var(--space-2);
  color: var(--text-main);
  font: 700 34px/1 var(--font-mono);
}

.score-value small {
  font-size: var(--font-size-16);
  color: var(--text-faint);
}

.score-note {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--line-soft);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--text-subtle);
}

.section {
  padding-top: var(--space-6);
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
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.section-title:not(.workspace-tab-heading__title) {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-20);
  line-height: 1.2;
  color: var(--text-main);
}

.description,
.solution-preview,
.tool-copy {
  color: var(--journal-muted);
}

.description {
  font-size: var(--font-size-15);
  line-height: 1.92;
  color: var(--text-subtle);
}

.challenge-prose :deep(p),
.challenge-prose :deep(ul),
.challenge-prose :deep(ol) {
  margin-bottom: var(--space-4);
}

.challenge-prose :deep(pre) {
  overflow: auto;
  margin: var(--space-5) 0;
  padding: var(--space-4-5) var(--space-5);
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--bg-panel) 72%, var(--color-bg-base));
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
  gap: var(--space-3-5);
  padding: var(--space-4) 0;
  border-top: 1px dashed var(--line-soft);
}

.hint-line:first-of-type {
  padding-top: 0;
  border-top: 0;
}

.hint-label {
  font-size: var(--font-size-14);
  font-weight: 600;
  color: var(--text-main);
}

.hint-copy {
  margin-top: var(--space-2-5);
  font-size: var(--font-size-14);
  line-height: 1.8;
  color: var(--text-subtle);
}

.solution-layout {
  display: grid;
  grid-template-columns: minmax(240px, 0.54fr) minmax(0, 1fr);
  gap: var(--space-6);
}

.solution-nav {
  padding-right: var(--space-5);
  border-right: 1px solid var(--line-soft);
}

.solution-tab {
  min-width: fit-content;
}

.solution-item,
.solution-list-item {
  width: 100%;
  text-align: left;
  padding: var(--space-3-5) 0 var(--space-4) var(--space-3-5);
  border: 0;
  border-left: 2px solid transparent;
  border-bottom: 1px solid var(--line-soft);
  background: transparent;
}

.solution-item strong,
.solution-list-item strong {
  display: block;
  font-size: var(--font-size-14);
  color: var(--text-main);
}

.solution-item span,
.solution-list-item span {
  display: block;
  margin-top: var(--space-1-5);
  font-size: var(--font-size-12);
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
  font-size: var(--font-size-14);
  line-height: 1.9;
  color: var(--text-subtle);
}

.solution-preview__content {
  min-height: 15rem;
}

.solution-preview__content :deep(h1),
.solution-preview__content :deep(h2),
.solution-preview__content :deep(h3) {
  margin-top: var(--space-5);
}

.inline-note {
  padding-left: var(--space-4);
  border-left: 2px solid var(--line-soft);
  font-size: var(--font-size-0-90);
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
  gap: var(--space-4-5);
  align-items: center;
  padding: var(--space-4-5) 0;
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
  font-size: var(--font-size-13);
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
  gap: var(--space-4);
}

.field label,
.flag-label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-12);
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.field .ui-control-wrap,
.flag-input-wrap {
  border-color: var(--line-strong);
  background: var(--bg-panel);
}

.field .ui-control-wrap:not(.writeup-textarea-wrap),
.flag-input-wrap {
  --ui-control-height: 3.125rem;
}

.field .ui-control-wrap > input,
.writeup-textarea-wrap > textarea {
  font: 500 14px/1.6 var(--font-sans);
}

.writeup-textarea-wrap > textarea {
  min-height: 260px;
  resize: vertical;
}

.flag-input-wrap > input {
  font: 500 15px/1 var(--font-mono);
}

.writeup-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.writeup-actions {
  display: flex;
  gap: var(--space-3);
  justify-content: flex-end;
}

.flag-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-3);
}

.tool-group + .tool-group,
:deep(.instance-shell) {
  margin-top: var(--space-6);
}

.tool-group + .tool-group {
  margin-top: var(--space-6);
  padding-top: var(--space-6);
  border-top: 1px solid var(--line-soft);
}

.tool-title {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-18);
  line-height: 1.25;
  color: var(--text-main);
}

.tool-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--text-subtle);
}

.hint-toggle {
  --ui-btn-height: 2.5rem;
}

.status-inline {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-3);
  font-size: var(--font-size-14);
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

.solution-list-item:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--brand) 44%, var(--color-bg-base));
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
    padding-top: var(--space-4-5);
    border-left: 0;
    border-top: 1px solid var(--line-soft);
  }
}

@media (max-width: 760px) {
  .detail-content > .workspace-tabbar,
  .content-pane,
  .tool-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }

  .detail-content > .workspace-tabbar {
    gap: var(--space-5-5);
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
    gap: var(--space-2-5);
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

:global([data-theme='dark']) .workspace-shell {
  --workspace-shell-radial-strength: 14%;
  --workspace-shell-radial-size: 24rem;
  --workspace-shell-top-strength: 97%;
}

</style>
