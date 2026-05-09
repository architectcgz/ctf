<script setup lang="ts">
import { Activity, Clock3, Eye, Search, Trash2 } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherInstanceItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  instances: TeacherInstanceItem[]
  className: string
  keyword: string
  studentNo: string
  loadingClasses: boolean
  loadingInstances: boolean
  destroyingId: string
  error: string | null
  isAdmin: boolean
  totalCount: number
  runningCount: number
  expiringSoonCount: number
  page: number
  totalPages: number
}>()

const emit = defineEmits<{
  retry: []
  openDashboard: []
  updateClassName: [value: string]
  updateKeyword: [value: string]
  updateStudentNo: [value: string]
  destroy: [id: string]
  changePage: [page: number]
}>()

const instanceTableColumns = [
  {
    key: 'student',
    label: '学生',
    widthClass: 'w-[20%] min-w-[14rem]',
    cellClass: 'teacher-instance-table__student-cell',
  },
  {
    key: 'challenge',
    label: '题目',
    widthClass: 'w-[17%] min-w-[12rem]',
    cellClass: 'teacher-instance-table__challenge-cell',
  },
  {
    key: 'status',
    label: '状态',
    widthClass: 'w-[9%] min-w-[6rem]',
    cellClass: 'teacher-instance-table__status-cell',
  },
  {
    key: 'created_at',
    label: '创建时间',
    widthClass: 'w-[12%] min-w-[9rem]',
    cellClass: 'teacher-instance-table__time-cell',
  },
  {
    key: 'expires_at',
    label: '到期时间',
    widthClass: 'w-[12%] min-w-[9rem]',
    cellClass: 'teacher-instance-table__time-cell',
  },
  {
    key: 'extends',
    label: '延期',
    align: 'center' as const,
    widthClass: 'w-[7%] min-w-[5rem]',
    cellClass: 'teacher-instance-table__compact-cell',
  },
  {
    key: 'remaining',
    label: '剩余时间',
    widthClass: 'w-[9%] min-w-[7rem]',
    cellClass: 'teacher-instance-table__compact-cell',
  },
  {
    key: 'access_url',
    label: '访问地址',
    widthClass: 'w-[14%] min-w-[12rem]',
    cellClass: 'teacher-instance-table__url-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[8rem]',
    cellClass: 'teacher-instance-table__actions-cell',
  },
]

