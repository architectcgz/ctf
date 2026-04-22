<script setup lang="ts">
import { ArrowRight, Search, Activity } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'

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
  <div class="workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <main class="content-pane">
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading">
            <div class="teacher-surface-eyebrow journal-eyebrow">
              Student Directory
            </div>
            <h1 class="teacher-title">
              学生管理
            </h1>
            <p class="teacher-copy">
              按班级筛选、搜索并进入学员分析。
            </p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="emit('openClassManagement')"
            >
              班级管理
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>
        </header>

        <section class="teacher-summary metric-panel-default-surface">
          <div class="teacher-summary-title">
            <span>Directory Snapshot</span>
          </div>
          <div class="teacher-summary-grid progress-strip metric-panel-grid">
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                可访问班级
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ classes.length }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前教师可切换的班级数量
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                当前班级学生
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ totalStudents }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前选中班级的学生总数
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                搜索结果
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ filteredStudents.length }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前搜索条件下匹配的学生数量
              </div>
            </article>
          </div>
        </section>

        <section
          class="workspace-directory-section teacher-directory-section"
          aria-label="学生目录"
        >
          <header class="list-heading">
            <div>
              <div class="journal-note-label">
                Student Directory
              </div>
              <h3 class="list-heading__title">
                学生目录
              </h3>
            </div>
            <div class="teacher-directory-meta">
              共 {{ filteredTotal }} 名学生
            </div>
          </header>

          <section
            class="teacher-directory-filters"
            aria-label="学生过滤"
          >
            <div class="teacher-filter-grid">
              <label class="teacher-field">
                <span class="teacher-field-label">班级</span>
                <div class="teacher-field-control teacher-filter-control teacher-filter-control--select">
                  <select
                    :value="selectedClassName"
                    class="teacher-input teacher-select"
                    :disabled="loadingClasses"
                    @change="emit('selectClass', ($event.target as HTMLSelectElement).value)"
                  >
                    <option value="">全部班级</option>
                    <option
                      v-for="item in classes"
                      :key="item.name"
                      :value="item.name"
                    >
                      {{ item.name }} · {{ item.student_count || 0 }}
                    </option>
                  </select>
                </div>
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
                  >
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
                  >
                </div>
              </label>
            </div>
          </section>

          <div
            v-if="loadingStudents"
            class="teacher-skeleton-list workspace-directory-loading"
          >
            <div
              v-for="index in 6"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-elevated)]"
            />
          </div>

          <AppEmpty
            v-else-if="filteredStudents.length === 0"
            class="teacher-empty-state workspace-directory-empty"
            icon="Users"
            title="没有匹配学生"
            description="调整搜索词或切换班级后再试。"
          />

          <section
            v-else
            class="teacher-directory"
          >
            <div class="teacher-directory-head">
              <span class="teacher-directory-head-cell teacher-directory-head-cell-student-no">
                学号
              </span>
              <span class="teacher-directory-head-cell teacher-directory-head-cell-name">
                学生名称
              </span>
              <span class="teacher-directory-head-cell teacher-directory-head-cell-alias">
                昵称
              </span>
              <span>薄弱项</span>
              <span>做题数</span>
              <span>得分数</span>
              <span>操作</span>
            </div>

            <button
              v-for="student in filteredStudents"
              :key="student.id"
              type="button"
              class="teacher-directory-row group"
              :aria-label="`${student.name || student.username}，${student.solved_count ?? 0} 题，${student.total_score ?? 0} 分，查看学员分析`"
              @click="emit('openStudent', student.id)"
            >
              <div class="teacher-directory-cell teacher-directory-cell-student-no">
                {{ student.student_no || '未设置' }}
              </div>

              <div class="teacher-directory-cell teacher-directory-cell-name">
                <h4
                  class="teacher-directory-row-title"
                  :title="student.name || '未设置姓名'"
                >
                  {{ student.name || '未设置姓名' }}
                </h4>
              </div>

              <div class="teacher-directory-cell teacher-directory-cell-alias">
                <div
                  class="teacher-directory-row-points"
                  :title="student.username"
                >
                  {{ student.username }}
                </div>
              </div>

              <div class="teacher-directory-row-tags">
                <span class="teacher-directory-chip teacher-directory-chip-muted">
                  {{ student.weak_dimension || '暂无薄弱项' }}
                </span>
              </div>

              <div class="teacher-directory-row-solved">
                {{ student.solved_count ?? 0 }}
              </div>

              <div class="teacher-directory-row-score">
                {{ student.total_score ?? 0 }}
              </div>

              <div class="teacher-directory-row-cta">
                <span>学员分析</span>
                <ArrowRight class="h-4 w-4" />
              </div>
            </button>

            <div
              v-if="filteredTotal > 0"
              class="teacher-directory-pagination workspace-directory-pagination"
            >
              <PagePaginationControls
                :page="page"
                :total-pages="totalPages"
                :total="filteredTotal"
                :total-label="`共 ${filteredTotal} 名学生`"
                @change-page="emit('changePage', $event)"
              />
            </div>
          </section>
        </section>
        <div
          v-if="error"
          class="teacher-surface-error"
        >
          {{ error }}
          <button
            type="button"
            class="ml-3 font-medium underline"
            @click="emit('retry')"
          >
            重试
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --teacher-directory-columns: var(--teacher-student-directory-columns);
  --teacher-student-directory-columns: minmax(7.5rem, 0.7fr) minmax(10rem, 1fr) minmax(10rem, 0.9fr)
    minmax(12rem, 0.95fr) minmax(6rem, 0.55fr) minmax(6rem, 0.55fr) minmax(8.5rem, 0.85fr);
  font-family: var(--font-family-sans);
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-directory-section {
  margin-top: var(--space-6);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 900;
  color: var(--color-text-primary);
}

