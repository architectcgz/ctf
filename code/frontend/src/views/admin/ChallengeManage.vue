<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Book,
  CheckCircle,
  Zap,
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

import ChallengePackageImportEntry from '@/components/admin/challenge/ChallengePackageImportEntry.vue'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useAdminChallenges, type AdminChallengeListRow } from '@/composables/useAdminChallenges'
import { useChallengeManagePresentation } from '@/composables/useChallengeManagePresentation'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'

type ChallengePanelKey = 'manage' | 'import' | 'queue'
type ChallengeSortOption = WorkspaceDirectorySortOption & {
  order: 'asc' | 'desc'
}

const panelTabs: Array<{ key: ChallengePanelKey; label: string; panelId: string; tabId: string }> =
  [
    {
      key: 'manage',
      label: '题目管理',
      panelId: 'challenge-panel-manage',
      tabId: 'challenge-tab-manage',
    },
    {
      key: 'import',
      label: '导入题目包',
      panelId: 'challenge-panel-import',
      tabId: 'challenge-tab-import',
    },
    {
      key: 'queue',
      label: '待确认导入',
      panelId: 'challenge-panel-queue',
      tabId: 'challenge-tab-queue',
    },
  ]

const route = useRoute()
const router = useRouter()

const {
  list,
  total,
  page,
  pageSize,
  loading,
  keyword,
  categoryFilter,
  difficultyFilter,
  statusFilter,
  clearFilters,
  changePage,
  publish,
  remove,
} = useAdminChallenges()

const {
  uploading,
  queueLoading,
  selectedFileName,
  queue,
  uploadResults,
  refreshQueue,
  selectPackages,
} = useChallengePackageImport()

const publishedCount = computed(
  () => list.value.filter((item) => item.status === 'published').length
)
const draftCount = computed(() => list.value.filter((item) => item.status === 'draft').length)
const hasActiveFilters = computed(() =>
  Boolean(
    keyword.value.trim() || categoryFilter.value || difficultyFilter.value || statusFilter.value
  )
)
const manageEmptyMessage = computed(() =>
  hasActiveFilters.value ? '当前筛选条件下没有匹配题目。' : '当前还没有题目，请先导入题目包。'
)

const panelTabOrder = panelTabs.map((tab) => tab.key) as ChallengePanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<ChallengePanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'manage',
  routeName: 'ChallengeManage',
})

const queueCount = computed(() => queue.value.length)
const {
  openActionMenuId,
  getCategoryLabel,
  getDifficultyLabel,
  formatDateTime,
  inspectImportTask,
  toggleActionMenu,
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

const actionMenuPanelRef = ref<HTMLDivElement | null>(null)
const actionMenuStyle = ref<Record<string, string>>({})
const actionMenuButtonRefs = new Map<string, HTMLButtonElement>()

const sortConfig = ref({ key: 'updateTime', order: 'desc', label: '最近更新' })
const sortOptions: ChallengeSortOption[] = [
  { key: 'updateTime', order: 'desc', label: '最近更新', icon: Calendar },
  { key: 'points', order: 'desc', label: '分值由高到低', icon: ArrowDownWideNarrow },
  { key: 'points', order: 'asc', label: '分值由低到高', icon: ArrowUpNarrowWide },
  { key: 'title', order: 'asc', label: '标题 A-Z', icon: SortAsc },
]
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
const activeActionRow = computed(() =>
  list.value.find((item) => item.id === openActionMenuId.value) ?? null
)

function setSort(option: WorkspaceDirectorySortOption) {
  const matchedOption =
    sortOptions.find((item) => item.key === option.key && item.label === option.label) ??
    sortOptions[0]

  if (!matchedOption) {
    return
  }

  sortConfig.value = matchedOption
}

function setActionMenuButtonRef(
  challengeId: string,
  element: Element | ComponentPublicInstance | null
): void {
  if (element instanceof HTMLButtonElement) {
    actionMenuButtonRefs.set(challengeId, element)
    return
  }

  actionMenuButtonRefs.delete(challengeId)
}

function updateActionMenuPosition(): void {
  if (!openActionMenuId.value) {
    return
  }

  const trigger = actionMenuButtonRefs.get(openActionMenuId.value)
  if (!trigger) {
    return
  }

  const rect = trigger.getBoundingClientRect()
  const viewportPadding = 12
  const gap = 8
  const panelWidth = actionMenuPanelRef.value?.offsetWidth ?? 176
  const panelHeight = actionMenuPanelRef.value?.offsetHeight ?? 220
  const maxLeft = Math.max(viewportPadding, window.innerWidth - panelWidth - viewportPadding)
  const left = Math.min(Math.max(viewportPadding, rect.right - panelWidth), maxLeft)
  const spaceBelow = window.innerHeight - rect.bottom - viewportPadding
  const spaceAbove = rect.top - viewportPadding
  const shouldOpenUpward = spaceBelow < panelHeight + gap && spaceAbove > spaceBelow
  const maxTop = Math.max(viewportPadding, window.innerHeight - panelHeight - viewportPadding)
  const top = shouldOpenUpward
    ? Math.max(viewportPadding, rect.top - panelHeight - gap)
    : Math.min(rect.bottom + gap, maxTop)

  actionMenuStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
    width: `${panelWidth}px`,
  }
}

