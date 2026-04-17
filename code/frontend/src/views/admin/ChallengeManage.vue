<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { useRouter } from 'vue-router'
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
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useAdminChallenges, type AdminChallengeListRow } from '@/composables/useAdminChallenges'
import { useChallengeManagePresentation } from '@/composables/useChallengeManagePresentation'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'

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
const importWorkspaceRef = ref<HTMLElement | null>(null)
const queueWorkspaceRef = ref<HTMLElement | null>(null)

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

function scrollToWorkspace(section: 'import' | 'queue'): void {
  const target = section === 'import' ? importWorkspaceRef.value : queueWorkspaceRef.value
  target?.scrollIntoView({ behavior: 'smooth', block: 'start' })
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
  <div class="workspace-shell challenge-manage-shell journal-shell journal-shell-admin journal-notes-card">
    <div class="workspace-grid">
      <main class="content-pane challenge-manage-content">
        <section class="challenge-manage-panel">
          <div class="workspace-tab-heading challenge-manage-actions">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Challenge Workspace</div>
              <h1 class="workspace-page-title">题目资源管理中心</h1>
              <p class="workspace-page-copy">集中查看题目目录、发布状态与资源包导入流程。</p>
            </div>
            <div class="challenge-manage-hero-actions">
              <button
                type="button"
                class="challenge-manage-action"
                @click="openPackageFormatGuide"
              >
                <FileSearch class="mr-1.5 h-3.5 w-3.5" />
                题目包规范
              </button>
              <button
                type="button"
                class="challenge-manage-action challenge-manage-action--primary"
                @click="scrollToWorkspace('import')"
              >
                <Plus class="mr-1.5 h-4 w-4" />
                导入资源包
              </button>
              <button
                type="button"
                class="challenge-manage-action"
                @click="scrollToWorkspace('queue')"
              >
                <Zap class="mr-1.5 h-3.5 w-3.5" />
                待确认导入
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

          <section class="workspace-directory-section challenge-manage-directory">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">Challenge Directory</div>
                <h2 class="list-heading__title">题目目录</h2>
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
              :rows="sortedChallenges"
              row-key="id"
              row-class="challenge-table-row group"
            >
              <template #cell-title="{ row }">
                <div class="challenge-table-title" :title="getChallengeRow(row).title">
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

            <WorkspaceDirectoryPagination
              :page="page"
              :total-pages="Math.max(1, Math.ceil(total / pageSize))"
              :total="total"
              :total-label="`共 ${total} 道题目`"
              @change-page="changePage"
            />
          </section>

          <section
            id="challenge-import-workspace"
            ref="importWorkspaceRef"
            class="workspace-directory-section challenge-manage-directory challenge-workspace-section"
          >
            <header class="list-heading challenge-section-heading">
              <div>
                <div class="workspace-overline">Challenge Package</div>
                <h2 class="list-heading__title">导入题目包</h2>
                <p class="challenge-section-copy">
                  上传压缩包后先进入预览，再确认是否写入题库。格式规则单独维护，页面只保留导入主流程。
                </p>
              </div>
              <div class="challenge-manage-hero-actions">
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
            </header>

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
                  class="challenge-upload-result"
                  :class="
                    result.status === 'success'
                      ? 'challenge-upload-result--success'
                      : 'challenge-upload-result--error'
                  "
                >
                  <div class="challenge-upload-result__head">
                    <span
                      class="challenge-upload-result__status"
                      :class="
                        result.status === 'success'
                          ? 'challenge-upload-result__status--success'
                          : 'challenge-upload-result__status--error'
                      "
                    >
                      {{ result.status === 'success' ? '成功' : '失败' }}
                    </span>
                    <strong class="challenge-upload-result__title" :title="result.fileName">
                      {{ result.fileName }}
                    </strong>
                  </div>
                  <p class="challenge-upload-result__copy">{{ result.message }}</p>
                  <div class="challenge-upload-result__meta">
                    <span>{{ formatDateTime(result.createdAt) }}</span>
                    <span v-if="result.code !== undefined">错误码 {{ result.code }}</span>
                    <span v-if="result.requestId">请求ID {{ result.requestId }}</span>
                  </div>
                </article>
              </div>
            </section>
          </section>

          <section
            id="challenge-queue-workspace"
            ref="queueWorkspaceRef"
            class="workspace-directory-section challenge-manage-directory challenge-workspace-section"
          >
            <div class="list-heading challenge-directory-head challenge-section-heading">
              <div>
                <div class="workspace-overline">Import Review</div>
                <h2 class="list-heading__title">待确认导入</h2>
                <p class="challenge-section-copy">
                  这里列出已生成预览、但还没正式导入题库的题目包。确认无误后，可继续查看预览并完成导入。
                </p>
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
                    <h2 class="challenge-queue-title" :title="item.title">
                      {{ item.title }}
                    </h2>
                    <p class="challenge-queue-file" :title="item.file_name">
                      {{ item.file_name }}
                    </p>
                    <div class="mt-3 flex flex-wrap gap-2">
                      <span class="challenge-table-pill challenge-table-pill--category">
                        {{ getCategoryLabel(item.category) }}
                      </span>
                      <span class="challenge-table-pill challenge-table-pill--neutral">
                        {{ getDifficultyLabel(item.difficulty) }}
                      </span>
                      <span class="challenge-queue-points">{{ item.points }} pts</span>
                    </div>
                  </div>
                </div>

                <div class="flex flex-col items-start gap-2 md:items-end">
                  <div class="challenge-queue-time">
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
  --challenge-page-bg: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  --challenge-page-surface: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-surface));
  --challenge-page-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 82%,
    var(--color-bg-surface)
  );
  --challenge-page-surface-elevated: color-mix(
    in srgb,
    var(--color-bg-elevated) 90%,
    var(--color-bg-surface)
  );
  --challenge-page-line: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, var(--journal-border) 92%, transparent);
  --challenge-page-text: color-mix(in srgb, var(--journal-ink) 94%, transparent);
  --challenge-page-muted: color-mix(in srgb, var(--journal-muted) 92%, transparent);
  --challenge-page-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
  --challenge-page-accent: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  --challenge-page-accent-soft: color-mix(
    in srgb,
    var(--workspace-brand) 10%,
    var(--challenge-page-surface)
  );
  --challenge-page-success-soft: color-mix(in srgb, var(--color-success) 12%, var(--challenge-page-surface));
  --challenge-page-danger-soft: color-mix(in srgb, var(--color-danger) 10%, var(--challenge-page-surface));
  border: none;
  background: var(--challenge-page-bg);
}

