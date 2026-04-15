<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Book,
  CheckCircle,
  Zap,
  Edit3,
  Search,
  Filter,
  ChevronDown,
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

import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import ChallengePackageImportEntry from '@/components/admin/challenge/ChallengePackageImportEntry.vue'
import { useAdminChallenges } from '@/composables/useAdminChallenges'
import { useChallengeManagePresentation } from '@/composables/useChallengeManagePresentation'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'

type ChallengePanelKey = 'manage' | 'import' | 'queue'

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

const isFilterOpen = ref(false)
const isSortOpen = ref(false)
const filterToggleRef = ref<HTMLButtonElement | null>(null)
const filterPanelRef = ref<HTMLDivElement | null>(null)
const sortButtonRef = ref<HTMLButtonElement | null>(null)
const sortMenuRef = ref<HTMLDivElement | null>(null)

const sortConfig = ref({ key: 'updateTime', order: 'desc', label: '最近更新' })
const sortOptions = [
  { key: 'updateTime', order: 'desc', label: '最近更新', icon: Calendar },
  { key: 'points', order: 'desc', label: '分值由高到低', icon: ArrowDownWideNarrow },
  { key: 'points', order: 'asc', label: '分值由低到高', icon: ArrowUpNarrowWide },
  { key: 'title', order: 'asc', label: '标题 A-Z', icon: SortAsc },
]

function setSort(opt: (typeof sortOptions)[number]) {
  sortConfig.value = opt
  isSortOpen.value = false
}

let removeWindowListeners: (() => void) | null = null

onMounted(() => {
  void refreshQueue()
  const handleClickOutside = (event: MouseEvent) => {
    const target = event.target
    if (!(target instanceof Node)) {
      isSortOpen.value = false
      isFilterOpen.value = false
      return
    }

    const clickedInsideFilter =
      filterToggleRef.value?.contains(target) || filterPanelRef.value?.contains(target)
    const clickedInsideSort =
      sortButtonRef.value?.contains(target) || sortMenuRef.value?.contains(target)

    if (!clickedInsideFilter) {
      isFilterOpen.value = false
    }

    if (!clickedInsideSort) {
      isSortOpen.value = false
    }
  }
  window.addEventListener('click', handleClickOutside)
  removeWindowListeners = () => {
    window.removeEventListener('click', handleClickOutside)
  }
})

