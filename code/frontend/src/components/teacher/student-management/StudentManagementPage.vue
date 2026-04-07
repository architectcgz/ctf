<script setup lang="ts">
import { ArrowRight, Search } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  searchQuery: string
  studentNoQuery: string
  filteredStudents: TeacherStudentItem[]
  filteredTotal: number
  totalStudents: number
  page: number
  totalPages: number
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
  changePage: [page: number]
  openStudent: [studentId: string]
}>()
</script>

<template>
  <div class="teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
    >
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading">
            <div class="teacher-surface-eyebrow journal-eyebrow">Student Directory</div>
            <h1 class="teacher-title">学生管理</h1>
            <p class="teacher-copy">按班级筛选、搜索并进入学员分析。</p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('openClassManagement')"
            >
              班级管理
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>
        </header>

        <section class="teacher-summary">
          <div class="teacher-summary-title">
            <span>Directory Snapshot</span>
          </div>
          <div class="teacher-summary-grid">
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">可访问班级</div>
              <div class="teacher-summary-value">{{ classes.length }}</div>
              <div class="teacher-summary-helper">当前教师可切换的班级数量</div>
            </div>
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">当前班级学生</div>
              <div class="teacher-summary-value">{{ totalStudents }}</div>
              <div class="teacher-summary-helper">当前选中班级的学生总数</div>
            </div>
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">搜索结果</div>
              <div class="teacher-summary-value">{{ filteredStudents.length }}</div>
              <div class="teacher-summary-helper">当前搜索条件下匹配的学生数量</div>
            </div>
          </div>
        </section>

        <section class="teacher-controls">
          <div class="teacher-controls-bar">
            <div class="teacher-controls-heading">
              <div class="teacher-surface-eyebrow journal-eyebrow">Student Filters</div>
              <h3 class="teacher-controls-title">学生筛选</h3>
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
                <option value="">全部班级</option>
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

        <div class="teacher-hero-divider" />

        <div v-if="loadingStudents" class="teacher-skeleton-list">
          <div
            v-for="index in 6"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
          />
        </div>

        <AppEmpty
          v-else-if="filteredStudents.length === 0"
          class="teacher-empty-state"
          icon="Users"
          title="没有匹配学生"
          description="调整搜索词或切换班级后再试。"
        />

        <section v-else class="teacher-directory" aria-label="学生目录">
          <div class="teacher-directory-top">
            <h3 class="teacher-directory-title">学生目录</h3>
            <div class="teacher-directory-meta">共 {{ filteredTotal }} 名学生</div>
          </div>

          <div class="teacher-directory-head">
            <span class="teacher-directory-head-cell teacher-directory-head-cell-student-no"
              >学号</span
            >
            <span class="teacher-directory-head-cell teacher-directory-head-cell-name"
              >学生名称</span
            >
            <span class="teacher-directory-head-cell teacher-directory-head-cell-alias">昵称</span>
            <span>薄弱项</span>
            <span>数据</span>
            <span>操作</span>
          </div>

          <button
            v-for="student in filteredStudents"
            :key="student.id"
            type="button"
            class="teacher-directory-row"
            :aria-label="`${student.name || student.username}，${student.solved_count ?? 0} 题，${student.total_score ?? 0} 分，查看学员分析`"
            @click="emit('openStudent', student.id)"
          >
            <div class="teacher-directory-cell teacher-directory-cell-student-no">
              {{ student.student_no || '未设置学号' }}
            </div>

            <div class="teacher-directory-cell teacher-directory-cell-name">
              <h4 class="teacher-directory-row-title" :title="student.name || '未设置姓名'">
                {{ student.name || '未设置姓名' }}
              </h4>
            </div>

            <div class="teacher-directory-cell teacher-directory-cell-alias">
              <div class="teacher-directory-row-points" :title="`@${student.username}`">
                @{{ student.username }}
              </div>
            </div>

            <div class="teacher-directory-row-tags">
              <span class="teacher-directory-chip teacher-directory-chip-muted">
                {{ student.weak_dimension || '暂无薄弱项' }}
              </span>
            </div>

            <div class="teacher-directory-row-metrics">
              <span>{{ student.solved_count ?? 0 }} 题</span>
              <span>{{ student.total_score ?? 0 }} 分</span>
            </div>

            <div class="teacher-directory-row-cta">
              <span>查看学员分析</span>
              <ArrowRight class="h-4 w-4" />
            </div>
          </button>

          <div
            v-if="filteredTotal > 0"
            class="teacher-directory-pagination workspace-directory-pagination"
          >
            <span>共 {{ filteredTotal }} 名学生</span>
            <div class="teacher-directory-pagination-actions">
              <button
                type="button"
                class="teacher-btn teacher-btn--ghost teacher-directory-pagination-button"
                :disabled="page === 1"
                @click="emit('changePage', page - 1)"
              >
                上一页
              </button>
              <span>{{ page }} / {{ totalPages }}</span>
              <button
                type="button"
                class="teacher-btn teacher-btn--ghost teacher-directory-pagination-button"
                :disabled="page >= totalPages"
                @click="emit('changePage', page + 1)"
              >
                下一页
              </button>
            </div>
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
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --teacher-student-directory-columns: minmax(7.5rem, 0.7fr) minmax(10rem, 1fr) minmax(10rem, 0.9fr)
    minmax(12rem, 0.95fr) minmax(8rem, 0.8fr) minmax(8.5rem, 0.85fr);
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 7%, transparent),
      transparent 22rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      var(--journal-surface)
    );
  box-shadow: 0 22px 50px var(--color-shadow-soft);
}

