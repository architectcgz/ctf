<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FileUp, RefreshCw, UserPlus, UsersRound, UserRoundCheck } from 'lucide-vue-next'

import type { AdminUserImportData, AdminUserListItem, UserStatus } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
import type { UserRole } from '@/utils/constants'

type UserFilterRole = UserRole | 'all'
type UserFilterStatus = UserStatus | 'all'
type UserPanelKey = 'overview' | 'directory' | 'import'

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

const panelTabs: Array<{ key: UserPanelKey; label: string; panelId: string; tabId: string }> = [
  {
    key: 'overview',
    label: '总览',
    panelId: 'user-overview-summary',
    tabId: 'user-tab-overview',
  },
  {
    key: 'directory',
    label: '用户列表',
    panelId: 'user-directory-filters',
    tabId: 'user-tab-directory',
  },
  { key: 'import', label: '导入用户', panelId: 'user-import-start', tabId: 'user-tab-import' },
]
const panelTabOrder = panelTabs.map((tab) => tab.key) as UserPanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<UserPanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'overview',
  routeName: 'UserManage',
})

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const activeCount = computed(() => props.list.filter((item) => item.status === 'active').length)
const teacherCount = computed(
  () => props.list.filter((item) => item.roles.includes('teacher')).length
)
const importSummary = computed(() => {
  if (!props.importResult) return '暂无导入记录'
  return `创建 ${props.importResult.created} / 更新 ${props.importResult.updated}`
})
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
    class="journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="workspace-overline">User Governance</div>

    <nav class="top-tabs" role="tablist" aria-label="用户治理标签页">
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
        @click="switchPanel(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <section
      id="user-overview-summary"
      class="tab-panel"
      role="tabpanel"
      aria-labelledby="user-tab-overview"
      :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
    >
      <template v-if="activePanel === 'overview'">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">User Governance</div>
          <h1 class="workspace-page-title">用户治理台</h1>
        </div>
        <p class="workspace-page-copy">在这里筛选账号、批量导入并处理用户状态。</p>

        <div class="user-overview-summary">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <UsersRound class="h-5 w-5 text-[var(--journal-accent)]" />
            当前用户概况
          </div>
          <div
            class="admin-summary-grid user-overview-grid progress-strip metric-panel-grid metric-panel-default-surface mt-5"
          >
            <div class="journal-note user-overview-stat progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">用户总量</div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ total }}
              </div>
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
        </div>
      </template>
    </section>

    <section
      id="user-directory-filters"
      class="tab-panel space-y-4"
      role="tabpanel"
      aria-labelledby="user-tab-directory"
      :aria-hidden="activePanel === 'directory' ? 'false' : 'true'"
      v-show="activePanel === 'directory'"
    >
      <div class="list-heading user-directory-head">
        <div>
          <div class="journal-note-label">User Directory</div>
          <h2 class="list-heading__title">用户目录</h2>
        </div>

        <div class="user-directory-actions">
          <div class="user-directory-meta">共 {{ total }} 个用户</div>
          <button type="button" class="admin-btn admin-btn-ghost" @click="emit('refresh')">
            <RefreshCw class="h-4 w-4" />
            刷新列表
          </button>
          <button
            type="button"
            class="admin-btn admin-btn-primary"
            @click="emit('openCreateDialog')"
          >
            <UserPlus class="h-4 w-4" />
            创建用户
          </button>
        </div>
      </div>

      <div class="mt-5 grid gap-4">
        <label class="space-y-2">
          <span class="text-sm text-[var(--journal-muted)]">关键词</span>
          <input
            :value="keyword"
            type="text"
            class="admin-input"
            placeholder="用户名 / 邮箱 / 班级 / 学号 / 工号"
            @input="emit('updateKeyword', ($event.target as HTMLInputElement).value)"
          />
        </label>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="space-y-2">
            <span class="text-sm text-[var(--journal-muted)]">角色</span>
            <select
              :value="roleFilter"
              class="admin-input"
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

          <label class="space-y-2">
            <span class="text-sm text-[var(--journal-muted)]">状态</span>
            <select
              :value="statusFilter"
              class="admin-input"
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
      </div>
    </section>

    <div v-show="activePanel === 'directory'" class="journal-divider mt-6" aria-hidden="true" />

    <section v-show="activePanel === 'directory'" class="workspace-directory-section">
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
            class="admin-btn admin-btn-primary"
            @click="emit('openCreateDialog')"
          >
            创建第一个用户
          </button>
        </template>
      </AppEmpty>

      <template v-else>
        <div class="user-table-shell workspace-directory-list">
          <table class="user-table min-w-full text-sm">
            <thead class="user-table-head">
              <tr>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  用户
                </th>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  姓名
                </th>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  邮箱
                </th>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  角色
                </th>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  状态
                </th>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  班级
                </th>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  学号 / 工号
                </th>
                <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                  创建时间
                </th>
                <th class="px-4 py-3 text-right font-medium text-[var(--color-text-secondary)]">
                  操作
                </th>
              </tr>
            </thead>
            <tbody class="user-table-body">
              <tr v-for="user in list" :key="user.id" class="user-table-row">
                <td class="px-4 py-3 align-top">
                  <div class="min-w-0">
                    <span class="text-sm text-[var(--journal-muted)]">@{{ user.username }}</span>
                  </div>
                </td>
                <td class="px-4 py-3 align-top text-[var(--journal-ink)]">
                  {{ user.name || user.username }}
                </td>
                <td class="px-4 py-3 align-top text-[var(--journal-muted)]">
                  {{ user.email || '未填写邮箱' }}
                </td>
                <td class="px-4 py-3 align-top">
                  <div class="flex flex-wrap gap-2">
                    <span
                      v-for="role in user.roles"
                      :key="`${user.id}-${role}`"
                      class="admin-role-chip"
                    >
                      <UserRoundCheck class="h-3.5 w-3.5" />
                      {{ role }}
                    </span>
                  </div>
                </td>
                <td class="px-4 py-3 align-top">
                  <span class="admin-status-chip" :style="getUserStatusStyle(user.status)">
                    {{ user.status }}
                  </span>
                </td>
                <td class="px-4 py-3 align-top text-[var(--journal-muted)]">
                  {{ user.class_name || '未分配班级' }}
                </td>
                <td class="px-4 py-3 align-top text-[var(--journal-muted)]">
                  <div class="text-sm">
                    {{ getUserIdentity(user) }}
                  </div>
                </td>
                <td class="px-4 py-3 align-top text-[var(--journal-muted)]">
                  {{ new Date(user.created_at).toLocaleString('zh-CN') }}
                </td>
                <td class="px-4 py-3 align-top">
                  <div class="flex justify-end gap-2">
                    <button
                      type="button"
                      class="admin-btn admin-btn-ghost admin-btn-compact user-action-btn"
                      @click="emit('openEditDialog', user)"
                    >
                      编辑
                    </button>
                    <button
                      type="button"
                      class="admin-btn admin-btn-danger admin-btn-compact user-action-btn"
                      @click="emit('deleteUser', user.id)"
                    >
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="admin-pagination workspace-directory-pagination">
          <AdminPaginationControls
            :page="page"
            :total-pages="totalPages"
            :total="total"
            :total-label="`共 ${total} 个用户`"
            @change-page="emit('changePage', $event)"
          />
        </div>
      </template>
    </section>

    <section
      id="user-import-start"
      class="tab-panel space-y-4"
      role="tabpanel"
      aria-labelledby="user-tab-import"
      :aria-hidden="activePanel === 'import' ? 'false' : 'true'"
      v-show="activePanel === 'import'"
    >
      <div class="list-heading admin-section-head-intro">
        <div>
          <div class="journal-note-label">User Import</div>
          <h2 class="list-heading__title">导入用户</h2>
        </div>

        <button type="button" class="admin-btn admin-btn-primary" @click="triggerImport">
          <FileUp class="h-4 w-4" />
          批量导入
        </button>
      </div>

      <div class="journal-note">
        <div class="journal-note-label">CSV 格式</div>
        <div class="journal-note-helper">
          列顺序：`username,password,email,class_name,role,status,student_no,teacher_no,name`
        </div>
      </div>
    </section>

    <div v-show="activePanel === 'import'" class="journal-divider mt-6" aria-hidden="true" />

    <section v-show="activePanel === 'import'" class="space-y-4">
      <div class="list-heading user-import-receipt-head">
        <div>
          <div class="journal-note-label">Import Receipt</div>
          <h2 class="list-heading__title">导入回执</h2>
        </div>
      </div>

      <div v-if="importResult" class="admin-receipt">
        <p>
          创建 {{ importResult.created }}，更新 {{ importResult.updated }}，失败
          {{ importResult.failed }}
        </p>
        <ul v-if="importResult.errors?.length" class="mt-3 space-y-2 text-[var(--color-danger)]">
          <li v-for="item in importResult.errors.slice(0, 5)" :key="`${item.row}-${item.message}`">
            第 {{ item.row }} 行：{{ item.message }}
          </li>
        </ul>
      </div>
      <div v-else class="admin-empty">还没有导入记录。</div>
    </section>

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
  --page-top-tabs-gap: var(--space-7);
  --page-top-tabs-margin: 0 calc(var(--space-6) * -1) var(--space-6);
  --page-top-tabs-padding: 0 var(--space-6);
  --page-top-tabs-border: color-mix(in srgb, var(--journal-ink) 10%, transparent);
  --page-top-tab-min-height: 52px;
  --page-top-tab-padding: var(--space-2-5) 0 var(--space-3-5);
  --page-top-tab-font-size: var(--font-size-15);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  --journal-note-label-weight: 600;
  --journal-note-label-spacing: 0.15em;
  --journal-note-label-color: var(--journal-muted);
  --journal-divider-border: 1px dashed color-mix(in srgb, var(--journal-border) 70%, transparent);
  --journal-shell-dark-accent: var(--color-primary-hover);
}

