<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, FolderKanban, RefreshCcw, Waypoints } from 'lucide-vue-next'
import { useRoute } from 'vue-router'

import AppEmpty from '@/components/common/AppEmpty.vue'
import { useTeacherAwdReviewIndex } from '@/composables/useTeacherAwdReviewIndex'
import { formatDate } from '@/utils/format'

const { router, loading, error, contests, filters, hasContests, loadContests, openContest } =
  useTeacherAwdReviewIndex()
const route = useRoute()
const isAdminRoute = computed(() => route.name === 'AdminAWDReviewIndex')
const overviewRouteName = computed(() => (isAdminRoute.value ? 'AdminDashboard' : 'TeacherDashboard'))
const overviewLabel = computed(() => (isAdminRoute.value ? '平台概览' : '教学概览'))

const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'running', label: '进行中' },
  { value: 'ended', label: '已结束' },
  { value: 'frozen', label: '冻结中' },
]

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
</script>

<template>
  <div class="teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
    >
      <div class="teacher-page">
        <header class="teacher-topbar workspace-tab-heading awd-review-index-header">
          <div class="teacher-heading workspace-tab-heading__main">
            <div class="workspace-overline">AWD Review</div>
            <h1 class="teacher-title workspace-page-title">AWD复盘</h1>
            <p class="teacher-copy workspace-page-copy">
              集中查看赛事轮次、状态与导出就绪度，从统一入口进入整场或单轮复盘。
            </p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="router.push({ name: overviewRouteName })"
            >
              {{ overviewLabel }}
            </button>
            <button type="button" class="teacher-btn teacher-btn--primary" @click="loadContests">
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
              <div class="progress-card-value metric-panel-value">
                {{ contests.filter((item) => item.status === 'running').length }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                仍在持续产出实时攻防信号的赛事
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">可导出教师报告</div>
              <div class="progress-card-value metric-panel-value">
                {{ contests.filter((item) => item.export_ready).length }}
              </div>
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
            <div class="teacher-directory-meta">共 {{ contests.length }} 场赛事</div>
          </header>

          <section class="teacher-directory-filters" aria-label="赛事过滤">
            <div class="awd-review-filter-grid">
              <label class="awd-review-field">
                <span class="awd-review-field__label">赛事状态</span>
                <select v-model="filters.status" class="awd-review-field__control">
                  <option
                    v-for="option in statusOptions"
                    :key="option.value || 'all'"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
              </label>

              <label class="awd-review-field awd-review-field--wide">
                <span class="awd-review-field__label">关键词</span>
                <input
                  v-model="filters.keyword"
                  type="text"
                  class="awd-review-field__control"
                  placeholder="搜索赛事标题"
                />
              </label>
            </div>
          </section>

          <div v-if="loading" class="teacher-skeleton-list workspace-directory-loading">
            <div
              v-for="index in 3"
              :key="index"
              class="h-28 animate-pulse rounded-[22px] bg-[color-mix(in_srgb,var(--journal-surface-subtle)_92%,transparent)]"
            />
          </div>

          <AppEmpty
            v-else-if="error"
            class="teacher-empty-state workspace-directory-empty"
            icon="AlertTriangle"
            title="AWD复盘目录加载失败"
            :description="error"
          >
            <template #action>
              <button type="button" class="teacher-btn teacher-btn--primary" @click="loadContests">
                重新加载
              </button>
            </template>
          </AppEmpty>

          <AppEmpty
            v-else-if="!hasContests"
            class="teacher-empty-state workspace-directory-empty"
            icon="Waypoints"
            title="暂无 AWD 赛事"
            description="当前还没有可进入复盘的 AWD 赛事。"
          />

          <section v-else class="teacher-directory">
            <div class="teacher-directory-head">
              <span class="teacher-directory-head-cell teacher-directory-head-cell-code">代号</span>
              <span class="teacher-directory-head-cell teacher-directory-head-cell-name">赛事</span>
              <span>轮次</span>
              <span>队伍</span>
              <span>状态</span>
              <span>操作</span>
            </div>

            <button
              v-for="contest in contests"
              :key="contest.id"
              type="button"
              class="teacher-directory-row"
              @click="openContest(contest.id)"
            >
              <div class="teacher-directory-cell teacher-directory-cell-code">
                AWD-{{ contest.id }}
              </div>

              <div class="teacher-directory-cell teacher-directory-cell-name">
                <h4 class="teacher-directory-row-title">{{ contest.title }}</h4>
                <p class="teacher-directory-row-copy">
                  最近信号
                  {{ contest.latest_evidence_at ? formatDate(contest.latest_evidence_at) : '暂无' }}
                </p>
              </div>

              <div class="teacher-directory-row-metrics">
                <span>{{
                  contest.current_round ? `第 ${contest.current_round} 轮` : '未开始'
                }}</span>
                <span>共 {{ contest.round_count }} 轮</span>
              </div>

              <div class="teacher-directory-row-metrics">
                <span>{{ contest.team_count }} 支队伍</span>
                <span>{{ contest.mode.toUpperCase() }}</span>
              </div>

              <div class="teacher-directory-row-tags">
                <span class="teacher-directory-chip">
                  {{ contestStatusLabel(contest.status) }}
                </span>
                <span
                  class="teacher-directory-chip"
                  :class="contest.export_ready ? '' : 'teacher-directory-chip-muted'"
                >
                  {{ contest.export_ready ? '可导出' : '实时复盘' }}
                </span>
              </div>

              <div class="teacher-directory-row-cta">
                <span>进入复盘</span>
                <ArrowRight class="h-4 w-4" />
              </div>
            </button>
          </section>
        </section>
      </div>
    </section>
  </div>
</template>

<style scoped>
.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
  --awd-review-directory-columns: minmax(0, 7rem) minmax(0, 2.1fr) minmax(0, 1fr) minmax(0, 0.85fr)
    minmax(0, 1fr) auto;
}

