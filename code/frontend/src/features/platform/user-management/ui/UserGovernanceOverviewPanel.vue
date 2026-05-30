<script setup lang="ts">
import { computed } from 'vue'
import { FileUp, GraduationCap, RefreshCw, UserPlus, UserRoundCheck, Users } from 'lucide-vue-next'

import type { AdminUserImportData, AdminUserListItem, UserStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import type { UserRole } from '@/utils/constants'

type UserFilterRole = UserRole | 'all'
type UserFilterStatus = UserStatus | 'all'

const props = defineProps<{
  list: AdminUserListItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  keyword: string
  studentNo: string
  teacherNo: string
  roleFilter: UserFilterRole
  statusFilter: UserFilterStatus
  importResult: AdminUserImportData | null
}>()

const emit = defineEmits<{
  refresh: []
  openImport: []
  openCreateDialog: []
  updateKeyword: [value: string]
  updateStudentNo: [value: string]
  updateTeacherNo: [value: string]
  updateRoleFilter: [value: UserFilterRole]
  updateStatusFilter: [value: UserFilterStatus]
  openEditDialog: [user: AdminUserListItem]
  openUserDetail: [user: AdminUserListItem]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const activeCount = computed(() => props.list.filter((item) => item.status === 'active').length)
const teacherCount = computed(() => props.list.filter((item) => item.roles.includes('teacher')).length)
const listCount = computed(() => props.list.length)
const hasActiveFilters = computed(() =>
  Boolean(
    props.keyword.trim() ||
    props.studentNo.trim() ||
    props.teacherNo.trim() ||
    props.roleFilter !== 'all' ||
    props.statusFilter !== 'all'
  )
)
const importSummary = computed(() => {
  if (!props.importResult) return '暂无导入记录'
  return `创建 ${props.importResult.created} / 更新 ${props.importResult.updated}`
})

const userTableColumns = [
  {
    key: 'username',
    label: '用户',
    widthClass: 'w-[20%] min-w-[10rem]',
    cellClass: 'user-table__username-cell',
  },
  {
    key: 'name',
    label: '姓名',
    widthClass: 'w-[20%] min-w-[10rem]',
    cellClass: 'user-table__name-cell',
  },
  {
    key: 'roles',
    label: '角色',
    widthClass: 'w-[24%] min-w-[13rem]',
    cellClass: 'user-table__roles-cell',
  },
  {
    key: 'status',
    label: '状态',
    widthClass: 'w-[14%] min-w-[8rem]',
    cellClass: 'user-table__status-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[12rem]',
    cellClass: 'user-table__actions-cell',
  },
]

const userStatusAccentMap: Record<UserStatus, string> = {
  active: 'var(--color-primary)',
  locked: 'var(--color-warning)',
  banned: 'var(--color-danger)',
  inactive: 'color-mix(in srgb, var(--journal-muted) 84%, var(--journal-ink))',
}

function getUserAccentColor(status: UserStatus): string {
  return userStatusAccentMap[status] ?? 'var(--color-primary)'
}

function getUserStatusStyle(status: UserStatus): Record<string, string> {
  const accent = getUserAccentColor(status)
  return {
    color: accent,
    borderColor: `color-mix(in srgb, ${accent} 18%, transparent)`,
    backgroundColor: `color-mix(in srgb, ${accent} 10%, var(--journal-surface))`,
  }
}

function resetDirectoryFilters(): void {
  emit('updateKeyword', '')
  emit('updateStudentNo', '')
  emit('updateTeacherNo', '')
  emit('updateRoleFilter', 'all')
  emit('updateStatusFilter', 'all')
}
</script>

<template>
  <section class="user-panel user-panel--workspace">
    <header class="workspace-panel-header user-overview-head">
      <div class="workspace-panel-header__intro">
        <div class="workspace-overline">
          User Workspace
        </div>
        <h1 class="workspace-page-title">用户治理台</h1>
      </div>

      <div class="workspace-panel-header__actions header-actions user-panel-actions">
        <button type="button" class="header-btn header-btn--ghost" @click="emit('refresh')">
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <button
          id="user-open-import"
          type="button"
          class="header-btn header-btn--ghost"
          @click="emit('openImport')"
        >
          <FileUp class="h-4 w-4" />
          导入用户
        </button>
        <button
          id="user-open-create"
          type="button"
          class="header-btn header-btn--primary"
          @click="emit('openCreateDialog')"
        >
          <UserPlus class="h-4 w-4" />
          创建用户
        </button>
      </div>
      <div class="workspace-panel-header__summary admin-summary-grid user-overview-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>用户总量</span>
            <Users class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ total.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前条件下的用户总数
          </div>
        </article>

        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>活跃账号</span>
            <UserPlus class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ activeCount.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前活跃状态的账号
          </div>
        </article>

        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>教师角色</span>
            <GraduationCap class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ teacherCount.toString().padStart(2, '0') }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前页教师账号数量
          </div>
        </article>

        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">
            <span>导入回执</span>
            <FileUp class="h-4 w-4" />
          </div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ props.importResult ? (props.importResult.created + props.importResult.updated).toString().padStart(2, '0') : '00' }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            {{ importSummary }}
          </div>
        </article>
      </div>
    </header>

    <div class="workspace-panel-divider" aria-hidden="true" />

    <section class="workspace-directory-section user-directory-section">
      <header class="list-heading user-directory-head">
        <div>
          <div class="journal-note-label">
            User Directory
          </div>
          <h2 class="list-heading__title">
            全部用户
          </h2>
        </div>
        <div class="user-directory-meta">
          当前页 {{ listCount }} 个用户
        </div>
      </header>

      <WorkspaceDirectoryToolbar
        :model-value="keyword"
        :total="total"
        selected-sort-label=""
        :sort-options="[]"
        search-placeholder="用户名 / 邮箱 / 班级 / 学号 / 工号"
        filter-panel-title="用户筛选"
        total-suffix="个用户"
        reset-label="重置筛选"
        :reset-disabled="!hasActiveFilters"
        @update:model-value="emit('updateKeyword', $event)"
        @reset-filters="resetDirectoryFilters"
      >
        <template #filter-panel>
          <div class="user-filter-grid">
            <label class="user-filter-field">
              <span class="user-filter-label">角色</span>
              <select
                :value="roleFilter"
                class="admin-input workspace-directory-filter-control user-filter-control"
                @change="emit('updateRoleFilter', ($event.target as HTMLSelectElement).value as UserFilterRole)"
              >
                <option value="all">全部角色</option>
                <option value="student">student</option>
                <option value="teacher">teacher</option>
                <option value="admin">admin</option>
              </select>
            </label>

            <label class="user-filter-field">
              <span class="user-filter-label">状态</span>
              <select
                :value="statusFilter"
                class="admin-input workspace-directory-filter-control user-filter-control"
                @change="emit('updateStatusFilter', ($event.target as HTMLSelectElement).value as UserFilterStatus)"
              >
                <option value="all">全部状态</option>
                <option value="active">active</option>
                <option value="inactive">inactive</option>
                <option value="locked">locked</option>
                <option value="banned">banned</option>
              </select>
            </label>
          </div>
        </template>
      </WorkspaceDirectoryToolbar>

      <div
        v-if="loading && list.length === 0"
        class="workspace-directory-loading flex justify-center py-10"
      >
        <AppLoading>正在同步用户列表...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="list.length === 0"
        class="workspace-directory-empty"
        title="暂无用户"
        description="当前筛选条件下没有匹配用户。"
        icon="UsersRound"
      >
        <template #action>
          <button type="button" class="ui-btn ui-btn--primary" @click="emit('openCreateDialog')">
            创建第一个用户
          </button>
        </template>
      </AppEmpty>

      <WorkspaceDataTable
        v-else
        class="user-table-shell workspace-directory-list user-list"
        :columns="userTableColumns"
        :rows="list"
        row-key="id"
        row-class="user-table-row"
      >
        <template #cell-username="{ row }">
          <div class="user-row__username">
            <span class="user-row__username-handle">@{{ (row as AdminUserListItem).username }}</span>
          </div>
        </template>

        <template #cell-name="{ row }">
          <span class="user-row__name">
            {{ (row as AdminUserListItem).name || (row as AdminUserListItem).username }}
          </span>
        </template>

        <template #cell-roles="{ row }">
          <div class="user-row__roles">
            <span
              v-for="role in (row as AdminUserListItem).roles"
              :key="`${(row as AdminUserListItem).id}-${role}`"
              class="admin-role-chip"
            >
              <UserRoundCheck class="h-3.5 w-3.5" />
              {{ role }}
            </span>
          </div>
        </template>

        <template #cell-status="{ row }">
          <span class="admin-status-chip" :style="getUserStatusStyle((row as AdminUserListItem).status)">
            {{ (row as AdminUserListItem).status }}
          </span>
        </template>

        <template #cell-actions="{ row }">
          <div class="user-row__actions">
            <button
              :id="`user-row-detail-${(row as AdminUserListItem).id}`"
              type="button"
              class="ui-btn ui-btn--secondary user-action-btn"
              @click="emit('openUserDetail', row as AdminUserListItem)"
            >
              详情
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--secondary user-action-btn"
              @click="emit('openEditDialog', row as AdminUserListItem)"
            >
              编辑
            </button>
          </div>
        </template>
      </WorkspaceDataTable>

      <div v-if="list.length > 0" class="admin-pagination workspace-directory-pagination">
        <PagePaginationControls
          :page="page"
          :total-pages="totalPages"
          :total="total"
          :total-label="`共 ${total} 个用户`"
          :show-jump="true"
          @change-page="emit('changePage', $event)"
        />
      </div>
    </section>
  </section>