.teacher-directory-filters {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) 0;
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: 220px minmax(0, 1fr) minmax(0, 1fr);
}

.teacher-filter-control--select { justify-content: flex-start; }
.teacher-select { min-height: 1.75rem; border: 0; appearance: none; cursor: pointer; background: transparent; width: 100%; outline: none; }

.teacher-skeleton-list {
  display: grid;
  gap: var(--space-3);
}

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--teacher-student-directory-columns);
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-5) 0;
  border: 0;
  border-bottom: 1px solid var(--color-border-subtle);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: var(--color-primary-soft);
  box-shadow: inset 3px 0 0 var(--color-primary);
  outline: none;
}

.teacher-directory-cell {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
  align-content: center;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-cell-alias .teacher-directory-row-points,
.teacher-directory-row-points {
  font-family: var(--font-family-mono);
}

.teacher-directory-cell-student-no {
  font-size: var(--font-size-0-76);
  font-weight: 800;
  letter-spacing: 0.02em;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-size: var(--font-size-0-98);
  font-weight: 800;
  line-height: 1.35;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.group:hover .teacher-directory-row-title {
  color: var(--color-primary);
}

.teacher-directory-row-points {
  font-size: var(--font-size-0-80);
  font-weight: 800;
  color: var(--color-text-secondary);
}

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.65rem;
  padding: 0 var(--space-2-5);
  border-radius: 0.5rem;
  background: var(--color-bg-elevated);
  font-size: var(--font-size-0-75);
  font-weight: 800;
  color: var(--color-text-secondary);
}

.teacher-directory-row-solved,
.teacher-directory-row-score {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-81);
  font-weight: 800;
  color: var(--color-text-primary);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-0-82);
  font-weight: 800;
  color: var(--color-primary);
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.2s ease;
}

.teacher-directory-row:hover .teacher-directory-row-cta {
  opacity: 1;
  transform: translateX(0);
}

@media (max-width: 1080px) {
  .teacher-topbar,
  .list-heading {
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
    gap: var(--space-3);
    padding: var(--space-4) 0;
  }
  
  .teacher-directory-row-cta {
    opacity: 1;
    transform: none;
  }
}
</style>