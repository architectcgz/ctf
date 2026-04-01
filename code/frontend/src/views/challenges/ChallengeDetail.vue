<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-col space-y-6 rounded-[30px] border p-6 md:p-8"
  >
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"></div>
    </div>

    <div v-else-if="challenge" class="space-y-6">
      <div class="challenge-panel p-6 md:p-8">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-3">
            <div class="journal-eyebrow">Challenge Detail</div>
            <h1 class="text-3xl font-bold text-[var(--journal-ink)]">{{ challenge.title }}</h1>
            <div class="flex flex-wrap gap-2">
              <span
                class="rounded-full px-3 py-1 text-sm font-medium"
                :style="{ backgroundColor: getCategoryColor(challenge.category) + '22', color: getCategoryColor(challenge.category) }"
              >
                {{ getCategoryLabel(challenge.category) }}
              </span>
              <span
                class="rounded-full px-3 py-1 text-sm font-medium"
                :style="{ backgroundColor: getDifficultyColor(challenge.difficulty) + '22', color: getDifficultyColor(challenge.difficulty) }"
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
            <div class="text-[11px] uppercase tracking-[0.22em] text-[var(--journal-muted)]">Score</div>
            <div class="mt-1 font-mono text-2xl font-bold text-[var(--journal-ink)]">{{ challenge.points }}<span class="ml-1 text-lg text-[var(--journal-muted)]">pts</span></div>
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
              <h2 class="text-lg font-semibold text-[var(--journal-ink)]">挑战描述</h2>
              <button
                :disabled="writeupLoading"
                class="challenge-btn-outline"
                @click="toggleWriteup"
              >
                {{ writeupLoading ? '加载题解中...' : writeupVisible ? '收起题解' : '查看题解' }}
              </button>
            </div>
            <div v-html="sanitizedDescription" class="prose challenge-prose max-w-none"></div>
            <button
              v-if="challenge.attachment_url"
              class="challenge-btn-outline mt-4"
              @click="downloadAttachment"
            >
              下载附件
            </button>
          </div>

          <div class="challenge-panel writeup-workbench p-6">
            <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
              <div class="max-w-2xl space-y-3">
                <div class="journal-eyebrow">Writeup Studio</div>
                <div>
                  <h2 class="text-lg font-semibold text-[var(--journal-ink)]">解题过程复盘</h2>
                  <p class="mt-2 text-sm leading-7 text-[var(--journal-muted)]">
                    把思路、关键利用步骤和踩坑点整理成一份自己的 writeup。草稿可反复保存，正式提交后教师会在分析页看到你的状态与评语。
                  </p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <span class="writeup-status-pill writeup-status-pill--primary">
                    {{ submissionStatusLabel(myWriteup?.submission_status) }}
                  </span>
                  <span
                    v-if="myWriteup"
                    class="writeup-status-pill"
                    :class="reviewStatusClass(myWriteup.review_status)"
                  >
                    {{ reviewStatusLabel(myWriteup.review_status) }}
                  </span>
                  <span
                    v-if="myWriteup?.submitted_at"
                    class="writeup-status-pill writeup-status-pill--muted"
                  >
                    提交于 {{ formatWriteupTime(myWriteup.submitted_at) }}
                  </span>
                </div>
              </div>

              <div class="writeup-side-note">
                <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-[var(--journal-accent)]">
                  Review Focus
                </div>
                <div class="mt-3 text-sm leading-6 text-[var(--journal-ink)]">
                  重点说明你如何定位突破口、怎样验证利用链、哪里走过弯路。
                </div>
              </div>
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
                />
              </label>

              <div class="writeup-meta-grid">
                <div class="writeup-meta-card">
                  <div class="writeup-meta-label">当前状态</div>
                  <div class="writeup-meta-value">{{ submissionStatusLabel(myWriteup?.submission_status) }}</div>
                </div>
                <div class="writeup-meta-card">
                  <div class="writeup-meta-label">评阅结果</div>
                  <div class="writeup-meta-value">{{ reviewStatusLabel(myWriteup?.review_status) }}</div>
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

            <div
              v-if="myWriteup?.review_comment"
              class="writeup-feedback-panel mt-5"
            >
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div class="text-xs font-semibold uppercase tracking-[0.22em] text-[var(--journal-accent)]">
                    Teacher Feedback
                  </div>
                  <div class="mt-2 text-sm text-[var(--journal-muted)]">
                    {{ reviewStatusLabel(myWriteup.review_status) }}
                  </div>
                </div>
                <div
                  v-if="myWriteup.reviewed_at"
                  class="text-xs text-[var(--journal-muted)]"
                >
                  {{ formatWriteupTime(myWriteup.reviewed_at) }}
                </div>
              </div>
              <p class="mt-4 whitespace-pre-wrap text-sm leading-7 text-[var(--journal-ink)]">
                {{ myWriteup.review_comment }}
              </p>
            </div>

            <div class="mt-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="text-sm text-[var(--journal-muted)]">
                {{
                  submissionLoading
                    ? '正在同步你的 writeup...'
                    : myWriteup?.updated_at
                      ? `最近更新：${formatWriteupTime(myWriteup.updated_at)}`
                      : '还没有提交记录，可以先保存草稿。'
                }}
              </div>
              <div class="flex flex-wrap gap-3">
                <button
                  :disabled="submissionLoading || submissionSaving !== null"
                  class="challenge-btn-outline"
                  @click="saveWriteup('draft')"
                >
                  {{ submissionSaving === 'draft' ? '保存中...' : '保存草稿' }}
                </button>
                <button
                  :disabled="submissionLoading || submissionSaving !== null"
                  class="challenge-btn-primary rounded-xl px-5 py-3 text-sm font-medium text-white transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                  @click="saveWriteup('submitted')"
                >
                  {{ submissionSaving === 'submitted' ? '提交中...' : '正式提交' }}
                </button>
              </div>
            </div>
          </div>

          <div class="challenge-panel p-6">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-[var(--journal-ink)]">Flag 提交</h2>
                <div class="mt-1 text-sm text-[var(--journal-muted)]">保持在当前题目页即可提交答案。</div>
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
                  placeholder="flag{...}"
                  :disabled="challenge.is_solved"
                  class="challenge-input flex-1 rounded-xl border px-4 py-3 font-mono transition-colors focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
                  :class="submitResult?.success ? 'border-[var(--color-success)]' : 'border-[#0891b2]'"
                  @keyup.enter="submitFlagHandler"
                />
                <button
                  :disabled="challenge.is_solved || submitting"
                  class="challenge-btn-primary rounded-xl px-6 py-3 text-sm font-medium text-white transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                  @click="submitFlagHandler"
                >
                  {{ submitting ? '提交中...' : '提交' }}
                </button>
              </div>
              <div v-if="submitResult" :class="submitResult.success ? 'text-[var(--color-success)]' : 'text-[var(--color-danger)]'" class="text-sm">
                {{ submitResult.message }}
              </div>
            </div>
          </div>

          <div
            v-if="challenge.hints.length > 0"
            class="challenge-panel p-6"
          >
            <h2 class="mb-4 text-lg font-semibold text-[var(--journal-ink)]">提示系统</h2>
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
                    <div v-if="hint.cost_points" class="mt-1 text-xs text-[var(--journal-muted)]">
                      解锁消耗：{{ hint.cost_points }} 分
                    </div>
                  </div>
                  <button
                    v-if="!hint.is_unlocked"
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
                <div v-if="hint.is_unlocked" class="mt-3 text-sm leading-6 text-[var(--journal-muted)]">
                  {{ hint.content }}
                </div>
                <div v-else class="mt-3 text-sm text-[var(--journal-muted)]">
                  解锁后显示提示内容
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="writeupVisible && writeup"
            class="challenge-panel p-6"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-[var(--journal-ink)]">题解</h2>
                <div class="mt-2 text-sm text-[var(--journal-muted)]">{{ writeup.title }}</div>
              </div>
              <div
                class="rounded bg-[var(--journal-surface-subtle)] px-3 py-1 text-xs uppercase tracking-[0.18em] text-[var(--journal-muted)]"
              >
                {{ writeup.visibility }}
              </div>
            </div>
            <div
              v-if="writeup.requires_spoiler_warning"
              class="mt-4 rounded-lg border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 px-4 py-3 text-sm text-[var(--color-warning)]"
            >
              你尚未完成该题。以下内容可能直接暴露解题思路，请谨慎阅读。
            </div>
            <div
              v-html="sanitizedWriteup"
              class="prose challenge-prose mt-4 max-w-none"
            ></div>
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
  getMyChallengeWriteupSubmission,
  getChallengeWriteup,
  upsertChallengeWriteupSubmission,
  submitFlag,
  unlockHint,
} from '@/api/challenge'
import ChallengeInstanceCard from '@/components/challenge/ChallengeInstanceCard.vue'
import { useChallengeInstance } from '@/composables/useChallengeInstance'
import { useSanitize } from '@/composables/useSanitize'
import { useToast } from '@/composables/useToast'
import type {
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeDetailData,
  ChallengeWriteupData,
  SubmissionWriteupData,
  SubmissionWriteupReviewStatus,
  SubmissionWriteupStatus,
} from '@/api/contracts'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { sanitizeHtml } = useSanitize()

