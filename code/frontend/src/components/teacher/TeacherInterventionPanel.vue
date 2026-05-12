<script setup lang="ts">
import { toRef } from 'vue'
import { useRouter } from 'vue-router'

import type { TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import {
  ChallengeCategoryDifficultyPills,
  ChallengeCategoryPill,
  toChallengeCategory,
} from '@/entities/challenge'
import { useTeacherInterventionRecommendations } from '@/features/teacher-student-analysis'

const props = defineProps<{
  students: TeacherStudentItem[]
  className?: string
  bare?: boolean
}>()

const router = useRouter()

const {
  candidates,
  getRecommendation,
  isRecommendationLoading,
  getCandidateClass,
  getCandidatePriorityLabel,
} = useTeacherInterventionRecommendations({
  students: toRef(props, 'students'),
})

function openStudent(studentId: string): void {
  if (!props.className) return
  router.push({
    name: 'TeacherStudentAnalysis',
    params: {
      className: props.className,
      studentId,
    },
  })
}

function weakDimensionCategory(value?: string | null) {
  return toChallengeCategory(value)
}
</script>

<template>
  <section class="teacher-panel" :class="{ 'teacher-panel--shellless': bare }">
    <header v-if="!bare" class="teacher-panel__header">
      <h2 class="teacher-panel__title">优先介入学生</h2>
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

    <div v-else class="intervention-list intervention-list--premium">
      <article
        v-for="item in candidates"
        :key="item.student.id"
        :class="getCandidateClass(item.accent)"
      >
        <div class="intervention-item__header">
          <div class="intervention-item__identity">
            <div class="intervention-item__name-row">
              <button
                type="button"
                class="intervention-item__name-button"
                :aria-label="`${item.student.name || item.student.username}，查看学员分析`"
                :disabled="!className"
                @click="openStudent(item.student.id)"
              >
                <span class="intervention-item__name">{{
                  item.student.name || item.student.username
                }}</span>
              </button>
              <span class="intervention-item__priority">
                {{ getCandidatePriorityLabel(item.accent) }}
              </span>
            </div>
            <div class="intervention-item__summary-row">
              <span class="intervention-item__signal-inline">
                {{ item.reason }}
              </span>
              <span class="intervention-item__meta-inline intervention-item__meta-inline--username">
                {{ item.student.username }}
              </span>
              <span
                v-if="item.student.weak_dimension"
                class="intervention-item__meta-inline intervention-item__weak-inline"
              >
                <span>薄弱项</span>
                <ChallengeCategoryPill
                  v-if="weakDimensionCategory(item.student.weak_dimension)"
                  :category="weakDimensionCategory(item.student.weak_dimension)!"
                />
                <span v-else>{{ item.student.weak_dimension }}</span>
              </span>
            </div>
          </div>

          <div class="intervention-item__stats intervention-item__stats--premium">
            <div class="stat-row">
              <span class="stat-label">动作</span>
              <span class="stat-value">{{ item.student.recent_event_count ?? 0 }}</span>
            </div>
            <div class="stat-row">
              <span class="stat-label">解题</span>
              <span class="stat-value">{{ item.student.solved_count ?? 0 }}</span>
            </div>
          </div>
        </div>

        <div class="intervention-item__main">
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
            <div class="intervention-item__recommendation-heading">
              <div class="intervention-item__recommendation-label">建议训练题</div>
              <div class="intervention-item__recommendation-kicker">可直接布置</div>
            </div>
            <div class="intervention-item__recommendation-body">
              <div class="recommendation-info">
                <div class="intervention-item__recommendation-title">
                  {{ getRecommendation(item.student.id)?.title }}
                </div>
                <ChallengeCategoryDifficultyPills
                  v-if="getRecommendation(item.student.id)"
                  class="intervention-item__recommendation-meta"
                  :category="getRecommendation(item.student.id)!.category"
                  :difficulty="getRecommendation(item.student.id)!.difficulty"
                />
              </div>
              <div class="intervention-item__recommendation-reason">
                {{ getRecommendation(item.student.id)?.summary }}
              </div>
            </div>
            <div
              v-if="getRecommendation(item.student.id)?.evidence"
              class="intervention-item__recommendation-evidence"
            >
              {{ getRecommendation(item.student.id)?.evidence }}
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
  border-radius: 18px;
  border: 1px solid var(--panel-border);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--panel-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--panel-surface-subtle) 94%, var(--color-bg-base))
    ),
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--intervention-accent) 8%, transparent),
      transparent 34%
    );
  padding: var(--space-4) var(--space-5);
  box-shadow:
    0 1px 2px rgb(15 23 42 / 0.08),
    0 10px 22px rgb(15 23 42 / 0.05);
  transition:
    border-color 0.24s ease,
    box-shadow 0.24s ease,
    transform 0.24s ease;
}

