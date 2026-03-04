<template>
  <div class="space-y-6">
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else-if="challenge" class="space-y-6">
      <div class="flex items-start justify-between">
        <div class="space-y-3">
          <h1 class="text-3xl font-bold text-[#c9d1d9]">{{ challenge.title }}</h1>
          <div class="flex flex-wrap gap-2">
            <span
              class="rounded px-3 py-1 text-sm font-medium"
              :style="{ backgroundColor: getCategoryColor(challenge.category) + '20', color: getCategoryColor(challenge.category) }"
            >
              {{ getCategoryLabel(challenge.category) }}
            </span>
            <span
              class="rounded px-3 py-1 text-sm font-medium"
              :style="{ backgroundColor: getDifficultyColor(challenge.difficulty) + '20', color: getDifficultyColor(challenge.difficulty) }"
            >
              {{ getDifficultyLabel(challenge.difficulty) }}
            </span>
            <span v-for="tag in challenge.tags" :key="tag" class="rounded bg-[#21262d] px-3 py-1 text-sm text-[#8b949e]">
              {{ tag }}
            </span>
          </div>
          <div class="text-sm text-[#8b949e]">{{ challenge.solved_count }} 人解出</div>
        </div>
        <div class="text-right">
          <div class="font-mono text-3xl font-bold text-[#0891b2]">{{ challenge.points }}pts</div>
          <span v-if="challenge.is_solved" class="mt-2 inline-block rounded bg-green-500/20 px-3 py-1 text-sm font-medium text-green-500">
            已完成 ✓
          </span>
        </div>
      </div>

      <div class="rounded-lg border border-[#30363d] bg-[#161b22] p-6">
        <h2 class="mb-4 text-lg font-semibold text-[#c9d1d9]">挑战描述</h2>
        <div v-html="sanitizedDescription" class="prose prose-invert max-w-none text-[#8b949e]"></div>
        <button
          v-if="challenge.attachment_url"
          class="mt-4 rounded-lg bg-[#21262d] px-4 py-2 text-sm text-[#c9d1d9] transition-colors hover:bg-[#30363d]"
          @click="downloadAttachment"
        >
          下载附件
        </button>
      </div>

      <div class="rounded-lg border border-[#30363d] bg-[#161b22] p-6">
        <h2 class="mb-4 text-lg font-semibold text-[#c9d1d9]">Flag 提交</h2>
        <div class="space-y-4">
          <div class="flex gap-2">
            <input
              v-model="flagInput"
              type="text"
              placeholder="flag{...}"
              :disabled="challenge.is_solved"
              class="flex-1 rounded-lg border bg-[#0d1117] px-4 py-2 font-mono text-[#c9d1d9] placeholder-[#6e7681] transition-colors focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
              :class="submitResult?.success ? 'border-green-500' : 'border-[#0891b2]'"
              @keyup.enter="submitFlagHandler"
            />
            <button
              :disabled="challenge.is_solved || submitting"
              class="rounded-lg bg-[#0891b2] px-6 py-2 text-sm font-medium text-white transition-colors hover:bg-[#0891b2]/90 disabled:cursor-not-allowed disabled:opacity-50"
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

      <div v-if="!challenge.is_solved" class="rounded-lg border border-[#30363d] bg-[#161b22] p-6">
        <h2 class="mb-4 text-lg font-semibold text-[#c9d1d9]">靶机实例</h2>
        <div class="space-y-4">
          <button
            :disabled="creating"
            class="rounded-lg bg-[#0891b2] px-6 py-2.5 text-sm font-medium text-white transition-colors hover:bg-[#0891b2]/90 disabled:cursor-not-allowed disabled:opacity-50"
            @click="startChallenge"
          >
            {{ creating ? '正在创建实例...' : '启动靶机' }}
          </button>
          <div class="text-sm text-[#6e7681]">点击按钮创建专属实例，实例有效期为 2 小时</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getChallengeDetail, submitFlag, createInstance } from '@/api/challenge'
import { useSanitize } from '@/composables/useSanitize'
import { useToast } from '@/composables/useToast'
import type { ChallengeCategory, ChallengeDifficulty, ChallengeDetailData } from '@/api/contracts'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { sanitizeHtml } = useSanitize()

const challenge = ref<ChallengeDetailData | null>(null)
const loading = ref(false)
const creating = ref(false)
const submitting = ref(false)
const flagInput = ref('')
const submitResult = ref<{ success: boolean; message: string } | null>(null)

const sanitizedDescription = computed(() => {
  return challenge.value ? sanitizeHtml(challenge.value.description) : ''
})

async function loadChallenge() {
  const id = route.params.id as string
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

async function startChallenge() {
  if (!challenge.value) return
  creating.value = true
  try {
    const instance = await createInstance(challenge.value.id)
    toast.success('实例创建成功')
    router.push('/instances')
  } catch (error) {
    toast.error('创建实例失败')
  } finally {
    creating.value = false
  }
}

async function submitFlagHandler() {
  if (!challenge.value || !flagInput.value.trim()) return
  submitting.value = true
  submitResult.value = null
  try {
    const result = await submitFlag(challenge.value.id, flagInput.value.trim())
    if (result.correct) {
      submitResult.value = { success: true, message: '恭喜！Flag 正确！' }
      toast.success('Flag 正确！')
      challenge.value.is_solved = true
    } else {
      submitResult.value = { success: false, message: 'Flag 错误，请重试' }
    }
  } catch (error) {
    submitResult.value = { success: false, message: '提交失败，请重试' }
  } finally {
    submitting.value = false
  }
}

function downloadAttachment() {
  if (challenge.value?.attachment_url) {
    window.open(challenge.value.attachment_url, '_blank')
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
    hell: '地狱',
  }
  return labels[difficulty]
}

function getDifficultyColor(difficulty: ChallengeDifficulty): string {
  const colors: Record<ChallengeDifficulty, string> = {
    beginner: '#10b981',
    easy: '#3b82f6',
    medium: '#f59e0b',
    hard: '#ef4444',
    hell: '#7c3aed',
  }
  return colors[difficulty]
}

onMounted(() => {
  loadChallenge()
})
</script>
