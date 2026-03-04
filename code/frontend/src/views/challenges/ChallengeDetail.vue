<template>
  <div v-loading="loading" class="space-y-6">
    <div v-if="challenge" class="space-y-6">
      <div class="flex items-start justify-between">
        <div class="space-y-2">
          <h1 class="text-3xl font-bold">{{ challenge.title }}</h1>
          <div class="flex flex-wrap gap-2">
            <ElTag :type="getCategoryColor(challenge.category)" size="large">{{ getCategoryLabel(challenge.category) }}</ElTag>
            <ElTag :type="getDifficultyColor(challenge.difficulty)" size="large">{{ getDifficultyLabel(challenge.difficulty) }}</ElTag>
            <ElTag v-for="tag in challenge.tags" :key="tag" size="large">{{ tag }}</ElTag>
          </div>
        </div>
        <div class="text-right">
          <div class="text-3xl font-bold text-primary">{{ challenge.points }} 分</div>
          <ElTag v-if="challenge.is_solved" type="success" size="large" class="mt-2">已完成</ElTag>
        </div>
      </div>

      <ElCard>
        <template #header>
          <span class="font-semibold">挑战描述</span>
        </template>
        <div v-html="sanitizedDescription" class="prose max-w-none"></div>
        <div v-if="challenge.attachment_url" class="mt-4">
          <ElButton @click="downloadAttachment">下载附件</ElButton>
        </div>
      </ElCard>

      <ElCard v-if="!challenge.is_solved">
        <template #header>
          <span class="font-semibold">开始挑战</span>
        </template>
        <div class="space-y-4">
          <ElButton type="primary" size="large" :loading="creating" @click="startChallenge">
            {{ creating ? '正在创建实例...' : '开始挑战' }}
          </ElButton>
          <div class="text-sm text-gray-600">点击按钮创建专属实例，实例有效期为 2 小时</div>
        </div>
      </ElCard>

      <ElCard>
        <template #header>
          <span class="font-semibold">提交 Flag</span>
        </template>
        <div class="space-y-4">
          <ElInput v-model="flagInput" placeholder="flag{...}" :disabled="challenge.is_solved" @keyup.enter="submitFlagHandler">
            <template #append>
              <ElButton :loading="submitting" :disabled="challenge.is_solved" @click="submitFlagHandler">提交</ElButton>
            </template>
          </ElInput>
          <div v-if="submitResult" :class="submitResult.success ? 'text-green-600' : 'text-red-600'">
            {{ submitResult.message }}
          </div>
        </div>
      </ElCard>
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
    web: 'primary',
    pwn: 'danger',
    reverse: 'warning',
    crypto: 'success',
    misc: 'info',
    forensics: '',
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
    beginner: 'info',
    easy: 'success',
    medium: 'warning',
    hard: 'danger',
    hell: 'danger',
  }
  return colors[difficulty]
}

onMounted(() => {
  loadChallenge()
})
</script>

