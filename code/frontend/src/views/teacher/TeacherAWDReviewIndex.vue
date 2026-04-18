<script setup lang="ts">
import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import type { WorkspaceDirectorySortOption } from '@/components/common/WorkspaceDirectoryToolbar.vue'
import type { WorkspaceDataTableColumn } from '@/components/common/WorkspaceDataTable.vue'

import { computed } from 'vue'
import { ArrowRight, FolderKanban, RefreshCcw, Shield, Waypoints } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useTeacherAwdReviewIndex } from '@/composables/useTeacherAwdReviewIndex'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'
import { resolveTeachingDashboardRouteName } from '@/utils/teachingWorkspaceRouting'

const authStore = useAuthStore()
const { router, loading, error, contests, filters, hasContests, loadContests, openContest } =
  useTeacherAwdReviewIndex()
const isAdminView = computed(() => authStore.user?.role === 'admin')

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'running', label: '进行中' },
  { value: 'ended', label: '已结束' },
  { value: 'frozen', label: '冻结中' },
]

const reviewTableColumns: WorkspaceDataTableColumn[] = [
  {
    key: 'id',
    label: '赛事编号',
    widthClass: 'w-[16%] min-w-[8rem]',
    cellClass: 'awd-review-table__mono',
  },
  {
    key: 'title',
    label: '赛事名称',
    widthClass: 'w-[32%] min-w-[16rem]',
    cellClass: 'awd-review-table__name-cell',
  },
  {
    key: 'round_count',
    label: '轮次',
    widthClass: 'w-[15%] min-w-[8rem]',
    cellClass: 'awd-review-table__metrics-cell',
  },
  {
    key: 'team_count',
    label: '队伍',
    widthClass: 'w-[15%] min-w-[8rem]',
    cellClass: 'awd-review-table__metrics-cell',
  },
  {
    key: 'status',
    label: '状态',
    align: 'center',
    widthClass: 'w-[14%] min-w-[9rem]',
    cellClass: 'awd-review-table__status-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right',
    widthClass: 'w-[12rem]',
    cellClass: 'awd-review-table__action-cell',
  },
]

const toolbarSortOptions: WorkspaceDirectorySortOption[] = []
const runningContestCount = computed(() => contests.value.filter((item) => item.status === 'running').length)
const exportReadyContestCount = computed(() => contests.value.filter((item) => item.export_ready).length)
const hasActiveFilters = computed(() => Boolean(filters.value.status || filters.value.keyword.trim()))
const overviewLabel = computed(() => (authStore.user?.role === 'admin' ? '平台概览' : '教学概览'))
const rootClasses = computed(() =>
  isAdminView.value
    ? 'workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col'
    : 'workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col'
)
const shellTag = computed(() => 'main')
const shellClasses = computed(() =>
  isAdminView.value ? 'content-pane awd-review-admin-pane' : 'content-pane'
)
const ghostActionClass = computed(() =>
  isAdminView.value ? 'ui-btn ui-btn--ghost' : 'teacher-btn teacher-btn--ghost'
)
const primaryActionClass = computed(() =>
  isAdminView.value ? 'ui-btn ui-btn--primary' : 'teacher-btn teacher-btn--primary'
)

function resetFilters(): void {
  filters.value.status = ''
  filters.value.keyword = ''
}

function contestStatusLabel(status: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'ended':
      return '已结束'
    case 'frozen':
      return '冻结中'
    case 'published':
      return '已发布'
    default:
      return status || '未开始'
  }
}

function latestEvidenceLabel(contest: TeacherAWDReviewContestItemData): string {
  return contest.latest_evidence_at ? formatDate(contest.latest_evidence_at) : '暂无'
}
</script>