async function handleActionMenuToggle(challengeId: string): Promise<void> {
  const shouldOpen = openActionMenuId.value !== challengeId
  toggleActionMenu(challengeId)

  if (!shouldOpen) {
    actionMenuStyle.value = {}
    return
  }

  await nextTick()
  updateActionMenuPosition()
}

onMounted(() => {
  void refreshQueue()
})

watch(openActionMenuId, async (challengeId, _previousId, onCleanup) => {
  if (!challengeId) {
    actionMenuStyle.value = {}
    return
  }

  await nextTick()
  updateActionMenuPosition()

  const handleViewportChange = () => {
    updateActionMenuPosition()
  }
  const handleEscape = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      closeActionMenu()
    }
  }

  window.addEventListener('resize', handleViewportChange)
  window.addEventListener('scroll', handleViewportChange, true)
  window.addEventListener('keydown', handleEscape)

  onCleanup(() => {
    window.removeEventListener('resize', handleViewportChange)
    window.removeEventListener('scroll', handleViewportChange, true)
    window.removeEventListener('keydown', handleEscape)
  })
})

async function handleSelectPackage(files: File[]) {
  const selectedPreview = await selectPackages(files, { parallel: files.length > 1 })
  if (!selectedPreview?.id) {
    return
  }
  await router.push({
    name: 'AdminChallengeImportPreview',
    params: { importId: selectedPreview.id },
  })
}

async function openPackageFormatGuide(): Promise<void> {
  await router.push({ name: 'AdminChallengePackageFormat' })
}

function handleTabChange(key: ChallengePanelKey) {
  switchPanel(key)
}

function resolveChallengeCategoryLabel(value: unknown): string {
  return getCategoryLabel(String(value) as never)
}

function resolveChallengeDifficultyLabel(value: unknown): string {
  return getDifficultyLabel(String(value) as never)
}

function getChallengeRow(row: unknown): AdminChallengeListRow {
  return row as AdminChallengeListRow
}
</script>

