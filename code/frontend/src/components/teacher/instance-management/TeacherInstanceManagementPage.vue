<script setup lang="ts">
import { Search, Trash2 } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherInstanceItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'

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
        chipClass:
          'border-[var(--color-success)]/25 bg-[var(--color-success)]/10 text-[var(--color-success)]',
      }
    case 'creating':
      return {
        label: '创建中',
        chipClass:
          'border-[var(--color-primary)]/25 bg-[var(--color-primary)]/10 text-[var(--color-primary)]',
      }
    case 'expired':
      return {
        label: '已过期',
        chipClass:
          'border-[var(--color-warning)]/25 bg-[var(--color-warning)]/10 text-[var(--color-warning)]',
      }
    case 'failed':
      return {
        label: '异常',
        chipClass:
          'border-[var(--color-danger)]/25 bg-[var(--color-danger)]/10 text-[var(--color-danger)]',
      }
    default:
      return { label: status, chipClass: 'border-border bg-elevated/70 text-text-secondary' }
  }
}
</script>

<template>
  <div class="workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <main class="content-pane">
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading">
            <div class="teacher-surface-eyebrow journal-eyebrow">
              Teacher Instance Ops
            </div>
            <h1 class="teacher-title">
              实例管理
            </h1>
            <p class="teacher-copy">
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
          <div class="teacher-summary-grid progress-strip metric-panel-grid">
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                当前可见
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ totalCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                符合当前筛选条件的实例数量
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                运行中
              </div>
              <div class="progress-card-value metric-panel-value">
                {{ runningCount }}
              </div>
              <div class="progress-card-hint metric-panel-helper">
                仍在占用环境资源的实例数量
              </div>
            </article>
            <article class="progress-card metric-panel-card">
              <div class="progress-card-label metric-panel-label">
                即将到期
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
          <header class="list-heading">
            <div>
              <div class="journal-note-label">
                Instance Directory
              </div>
              <h3 class="list-heading__title">
                实例目录
              </h3>
            </div>
            <div class="teacher-directory-meta">
              共 {{ totalCount }} 条记录
            </div>
          </header>

          <section
            class="teacher-directory-filters"
            aria-label="实例过滤"
          >
            <div class="teacher-filter-grid">
              <label class="teacher-field">
                <span class="teacher-field-label">班级</span>
                <select
                  :value="className"
                  class="teacher-field-control"
                  :disabled="loadingClasses || (!isAdmin && classes.length <= 1)"
                  @change="emit('updateClassName', ($event.target as HTMLSelectElement).value)"
                >
                  <option
                    v-if="isAdmin"
                    value=""
                  >全部班级</option>
                  <option
                    v-for="item in classes"
                    :key="item.name"
                    :value="item.name"
                  >
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
                  >
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
                  >
                </div>
              </label>
            </div>
          </section>

          <div
            v-if="loadingInstances"
            class="teacher-skeleton-list workspace-directory-loading"
          >
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

          <section
            v-else
            class="teacher-directory"
            aria-label="实例目录"
          >
            <div class="teacher-directory-head">
              <span>学生</span>
              <span>题目</span>
              <span>标签</span>
              <span>状态</span>
              <span>创建时间</span>
              <span>到期时间</span>
              <span>延期</span>
              <span>剩余时间</span>
              <span>访问地址</span>
              <span>操作</span>
            </div>

            <div
              v-for="item in instances"
              :key="item.id"
              class="teacher-directory-row"
            >
              <div class="teacher-directory-row-main">
                <div class="teacher-directory-row-index">
                  {{ item.student_no || `@${item.student_username}` }}
                </div>
                <h4
                  class="teacher-directory-row-title"
                  :title="item.student_name || item.student_username"
                >
                  {{ item.student_name || item.student_username }}
                </h4>
                <div
                  class="teacher-directory-row-copy"
                  :title="`@${item.student_username} · ${item.class_name}`"
                >
                  @{{ item.student_username }} · {{ item.class_name }}
                </div>
              </div>

              <div
                class="teacher-directory-row-challenge"
                :title="item.challenge_title"
              >
                {{ item.challenge_title }}
              </div>

              <div class="teacher-directory-row-tags">
                <span class="teacher-directory-chip">Instance</span>
              </div>

              <div class="teacher-directory-row-status">
                <span
                  class="teacher-directory-state-chip border"
                  :class="statusMeta(item.status).chipClass"
                >
                  {{ statusMeta(item.status).label }}
                </span>
              </div>

              <div class="teacher-directory-row-created">
                {{ formatDateTime(item.created_at) }}
              </div>

              <div class="teacher-directory-row-expires-at">
                {{ formatDateTime(item.expires_at) }}
              </div>

              <div class="teacher-directory-row-extends">
                {{ item.extend_count }} / {{ item.max_extends }}
              </div>

              <div class="teacher-directory-row-remaining">
                {{ formatRemainingTime(item.remaining_time) }}
              </div>

              <div
                class="teacher-directory-row-url"
                :title="item.access_url || '暂未分配访问地址'"
              >
                {{ item.access_url || '暂未分配访问地址' }}
              </div>

              <div class="teacher-directory-row-cta">
                <button
                  type="button"
                  class="teacher-row-btn teacher-row-btn--danger"
                  :disabled="destroyingId === item.id"
                  :data-instance-id="item.id"
                  @click="emit('destroy', item.id)"
                >
                  <Trash2 class="h-4 w-4" />
                  {{ destroyingId === item.id ? '销毁中...' : '销毁实例' }}
                </button>
              </div>
            </div>

            <div
              v-if="totalCount > 0"
              class="teacher-directory-pagination workspace-directory-pagination"
            >
              <PagePaginationControls
                :page="page"
                :total-pages="totalPages"
                :total="totalCount"
                :total-label="`共 ${totalCount} 条实例`"
                @change-page="emit('changePage', $event)"
              />
            </div>
          </section>
        </section>
        <div
          v-if="error"
          class="teacher-surface-error"
        >
          {{ error }}
          <button
            type="button"
            class="ml-3 font-medium underline"
            @click="emit('retry')"
          >
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
  --teacher-directory-columns: minmax(0, 1.1fr) minmax(0, 0.92fr) 6rem 7rem minmax(132px, 0.7fr)
    minmax(132px, 0.7fr) 5rem 7rem minmax(180px, 1fr) 8rem;
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
  margin-top: var(--space-6);
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