.challenge-row-menu-button,
.challenge-row-menu {
  --challenge-action-surface: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
  --challenge-action-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 82%,
    var(--color-bg-surface)
  );
  --challenge-action-surface-elevated: color-mix(
    in srgb,
    var(--color-bg-elevated) 90%,
    var(--color-bg-surface)
  );
  --challenge-action-line: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --challenge-action-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --challenge-action-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --challenge-action-muted: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
  --challenge-action-accent: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-action-text));
  --challenge-action-accent-soft: color-mix(
    in srgb,
    var(--workspace-brand) 10%,
    var(--challenge-action-surface)
  );
}

.challenge-manage-content {
  display: grid;
  gap: 2rem;
  background: transparent;
}

.challenge-manage-panel {
  display: grid;
  gap: 2rem;
  min-width: 0;
}

.challenge-manage-actions {
  align-items: flex-end;
}

.challenge-manage-hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.challenge-workspace-section {
  scroll-margin-top: 6rem;
}

.challenge-section-heading {
  align-items: flex-start;
  gap: 1rem;
}

.challenge-section-copy {
  margin: 0.5rem 0 0;
  max-width: 44rem;
  font-size: 0.92rem;
  line-height: 1.6;
  color: var(--challenge-page-muted);
}

.challenge-panel-stack {
  display: grid;
  gap: 1rem;
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
  color: var(--challenge-page-faint);
}

.challenge-filter-select {
  width: 100%;
  min-height: 2.25rem;
  padding: 0 0.75rem;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid var(--challenge-page-line);
  border-radius: 8px;
  background: var(--challenge-page-surface-subtle);
  color: var(--challenge-page-text);
}

.challenge-row-menu__title {
  padding: 0.75rem 1rem 0.5rem;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--challenge-action-muted);
  background: color-mix(in srgb, var(--workspace-brand) 5%, var(--challenge-action-surface-subtle));
  border-bottom: 1px solid color-mix(in srgb, var(--challenge-action-line) 78%, transparent);
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

.challenge-table-pill--neutral {
  background: color-mix(in srgb, var(--challenge-page-line) 18%, var(--challenge-page-surface));
  color: var(--challenge-page-muted);
  border: 1px solid color-mix(in srgb, var(--challenge-page-line-strong) 78%, transparent);
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
  background: #10b981;
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

.challenge-row-menu-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.85rem;
  height: 1.85rem;
  border: 1px solid var(--challenge-action-line);
  border-radius: 8px;
  background: var(--challenge-action-surface);
  color: var(--challenge-action-muted);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    box-shadow 0.2s ease;
}

.challenge-row-menu-button:hover,
.challenge-row-menu-button--active {
  border-color: color-mix(in srgb, var(--workspace-brand) 26%, var(--challenge-action-line-strong));
  background: var(--challenge-action-accent-soft);
  color: var(--challenge-action-accent);
  box-shadow: 0 12px 26px color-mix(in srgb, var(--workspace-brand) 12%, transparent);
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
  border: 1px solid var(--challenge-action-line);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--challenge-action-surface) 98%, transparent),
    color-mix(in srgb, var(--challenge-action-surface-subtle) 96%, transparent)
  );
  box-shadow:
    0 24px 60px color-mix(in srgb, var(--color-shadow-strong) 20%, transparent),
    0 10px 24px color-mix(in srgb, var(--color-shadow-soft) 18%, transparent);
}

