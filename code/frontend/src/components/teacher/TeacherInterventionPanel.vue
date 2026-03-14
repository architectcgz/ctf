<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { getStudentRecommendations } from '@/api/teacher'
import type { RecommendationItem, TeacherStudentItem } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import SectionCard from '@/components/common/SectionCard.vue'

const props = defineProps<{
  students: TeacherStudentItem[]
  className?: string
}>()

interface InterventionCandidate {
  student: TeacherStudentItem
  reason: string
  accent: 'danger' | 'warning' | 'primary'
  score: number
}

const recommendationMap = ref<Record<string, RecommendationItem | null>>({})
const recommendationLoadingMap = ref<Record<string, boolean>>({})

const candidates = computed<InterventionCandidate[]>(() =>
  props.students
    .map((student) => {
      const recentEventCount = student.recent_event_count ?? 0
      const solvedCount = student.solved_count ?? 0
      const weakDimension = student.weak_dimension?.trim()

      let score = 0
      let reason = ''
      let accent: InterventionCandidate['accent'] = 'primary'

      if (recentEventCount === 0) {
        score += 100
        reason = weakDimension
          ? `近 7 天无训练动作，且薄弱项集中在 ${weakDimension}`
          : '近 7 天无训练动作'
        accent = 'danger'
      } else if (recentEventCount <= 1) {
        score += 70
        reason = weakDimension
          ? `近 7 天训练动作偏少，建议优先补强 ${weakDimension}`
          : '近 7 天训练动作偏少'
        accent = 'warning'
      }

      if (solvedCount === 0) {
        score += 60
        reason =
          reason ||
          (weakDimension ? `尚未解出题目，建议从 ${weakDimension} 基础题开始` : '尚未解出题目')
        accent = accent === 'danger' ? accent : 'warning'
      } else if (solvedCount <= 1) {
        score += 25
        if (!reason && weakDimension) {
          reason = `解题数偏少，建议定向补强 ${weakDimension}`
        }
      }

      if (!reason && weakDimension) {
        score += 10
        reason = `当前薄弱项为 ${weakDimension}，适合定向布置补强训练`
      }

      return {
        student,
        reason,
        accent,
        score,
      }
    })
    .filter((item) => item.score > 0)
    .sort((left, right) => {
      const scoreGap = right.score - left.score
      if (scoreGap !== 0) return scoreGap
      const eventGap =
        (left.student.recent_event_count ?? 0) - (right.student.recent_event_count ?? 0)
      if (eventGap !== 0) return eventGap
      const solvedGap = (left.student.solved_count ?? 0) - (right.student.solved_count ?? 0)
      if (solvedGap !== 0) return solvedGap
      return (left.student.username || '').localeCompare(right.student.username || '')
    })
    .slice(0, 5)
)

const recommendationTargets = computed(() =>
  candidates.value.filter((item) => item.score >= 25).slice(0, 3)
)

function getRecommendation(studentId: string): RecommendationItem | null {
  return recommendationMap.value[studentId] ?? null
}

function isRecommendationLoading(studentId: string): boolean {
  return recommendationLoadingMap.value[studentId] ?? false
}

watch(
  recommendationTargets,
  async (nextTargets, _, onCleanup) => {
    let cancelled = false
    onCleanup(() => {
      cancelled = true
    })

    const targetIds = nextTargets.map((item) => item.student.id)
    recommendationMap.value = Object.fromEntries(
      targetIds.map((id) => [id, recommendationMap.value[id] ?? null])
    )
    recommendationLoadingMap.value = Object.fromEntries(targetIds.map((id) => [id, true]))

    if (targetIds.length === 0) {
      return
    }

    await Promise.all(
      targetIds.map(async (studentId) => {
        try {
          const [recommendation] = await getStudentRecommendations(studentId)
          if (cancelled) {
            return
          }
          recommendationMap.value = {
            ...recommendationMap.value,
            [studentId]: recommendation ?? null,
          }
        } catch {
          if (cancelled) {
            return
          }
          recommendationMap.value = {
            ...recommendationMap.value,
            [studentId]: null,
          }
        } finally {
          if (cancelled) {
            return
          }
          recommendationLoadingMap.value = {
            ...recommendationLoadingMap.value,
            [studentId]: false,
          }
        }
      })
    )
  },
  { immediate: true }
)
</script>

<template>
  <SectionCard
    title="优先介入学生"
    :subtitle="
      className
        ? `${className} 当前最值得优先跟进的学生名单。`
        : '当前班级最值得优先跟进的学生名单。'
    "
  >
    <AppEmpty
      v-if="candidates.length === 0"
      icon="GraduationCap"
      title="暂无高优先级介入对象"
      description="当前班级学生的训练活跃度和解题表现暂时没有明显风险。"
    />

    <div v-else class="space-y-3">
      <AppCard
        v-for="item in candidates"
        :key="item.student.id"
        variant="action"
        :accent="item.accent"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <div class="font-semibold text-text-primary">
              {{ item.student.name || item.student.username }}
            </div>
            <div class="mt-1 text-sm text-text-secondary">@{{ item.student.username }}</div>
            <div class="mt-3 text-sm leading-6 text-text-secondary">{{ item.reason }}</div>

            <div
              v-if="isRecommendationLoading(item.student.id)"
              class="mt-3 rounded-2xl border border-border bg-base/60 px-4 py-3 text-sm text-text-secondary"
            >
              正在匹配建议训练题...
            </div>

            <div
              v-else-if="getRecommendation(item.student.id)"
              class="mt-3 rounded-2xl border border-border bg-base/60 px-4 py-3"
            >
              <div class="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary/80">
                建议训练题
              </div>
              <div class="mt-2 text-sm font-semibold text-text-primary">
                {{ getRecommendation(item.student.id)?.title }}
              </div>
              <div class="mt-1 text-xs text-text-secondary">
                {{ getRecommendation(item.student.id)?.category }} /
                {{ getRecommendation(item.student.id)?.difficulty }}
              </div>
              <div class="mt-2 text-sm leading-6 text-text-secondary">
                {{ getRecommendation(item.student.id)?.reason }}
              </div>
            </div>
          </div>

          <div class="text-right text-xs text-text-secondary">
            <div>{{ item.student.recent_event_count ?? 0 }} 次动作</div>
            <div class="mt-1">
              {{ item.student.solved_count ?? 0 }} 题 / {{ item.student.total_score ?? 0 }} 分
            </div>
          </div>
        </div>
      </AppCard>
    </div>
  </SectionCard>
</template>