<template>
  <div :class="rootClasses">
    <component :is="shellTag" :class="shellClasses">
      <div class="teacher-page">
        <header class="teacher-topbar workspace-tab-heading awd-review-index-header">
          <div class="teacher-heading workspace-tab-heading__main">
            <div class="workspace-overline awd-review-index-overline">AWD Review</div>
            <h1 class="teacher-title workspace-page-title">AWD复盘</h1>
            <p class="teacher-copy workspace-page-copy">
              集中查看赛事轮次、状态与导出就绪度，从统一入口进入整场或单轮复盘。
            </p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              :class="ghostActionClass"
              @click="router.push({ name: resolveTeachingDashboardRouteName(authStore.user?.role) })"
            >
              {{ overviewLabel }}
            </button>
            <button type="button" :class="primaryActionClass" @click="loadContests">
              <RefreshCcw class="h-4 w-4" />
              刷新目录
            </button>
          </div>
        </header>

        <section class="teacher-summary teacher-summary--flat metric-panel-default-surface">
          <div class="teacher-summary-title">
            <FolderKanban class="h-4 w-4" />
            <span>Review Snapshot</span>
          </div>
          <div class="teacher-summary-grid progress-strip metric-panel-grid">
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">赛事数量</div>
              <div class="progress-card-value metric-panel-value">{{ contests.length }}</div>
              <div class="progress-card-hint metric-panel-helper">
                当前可进入 AWD 复盘的赛事总数
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">进行中</div>
              <div class="progress-card-value metric-panel-value">{{ runningContestCount }}</div>
              <div class="progress-card-hint metric-panel-helper">
                仍在持续产出实时攻防信号的赛事
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">可导出教师报告</div>
              <div class="progress-card-value metric-panel-value">{{ exportReadyContestCount }}</div>
              <div class="progress-card-hint metric-panel-helper">
                已结束并允许生成教师复盘报告的赛事
              </div>
            </article>
          </div>
        </section>

        <section
          class="workspace-directory-section teacher-directory-section"
          aria-label="AWD 赛事目录"
        >
          <header class="list-heading">
            <div>
              <div class="journal-note-label">Review Directory</div>
              <h3 class="list-heading__title">赛事目录</h3>
            </div>
          </header>

          <WorkspaceDirectoryToolbar
            v-model="filters.keyword"
            :total="contests.length"
            selected-sort-label=""
            :sort-options="toolbarSortOptions"
            search-placeholder="搜索赛事标题"
            filter-button-label="状态"
            total-suffix="场赛事"
            filter-panel-kicker="Review Filters"
            filter-panel-title="筛选赛事"
            reset-label="重置筛选"
            :reset-disabled="!hasActiveFilters"
            @reset-filters="resetFilters"
          >
            <template #filter-panel>
              <div class="awd-review-filter-grid">
                <label class="awd-review-filter-field">
                  <span class="awd-review-filter-field__label">赛事状态</span>
                  <select v-model="filters.status" class="awd-review-filter-field__control">
                    <option
                      v-for="option in statusOptions"
                      :key="option.value || 'all'"
                      :value="option.value"
                    >
                      {{ option.label }}
                    </option>
                  </select>
                </label>
              </div>
            </template>
          </WorkspaceDirectoryToolbar>

          <div v-if="loading" class="teacher-skeleton-list workspace-directory-loading">
            <div
              v-for="index in 3"
              :key="index"
              class="h-28 animate-pulse rounded-[22px] bg-[color-mix(in_srgb,var(--journal-surface-subtle)_92%,transparent)]"
            />
          </div>

          <AppEmpty
            v-else-if="error"
            class="teacher-empty-state workspace-directory-empty awd-review-directory__empty"
            icon="AlertTriangle"
            title="AWD复盘目录加载失败"
            :description="error"
          >
            <template #action>
              <button type="button" :class="primaryActionClass" @click="loadContests">
                重新加载
              </button>
            </template>
          </AppEmpty>

          <AppEmpty
            v-else-if="!hasContests"
            class="teacher-empty-state workspace-directory-empty awd-review-directory__empty"
            icon="Waypoints"
            title="暂无 AWD 赛事"
            description="当前还没有可进入复盘的 AWD 赛事。"
          />

          <WorkspaceDataTable
            v-else
            class="workspace-directory-list awd-review-table"
            :columns="reviewTableColumns"
            :rows="contests"
            row-key="id"
            row-class="awd-review-table__row"
          >
            <template #cell-id="{ row }">
              <span class="awd-review-table__code">
                AWD-{{ (row as TeacherAWDReviewContestItemData).id }}
              </span>
            </template>

            <template #cell-title="{ row }">
              <div class="awd-review-table__name-wrap">
                <span
                  class="awd-review-table__name"
                  :title="(row as TeacherAWDReviewContestItemData).title"
                >
                  {{ (row as TeacherAWDReviewContestItemData).title }}
                </span>
                <span class="awd-review-table__meta">
                  最近信号 {{ latestEvidenceLabel(row as TeacherAWDReviewContestItemData) }}
                </span>
              </div>
            </template>

            <template #cell-round_count="{ row }">
              <div class="awd-review-table__metrics">
                <span>
                  {{
                    (row as TeacherAWDReviewContestItemData).current_round
                      ? `第 ${(row as TeacherAWDReviewContestItemData).current_round} 轮`
                      : '未开始'
                  }}
                </span>
                <span>共 {{ (row as TeacherAWDReviewContestItemData).round_count }} 轮</span>
              </div>
            </template>

            <template #cell-team_count="{ row }">
              <div class="awd-review-table__metrics">
                <span>{{ (row as TeacherAWDReviewContestItemData).team_count }} 支队伍</span>
                <span>{{ (row as TeacherAWDReviewContestItemData).mode.toUpperCase() }}</span>
              </div>
            </template>

            <template #cell-status="{ row }">
              <div class="awd-review-table__status-stack">
                <span class="awd-review-table__status-pill">
                  {{ contestStatusLabel((row as TeacherAWDReviewContestItemData).status) }}
                </span>
                <span
                  class="awd-review-table__status-pill awd-review-table__status-pill--muted"
                  :class="{
                    'awd-review-table__status-pill--accent':
                      (row as TeacherAWDReviewContestItemData).export_ready,
                  }"
                >
                  {{ (row as TeacherAWDReviewContestItemData).export_ready ? '可导出' : '实时复盘' }}
                </span>
              </div>
            </template>

            <template #cell-actions="{ row }">
              <button
                type="button"
                class="awd-review-table__action"
                @click="openContest((row as TeacherAWDReviewContestItemData).id)"
              >
                <span>进入复盘</span>
                <ArrowRight class="h-4 w-4" />
              </button>
            </template>
          </WorkspaceDataTable>
        </section>
      </div>
    </component>
  </div>
