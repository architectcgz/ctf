<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, FolderKanban, RefreshCcw } from 'lucide-vue-next'

import type { TeacherAWDReviewContestItemData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { formatDate } from '@/utils/format'
import TeacherAWDReviewIndexFilters from './TeacherAWDReviewIndexFilters.vue'
import TeacherAWDReviewSummaryPanel from './TeacherAWDReviewSummaryPanel.vue'
import TeacherAWDReviewSurfaceShell from './TeacherAWDReviewSurfaceShell.vue'
import TeacherAWDReviewWorkspaceHeader from './TeacherAWDReviewWorkspaceHeader.vue'

interface ContestSummary {
  totalCount: number
  runningCount: number
  exportReadyCount: number
}

type ContestStatusOption = {
  value: '' | TeacherAWDReviewContestItemData['status']
  label: string
}

const props = defineProps<{
  loading: boolean
  error: string | null
  contests: TeacherAWDReviewContestItemData[]
  hasContests: boolean
  statusOptions: readonly ContestStatusOption[]
  contestSummary: ContestSummary
  statusFilter: '' | TeacherAWDReviewContestItemData['status']
  keywordFilter: string
  contestStatusLabel: (status: string) => string
}>()

const emit = defineEmits<{
  openDashboard: []
  refresh: []
  reload: []
  openContest: [contestId: string]
  updateStatusFilter: [status: '' | TeacherAWDReviewContestItemData['status']]
  updateKeywordFilter: [keyword: string]
}>()

const summaryItems = computed(() => [
  {
    label: '赛事数量',
    value: props.contestSummary.totalCount,
    hint: '当前可进入 AWD 复盘的赛事总数',
  },
  {
    label: '进行中',
    value: props.contestSummary.runningCount,
    hint: '仍在持续产出实时攻防信号的赛事',
  },
  {
    label: '可导出教师报告',
    value: props.contestSummary.exportReadyCount,
    hint: '已结束并允许生成教师复盘报告的赛事',
  },
])
</script>

<template>
  <TeacherAWDReviewSurfaceShell>
    <div class="teacher-page">
      <TeacherAWDReviewWorkspaceHeader
        overline="AWD Review"
        title="AWD复盘"
        header-class="awd-review-index-header"
        overline-class="awd-review-index-overline"
      >
        <template #description>
          集中查看赛事轮次、状态与导出就绪度，从统一入口进入整场或单轮复盘。
        </template>

        <template #actions>
          <button
            type="button"
            class="teacher-btn teacher-btn--ghost"
            @click="emit('openDashboard')"
          >
            教学概览
          </button>
          <button
            type="button"
            class="teacher-btn teacher-btn--primary"
            @click="emit('refresh')"
          >
            <RefreshCcw class="h-4 w-4" />
            刷新目录
          </button>
        </template>
      </TeacherAWDReviewWorkspaceHeader>

      <TeacherAWDReviewSummaryPanel
        title="Review Snapshot"
        :items="summaryItems"
      >
        <template #title-prefix>
          <FolderKanban class="h-4 w-4" />
        </template>
      </TeacherAWDReviewSummaryPanel>

      <section
        class="workspace-directory-section teacher-directory-section"
        aria-label="AWD 赛事目录"
      >
        <header class="list-heading">
          <div>
            <div class="journal-note-label">
              Review Directory
            </div>
            <h3 class="list-heading__title">
              赛事目录
            </h3>
          </div>
          <div class="teacher-directory-meta">
            共 {{ contests.length }} 场赛事
          </div>
        </header>

        <TeacherAWDReviewIndexFilters
          :status-options="statusOptions"
          :status-filter="statusFilter"
          :keyword-filter="keywordFilter"
          @update-status-filter="emit('updateStatusFilter', $event)"
          @update-keyword-filter="emit('updateKeywordFilter', $event)"
        />

        <div
          v-if="loading"
          class="teacher-skeleton-list workspace-directory-loading"
        >
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
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('reload')"
            >
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

        <section
          v-else
          class="teacher-directory"
        >
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
            @click="emit('openContest', contest.id)"
          >
            <div class="teacher-directory-cell teacher-directory-cell-code">
              AWD-{{ contest.id }}
            </div>

            <div class="teacher-directory-cell teacher-directory-cell-name">
              <h4 class="teacher-directory-row-title">
                {{ contest.title }}
              </h4>
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
  </TeacherAWDReviewSurfaceShell>
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

.awd-review-index-overline {
  font-size: var(--journal-overline-font-size, var(--font-size-0-70));
  font-weight: 700;
  letter-spacing: var(--journal-overline-letter-spacing, 0.2em);
  text-transform: uppercase;
  color: var(--journal-accent, var(--color-primary));
}

.teacher-directory-section {
  margin-top: var(--workspace-directory-page-block-gap, var(--space-5));
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
