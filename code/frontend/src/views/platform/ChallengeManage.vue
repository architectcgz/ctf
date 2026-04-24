<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowDownWideNarrow,
  ArrowUpNarrowWide,
  Calendar,
  SortAsc,
} from 'lucide-vue-next'

import ChallengeManageDirectoryPanel from '@/components/platform/challenge/ChallengeManageDirectoryPanel.vue'
import ChallengeManageHeroPanel from '@/components/platform/challenge/ChallengeManageHeroPanel.vue'
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
        <div class="challenge-manage-panel">
          <ChallengeManageHeroPanel
            :total="total"
            :published-count="publishedCount"
            :draft-count="draftCount"
            :archived-count="archivedCount"
            @import="void openImportWorkspace()"
          />

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
</style>
