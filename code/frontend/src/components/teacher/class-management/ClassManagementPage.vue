<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowRight, FolderKanban, Search } from 'lucide-vue-next'

import type { TeacherClassItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import { useUrlSyncedTabs } from '@/composables/useUrlSyncedTabs'

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

const workspaceTabOrder = workspaceTabs.map((tab) => tab.key) as WorkspaceTab[]
const { activeTab, setTabButtonRef, selectTab, handleTabKeydown } = useUrlSyncedTabs<WorkspaceTab>(
  {
    orderedTabs: workspaceTabOrder,
    defaultTab: 'overview',
  }
)

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
                导出班级报告
              </button>
            </div>
          </header>

          <section class="teacher-summary metric-panel-default-surface">
            <div class="teacher-summary-title">
              <FolderKanban class="h-4 w-4" />
              <span>Directory Snapshot</span>
            </div>
            <div class="teacher-summary-grid metric-panel-grid">
              <div class="teacher-summary-item metric-panel-card">
                <div class="teacher-summary-label metric-panel-label">班级数量</div>
                <div class="teacher-summary-value metric-panel-value">{{ total }}</div>
                <div class="teacher-summary-helper metric-panel-helper">当前可管理班级总数</div>
              </div>
              <div class="teacher-summary-item metric-panel-card">
                <div class="teacher-summary-label metric-panel-label">当前页学生数</div>
                <div class="teacher-summary-value metric-panel-value">{{ currentPageStudentCount }}</div>
                <div class="teacher-summary-helper metric-panel-helper">当前分页已加载班级的学生数汇总</div>
              </div>
              <div class="teacher-summary-item metric-panel-card">
                <div class="teacher-summary-label metric-panel-label">当前状态</div>
                <div class="teacher-summary-value metric-panel-value">
                  {{ loading ? '同步中' : '已就绪' }}
                </div>
                <div class="teacher-summary-helper metric-panel-helper">班级目录与入口操作已同步</div>
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
              <div class="teacher-controls-heading workspace-tab-heading__main">
                <div class="teacher-surface-eyebrow journal-eyebrow">Class Filters</div>
                <h3 class="teacher-controls-title workspace-tab-heading__title">班级筛选</h3>
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
              <PagePaginationControls
                :page="page"
                :total-pages="totalPages"
                :total="total"
                :total-label="`共 ${total} 个班级`"
                @change-page="emit('changePage', $event)"
              />
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
  --teacher-management-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --teacher-management-accent-strong: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --teacher-directory-columns: var(--teacher-class-directory-columns);
  --teacher-class-directory-columns: minmax(7rem, 0.7fr) minmax(11rem, 1.15fr) minmax(7rem, 0.7fr)
    minmax(7rem, 0.7fr) minmax(7rem, 0.75fr);
  --page-top-tabs-gap: var(--space-5);
  --page-top-tabs-margin: 0 calc(var(--space-6) * -1) var(--space-5);
  --page-top-tabs-padding: 0 var(--space-6);
  --page-top-tabs-border: color-mix(in srgb, var(--journal-border) 88%, transparent);
  --page-top-tab-min-height: 3rem;
  --page-top-tab-padding: var(--space-1-5) 0 var(--space-3);
  --page-top-tab-font-size: var(--font-size-0-92);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
  font-family: var(--font-family-sans);
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-skeleton-list {
  margin-top: var(--space-6);
  display: grid;
  gap: var(--space-3);
}

.teacher-empty-state {
  margin-top: var(--space-6);
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-hero-divider {
  margin-top: var(--space-6);
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 22rem);
}

.teacher-filter-grid--single {
  justify-content: start;
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--teacher-class-directory-columns);
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

.teacher-directory-cell-class-code,
.teacher-directory-row-points {
  font-family: var(--font-family-mono);
}

.teacher-directory-cell-class-code {
  font-size: var(--font-size-0-76);
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--journal-muted);
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-1-08);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-points {
  font-size: var(--font-size-1-00);
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
  font-size: var(--font-size-0-84);
  line-height: 1.6;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-directory-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.teacher-directory-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.65rem;
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

.teacher-directory-row-status {
  display: flex;
  justify-content: flex-start;
}

.teacher-directory-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 var(--space-2-5);
  border-radius: 0.5rem;
  font-size: var(--font-size-0-75);
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
  gap: var(--space-1);
  font-size: var(--font-size-0-81);
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  font-size: var(--font-size-0-82);
  font-weight: 700;
  color: var(--journal-accent-strong);
}

@media (max-width: 960px) {
  .top-tabs {
    margin-left: calc(var(--space-4) * -1);
    margin-right: calc(var(--space-4) * -1);
    padding: 0 var(--space-4);
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
    gap: var(--space-3);
    padding: var(--space-4) 0;
  }
}
</style>
