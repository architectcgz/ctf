<script setup lang="ts">
import { computed } from 'vue'

import SkillRadar from '@/components/common/SkillRadar.vue'
import type { MyProgressData, RecommendationItem, SkillProfileData, TeacherStudentItem } from '@/api/contracts'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { getWeakDimensions, toRadarScores } from '@/utils/skillProfile'

const props = defineProps<{
  student: TeacherStudentItem | null
  progress: MyProgressData | null
  profile: SkillProfileData | null
  recommendations: RecommendationItem[]
  loading: boolean
  emptyText?: string
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
}>()

const radarScores = computed(() => toRadarScores(props.profile))
const weakDimensions = computed(() => getWeakDimensions(props.profile))

function openChallenge(challengeId: string): void {
  emit('openChallenge', challengeId)
}
</script>

<template>
  <div class="space-y-6">
    <div
      v-if="!student && !loading"
      class="rounded-2xl border border-dashed border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-6 py-12 text-center text-sm text-[var(--color-text-secondary)]"
    >
      {{ emptyText || '请先选择学员。' }}
    </div>

    <template v-else>
      <div v-if="loading" class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
          <div class="h-6 w-36 animate-pulse rounded bg-[var(--color-bg-base)]"></div>
          <div class="mt-6 space-y-3">
            <div class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]"></div>
            <div class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]"></div>
          </div>
        </div>
        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6">
          <div class="h-[280px] animate-pulse rounded-2xl bg-[var(--color-bg-base)]"></div>
        </div>
      </div>

      <template v-else-if="student">
        <div class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
          <section class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="text-sm text-[var(--color-text-secondary)]">当前学员</p>
                <h3 class="mt-1 text-2xl font-semibold text-[var(--color-text-primary)]">
                  {{ student.name || student.username }}
                </h3>
                <p class="mt-1 text-sm text-[var(--color-text-secondary)]">@{{ student.username }}</p>
              </div>

              <div class="rounded-2xl bg-[var(--color-bg-base)] px-5 py-4 text-center">
                <p class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-secondary)]">Solved Rate</p>
                <p class="mt-2 text-3xl font-semibold text-[var(--color-primary)]">
                  {{ progress?.total_challenges ? Math.round((progress.solved_challenges / progress.total_challenges) * 100) : 0 }}%
                </p>
              </div>
            </div>

            <div class="mt-6 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <div class="rounded-xl bg-[var(--color-bg-base)] px-4 py-4">
                <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">总题量</p>
                <p class="mt-2 text-xl font-semibold text-[var(--color-text-primary)]">{{ progress?.total_challenges ?? 0 }}</p>
              </div>
              <div class="rounded-xl bg-[var(--color-bg-base)] px-4 py-4">
                <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">已完成</p>
                <p class="mt-2 text-xl font-semibold text-[var(--color-text-primary)]">{{ progress?.solved_challenges ?? 0 }}</p>
              </div>
              <div class="rounded-xl bg-[var(--color-bg-base)] px-4 py-4">
                <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">薄弱维度</p>
                <p class="mt-2 text-sm font-medium text-[var(--color-text-primary)]">
                  {{ weakDimensions.length > 0 ? weakDimensions.join('、') : '暂无' }}
                </p>
              </div>
              <div class="rounded-xl bg-[var(--color-bg-base)] px-4 py-4">
                <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">推荐题目</p>
                <p class="mt-2 text-xl font-semibold text-[var(--color-text-primary)]">{{ recommendations.length }}</p>
              </div>
            </div>

            <div class="mt-6 grid gap-4 xl:grid-cols-2">
              <div class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] p-4">
                <h4 class="text-sm font-semibold text-[var(--color-text-primary)]">分类进度</h4>
                <div class="mt-4 space-y-3">
                  <div
                    v-for="(value, key) in progress?.by_category || {}"
                    :key="key"
                    class="rounded-lg bg-[var(--color-bg-surface)] px-3 py-3"
                  >
                    <div class="flex items-center justify-between text-sm">
                      <span class="font-medium text-[var(--color-text-primary)]">{{ key }}</span>
                      <span class="text-[var(--color-text-secondary)]">{{ value.solved }} / {{ value.total }}</span>
                    </div>
                    <div class="mt-2 h-2 overflow-hidden rounded-full bg-[var(--color-border-default)]">
                      <div
                        class="h-full rounded-full bg-[var(--color-primary)]"
                        :style="{ width: `${value.total ? Math.round((value.solved / value.total) * 100) : 0}%` }"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <div class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] p-4">
                <h4 class="text-sm font-semibold text-[var(--color-text-primary)]">难度进度</h4>
                <div class="mt-4 space-y-3">
                  <div
                    v-for="(value, key) in progress?.by_difficulty || {}"
                    :key="key"
                    class="flex items-center justify-between rounded-lg bg-[var(--color-bg-surface)] px-3 py-3 text-sm"
                  >
                    <span class="font-medium text-[var(--color-text-primary)]">{{ difficultyLabel(key) }}</span>
                    <span class="text-[var(--color-text-secondary)]">{{ value.solved }} / {{ value.total }}</span>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
            <h4 class="text-sm font-semibold text-[var(--color-text-primary)]">能力画像</h4>
            <div class="mt-4">
              <SkillRadar :scores="radarScores" />
            </div>
          </section>
        </div>

        <section class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
          <div class="flex items-center justify-between gap-4">
            <div>
              <h4 class="text-lg font-semibold text-[var(--color-text-primary)]">推荐训练任务</h4>
              <p class="mt-1 text-sm text-[var(--color-text-secondary)]">根据当前能力薄弱维度筛出的优先训练题目。</p>
            </div>
          </div>

          <div v-if="recommendations.length === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
            暂无推荐题目。
          </div>

          <div v-else class="mt-5 grid gap-3 lg:grid-cols-2">
            <button
              v-for="item in recommendations"
              :key="item.challenge_id"
              type="button"
              class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] p-4 text-left transition hover:border-[var(--color-primary)]/60 hover:shadow-sm"
              @click="openChallenge(item.challenge_id)"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h5 class="font-semibold text-[var(--color-text-primary)]">{{ item.title }}</h5>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">{{ item.reason }}</p>
                </div>
                <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="difficultyClass(item.difficulty)">
                  {{ difficultyLabel(item.difficulty) }}
                </span>
              </div>
            </button>
          </div>
        </section>
      </template>
    </template>
  </div>
</template>
