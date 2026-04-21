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
  <div
    class="workspace-shell challenge-manage-shell journal-shell journal-shell-admin journal-notes-card journal-hero"
  >
    <div class="workspace-grid">
      <main class="content-pane challenge-manage-content">
        <section class="challenge-manage-panel">
          <div class="workspace-tab-heading challenge-manage-actions">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">
                Challenge Workspace
              </div>
              <h1 class="workspace-page-title">
                题目资源管理中心
              </h1>
              <p class="workspace-page-copy">
                集中查看题目目录、发布状态与题库变更。
              </p>
            </div>
            <div class="challenge-manage-hero-actions">
              <button
                type="button"
                class="challenge-manage-action challenge-manage-action--primary"
                @click="openImportWorkspace"
              >
                <Plus class="mr-1.5 h-4 w-4" />
                导入资源包
              </button>
            </div>
          </div>

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
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.challenge-manage-shell {
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--journal-muted) 92%, transparent);
  --workspace-line-soft: color-mix(in srgb, var(--journal-border) 82%, transparent);
  --workspace-shell-bg: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  --workspace-page: color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  --workspace-shadow-shell: 0 22px 52px color-mix(in srgb, var(--color-shadow-soft) 54%, transparent);
  --workspace-shadow-panel: 0 14px 34px color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-side-padding: 2rem;
  --workspace-content-padding: 2rem;
  --journal-shell-hero-radial-strength: 10%;
  --journal-shell-hero-radial-size: 20rem;
  --journal-shell-dark-accent: var(--color-primary-hover);
  --challenge-page-bg: var(--workspace-page);
  --challenge-page-surface: var(--workspace-panel);
  --challenge-page-surface-subtle: color-mix(
    in srgb,
    var(--workspace-panel-soft) 94%,
    var(--color-bg-base)
  );
  --challenge-page-surface-elevated: color-mix(
    in srgb,
    var(--workspace-panel) 98%,
    var(--journal-surface)
  );
  --challenge-page-line: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, var(--journal-border) 92%, transparent);
  --challenge-page-text: color-mix(in srgb, var(--journal-ink) 94%, transparent);
  --challenge-page-muted: color-mix(in srgb, var(--journal-muted) 92%, transparent);
  --challenge-page-faint: color-mix(in srgb, var(--journal-muted) 72%, var(--color-bg-base));
  --challenge-page-accent: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  --challenge-directory-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --challenge-directory-row-divider: color-mix(in srgb, var(--journal-border) 58%, transparent);
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --challenge-page-accent-soft: color-mix(
    in srgb,
    var(--workspace-brand) 10%,
    var(--challenge-page-surface)
  );
  background: var(--challenge-page-bg);
}

.challenge-manage-content {
  display: grid;
  gap: var(--space-6);
  background: transparent;
}

.challenge-manage-panel {
  display: grid;
  gap: var(--space-section-gap-compact, var(--space-4));
  min-width: 0;
}

.challenge-manage-actions {
  align-items: flex-end;
}

.challenge-manage-shell .manage-summary-grid {
  width: 100%;
}

.challenge-manage-hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.challenge-manage-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.5rem;
  padding: 0 1.25rem;
  border: 1px solid var(--challenge-page-line);
  border-radius: 12px;
  background: var(--challenge-page-surface);
  font-size: 12px;
  font-weight: 700;
  color: var(--challenge-page-muted);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
  box-shadow: 0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
}

.challenge-manage-action:hover {
  border-color: var(--challenge-page-line-strong);
  background: var(--challenge-page-surface-elevated);
  color: var(--challenge-page-text);
  transform: translateY(-1px);
}

.challenge-manage-action--primary {
  border-color: color-mix(in srgb, var(--workspace-brand) 42%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  color: white;
  box-shadow: 0 10px 24px color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}

.challenge-manage-action--primary:hover {
  color: white;
  background: color-mix(in srgb, var(--workspace-brand-ink) 92%, var(--challenge-page-text));
  border-color: color-mix(in srgb, var(--workspace-brand-ink) 62%, transparent);
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
  border: 1px solid var(--admin-control-border);
  border-radius: 0.95rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
}

.challenge-filter-select:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.challenge-manage-directory {
  display: grid;
  gap: var(--space-4);
}

.challenge-manage-directory :deep(.workspace-directory-toolbar) {
  margin-bottom: 0.5rem;
}

.challenge-list {
  --workspace-directory-shell-border: var(--challenge-directory-border);
  --workspace-directory-head-divider: var(--challenge-directory-border);
  --workspace-directory-row-divider: var(--challenge-directory-row-divider);
}

.challenge-table-row {
  position: relative;
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
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.challenge-table-pill--category {
  background: color-mix(in srgb, var(--workspace-brand) 10%, var(--challenge-page-surface));
  color: var(--challenge-page-accent);
  border: 1px solid color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}

.challenge-table__title-cell {
  min-width: 0;
}

.challenge-table__compact-cell,
.challenge-table__points-cell,
.challenge-table__actions-cell {
  font-size: 13px;
}

.challenge-table-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 15px;
  font-weight: 700;
  color: var(--challenge-page-text);
  transition: color 0.2s ease;
}

.group:hover .challenge-table-title {
  color: var(--challenge-page-accent);
}

.challenge-table-difficulty {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--challenge-page-muted);
}

.challenge-table-points {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 15px;
  font-weight: 900;
  letter-spacing: -0.03em;
  color: var(--challenge-page-text);
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
  background: color-mix(in srgb, var(--color-success) 88%, transparent);
  animation: challengeStatusPulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.challenge-table-status__dot--idle {
  background: color-mix(in srgb, var(--challenge-page-line-strong) 88%, transparent);
}

.challenge-table-status__label {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--challenge-page-text) 84%, var(--challenge-page-muted));
}

.challenge-table-actions {
  position: relative;
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
  border: 1px solid color-mix(in srgb, var(--workspace-brand) 24%, var(--challenge-page-line));
  border-radius: 8px;
  font-size: 12px;
  font-weight: 800;
  background: color-mix(in srgb, var(--workspace-brand) 7%, var(--challenge-page-surface));
  color: var(--challenge-page-accent);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease;
}

.challenge-row-action:hover {
  border-color: color-mix(in srgb, var(--workspace-brand) 42%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  color: white;
  box-shadow: 0 8px 20px color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}

@keyframes challengeStatusPulse {
  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.45;
  }
}

@media (max-width: 767px) {
  .challenge-search-input {
    width: 100%;
  }
}
</style>
