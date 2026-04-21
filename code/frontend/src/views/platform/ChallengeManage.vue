<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Book,
  CheckCircle,
  Edit3,
  Eye,
  MoreHorizontal,
  Trash2,
  FileSearch,
  Plus,
  ArrowUpNarrowWide,
  ArrowDownWideNarrow,
  SortAsc,
  Calendar,
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import CActionMenu from '@/components/common/menus/CActionMenu.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar, {
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

const sortConfig = ref({ key: 'updateTime', order: 'desc', label: '最近更新' })
const sortOptions: ChallengeSortOption[] = [
  { key: 'updateTime', order: 'desc', label: '最近更新', icon: Calendar },
  { key: 'points', order: 'desc', label: '分值由高到低', icon: ArrowDownWideNarrow },
  { key: 'points', order: 'asc', label: '分值由低到高', icon: ArrowUpNarrowWide },
  { key: 'title', order: 'asc', label: '标题 A-Z', icon: SortAsc },
]
const sortedChallenges = computed(() => {
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
const challengeTableColumns = [
  {
    key: 'title',
    label: '题目名称',
    widthClass: 'w-[42%] min-w-[18rem]',
    cellClass: 'challenge-table__title-cell',
  },
  {
    key: 'category',
    label: '分类',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[6rem]',
    cellClass: 'challenge-table__compact-cell',
  },
  {
    key: 'difficulty',
    label: '难度',
    align: 'center' as const,
    widthClass: 'w-[11%] min-w-[5.5rem]',
    cellClass: 'challenge-table__compact-cell',
  },
  {
    key: 'points',
    label: '分值',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[5rem]',
    cellClass: 'challenge-table__points-cell',
  },
  {
    key: 'status',
    label: '状态',
    widthClass: 'w-[13%] min-w-[7rem]',
    cellClass: 'challenge-table__compact-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[12rem]',
    cellClass: 'challenge-table__actions-cell',
  },
]
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

function resolveChallengeCategoryLabel(value: unknown): string {
  return getCategoryLabel(String(value) as never)
}

function resolveChallengeDifficultyLabel(value: unknown): string {
  return getDifficultyLabel(String(value) as never)
}

function getChallengeRow(row: unknown): PlatformChallengeListRow {
  return row as PlatformChallengeListRow
}
</script>

<template>
  <div class="workspace-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Challenge Workspace
            </div>
            <h1 class="hero-title">
              题目资源管理
            </h1>
            <p class="hero-summary">
              集中查看题目目录、发布状态与题库变更。
            </p>
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

        <div class="challenge-manage-body mt-10 space-y-10">
          <div class="metric-panel-grid metric-panel-grid--premium cols-4 manage-summary-grid">
            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>题目总量</span>
                <Book class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ total.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                题目资源总计
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>已发布</span>
                <CheckCircle class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ publishedCount.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                线上公开题目
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>草稿存量</span>
                <Edit3 class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ draftCount.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                导入后仍待发布
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>已归档</span>
                <Calendar class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ archivedCount.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                只读保留题目
              </div>
            </article>
          </div>

          <section class="workspace-directory-section challenge-manage-directory">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">
                  Challenge Directory
                </div>
                <h2 class="list-heading__title">
                  题目目录
                </h2>
              </div>
            </header>

            <WorkspaceDirectoryToolbar
              v-model="keyword"
              :total="total"
              :selected-sort-label="sortConfig.label"
              :sort-options="sortOptions"
              search-placeholder="检索题目名称..."
              :reset-disabled="!hasActiveFilters"
              @select-sort="setSort"
              @reset-filters="clearFilters"
            >
              <template #filter-panel>
                <div class="challenge-filter-grid">
                  <label class="challenge-filter-field">
                    <span class="challenge-filter-label">题目分类</span>
                    <select
                      v-model="categoryFilter"
                      class="challenge-filter-select"
                    >
                      <option value="">全部分类</option>
                      <option value="web">Web</option>
                      <option value="pwn">Pwn</option>
                      <option value="reverse">逆向</option>
                      <option value="crypto">密码</option>
                      <option value="misc">杂项</option>
                      <option value="forensics">取证</option>
                    </select>
                  </label>

                  <label class="challenge-filter-field">
                    <span class="challenge-filter-label">难度等级</span>
                    <select
                      v-model="difficultyFilter"
                      class="challenge-filter-select"
                    >
                      <option value="">全部难度</option>
                      <option value="beginner">入门</option>
                      <option value="easy">简单</option>
                      <option value="medium">中等</option>
                      <option value="hard">困难</option>
                      <option value="insane">地狱</option>
                    </select>
                  </label>

                  <label class="challenge-filter-field">
                    <span class="challenge-filter-label">发布状态</span>
                    <select
                      v-model="statusFilter"
                      class="challenge-filter-select"
                    >
                      <option value="">全部状态</option>
                      <option value="draft">草稿</option>
                      <option value="published">已发布</option>
                      <option value="archived">已归档</option>
                    </select>
                  </label>
                </div>
              </template>
            </WorkspaceDirectoryToolbar>

            <div
              v-if="loading"
              class="workspace-directory-loading"
            >
              正在同步题目目录...
            </div>
            <AppEmpty
              v-else-if="hasLoadError"
              class="workspace-directory-empty"
              icon="AlertTriangle"
              title="题目目录加载失败"
              :description="loadErrorMessage"
            >
              <template #action>
                <button
                  type="button"
                  class="ui-btn ui-btn--secondary"
                  @click="void refresh()"
                >
                  重新加载
                </button>
              </template>
            </AppEmpty>
            <AppEmpty
              v-else-if="list.length === 0"
              class="workspace-directory-empty"
              icon="BookOpen"
              :title="manageEmptyTitle"
              :description="manageEmptyMessage"
            />
            <WorkspaceDataTable
              v-else
              class="challenge-list workspace-directory-list"
              :columns="challengeTableColumns"
              :rows="sortedChallenges"
              row-key="id"
              row-class="challenge-table-row group"
            >
              <template #cell-title="{ row }">
                <div
                  class="challenge-table-title"
                  :title="getChallengeRow(row).title"
                >
                  {{ getChallengeRow(row).title }}
                </div>
              </template>

              <template #cell-category="{ row }">
                <span class="challenge-table-pill challenge-table-pill--category">
                  {{ resolveChallengeCategoryLabel(getChallengeRow(row).category) }}
                </span>
              </template>

              <template #cell-difficulty="{ row }">
                <span class="challenge-table-difficulty">
                  {{ resolveChallengeDifficultyLabel(getChallengeRow(row).difficulty) }}
                </span>
              </template>

              <template #cell-points="{ row }">
                <span class="challenge-table-points">{{ getChallengeRow(row).points }}</span>
              </template>

              <template #cell-status="{ row }">
                <div class="challenge-table-status">
                  <div
                    class="challenge-table-status__dot"
                    :class="
                      getChallengeRow(row).status === 'published'
                        ? 'challenge-table-status__dot--published'
                        : 'challenge-table-status__dot--idle'
                    "
                  />
                  <span class="challenge-table-status__label">
                    {{
                      getChallengeRow(row).status === 'published'
                        ? '已发布'
                        : getChallengeRow(row).status === 'archived'
                          ? '已归档'
                          : '草稿'
                    }}
                  </span>
                </div>
              </template>

              <template #cell-actions="{ row }">
                <div class="challenge-table-actions">
                  <button
                    type="button"
                    class="challenge-row-action"
                    @click="openChallengeDetail(getChallengeRow(row).id)"
                  >
                    <Eye class="h-3 w-3" />
                    查看
                  </button>

                  <CActionMenu
                    :open="openActionMenuId === getChallengeRow(row).id"
                    title="Management"
                    menu-label="题目更多操作"
                    @update:open="setActionMenuOpen(getChallengeRow(row).id, $event)"
                  >
                    <template #trigger="{ open, toggle, setTriggerRef }">
                      <button
                        :ref="setTriggerRef"
                        type="button"
                        class="c-action-menu__trigger c-action-menu__trigger--icon"
                        :aria-expanded="open ? 'true' : 'false'"
                        aria-haspopup="menu"
                        aria-label="题目更多操作"
                        @click.stop="toggle"
                      >
                        <MoreHorizontal class="h-3.5 w-3.5" />
                      </button>
                    </template>

                    <template #default>
                      <button
                        type="button"
                        class="c-action-menu__item"
                        @click="openChallengeTopology(getChallengeRow(row).id)"
                      >
                        <FileSearch class="h-3 w-3" />
                        编排拓扑
                      </button>
                      <button
                        type="button"
                        class="c-action-menu__item"
                        @click="openChallengeWriteup(getChallengeRow(row).id)"
                      >
                        <Book class="h-3 w-3" />
                        题解与提示
                      </button>
                      <button
                        v-if="getChallengeRow(row).status !== 'published'"
                        type="button"
                        class="c-action-menu__item c-action-menu__item--success"
                        @click="submitPublishCheck(getChallengeRow(row))"
                      >
                        <CheckCircle class="h-3 w-3" />
                        提交发布检查
                      </button>
                      <button
                        type="button"
                        class="c-action-menu__item c-action-menu__item--danger"
                        @click="removeChallenge(getChallengeRow(row).id)"
                      >
                        <Trash2 class="h-3 w-3" />
                        永久删除
                      </button>
                    </template>
                  </CActionMenu>
                </div>
              </template>
            </WorkspaceDataTable>

            <WorkspaceDirectoryPagination
              :page="page"
              :total-pages="Math.max(1, Math.ceil(total / pageSize))"
              :total="total"
              :total-label="`共 ${total} 道题目`"
              @change-page="changePage"
            />
          </section>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--journal-ink);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
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

.challenge-manage-body {
  min-width: 0;
}

.challenge-filter-grid {
  display: grid;
  gap: var(--space-4);
}

.challenge-filter-field {
  display: grid;
  gap: var(--space-2);
}

.challenge-filter-label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-filter-select {
  width: 100%;
  min-height: 2.75rem;
  padding: 0 var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 500;
  border: 1px solid var(--color-border-default);
  border-radius: 0.95rem;
  background: var(--color-bg-surface);
  color: var(--color-text-primary);
  outline: none;
  transition: all 150ms ease;
}

.challenge-filter-select:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 12%, transparent);
}