.user-overview-summary {
  margin-top: var(--space-6);
  display: grid;
  gap: var(--space-4);
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

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.user-directory-head {
  gap: var(--space-4);
}

.user-directory-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
}

.user-directory-meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.admin-section-head-intro {
  position: relative;
  padding: var(--space-4) var(--space-4-5) var(--space-4) var(--space-5-5);
  border: 1px dashed color-mix(in srgb, var(--journal-border) 82%, transparent);
  border-radius: 18px;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-accent) 10%, transparent),
    transparent 72%
  );
}

.admin-section-head-intro::before {
  content: '';
  position: absolute;
  left: 0.82rem;
  top: 0.95rem;
  bottom: 0.95rem;
  width: 3px;
  border-radius: 999px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink)),
    color-mix(in srgb, var(--journal-accent) 20%, transparent)
  );
}

.admin-section-head-intro .journal-note-label {
  color: var(--journal-accent);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-compact {
  min-height: 2.35rem;
  padding: var(--space-2) var(--space-3-5);
}

.user-action-btn {
  min-height: 2rem;
  padding: var(--space-1-5) var(--space-3);
  border-radius: 0.8rem;
  font-size: var(--font-size-0-8125);
}

.admin-btn-primary {
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-primary:hover {
  background: var(--color-primary-hover);
}

.admin-btn-ghost {
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  color: var(--journal-accent);
}

.admin-btn-danger {
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
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
  overflow: hidden;
  border: 1px solid var(--user-table-border);
  border-radius: 18px;
  background: var(--journal-surface);
}

.user-table {
  border-collapse: collapse;
}

.user-table-head {
  background: var(--journal-surface-subtle);
}

.user-table-body {
  background: var(--journal-surface);
}

.user-table-row {
  border-top: 1px solid var(--user-row-divider);
  transition: background 180ms ease;
}

.user-table-row:hover,
.user-table-row:focus-within {
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--journal-surface));
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

:global([data-theme='dark']) .admin-section-head-intro {
  border-color: color-mix(in srgb, var(--journal-accent) 22%, var(--journal-border));
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-accent) 14%, transparent),
    transparent 72%
  );
}

@media (max-width: 767px) {
  .journal-hero {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .top-tabs {
    gap: var(--space-4-5);
    margin-left: calc(var(--space-4) * -1);
    margin-right: calc(var(--space-4) * -1);
    padding: 0 var(--space-4);
  }

  .user-table-shell {
    overflow-x: auto;
  }

  .user-overview-grid {
    --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
  }

  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .user-directory-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 560px) {
  .user-overview-grid {
    --admin-summary-grid-columns: 1fr;
  }
}
</style>
