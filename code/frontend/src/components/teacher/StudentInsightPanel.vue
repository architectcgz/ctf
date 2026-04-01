<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight } from 'lucide-vue-next'

import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import StudentTimelinePage from '@/components/dashboard/student/StudentTimelinePage.vue'
import SkillRadar from '@/components/common/SkillRadar.vue'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherEvidenceData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { getWeakDimensions, toRadarScores } from '@/utils/skillProfile'

const props = defineProps<{
  student: TeacherStudentItem | null
  progress: MyProgressData | null
  profile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: TeacherEvidenceData | null
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
  <div class="student-insight-shell space-y-6">
    <AppEmpty
      v-if="!student && !loading"
      title="尚未选择学员"
      :description="emptyText || '请先选择学员。'"
      icon="GraduationCap"
    />

    <template v-else>
      <div v-if="loading" class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <AppCard variant="panel" accent="neutral">
          <div class="h-6 w-36 animate-pulse rounded bg-[var(--color-bg-base)]" />
          <div class="mt-6 space-y-3">
            <div class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]" />
            <div class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]" />
          </div>
        </AppCard>
        <AppCard variant="panel" accent="neutral">
          <div class="h-[280px] animate-pulse rounded-2xl bg-[var(--color-bg-base)]" />
        </AppCard>
      </div>

      <template v-else-if="student">
        <div class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
          <SectionCard title="当前学员" subtitle="聚合进度、难度完成情况和薄弱维度。">
            <AppCard
              variant="hero"
              accent="primary"
              eyebrow="Student Snapshot"
              :title="student.name || student.username"
              subtitle="查看当前学员的关键指标和推荐方向。"
            >
              <template #header>
                <span
                  class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
                  style="
                    border-color: color-mix(
                      in srgb,
                      var(--color-primary) 18%,
                      var(--color-border-default)
                    );
                    background-color: var(--color-primary-soft);
                    color: var(--color-primary);
                  "
                >
                  @{{ student.username }}
                </span>
              </template>

              <div class="insight-kpi-grid grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                <article class="insight-kpi-card insight-kpi-card--primary">
                  <div class="insight-kpi-label">总题量</div>
                  <div class="insight-kpi-value">{{ progress?.total_challenges ?? 0 }}</div>
                  <div class="insight-kpi-hint">该学员当前纳入统计的挑战总数</div>
                </article>
                <article class="insight-kpi-card insight-kpi-card--success">
                  <div class="insight-kpi-label">已完成</div>
                  <div class="insight-kpi-value">{{ progress?.solved_challenges ?? 0 }}</div>
                  <div class="insight-kpi-hint">已成功完成的挑战数量</div>
                </article>
                <article class="insight-kpi-card insight-kpi-card--warning">
                  <div class="insight-kpi-label">薄弱维度</div>
                  <div class="insight-kpi-value">
                    {{ weakDimensions.length > 0 ? weakDimensions.join('、') : '暂无' }}
                  </div>
                  <div class="insight-kpi-hint">基于能力画像提炼的风险点</div>
                </article>
                <article class="insight-kpi-card insight-kpi-card--primary">
                  <div class="insight-kpi-label">推荐题目</div>
                  <div class="insight-kpi-value">{{ recommendations.length }}</div>
                  <div class="insight-kpi-hint">可立即布置的补强任务数量</div>
                </article>
              </div>
            </AppCard>

            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div class="rounded-2xl bg-[var(--color-bg-base)] px-5 py-4 text-center">
                <p class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-secondary)]">
                  Solved Rate
                </p>
                <p class="mt-2 text-3xl font-semibold text-[var(--color-primary)]">
                  {{
                    progress?.total_challenges
                      ? Math.round(
                          ((progress.solved_challenges ?? 0) / progress.total_challenges) * 100
                        )
                      : 0
                  }}%
                </p>
              </div>
            </div>

            <div class="mt-6 grid gap-4 xl:grid-cols-2">
              <AppCard
                variant="panel"
                accent="primary"
                eyebrow="分类进度"
                subtitle="按知识方向查看当前完成覆盖率。"
              >
                <div class="mt-4 space-y-3">
                  <div
                    v-for="(value, key) in progress?.by_category || {}"
                    :key="key"
                    class="rounded-lg bg-[var(--color-bg-surface)] px-3 py-3"
                  >
                    <div class="flex items-center justify-between text-sm">
                      <span class="font-medium text-[var(--color-text-primary)]">{{ key }}</span>
                      <span class="text-[var(--color-text-secondary)]"
                        >{{ value.solved }} / {{ value.total }}</span
                      >
                    </div>
                    <div
                      class="mt-2 h-2 overflow-hidden rounded-full bg-[var(--color-border-default)]"
                    >
                      <div
                        class="h-full rounded-full bg-[var(--color-primary)]"
                        :style="{
                          width: `${value.total ? Math.round((value.solved / value.total) * 100) : 0}%`,
                        }"
                      />
                    </div>
                  </div>
                </div>
              </AppCard>

              <AppCard
                variant="panel"
                accent="warning"
                eyebrow="难度进度"
                subtitle="按题目难度查看学员当前突破情况。"
              >
                <div class="mt-4 space-y-3">
                  <div
                    v-for="(value, key) in progress?.by_difficulty || {}"
                    :key="key"
                    class="flex items-center justify-between rounded-lg bg-[var(--color-bg-surface)] px-3 py-3 text-sm"
                  >
                    <span class="font-medium text-[var(--color-text-primary)]">{{
                      difficultyLabel(key)
                    }}</span>
                    <span class="text-[var(--color-text-secondary)]"
                      >{{ value.solved }} / {{ value.total }}</span
                    >
                  </div>
                </div>
              </AppCard>
            </div>
          </SectionCard>

          <SectionCard title="能力画像" subtitle="以雷达图观察当前能力维度分布。">
            <div class="mt-4">
              <SkillRadar :scores="radarScores" />
            </div>
          </SectionCard>
        </div>

        <SectionCard title="推荐训练任务" subtitle="根据当前能力薄弱维度筛出的优先训练题目。">
          <AppEmpty
            v-if="recommendations.length === 0"
            title="暂无推荐题目"
            description="当前画像还没有生成新的推荐训练任务。"
            icon="BookOpen"
          />

          <div v-else class="mt-5 grid gap-3 lg:grid-cols-2">
            <AppCard
              v-for="item in recommendations"
              :key="item.challenge_id"
              as="button"
              variant="action"
              accent="primary"
              interactive
              class="text-left"
              @click="openChallenge(item.challenge_id)"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <h5 class="font-semibold text-[var(--color-text-primary)]">{{ item.title }}</h5>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">{{ item.reason }}</p>
                </div>
                <span
                  class="rounded-full px-2.5 py-1 text-xs font-medium"
                  :class="difficultyClass(item.difficulty)"
                >
                  {{ difficultyLabel(item.difficulty) }}
                </span>
              </div>
              <div
                class="mt-3 inline-flex items-center gap-1 text-sm font-medium text-[var(--color-primary)]"
              >
                打开挑战
                <ArrowRight class="h-4 w-4" />
              </div>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard title="攻防证据链" subtitle="教师按关键动作查看该学员的利用过程。">
          <AppEmpty
            v-if="!evidence || evidence.events.length === 0"
            title="暂无证据链数据"
            description="当前学员还没有可用于复盘的攻击过程记录。"
            icon="NotebookText"
          />

          <template v-else>
            <div class="grid gap-3 md:grid-cols-4">
              <article class="insight-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">总事件数</div>
                <div class="insight-kpi-value">{{ evidence.summary.total_events }}</div>
                <div class="insight-kpi-hint">纳入教师复盘的动作总数</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--warning">
                <div class="insight-kpi-label">利用请求</div>
                <div class="insight-kpi-value">{{ evidence.summary.proxy_request_count }}</div>
                <div class="insight-kpi-hint">经平台代理的利用请求次数</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--success">
                <div class="insight-kpi-label">提交次数</div>
                <div class="insight-kpi-value">{{ evidence.summary.submit_count }}</div>
                <div class="insight-kpi-hint">当前题目的提交动作统计</div>
              </article>
              <article class="insight-kpi-card insight-kpi-card--primary">
                <div class="insight-kpi-label">成功次数</div>
                <div class="insight-kpi-value">{{ evidence.summary.success_count }}</div>
                <div class="insight-kpi-hint">提交命中或利用成功的次数</div>
              </article>
            </div>

            <div class="mt-5 space-y-3">
              <AppCard
                v-for="(event, index) in evidence.events"
                :key="`${event.type}-${event.challenge_id}-${event.timestamp}-${index}`"
                variant="panel"
                accent="neutral"
              >
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <div class="text-sm font-semibold text-[var(--color-text-primary)]">{{ event.title }}</div>
                    <div class="mt-1 text-sm text-[var(--color-text-secondary)]">{{ event.detail }}</div>
                    <div class="mt-2 flex flex-wrap gap-2 text-xs text-[var(--color-text-secondary)]">
                      <span
                        class="rounded-full border border-[var(--color-border-default)] px-2.5 py-1"
                      >
                        {{ String(event.meta?.event_stage || 'trace') }}
                      </span>
                      <span
                        v-if="typeof event.meta?.method === 'string'"
                        class="rounded-full border border-[var(--color-border-default)] px-2.5 py-1"
                      >
                        {{ String(event.meta?.method) }}
                      </span>
                    </div>
                  </div>
                  <div class="text-right text-xs text-[var(--color-text-secondary)]">
                    <div>{{ new Date(event.timestamp).toLocaleDateString('zh-CN') }}</div>
                    <div class="mt-1">{{ new Date(event.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) }}</div>
                  </div>
                </div>
              </AppCard>
            </div>
          </template>
        </SectionCard>

        <StudentTimelinePage :timeline="timeline" />
      </template>
    </template>
  </div>
