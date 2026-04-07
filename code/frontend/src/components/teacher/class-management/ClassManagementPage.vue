<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowRight, FolderKanban, Search } from 'lucide-vue-next'

import type { TeacherClassItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  changePage: [page: number]
  openDashboard: []
  openReportExport: []
  openClass: [className: string]
}>()

const filterQuery = ref('')

type WorkspaceTab = 'overview' | 'directory'

interface WorkspaceTabItem {
  key: WorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

const workspaceTabs: WorkspaceTabItem[] = [
  {
    key: 'overview',
    label: '总览',
    buttonId: 'class-manage-tab-overview',
    panelId: 'class-manage-overview',
  },
  {
    key: 'directory',
    label: '班级列表',
    buttonId: 'class-manage-tab-directory',
    panelId: 'class-manage-directory',
  },
]

const activeTab = ref<WorkspaceTab>('overview')
const tabButtonRefs: Partial<Record<WorkspaceTab, HTMLButtonElement | null>> = {}

function setTabButtonRef(tab: WorkspaceTab, element: HTMLButtonElement | null): void {
  tabButtonRefs[tab] = element
}

function selectTab(tab: WorkspaceTab): void {
  activeTab.value = tab
}

function focusTab(tab: WorkspaceTab): void {
  tabButtonRefs[tab]?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (
    event.key !== 'ArrowRight' &&
    event.key !== 'ArrowLeft' &&
    event.key !== 'Home' &&
    event.key !== 'End'
  ) {
    return
  }

  event.preventDefault()

  if (event.key === 'Home') {
    selectTab(workspaceTabs[0].key)
    focusTab(workspaceTabs[0].key)
    return
  }

  if (event.key === 'End') {
    const lastTab = workspaceTabs[workspaceTabs.length - 1]
    selectTab(lastTab.key)
    focusTab(lastTab.key)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + workspaceTabs.length) % workspaceTabs.length
  const nextTab = workspaceTabs[nextIndex]
  selectTab(nextTab.key)
  focusTab(nextTab.key)
}

const classEntries = computed(() =>
  props.classes.map((item, index) => ({
    item,
    code: `CL-${String(index + 1).padStart(2, '0')}`,
  }))
)

const filteredClassEntries = computed(() => {
  const keyword = filterQuery.value.trim().toLowerCase()
  if (!keyword) return classEntries.value

  return classEntries.value.filter(({ item, code }) => {
    return code.toLowerCase().includes(keyword) || item.name.toLowerCase().includes(keyword)
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / Math.max(props.pageSize, 1))))
const currentPageStudentCount = computed(() =>
  props.classes.reduce((sum, item) => sum + (item.student_count || 0), 0)
)
</script>

<template>
  <div class="teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
    >
      <div class="teacher-page">
        <nav class="top-tabs" role="tablist" aria-label="班级管理标签页">
          <button
            v-for="(tab, index) in workspaceTabs"
            :id="tab.buttonId"
            :key="tab.key"
            :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
            class="top-tab"
            :class="{ active: activeTab === tab.key }"
            type="button"
            role="tab"
            :tabindex="activeTab === tab.key ? 0 : -1"
            :aria-selected="activeTab === tab.key ? 'true' : 'false'"
            :aria-controls="tab.panelId"
            @click="selectTab(tab.key)"
            @keydown="handleTabKeydown($event, index)"
          >
            {{ tab.label }}
          </button>
        </nav>

        <section
          id="class-manage-overview"
          class="tab-panel"
          role="tabpanel"
          aria-labelledby="class-manage-tab-overview"
          :aria-hidden="activeTab === 'overview' ? 'false' : 'true'"
          v-show="activeTab === 'overview'"
        >
          <header class="teacher-topbar">
            <div class="teacher-heading">
              <div class="teacher-surface-eyebrow journal-eyebrow">Class Directory</div>
              <h1 class="teacher-title">班级管理</h1>
              <p class="teacher-copy">查看当前可管理班级，并进入对应班级继续查看学生和训练表现。</p>
            </div>

            <div class="teacher-actions">
              <button
                type="button"
                class="teacher-btn teacher-btn--primary"
                @click="emit('openDashboard')"
              >
                教学概览
              </button>
              <button
                type="button"
                class="teacher-btn teacher-btn--ghost"
                @click="emit('openReportExport')"
              >
                导出报告
              </button>
            </div>
          </header>

          <section class="teacher-summary">
            <div class="teacher-summary-title">
              <FolderKanban class="h-4 w-4" />
              <span>Directory Snapshot</span>
            </div>
            <div class="teacher-summary-grid">
              <div class="teacher-summary-item">
                <div class="teacher-summary-label">班级数量</div>
                <div class="teacher-summary-value">{{ total }}</div>
                <div class="teacher-summary-helper">当前可管理班级总数</div>
              </div>
              <div class="teacher-summary-item">
                <div class="teacher-summary-label">当前页学生数</div>
                <div class="teacher-summary-value">{{ currentPageStudentCount }}</div>
                <div class="teacher-summary-helper">当前分页已加载班级的学生数汇总</div>
              </div>
              <div class="teacher-summary-item">
                <div class="teacher-summary-label">当前状态</div>
                <div class="teacher-summary-value">{{ loading ? '同步中' : '已就绪' }}</div>
                <div class="teacher-summary-helper">班级目录与入口操作已同步</div>
              </div>
            </div>
          </section>
        </section>

        <section
          id="class-manage-directory"
          class="tab-panel"
          role="tabpanel"
          aria-labelledby="class-manage-tab-directory"
          :aria-hidden="activeTab === 'directory' ? 'false' : 'true'"
          v-show="activeTab === 'directory'"
        >
          <section class="teacher-controls">
            <div class="teacher-controls-bar">
              <div class="teacher-controls-heading">
                <div class="teacher-surface-eyebrow journal-eyebrow">Class Filters</div>
                <h3 class="teacher-controls-title">班级筛选</h3>
                <p class="teacher-controls-copy">支持按班级编号或班级名称快速定位班级入口。</p>
              </div>
            </div>

            <div class="teacher-filter-grid teacher-filter-grid--single">
              <label class="teacher-field">
                <span class="teacher-field-label">搜索班级</span>
                <div class="teacher-field-control teacher-filter-control">
                  <Search class="h-4 w-4 text-text-muted" />
                  <input
                    v-model="filterQuery"
                    type="text"
                    placeholder="搜索班级编号或名称"
                    class="teacher-input"
                  />
                </div>
              </label>
            </div>
          </section>

          <div class="teacher-hero-divider" />

          <div v-if="loading" class="teacher-skeleton-list workspace-directory-loading">
            <div
              v-for="index in 5"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
            />
          </div>

          <AppEmpty
            v-else-if="classes.length === 0"
            class="teacher-empty-state workspace-directory-empty"
            icon="Users"
            title="暂无班级"
            description="当前教师账号下还没有可访问的班级。"
          />

          <section
            v-else
            class="teacher-directory workspace-directory-section"
            aria-label="班级目录"
          >
            <div class="teacher-directory-top">
              <div>
                <h3 class="teacher-directory-title">班级目录</h3>
                <div class="teacher-directory-meta">
                  本页 {{ filteredClassEntries.length }} / {{ classes.length }} 个班级，共
                  {{ total }} 个班级
                </div>
              </div>
            </div>

            <div v-if="filteredClassEntries.length > 0" class="teacher-directory-head">
              <span class="teacher-directory-head-cell teacher-directory-head-cell-class-code"
                >班级编号</span
              >
              <span class="teacher-directory-head-cell teacher-directory-head-cell-class-name"
                >班级名称</span
              >
              <span class="teacher-directory-head-cell teacher-directory-head-cell-student-count"
                >学生数</span
              >
              <span>状态</span>
              <span>操作</span>
            </div>

            <AppEmpty
              v-if="filteredClassEntries.length === 0"
              class="teacher-empty-state workspace-directory-empty"
              icon="Search"
              title="没有匹配班级"
              description="调整搜索关键词后再试。"
            />

            <div v-if="filteredClassEntries.length > 0" class="workspace-directory-list">
              <button
                v-for="{ item, code } in filteredClassEntries"
                :key="item.name"
                type="button"
                class="teacher-directory-row"
                :aria-label="`${item.name}，${item.student_count || 0} 名学生，进入班级`"
                @click="emit('openClass', item.name)"
              >
                <div class="teacher-directory-cell teacher-directory-cell-class-code">
                  {{ code }}
                </div>

                <div class="teacher-directory-cell teacher-directory-cell-class-name">
                  <h4 class="teacher-directory-row-title" :title="item.name">{{ item.name }}</h4>
                </div>

                <div class="teacher-directory-cell teacher-directory-cell-student-count">
                  <div class="teacher-directory-row-points">{{ item.student_count || 0 }}</div>
                </div>

                <div class="teacher-directory-row-status">
                  <span
                    class="teacher-directory-state-chip"
                    :class="
                      (item.student_count || 0) > 0
                        ? 'teacher-directory-state-chip-ready'
                        : 'teacher-directory-state-chip-empty'
                    "
                  >
                    {{ (item.student_count || 0) > 0 ? '可查看' : '待入班' }}
                  </span>
                </div>

                <div class="teacher-directory-row-cta">
                  <span>进入班级</span>
                  <ArrowRight class="h-4 w-4" />
                </div>
              </button>
            </div>

            <div
              v-if="total > 0"
              class="teacher-directory-pagination workspace-directory-pagination"
            >
              <span>共 {{ total }} 个班级</span>
              <div class="teacher-directory-pagination-actions">
                <button
                  type="button"
                  class="teacher-btn teacher-btn--ghost teacher-directory-pagination-button"
                  :disabled="page === 1"
                  @click="emit('changePage', page - 1)"
                >
                  上一页
                </button>
                <span>{{ page }} / {{ totalPages }}</span>
                <button
                  type="button"
                  class="teacher-btn teacher-btn--ghost teacher-directory-pagination-button"
                  :disabled="page >= totalPages"
                  @click="emit('changePage', page + 1)"
                >
                  下一页
                </button>
              </div>
            </div>
          </section>
        </section>
      </div>
    </section>

    <div v-if="error" class="teacher-surface-error">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --journal-accent-strong: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --teacher-class-directory-columns: minmax(7rem, 0.7fr) minmax(11rem, 1.15fr) minmax(7rem, 0.7fr)
    minmax(7rem, 0.7fr) minmax(7rem, 0.75fr);
  --page-top-tabs-gap: 1.2rem;
  --page-top-tabs-margin: 0 -1.5rem 1.25rem;
  --page-top-tabs-padding: 0 1.5rem;
  --page-top-tabs-border: color-mix(in srgb, var(--journal-border) 88%, transparent);
  --page-top-tab-min-height: 3rem;
  --page-top-tab-padding: 0.4rem 0 0.75rem;
  --page-top-tab-font-size: 0.92rem;
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
  font-family:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
}

.teacher-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 7%, transparent),
      transparent 22rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      var(--journal-surface)
    );
  box-shadow: 0 22px 50px var(--color-shadow-soft);
}

