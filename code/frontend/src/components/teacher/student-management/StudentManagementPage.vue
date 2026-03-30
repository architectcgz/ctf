<script setup lang="ts">
import { Search, Users } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  searchQuery: string
  studentNoQuery: string
  filteredStudents: TeacherStudentItem[]
  totalStudents: number
  loadingClasses: boolean
  loadingStudents: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openReportExport: []
  updateSearchQuery: [value: string]
  updateStudentNoQuery: [value: string]
  selectClass: [className: string]
  openStudent: [studentId: string]
}>()
</script>

<template>
  <div class="teacher-management-shell space-y-6">
    <section class="teacher-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-eyebrow">Student Directory</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            学生管理
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            按班级筛选、搜索并进入学员分析。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="teacher-btn" @click="emit('openClassManagement')">
              班级管理
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

        <article class="teacher-brief rounded-[24px] border px-5 py-5">
          <div class="teacher-brief-title">当前学生概况</div>
          <div class="teacher-kpi-grid mt-5 grid gap-3 sm:grid-cols-3">
            <article class="teacher-kpi-card teacher-kpi-card--primary">
              <div class="teacher-kpi-label">可访问班级</div>
              <div class="teacher-kpi-value">{{ classes.length }}</div>
              <div class="teacher-kpi-hint">当前教师可切换的班级数量</div>
            </article>
            <article class="teacher-kpi-card teacher-kpi-card--success">
              <div class="teacher-kpi-label">当前班级学生</div>
              <div class="teacher-kpi-value">{{ totalStudents }}</div>
              <div class="teacher-kpi-hint">当前选中班级的学生总数</div>
            </article>
            <article class="teacher-kpi-card teacher-kpi-card--warning">
              <div class="teacher-kpi-label">搜索结果</div>
              <div class="teacher-kpi-value">{{ filteredStudents.length }}</div>
              <div class="teacher-kpi-hint">当前搜索条件下匹配的学生数量</div>
            </article>
          </div>
        </article>
      </div>

      <div class="teacher-hero-divider" />

      <div class="teacher-hero-section">
        <div class="teacher-hero-section-head">
          <div>
            <div class="teacher-eyebrow teacher-eyebrow--soft">Student Filters</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">学生筛选与列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              先选班级，再按姓名、用户名或学号定位学生。
            </p>
          </div>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-[220px_1fr_1fr]">
          <label class="space-y-2">
            <span class="text-sm text-text-secondary">班级</span>
            <select
              :value="selectedClassName"
              class="teacher-filter-field w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-text-primary outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="loadingClasses"
              @change="emit('selectClass', ($event.target as HTMLSelectElement).value)"
            >
              <option v-for="item in classes" :key="item.name" :value="item.name">
                {{ item.name }} · {{ item.student_count || 0 }}
              </option>
            </select>
          </label>

          <label class="space-y-2">
            <span class="text-sm text-text-secondary">搜索姓名或用户名</span>
            <div
              class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
            >
              <Search class="h-4 w-4 text-text-muted" />
              <input
                :value="searchQuery"
                type="text"
                placeholder="搜索姓名或用户名"
                class="w-full bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
                @input="emit('updateSearchQuery', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </label>

          <label class="space-y-2">
            <span class="text-sm text-text-secondary">按学号查询</span>
            <div
              class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
            >
              <Search class="h-4 w-4 text-text-muted" />
              <input
                :value="studentNoQuery"
                type="text"
                placeholder="输入学号精确查询"
                class="w-full bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
                @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </label>
        </div>

        <div class="teacher-hero-divider teacher-hero-divider--inner" />

        <div v-if="loadingStudents" class="space-y-3">
          <div
            v-for="index in 6"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]"
          />
        </div>

        <AppEmpty
          v-else-if="filteredStudents.length === 0"
          class="mt-5"
          icon="Users"
          title="没有匹配学生"
          description="调整搜索词或切换班级后再试。"
        />

        <div v-else class="mt-5">
          <ElTable
            :data="filteredStudents"
            row-key="id"
            class="teacher-student-table"
            empty-text="没有匹配学生"
          >
            <ElTableColumn label="姓名" min-width="220">
              <template #default="{ row }">
                <div class="py-1">
                  <div class="font-semibold text-text-primary">{{ row.name || row.username }}</div>
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
                <span class="text-sm text-text-secondary">{{ row.weak_dimension || '暂无' }}</span>
              </template>
            </ElTableColumn>

            <ElTableColumn label="操作" width="180" align="right">
              <template #default="{ row }">
                <ElButton type="primary" plain @click="emit('openStudent', row.id)"
                  >查看学员分析</ElButton
                >
              </template>
            </ElTableColumn>
          </ElTable>
        </div>
      </div>
    </section>

    <div
      v-if="error"
      class="rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 px-5 py-4 text-sm text-[var(--color-danger)]"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>
  </div>
</template>

<style scoped>
:deep(.teacher-filter-field) {
  color: var(--journal-ink);
  border-color: var(--journal-border) !important;
  background: var(--journal-surface) !important;
}

:deep(.teacher-filter-field option) {
  background-color: var(--journal-surface);
  color: var(--journal-ink);
}

:deep(.teacher-filter-field select),
:deep(.teacher-filter-field input) {
  color: var(--journal-ink);
}

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
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  --color-primary: #4f46e5;
  --color-primary-hover: #4338ca;
  --color-text-primary: var(--journal-ink);
  --color-text-secondary: var(--journal-muted);
  --color-text-muted: #94a3b8;
  --color-border-default: var(--journal-border);
  --color-border-subtle: rgba(226, 232, 240, 0.74);
  --color-bg-surface: var(--journal-surface);
  --color-bg-base: #f8fafc;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.teacher-eyebrow--soft {
  opacity: 0.88;
}

.teacher-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.teacher-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.teacher-brief-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-hero-divider {
  margin-top: 1.5rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.teacher-hero-divider--inner {
  margin-top: 1.25rem;
}

.teacher-hero-section {
  margin-top: 1.5rem;
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
  background: rgba(99, 102, 241, 0.06);
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 12px 24px rgba(79, 70, 229, 0.18);
}

.teacher-btn--primary:hover {
  border-color: transparent;
  background: var(--color-primary-hover);
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
