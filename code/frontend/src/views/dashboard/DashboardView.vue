<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getMyProgress, getMyTimeline, getRecommendations, getSkillProfile } from '@/api/assessment'
import type { MyProgressData, RecommendationItem, SkillProfileData, TimelineEvent } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { formatDate } from '@/utils/format'
import { getWeakDimensions } from '@/utils/skillProfile'

const authStore = useAuthStore()
const router = useRouter()

const loading = ref(false)
const error = ref<string | null>(null)
const progress = ref<MyProgressData | null>(null)
const timeline = ref<TimelineEvent[]>([])
const recommendations = ref<RecommendationItem[]>([])
const skillProfile = ref<SkillProfileData | null>(null)

const displayName = computed(() => authStore.user?.name || authStore.user?.username || '选手')
const weakDimensions = computed(() => getWeakDimensions(skillProfile.value).slice(0, 3))
const completionRate = computed(() => {
  const solved = progress.value?.total_solved ?? 0
  const total = progress.value?.category_stats?.reduce((sum, item) => sum + item.total, 0) ?? 0
  if (!total) return 0
  return Math.round((solved / total) * 100)
})

function categoryRate(total: number, solved: number): number {
  if (!total) return 0
  return Math.round((solved / total) * 100)
}

function timelineSummary(event: TimelineEvent): string {
  if (event.type === 'solve') {
    return `成功解出题目${event.points ? `，获得 ${event.points} 分` : ''}`
  }
  if (event.type === 'submit') {
    return '提交过 Flag，当前记录未判定为成功'
  }
  if ((event.meta?.raw_type as string | undefined) === 'instance_destroy') {
    return '结束了一个练习实例'
  }
  return '启动或操作了练习实例'
}

