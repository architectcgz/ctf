<template>
  <div class="mx-auto max-w-7xl space-y-6">
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else-if="challenge" class="space-y-6">
      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(8,47,73,0.88),rgba(15,23,42,0.96))] p-6 shadow-[0_24px_60px_rgba(15,23,42,0.22)]">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-3">
            <h1 class="text-3xl font-bold text-white">{{ challenge.title }}</h1>
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
                class="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-sm text-slate-200"
              >
                {{ tag }}
              </span>
            </div>
          </div>
          <div class="rounded-xl border border-cyan-200/12 bg-slate-950/45 px-4 py-3 text-left lg:min-w-[148px] lg:text-right">
            <div class="text-[11px] uppercase tracking-[0.22em] text-slate-200">Score</div>
            <div class="mt-1 font-mono text-2xl font-bold text-white">{{ challenge.points }}<span class="ml-1 text-lg text-cyan-200">pts</span></div>
            <span
              v-if="challenge.is_solved"
              class="mt-3 inline-flex rounded-full bg-emerald-500/18 px-3 py-1 text-sm font-medium text-emerald-200"
            >
              已完成 ✓
            </span>
          </div>
        </div>
      </div>

      <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
        <main class="space-y-6">
          <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
            <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
              <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">挑战描述</h2>
              <button
                :disabled="writeupLoading"
                class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[#21262d] disabled:cursor-not-allowed disabled:opacity-50"
                @click="toggleWriteup"
              >
                {{ writeupLoading ? '加载题解中...' : writeupVisible ? '收起题解' : '查看题解' }}
              </button>
            </div>
            <div v-html="sanitizedDescription" class="prose prose-invert max-w-none text-[var(--color-text-secondary)]"></div>
            <button
              v-if="challenge.attachment_url"
              class="mt-4 rounded-lg bg-[#21262d] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[#30363d]"
              @click="downloadAttachment"
            >
              下载附件
            </button>
          </div>

          <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">Flag 提交</h2>
                <div class="mt-1 text-sm text-[var(--color-text-secondary)]">保持在当前题目页即可提交答案，不需要离开题面。</div>
              </div>
              <span
                v-if="challenge.is_solved"
                class="rounded-full bg-emerald-500/18 px-3 py-1 text-xs font-medium text-emerald-300"
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
                  class="flex-1 rounded-xl border bg-[var(--color-bg-base)] px-4 py-3 font-mono text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] transition-colors focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
                  :class="submitResult?.success ? 'border-green-500' : 'border-[#0891b2]'"
                  @keyup.enter="submitFlagHandler"
                />
                <button
                  :disabled="challenge.is_solved || submitting"
                  class="rounded-xl bg-[var(--color-primary)] px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-[var(--color-primary)]/90 disabled:cursor-not-allowed disabled:opacity-50"
                  @click="submitFlagHandler"
                >
                  {{ submitting ? '提交中...' : '提交' }}
                </button>
              </div>
              <div v-if="submitResult" :class="submitResult.success ? 'text-green-500' : 'text-red-500'" class="text-sm">
                {{ submitResult.message }}
              </div>
            </div>
          </div>

          <div
            v-if="challenge.hints.length > 0"
            class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
          >
            <h2 class="mb-4 text-lg font-semibold text-[var(--color-text-primary)]">提示系统</h2>
            <div class="space-y-4">
              <div
                v-for="hint in challenge.hints"
                :key="hint.id"
                class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-base)] p-4"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <div class="text-sm font-medium text-[var(--color-text-primary)]">
                      Level {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}
                    </div>
                    <div v-if="hint.cost_points" class="mt-1 text-xs text-[var(--color-text-secondary)]">
                      解锁消耗：{{ hint.cost_points }} 分
                    </div>
                  </div>
                  <button
                    v-if="!hint.is_unlocked"
                    :disabled="unlockingLevel === hint.level"
                    class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-xs font-medium text-white transition-colors hover:bg-[var(--color-primary)]/90 disabled:cursor-not-allowed disabled:opacity-50"
                    @click="unlockHintHandler(hint.level)"
                  >
                    {{ unlockingLevel === hint.level ? '解锁中...' : '解锁提示' }}
                  </button>
                  <span
                    v-else
                    class="rounded bg-emerald-500/20 px-3 py-1 text-xs font-medium text-emerald-500"
                  >
                    已解锁
                  </span>
                </div>
                <div v-if="hint.is_unlocked" class="mt-3 text-sm leading-6 text-[var(--color-text-secondary)]">
                  {{ hint.content }}
                </div>
                <div v-else class="mt-3 text-sm text-[var(--color-text-muted)]">
                  解锁后显示提示内容
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="writeupVisible && writeup"
            class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">题解</h2>
                <div class="mt-2 text-sm text-[var(--color-text-secondary)]">{{ writeup.title }}</div>
              </div>
              <div
                class="rounded bg-[var(--color-bg-base)] px-3 py-1 text-xs uppercase tracking-[0.18em] text-[var(--color-text-muted)]"
              >
                {{ writeup.visibility }}
              </div>
            </div>
            <div
              v-if="writeup.requires_spoiler_warning"
              class="mt-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-300"
            >
              你尚未完成该题。以下内容可能直接暴露解题思路，请谨慎阅读。
            </div>
            <div
              v-html="sanitizedWriteup"
              class="prose prose-invert mt-4 max-w-none text-[var(--color-text-secondary)]"
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
            class="rounded-2xl border border-emerald-400/30 bg-emerald-500/10 p-5 text-sm text-emerald-200"
          >
            该题目不需要靶机，可直接分析题面并提交 Flag。
          </section>
        </aside>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { marked } from 'marked'
import { useRoute, useRouter } from 'vue-router'
import {
  downloadAttachment as downloadChallengeAttachment,
  getChallengeDetail,
  getChallengeWriteup,
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
    flagInput.value = ''
    submitResult.value = null
    void loadChallenge()
  },
  { immediate: true }
)
</script>
