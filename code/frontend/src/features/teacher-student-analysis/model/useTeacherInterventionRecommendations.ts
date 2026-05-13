import { computed, ref, watch, type Ref } from 'vue'

import { getStudentRecommendations } from '@/api/teacher'
import type { RecommendationItem, TeacherStudentItem } from '@/api/contracts'

export interface InterventionCandidate {
  student: TeacherStudentItem
  reason: string
  accent: 'danger' | 'warning' | 'primary'
  score: number
}

interface UseTeacherInterventionRecommendationsOptions {
  students: Readonly<Ref<TeacherStudentItem[]>>
}

export function useTeacherInterventionRecommendations(
  options: UseTeacherInterventionRecommendationsOptions
) {
  const recommendationMap = ref<Record<string, RecommendationItem | null>>({})
  const recommendationLoadingMap = ref<Record<string, boolean>>({})
  const recommendationErrorMap = ref<Record<string, boolean>>({})

  const candidates = computed<InterventionCandidate[]>(() =>
    options.students.value
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

  function hasRecommendationError(studentId: string): boolean {
    return recommendationErrorMap.value[studentId] ?? false
  }

  function getCandidateClass(accent: InterventionCandidate['accent']): string {
    if (accent === 'danger') return 'intervention-item intervention-item--danger'
    if (accent === 'warning') return 'intervention-item intervention-item--warning'
    return 'intervention-item intervention-item--primary'
  }

  function getCandidatePriorityLabel(accent: InterventionCandidate['accent']): string {
    if (accent === 'danger') return '立即跟进'
    if (accent === 'warning') return '本周关注'
    return '持续观察'
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
      recommendationErrorMap.value = Object.fromEntries(targetIds.map((id) => [id, false]))

      if (targetIds.length === 0) {
        return
      }

      await Promise.all(
        targetIds.map(async (studentId) => {
          try {
            const { challenges } = await getStudentRecommendations(studentId)
            const recommendation = challenges[0] ?? null
            if (cancelled) return
            recommendationMap.value = {
              ...recommendationMap.value,
              [studentId]: recommendation,
            }
            recommendationErrorMap.value = {
              ...recommendationErrorMap.value,
              [studentId]: false,
            }
          } catch {
            if (cancelled) return
            recommendationMap.value = {
              ...recommendationMap.value,
              [studentId]: null,
            }
            recommendationErrorMap.value = {
              ...recommendationErrorMap.value,
              [studentId]: true,
            }
          } finally {
            if (cancelled) return
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

  return {
    candidates,
    getCandidateClass,
    getCandidatePriorityLabel,
    getRecommendation,
    hasRecommendationError,
    isRecommendationLoading,
  }
}
