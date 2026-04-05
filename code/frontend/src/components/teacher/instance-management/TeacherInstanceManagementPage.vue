<script setup lang="ts">
import { computed } from 'vue'
import { Search, Trash2 } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherInstanceItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

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
}>()

const emit = defineEmits<{
  retry: []
  submit: []
  reset: []
  openDashboard: []
  updateClassName: [value: string]
  updateKeyword: [value: string]
  updateStudentNo: [value: string]
  destroy: [id: string]
}>()

const selectedClassLabel = computed(() => {
  if (!props.className) return props.isAdmin ? '全部班级' : '未设置班级'
  return props.className
})

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

function remainingExtends(item: TeacherInstanceItem): number {
  return Math.max(0, item.max_extends - item.extend_count)
}
</script>

<template>
  <div class="teacher-management-shell teacher-surface">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <header class="teacher-header">
        <div class="teacher-header__main">
          <div class="teacher-surface-eyebrow journal-eyebrow">Teacher Instance Ops</div>
          <h2 class="teacher-title">实例管理</h2>
          <p class="teacher-copy">先筛班级与学员，再快速定位异常或即将到期的训练实例。</p>

          <div class="teacher-actions">
            <button type="button" class="teacher-btn teacher-surface-btn" @click="emit('openDashboard')">
              返回教学概览
            </button>
          </div>
        </div>

        <div class="teacher-badge-grid">
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">当前可见</div>
            <div class="teacher-badge-value">{{ totalCount }}</div>
            <div class="teacher-badge-hint">符合当前筛选条件的实例数量</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">运行中</div>
            <div class="teacher-badge-value">{{ runningCount }}</div>
            <div class="teacher-badge-hint">仍在占用环境资源的实例数量</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">即将到期</div>
            <div class="teacher-badge-value">{{ expiringSoonCount }}</div>
            <div class="teacher-badge-hint">剩余时间不足 10 分钟的实例数量</div>
          </article>
        </div>
      </header>

      <div class="teacher-hero-divider" />

      <div class="teacher-tip-block">
        <div class="teacher-tip-label">当前范围</div>
        <div class="teacher-tip-copy">{{ selectedClassLabel }}。支持按班级、用户名关键字、学号精确筛选。</div>
      </div>

      <div class="teacher-surface-board">
        <section class="teacher-surface-section teacher-filter-panel">
          <div class="teacher-section-head">
            <div>
              <div class="teacher-surface-eyebrow journal-eyebrow">Instance Filters</div>
              <h3 class="teacher-section-title">实例筛选</h3>
            </div>
          </div>

          <form class="teacher-filter-grid" @submit.prevent="emit('submit')">
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
              <span class="teacher-field-label">用户名关键字</span>
              <div class="teacher-field-control teacher-filter-control">
                <Search class="h-4 w-4 text-text-muted" />
                <input
                  :value="keyword"
                  type="text"
                  placeholder="按用户名关键字搜索"
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

            <div class="teacher-filter-actions">
              <button type="button" class="teacher-btn teacher-surface-btn" @click="emit('reset')">
                重置筛选
              </button>
              <button type="submit" class="teacher-btn teacher-surface-btn teacher-btn--primary">
                查询实例
              </button>
            </div>
          </form>
        </section>

        <section class="teacher-surface-section">
          <div class="teacher-section-head">
            <div>
              <div class="teacher-surface-eyebrow journal-eyebrow">Instance List</div>
              <h3 class="teacher-section-title">实例列表</h3>
            </div>
            <div class="teacher-section-meta">当前 {{ instances.length }} 条记录</div>
          </div>

          <div v-if="loadingInstances" class="teacher-skeleton-list">
            <div
              v-for="index in 6"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
            />
          </div>

          <AppEmpty
            v-else-if="instances.length === 0"
            class="mt-5"
            icon="Inbox"
            title="当前没有匹配到实例"
            description="可以调整筛选条件，或等待学员创建新的训练环境后再查看。"
          />

          <div v-else class="mt-5 teacher-table-shell">
            <ElTable
              :data="instances"
              row-key="id"
              class="teacher-surface-table teacher-instance-table"
              empty-text="没有匹配实例"
            >
              <ElTableColumn label="学生 / 班级" min-width="220">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="teacher-row-title">{{ row.student_name || row.student_username }}</div>
                    <div class="teacher-row-copy">
                      @{{ row.student_username }}
                      <span class="mx-1 text-text-muted">·</span>
                      {{ row.class_name }}
                      <span v-if="row.student_no" class="mx-1 text-text-muted">·</span>
                      <span v-if="row.student_no">学号 {{ row.student_no }}</span>
                    </div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn label="题目" min-width="220">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="teacher-row-title">{{ row.challenge_title }}</div>
                    <div class="mt-1">
                      <span
                        class="rounded-full border px-3 py-1 text-xs font-semibold"
                        :class="statusMeta(row.status).chipClass"
                      >
                        {{ statusMeta(row.status).label }}
                      </span>
                    </div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn label="访问地址" min-width="260">
                <template #default="{ row }">
                  <div class="break-all py-1 font-mono text-sm text-text-primary">
                    {{ row.access_url || '暂未分配' }}
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn label="到期时间" width="180">
                <template #default="{ row }">
                  <span class="teacher-row-copy">{{ formatDateTime(row.expires_at) }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="剩余时间" width="130" align="center">
                <template #default="{ row }">
                  <span class="font-mono text-sm font-medium text-text-primary">{{
                    formatRemainingTime(row.remaining_time)
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="延期" width="120" align="center">
                <template #default="{ row }">
                  <span class="text-sm font-medium text-text-primary"
                    >{{ remainingExtends(row) }} / {{ row.max_extends }}</span
                  >
                </template>
              </ElTableColumn>

              <ElTableColumn label="创建时间" width="180">
                <template #default="{ row }">
                  <span class="teacher-row-copy">{{ formatDateTime(row.created_at) }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="操作" width="160" align="right">
                <template #default="{ row }">
                  <button
                    type="button"
                    class="teacher-row-btn teacher-row-btn--danger"
                    :disabled="destroyingId === row.id"
                    :data-instance-id="row.id"
                    @click="emit('destroy', row.id)"
                  >
                    <Trash2 class="h-4 w-4" />
                    {{ destroyingId === row.id ? '销毁中...' : '销毁实例' }}
                  </button>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
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
  --journal-accent: #2563eb;
  --journal-accent-strong: #1d4ed8;
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-hero {
  border-color: var(--teacher-card-border);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 12%, transparent),
      transparent 18rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-base))
    );
}

