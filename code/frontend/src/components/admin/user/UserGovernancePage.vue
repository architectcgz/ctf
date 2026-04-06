<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import {
  FileUp,
  RefreshCw,
  UserPlus,
  UsersRound,
  UserRoundCheck,
} from 'lucide-vue-next'

import type { AdminUserImportData, AdminUserListItem, UserStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
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

const importInput = useTemplateRef<HTMLInputElement>('importInput')
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const activeCount = computed(() => props.list.filter((item) => item.status === 'active').length)
const teacherCount = computed(() => props.list.filter((item) => item.roles.includes('teacher')).length)
const importSummary = computed(() => {
  if (!props.importResult) return '暂无导入记录'
  return `创建 ${props.importResult.created} / 更新 ${props.importResult.updated}`
})

function getUserAccentColor(status: UserStatus): string {
  switch (status) {
    case 'active':
      return 'var(--color-primary)'
    case 'locked':
      return '#f59e0b'
    case 'banned':
      return '#dc2626'
    case 'inactive':
      return '#64748b'
    default:
      return 'var(--color-primary)'
  }
}

function getUserStatusStyle(status: UserStatus): Record<string, string> {
  const accent = getUserAccentColor(status)
  return {
    color: accent,
    borderColor: `color-mix(in srgb, ${accent} 18%, transparent)`,
    backgroundColor: `color-mix(in srgb, ${accent} 10%, white)`,
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
  <section class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">User Governance</div>
          <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            用户治理台
          </h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            在这里筛选账号、批量导入并处理用户状态。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="admin-btn admin-btn-ghost" @click="emit('refresh')">
              <RefreshCw class="h-4 w-4" />
              刷新列表
            </button>
            <button type="button" class="admin-btn admin-btn-ghost" @click="triggerImport">
              <FileUp class="h-4 w-4" />
              批量导入
            </button>
            <button type="button" class="admin-btn admin-btn-primary" @click="emit('openCreateDialog')">
              <UserPlus class="h-4 w-4" />
              创建用户
            </button>
          </div>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <UsersRound class="h-5 w-5 text-[var(--journal-accent)]" />
            当前治理概况
          </div>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div class="journal-note">
              <div class="journal-note-label">用户总量</div>
              <div class="journal-note-value">{{ total }}</div>
              <div class="journal-note-helper">当前筛选条件下的用户总数</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">活跃账号</div>
              <div class="journal-note-value">{{ activeCount }}</div>
              <div class="journal-note-helper">当前页处于 active 的账号</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">教师角色</div>
              <div class="journal-note-value">{{ teacherCount }}</div>
              <div class="journal-note-helper">当前页教师账号数量</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">导入回执</div>
              <div class="journal-note-value">{{ importSummary }}</div>
              <div class="journal-note-helper">最近一次导入结果</div>
            </div>
          </div>
        </article>
      </div>
      <div class="journal-divider mt-6" />

      <div class="admin-section-head admin-section-head-intro">
        <div>
          <div class="journal-note-label">Filters</div>
          <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">筛选与导入</h2>
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
            <span class="text-sm text-[var(--journal-muted)]">学生学号</span>
            <input
              :value="studentNo"
              type="text"
              class="admin-input"
              placeholder="按学号筛选"
              @input="emit('updateStudentNo', ($event.target as HTMLInputElement).value)"
            />
          </label>

          <label class="space-y-2">
            <span class="text-sm text-[var(--journal-muted)]">教师工号</span>
            <input
              :value="teacherNo"
              type="text"
              class="admin-input"
              placeholder="按工号筛选"
              @input="emit('updateTeacherNo', ($event.target as HTMLInputElement).value)"
            />
          </label>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="space-y-2">
            <span class="text-sm text-[var(--journal-muted)]">角色</span>
            <select
              :value="roleFilter"
              class="admin-input"
              @change="emit('updateRoleFilter', ($event.target as HTMLSelectElement).value as UserFilterRole)"
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

      <div class="journal-divider mt-6" />

      <section class="space-y-4">
        <div class="admin-section-head">
          <div>
            <div class="journal-note-label">Users</div>
            <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">用户列表</h2>
          </div>
        </div>

        <div v-if="loading && list.length === 0" class="flex justify-center py-10">
          <AppLoading>正在同步用户列表...</AppLoading>
        </div>

        <AppEmpty
          v-else-if="list.length === 0"
          title="暂无用户"
          description="当前筛选条件下没有匹配用户。"
          icon="UsersRound"
        >
          <template #action>
            <button type="button" class="admin-btn admin-btn-primary" @click="emit('openCreateDialog')">
              创建第一个用户
            </button>
          </template>
        </AppEmpty>

        <template v-else>
          <div class="user-table-shell">
            <table class="user-table min-w-full text-sm">
              <thead class="user-table-head">
                <tr>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">用户</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">姓名</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">邮箱</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">角色</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">状态</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">班级</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">学号 / 工号</th>
                  <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">创建时间</th>
                  <th class="px-4 py-3 text-right font-medium text-[var(--color-text-secondary)]">操作</th>
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

          <div class="admin-pagination">
            <span>共 {{ total }} 个用户</span>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="page <= 1"
                @click="emit('changePage', page - 1)"
              >
                上一页
              </button>
              <span>{{ page }} / {{ totalPages }}</span>
              <button
                type="button"
                class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="page >= totalPages"
                @click="emit('changePage', page + 1)"
              >
                下一页
              </button>
            </div>
          </div>
        </template>
      </section>

      <div class="journal-divider mt-6" />

      <section class="space-y-4">
        <div class="admin-section-head">
          <div>
            <div class="journal-note-label">Import Receipt</div>
            <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">导入回执</h2>
          </div>
        </div>

        <div class="journal-note">
          <div class="journal-note-helper">
            CSV 列顺序：`username,password,email,class_name,role,status,student_no,teacher_no,name`
          </div>
        </div>

        <div v-if="importResult" class="admin-receipt">
          <p>
            创建 {{ importResult.created }}，更新 {{ importResult.updated }}，失败 {{ importResult.failed }}
          </p>
          <ul v-if="importResult.errors?.length" class="mt-3 space-y-2 text-[var(--color-danger)]">
            <li v-for="item in importResult.errors.slice(0, 5)" :key="`${item.row}-${item.message}`">
              第 {{ item.row }} 行：{{ item.message }}
            </li>
          </ul>
        </div>
        <div v-else class="admin-empty">
          还没有导入记录。
        </div>
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
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --user-table-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --user-row-divider: color-mix(in srgb, var(--journal-border) 58%, transparent);
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    );
  border-radius: 16px !important;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.journal-divider,
.journal-divider {
  border-top: 1px dashed rgba(148, 163, 184, 0.7);
}

.admin-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-section-head-intro {
  position: relative;
  padding: 1rem 1.1rem 1rem 1.35rem;
  border: 1px dashed color-mix(in srgb, var(--journal-border) 82%, transparent);
  border-radius: 18px;
  background: linear-gradient(90deg, color-mix(in srgb, var(--journal-accent) 10%, transparent), transparent 72%);
}

.admin-section-head-intro::before {
  content: '';
  position: absolute;
  left: 0.82rem;
  top: 0.95rem;
  bottom: 0.95rem;
  width: 3px;
  border-radius: 999px;
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink)), color-mix(in srgb, var(--journal-accent) 20%, transparent));
}