.intervention-item:hover {
  transform: translateY(calc(var(--space-0-5) * -1));
  box-shadow:
    0 2px 4px rgb(15 23 42 / 0.1),
    0 14px 28px color-mix(in srgb, var(--intervention-accent) 8%, rgb(15 23 42 / 0.08));
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

.intervention-item__header {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  align-items: flex-start;
}

.intervention-item__identity,
.intervention-item__main {
  min-width: 0;
}

.intervention-item__name-row {
  display: flex;
  align-items: center;
  gap: var(--space-2-5);
  flex-wrap: wrap;
}

.intervention-item__name-button {
  display: inline-flex;
  align-items: center;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.intervention-item__name-button:disabled {
  cursor: default;
}

.intervention-item__name-button:hover .intervention-item__name,
.intervention-item__name-button:focus-visible .intervention-item__name {
  color: color-mix(in srgb, var(--intervention-accent) 78%, var(--panel-ink));
}

.intervention-item__name-button:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--intervention-accent) 38%, transparent);
  outline-offset: 2px;
  border-radius: 6px;
}

.intervention-item__name {
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--panel-ink);
  transition: color 0.2s ease;
}

.intervention-item__priority {
  display: inline-flex;
  align-items: center;
  min-height: 1.5rem;
  border-radius: 999px;
  padding: 0 var(--space-2-5);
  background: color-mix(in srgb, var(--intervention-accent) 16%, transparent);
  color: color-mix(in srgb, var(--intervention-accent) 72%, var(--panel-ink));
  font-size: var(--font-size-12);
  font-weight: 700;
  letter-spacing: 0.04em;
}

.intervention-item__summary-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-1-5);
  align-items: center;
}

.intervention-item__signal-inline {
  display: inline-flex;
  align-items: center;
  min-height: 1.5rem;
  max-width: 100%;
  border-radius: 999px;
  padding: 0 var(--space-2-5);
  background: color-mix(in srgb, var(--intervention-accent) 10%, transparent);
  font-size: var(--font-size-12);
  font-weight: 600;
  line-height: 1.4;
  color: color-mix(in srgb, var(--intervention-accent) 76%, var(--panel-ink));
}

.intervention-item__meta-inline {
  font-size: var(--font-size-12);
  color: var(--panel-muted);
}

.intervention-item__meta-inline--username {
  font-family: var(--font-family-mono);
  color: var(--panel-ink);
}

.intervention-item__weak-inline {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
}

.intervention-item__recommendation--premium {
  margin-top: var(--space-3);
  border: 1px solid var(--panel-border);
  border-radius: 14px;
  background: color-mix(in srgb, var(--panel-surface-subtle) 90%, transparent);
  padding: var(--space-3) var(--space-4);
}

.intervention-item__recommendation--loading {
  margin-top: var(--space-3);
  padding: var(--space-2-5) var(--space-3);
  border-radius: 14px;
  background: color-mix(in srgb, var(--panel-surface-subtle) 84%, transparent);
  font-size: var(--font-size-12);
  color: var(--panel-muted);
}

.intervention-item__recommendation-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.intervention-item__recommendation-label {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--intervention-accent) 76%, var(--panel-muted));
}

.intervention-item__recommendation-kicker {
  font-size: var(--font-size-12);
  color: var(--panel-muted);
}

.intervention-item__recommendation-body {
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: var(--space-4);
  margin-top: var(--space-2);
}

.intervention-item__recommendation-title {
  font-size: var(--font-size-14);
  font-weight: 800;
  color: var(--panel-ink);
}

.intervention-item__recommendation-meta {
  margin-top: var(--space-0-5);
  font-size: var(--font-size-13);
  color: var(--panel-muted);
}

.intervention-item__recommendation-reason {
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--panel-muted);
}

.intervention-item__recommendation-evidence {
  margin-top: var(--space-2);
  font-size: var(--font-size-12);
  line-height: 1.6;
  color: color-mix(in srgb, var(--panel-muted) 82%, var(--panel-ink));
}

.intervention-item__stats--premium {
  flex-shrink: 0;
  display: flex;
  gap: var(--space-3);
  padding-top: var(--space-0-5);
}

.stat-row {
  display: flex;
  align-items: baseline;
  gap: var(--space-1-5);
}

.stat-value {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-15);
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
  .intervention-item__header {
    flex-direction: column;
    gap: var(--space-2-5);
  }

  .intervention-item__stats--premium {
    width: 100%;
    padding-top: 0;
  }

  .stat-row {
    flex: 0 0 auto;
  }

  .intervention-item__recommendation-body {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }

  .intervention-item__recommendation-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