const challengeId = computed(() => String(route.params.id ?? ''))
const challenge = ref<ChallengeDetailData | null>(null)
const loading = ref(false)
const submitting = ref(false)
const writeupVisible = ref(false)
const writeupLoading = ref(false)
const writeup = ref<ChallengeWriteupData | null>(null)
const myWriteup = ref<SubmissionWriteupData | null>(null)
const submissionLoading = ref(false)
const submissionSaving = ref<SubmissionWriteupStatus | null>(null)
const writeupTitle = ref('')
const writeupContent = ref('')
const flagInput = ref('')
const submitResult = ref<{ success: boolean; message: string } | null>(null)
const unlockingLevel = ref<number | null>(null)
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

const sanitizedDescription = computed(() => {
  if (!challenge.value) return ''
  const html = marked.parse(challenge.value.description, {
    gfm: true,
    breaks: true,
  })
  return sanitizeHtml(typeof html === 'string' ? html : challenge.value.description)
})

const sanitizedWriteup = computed(() => {
  return writeup.value ? sanitizeHtml(writeup.value.content) : ''
})

const needTarget = computed(() => challenge.value?.need_target ?? true)

function hydrateSubmissionForm(item: SubmissionWriteupData | null): void {
  writeupTitle.value = item?.title ?? ''
  writeupContent.value = item?.content ?? ''
}