</template>

<style scoped>
.user-panel {
  display: grid;
  gap: var(--space-4);
}

.user-panel--workspace {
  gap: 0;
}

.user-overview-head {
  gap: var(--space-3);
  --workspace-panel-divider-gap: var(--space-divider-gap);
}

.user-overview-grid {
  --admin-summary-grid-gap: var(--space-3-5);
  --admin-summary-grid-columns: repeat(4, minmax(0, 1fr));
}

.user-directory-head {
  gap: var(--space-4);
}

.user-directory-meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.user-row__actions > .ui-btn {
  --ui-btn-height: 2.75rem;
  --ui-btn-radius: 1rem;
  --ui-btn-padding: var(--space-2-5) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.user-action-btn {
  --ui-btn-height: 2rem;
  --ui-btn-padding: var(--space-1-5) var(--space-3);
  --ui-btn-radius: 0.8rem;
  --ui-btn-font-size: var(--font-size-0-8125);
}

.user-row__actions > .ui-btn.ui-btn--secondary {
  --ui-btn-border: var(--admin-control-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

.admin-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--admin-control-border);
  background: var(--journal-surface);
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
  outline: none;
  transition: border-color 150ms ease;
}

.admin-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
}

.user-filter-grid {
  display: grid;
  gap: var(--space-4);
}