</template>

<style scoped>
.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-summary--flat {
  border-bottom: 0;
}

.awd-review-index-overline {
  font-size: var(--journal-overline-font-size, var(--font-size-0-70));
  font-weight: 700;
  letter-spacing: var(--journal-overline-letter-spacing, 0.2em);
  text-transform: uppercase;
  color: var(--journal-accent, var(--color-primary));
}

.teacher-directory-section {
  margin-top: var(--space-6);
}

.teacher-directory-section > .list-heading {
  margin-bottom: 1.1rem;
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
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-skeleton-list {
  display: grid;
  gap: var(--space-3);
}

.awd-review-filter-grid {
  display: grid;
  gap: 0.85rem;
}

.awd-review-filter-field {
  display: grid;
  gap: 0.45rem;
}

.awd-review-filter-field__label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.awd-review-filter-field__control {
  min-height: 2.75rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
  padding: 0 0.9rem;
  color: var(--journal-ink);
}

.awd-review-directory__empty {
  border: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 1.25rem;
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
}

.awd-review-table {
  --awd-review-line-soft: color-mix(in srgb, var(--journal-border) 74%, transparent);
  border: 1px solid var(--awd-review-line-soft);
  border-radius: 1.35rem;
  background: color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base));
  padding: 0.25rem 0.9rem 0.4rem;
}

.awd-review-table :deep(.workspace-data-table__head-cell) {
  border-bottom-color: var(--awd-review-line-soft);
}

.awd-review-table :deep(.workspace-data-table__row) {
  border-bottom-color: var(--awd-review-line-soft);
}

.awd-review-table :deep(.workspace-data-table__body tr:last-child) {
  border-bottom-color: transparent;
}

.awd-review-table :deep(.workspace-data-table__row:hover) {
  background: color-mix(in srgb, var(--color-primary) 6%, transparent);
}

.awd-review-table__code,
.awd-review-table :deep(.awd-review-table__mono) {
  font-family: var(--font-family-mono);
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.awd-review-table__name-wrap {
  display: grid;
  min-width: 0;
  gap: 0.28rem;
}

.awd-review-table__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.awd-review-table__meta {
  font-size: 0.86rem;
  line-height: 1.5;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.awd-review-table__metrics {
  display: grid;
  gap: 0.2rem;
  font-size: 0.88rem;
  color: var(--journal-muted);
}

.awd-review-table__status-stack {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.45rem;
}

.awd-review-table__status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.95rem;
  min-width: 5.25rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 22%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  padding: 0 0.72rem;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--color-primary);
}

.awd-review-table__status-pill--muted {
  border-color: color-mix(in srgb, var(--journal-border) 80%, transparent);
  background: color-mix(in srgb, var(--journal-surface-subtle) 78%, transparent);
  color: var(--journal-muted);
}

.awd-review-table__status-pill--accent {
  border-color: color-mix(in srgb, var(--color-success) 22%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: color-mix(in srgb, var(--color-success) 82%, var(--journal-ink));
}

.awd-review-table__action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  min-height: 2.4rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 18%, var(--journal-border));
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 8%, var(--journal-surface));
  padding: 0 0.95rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-primary);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.awd-review-table__action:hover {
  border-color: color-mix(in srgb, var(--color-primary) 32%, var(--journal-border));
  background: color-mix(in srgb, var(--color-primary) 12%, var(--journal-surface));
}

.awd-review-table__action:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 24%, transparent);
  outline-offset: 2px;
}

@media (max-width: 1080px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 768px) {
  .awd-review-table {
    padding-inline: 0.65rem;
  }
}
</style>
