<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FileUp, RefreshCw, UserPlus, UserRoundCheck } from 'lucide-vue-next'

import type { AdminUserImportData, AdminUserListItem, UserStatus } from '@/api/contracts'
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import type { UserRole } from '@/utils/constants'

type UserFilterRole = UserRole | 'all'
type UserFilterStatus = UserStatus | 'all'
type UserPanelKey = 'overview' | 'import'

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
  updateKeyword: [value: string]
  updateStudentNo: [value: string]
  updateTeacherNo: [value: string]
  updateRoleFilter: [value: UserFilterRole]
  updateStatusFilter: [value: UserFilterStatus]
  openCreateDialog: []
  openEditDialog: [user: AdminUserListItem]
  deleteUser: [userId: string]
  changePage: [page: number]
  importFile: [file: File]
}>()

const route = useRoute()
const router = useRouter()
const importInput = useTemplateRef<HTMLInputElement>('importInput')

const activePanel = computed<UserPanelKey>(() => {
  const rawPanel = route.query.panel
  const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
  if (panel === 'import') {
    return 'import'
  }
  return 'overview'
})

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const activeCount = computed(() => props.list.filter((item) => item.status === 'active').length)
const teacherCount = computed(
  () => props.list.filter((item) => item.roles.includes('teacher')).length
)
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
    widthClass: 'w-[12%] min-w-[8rem]',
    cellClass: 'user-table__username-cell',
  },
  {
    key: 'name',
    label: '姓名',
    widthClass: 'w-[13%] min-w-[8rem]',
    cellClass: 'user-table__name-cell',
  },
  {
    key: 'email',
    label: '邮箱',
    widthClass: 'w-[16%] min-w-[10rem]',
    cellClass: 'user-table__email-cell',
  },
  {
    key: 'roles',
    label: '角色',
    widthClass: 'w-[15%] min-w-[9rem]',
    cellClass: 'user-table__roles-cell',
  },
  {
    key: 'status',
    label: '状态',
    widthClass: 'w-[10%] min-w-[7rem]',
    cellClass: 'user-table__status-cell',
  },
  {
    key: 'class_name',
    label: '班级',
    widthClass: 'w-[10%] min-w-[7rem]',
    cellClass: 'user-table__class-cell',
  },
  {
    key: 'identity',
    label: '学号 / 工号',
    widthClass: 'w-[12%] min-w-[8rem]',
    cellClass: 'user-table__identity-cell',
  },
  {
    key: 'created_at',
    label: '创建时间',
    widthClass: 'w-[12%] min-w-[9rem]',
    cellClass: 'user-table__time-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[10rem]',
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

function getUserIdentity(user: AdminUserListItem): string {
  if (user.roles.includes('admin') || user.roles.includes('teacher')) {
    return user.teacher_no || '未设置'
  }
  if (user.roles.includes('student')) {
    return user.student_no || '未设置'
  }
  return '未设置'
}

function formatCreatedAt(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}

function resetDirectoryFilters(): void {
  emit('updateKeyword', '')
  emit('updateStudentNo', '')
  emit('updateTeacherNo', '')
  emit('updateRoleFilter', 'all')
  emit('updateStatusFilter', 'all')
}

async function switchPanel(panel: UserPanelKey): Promise<void> {
  if (activePanel.value === panel) return

  const nextQuery = { ...route.query }
  if (panel === 'overview') {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panel
  }

  await router.replace({ name: 'UserManage', query: nextQuery })
}

function triggerImport(): void {
  importInput.value?.click()
}

function handleImportChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  emit('importFile', file)
  input.value = ''
}
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <section
        id="user-panel-overview"
        v-show="activePanel === 'overview'"
        class="user-panel user-panel--workspace"
        :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
      >
        <header class="workspace-tab-heading user-overview-head">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">User Workspace</div>
            <h1 class="workspace-page-title">用户治理台</h1>
            <p class="workspace-page-copy">
              上面直接查看用户规模和导入回执，下面围绕具体账号完成搜索、筛选、编辑与治理操作。
            </p>
          </div>

          <div class="user-panel-actions">
            <button type="button" class="ui-btn ui-btn--ghost" @click="emit('refresh')">
              <RefreshCw class="h-4 w-4" />
              刷新列表
            </button>
            <button
              id="user-open-import"
              type="button"
              class="ui-btn ui-btn--ghost"
              @click="switchPanel('import')"
            >
              <FileUp class="h-4 w-4" />
              导入用户
            </button>
            <button
              id="user-open-create"
              type="button"
              class="ui-btn ui-btn--primary"
              @click="emit('openCreateDialog')"
            >
              <UserPlus class="h-4 w-4" />
              创建用户
            </button>
          </div>
        </header>

        <div
          class="admin-summary-grid user-overview-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
        >
          <div class="journal-note user-overview-stat progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">用户总量</div>
            <div class="journal-note-value progress-card-value metric-panel-value">{{ total }}</div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前筛选条件下的用户总数
            </div>
          </div>
          <div class="journal-note user-overview-stat progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">活跃账号</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ activeCount }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前页处于 active 的账号
            </div>
          </div>
          <div class="journal-note user-overview-stat progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">教师角色</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ teacherCount }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前页教师账号数量
            </div>
          </div>
          <div class="journal-note user-overview-stat progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">导入回执</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ importSummary }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              最近一次导入结果
            </div>
          </div>
        </div>

        <section class="workspace-directory-section user-directory-section">
          <header class="list-heading user-directory-head">
            <div>
              <div class="journal-note-label">User Directory</div>
              <h2 class="list-heading__title">全部用户</h2>
            </div>
            <div class="user-directory-meta">当前页 {{ listCount }} 个用户</div>
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
                    class="admin-input user-filter-control"
                    @change="
                      emit(
                        'updateRoleFilter',
                        ($event.target as HTMLSelectElement).value as UserFilterRole
                      )
                    "
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
                    class="admin-input user-filter-control"
                    @change="
                      emit(
                        'updateStatusFilter',
                        ($event.target as HTMLSelectElement).value as UserFilterStatus
                      )
                    "
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
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="emit('openCreateDialog')"
              >
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
                <span class="user-row__username-handle"
                  >@{{ (row as AdminUserListItem).username }}</span
                >
              </div>
            </template>

            <template #cell-name="{ row }">
              <span class="user-row__name">
                {{ (row as AdminUserListItem).name || (row as AdminUserListItem).username }}
              </span>
            </template>

            <template #cell-email="{ row }">
              <span class="user-row__email">
                {{ (row as AdminUserListItem).email || '未填写邮箱' }}
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
              <span
                class="admin-status-chip"
                :style="getUserStatusStyle((row as AdminUserListItem).status)"
              >
                {{ (row as AdminUserListItem).status }}
              </span>
            </template>

            <template #cell-class_name="{ row }">
              <span class="user-row__class">
                {{ (row as AdminUserListItem).class_name || '未分配班级' }}
              </span>
            </template>

            <template #cell-identity="{ row }">
              <span class="user-row__identity">
                {{ getUserIdentity(row as AdminUserListItem) }}
              </span>
            </template>

            <template #cell-created_at="{ row }">
              <span class="user-row__time">
                {{ formatCreatedAt((row as AdminUserListItem).created_at) }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="user-row__actions">
                <button
                  type="button"
                  class="ui-btn ui-btn--secondary user-action-btn"
                  @click="emit('openEditDialog', row as AdminUserListItem)"
                >
                  编辑
                </button>
                <button
                  type="button"
                  class="ui-btn ui-btn--danger user-action-btn"
                  @click="emit('deleteUser', (row as AdminUserListItem).id)"
                >
                  删除
                </button>
              </div>
            </template>
          </WorkspaceDataTable>

          <div v-if="list.length > 0" class="admin-pagination workspace-directory-pagination">
            <PlatformPaginationControls
              :page="page"
              :total-pages="totalPages"
              :total="total"
              :total-label="`共 ${total} 个用户`"
              @change-page="emit('changePage', $event)"
            />
          </div>
        </section>
      </section>

      <section
        id="user-panel-import"
        v-show="activePanel === 'import'"
        class="user-panel user-panel--import"
        :aria-hidden="activePanel === 'import' ? 'false' : 'true'"
      >
        <section class="workspace-directory-section user-import-panel">
          <header class="workspace-tab-heading user-import-head">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">User Import</div>
              <h2 class="workspace-page-title">导入用户</h2>
              <p class="workspace-page-copy">
                统一导入账号、角色与班级归属，导入完成后可回到工作台继续筛选和治理具体用户。
              </p>
            </div>

            <div class="user-panel-actions">
              <button
                id="user-return-overview"
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="switchPanel('overview')"
              >
                返回工作台
              </button>
              <button type="button" class="ui-btn ui-btn--primary" @click="triggerImport">
                <FileUp class="h-4 w-4" />
                批量导入
              </button>
            </div>
          </header>

          <div class="journal-note user-import-format">
            <div class="journal-note-label">CSV 格式</div>
            <div class="journal-note-helper">
              列顺序：`username,password,email,class_name,role,status,student_no,teacher_no,name`
            </div>
          </div>

          <section class="workspace-directory-section user-import-receipt-section">
            <header class="list-heading user-import-receipt-head">
              <div>
                <div class="journal-note-label">Import Receipt</div>
                <h2 class="list-heading__title">导入回执</h2>
              </div>
            </header>

            <div v-if="importResult" class="admin-receipt">
              <p>
                创建 {{ importResult.created }}，更新 {{ importResult.updated }}，失败
                {{ importResult.failed }}
              </p>
              <ul
                v-if="importResult.errors?.length"
                class="mt-3 space-y-2 text-[var(--color-danger)]"
              >
                <li
                  v-for="item in importResult.errors.slice(0, 5)"
                  :key="`${item.row}-${item.message}`"
                >
                  第 {{ item.row }} 行：{{ item.message }}
                </li>
              </ul>
            </div>
            <div v-else class="admin-empty">还没有导入记录。</div>
          </section>
        </section>
      </section>
    </main>

    <input
      ref="importInput"
      type="file"
      accept=".csv,text/csv"
      class="hidden"
      @change="handleImportChange"
    />
  </section>
</template>

<style scoped>
.journal-shell {
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --user-table-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --user-row-divider: color-mix(in srgb, var(--journal-border) 58%, transparent);
  --journal-note-label-weight: 600;
  --journal-note-label-spacing: 0.15em;
  --journal-note-label-color: var(--journal-muted);
  --journal-shell-dark-accent: var(--color-primary-hover);
}

.user-panel {
  display: grid;
  gap: var(--space-4);
}

.user-overview-head,
.user-import-head {
  gap: var(--space-3);
}

.user-panel-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
}