async function loadDashboard(): Promise<void> {
  const role = authStore.user?.role
  if (role === 'teacher') {
    await router.replace({ name: 'TeacherDashboard' })
    return
  }
  if (role === 'admin') {
    await router.replace({ name: 'AdminDashboard' })
    return
  }

  loading.value = true
  error.value = null
  try {
    const [progressPayload, timelinePayload, recommendationPayload, profilePayload] = await Promise.all([
      getMyProgress(),
      getMyTimeline(),
      getRecommendations(),
      getSkillProfile(),
    ])

    progress.value = progressPayload
    timeline.value = timelinePayload.slice(0, 6)
    recommendations.value = recommendationPayload.slice(0, 4)
    skillProfile.value = profilePayload
  } catch (err) {
    console.error('加载学生仪表盘失败:', err)
    error.value = '加载仪表盘失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<template>
  <div class="space-y-6">
    <section class="rounded-[28px] border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(14,116,144,0.10),rgba(34,197,94,0.12))] p-7 shadow-sm">
      <p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--color-primary)]/85">Student Workspace</p>
      <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--color-text-primary)]">{{ displayName }} 的训练仪表盘</h1>
      <p class="mt-2 max-w-3xl text-sm leading-6 text-[var(--color-text-secondary)]">
        汇总当前得分、解题进度、近期训练动态与推荐靶场，方便你快速决定下一步训练重点。
      </p>
      <p v-if="authStore.user?.class_name" class="mt-3 text-sm text-[var(--color-text-secondary)]">
        所属班级：<span class="font-medium text-[var(--color-text-primary)]">{{ authStore.user.class_name }}</span>
      </p>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="loadDashboard">重试</button>
    </div>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div v-for="index in 4" :key="index" class="h-32 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]"></div>
    </div>

    <template v-else-if="progress">
      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">总得分</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ progress.total_score ?? 0 }}</p>
        </div>
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">已解题数</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ progress.total_solved ?? 0 }}</p>
        </div>
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">当前排名</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-primary)]">#{{ progress.rank ?? '-' }}</p>
        </div>
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">完成率</p>
          <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ completionRate }}%</p>
        </div>
      </section>

      <section class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div class="space-y-6">
          <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
            <div class="flex items-center justify-between gap-4">
              <div>
                <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">分类进度</h2>
                <p class="mt-1 text-sm text-[var(--color-text-secondary)]">看看哪些方向已经形成稳定解题能力。</p>
              </div>
              <button
                type="button"
                class="rounded-xl border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)]"
                @click="router.push({ name: 'SkillProfile' })"
              >
                查看能力画像
              </button>
            </div>

            <div v-if="(progress.category_stats?.length ?? 0) === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
              当前还没有分类统计数据，先去完成几道题再回来查看。
            </div>

            <div v-else class="mt-5 space-y-4">
              <div v-for="item in progress.category_stats" :key="item.category" class="space-y-2">
                <div class="flex items-center justify-between gap-4 text-sm">
                  <span class="font-medium uppercase text-[var(--color-text-primary)]">{{ item.category }}</span>
                  <span class="text-[var(--color-text-secondary)]">{{ item.solved }} / {{ item.total }}</span>
                </div>
                <div class="h-2 rounded-full bg-[var(--color-bg-base)]">
                  <div class="h-2 rounded-full bg-[var(--color-primary)]" :style="{ width: `${categoryRate(item.total, item.solved)}%` }"></div>
                </div>
              </div>
            </div>
          </div>

          <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
            <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">近期动态</h2>
            <p class="mt-1 text-sm text-[var(--color-text-secondary)]">保留最近几次实例操作和提交记录，方便回看训练节奏。</p>

            <div v-if="timeline.length === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
              当前还没有训练动态。
            </div>

            <div v-else class="mt-5 space-y-4">
              <div v-for="event in timeline" :key="event.id" class="flex gap-4 rounded-xl border border-[var(--color-border-default)] px-4 py-4">
                <div class="mt-1 h-2.5 w-2.5 flex-shrink-0 rounded-full" :class="event.type === 'solve' ? 'bg-emerald-500' : event.type === 'submit' ? 'bg-amber-500' : 'bg-sky-500'"></div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-3">
                    <p class="font-medium text-[var(--color-text-primary)]">{{ event.title }}</p>
                    <span class="text-xs text-[var(--color-text-secondary)]">{{ formatDate(event.created_at) }}</span>
                  </div>
                  <p class="mt-1 text-sm leading-6 text-[var(--color-text-secondary)]">{{ timelineSummary(event) }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="space-y-6">
          <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
            <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">训练建议</h2>
            <p class="mt-1 text-sm text-[var(--color-text-secondary)]">基于当前画像推荐更适合补短板的题目。</p>

            <div v-if="weakDimensions.length > 0" class="mt-4 flex flex-wrap gap-2">
              <span
                v-for="item in weakDimensions"
                :key="item"
                class="rounded-full bg-amber-500/10 px-3 py-1 text-xs font-medium text-amber-700"
              >
                待加强：{{ item }}
              </span>
            </div>

            <div v-if="recommendations.length === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
              暂无推荐题目，可以先去挑战列表挑一道新题。
            </div>

            <div v-else class="mt-5 space-y-3">
              <button
                v-for="item in recommendations"
                :key="item.challenge_id"
                type="button"
                class="w-full rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-4 text-left transition hover:-translate-y-0.5 hover:border-[var(--color-primary)]"
                @click="router.push(`/challenges/${item.challenge_id}`)"
              >
                <div class="flex items-center justify-between gap-3">
                  <p class="font-medium text-[var(--color-text-primary)]">{{ item.title }}</p>
                  <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="difficultyClass(item.difficulty)">
                    {{ difficultyLabel(item.difficulty) }}
                  </span>
                </div>
                <p class="mt-2 text-sm leading-6 text-[var(--color-text-secondary)]">{{ item.reason }}</p>
              </button>
            </div>
          </div>

          <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
            <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">难度分布</h2>
            <p class="mt-1 text-sm text-[var(--color-text-secondary)]">观察自己在不同难度上的完成情况，避免训练结构失衡。</p>

            <div v-if="(progress.difficulty_stats?.length ?? 0) === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
              暂无难度统计数据。
            </div>

            <div v-else class="mt-5 space-y-4">
              <div v-for="item in progress.difficulty_stats" :key="item.difficulty" class="space-y-2">
                <div class="flex items-center justify-between gap-4 text-sm">
                  <span class="font-medium text-[var(--color-text-primary)]">{{ difficultyLabel(item.difficulty) }}</span>
                  <span class="text-[var(--color-text-secondary)]">{{ item.solved }} / {{ item.total }}</span>
                </div>
                <div class="h-2 rounded-full bg-[var(--color-bg-base)]">
                  <div class="h-2 rounded-full bg-emerald-500" :style="{ width: `${categoryRate(item.total, item.solved)}%` }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
