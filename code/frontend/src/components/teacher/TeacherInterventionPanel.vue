<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { getStudentRecommendations } from '@/api/teacher'
import type { RecommendationItem, TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  students: TeacherStudentItem[]
  className?: string
  bare?: boolean
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

function getCandidateClass(accent: InterventionCandidate['accent']): string {
  if (accent === 'danger') return 'intervention-item intervention-item--danger'
  if (accent === 'warning') return 'intervention-item intervention-item--warning'
  return 'intervention-item intervention-item--primary'
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
  <section
    class="teacher-panel"
    :class="{ 'teacher-panel--shellless': bare }"
  >
    <header
      v-if="!bare"
      class="teacher-panel__header"
    >
      <div class="journal-eyebrow">
        Intervention
      </div>
      <h2 class="teacher-panel__title">
        优先介入学生
      </h2>
      <p class="teacher-panel__subtitle">
        {{
          className
            ? `${className} 当前最值得优先跟进的学生名单。`
            : '当前班级最值得优先跟进的学生名单。'
        }}
      </p>
    </header>

    <AppEmpty
      v-if="candidates.length === 0"
      icon="GraduationCap"
      title="暂无高优先级介入对象"
      description="当前班级学生的训练活跃度和解题表现暂时没有明显风险。"
    />

    <div
      v-else
      class="intervention-list intervention-list--premium"
    >
      <article
        v-for="item in candidates"
        :key="item.student.id"
        :class="getCandidateClass(item.accent)"
      >
        <div class="intervention-item__layout">
          <div class="intervention-item__main">
            <div class="intervention-item__name-row">
              <span class="intervention-item__name">{{
                item.student.name || item.student.username
              }}</span>
              <span class="intervention-item__username">@{{ item.student.username }}</span>
            </div>
            <div class="intervention-item__reason">
              {{ item.reason }}
            </div>

            <div
              v-if="isRecommendationLoading(item.student.id)"
              class="intervention-item__recommendation intervention-item__recommendation--loading"
            >
              正在匹配建议训练题...
            </div>

            <div
              v-else-if="getRecommendation(item.student.id)"
              class="intervention-item__recommendation intervention-item__recommendation--premium"
            >
              <div class="intervention-item__recommendation-label">
                建议训练题
              </div>
              <div class="intervention-item__recommendation-body">
                <div class="recommendation-info">
                  <div class="intervention-item__recommendation-title">
                    {{ getRecommendation(item.student.id)?.title }}
                  </div>
                  <div class="intervention-item__recommendation-meta">
                    {{ getRecommendation(item.student.id)?.category }} ·
                    {{ getRecommendation(item.student.id)?.difficulty }}
                  </div>
                </div>
                <div class="intervention-item__recommendation-reason">
                  {{ getRecommendation(item.student.id)?.reason }}
                </div>
              </div>
            </div>
          </div>

          <div class="intervention-item__stats intervention-item__stats--premium">
            <div class="stat-row">
              <span class="stat-value">{{ item.student.recent_event_count ?? 0 }}</span>
              <span class="stat-label">动作</span>
            </div>
            <div class="stat-row">
              <span class="stat-value">{{ item.student.solved_count ?? 0 }}</span>
              <span class="stat-label">解题</span>
            </div>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
@import './teacher-panel-shell.css';

.intervention-list {
  display: grid;
  gap: var(--space-4);
}

.intervention-item {
  --intervention-accent: var(--panel-accent);
  border-radius: 24px;
  border: 1px solid color-mix(in srgb, var(--intervention-accent) 12%, var(--panel-border));
  border-left: 5px solid color-mix(in srgb, var(--intervention-accent) 64%, transparent);
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--panel-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--panel-surface-subtle) 96%, var(--color-bg-base))
  );
  padding: var(--space-6) var(--space-7);
  box-shadow:
    0 1px 3px 0 rgb(0 0 0 / 0.1),
    0 1px 2px -1px rgb(0 0 0 / 0.1);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.intervention-item:hover {
  transform: translateX(6px);
  box-shadow:
    0 10px 15px -3px rgb(0 0 0 / 0.1),
    0 4px 6px -4px rgb(0 0 0 / 0.1);
  border-color: color-mix(in srgb, var(--intervention-accent) 30%, var(--panel-border));
}

.intervention-item--primary {
  --intervention-accent: var(--panel-accent);
}

.intervention-item--warning {
  --intervention-accent: var(--color-warning);
}

.intervention-item--danger {
  --intervention-accent: var(--color-danger);
}

.intervention-item__layout {
  display: flex;
  justify-content: space-between;
  gap: var(--space-5);
}

.intervention-item__main {
  min-width: 0;
}

.intervention-item__name-row {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
}

.intervention-item__name {
  font-size: var(--font-size-17);
  font-weight: 800;
  color: var(--panel-ink);
}

.intervention-item__username {
  font-size: var(--font-size-13);
  color: var(--panel-muted);
}

.intervention-item__reason {
  margin-top: var(--space-2);
  font-size: var(--font-size-15);
  line-height: 1.7;
  color: var(--panel-muted);
}

.intervention-item__recommendation--premium {
  margin-top: var(--space-5);
  border-top: 1px solid color-mix(in srgb, var(--intervention-accent) 12%, var(--panel-border));
  padding-top: var(--space-4);
}

.intervention-item__recommendation--loading {
  margin-top: var(--space-3);
  font-size: var(--font-size-13);
  font-style: italic;
  color: var(--panel-muted);
}

.intervention-item__recommendation-label {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--intervention-accent) 76%, var(--panel-muted));
}

.intervention-item__recommendation-body {
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: var(--space-5);
  margin-top: var(--space-2);
}

.intervention-item__recommendation-title {
  font-size: var(--font-size-15);
  font-weight: 800;
  color: var(--panel-ink);
}

.intervention-item__recommendation-meta {
  margin-top: var(--space-0-5);
  font-size: var(--font-size-13);
  color: var(--panel-muted);
}

.intervention-item__recommendation-reason {
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--panel-muted);
}

.intervention-item__stats--premium {
  flex-shrink: 0;
  text-align: right;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: var(--space-2);
  padding-left: var(--space-5);
  border-left: 1px solid var(--panel-divider);
}

.stat-row {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.stat-value {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-18);
  font-weight: 800;
  color: var(--panel-ink);
}

.stat-label {
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--panel-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

@media (max-width: 768px) {
  .intervention-item__layout {
    flex-direction: column;
    gap: var(--space-3);
  }

  .intervention-item__stats--premium {
    flex-direction: row;
    justify-content: flex-start;
    padding-left: 0;
    padding-top: var(--space-3);
    border-left: 0;
    border-top: 1px solid var(--panel-divider);
    gap: var(--space-6);
  }

  .stat-row {
    align-items: flex-start;
  }

  .intervention-item__recommendation-body {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
}
</style>