onUnmounted(() => {
  removeWindowListeners?.()
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

          <div class="challenge-metric-grid">
            <article class="challenge-metric-card">
              <div class="challenge-metric-head">
                <span class="challenge-metric-label">题目总量</span>
                <Book class="h-4 w-4" />
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="challenge-metric-value">{{ total.toString().padStart(2, '0') }}</div>
                <div class="challenge-metric-trend">题目资源总计</div>
              </div>
            </article>

            <article class="challenge-metric-card">
              <div class="challenge-metric-head">
                <span class="challenge-metric-label">已发布</span>
                <CheckCircle class="h-4 w-4" />
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="challenge-metric-value text-emerald-600">
                  {{ publishedCount.toString().padStart(2, '0') }}
                </div>
                <div class="challenge-metric-trend">线上公开题目</div>
              </div>
            </article>

            <article class="challenge-metric-card">
              <div class="challenge-metric-head">
                <span class="challenge-metric-label">待确认任务</span>
                <Zap class="h-4 w-4" />
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="challenge-metric-value text-purple-600">{{ queueCount.toString().padStart(2, '0') }}</div>
                <div class="challenge-metric-trend">待导入任务队列</div>
              </div>
            </article>

            <article class="challenge-metric-card">
              <div class="challenge-metric-head">
                <span class="challenge-metric-label">草稿存量</span>
                <Edit3 class="h-4 w-4" />
              </div>
              <div class="challenge-metric-value-wrap">
                <div class="challenge-metric-value text-orange-500">{{ draftCount.toString().padStart(2, '0') }}</div>
                <div class="challenge-metric-trend">导入后仍待发布</div>
              </div>
            </article>
          </div>

          <section class="workspace-directory-section challenge-manage-directory">
            <div class="challenge-filter-bar">
              <div class="challenge-filter-main">
                <label class="challenge-search-wrap">
                  <Search class="challenge-search-icon h-3.5 w-3.5" />
                  <input
                    v-model="keyword"
                    type="text"
                    class="challenge-search-input"
                    placeholder="检索题目 ID 或名称..."
                  />
                </label>

                <button
                  ref="filterToggleRef"
                  type="button"
                  class="challenge-filter-toggle"
                  :class="{ 'challenge-filter-toggle--active': isFilterOpen }"
                  @click.stop="
                    ;isFilterOpen = !isFilterOpen
                    ;isSortOpen = false
                  "
                >
                  <Filter class="h-3.5 w-3.5" />
                  筛选
                </button>
              </div>

              <div class="challenge-filter-meta">
                <div class="challenge-sort-wrap">
                  <span class="mr-2 text-[10px] font-bold text-slate-400 uppercase tracking-tight">排序:</span>
                  <button
                    ref="sortButtonRef"
                    type="button"
                    class="challenge-sort-button"
                    @click.stop="
                      ;isSortOpen = !isSortOpen
                      ;isFilterOpen = false
                    "
                  >
                    <span class="font-black tracking-wider">{{ sortConfig.label }}</span>
                    <ChevronDown class="h-3.5 w-3.5 transition-transform" :class="{ 'rotate-180': isSortOpen }" />
                  </button>

                  <div v-if="isSortOpen" ref="sortMenuRef" class="challenge-filter-menu">
                    <div class="challenge-filter-menu__title">Sort Strategy</div>
                    <div class="challenge-filter-menu__list">
                      <button
                        v-for="opt in sortOptions"
                        :key="opt.label"
                        type="button"
                        class="challenge-filter-menu__item"
                        :class="{ 'challenge-filter-menu__item--active': sortConfig.label === opt.label }"
                        @click="setSort(opt)"
                      >
                        <div class="flex items-center gap-2">
                          <component :is="opt.icon" class="h-3.5 w-3.5" />
                          {{ opt.label }}
                        </div>
                      </button>
                    </div>
                  </div>
                </div>

                <div class="challenge-count-pill">
                  共 <span class="font-mono font-black text-slate-900">{{ total }}</span> 项
                </div>
              </div>

              <div v-if="isFilterOpen" ref="filterPanelRef" class="challenge-filter-panel">
                <div class="challenge-filter-panel__header">
                  <div>
                    <div class="workspace-overline">Filter Stack</div>
                    <h3 class="challenge-filter-panel__title">高级筛选</h3>
                  </div>
                  <button
                    type="button"
                    class="challenge-filter-reset"
                    :disabled="!hasActiveFilters"
                    @click="clearFilters"
                  >
                    清空筛选
                  </button>
                </div>

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
              </div>
            </div>

            <div v-if="loading" class="challenge-directory-state">正在同步题目目录...</div>
            <div v-else-if="list.length === 0" class="challenge-directory-state">
              {{ manageEmptyMessage }}
            </div>
            <div v-else class="challenge-table-shell workspace-directory-list">
              <table class="challenge-table">
                <thead class="challenge-table-head">
                  <tr>
                    <th class="w-[30%] min-w-[180px] px-2">题目名称</th>
                    <th class="w-[18%] px-2 text-center">题目 ID</th>
                    <th class="w-24 px-2 text-center">分类</th>
                    <th class="w-20 px-2 text-center">难度</th>
                    <th class="w-20 px-2 text-center">分值</th>
                    <th class="w-32 px-4">状态</th>
                    <th class="w-40 px-2 text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(row, index) in list"
                    :key="row.id"
                    class="challenge-table-row group"
                  >
                    <td class="truncate px-2 py-3.5 font-bold text-slate-900 group-hover:text-blue-600 transition-colors" :title="row.title">
                      {{ row.title }}
                    </td>
                    <td
                      class="truncate px-2 py-3.5 text-center font-mono text-[10px] font-bold uppercase tracking-tighter text-slate-400 group-hover:text-slate-600"
                      :title="row.id"
                    >
                      {{ row.id.split('-').pop()?.substring(0, 10) || row.id }}
                    </td>
                    <td class="px-2 py-3.5 text-center">
                      <span class="challenge-table-pill challenge-table-pill--category">
                        {{ getCategoryLabel(row.category) }}
                      </span>
                    </td>
                    <td class="px-2 py-3.5 text-center">
                      <span class="text-[10px] font-bold uppercase text-slate-500">
                        {{ getDifficultyLabel(row.difficulty) }}
                      </span>
                    </td>
                    <td
                      class="px-2 py-3.5 text-center font-mono text-sm font-black tracking-tighter text-slate-900"
                    >
                      {{ row.points }}
                    </td>
                    <td class="px-4 py-3.5">
                      <div class="flex items-center gap-2">
                        <div
                          class="h-1.5 w-1.5 rounded-full"
                          :class="row.status === 'published' ? 'bg-emerald-500 animate-pulse' : 'bg-slate-300'"
                        />
                        <span class="text-[11px] font-bold uppercase text-slate-700">
                          {{
                            row.status === 'published'
                              ? '已发布'
                              : row.status === 'archived'
                                ? '已归档'
                                : '草稿'
                          }}
                        </span>
                      </div>
                    </td>
                    <td class="relative px-2 py-3.5 text-right">
                      <div class="flex items-center justify-end gap-1.5">
                        <button
                          type="button"
                          class="challenge-row-action"
                          @click="openChallengeDetail(row.id)"
                        >
                          <Eye class="h-3 w-3" />
                          查看
                        </button>

                        <div class="relative inline-block text-left">
                          <button
                            type="button"
                            class="challenge-row-menu-button"
                            :class="{ 'challenge-row-menu-button--active': openActionMenuId === row.id }"
                            @click.stop="toggleActionMenu(row.id)"
                          >
                            <MoreHorizontal class="h-3.5 w-3.5" />
                          </button>

                          <div
                            v-if="openActionMenuId === row.id"
                            class="challenge-row-menu shadow-2xl"
                            :class="
                              index >= list.length - 2 && list.length > 2
                                ? 'challenge-row-menu--up'
                                : 'challenge-row-menu--down'
                            "
                          >
                            <div class="challenge-row-menu__title">Management</div>
                            <button
                              type="button"
                              class="challenge-row-menu__item"
                              @click="openChallengeTopology(row.id)"
                            >
                              <FileSearch class="h-3 w-3" />
                              编排拓扑
                            </button>
                            <button
                              type="button"
                              class="challenge-row-menu__item"
                              @click="openChallengeWriteup(row.id)"
                            >
                              <Book class="h-3 w-3" />
                              题解与提示
                            </button>
                            <button
                              v-if="row.status !== 'published'"
                              type="button"
                              class="challenge-row-menu__item challenge-row-menu__item--success"
                              @click="submitPublishCheck(row)"
                            >
                              <CheckCircle class="h-3 w-3" />
                              提交发布检查
                            </button>
                            <button
                              type="button"
                              class="challenge-row-menu__item challenge-row-menu__item--danger"
                              @click="removeChallenge(row.id)"
                            >
                              <Trash2 class="h-3 w-3" />
                              永久删除
                            </button>
                          </div>
                        </div>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="total > 0" class="workspace-directory-pagination challenge-manage-pagination">
              <AdminPaginationControls
                :page="page"
                :total-pages="Math.max(1, Math.ceil(total / pageSize))"
                :total="total"
                :total-label="`共 ${total} 条`"
                @change-page="changePage"
              />
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

.challenge-metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1rem;
  margin-top: 1.25rem;
  margin-bottom: 2.5rem;
}

.challenge-metric-card {
  position: relative;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: white;
  padding: 1rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  transition: all 0.2s ease;
}

.challenge-metric-card:hover {
  border-color: #2563eb;
  transform: translateY(-1px);
}

.challenge-metric-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #94a3b8;
  margin-bottom: 1rem;
}