</template>

<style scoped>
.student-insight-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  --color-primary: #4f46e5;
  --color-primary-soft: rgba(79, 70, 229, 0.08);
  --color-text-primary: var(--journal-ink);
  --color-text-secondary: var(--journal-muted);
  --color-border-default: var(--journal-border);
  --color-bg-surface: var(--journal-surface);
  --color-bg-base: #f8fafc;
}

:deep(.section-card) {
  padding: 1.1rem 1.1rem 1.05rem;
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  border-top: 1px solid var(--journal-border);
  background: var(--journal-surface-subtle);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.035);
}

:deep(.section-card__header) {
  margin-bottom: 1rem;
  border-bottom: 1px dashed rgba(148, 163, 184, 0.58);
  padding-bottom: 0.75rem;
}

:deep(.section-card__body) {
  padding-left: 0;
}

.insight-kpi-grid {
  align-items: stretch;
}

.insight-kpi-card {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: var(--journal-surface-subtle);
  padding: 0.9rem 0.95rem;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.insight-kpi-card--primary {
  border-top: 3px solid rgba(79, 70, 229, 0.42);
}

.insight-kpi-card--success {
  border-top: 3px solid rgba(16, 185, 129, 0.36);
}

.insight-kpi-card--warning {
  border-top: 3px solid rgba(245, 158, 11, 0.38);
}

.insight-kpi-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.insight-kpi-value {
  margin-top: 0.45rem;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.5;
  color: var(--journal-ink);
}

.insight-kpi-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}
</style>