.user-filter-field {
  display: grid;
  gap: var(--space-2);
}

.user-filter-label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.user-filter-control {
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
}

.user-table-shell {
  --workspace-directory-shell-border: var(--user-table-border);
  --workspace-directory-head-divider: var(--user-table-border);
  --workspace-directory-row-divider: var(--user-row-divider);
}

.user-table-row {
  border-top: 1px solid var(--user-row-divider);
  transition: background 180ms ease;
}

.user-table-shell :deep(.workspace-data-table__body-cell) {
  vertical-align: top;
}

.user-table-row:hover,
.user-table-row:focus-within {
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--journal-surface));
}

.user-row__username,
.user-row__roles,
.user-row__actions {
  display: flex;
}

.user-row__username {
  min-width: 0;
}

.user-row__username-handle {
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.user-row__name {
  color: var(--journal-ink);
  font-weight: 600;
}

.user-row__roles {
  flex-wrap: wrap;
  gap: var(--space-2);
}

.user-row__actions {
  justify-content: flex-end;
  gap: var(--space-2);
}

.admin-status-chip,
.admin-role-chip {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  border-radius: 999px;
  padding: var(--space-1-5) var(--space-3);
  font-size: var(--font-size-0-72);
  font-weight: 600;
}

.admin-status-chip {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 14%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.admin-role-chip {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 16%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

@media (max-width: 767px) {
  .user-overview-grid {
    --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
  }

  .user-panel-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 560px) {
  .user-overview-grid {
    --admin-summary-grid-columns: 1fr;
  }
}
</style>