.teacher-directory-filters {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) 0;
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

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-head,
.teacher-directory-row {
  display: grid;
  grid-template-columns: var(--teacher-directory-columns);
  gap: var(--space-4);
}

.teacher-directory-row {
  width: 100%;
  align-items: center;
  border: 0;
  padding: var(--space-4-5) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
}

.teacher-directory-row-main {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.teacher-directory-row-index,
.teacher-directory-row-challenge,
.teacher-directory-row-url {
  font-family: var(--font-family-mono);
}

.teacher-directory-row-index {
  font-size: var(--font-size-0-76);
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--journal-muted);
}

.teacher-directory-row-title {
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

.teacher-directory-row-challenge {
  min-width: 0;
  font-size: var(--font-size-0-80);
  font-weight: 700;
  color: var(--journal-accent-strong);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-directory-row-copy {
  display: -webkit-box;
  font-size: var(--font-size-0-84);
  line-height: 1.6;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
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

.teacher-directory-row-created,
.teacher-directory-row-expires-at,
.teacher-directory-row-extends,
.teacher-directory-row-remaining,
.teacher-directory-row-url {
  min-width: 0;
  font-size: var(--font-size-0-81);
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-directory-row-url {
  overflow-wrap: anywhere;
  word-break: break-word;
}

.teacher-directory-row-cta {
  display: flex;
  justify-content: flex-start;
}

.teacher-row-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1-5);
  border-radius: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, var(--teacher-control-border));
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-row-btn:hover:not(:disabled),
.teacher-row-btn:focus-visible:not(:disabled) {
  border-color: color-mix(in srgb, var(--journal-accent) 38%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface));
  color: var(--journal-accent);
}

.teacher-row-btn--danger {
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--teacher-control-border));
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: var(--color-danger);
}

.teacher-row-btn--danger:hover:not(:disabled),
.teacher-row-btn--danger:focus-visible:not(:disabled) {
  border-color: color-mix(in srgb, var(--color-danger) 40%, transparent);
  background: color-mix(in srgb, var(--color-danger) 16%, var(--journal-surface));
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
