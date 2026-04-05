<script setup lang="ts">
import { ArrowRight, ChevronLeft } from 'lucide-vue-next'
import { computed } from 'vue'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherInterventionPanel from '@/components/teacher/TeacherInterventionPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  students: TeacherStudentItem[]
  review: TeacherClassReviewData | null
  summary: TeacherClassSummaryData | null
  trend: TeacherClassTrendData | null
  studentNoQuery: string
  loadingStudents: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openDashboard: []
  openReportExport: []
  updateStudentNoQuery: [value: string]
  openStudent: [studentId: string]
}>()

const averageSolvedText = computed(() => {
  if (!props.summary) return '--'
  return props.summary.average_solved.toFixed(1)
})

const activeRateText = computed(() => {
  if (!props.summary) return '--'
  return `${Math.round(props.summary.active_rate)}%`
})
</script>

<template>
  <div class="teacher-management-shell teacher-surface">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <header class="teacher-header">
        <div class="teacher-header__main">
          <div class="teacher-eyebrow-row">
            <div class="teacher-surface-eyebrow journal-eyebrow">Class Students</div>
            <span class="teacher-class-chip teacher-surface-chip">{{
              selectedClassName || '未选择班级'
            }}</span>
          </div>

          <h2 class="teacher-title">
            {{ selectedClassName ? `${selectedClassName} · 学生列表` : '班级学生' }}
          </h2>
          <p class="teacher-copy">查看当前班级学生名单，并继续进入学员分析。</p>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-surface-btn"
              @click="emit('openClassManagement')"
            >
              返回班级管理
            </button>
            <button
              type="button"
              class="teacher-btn teacher-surface-btn"
              @click="emit('openDashboard')"
            >
              教学概览
            </button>
            <button
              type="button"
              class="teacher-btn teacher-surface-btn teacher-btn--primary"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>
        </div>

        <div class="teacher-badge-grid">
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">可访问班级</div>
            <div class="teacher-badge-value">{{ classes.length }}</div>
            <div class="teacher-badge-hint">当前教师可切换的班级数量</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">班级人数</div>
            <div class="teacher-badge-value">{{ props.summary?.student_count ?? students.length }}</div>
            <div class="teacher-badge-hint">当前班级纳入统计的学生数量</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">平均解题</div>
            <div class="teacher-badge-value">{{ averageSolvedText }}</div>
            <div class="teacher-badge-hint">班级当前平均完成情况</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">近 7 天活跃率</div>
            <div class="teacher-badge-value">{{ activeRateText }}</div>
            <div class="teacher-badge-hint">当前班级近 7 天训练参与情况</div>
          </article>
        </div>
      </header>

      <div class="teacher-hero-divider" />

      <div class="teacher-tip-block">
        <div class="teacher-tip-label">当前焦点</div>
        <div class="teacher-tip-copy">
          先看班级趋势和复盘结论，再从学生名单进入单个学员分析。
        </div>
      </div>

      <div class="teacher-summary-grid">
        <article class="teacher-summary-card">
          <div class="teacher-summary-label">近 7 天训练事件</div>
          <div class="teacher-summary-value">{{ props.summary?.recent_event_count ?? '--' }}</div>
          <div class="teacher-summary-hint">提交、实例启动与销毁等动作总数</div>
        </article>
        <article class="teacher-summary-card">
          <div class="teacher-summary-label">当前学生记录</div>
          <div class="teacher-summary-value">{{ students.length }}</div>
          <div class="teacher-summary-hint">当前列表内可直接进入分析的学生数量</div>
        </article>
      </div>

      <div class="teacher-surface-board">
        <div v-if="error" class="teacher-surface-error" role="alert" aria-live="polite">
          {{ error }}
          <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
        </div>

        <section id="teacher-trend" class="teacher-anchor-section">
          <TeacherClassTrendPanel
            :trend="trend"
            title="班级近 7 天训练趋势"
            subtitle="先看整体节奏，再下钻到具体学生。"
          />
        </section>

        <section id="teacher-review" class="teacher-anchor-section">
          <TeacherClassReviewPanel :review="review" :class-name="selectedClassName" />
        </section>

        <section id="teacher-insight" class="teacher-anchor-section">
          <TeacherClassInsightsPanel :students="students" :class-name="selectedClassName" />
        </section>

        <section id="teacher-intervention" class="teacher-anchor-section">
          <TeacherInterventionPanel :students="students" :class-name="selectedClassName" />
        </section>

        <section class="teacher-surface-section teacher-student-list-section">
          <div class="teacher-section-head">
            <div>
              <div class="teacher-surface-eyebrow journal-eyebrow">Students</div>
              <h3 class="teacher-section-title">学生列表</h3>
              <p class="teacher-section-copy">选择学生后进入学员分析。</p>
            </div>
            <button type="button" class="teacher-inline-link" @click="emit('openClassManagement')">
              <ChevronLeft class="h-4 w-4" />
              返回班级列表
            </button>
          </div>

          <div class="teacher-student-toolbar">
            <div class="teacher-section-meta">共 {{ students.length }} 名学生</div>
            <label class="teacher-search-field">
              <span class="teacher-field-label">按学号查询</span>
              <input
                :value="studentNoQuery"
                type="text"
                placeholder="输入学号后实时查询"
                class="teacher-search-input"
                @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
              />
            </label>
          </div>

          <div v-if="loadingStudents" class="teacher-skeleton-list">
            <div
              v-for="index in 6"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
            />
          </div>

          <AppEmpty
            v-else-if="students.length === 0"
            icon="Users"
            title="当前班级暂无学生"
            description="该班级下还没有可用学生记录。"
          />

          <div v-else class="teacher-table-shell">
            <ElTable
              :data="students"
              row-key="id"
              class="teacher-surface-table teacher-student-table"
              empty-text="当前班级暂无学生"
            >
              <ElTableColumn label="姓名" min-width="220">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="teacher-row-title">{{ row.name || row.username }}</div>
                    <div class="teacher-row-copy">@{{ row.username }}</div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn prop="username" label="用户名" min-width="220">
                <template #default="{ row }">
                  <span class="teacher-row-copy">@{{ row.username }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="学号" min-width="180">
                <template #default="{ row }">
                  <span class="teacher-row-copy">{{ row.student_no || '未设置' }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="解题数" width="120" align="center">
                <template #default="{ row }">
                  <span class="teacher-row-stat">{{ row.solved_count ?? 0 }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="得分" width="120" align="center">
                <template #default="{ row }">
                  <span class="teacher-row-stat">{{ row.total_score ?? 0 }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="薄弱项" min-width="160">
                <template #default="{ row }">
                  <span class="teacher-row-copy">{{ row.weak_dimension || '暂无' }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="操作" width="180" align="right">
                <template #default="{ row }">
                  <button
                    type="button"
                    class="teacher-row-btn"
                    @click="emit('openStudent', row.id)"
                  >
                    查看学员分析
                    <ArrowRight class="h-4 w-4" />
                  </button>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        </section>
      </div>
    </section>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #2563eb;
  --journal-accent-strong: #1d4ed8;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-hero {
  border-color: var(--teacher-card-border);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 12%, transparent),
      transparent 18rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-base))
    );
}

