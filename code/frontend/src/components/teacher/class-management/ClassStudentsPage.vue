<script setup lang="ts">
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

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
  <div class="teacher-management-shell teacher-surface space-y-6">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-eyebrow-row">
            <div class="teacher-surface-eyebrow">Class Students</div>
            <span class="teacher-class-chip teacher-surface-chip">{{ selectedClassName || '未选择班级' }}</span>
          </div>

          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            {{ selectedClassName ? `${selectedClassName} · 学生列表` : '班级学生' }}
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            查看当前班级学生名单，并继续进入学员分析。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="teacher-btn" @click="emit('openClassManagement')">
              返回班级管理
            </button>
            <button type="button" class="teacher-btn" @click="emit('openDashboard')">
              教学概览
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>
        </div>

        <article class="teacher-brief teacher-surface-brief px-5 py-5">
          <div class="text-sm font-medium text-[var(--journal-ink)]">当前班级概况</div>
          <div class="teacher-badge-grid mt-5">
            <div class="teacher-badge-card teacher-surface-metric">
              <div class="teacher-badge-label">可访问班级</div>
              <div class="teacher-badge-value">{{ classes.length }}</div>
            </div>
            <div class="teacher-badge-card teacher-surface-metric">
              <div class="teacher-badge-label">班级人数</div>
              <div class="teacher-badge-value">
                {{ props.summary?.student_count ?? students.length }}
              </div>
            </div>
            <div class="teacher-badge-card teacher-surface-metric">
              <div class="teacher-badge-label">平均解题</div>
              <div class="teacher-badge-value">{{ averageSolvedText }}</div>
            </div>
            <div class="teacher-badge-card teacher-surface-metric">
              <div class="teacher-badge-label">近 7 天活跃率</div>
              <div class="teacher-badge-value">{{ activeRateText }}</div>
            </div>
          </div>

          <div class="teacher-tip-block mt-5">
            <div class="teacher-tip-title">当前筛查重点</div>
            <ul class="teacher-tip-list mt-3">
              <li class="teacher-tip-item">
                <span class="teacher-tip-index">1</span>
                <span>先从整体趋势和复盘结论判断班级节奏。</span>
              </li>
              <li class="teacher-tip-item">
                <span class="teacher-tip-index">2</span>
                <span>再结合学生名单按学号或薄弱项继续下钻。</span>
              </li>
              <li class="teacher-tip-item">
                <span class="teacher-tip-index">3</span>
                <span>确认重点对象后直接进入学员分析页。</span>
              </li>
            </ul>
          </div>
        </article>
      </div>

      <div class="teacher-metric-grid mt-6">
        <article class="teacher-surface-metric teacher-kpi-card teacher-kpi-card--primary">
          <div class="teacher-kpi-label">近 7 天训练事件</div>
          <div class="teacher-kpi-value">{{ props.summary?.recent_event_count ?? '--' }}</div>
          <div class="teacher-kpi-hint">提交、实例启动与销毁等动作总数</div>
        </article>
        <article class="teacher-surface-metric teacher-kpi-card teacher-kpi-card--success">
          <div class="teacher-kpi-label">学生记录</div>
          <div class="teacher-kpi-value">{{ students.length }}</div>
          <div class="teacher-kpi-hint">当前列表内可直接进入分析的学生数量</div>
        </article>
      </div>

      <div class="teacher-board teacher-surface-board">
        <div v-if="error" class="teacher-error-card" role="alert" aria-live="polite">
          {{ error }}
          <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">
            重试
          </button>
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

        <section class="teacher-student-list-section teacher-surface-section">
          <div class="teacher-section-head">
            <div>
              <div class="teacher-surface-eyebrow">Students</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">学生名单</h3>
              <p class="mt-2 text-sm leading-7 text-[var(--journal-muted)]">
                选择学生后进入学员分析。
              </p>
            </div>
            <button type="button" class="teacher-inline-link" @click="emit('openClassManagement')">
              <ChevronLeft class="h-4 w-4" />
              返回班级列表
            </button>
          </div>

          <div class="teacher-student-toolbar">
            <div class="text-sm text-text-secondary">共 {{ students.length }} 名学生</div>

            <label class="teacher-search-field">
              <span class="text-sm text-text-secondary">按学号查询</span>
              <input
                :value="studentNoQuery"
                type="text"
                placeholder="输入学号后实时查询"
                class="teacher-search-input"
                @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
              />
            </label>
          </div>

          <div v-if="loadingStudents" class="space-y-3">
            <div
              v-for="index in 6"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]"
            />
          </div>

          <AppEmpty
            v-else-if="students.length === 0"
            icon="Users"
            title="当前班级暂无学生"
            description="该班级下还没有可用学生记录。"
          />

          <div v-else class="teacher-table-shell teacher-surface-table">
            <ElTable
              :data="students"
              row-key="id"
              class="teacher-student-table"
              empty-text="当前班级暂无学生"
            >
              <ElTableColumn label="姓名" min-width="220">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="font-semibold text-text-primary">
                      {{ row.name || row.username }}
                    </div>
                    <div class="mt-1 text-sm text-text-secondary">@{{ row.username }}</div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn prop="username" label="用户名" min-width="220">
                <template #default="{ row }">
                  <span class="text-sm text-text-secondary">@{{ row.username }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="学号" min-width="180">
                <template #default="{ row }">
                  <span class="text-sm text-text-secondary">{{ row.student_no || '未设置' }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="解题数" width="120" align="center">
                <template #default="{ row }">
                  <span class="text-sm font-medium text-text-primary">{{
                    row.solved_count ?? 0
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="得分" width="120" align="center">
                <template #default="{ row }">
                  <span class="text-sm font-medium text-text-primary">{{
                    row.total_score ?? 0
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="薄弱项" min-width="160">
                <template #default="{ row }">
                  <span class="text-sm text-text-secondary">{{
                    row.weak_dimension || '暂无'
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="操作" width="180" align="right">
                <template #default="{ row }">
                  <ElButton type="primary" plain @click="emit('openStudent', row.id)">
                    查看学员分析
                    <ChevronRight class="ml-1 h-4 w-4" />
                  </ElButton>
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
:deep(.teacher-student-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-expanded-cell-bg-color: transparent;
  --el-table-header-bg-color: var(--journal-surface);
  --el-table-border-color: var(--journal-border);
  --el-table-row-hover-bg-color: rgba(99, 102, 241, 0.06);
  --el-table-text-color: var(--journal-ink);
  --el-table-header-text-color: var(--journal-muted);
}

:deep(.teacher-student-table th.el-table__cell) {
  background: var(--journal-surface);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

:deep(.teacher-student-table td.el-table__cell),
:deep(.teacher-student-table th.el-table__cell) {
  border-bottom-color: var(--journal-border);
}

:deep(.teacher-student-table .el-table__inner-wrapper::before) {
  display: none;
}

.teacher-management-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-hero {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 14%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-base))
    );
  overflow: hidden;
}

.teacher-brief {
  overflow: hidden;
}

.teacher-eyebrow-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.5rem;
  border-radius: 0.9rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.55rem 1.1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    background 0.18s ease;
}

.teacher-btn:hover {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 7%, var(--journal-surface));
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
}

.teacher-btn--primary:hover {
  background: #4338ca;
  border-color: transparent;
}

.teacher-badge-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.teacher-badge-card {
  padding: 0.9rem 0.95rem;
}

.teacher-badge-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-badge-value {
  margin-top: 0.55rem;
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-tip-block {
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
  padding-top: 1rem;
}

.teacher-tip-title {
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-tip-list {
  display: grid;
  gap: 0.6rem;
}

.teacher-tip-item {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  font-size: 0.83rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.teacher-tip-index {
  display: inline-flex;
  min-width: 1.2rem;
  justify-content: center;
  margin-top: 0.04rem;
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
    monospace;
  font-size: 0.76rem;
  font-weight: 700;
  color: var(--journal-accent);
}

.teacher-kpi-grid {
  align-items: stretch;
}

.teacher-metric-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.teacher-kpi-card {
  padding: 0.95rem 1rem;
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

.teacher-board {
  margin-top: 1.5rem;
  padding-top: 1.25rem;
}

.teacher-board > * + * {
  margin-top: 1.25rem;
  padding-top: 1.25rem;
}

.teacher-error-card {
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--color-danger) 22%, var(--journal-border));
  background: color-mix(in srgb, var(--color-danger) 6%, transparent);
  padding: 1rem 1rem 1.1rem;
  font-size: 0.875rem;
  color: var(--color-danger);
}

.teacher-anchor-section {
  scroll-margin-top: 84px;
}

.teacher-student-list-section {
  --panel-border: var(--journal-border);
  --panel-surface: var(--journal-surface);
  --panel-surface-subtle: var(--journal-surface-subtle);
}

.teacher-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
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
  cursor: pointer;
}

.teacher-student-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

.teacher-search-field {
  display: grid;
  gap: 0.5rem;
  min-width: min(100%, 20rem);
}

.teacher-search-input {
  width: 100%;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.72rem 0.95rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.teacher-search-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.teacher-table-shell {
  padding: 0;
}

@media (max-width: 767px) {
  .teacher-badge-grid,
  .teacher-metric-grid {
    grid-template-columns: 1fr;
  }
}
</style>