<template>
  <div class="workspace-shell challenge-manage-shell">
    <nav class="top-tabs" role="tablist" aria-label="题目管理工作区切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.tabId"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: activePanel === tab.key }"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="activePanel === tab.key ? 0 : -1"
        @click="handleTabChange(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <div class="workspace-grid">
      <main class="content-pane challenge-manage-content">
        <section
          id="challenge-panel-manage"
          v-show="activePanel === 'manage'"
          class="tab-panel challenge-manage-panel"
          :class="{ active: activePanel === 'manage' }"
          role="tabpanel"
          aria-labelledby="challenge-tab-manage"
          :aria-hidden="activePanel === 'manage' ? 'false' : 'true'"
        >
          <div class="workspace-tab-heading challenge-manage-actions">
          <div class="workspace-tab-heading__main">
            <h1 class="workspace-page-title">题目资源管理中心</h1>
            <p class="workspace-page-copy uppercase tracking-wider font-bold text-[10px] text-slate-400 mt-1">
              Inventory / Challenge Management
            </p>
          </div>
            <div class="challenge-manage-hero-actions">
              <button
                type="button"
                class="challenge-manage-action"
                @click="openPackageFormatGuide"
              >
                <FileSearch class="h-3.5 w-3.5 mr-1.5" />
                审计日志
              </button>
              <button
                type="button"
                class="challenge-manage-action challenge-manage-action--primary"
                @click="handleTabChange('import')"
              >
                <Plus class="h-4 w-4 mr-1.5" />
                导入资源包
              </button>
            </div>
          </div>

          <div class="manage-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
            <article class="journal-note progress-card metric-panel-card">
              <div class="challenge-metric-head">
                <span class="journal-note-label progress-card-label metric-panel-label">题目总量</span>
                <Book class="h-4 w-4" />
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
                <span class="journal-note-label progress-card-label metric-panel-label">已发布</span>
                <CheckCircle class="h-4 w-4" />
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
                <span class="journal-note-label progress-card-label metric-panel-label">待确认任务</span>
                <Zap class="h-4 w-4" />
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="journal-note-value progress-card-value metric-panel-value">
                  {{ queueCount.toString().padStart(2, '0') }}
                </div>
                <div class="journal-note-helper progress-card-hint metric-panel-helper">
                  待导入任务队列
                </div>
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="challenge-metric-head">
                <span class="journal-note-label progress-card-label metric-panel-label">草稿存量</span>
                <Edit3 class="h-4 w-4" />
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
          </div>

          <section class="workspace-directory-section">
            <div class="challenge-manage-directory">
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
                    <select v-model="categoryFilter" class="challenge-filter-select">
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
                    <select v-model="difficultyFilter" class="challenge-filter-select">
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
                    <select v-model="statusFilter" class="challenge-filter-select">
                      <option value="">全部状态</option>
                      <option value="draft">草稿</option>
                      <option value="published">已发布</option>
                      <option value="archived">已归档</option>
                    </select>
                  </label>
                </div>
              </template>
              </WorkspaceDirectoryToolbar>

              <div v-if="loading" class="challenge-directory-state">正在同步题目目录...</div>
              <div v-else-if="list.length === 0" class="challenge-directory-state">
                {{ manageEmptyMessage }}
              </div>
              <WorkspaceDataTable
                v-else
                class="challenge-list workspace-directory-list"
                :columns="challengeTableColumns"
                :rows="list"
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

              <template #cell-actions="{ row, index }">
                <div class="challenge-table-actions">
                  <button
                    type="button"
                    class="challenge-row-action"
                    @click="openChallengeDetail(getChallengeRow(row).id)"
                  >
                    <Eye class="h-3 w-3" />
                    查看
                  </button>

                  <div class="relative inline-block text-left">
                    <button
                      :ref="(element) => setActionMenuButtonRef(getChallengeRow(row).id, element)"
                      type="button"
                      class="challenge-row-menu-button"
                      :aria-expanded="
                        openActionMenuId === getChallengeRow(row).id ? 'true' : 'false'
                      "
                      aria-haspopup="menu"
                      :class="{
                        'challenge-row-menu-button--active':
                          openActionMenuId === getChallengeRow(row).id,
                      }"
                      @click.stop="void handleActionMenuToggle(getChallengeRow(row).id)"
                    >
                      <MoreHorizontal class="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>
              </template>
              </WorkspaceDataTable>

              <Teleport to="body">
                <div
                  v-if="activeActionRow"
                  class="challenge-row-menu-layer"
                  @click="closeActionMenu"
                >
                  <div
                    ref="actionMenuPanelRef"
                    class="challenge-row-menu shadow-2xl"
                    :style="actionMenuStyle"
                    role="menu"
                    aria-label="题目更多操作"
                    @click.stop
                  >
                    <div class="challenge-row-menu__title">Management</div>
                    <button
                      type="button"
                      class="challenge-row-menu__item"
                      @click="openChallengeTopology(activeActionRow.id)"
                    >
                      <FileSearch class="h-3 w-3" />
                      编排拓扑
                    </button>
                    <button
                      type="button"
                      class="challenge-row-menu__item"
                      @click="openChallengeWriteup(activeActionRow.id)"
                    >
                      <Book class="h-3 w-3" />
                      题解与提示
                    </button>
                    <button
                      v-if="activeActionRow.status !== 'published'"
                      type="button"
                      class="challenge-row-menu__item challenge-row-menu__item--success"
                      @click="submitPublishCheck(activeActionRow)"
                    >
                      <CheckCircle class="h-3 w-3" />
                      提交发布检查
                    </button>
                    <button
                      type="button"
                      class="challenge-row-menu__item challenge-row-menu__item--danger"
                      @click="removeChallenge(activeActionRow.id)"
                    >
                      <Trash2 class="h-3 w-3" />
                      永久删除
                    </button>
                  </div>
                </div>
              </Teleport>

              <div class="admin-pagination workspace-directory-pagination">
                <AdminPaginationControls
                  :page="page"
                  :total-pages="Math.max(1, Math.ceil(total / pageSize))"
                  :total="total"
                  :total-label="`共 ${total} 条`"
                  @change-page="changePage"
                />
              </div>
            </div>
          </section>
        </section>

        <section
          id="challenge-panel-import"
          v-show="activePanel === 'import'"
          class="tab-panel challenge-manage-panel"
          :class="{ active: activePanel === 'import' }"
          role="tabpanel"
          aria-labelledby="challenge-tab-import"
          :aria-hidden="activePanel === 'import' ? 'false' : 'true'"
        >
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Challenge Package</div>
            <h1 class="workspace-page-title">导入题目包</h1>
            <p class="workspace-page-copy">
              上传压缩包后先进入预览，再确认是否写入题库。页面内只保留导入主流程，格式规则单独维护。
            </p>
          </div>

          <div class="challenge-panel-stack">
            <section class="challenge-plain-section">
              <div class="list-heading">
                <div>
                  <div class="workspace-overline">Package Guide</div>
                  <h2 class="list-heading__title">题目包示例</h2>
                </div>
              </div>

              <p class="workspace-page-copy">
                导入页只保留上传和预览流程，目录结构与 `challenge.yml` 示例统一放到独立说明页，避免同一份规则重复维护。
              </p>

              <div class="challenge-manage-hero-actions mt-4">
                <a
                  class="challenge-manage-action"
                  href="/downloads/challenge-package-sample-v1.zip"
                  download="challenge-package-sample-v1.zip"
                >
                  下载示例题目包
                </a>
                <button
                  type="button"
                  class="challenge-manage-action"
                  @click="openPackageFormatGuide"
                >
                  查看题目包示例
                </button>
              </div>
            </section>

            <ChallengePackageImportEntry
              :hide-header="true"
              :uploading="uploading"
              :selected-file-name="selectedFileName"
              @select="handleSelectPackage"
            />

            <section
              v-if="uploadResults.length > 0"
              class="workspace-directory-section challenge-plain-section"
            >
              <div class="list-heading">
                <div>
                  <div class="workspace-overline">Upload Receipt</div>
                  <h2 class="list-heading__title">最近上传结果</h2>
                </div>
              </div>

              <div class="challenge-panel-stack">
                <article
                  v-for="result in uploadResults"
                  :key="result.id"
                  class="rounded-2xl border px-4 py-4"
                  :class="
                    result.status === 'success'
                      ? 'border-emerald-200 bg-emerald-50/70'
                      : 'border-red-200 bg-red-50/70'
                  "
                >
                  <div class="mb-2 flex items-center gap-2">
                    <span
                      class="rounded-full px-2 py-0.5 text-xs font-bold"
                      :class="
                        result.status === 'success'
                          ? 'bg-emerald-100 text-emerald-700'
                          : 'bg-red-100 text-red-700'
                      "
                    >
                      {{ result.status === 'success' ? '成功' : '失败' }}
                    </span>
                    <strong class="truncate text-sm text-slate-900" :title="result.fileName">
                      {{ result.fileName }}
                    </strong>
                  </div>
                  <p class="mb-2 text-xs text-slate-600">{{ result.message }}</p>
                  <div class="flex flex-wrap gap-4 text-[10px] font-medium text-slate-400">
                    <span>{{ formatDateTime(result.createdAt) }}</span>
                    <span v-if="result.code !== undefined">错误码 {{ result.code }}</span>
                    <span v-if="result.requestId">请求ID {{ result.requestId }}</span>
                  </div>
                </article>
              </div>
            </section>
          </div>
        </section>

        <section
          id="challenge-panel-queue"
          v-show="activePanel === 'queue'"
          class="tab-panel challenge-manage-panel"
          :class="{ active: activePanel === 'queue' }"
          role="tabpanel"
          aria-labelledby="challenge-tab-queue"
          :aria-hidden="activePanel === 'queue' ? 'false' : 'true'"
        >
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Import Review</div>
            <h1 class="workspace-page-title">待确认导入</h1>
            <p class="workspace-page-copy">
              这里列出已生成预览、但还没正式导入题库的题目包。确认无误后，可继续查看预览并完成导入。
            </p>
          </div>

          <section class="workspace-directory-section challenge-manage-directory">
            <div class="list-heading challenge-directory-head">
              <div>
                <div class="workspace-overline">Import Queue</div>
                <h2 class="list-heading__title">待确认目录</h2>
              </div>
              <div class="challenge-directory-meta">共 {{ queueCount }} 个待处理任务</div>
            </div>

            <div v-if="queueLoading" class="challenge-directory-state">正在同步导入队列...</div>
            <div v-else-if="queue.length === 0" class="challenge-directory-state">
              当前没有待确认的导入任务。
            </div>

            <div v-else class="challenge-panel-stack">
              <article
                v-for="item in queue"
                :key="item.id"
                class="challenge-plain-section challenge-queue-item"
              >
                <div class="flex min-w-0 items-start gap-4">
                  <div class="challenge-queue-id">IMP-{{ item.id.slice(0, 6).toUpperCase() }}</div>
                  <div class="min-w-0 flex-1">
                    <h2 class="truncate text-base font-bold text-slate-900" :title="item.title">
                      {{ item.title }}
                    </h2>
                    <p class="mt-1 truncate text-sm text-slate-500" :title="item.file_name">
                      {{ item.file_name }}
                    </p>
                    <div class="mt-3 flex flex-wrap gap-2">
                      <span class="challenge-table-pill challenge-table-pill--category">
                        {{ getCategoryLabel(item.category) }}
                      </span>
                      <span class="challenge-table-pill challenge-table-pill--neutral">
                        {{ getDifficultyLabel(item.difficulty) }}
                      </span>
                      <span class="text-[11px] font-mono text-slate-500">{{ item.points }} pts</span>
                    </div>
                  </div>
                </div>

                <div class="flex flex-col items-start gap-2 md:items-end">
                  <div class="text-[10px] font-medium text-slate-400">
                    {{ formatDateTime(item.created_at) }}
                  </div>
                  <button
                    type="button"
                    class="challenge-manage-action challenge-manage-action--primary"
                    @click="inspectImportTask(item)"
                  >
                    继续查看预览
                  </button>
                </div>
              </article>
            </div>
          </section>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.challenge-manage-shell {
  --workspace-brand: #2563eb;
  --workspace-brand-ink: #1e40af;
  --workspace-brand-soft: #eff6ff;
  --workspace-faint: #f8fafc;
  --workspace-shell-border: #e2e8f0;
  --workspace-shadow-shell: 0 1px 3px rgba(0, 0, 0, 0.05);
  --workspace-side-padding: 2rem;
  --workspace-content-padding: 2rem;
  --workspace-tabs-offset-top: 1rem;
  background: #f8fafc;
}

