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
  <div class="teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
    >
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading">
            <div class="teacher-surface-eyebrow journal-eyebrow">Teacher Instance Ops</div>
            <h1 class="teacher-title">实例管理</h1>
            <p class="teacher-copy">先筛班级与学员，再快速定位异常或即将到期的训练实例。</p>
          </div>

          <div class="teacher-actions">
            <button type="button" class="teacher-btn teacher-btn--primary" @click="emit('openDashboard')">
              返回教学概览
            </button>
          </div>
        </header>

        <section class="teacher-summary">
          <div class="teacher-summary-title">
            <span>Instance Snapshot</span>
          </div>
          <div class="teacher-summary-grid">
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">当前可见</div>
              <div class="teacher-summary-value">{{ totalCount }}</div>
              <div class="teacher-summary-helper">符合当前筛选条件的实例数量</div>
            </div>
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">运行中</div>
              <div class="teacher-summary-value">{{ runningCount }}</div>
              <div class="teacher-summary-helper">仍在占用环境资源的实例数量</div>
            </div>
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">即将到期</div>
              <div class="teacher-summary-value">{{ expiringSoonCount }}</div>
              <div class="teacher-summary-helper">剩余时间不足 10 分钟的实例数量</div>
            </div>
          </div>
        </section>

        <section class="teacher-controls">
          <div class="teacher-controls-bar">
            <div class="teacher-controls-heading">
              <div class="teacher-surface-eyebrow journal-eyebrow">Instance Filters</div>
              <h3 class="teacher-controls-title">实例筛选</h3>
              <p class="teacher-controls-copy">
                {{ selectedClassLabel }}。支持按班级、用户名关键字、学号精确筛选。
              </p>
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
              <button type="button" class="teacher-btn teacher-btn--ghost" @click="emit('reset')">
                重置筛选
              </button>
              <button type="submit" class="teacher-btn teacher-btn--primary">查询实例</button>
            </div>
          </form>
        </section>

        <div class="teacher-hero-divider" />

        <div v-if="loadingInstances" class="teacher-skeleton-list">
          <div
            v-for="index in 6"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
          />
        </div>

        <AppEmpty
          v-else-if="instances.length === 0"
          class="teacher-empty-state"
          icon="Inbox"
          title="当前没有匹配到实例"
          description="可以调整筛选条件，或等待学员创建新的训练环境后再查看。"
        />

        <section v-else class="teacher-directory" aria-label="实例目录">
          <div class="teacher-directory-top">
            <h3 class="teacher-directory-title">实例目录</h3>
            <div class="teacher-directory-meta">当前 {{ instances.length }} 条记录</div>
          </div>

          <div class="teacher-directory-head">
            <span>学生 / 题目</span>
            <span>标签</span>
            <span>状态</span>
            <span>数据</span>
            <span>操作</span>
          </div>

          <div v-for="item in instances" :key="item.id" class="teacher-directory-row">
            <div class="teacher-directory-row-main">
              <div class="teacher-directory-row-index">{{ item.student_no || `@${item.student_username}` }}</div>
              <div class="teacher-directory-row-title-group">
                <h4 class="teacher-directory-row-title">{{ item.student_name || item.student_username }}</h4>
                <div class="teacher-directory-row-points">{{ item.challenge_title }}</div>
              </div>
              <div class="teacher-directory-row-copy">
                @{{ item.student_username }} · {{ item.class_name }} ·
                <span class="teacher-directory-url">{{ item.access_url || '暂未分配访问地址' }}</span>
              </div>
            </div>

            <div class="teacher-directory-row-tags">
              <span class="teacher-directory-chip">Instance</span>
              <span class="teacher-directory-chip teacher-directory-chip-muted">
                创建于 {{ formatDateTime(item.created_at) }}
              </span>
            </div>

            <div class="teacher-directory-row-status">
              <span class="teacher-directory-state-chip border" :class="statusMeta(item.status).chipClass">
                {{ statusMeta(item.status).label }}
              </span>
            </div>

            <div class="teacher-directory-row-metrics">
              <span>到期 {{ formatDateTime(item.expires_at) }}</span>
              <span>剩余 {{ formatRemainingTime(item.remaining_time) }}</span>
              <span>延期 {{ remainingExtends(item) }} / {{ item.max_extends }}</span>
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
  font-family:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
}

.teacher-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 7%, transparent), transparent 22rem),
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
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-heading {
  min-width: 0;
}

.teacher-title {
  margin-top: 0.75rem;
  font-size: clamp(32px, 4vw, 46px);
  line-height: 1.02;
  letter-spacing: -0.04em;
  color: var(--journal-ink);
}

.teacher-copy {
  margin-top: 0.75rem;
  max-width: 42.5rem;
  font-size: 0.9rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.teacher-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.5rem;
  padding: 0 0.95rem;
  border: 1px solid var(--teacher-control-border);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.teacher-btn:hover,
.teacher-btn:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  outline: none;
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: var(--color-bg-base);
}

.teacher-btn--primary:hover,
.teacher-btn--primary:focus-visible {
  border-color: transparent;
  background: var(--journal-accent-strong);
  color: var(--color-bg-base);
}

.teacher-btn--ghost {
  background: color-mix(in srgb, var(--journal-surface) 84%, transparent);
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

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
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

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}

.teacher-filter-grid {
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
  margin-top: 1.5rem;
  display: grid;
  gap: 0.75rem;
}

.teacher-empty-state {
  margin-top: 1.5rem;
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

.teacher-directory-head {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(220px, 0.9fr) 8rem 11rem 9rem;
  gap: 1rem;
  padding: 0 0 0.75rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-directory-row {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(220px, 0.9fr) 8rem 11rem 9rem;
  gap: 1rem;
  align-items: center;
  padding: 1.1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-directory-row-main {
  display: grid;
  gap: 0.5rem;
  min-width: 0;
}

.teacher-directory-row-index,
.teacher-directory-row-points,
.teacher-directory-url {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.teacher-directory-row-index {
  font-size: 0.76rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--journal-muted);
}

.teacher-directory-row-title-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.6rem 0.85rem;
}

.teacher-directory-row-title {
  min-width: 0;
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
  font-size: 1.08rem;
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
}

.teacher-directory-row-points {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.teacher-directory-row-copy {
  font-size: 0.84rem;
  line-height: 1.6;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-directory-url {
  word-break: break-all;
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

.teacher-directory-row-metrics {
  display: grid;
  gap: 0.25rem;
  font-size: 0.81rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: flex;
  justify-content: flex-start;
}

.teacher-row-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.42rem;
  border-radius: 0.75rem;
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
  .teacher-topbar {
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
    gap: 0.85rem;
    padding: 1rem 0;
  }
}
</style>