.journal-eyebrow {
  letter-spacing: 0.08em;
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-topbar {
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-summary {
  display: grid;
  gap: 1.1rem;
  padding: 1.5rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-summary-title {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.teacher-summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.teacher-summary-item {
  min-width: 0;
  padding-left: 1rem;
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-summary-label {
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-summary-value {
  margin-top: 0.55rem;
  font-size: 1.35rem;
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--journal-ink);
}

.teacher-summary-helper {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.teacher-skeleton-list {
  margin-top: 1.5rem;
  display: grid;
  gap: 0.75rem;
}

.teacher-empty-state {
  margin-top: 1.5rem;
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-directory {
  display: flex;
  flex-direction: column;
  margin-top: 1.5rem;
}

.teacher-directory-top {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 0.5rem 1rem;
  padding-bottom: 0.9rem;
}

.teacher-directory-title {
  font-size: 1.1rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--journal-ink);
}

.teacher-directory-meta {
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.teacher-controls {
  display: grid;
  gap: 1rem;
  padding: 1.5rem 0 0;
}

.teacher-controls-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 0.85rem;
}

.teacher-controls-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-controls-copy {
  margin-top: 0.45rem;
  font-size: 0.84rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-hero-divider {
  margin-top: 1.5rem;
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-filter-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 22rem);
}

.teacher-filter-grid--single {
  justify-content: start;
}

.teacher-field {
  display: grid;
  gap: 0.45rem;
}

.teacher-field-label {
  font-size: 0.84rem;
  color: var(--journal-muted);
}

.teacher-field-control {
  width: 100%;
  min-height: 2.9rem;
  border: 1px solid var(--teacher-control-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.72rem 0.95rem;
  color: var(--journal-ink);
  transition:
    border-color 0.18s ease,
    background 0.18s ease;
}

.teacher-field-control:focus-within,
.teacher-field-control:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 5%, var(--journal-surface));
  outline: none;
}

.teacher-filter-control {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.teacher-input {
  width: 100%;
  background: transparent;
  color: var(--journal-ink);
  outline: none;
}

.teacher-input::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 76%, transparent);
}

.teacher-directory-head {
  display: grid;
  grid-template-columns: var(--teacher-class-directory-columns);
  gap: 1rem;
  padding: 0 0 0.75rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-directory-head-cell {
  min-width: 0;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--teacher-class-directory-columns);
  gap: 1rem;
  align-items: center;
  width: 100%;
  padding: 1.1rem 0;
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
  gap: 0.5rem;
  min-width: 0;
  align-content: center;
  justify-self: stretch;
  text-align: left;
}

.teacher-directory-cell-class-code,
.teacher-directory-row-points {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.teacher-directory-cell-class-code {
  font-size: 0.76rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--journal-muted);
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
  font-size: 1.08rem;
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-points {
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.teacher-directory-head-cell-class-code,
.teacher-directory-head-cell-class-name,
.teacher-directory-head-cell-student-count,
.teacher-directory-cell-class-code,
.teacher-directory-cell-class-name,
.teacher-directory-cell-student-count {
  justify-self: start;
  width: 100%;
}

.teacher-directory-row-copy {
  font-size: 0.84rem;
  line-height: 1.6;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-directory-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.65rem;
  padding: 0 0.62rem;
  border-radius: 0.5rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-directory-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.teacher-directory-row-status {
  display: flex;
  justify-content: flex-start;
}

.teacher-directory-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 0.62rem;
  border-radius: 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.teacher-directory-state-chip-ready {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

.teacher-directory-state-chip-empty {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.teacher-directory-row-metrics {
  display: grid;
  gap: 0.25rem;
  font-size: 0.81rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.teacher-directory-pagination-actions {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
}

.teacher-directory-pagination-button {
  min-height: 2.2rem;
  padding: 0 0.85rem;
}

.teacher-directory-pagination-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

@media (max-width: 960px) {
  .top-tabs {
    margin-left: -1rem;
    margin-right: -1rem;
    padding: 0 1rem;
  }

  .teacher-topbar {
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
    gap: 0.85rem;
    padding: 1rem 0;
  }

  .teacher-directory-pagination-actions {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