function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatRemainingTime(seconds: number): string {
  if (seconds <= 0) return '已到期'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const restSeconds = seconds % 60
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(restSeconds).padStart(2, '0')}`
}

function statusMeta(status: string): { label: string; chipClass: string } {
  switch (status) {
    case 'running':
      return {
        label: '运行中',
        chipClass: 'instance-status-pill--running',
      }
    case 'creating':
      return {
        label: '创建中',
        chipClass: 'instance-status-pill--pending',
      }
    case 'expired':
      return {
        label: '已过期',
        chipClass: 'instance-status-pill--inactive',
      }
    case 'failed':
      return {
        label: '异常',
        chipClass: 'instance-status-pill--danger',
      }
    default:
      return { label: status, chipClass: 'instance-status-pill--inactive' }
  }
}
</script>

<template>
  <div
    class="workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading workspace-tab-heading__main">
            <div class="workspace-overline">Teacher Instance Ops</div>
            <h1 class="teacher-title workspace-page-title">实例管理</h1>
            <p class="teacher-copy workspace-page-copy">
              先筛班级与学员，再快速定位异常或即将到期的训练实例。
            </p>
          </div>

          <div class="teacher-actions">
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('openDashboard')"
            >
              返回教学概览
            </button>
          </div>
        </header>

        <section class="teacher-summary metric-panel-default-surface">
          <div class="teacher-summary-title">
            <span>Instance Snapshot</span>
          </div>
          <div
            class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"
          >
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                <span>当前可见</span>
                <Eye class="h-4 w-4" />
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ totalCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">符合当前筛选条件的实例数量</div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                <span>运行中</span>
                <Activity class="h-4 w-4" />
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ runningCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">仍在占用环境资源的实例数量</div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                <span>即将到期</span>
                <Clock3 class="h-4 w-4" />
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ expiringSoonCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                剩余时间不足 10 分钟的实例数量
              </div>
            </article>
          </div>
        </section>

        <section
          class="workspace-directory-section teacher-directory-section"
          aria-label="实例目录"
        >
          <section class="teacher-directory-shell workspace-directory-list">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">Instance Directory</div>
                <h3 class="list-heading__title">实例目录</h3>
              </div>
              <div class="teacher-directory-meta">共 {{ totalCount }} 条记录</div>
            </header>

            <section class="teacher-directory-filters" aria-label="实例过滤">
              <div class="teacher-filter-grid">
                <label class="teacher-field">
                  <span class="teacher-field-label">班级</span>
                  <select
                    :value="className"
                    class="teacher-field-control"
                    :disabled="loadingClasses || (!isAdmin && classes.length <= 1)"
                    @change="emit('updateClassName', ($event.target as HTMLSelectElement).value)"
                  >
                    <option v-if="isAdmin" value="">全部班级</option>
                    <option v-for="item in classes" :key="item.name" :value="item.name">
                      {{ item.name }} · {{ item.student_count || 0 }}
                    </option>
                  </select>
                </label>

                <label class="teacher-field">
                  <span class="teacher-field-label">用户关键字</span>
                  <div class="teacher-field-control teacher-filter-control">
                    <Search class="h-4 w-4 text-text-muted" />
                    <input
                      :value="keyword"
                      type="text"
                      placeholder="按用户名或学号搜索"
                      class="teacher-input"
                      @input="emit('updateKeyword', ($event.target as HTMLInputElement).value)"
                    />
                  </div>
                </label>

                <label class="teacher-field">
                  <span class="teacher-field-label">按学号查询</span>
                  <div class="teacher-field-control teacher-filter-control">
                    <Search class="h-4 w-4 text-text-muted" />
                    <input
                      :value="studentNo"
                      type="text"
                      placeholder="输入学号精确查询"
                      class="teacher-input"
                      @input="emit('updateStudentNo', ($event.target as HTMLInputElement).value)"
                    />
                  </div>
                </label>
              </div>
            </section>

            <div v-if="loadingInstances" class="teacher-skeleton-list workspace-directory-loading">
              <div
                v-for="index in 6"
                :key="index"
                class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
              />
            </div>

            <AppEmpty
              v-else-if="instances.length === 0"
              class="teacher-empty-state workspace-directory-empty"
              icon="Inbox"
              title="当前没有匹配到实例"
              description="可以调整筛选条件，或等待学员创建新的训练环境后再查看。"
            />

            <template v-else>
              <WorkspaceDataTable
                class="teacher-instance-list"
                :columns="instanceTableColumns"
                :rows="instances"
                row-key="id"
                row-class="teacher-directory-row teacher-instance-table-row"
              >
                <template #cell-student="{ row }">
                  <div class="teacher-instance-user-cell">
                    <span class="teacher-instance-user-meta">
                      {{
                        (row as TeacherInstanceItem).student_no ||
                        `@${(row as TeacherInstanceItem).student_username}`
                      }}
                    </span>
                    <span
                      class="teacher-instance-primary-text"
                      :title="
                        (row as TeacherInstanceItem).student_name ||
                        (row as TeacherInstanceItem).student_username
                      "
                    >
                      {{
                        (row as TeacherInstanceItem).student_name ||
                        (row as TeacherInstanceItem).student_username
                      }}
                    </span>
                    <span
                      class="teacher-instance-secondary-text"
                      :title="`@${(row as TeacherInstanceItem).student_username} · ${(row as TeacherInstanceItem).class_name}`"
                    >
                      @{{ (row as TeacherInstanceItem).student_username }} ·
                      {{ (row as TeacherInstanceItem).class_name }}
                    </span>
                  </div>
                </template>

                <template #cell-challenge="{ row }">
                  <span
                    class="teacher-instance-primary-text"
                    :title="(row as TeacherInstanceItem).challenge_title"
                  >
                    {{ (row as TeacherInstanceItem).challenge_title }}
                  </span>
                </template>

                <template #cell-status="{ row }">
                  <span
                    class="instance-status-pill"
                    :class="statusMeta((row as TeacherInstanceItem).status).chipClass"
                  >
                    {{ statusMeta((row as TeacherInstanceItem).status).label }}
                  </span>
                </template>

                <template #cell-created_at="{ row }">
                  <span class="teacher-instance-muted-text">
                    {{ formatDateTime((row as TeacherInstanceItem).created_at) }}
                  </span>
                </template>

                <template #cell-expires_at="{ row }">
                  <span class="teacher-instance-muted-text">
                    {{ formatDateTime((row as TeacherInstanceItem).expires_at) }}
                  </span>
                </template>

                <template #cell-extends="{ row }">
                  <span class="teacher-instance-muted-text">
                    {{ (row as TeacherInstanceItem).extend_count }} /
                    {{ (row as TeacherInstanceItem).max_extends }}
                  </span>
                </template>

                <template #cell-remaining="{ row }">
                  <span class="teacher-instance-muted-text">
                    {{ formatRemainingTime((row as TeacherInstanceItem).remaining_time) }}
                  </span>
                </template>

                <template #cell-access_url="{ row }">
                  <span
                    class="teacher-instance-url-text"
                    :title="(row as TeacherInstanceItem).access_url || '暂未分配访问地址'"
                  >
                    {{ (row as TeacherInstanceItem).access_url || '暂未分配访问地址' }}
                  </span>
                </template>

                <template #cell-actions="{ row }">
                  <div class="teacher-directory-row-cta">
                    <button
                      type="button"
                      class="ui-btn ui-btn--danger ui-btn--xs teacher-instance-danger-action"
                      :disabled="destroyingId === (row as TeacherInstanceItem).id"
                      :data-instance-id="(row as TeacherInstanceItem).id"
                      @click="emit('destroy', (row as TeacherInstanceItem).id)"
                    >
                      <Trash2 class="h-3 w-3" />
                      {{ destroyingId === (row as TeacherInstanceItem).id ? '销毁中' : '销毁' }}
                    </button>
                  </div>
                </template>
              </WorkspaceDataTable>

              <WorkspaceDirectoryPagination
                class="teacher-directory-pagination"
                :page="page"
                :total-pages="totalPages"
                :total="totalCount"
                :total-label="`共 ${totalCount} 条实例`"
                @change-page="emit('changePage', $event)"
              />
            </template>
          </section>
        </section>
        <div v-if="error" class="teacher-surface-error">
          {{ error }}
          <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">
            重试
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --teacher-management-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --teacher-management-accent-strong: color-mix(
    in srgb,
    var(--color-primary) 74%,
    var(--journal-ink)
  );
  --teacher-management-hero-border: var(--teacher-card-border);
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-directory-section {
  margin-top: var(--workspace-directory-page-block-gap, var(--space-5));
}

.teacher-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 calc(var(--space-4) + var(--space-0-5)) calc(var(--space-8) + var(--space-0-5))
    color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-directory-filters {
  display: grid;
  gap: var(--space-4);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: 220px minmax(0, 1fr) minmax(0, 1fr);
}

.teacher-skeleton-list {
  display: grid;
  gap: var(--space-3);
}

.teacher-instance-list {
  --workspace-directory-shell-border: color-mix(
    in srgb,
    var(--teacher-card-border) 86%,
    transparent
  );
}

.teacher-instance-user-cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-1);
}

.teacher-instance-primary-text,
.teacher-instance-secondary-text,
.teacher-instance-user-meta,
.teacher-instance-muted-text,
.teacher-instance-url-text {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-instance-primary-text {
  font-size: var(--font-size-0-875);
  font-weight: 700;
  color: var(--color-text-primary);
}

.teacher-instance-secondary-text,
.teacher-instance-muted-text {
  font-size: var(--font-size-0-8125);
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-instance-user-meta,
.teacher-instance-url-text {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-75);
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-instance-url-text {
  white-space: normal;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.instance-status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.4rem;
  padding: 0 var(--space-2);
  border-radius: 999px;
  border: 1px solid transparent;
  font-size: var(--font-size-10);
  font-weight: 700;
  text-transform: uppercase;
}

.instance-status-pill--running {
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-success) 24%, transparent);
  color: color-mix(in srgb, var(--color-success) 82%, var(--color-text-primary));
}

.instance-status-pill--pending {
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-primary) 24%, transparent);
  color: color-mix(in srgb, var(--color-primary) 82%, var(--color-text-primary));
}

.instance-status-pill--danger {
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-danger) 24%, transparent);
  color: color-mix(in srgb, var(--color-danger) 82%, var(--color-text-primary));
}

.instance-status-pill--inactive {
  background: color-mix(in srgb, var(--color-text-muted) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  color: var(--color-text-secondary);
}

.teacher-directory-row-cta {
  display: flex;
  justify-content: flex-end;
}

.teacher-instance-danger-action {
  color: var(--color-danger);
}

.teacher-instance-danger-action:hover:not(:disabled),
.teacher-instance-danger-action:focus-visible:not(:disabled) {
  color: color-mix(in srgb, var(--color-danger) 90%, var(--journal-ink));
}

@media (max-width: 1080px) {
  .teacher-topbar,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid,
  .teacher-filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
