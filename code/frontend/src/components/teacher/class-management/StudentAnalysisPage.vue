<script setup lang="ts">
import { ArrowLeftRight, ChevronLeft, FileDown, GraduationCap, Users } from 'lucide-vue-next'

import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherEvidenceData,
  TeacherClassItem,
  TeacherSubmissionWriteupItemData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import StudentInsightPanel from '@/components/teacher/StudentInsightPanel.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  students: TeacherStudentItem[]
  selectedClassName: string
  selectedStudentId: string
  selectedStudent: TeacherStudentItem | null
  loadingClasses: boolean
  loadingStudents: boolean
  loadingDetails: boolean
  error: string | null
  progress: MyProgressData | null
  skillProfile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: TeacherEvidenceData | null
  writeupSubmissions: TeacherSubmissionWriteupItemData[]
  solvedRate: number
  weakDimensions: string[]
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openClassStudents: []
  openReportExport: []
  selectClass: [className: string]
  selectStudent: [studentId: string]
  openChallenge: [challengeId: string]
}>()
</script>

<template>
  <div class="teacher-analysis-shell space-y-6">
    <PageHeader
      eyebrow="Student Analysis"
      :title="selectedStudent?.name || selectedStudent?.username || '学员分析'"
      description="查看学员的能力画像、进度和推荐任务。"
    >
      <ElButton plain @click="emit('openClassManagement')">班级管理</ElButton>
      <ElButton plain @click="emit('openClassStudents')">返回学生列表</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section class="grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
      <article class="analysis-hero-card rounded-[30px] border px-6 py-6 md:px-8">
        <div class="analysis-hero-head">
          <div>
            <div class="analysis-eyebrow">Focused Student</div>
            <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">
              {{ selectedStudent?.name || selectedStudent?.username || '未选择学员' }}
            </h2>
            <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
              当前学员训练概览。
            </p>
          </div>
          <span class="analysis-class-chip">{{ selectedClassName || '未选择班级' }}</span>
        </div>

        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="analysis-note">
            <div class="analysis-note-label">当前学员</div>
            <div class="analysis-note-value">{{ selectedStudent?.username || '未选择' }}</div>
            <div class="analysis-note-helper">当前聚焦的学生对象</div>
          </div>
          <div class="analysis-note">
            <div class="analysis-note-label">完成率</div>
            <div class="analysis-note-value">{{ solvedRate }}%</div>
            <div class="analysis-note-helper">基于当前学员训练数据计算</div>
          </div>
          <div class="analysis-note">
            <div class="analysis-note-label">薄弱维度</div>
            <div class="analysis-note-value">{{ weakDimensions[0] || '暂无' }}</div>
            <div class="analysis-note-helper">当前最需要补强的方向</div>
          </div>
        </div>
      </article>

      <div class="teacher-kpi-grid grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <article class="teacher-kpi-card teacher-kpi-card--primary">
          <div class="teacher-kpi-label">同班学生</div>
          <div class="teacher-kpi-value">{{ students.length }}</div>
          <div class="teacher-kpi-hint">当前班级可切换的学生数量</div>
        </article>
        <article class="teacher-kpi-card teacher-kpi-card--success">
          <div class="teacher-kpi-label">推荐任务</div>
          <div class="teacher-kpi-value">{{ recommendations.length }}</div>
          <div class="teacher-kpi-hint">当前可布置的补强题目数</div>
        </article>
        <article class="teacher-kpi-card teacher-kpi-card--warning">
          <div class="teacher-kpi-label">查看方式</div>
          <div class="teacher-kpi-value">学生画像</div>
          <div class="teacher-kpi-hint">当前学员分析视图</div>
        </article>
      </div>
    </section>

    <div
      v-if="error"
      class="rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 px-5 py-4 text-sm text-[var(--color-danger)]"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <section class="grid gap-6 xl:grid-cols-[0.84fr_1.16fr]">
      <div class="space-y-6">
        <SectionCard title="班级与学生切换" subtitle="先定班级，再切换当前分析的学生。">
          <div class="space-y-4">
            <div class="flex flex-wrap gap-3">
              <button
                v-for="item in classes"
                :key="item.name"
                type="button"
                class="rounded-full px-4 py-2 text-sm font-medium transition"
                :class="
                  item.name === selectedClassName
                    ? 'bg-[var(--color-primary)] text-white'
                    : 'border border-[var(--color-border-default)] bg-[var(--color-bg-base)] text-[var(--color-text-primary)] hover:border-[var(--color-primary)]/60'
                "
                @click="emit('selectClass', item.name)"
              >
                {{ item.name }} · {{ item.student_count || 0 }}
              </button>
            </div>

            <div v-if="loadingClasses || loadingStudents" class="space-y-3">
              <div
                v-for="index in 4"
                :key="index"
                class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]"
              />
            </div>

            <div v-else class="grid gap-3">
              <AppCard
                v-for="student in students"
                :key="student.id"
                as="button"
                variant="action"
                :accent="student.id === selectedStudentId ? 'primary' : 'neutral'"
                interactive
                class="text-left"
                @click="emit('selectStudent', student.id)"
              >
                <div class="flex items-center justify-between gap-3">
                  <div class="flex items-center gap-3">
                    <div
                      class="flex h-10 w-10 items-center justify-center rounded-2xl border border-primary/16 bg-primary/10 text-primary"
                    >
                      <GraduationCap class="h-4 w-4" />
                    </div>
                    <div>
                      <div class="font-medium text-text-primary">
                        {{ student.name || student.username }}
                      </div>
                      <div class="mt-1 text-sm text-text-secondary">@{{ student.username }}</div>
                    </div>
                  </div>
                  <span
                    v-if="student.id === selectedStudentId"
                    class="rounded-full bg-primary/12 px-3 py-1 text-xs font-medium text-primary"
                  >
                    当前
                  </span>
                </div>
              </AppCard>
            </div>
          </div>
        </SectionCard>

        <SectionCard title="操作入口" subtitle="从分析页返回上一层，或者直接导出报告。">
          <div class="grid gap-3">
            <AppCard
              as="button"
              variant="action"
              accent="primary"
              interactive
              class="text-left"
              @click="emit('openClassStudents')"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div
                    class="flex h-10 w-10 items-center justify-center rounded-2xl border border-primary/16 bg-primary/10 text-primary"
                  >
                    <Users class="h-4 w-4" />
                  </div>
                  <div>
                    <div class="font-medium text-text-primary">返回学生列表</div>
                    <div class="mt-1 text-sm text-text-secondary">回到当前班级，查看全部学生。</div>
                  </div>
                </div>
                <ChevronLeft class="h-4 w-4 text-primary" />
              </div>
            </AppCard>

            <AppCard
              as="button"
              variant="action"
              accent="warning"
              interactive
              class="text-left"
              @click="emit('openReportExport')"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div
                    class="flex h-10 w-10 items-center justify-center rounded-2xl border border-[var(--color-warning)]/16 bg-[var(--color-warning)]/10 text-[var(--color-warning)]"
                  >
                    <FileDown class="h-4 w-4" />
                  </div>
                  <div>
                    <div class="font-medium text-text-primary">导出班级报告</div>
                    <div class="mt-1 text-sm text-text-secondary">
                      从当前教师路径直接进入报告导出。
                    </div>
                  </div>
                </div>
                <ArrowLeftRight class="h-4 w-4 text-[var(--color-warning)]" />
              </div>
            </AppCard>
          </div>
        </SectionCard>
      </div>

      <StudentInsightPanel
        :student="selectedStudent"
        :progress="progress"
        :profile="skillProfile"
        :recommendations="recommendations"
        :timeline="timeline"
        :evidence="evidence"
        :writeup-submissions="writeupSubmissions"
        :loading="loadingDetails"
        empty-text="请先从左侧选择一名学生。"
        @open-challenge="emit('openChallenge', $event)"
      />
    </section>
  </div>