.teacher-summary--flat {
  border-bottom: 0;
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
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-directory-filters {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) 0;
}

.awd-review-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(14rem, 16rem) minmax(16rem, 1fr);
}

.awd-review-field {
  display: grid;
  gap: var(--space-2);
}

.awd-review-field__label {
  font-size: var(--font-size-0-80);
  font-weight: 600;
  color: var(--journal-muted);
}

.awd-review-field__control {
  min-height: 2.8rem;
  width: 100%;
  border: 1px solid var(--teacher-control-border);
  border-radius: 16px;
  background: var(--journal-surface);
  padding: 0 0.95rem;
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 160ms ease,
    box-shadow 160ms ease;
}

.awd-review-field__control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 46%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.awd-review-field__control::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 80%, transparent);
}

.teacher-skeleton-list {
  display: grid;
  gap: var(--space-3);
}

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-meta {
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.teacher-directory-head {
  display: grid;
  grid-template-columns: var(--awd-review-directory-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px dashed var(--teacher-divider);
  color: var(--journal-muted);
  font-size: var(--font-size-0-76);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--awd-review-directory-columns);
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-4-5) 0;
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
  gap: var(--space-2);
  min-width: 0;
  align-content: center;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-cell-code {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-78);
  font-weight: 700;
  color: var(--journal-muted);
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-02);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-copy {
  margin: 0;
  font-size: var(--font-size-0-84);
  line-height: 1.6;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-directory-row-metrics {
  display: grid;
  gap: var(--space-1);
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.teacher-directory-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.7rem;
  padding: 0 var(--space-2-5);
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: var(--font-size-0-75);
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-directory-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  color: var(--journal-accent-strong);
  font-size: var(--font-size-0-82);
  font-weight: 700;
}

@media (max-width: 1080px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .awd-review-filter-grid,
  .teacher-summary-grid {
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
}
</style>