.admin-section-head-intro .journal-note-label {
  color: var(--journal-accent);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-compact {
  min-height: 2.35rem;
  padding: 0.5rem 0.85rem;
}

.user-action-btn {
  min-height: 2rem;
  padding: 0.35rem 0.7rem;
  border-radius: 0.8rem;
  font-size: 0.8125rem;
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
  border: 1px solid rgba(239, 68, 68, 0.2);
  background: rgba(254, 242, 242, 0.9);
  color: #dc2626;
}

.admin-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--admin-control-border);
  background: var(--journal-surface);
  padding: 0.7rem 1rem;
  font-size: 0.875rem;
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
  padding: 1rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
}

.admin-empty {
  border: 1px dashed rgba(148, 163, 184, 0.72);
  border-radius: 16px;
  padding: 1rem;
  font-size: 0.875rem;
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
  gap: 0.35rem;
  border-radius: 999px;
  padding: 0.34rem 0.75rem;
  font-size: 0.72rem;
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

.admin-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.72);
  padding: 1rem 1.1rem 1.05rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: color-mix(in srgb, var(--color-text-primary) 88%, var(--color-text-secondary));
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary-hover);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 16%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
}

:global([data-theme='dark']) .admin-section-head-intro {
  border-color: color-mix(in srgb, var(--journal-accent) 22%, var(--journal-border));
  background: linear-gradient(90deg, color-mix(in srgb, var(--journal-accent) 14%, transparent), transparent 72%);
}

@media (max-width: 767px) {
  .journal-hero {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .user-table-shell {
    overflow-x: auto;
  }
}
</style>