.challenge-list {
  --workspace-directory-shell-border: var(--color-border-default);
}

.challenge-table-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 1.4rem;
  padding: 0 0.5rem;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
}

.challenge-table-pill--category {
  background: var(--color-primary-soft);
  color: var(--color-primary);
  border: 1px solid color-mix(in srgb, var(--color-primary) 18%, transparent);
}

.challenge-table-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 15px;
  font-weight: 700;
  color: var(--color-text-primary);
}

.group:hover .challenge-table-title {
  color: var(--color-primary);
}

.challenge-table-difficulty {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.challenge-table-points {
  font-family: var(--font-family-mono);
  font-size: 15px;
  font-weight: 900;
  color: var(--color-text-primary);
}

.challenge-table-status {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.challenge-table-status__dot {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 999px;
}

.challenge-table-status__dot--published {
  background: var(--color-success);
}

.challenge-table-status__dot--idle {
  background: var(--color-border-default);
}

.challenge-table-status__label {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.challenge-table-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.375rem;
}

.challenge-row-action {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  height: 1.85rem;
  padding: 0 0.75rem;
  border: 1px solid var(--color-border-default);
  border-radius: 8px;
  font-size: 12px;
  font-weight: 800;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.challenge-row-action:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-soft);
  color: var(--color-primary);
}
</style>