.challenge-metric-label {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.challenge-metric-value-wrap {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
}

.challenge-metric-value {
  font-family: inherit;
  font-size: 1.85rem;
  font-weight: 900;
  line-height: 1;
  letter-spacing: -0.05em;
  color: #0f172a;
}

.challenge-metric-trend {
  font-size: 10px;
  font-weight: 700;
  color: #94a3b8;
}

.challenge-filter-bar {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.challenge-filter-main,
.challenge-filter-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.challenge-search-wrap {
  position: relative;
}

.challenge-search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #94a3b8;
}

.challenge-search-input {
  width: 20rem;
  min-height: 2.5rem;
  padding: 0 1rem 0 2.25rem;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid transparent;
  border-radius: 12px;
  background: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  outline: none;
  transition: all 0.2s ease;
}

.challenge-search-input:focus {
  background: white;
  box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.1);
}

.challenge-filter-toggle,
.challenge-sort-button,
.challenge-count-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.5rem;
  padding: 0 1rem;
  border: 1px solid transparent;
  border-radius: 12px;
  background: white;
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  transition: all 0.2s ease;
}

.challenge-filter-toggle--active {
  background: #0f172a;
  color: white;
}

.challenge-sort-button:hover,
.challenge-filter-toggle:hover {
  color: #2563eb;
}

