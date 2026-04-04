<script setup lang="ts">
import { Search } from 'lucide-vue-next'

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
  <div class="teacher-management-shell teacher-surface space-y-6">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-surface-eyebrow">Student Directory</div>
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

        <article class="teacher-brief teacher-surface-brief px-5 py-5">
          <div class="text-sm font-medium text-[var(--journal-ink)]">当前学生概况</div>
          <div class="teacher-metric-grid mt-5 grid gap-3 sm:grid-cols-3">
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--accent px-4 py-4"
            >
              <div class="teacher-metric-label">可访问班级</div>
              <div class="teacher-metric-value">{{ classes.length }}</div>
              <div class="teacher-metric-hint">当前教师可切换的班级数量</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--calm px-4 py-4"
            >
              <div class="teacher-metric-label">当前班级学生</div>
              <div class="teacher-metric-value">{{ totalStudents }}</div>
              <div class="teacher-metric-hint">当前选中班级的学生总数</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--soft px-4 py-4"
            >
              <div class="teacher-metric-label">搜索结果</div>
              <div class="teacher-metric-value">{{ filteredStudents.length }}</div>
              <div class="teacher-metric-hint">当前搜索条件下匹配的学生数量</div>
            </article>
          </div>
        </article>
      </div>

      <div class="teacher-surface-board mt-6">
        <section class="teacher-surface-section teacher-surface-filter">
          <div>
            <div class="teacher-surface-eyebrow">Student Filters</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">学生筛选</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              先选班级，再按姓名、用户名或学号定位学生。
            </p>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-[220px_1fr_1fr]">
            <label class="space-y-2">
              <span class="text-sm text-text-secondary">班级</span>
              <select
                :value="selectedClassName"
                class="w-full rounded-xl px-4 py-3 text-sm outline-none transition focus:border-[var(--journal-accent)] disabled:cursor-not-allowed disabled:opacity-60"
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
              <div class="teacher-filter-control">
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
              <div class="teacher-filter-control">
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
        </section>

        <section class="teacher-surface-section">
          <div>
            <div class="teacher-surface-eyebrow">Student List</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">学生列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              当前班级下的学生将按筛选条件实时收敛。
            </p>
          </div>

          <div v-if="loadingStudents" class="mt-5 space-y-3">
            <div
              v-for="index in 6"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
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
              class="teacher-surface-table teacher-student-table"
              empty-text="没有匹配学生"
            >
              <ElTableColumn label="姓名" min-width="220">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="teacher-student-name font-semibold">{{ row.name || row.username }}</div>
                    <div class="teacher-student-copy mt-1 text-sm">@{{ row.username }}</div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn prop="username" label="用户名" min-width="220">
                <template #default="{ row }">
                  <span class="teacher-student-copy text-sm">@{{ row.username }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="学号" min-width="180">
                <template #default="{ row }">
                  <span class="teacher-student-copy text-sm">{{ row.student_no || '未设置' }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="解题数" width="120" align="center">
                <template #default="{ row }">
                  <span class="teacher-student-stat text-sm font-medium">{{
                    row.solved_count ?? 0
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="得分" width="120" align="center">
                <template #default="{ row }">
                  <span class="teacher-student-stat text-sm font-medium">{{
                    row.total_score ?? 0
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="薄弱项" min-width="160">
                <template #default="{ row }">
                  <span class="teacher-student-copy text-sm">{{ row.weak_dimension || '暂无' }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="操作" width="180" align="right">
                <template #default="{ row }">
                  <ElButton class="teacher-student-action" @click="emit('openStudent', row.id)"
                    >查看学员分析</ElButton
                  >
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        </section>
      </div>
    </section>

    <div v-if="error" class="teacher-surface-error">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-filter-control {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid var(--journal-border);
  border-radius: 0.9rem;
  background: var(--journal-surface);
  padding: 0.75rem 1rem;
  transition:
    border-color 0.18s ease,
    background 0.18s ease;
}

.teacher-filter-control:focus-within {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 5%, var(--journal-surface));
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
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-btn:hover {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 12px 24px var(--color-shadow-soft);
}

.teacher-btn--primary:hover {
  border-color: transparent;
  background: var(--journal-accent-strong);
  color: #fff;
}

.teacher-metric-grid {
  align-items: stretch;
}

.teacher-metric-card {
  min-height: 100%;
  border-top: 3px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
}

.teacher-metric-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-metric-card--accent {
  border-top-color: color-mix(in srgb, var(--journal-accent) 28%, var(--journal-border));
}

.teacher-metric-card--calm {
  border-top-color: color-mix(in srgb, var(--color-success) 22%, var(--journal-border));
}

.teacher-metric-card--soft {
  border-top-color: color-mix(in srgb, var(--color-warning) 22%, var(--journal-border));
}

.teacher-metric-value {
  margin-top: 0.45rem;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-metric-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-student-name {
  color: color-mix(in srgb, var(--journal-ink) 88%, var(--journal-muted));
}

.teacher-student-copy {
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-student-stat {
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

:deep(.teacher-student-action.el-button) {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, var(--journal-border));
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  border-radius: 0.9rem;
  box-shadow: none;
}

:deep(.teacher-student-action.el-button:hover),
:deep(.teacher-student-action.el-button:focus-visible) {
  border-color: color-mix(in srgb, var(--journal-accent) 38%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface));
  color: var(--journal-accent);
}
</style>