.user-overview-grid {
  --admin-summary-grid-gap: var(--space-3-5);
  --admin-summary-grid-columns: repeat(4, minmax(0, 1fr));
}

.user-overview-stat {
  display: flex;
  min-height: 140px;
  flex-direction: column;
  justify-content: space-between;
}

.user-directory-head {
  gap: var(--space-4);
}

.user-directory-meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.user-panel-actions > .ui-btn,
.user-row__actions > .ui-btn,
.workspace-directory-empty .ui-btn {
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

.user-panel-actions > .ui-btn.ui-btn--primary,
.workspace-directory-empty .ui-btn.ui-btn--primary {
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: var(--color-primary-hover);
}

.user-panel-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--admin-control-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

.user-row__actions > .ui-btn.ui-btn--secondary {
  --ui-btn-border: var(--admin-control-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

.user-row__actions > .ui-btn.ui-btn--danger {
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 20%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
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

.user-import-format {
  border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 1.1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: var(--space-4);
}

.admin-receipt {
  border-radius: 16px;
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 95%, transparent);
  padding: var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
}

.admin-empty {
  border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 16px;
  padding: var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
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

.user-row__username-handle,
.user-row__email,
.user-row__class,
.user-row__identity,
.user-row__time {
  color: var(--journal-muted);
}

.user-row__username-handle,
.user-row__identity,
.user-row__time {
  font-family: var(--font-family-mono);
}

.user-row__name {
  color: var(--journal-ink);
  font-weight: 600;
}

.user-row__email,
.user-row__class,
.user-row__identity,
.user-row__time {
  display: block;
  line-height: 1.6;
}

.user-row__email,
.user-row__class {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
.admin-inline-chip,
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

.admin-inline-chip {
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--journal-surface));
  color: var(--journal-muted);
}

.admin-role-chip {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 16%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

@media (max-width: 767px) {
  .journal-hero {
    padding-left: 1rem;
    padding-right: 1rem;
  }

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