.challenge-filter-panel,
.challenge-filter-menu,
.challenge-row-menu {
  border: 1px solid #e2e8f0;
  background: white;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.1);
}

.challenge-filter-panel {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  z-index: 40;
  width: 24rem;
  border-radius: 16px;
  padding: 1.25rem;
}

.challenge-filter-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.challenge-filter-panel__title {
  font-size: 14px;
  font-weight: 800;
  color: #0f172a;
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

.challenge-filter-reset {
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
}

.challenge-filter-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 0.5rem);
  z-index: 40;
  width: 12rem;
  border-radius: 12px;
  overflow: hidden;
}

.challenge-filter-menu__title,
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

.challenge-filter-menu__item {
  display: flex;
  width: 100%;
  align-items: center;
  padding: 0.65rem 1rem;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  transition: all 0.2s ease;
}

.challenge-filter-menu__item:hover,
.challenge-filter-menu__item--active {
  background: #f8fafc;
  color: #2563eb;
}

.challenge-table-shell {
  border: none;
  background: transparent;
}

.challenge-table {
  width: 100%;
  border-collapse: collapse;
}

.challenge-table-head th {
  padding: 0.75rem 0.5rem;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: #94a3b8;
  border-bottom: 1px solid #e2e8f0;
  text-align: left;
}

.challenge-table-row {
  border-bottom: 1px solid #f1f5f9;
  background: transparent;
  transition: all 0.2s ease;
}

.challenge-table-row:hover {
  background: rgba(226, 232, 240, 0.4);
}

.challenge-table-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 1.4rem;
  padding: 0 0.5rem;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.challenge-table-pill--category {
  background: #eff6ff;
  color: #2563eb;
  border: 1px solid rgba(37, 99, 235, 0.1);
}

.challenge-row-action {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  height: 1.85rem;
  padding: 0 0.75rem;
  border-radius: 8px;
  font-size: 11px;
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

.challenge-row-menu {
  position: absolute;
  right: 0;
  z-index: 50;
  width: 11rem;
  border-radius: 12px;
  overflow: hidden;
}

.challenge-row-menu--down {
  top: calc(100% + 0.4rem);
}

.challenge-row-menu--up {
  bottom: calc(100% + 0.4rem);
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

.challenge-manage-pagination {
  margin-top: 1.5rem;
}

@media (max-width: 1023px) {
  .challenge-metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .challenge-metric-grid {
    grid-template-columns: 1fr;
  }
  .challenge-search-input {
    width: 100%;
  }
}
</style>