async function loadChallenge() {
  const id = challengeId.value
  loading.value = true
  try {
    challenge.value = await getChallengeDetail(id)
  } catch (error) {
    toast.error('加载挑战详情失败')
    router.push('/challenges')
  } finally {
    loading.value = false
  }
}

async function loadMyWriteupSubmission() {
  if (!challengeId.value) return
  submissionLoading.value = true
  try {
    myWriteup.value = await getMyChallengeWriteupSubmission(challengeId.value)
    hydrateSubmissionForm(myWriteup.value)
  } catch {
    toast.error('加载个人 writeup 失败')
  } finally {
    submissionLoading.value = false
  }
}

async function toggleWriteup() {
  if (!challenge.value) return
  if (writeupVisible.value) {
    writeupVisible.value = false
    return
  }
  if (writeup.value) {
    writeupVisible.value = true
    return
  }

  writeupLoading.value = true
  try {
    const result = await getChallengeWriteup(challenge.value.id)
    if (!result) {
      toast.info('当前题目暂未开放题解')
      return
    }
    writeup.value = result
    writeupVisible.value = true
  } catch {
    toast.error('加载题解失败')
  } finally {
    writeupLoading.value = false
  }
}

async function submitFlagHandler() {
  if (!challenge.value || !flagInput.value.trim()) return
  submitting.value = true
  submitResult.value = null
  try {
    const result = await submitFlag(challenge.value.id, flagInput.value.trim())
    if (result.is_correct) {
      submitResult.value = { success: true, message: result.message }
      toast.success('Flag 正确！')
      challenge.value.is_solved = true
    } else {
      submitResult.value = { success: false, message: result.message }
    }
  } catch (error) {
    submitResult.value = { success: false, message: '提交失败，请重试' }
  } finally {
    submitting.value = false
  }
}