.journal-eyebrow {
  letter-spacing: 0.08em;
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-topbar {
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-summary {
  display: grid;
  gap: 1.1rem;
  padding: 1.5rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-summary-title {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.teacher-summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.teacher-summary-item {
  min-width: 0;
  padding-left: 1rem;
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-summary-label {
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-summary-value {
  margin-top: 0.55rem;
  font-size: 1.35rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--journal-ink);
}

.teacher-summary-helper {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.teacher-hero-divider {
  margin-top: 1.5rem;
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-controls {
  display: grid;
  gap: 1rem;
  padding: 1.5rem 0 0;
}

.teacher-controls-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 0.85rem;
}

.teacher-controls-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-filter-grid {
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
  margin-top: 1.5rem;
  display: grid;
  gap: 0.75rem;
}

.teacher-empty-state {
  margin-top: 1.5rem;
}

.teacher-directory {
  display: flex;
  flex-direction: column;
  margin-top: 1.5rem;
}

.teacher-directory-pagination-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.teacher-directory-pagination-button {
  min-width: 5.5rem;
}

.teacher-directory-pagination-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.teacher-directory-top {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 0.5rem 1rem;
  padding-bottom: 0.9rem;
}

.teacher-directory-title {
  font-size: 1.1rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--journal-ink);
}

.teacher-directory-meta {
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.teacher-directory-head {
  display: grid;
  grid-template-columns: var(--teacher-student-directory-columns);
  gap: 1rem;
  padding: 0 0 0.75rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-directory-head-cell {
  min-width: 0;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--teacher-student-directory-columns);
  gap: 1rem;
  align-items: center;
  width: 100%;
  padding: 1.1rem 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  box-shadow: inset 2px 0 0 color-mix(in srgb, var(--journal-accent) 62%, transparent);
  outline: none;
}

.teacher-directory-cell {
  display: grid;
  gap: 0.5rem;
  min-width: 0;
  align-content: center;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-cell-alias .teacher-directory-row-points,
.teacher-directory-row-points {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.teacher-directory-cell-student-no {
  font-size: 0.76rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--journal-muted);
  font-family:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
  font-variant-numeric: tabular-nums;
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-size: 0.98rem;
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-head-cell-student-no,
.teacher-directory-head-cell-name,
.teacher-directory-head-cell-alias,
.teacher-directory-cell-student-no,
.teacher-directory-cell-name,
.teacher-directory-cell-alias {
  justify-self: start;
  width: 100%;
}

.teacher-directory-row-points {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-copy {
  font-size: 0.84rem;
  line-height: 1.6;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-directory-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.65rem;
  padding: 0 0.62rem;
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-directory-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.teacher-directory-row-metrics {
  display: grid;
  gap: 0.25rem;
  font-size: 0.81rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
}

@media (max-width: 1080px) {
  .teacher-topbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid,
  .teacher-filter-grid {
    grid-template-columns: 1fr;
  }

  .teacher-directory-head {
    display: none;
  }

  .teacher-directory-row {
    grid-template-columns: 1fr;
    gap: 0.85rem;
    padding: 1rem 0;
  }
}
</style>
