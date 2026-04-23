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
          <ChallengeQuestionPanel
            v-if="activeWorkspaceTab === 'question'"
            :challenge="challenge"
            :sanitized-description="sanitizedDescription"
            :score-rail-probe-message="scoreRailProbeMessage"
            :build-meta-pill-style="buildMetaPillStyle"
            :get-category-label="getCategoryLabel"
            :get-category-color="getCategoryColor"
            :get-difficulty-label="getDifficultyLabel"
            :get-difficulty-color="getDifficultyColor"
            :is-hint-expanded="isHintExpanded"
            @download-attachment="downloadAttachment"
            @toggle-hint="toggleHint"
            @score-rail-probe="handleScoreRailProbe"
          />

          <ChallengeSolutionsPanel
            v-else-if="activeWorkspaceTab === 'solution'"
            :challenge-solved="Boolean(challenge?.is_solved)"
            :recommended-solution-count="recommendedSolutions.length"
            :community-solution-count="communitySolutions.length"
            :active-solution-tab="activeSolutionTab"
            :displayed-solution-cards="displayedSolutionCards"
            :active-solution="activeSolution"
            :sanitized-active-solution-content="sanitizedActiveSolutionContent"
            :format-writeup-time="formatWriteupTime"
            :set-solution-tab-button-ref="setSolutionTabButtonRef"
            :handle-solution-tab-keydown="handleSolutionTabKeydown"
            @select-tab="selectSolutionTab"
            @select-solution="selectedSolutionId = $event"
          />

          <ChallengeSubmissionRecordsPanel
            v-else-if="activeWorkspaceTab === 'records'"
            :submission-records="submissionRecords"
            :paginated-submission-records="paginatedSubmissionRecords"
            :submission-record-page="submissionRecordPage"
            :submission-record-total="submissionRecordTotal"
            :submission-record-total-pages="submissionRecordTotalPages"
            :format-submission-time="formatSubmissionTime"
            :submission-record-message="submissionRecordMessage"
            :submission-status-text="submissionStatusText"
            @change-page="changeSubmissionRecordPage"
          />

          <ChallengeWriteupPanel
            v-else
            :challenge-solved="Boolean(challenge?.is_solved)"
            :my-writeup="myWriteup"
            :submission-loading="submissionLoading"
            :submission-saving="submissionSaving"
            :writeup-title="writeupTitle"
            :writeup-content="writeupContent"
            :format-writeup-time="formatWriteupTime"
            :submission-status-label="submissionStatusLabel"
            @update:writeup-title="writeupTitle = $event"
            @update:writeup-content="writeupContent = $event"
            @save="saveWriteup"
          />
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
import { computed, onBeforeUnmount, ref, watch } from 'vue'
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
import ChallengeInstanceCard from '@/components/challenge/ChallengeInstanceCard.vue'
import ChallengeQuestionPanel from '@/components/challenge/ChallengeQuestionPanel.vue'
import ChallengeSolutionsPanel from '@/components/challenge/ChallengeSolutionsPanel.vue'
import ChallengeSubmissionRecordsPanel from '@/components/challenge/ChallengeSubmissionRecordsPanel.vue'
import ChallengeWriteupPanel from '@/components/challenge/ChallengeWriteupPanel.vue'
import { useChallengeDetailInteractions } from '@/composables/useChallengeDetailInteractions'
import {
  useChallengeDetailPresentation,
  type ChallengeSolutionTab,
} from '@/composables/useChallengeDetailPresentation'
import { useChallengeInstance } from '@/composables/useChallengeInstance'
import { useProbeEasterEggs } from '@/composables/useProbeEasterEggs'
import { useSanitize } from '@/composables/useSanitize'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'
import { useToast } from '@/composables/useToast'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

type WorkspaceTab = 'question' | 'solution' | 'records' | 'writeup'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { sanitizeHtml } = useSanitize()
const { track } = useProbeEasterEggs()

const challengeId = computed(() => String(route.params.id ?? ''))
const challenge = ref<ChallengeDetailData | null>(null)
const loading = ref(false)
const scoreRailProbeMessage = ref('')
const recommendedSolutions = ref<RecommendedChallengeSolutionData[]>([])
const communitySolutions = ref<CommunityChallengeSolutionData[]>([])
const selectedSolutionId = ref<string | null>(null)
let scoreRailProbeTimer: number | null = null
let latestChallengeRequestId = 0
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

function showScoreRailProbe(message: string) {
  scoreRailProbeMessage.value = message
  if (scoreRailProbeTimer) {
    window.clearTimeout(scoreRailProbeTimer)
  }
  scoreRailProbeTimer = window.setTimeout(() => {
    scoreRailProbeMessage.value = ''
    scoreRailProbeTimer = null
  }, 2800)
}

function handleScoreRailProbe() {
  const result = track('challenge-side-rail', 4)
  if (!result.unlocked) {
    return
  }
  showScoreRailProbe('这块区域的情报价值，低于你现在的期待。')
}

onBeforeUnmount(() => {
  if (scoreRailProbeTimer) {
    window.clearTimeout(scoreRailProbeTimer)
  }
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

async function loadSolutions(id: string, requestId = latestChallengeRequestId): Promise<void> {
  try {
    const [recommended, communityPage] = await Promise.all([
      getRecommendedChallengeSolutions(id),
      getCommunityChallengeSolutions(id),
    ])
    if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
      return
    }
    recommendedSolutions.value = recommended
    communitySolutions.value = communityPage.list
  } catch {
    if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
      return
    }
    clearSolutions()
    toast.error('加载题解失败')
  }
}

async function loadChallenge(): Promise<void> {
  const id = challengeId.value
  const requestId = ++latestChallengeRequestId
  loading.value = true

  try {
    const detail = await getChallengeDetail(id)
    if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
      return
    }
    challenge.value = detail

    if (detail.is_solved) {
      await loadSolutions(id, requestId)
    } else {
      clearSolutions()
    }
  } catch {
    if (requestId !== latestChallengeRequestId || id !== challengeId.value) {
      return
    }
    toast.error('加载题目详情失败')
    void router.push('/challenges')
  } finally {
    if (requestId === latestChallengeRequestId && id === challengeId.value) {
      loading.value = false
    }
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
.tool-copy {
  font-size: var(--font-size-13);
  line-height: 1.75;
  color: var(--text-faint);
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


















.section-title:not(.workspace-tab-heading__title) {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-20);
  line-height: 1.2;
  color: var(--text-main);
}

.tool-copy {
  color: var(--journal-muted);
}



































.flag-label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-12);
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.flag-input-wrap {
  border-color: var(--line-strong);
  background: var(--bg-panel);
  --ui-control-height: 3.125rem;
}

.flag-input-wrap > input {
  font: 500 15px/1 var(--font-mono);
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

.writeup-status-pill {
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

.writeup-status-pill--success {
  border-color: color-mix(in srgb, var(--color-success) 18%, transparent);
  background: var(--journal-success-soft);
  color: var(--journal-success-ink);
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

  .flag-row {
    grid-template-columns: minmax(0, 1fr);
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
