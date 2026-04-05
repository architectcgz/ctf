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
  <div class="teacher-management-shell teacher-surface">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <header class="teacher-header">
        <div class="teacher-header__main">
          <div class="teacher-surface-eyebrow journal-eyebrow">Student Directory</div>
          <h2 class="teacher-title">学生管理</h2>
          <p class="teacher-copy">按班级筛选、搜索并进入学员分析。</p>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-surface-btn"
              @click="emit('openClassManagement')"
            >
              班级管理
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
            <div class="teacher-badge-label">当前班级学生</div>
            <div class="teacher-badge-value">{{ totalStudents }}</div>
            <div class="teacher-badge-hint">当前选中班级的学生总数</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">搜索结果</div>
            <div class="teacher-badge-value">{{ filteredStudents.length }}</div>
            <div class="teacher-badge-hint">当前搜索条件下匹配的学生数量</div>
          </article>
        </div>
      </header>

      <div class="teacher-hero-divider" />

      <div class="teacher-surface-board">
        <section class="teacher-surface-section teacher-filter-panel">
          <div class="teacher-section-head">
            <div>
              <div class="teacher-surface-eyebrow journal-eyebrow">Student Filters</div>
              <h3 class="teacher-section-title">学生筛选</h3>
            </div>
          </div>

          <div class="teacher-filter-grid">
            <label class="teacher-field">
              <span class="teacher-field-label">班级</span>
              <select
                :value="selectedClassName"
                class="teacher-field-control"
                :disabled="loadingClasses"
                @change="emit('selectClass', ($event.target as HTMLSelectElement).value)"
              >
                <option v-for="item in classes" :key="item.name" :value="item.name">
                  {{ item.name }} · {{ item.student_count || 0 }}
                </option>
              </select>
            </label>

            <label class="teacher-field">
              <span class="teacher-field-label">搜索姓名或用户名</span>
              <div class="teacher-field-control teacher-filter-control">
                <Search class="h-4 w-4 text-text-muted" />
                <input
                  :value="searchQuery"
                  type="text"
                  placeholder="搜索姓名或用户名"
                  class="teacher-input"
                  @input="emit('updateSearchQuery', ($event.target as HTMLInputElement).value)"
                />
              </div>
            </label>

            <label class="teacher-field">
              <span class="teacher-field-label">按学号查询</span>
              <div class="teacher-field-control teacher-filter-control">
                <Search class="h-4 w-4 text-text-muted" />
                <input
                  :value="studentNoQuery"
                  type="text"
                  placeholder="输入学号精确查询"
                  class="teacher-input"
                  @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
                />
              </div>
            </label>
          </div>
        </section>

        <section class="teacher-surface-section">
          <div class="teacher-section-head">
            <div>
              <div class="teacher-surface-eyebrow journal-eyebrow">Student List</div>
              <h3 class="teacher-section-title">学生列表</h3>
            </div>
            <div class="teacher-section-meta">共 {{ filteredStudents.length }} 名学生</div>
          </div>

          <div v-if="loadingStudents" class="teacher-skeleton-list">
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

          <div v-else class="mt-5 teacher-table-shell">
            <ElTable
              :data="filteredStudents"
              row-key="id"
              class="teacher-surface-table teacher-student-table"
              empty-text="没有匹配学生"
            >
              <ElTableColumn label="姓名" min-width="220">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="teacher-student-name">{{ row.name || row.username }}</div>
                    <div class="teacher-student-copy">@{{ row.username }}</div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn prop="username" label="用户名" min-width="220">
                <template #default="{ row }">
                  <span class="teacher-student-copy">@{{ row.username }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="学号" min-width="180">
                <template #default="{ row }">
                  <span class="teacher-student-copy">{{ row.student_no || '未设置' }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="解题数" width="120" align="center">
                <template #default="{ row }">
                  <span class="teacher-student-stat">{{ row.solved_count ?? 0 }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="得分" width="120" align="center">
                <template #default="{ row }">
                  <span class="teacher-student-stat">{{ row.total_score ?? 0 }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="薄弱项" min-width="160">
                <template #default="{ row }">
                  <span class="teacher-student-copy">{{ row.weak_dimension || '暂无' }}</span>
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
                  </button>
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
  --journal-accent: #2563eb;
  --journal-accent-strong: #1d4ed8;
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
  grid-template-columns: repeat(3, minmax(0, 1fr));
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
  margin: 1.35rem 0 0.2rem;
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-filter-panel {
  padding-top: 1.3rem;
}

.teacher-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.85rem;
}

.teacher-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-section-meta {
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.teacher-filter-grid {
  margin-top: 1rem;
  display: grid;
  gap: 1rem;
  grid-template-columns: 220px minmax(0, 1fr) minmax(0, 1fr);
}

.teacher-field {
  display: grid;
  gap: 0.45rem;
}

.teacher-field-label {
  font-size: 0.84rem;
  color: var(--journal-muted);
}

.teacher-field-control {
  width: 100%;
  min-height: 2.9rem;
  border: 1px solid var(--teacher-control-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.72rem 0.95rem;
  color: var(--journal-ink);
  transition:
    border-color 0.18s ease,
    background 0.18s ease;
}

.teacher-field-control:focus-within,
.teacher-field-control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 5%, var(--journal-surface));
}

.teacher-filter-control {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.teacher-input {
  width: 100%;
  background: transparent;
  color: var(--journal-ink);
  outline: none;
}

.teacher-input::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 76%, transparent);
}

.teacher-skeleton-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.75rem;
}

.teacher-table-shell {
  border: 1px solid var(--teacher-card-border);
  border-radius: 18px;
}

.teacher-student-name {
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 88%, var(--journal-muted));
}

.teacher-student-copy {
  font-size: 0.84rem;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-student-stat {
  font-size: 0.95rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

.teacher-row-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
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

@media (max-width: 1080px) {
  .teacher-badge-grid,
  .teacher-filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