.journal-eyebrow {
  letter-spacing: 0.08em;
}

.journal-brief,
.journal-metric {
  border-radius: 18px;
}

.teacher-header {
  display: grid;
  gap: 1.25rem;
}

.teacher-header__main {
  max-width: 42rem;
}

.teacher-title {
  margin-top: 0.85rem;
  font-size: clamp(2rem, 2vw, 2.45rem);
  font-weight: 700;
  line-height: 1.08;
  color: var(--journal-ink);
}

.teacher-copy {
  margin-top: 0.7rem;
  max-width: 42rem;
  font-size: 0.92rem;
  line-height: 1.72;
  color: var(--journal-muted);
}

.teacher-actions {
  margin-top: 1.3rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.75rem;
  border-radius: 999px;
  border: 1px solid var(--teacher-control-border);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.62rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-btn:hover,
.teacher-btn:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.teacher-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink));
}

.teacher-btn--primary:hover,
.teacher-btn--primary:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface));
}

.teacher-badge-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.teacher-badge-card {
  min-height: 100%;
  border: 1px solid var(--teacher-card-border);
  padding: 1rem 1.05rem 1.05rem;
}

.teacher-badge-label {
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-badge-value {
  margin-top: 0.55rem;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-badge-hint {
  margin-top: 0.5rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-hero-divider {
  margin: 1.35rem 0 1.15rem;
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-tip-block {
  display: grid;
  gap: 0.35rem;
  border-top: 1px dashed var(--teacher-divider);
  padding-top: 1rem;
}

.teacher-tip-label {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.teacher-tip-copy {
  font-size: 0.86rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.teacher-filter-panel {
  padding-top: 1.3rem;
}

.teacher-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.85rem;
}

.teacher-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-section-meta {
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.teacher-filter-grid {
  margin-top: 1rem;
  display: grid;
  gap: 1rem;
  grid-template-columns: 220px minmax(0, 1fr) minmax(0, 1fr);
}

.teacher-filter-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.75rem;
  grid-column: 1 / -1;
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

.teacher-skeleton-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.75rem;
}

.teacher-table-shell {
  border: 1px solid var(--teacher-card-border);
  border-radius: 18px;
}

.teacher-row-title {
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 88%, var(--journal-muted));
}

.teacher-row-copy {
  font-size: 0.84rem;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-row-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.42rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, var(--teacher-control-border));
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  padding: 0.58rem 0.95rem;
  font-size: 0.84rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-row-btn:hover,
.teacher-row-btn:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 38%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface));
  color: var(--journal-accent);
}

.teacher-row-btn--danger {
  border-color: color-mix(in srgb, var(--color-danger) 22%, var(--teacher-control-border));
  background: color-mix(in srgb, var(--color-danger) 8%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.teacher-row-btn--danger:hover,
.teacher-row-btn--danger:focus-visible {
  border-color: color-mix(in srgb, var(--color-danger) 36%, transparent);
  background: color-mix(in srgb, var(--color-danger) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 92%, var(--journal-ink));
}

.teacher-row-btn--danger:disabled {
  cursor: wait;
  opacity: 0.7;
}

:deep(.teacher-instance-table.el-table),
:deep(.teacher-instance-table .el-table__inner-wrapper),
:deep(.teacher-instance-table .el-scrollbar),
:deep(.teacher-instance-table .el-scrollbar__view),
:deep(.teacher-instance-table .el-table__body-wrapper),
:deep(.teacher-instance-table .el-table__header-wrapper),
:deep(.teacher-instance-table .el-table__empty-block) {
  background: var(--journal-surface);
}

@media (max-width: 1080px) {
  .teacher-badge-grid,
  .teacher-filter-grid {
    grid-template-columns: 1fr;
  }

  .teacher-filter-actions {
    justify-content: flex-start;
  }
}
</style>