.challenge-manage-top-note {
  font-size: 12px;
}

.challenge-manage-content {
  display: grid;
  gap: 1.5rem;
  background: #f8fafc;
}

.challenge-manage-panel {
  min-width: 0;
}

.challenge-manage-actions {
  align-items: flex-end;
  margin-bottom: 1.5rem;
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
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: white;
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  transition: all 0.2s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.challenge-manage-action:hover {
  border-color: #cbd5e1;
  color: #0f172a;
  transform: translateY(-1px);
}

.challenge-manage-action--primary {
  border-color: #2563eb;
  background: #2563eb;
  color: white;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
}

.challenge-manage-action--primary:hover {
  color: white;
  background: #1d4ed8;
  border-color: #1d4ed8;
}

.challenge-manage-shell .manage-summary-grid {
  --admin-summary-grid-columns: repeat(4, minmax(0, 1fr));
  margin-top: 1.25rem;
  margin-bottom: 2.5rem;
}

.challenge-manage-directory {
  display: grid;
  gap: 1.5rem;
}

.metric-panel-card.progress-card {
  position: relative;
  overflow: hidden;
  transition: all 0.2s ease;
}

.metric-panel-card.progress-card:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 32%, var(--metric-panel-border));
  transform: translateY(-1px);
}

.challenge-metric-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
  color: var(--journal-muted);
}

