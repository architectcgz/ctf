<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowRight, FolderKanban, Search } from 'lucide-vue-next'

import type { TeacherClassItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  changePage: [page: number]
  openDashboard: []
  openReportExport: []
  openClass: [className: string]
}>()

const filterQuery = ref('')

const classEntries = computed(() =>
  props.classes.map((item, index) => ({
    item,
    code: `CL-${String(index + 1).padStart(2, '0')}`,
  }))
)

const filteredClassEntries = computed(() => {
  const keyword = filterQuery.value.trim().toLowerCase()
  if (!keyword) return classEntries.value

  return classEntries.value.filter(({ item, code }) => {
    return code.toLowerCase().includes(keyword) || item.name.toLowerCase().includes(keyword)
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / Math.max(props.pageSize, 1))))
const currentPageStudentCount = computed(() =>
  props.classes.reduce((sum, item) => sum + (item.student_count || 0), 0)
)
</script>

<template>
  <div class="workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <main class="content-pane">
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading workspace-tab-heading__main">
            <div class="workspace-overline">
              Class Directory
            </div>
            <h1 class="teacher-title workspace-page-title">班级管理</h1>
            <p class="teacher-copy workspace-page-copy">
              查看当前可管理班级，并进入对应班级继续查看学生和训练表现。
            </p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('openDashboard')"
            >
              教学概览
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="emit('openReportExport')"
            >
              导出班级报告
            </button>
          </div>
        </header>

        <section class="teacher-summary metric-panel-default-surface">
          <div class="teacher-summary-title">
            <FolderKanban class="h-4 w-4" />
            <span>Directory Snapshot</span>
          </div>
          <div class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
            <div class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                班级数量
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ total }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前可管理班级总数
              </div>
            </div>
            <div class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                当前页学生数
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ currentPageStudentCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                当前分页已加载班级的学生数汇总
              </div>
            </div>
          </div>
        </section>

        <section
          class="workspace-directory-section teacher-directory-section"
          aria-label="班级目录"
        >
          <header class="list-heading">
            <div>
              <div class="workspace-overline">
                Class Directory
              </div>
              <h3 class="list-heading__title">
                班级目录
              </h3>
            </div>
            <div class="teacher-directory-meta">
              本页 {{ filteredClassEntries.length }} / {{ classes.length }} 个班级，共
              {{ total }} 个班级
            </div>
          </header>

          <section
            class="teacher-directory-filters"
            aria-label="班级过滤"
          >
            <div class="teacher-filter-grid teacher-filter-grid--single">
              <label class="teacher-field">
                <span class="teacher-field-label">搜索班级</span>
                <div class="teacher-field-control teacher-filter-control">
                  <Search class="h-4 w-4 text-text-muted" />
                  <input
                    v-model="filterQuery"
                    type="text"
                    placeholder="搜索班级编号或名称"
                    class="teacher-input"
                  >
                </div>
              </label>
            </div>
          </section>

          <div
            v-if="loading"
            class="teacher-skeleton-list workspace-directory-loading"
          >
            <div
              v-for="index in 5"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-elevated)]"
            />
          </div>

          <AppEmpty
            v-else-if="classes.length === 0"
            class="teacher-empty-state workspace-directory-empty"
            icon="Users"
            title="暂无班级"
            description="当前教师账号下还没有可访问的班级。"
          />

          <section
            v-else
            class="teacher-directory"
          >
            <div
              v-if="filteredClassEntries.length > 0"
              class="teacher-directory-head"
            >
              <span class="teacher-directory-head-cell teacher-directory-head-cell-class-code">
                班级编号
              </span>
              <span class="teacher-directory-head-cell teacher-directory-head-cell-class-name">
                班级名称
              </span>
              <span class="teacher-directory-head-cell teacher-directory-head-cell-student-count">
                学生数
              </span>
              <span>状态</span>
              <span>操作</span>
            </div>

            <AppEmpty
              v-if="filteredClassEntries.length === 0"
              class="teacher-empty-state workspace-directory-empty"
              icon="Search"
              title="没有匹配班级"
              description="调整搜索关键词后再试。"
            />

            <div
              v-if="filteredClassEntries.length > 0"
              class="workspace-directory-list"
            >
              <button
                v-for="{ item, code } in filteredClassEntries"
                :key="item.name"
                type="button"
                class="teacher-directory-row group"
                :aria-label="`${item.name}，${item.student_count || 0} 名学生，进入班级`"
                @click="emit('openClass', item.name)"
              >
                <div class="teacher-directory-cell teacher-directory-cell-class-code">
                  {{ code }}
                </div>

                <div class="teacher-directory-cell teacher-directory-cell-class-name">
                  <h4
                    class="teacher-directory-row-title"
                    :title="item.name"
                  >
                    {{ item.name }}
                  </h4>
                </div>

                <div class="teacher-directory-cell teacher-directory-cell-student-count">
                  <div class="teacher-directory-row-points">
                    {{ item.student_count || 0 }}
                  </div>
                </div>

                <div class="teacher-directory-state">
                  <span
                    class="teacher-directory-state-chip"
                    :class="
                      (item.student_count || 0) > 0
                        ? 'teacher-directory-state-chip-ready'
                        : 'teacher-directory-state-chip-empty'
                    "
                  >
                    {{ (item.student_count || 0) > 0 ? '可查看' : '待入班' }}
                  </span>
                </div>

                <div class="teacher-directory-row-cta">
                  <span>进入班级</span>
                  <ArrowRight class="h-4 w-4" />
                </div>
              </button>
            </div>

            <div
              v-if="total > 0 && filteredClassEntries.length > 0"
              class="teacher-directory-pagination workspace-directory-pagination"
            >
              <PagePaginationControls
                :page="page"
                :total-pages="totalPages"
                :total="total"
                :total-label="`共 ${total} 个班级`"
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
  --teacher-directory-columns: var(--teacher-class-directory-columns);
  --teacher-class-directory-columns: minmax(7rem, 0.7fr) minmax(11rem, 1.15fr) minmax(7rem, 0.7fr)
    minmax(7rem, 0.7fr) minmax(7rem, 0.75fr);
  font-family: var(--font-family-sans);
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-directory-section {
  margin-top: var(--workspace-directory-page-block-gap, var(--space-5));
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
  padding: var(--workspace-directory-gap-top) 0 var(--space-4);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 22rem);
}

.teacher-filter-grid--single {
  justify-content: start;
}

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
  grid-template-columns: var(--teacher-class-directory-columns);
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

.teacher-directory-cell-class-code,
.teacher-directory-row-points {
  font-family: var(--font-family-mono);
}

.teacher-directory-cell-class-code {
  font-size: var(--font-size-0-76);
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-family: var(--font-family-sans);
  font-size: var(--font-size-1-08);
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
  font-size: var(--font-size-1-00);
  font-weight: 900;
  color: var(--color-text-primary);
}

.teacher-directory-head-cell-class-code,
.teacher-directory-head-cell-class-name,
.teacher-directory-head-cell-student-count,
.teacher-directory-cell-class-code,
.teacher-directory-cell-class-name,
.teacher-directory-cell-student-count {
  justify-self: start;
  width: 100%;
}

.teacher-directory-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 var(--space-3);
  border-radius: 0.5rem;
  font-size: var(--font-size-0-75);
  font-weight: 800;
}

.teacher-directory-state-chip-ready {
  background: var(--color-primary-soft);
  color: var(--color-primary);
}

.teacher-directory-state-chip-empty {
  background: var(--color-bg-elevated);
  color: var(--color-text-muted);
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

@media (max-width: 960px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
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
