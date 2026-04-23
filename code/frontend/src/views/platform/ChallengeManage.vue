<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowDownWideNarrow,
  ArrowUpNarrowWide,
  Book,
  Calendar,
  CheckCircle,
  Edit3,
  Plus,
  SortAsc,
} from 'lucide-vue-next'

import ChallengeManageDirectoryPanel from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue'
import {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { usePlatformChallenges, type PlatformChallengeListRow } from '@/composables/usePlatformChallenges'
import { useChallengeManagePresentation } from '@/composables/useChallengeManagePresentation'

type ChallengeSortOption = WorkspaceDirectorySortOption & {
  order: 'asc' | 'desc'
}

const router = useRouter()

const {
  list,
  total,
  page,
  pageSize,
  loading,
  error,
  keyword,
  categoryFilter,
  difficultyFilter,
  statusFilter,
  clearFilters,
  changePage,
  refresh,
  publish,
  remove,
} = usePlatformChallenges()

const publishedCount = computed(
  () => list.value.filter((item) => item.status === 'published').length
)
const draftCount = computed(() => list.value.filter((item) => item.status === 'draft').length)
const archivedCount = computed(() => list.value.filter((item) => item.status === 'archived').length)
const hasActiveFilters = computed(() =>
  Boolean(
    keyword.value.trim() || categoryFilter.value || difficultyFilter.value || statusFilter.value
  )
)
const manageEmptyTitle = computed(() => (hasActiveFilters.value ? '没有匹配题目' : '暂无题目'))
const manageEmptyMessage = computed(() =>
  hasActiveFilters.value
    ? '当前筛选条件下没有匹配题目。'
    : '当前还没有题目，请先前往导入页上传题目包。'
)
const hasLoadError = computed(() => Boolean(error.value) && list.value.length === 0)
const loadErrorMessage = computed(() =>
  error.value instanceof Error && error.value.message.trim().length > 0
    ? error.value.message
    : '题目目录暂时无法加载，请稍后重试。'
)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const {
  openActionMenuId,
  getCategoryLabel,
  getDifficultyLabel,
  closeActionMenu,
  openChallengeDetail,
  openChallengeTopology,
  openChallengeWriteup,
  submitPublishCheck,
  removeChallenge,
} = useChallengeManagePresentation({
  router,
  publish,
  remove,
})

const sortConfig = ref<ChallengeSortOption>({
  key: 'updateTime',
  order: 'desc',
  label: '最近更新',
  icon: Calendar,
})
const sortOptions: ChallengeSortOption[] = [
  { key: 'updateTime', order: 'desc', label: '最近更新', icon: Calendar },
  { key: 'points', order: 'desc', label: '分值由高到低', icon: ArrowDownWideNarrow },
  { key: 'points', order: 'asc', label: '分值由低到高', icon: ArrowUpNarrowWide },
  { key: 'title', order: 'asc', label: '标题 A-Z', icon: SortAsc },
]
const sortedChallenges = computed<PlatformChallengeListRow[]>(() => {
  const nextRows = [...list.value]

  nextRows.sort((left, right) => {
    switch (sortConfig.value.key) {
      case 'points': {
        const delta = left.points - right.points
        return sortConfig.value.order === 'asc' ? delta : -delta
      }
      case 'title': {
        const delta = left.title.localeCompare(right.title, 'zh-CN')
        return sortConfig.value.order === 'asc' ? delta : -delta
      }
      case 'updateTime':
      default: {
        const leftTime = new Date(left.updated_at || left.created_at).getTime()
        const rightTime = new Date(right.updated_at || right.created_at).getTime()
        return sortConfig.value.order === 'asc' ? leftTime - rightTime : rightTime - leftTime
      }
    }
  })

  return nextRows
})

function setSort(option: WorkspaceDirectorySortOption) {
  const matchedOption =
    sortOptions.find((item) => item.key === option.key && item.label === option.label) ??
    sortOptions[0]

  if (!matchedOption) {
    return
  }

  sortConfig.value = matchedOption
}

function setActionMenuOpen(challengeId: string, nextOpen: boolean): void {
  if (nextOpen) {
    openActionMenuId.value = challengeId
    return
  }

  if (openActionMenuId.value === challengeId) {
    closeActionMenu()
  }
}

async function openImportWorkspace(): Promise<void> {
  await router.push({ name: 'PlatformChallengeImportManage' })
}
</script>

<template>
  <section class="workspace-shell challenge-manage-shell journal-shell journal-shell-admin journal-notes-card journal-hero">
    <div class="workspace-grid">
      <main class="content-pane challenge-manage-content">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Challenge Workspace</div>
            <h1 class="workspace-page-title">
              题目资源管理中心
            </h1>
            <p class="workspace-page-copy">集中查看题目目录、发布状态与题库变更。</p>
          </div>
          <div class="awd-library-hero-actions">
            <div class="quick-actions">
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="openImportWorkspace"
              >
                <Plus class="h-4 w-4" />
                导入资源包
              </button>
            </div>
          </div>
        </section>

        <div class="challenge-manage-panel">
          <div class="manage-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
            <article class="journal-note progress-card metric-panel-card">
              <div class="challenge-metric-head">
                <div class="journal-note-label progress-card-label metric-panel-label">
                  <span>题目总量</span>
                  <Book class="h-4 w-4" />
                </div>
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="journal-note-value progress-card-value metric-panel-value">
                  {{ total.toString().padStart(2, '0') }}
                </div>
                <div class="journal-note-helper progress-card-hint metric-panel-helper">
                  题目资源总计
                </div>
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="challenge-metric-head">
                <div class="journal-note-label progress-card-label metric-panel-label">
                  <span>已发布</span>
                  <CheckCircle class="h-4 w-4" />
                </div>
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="journal-note-value progress-card-value metric-panel-value">
                  {{ publishedCount.toString().padStart(2, '0') }}
                </div>
                <div class="journal-note-helper progress-card-hint metric-panel-helper">
                  线上公开题目
                </div>
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="challenge-metric-head">
                <div class="journal-note-label progress-card-label metric-panel-label">
                  <span>草稿存量</span>
                  <Edit3 class="h-4 w-4" />
                </div>
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="journal-note-value progress-card-value metric-panel-value">
                  {{ draftCount.toString().padStart(2, '0') }}
                </div>
                <div class="journal-note-helper progress-card-hint metric-panel-helper">
                  导入后仍待发布
                </div>
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="challenge-metric-head">
                <div class="journal-note-label progress-card-label metric-panel-label">
                  <span>已归档</span>
                  <Calendar class="h-4 w-4" />
                </div>
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="journal-note-value progress-card-value metric-panel-value">
                  {{ archivedCount.toString().padStart(2, '0') }}
                </div>
                <div class="journal-note-helper progress-card-hint metric-panel-helper">
                  只读保留题目
                </div>
              </div>
            </article>
          </div>

          <ChallengeManageDirectoryPanel
            :rows="sortedChallenges"
            :total="total"
            :page="page"
            :total-pages="totalPages"
            :loading="loading"
            :has-load-error="hasLoadError"
            :load-error-message="loadErrorMessage"
            :has-active-filters="hasActiveFilters"
            :manage-empty-title="manageEmptyTitle"
            :manage-empty-message="manageEmptyMessage"
            :keyword="keyword"
            :category-filter="categoryFilter"
            :difficulty-filter="difficultyFilter"
            :status-filter="statusFilter"
            :selected-sort-label="sortConfig.label"
            :sort-options="sortOptions"
            :open-action-menu-id="openActionMenuId"
            :get-category-label="getCategoryLabel"
            :get-difficulty-label="getDifficultyLabel"
            @update:keyword="keyword = $event"
            @update:category-filter="categoryFilter = $event"
            @update:difficulty-filter="difficultyFilter = $event"
            @update:status-filter="statusFilter = $event"
            @select-sort="setSort"
            @reset-filters="clearFilters"
            @retry="void refresh()"
            @change-page="changePage"
            @update-action-menu-open="setActionMenuOpen($event.challengeId, $event.open)"
            @open-detail="openChallengeDetail"
            @open-topology="openChallengeTopology"
            @open-writeup="openChallengeWriteup"
            @submit-publish-check="submitPublishCheck"
            @remove-challenge="removeChallenge"
          />
        </div>
      </main>
    </div>
  </section>
</template>

<style scoped>
.challenge-manage-shell {
  --challenge-page-bg: var(--journal-surface);
  --workspace-shell-bg: var(--challenge-page-bg);
  --workspace-shell-elevated-bg: var(--journal-surface-subtle);
  background: var(--workspace-shell-bg);
}

.challenge-manage-content {
  display: grid;
  gap: var(--space-6);
  gap: var(--workspace-directory-page-block-gap);
}

.challenge-manage-panel {
  display: grid;
  gap: var(--space-section-gap-compact, var(--space-4));
}

.challenge-manage-shell .manage-summary-grid {
  --metric-panel-columns: 4;
}

.challenge-metric-head {
  margin-bottom: var(--space-2);
}

.challenge-metric-value-wrap .metric-panel-value,
.challenge-metric-value-wrap .metric-panel-helper {
  margin-top: 0;
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.awd-library-hero-actions {
  display: flex;
  align-items: flex-end;
  padding-bottom: 0.5rem;
}

.quick-actions {
  display: flex;
  gap: 0.75rem;
}
</style>