.challenge-metric-value-wrap {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.manage-summary-grid > :nth-child(2) .metric-panel-value {
  color: var(--color-success);
}

.manage-summary-grid > :nth-child(3) .metric-panel-value {
  color: color-mix(in srgb, #7c3aed 78%, var(--journal-ink));
}

.manage-summary-grid > :nth-child(4) .metric-panel-value {
  color: color-mix(in srgb, #f97316 82%, var(--journal-ink));
}

.challenge-row-menu {
  border: 1px solid #e2e8f0;
  background: white;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.1);
}

.challenge-filter-grid {
  display: grid;
  gap: 1rem;
}

.challenge-filter-field {
  display: grid;
  gap: 0.35rem;
}

.challenge-filter-label {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: #94a3b8;
}

.challenge-filter-select {
  width: 100%;
  min-height: 2.25rem;
  padding: 0 0.75rem;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #f8fafc;
}

.challenge-row-menu__title {
  padding: 0.75rem 1rem 0.5rem;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: #94a3b8;
  background: #f8fafc;
  border-bottom: 1px solid #f1f5f9;
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
  background: #eff6ff;
  color: #2563eb;
  border: 1px solid rgba(37, 99, 235, 0.1);
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
  color: #0f172a;
  transition: color 0.2s ease;
}

.group:hover .challenge-table-title {
  color: #2563eb;
}

.challenge-table-difficulty {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: #64748b;
}

.challenge-table-points {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 15px;
  font-weight: 900;
  letter-spacing: -0.03em;
  color: #0f172a;
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
  background: #10b981;
  animation: challengeStatusPulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.challenge-table-status__dot--idle {
  background: #cbd5e1;
}

.challenge-table-status__label {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  color: #334155;
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
  border-radius: 8px;
  font-size: 12px;
  font-weight: 800;
  color: #2563eb;
  transition: all 0.2s ease;
}

.challenge-row-action:hover {
  background: #2563eb;
  color: white;
  box-shadow: 0 4px 10px rgba(37, 99, 235, 0.2);
}

.challenge-row-menu-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.85rem;
  height: 1.85rem;
  border-radius: 8px;
  color: #94a3b8;
  transition: all 0.2s ease;
}

.challenge-row-menu-button:hover,
.challenge-row-menu-button--active {
  background: #0f172a;
  color: white;
  box-shadow: 0 4px 10px rgba(15, 23, 42, 0.2);
}

.challenge-row-menu-layer {
  position: fixed;
  inset: 0;
  z-index: 120;
}

.challenge-row-menu {
  position: fixed;
  z-index: 130;
  width: 11rem;
  border-radius: 12px;
  overflow: hidden;
}

.challenge-row-menu__item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1rem;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  transition: all 0.2s ease;
}

.challenge-row-menu__item:hover {
  background: #f8fafc;
  color: #2563eb;
}

.challenge-row-menu__item--danger {
  color: #ef4444;
}

.challenge-row-menu__item--danger:hover {
  background: #fef2f2;
  color: #dc2626;
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

@media (max-width: 1023px) {
  .challenge-manage-shell .manage-summary-grid {
    --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .challenge-manage-shell .manage-summary-grid {
    --admin-summary-grid-columns: 1fr;
  }
  .challenge-search-input {
    width: 100%;
  }
}
</style>
