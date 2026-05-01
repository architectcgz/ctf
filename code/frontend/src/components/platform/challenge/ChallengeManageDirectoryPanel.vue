<script setup lang="ts">
import type {
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
} from '@/api/contracts'
import type { PlatformChallengeListRow } from '@/features/platform-challenges'
import {
  ChallengeCategoryPill,
  ChallengeDifficultyText,
} from '@/entities/challenge'
import AppEmpty from '@/components/common/AppEmpty.vue'
import CActionMenu from '@/components/common/menus/CActionMenu.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import {
  Book,
  CheckCircle,
  Eye,
  FileSearch,
  MoreHorizontal,
  Trash2,
} from 'lucide-vue-next'

type ChallengeManageStatusFilter = Extract<ChallengeStatus, 'draft' | 'published' | 'archived'>

interface Props {
  rows: PlatformChallengeListRow[]
  total: number
  page: number
  totalPages: number
  loading: boolean
  hasLoadError: boolean
  loadErrorMessage: string
  hasActiveFilters: boolean
  manageEmptyTitle: string
  manageEmptyMessage: string
  keyword: string
  categoryFilter: ChallengeCategory | ''
  difficultyFilter: ChallengeDifficulty | ''
  statusFilter: ChallengeManageStatusFilter | ''
  selectedSortLabel: string
  sortOptions: WorkspaceDirectorySortOption[]
  openActionMenuId: string | null
}

defineProps<Props>()

const emit = defineEmits<{
  'update:keyword': [value: string]
  'update:categoryFilter': [value: ChallengeCategory | '']
  'update:difficultyFilter': [value: ChallengeDifficulty | '']
  'update:statusFilter': [value: ChallengeManageStatusFilter | '']
  'select-sort': [option: WorkspaceDirectorySortOption]
  'reset-filters': []
  retry: []
  'change-page': [page: number]
  'update-action-menu-open': [payload: { challengeId: string; open: boolean }]
  'open-detail': [challengeId: string]
  'open-topology': [challengeId: string]
  'open-writeup': [challengeId: string]
  'submit-publish-check': [row: PlatformChallengeListRow]
  'remove-challenge': [challengeId: string]
}>()

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

function updateCategoryFilter(event: Event): void {
  emit('update:categoryFilter', (event.target as HTMLSelectElement).value as ChallengeCategory | '')
}

function updateDifficultyFilter(event: Event): void {
  emit(
    'update:difficultyFilter',
    (event.target as HTMLSelectElement).value as ChallengeDifficulty | ''
  )
}

function updateStatusFilter(event: Event): void {
  emit(
    'update:statusFilter',
    (event.target as HTMLSelectElement).value as ChallengeManageStatusFilter | ''
  )
}
</script>

<template>
  <section class="workspace-directory-section challenge-manage-directory">
    <header class="list-heading">
      <div>
        <div class="workspace-overline">Challenge Directory</div>
        <h2 class="list-heading__title">题目目录</h2>
      </div>
    </header>

    <WorkspaceDirectoryToolbar
      :model-value="keyword"
      :total="total"
      :selected-sort-label="selectedSortLabel"
      :sort-options="sortOptions"
      search-placeholder="检索题目名称..."
      :reset-disabled="!hasActiveFilters"
      @update:model-value="emit('update:keyword', $event)"
      @select-sort="emit('select-sort', $event)"
      @reset-filters="emit('reset-filters')"
    >
      <template #filter-panel>
        <div class="challenge-filter-grid">
          <label class="challenge-filter-field">
            <span class="challenge-filter-label">题目分类</span>
            <select
              :value="categoryFilter"
              class="challenge-filter-select"
              @change="updateCategoryFilter"
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
              :value="difficultyFilter"
              class="challenge-filter-select"
              @change="updateDifficultyFilter"
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
              :value="statusFilter"
              class="challenge-filter-select"
              @change="updateStatusFilter"
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
          @click="emit('retry')"
        >
          重新加载
        </button>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="rows.length === 0"
      class="workspace-directory-empty"
      icon="BookOpen"
      :title="manageEmptyTitle"
      :description="manageEmptyMessage"
    />

    <WorkspaceDataTable
      v-else
      class="challenge-list workspace-directory-list"
      :columns="challengeTableColumns"
      :rows="rows"
      row-key="id"
      row-class="challenge-table-row group"
    >
      <template #cell-title="{ row }">
        <div
          class="challenge-table-title"
          :title="(row as PlatformChallengeListRow).title"
        >
          {{ (row as PlatformChallengeListRow).title }}
        </div>
      </template>

      <template #cell-category="{ row }">
        <ChallengeCategoryPill :category="(row as PlatformChallengeListRow).category" />
      </template>

      <template #cell-difficulty="{ row }">
        <ChallengeDifficultyText :difficulty="(row as PlatformChallengeListRow).difficulty" />
      </template>

      <template #cell-points="{ row }">
        <span class="challenge-table-points">{{ (row as PlatformChallengeListRow).points }}</span>
      </template>

      <template #cell-status="{ row }">
        <div class="challenge-table-status">
          <div
            class="challenge-table-status__dot"
            :class="
              (row as PlatformChallengeListRow).status === 'published'
                ? 'challenge-table-status__dot--published'
                : 'challenge-table-status__dot--idle'
            "
          />
          <span class="challenge-table-status__label">
            {{
              (row as PlatformChallengeListRow).status === 'published'
                ? '已发布'
                : (row as PlatformChallengeListRow).status === 'archived'
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
            @click="emit('open-detail', (row as PlatformChallengeListRow).id)"
          >
            <Eye class="h-3 w-3" />
            查看
          </button>

          <CActionMenu
            :open="openActionMenuId === (row as PlatformChallengeListRow).id"
            title="Management"
            menu-label="题目更多操作"
            @update:open="
              emit('update-action-menu-open', {
                challengeId: (row as PlatformChallengeListRow).id,
                open: $event,
              })
            "
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
                @click="emit('open-topology', (row as PlatformChallengeListRow).id)"
              >
                <FileSearch class="h-3 w-3" />
                编排拓扑
              </button>
              <button
                type="button"
                class="c-action-menu__item"
                @click="emit('open-writeup', (row as PlatformChallengeListRow).id)"
              >
                <Book class="h-3 w-3" />
                题解与提示
              </button>
              <button
                v-if="(row as PlatformChallengeListRow).status !== 'published'"
                type="button"
                class="c-action-menu__item c-action-menu__item--success"
                @click="emit('submit-publish-check', row as PlatformChallengeListRow)"
              >
                <CheckCircle class="h-3 w-3" />
                提交发布检查
              </button>
              <button
                type="button"
                class="c-action-menu__item c-action-menu__item--danger"
                @click="emit('remove-challenge', (row as PlatformChallengeListRow).id)"
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
      :total-pages="totalPages"
      :total="total"
      :total-label="`共 ${total} 道题目`"
      @change-page="emit('change-page', $event)"
    />
  </section>
</template>

<style scoped>
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