.challenge-row-menu__item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.5rem;
  padding: 0.65rem 1rem;
  font-size: 12px;
  font-weight: 600;
  color: var(--challenge-action-text);
  transition:
    background-color 0.2s ease,
    color 0.2s ease;
}

.challenge-row-menu__item:hover {
  background: color-mix(in srgb, var(--workspace-brand) 7%, var(--challenge-action-surface-subtle));
  color: var(--challenge-action-accent);
}

.challenge-row-menu__item--success {
  color: color-mix(in srgb, var(--color-success) 84%, var(--challenge-action-text));
}

.challenge-row-menu__item--success:hover {
  background: color-mix(in srgb, var(--color-success) 10%, var(--challenge-action-surface-subtle));
  color: color-mix(in srgb, var(--color-success) 92%, var(--challenge-action-text));
}

.challenge-row-menu__item--danger {
  color: color-mix(in srgb, var(--color-danger) 88%, var(--challenge-action-text));
}

.challenge-row-menu__item--danger:hover {
  background: color-mix(in srgb, var(--color-danger) 10%, var(--challenge-action-surface-subtle));
  color: color-mix(in srgb, var(--color-danger) 96%, var(--challenge-action-text));
}

.challenge-directory-state,
.challenge-directory-meta,
.challenge-queue-file,
.challenge-queue-time,
.challenge-upload-result__copy,
.challenge-upload-result__meta {
  color: var(--challenge-page-muted);
}

.challenge-queue-title,
.challenge-upload-result__title {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 1rem;
  font-weight: 700;
  color: var(--challenge-page-text);
}

.challenge-queue-file {
  margin: 0.25rem 0 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.875rem;
}

.challenge-queue-id {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 5.5rem;
  height: 2rem;
  padding: 0 0.75rem;
  border: 1px solid color-mix(in srgb, var(--workspace-brand) 18%, var(--challenge-page-line));
  border-radius: 999px;
  background: color-mix(in srgb, var(--workspace-brand) 9%, var(--challenge-page-surface));
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--challenge-page-accent);
}

.challenge-queue-points {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--challenge-page-muted);
}

.challenge-upload-result {
  display: grid;
  gap: 0.5rem;
  padding: 1rem;
  border: 1px solid var(--challenge-page-line);
  border-radius: 1rem;
  background: var(--challenge-page-surface);
}

.challenge-upload-result--success {
  border-color: color-mix(in srgb, var(--color-success) 24%, var(--challenge-page-line));
  background: color-mix(in srgb, var(--color-success) 9%, var(--challenge-page-surface));
}

.challenge-upload-result--error {
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--challenge-page-line));
  background: color-mix(in srgb, var(--color-danger) 8%, var(--challenge-page-surface));
}

.challenge-upload-result__head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.challenge-upload-result__status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.challenge-upload-result__status--success {
  background: color-mix(in srgb, var(--color-success) 16%, transparent);
  color: color-mix(in srgb, var(--color-success) 92%, var(--challenge-page-text));
}

.challenge-upload-result__status--error {
  background: color-mix(in srgb, var(--color-danger) 14%, transparent);
  color: color-mix(in srgb, var(--color-danger) 92%, var(--challenge-page-text));
}

.challenge-upload-result__copy {
  margin: 0;
  font-size: 0.8rem;
}

.challenge-upload-result__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.65rem;
  font-weight: 600;
}

:global([data-theme='light']) .challenge-manage-shell {
  --challenge-page-bg: #f8fafc;
  --challenge-page-surface: white;
  --challenge-page-surface-subtle: #f8fafc;
  --challenge-page-surface-elevated: white;
  --challenge-page-line: color-mix(in srgb, #e2e8f0 90%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, #d4dde8 94%, transparent);
  --challenge-page-text: #0f172a;
  --challenge-page-muted: #64748b;
  --challenge-page-faint: #94a3b8;
}

:global([data-theme='dark']) .challenge-manage-shell {
  --challenge-page-bg: color-mix(in srgb, var(--color-bg-base) 92%, var(--color-bg-surface));
  --challenge-page-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --challenge-page-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-surface));
  --challenge-page-surface-elevated: color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-surface));
  --challenge-page-line: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --challenge-page-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --challenge-page-muted: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
  --challenge-page-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
}

:global([data-theme='dark']) .challenge-row-menu,
:global([data-theme='dark']) .challenge-row-menu-button {
  --challenge-action-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --challenge-action-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 84%,
    var(--color-bg-surface)
  );
  --challenge-action-surface-elevated: color-mix(
    in srgb,
    var(--color-bg-elevated) 92%,
    var(--color-bg-surface)
  );
  --challenge-action-line: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --challenge-action-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --challenge-action-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --challenge-action-muted: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
}

:global([data-theme='dark']) .challenge-row-menu {
  box-shadow:
    0 22px 56px color-mix(in srgb, var(--color-shadow-strong) 28%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-border-subtle) 46%, transparent);
}

:global([data-theme='dark']) .challenge-row-menu-button {
  background: var(--challenge-action-surface-elevated);
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