.journal-eyebrow {
  letter-spacing: 0.08em;
}

.journal-brief,
.journal-metric {
  border-radius: 18px;
}

.teacher-header {
  display: grid;
  gap: 1.25rem;
}

.teacher-header__main {
  max-width: 42rem;
}

.teacher-eyebrow-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.teacher-title {
  margin-top: 0.85rem;
  font-size: clamp(2rem, 2vw, 2.45rem);
  font-weight: 700;
  line-height: 1.08;
  color: var(--journal-ink);
}

.teacher-copy {
  margin-top: 0.7rem;
  max-width: 42rem;
  font-size: 0.92rem;
  line-height: 1.72;
  color: var(--journal-muted);
}

.teacher-actions {
  margin-top: 1.3rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.75rem;
  border-radius: 999px;
  border: 1px solid var(--teacher-control-border);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.62rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-btn:hover,
.teacher-btn:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.teacher-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink));
}

.teacher-btn--primary:hover,
.teacher-btn--primary:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface));
}

.teacher-badge-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.teacher-badge-card {
  min-height: 100%;
  border: 1px solid var(--teacher-card-border);
  padding: 1rem 1.05rem 1.05rem;
}

.teacher-badge-label {
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-badge-value {
  margin-top: 0.55rem;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-badge-hint {
  margin-top: 0.5rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-hero-divider {
  margin: 1.35rem 0 1.15rem;
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-tip-block {
  display: grid;
  gap: 0.35rem;
  border-top: 1px dashed var(--teacher-divider);
  padding-top: 1rem;
}

.teacher-tip-label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.teacher-tip-copy {
  font-size: 0.86rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-summary-grid {
  margin-top: 1.25rem;
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.teacher-summary-card {
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 96%, transparent);
  padding-top: 0.95rem;
}

.teacher-summary-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-summary-value {
  margin-top: 0.45rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-summary-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-anchor-section {
  scroll-margin-top: 84px;
}

.teacher-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.teacher-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-section-copy {
  margin-top: 0.45rem;
  font-size: 0.86rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-section-meta {
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.teacher-inline-link {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border: 0;
  background: transparent;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-accent);
}

.teacher-student-toolbar {
  margin: 1rem 0;
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
}

.teacher-search-field {
  display: grid;
  gap: 0.45rem;
  min-width: min(100%, 20rem);
}

.teacher-field-label {
  font-size: 0.84rem;
  color: var(--journal-muted);
}

.teacher-search-input {
  width: 100%;
  min-height: 2.9rem;
  border-radius: 1rem;
  border: 1px solid var(--teacher-control-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.72rem 0.95rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
}

.teacher-search-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 5%, var(--journal-surface));
}

.teacher-skeleton-list {
  display: grid;
  gap: 0.75rem;
}

.teacher-table-shell {
  border: 1px solid var(--teacher-card-border);
  border-radius: 18px;
}

.teacher-row-title {
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 88%, var(--journal-muted));
}

.teacher-row-copy {
  font-size: 0.84rem;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-row-stat {
  font-size: 0.95rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

.teacher-row-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.42rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, var(--teacher-control-border));
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  padding: 0.58rem 0.95rem;
  font-size: 0.84rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-row-btn:hover,
.teacher-row-btn:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 38%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface));
  color: var(--journal-accent);
}

:deep(.teacher-student-table.el-table),
:deep(.teacher-student-table .el-table__inner-wrapper),
:deep(.teacher-student-table .el-scrollbar),
:deep(.teacher-student-table .el-scrollbar__view),
:deep(.teacher-student-table .el-table__body-wrapper),
:deep(.teacher-student-table .el-table__header-wrapper),
:deep(.teacher-student-table .el-table__empty-block) {
  background: var(--journal-surface);
}

@media (max-width: 1100px) {
  .teacher-badge-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .teacher-badge-grid,
  .teacher-summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