</template>

<style scoped>
.teacher-analysis-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  --color-primary: #4f46e5;
  --color-primary-hover: #4338ca;
  --color-primary-soft: rgba(79, 70, 229, 0.08);
  --color-text-primary: var(--journal-ink);
  --color-text-secondary: var(--journal-muted);
  --color-text-muted: #94a3b8;
  --color-border-default: var(--journal-border);
  --color-border-subtle: rgba(226, 232, 240, 0.74);
  --color-bg-surface: var(--journal-surface);
  --color-bg-base: #f8fafc;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

:deep(.page-header) {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

:deep(.page-header__eyebrow) {
  border: 1px solid rgba(99, 102, 241, 0.18);
  border-left: 1px solid rgba(99, 102, 241, 0.18) !important;
  border-radius: 999px;
  background: rgba(99, 102, 241, 0.06);
  padding: 0.2rem 0.72rem;
  padding-left: 0.72rem !important;
  letter-spacing: 0.2em;
  color: var(--journal-accent);
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

.analysis-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.analysis-hero-card {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.analysis-hero-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.analysis-class-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.16);
  background: rgba(99, 102, 241, 0.06);
  padding: 0.3rem 0.75rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.analysis-note {
  border-radius: 16px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface-subtle);
  padding: 0.85rem 0.95rem;
}

.analysis-note-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.analysis-note-value {
  margin-top: 0.45rem;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.analysis-note-helper {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-kpi-grid {
  align-items: stretch;
}

.teacher-kpi-card {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: var(--journal-surface-subtle);
  padding: 0.95rem 1rem;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.teacher-kpi-card--primary {
  border-top: 3px solid rgba(79, 70, 229, 0.42);
}

.teacher-kpi-card--success {
  border-top: 3px solid rgba(16, 185, 129, 0.36);
}

.teacher-kpi-card--warning {
  border-top: 3px solid rgba(245, 158, 11, 0.38);
}

.teacher-kpi-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-kpi-value {
  margin-top: 0.45rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-kpi-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}
</style>