async function unlockHintHandler(level: number) {
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

async function downloadAttachment() {
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

async function saveWriteup(status: SubmissionWriteupStatus) {
  if (!challenge.value) return
  if (!writeupTitle.value.trim() || !writeupContent.value.trim()) {
    toast.error('请先补全 writeup 标题和正文')
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
    if (saved.submission_status === 'submitted' && status === 'draft') {
      toast.success('writeup 已更新')
    } else {
      toast.success(status === 'submitted' ? 'writeup 已正式提交' : '草稿已保存')
    }
  } catch {
    toast.error(status === 'submitted' ? '提交 writeup 失败' : '保存草稿失败')
  } finally {
    submissionSaving.value = null
  }
}

function submissionStatusLabel(status?: SubmissionWriteupStatus): string {
  if (status === 'submitted') return '已提交'
  if (status === 'draft') return '草稿'
  return '未开始'
}

function reviewStatusLabel(status?: SubmissionWriteupReviewStatus): string {
  switch (status) {
    case 'reviewed':
      return '已评阅'
    case 'excellent':
      return '优秀'
    case 'needs_revision':
      return '待修改'
    case 'pending':
      return '待评阅'
    default:
      return '未评阅'
  }
}

function reviewStatusClass(status?: SubmissionWriteupReviewStatus): string {
  switch (status) {
    case 'excellent':
      return 'writeup-status-pill--success'
    case 'needs_revision':
      return 'writeup-status-pill--warning'
    case 'reviewed':
      return 'writeup-status-pill--primary'
    default:
      return 'writeup-status-pill--muted'
  }
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
  challengeId,
  () => {
    writeupVisible.value = false
    writeup.value = null
    myWriteup.value = null
    writeupTitle.value = ''
    writeupContent.value = ''
    flagInput.value = ''
    submitResult.value = null
    void Promise.all([loadChallenge(), loadMyWriteupSubmission()])
  },
  { immediate: true }
)
</script>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: #ffffff;
  --journal-surface-subtle: rgba(248, 250, 252, 0.92);
}

.journal-hero {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.98), rgba(241, 245, 249, 0.95));
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
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.56);
}

.challenge-panel {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.94));
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
  background: rgba(248, 250, 252, 0.92);
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
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(243, 244, 255, 0.96));
}

.writeup-side-note {
  max-width: 18rem;
  border: 1px solid rgba(99, 102, 241, 0.14);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.72);
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
  border: 1px solid rgba(148, 163, 184, 0.18);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.75);
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
  background: rgba(148, 163, 184, 0.16);
  color: #475569;
}

.writeup-feedback-panel {
  border: 1px solid rgba(79, 70, 229, 0.16);
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.74);
  padding: 1.1rem 1.2rem;
}

.challenge-prose :deep(h1),
.challenge-prose :deep(h2),
.challenge-prose :deep(h3),
.challenge-prose :deep(strong),
.challenge-prose :deep(code) {
  color: var(--journal-ink);
}

.hint-list {
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(255, 255, 255, 0.42);
}

.hint-item {
  padding: 1rem 1.1rem;
}

.hint-item + .hint-item {
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f1f5f9;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.85);
  --journal-surface-subtle: rgba(30, 41, 59, 0.6);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .challenge-score-card,
:global([data-theme='dark']) .challenge-panel,
:global([data-theme='dark']) .hint-list,
:global([data-theme='dark']) .challenge-btn-outline,
:global([data-theme='dark']) .challenge-input {
  background: rgba(15, 23, 42, 0.42);
}
</style>
