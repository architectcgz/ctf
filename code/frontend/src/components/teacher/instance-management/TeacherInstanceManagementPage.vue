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

const overviewBadges = computed(() => [
  { key: 'visible', label: '当前可见', value: `${props.totalCount} 个实例` },
  { key: 'running', label: '运行中', value: `${props.runningCount} 个实例` },
  { key: 'scope', label: '当前范围', value: selectedClassLabel.value },
])

const instanceTips = computed(() => [
  `当前筛选范围为 ${selectedClassLabel.value}。`,
  props.expiringSoonCount > 0
    ? `有 ${props.expiringSoonCount} 个实例即将到期，建议优先跟进。`
    : '当前没有即将到期实例，可优先关注异常或新建中的环境。',
  '先筛班级与学员，再进入实例列表定位异常、到期与销毁动作。',
])

const overviewMetrics = computed(() => [
  { key: 'visible', label: '当前可见', value: props.totalCount, hint: '符合当前筛选条件的实例数量' },
  { key: 'running', label: '运行中', value: props.runningCount, hint: '仍在占用环境资源的实例数量' },
  { key: 'expiring', label: '即将到期', value: props.expiringSoonCount, hint: '剩余时间不足 10 分钟的实例数量' },
  { key: 'class-count', label: '班级池', value: props.classes.length, hint: '当前可用于筛选的班级数量' },
])

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
  <div class="teacher-management-shell teacher-surface space-y-6">
    <section class="teacher-hero teacher-surface-hero rounded-[30px] px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-eyebrow-row">
            <div class="journal-eyebrow">Teacher Instance Ops</div>
            <span class="teacher-class-chip">{{ selectedClassLabel }}</span>
          </div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            实例管理
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            聚焦教师对学生训练实例的查看与处置，先筛班级与学员，再快速定位异常或即将到期的实例。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="teacher-btn teacher-surface-btn" @click="emit('openDashboard')">
              返回教学概览
            </button>
          </div>
        </div>

        <article class="teacher-brief teacher-surface-brief journal-brief rounded-[24px] border px-5 py-5">
          <div class="teacher-brief-title">当前实例概况</div>
          <div class="teacher-badge-grid mt-5">
            <div v-for="badge in overviewBadges" :key="badge.key" class="teacher-badge-card">
              <div class="teacher-badge-label">{{ badge.label }}</div>
              <div class="teacher-badge-value">{{ badge.value }}</div>
            </div>
          </div>

          <div class="teacher-tip-block mt-5">
            <div class="teacher-tip-title">当前处理建议</div>
            <ul class="teacher-tip-list mt-3">
              <li v-for="(tip, index) in instanceTips" :key="tip" class="teacher-tip-item">
                <span class="teacher-tip-index">{{ index + 1 }}</span>
                <span>{{ tip }}</span>
              </li>
            </ul>
          </div>
        </article>
      </div>

      <div class="teacher-metric-grid mt-6">
        <article
          v-for="item in overviewMetrics"
          :key="item.key"
          class="teacher-metric-card journal-metric rounded-[20px] border px-4 py-4"
        >
          <div class="teacher-metric-label">{{ item.label }}</div>
          <div class="teacher-metric-value">{{ item.value }}</div>
          <div class="teacher-metric-hint">{{ item.hint }}</div>
        </article>
      </div>

      <div class="teacher-board teacher-surface-board">
        <section class="teacher-anchor-section teacher-surface-section">
          <div>
            <div class="journal-eyebrow teacher-eyebrow--soft">Instance Filters</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">实例筛选</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              当前范围：{{ selectedClassLabel }}。支持按班级、用户名关键字、学号精确筛选。
            </p>
          </div>

          <form class="mt-5 grid gap-4 md:grid-cols-[220px_1fr_1fr]" @submit.prevent="emit('submit')">
          <label class="space-y-2">
            <span class="text-sm text-text-secondary">班级</span>
            <select
              :value="className"
              class="teacher-filter-field teacher-surface-filter w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-text-primary outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="loadingClasses || (!isAdmin && classes.length <= 1)"
              @change="emit('updateClassName', ($event.target as HTMLSelectElement).value)"
            >
              <option v-if="isAdmin" value="">全部班级</option>
              <option v-for="item in classes" :key="item.name" :value="item.name">
                {{ item.name }} · {{ item.student_count || 0 }}
              </option>
            </select>
          </label>

          <label class="space-y-2">
            <span class="text-sm text-text-secondary">用户名关键字</span>
            <div
              class="teacher-filter-field teacher-surface-filter flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
            >
              <Search class="h-4 w-4 text-text-muted" />
              <input
                :value="keyword"
                type="text"
                placeholder="按用户名关键字搜索"
                class="w-full bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
                @input="emit('updateKeyword', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </label>

          <label class="space-y-2">
            <span class="text-sm text-text-secondary">按学号查询</span>
            <div
              class="teacher-filter-field teacher-surface-filter flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
            >
              <Search class="h-4 w-4 text-text-muted" />
              <input
                :value="studentNo"
                type="text"
                placeholder="输入学号精确查询"
                class="w-full bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
                @input="emit('updateStudentNo', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </label>

          <div class="md:col-span-3 flex flex-wrap items-center justify-end gap-3">
            <button type="button" class="teacher-btn teacher-surface-btn" @click="emit('reset')">
              重置筛选
            </button>
            <button
              type="submit"
              class="teacher-btn teacher-btn--primary teacher-surface-btn teacher-surface-btn--primary"
            >
              查询实例
            </button>
          </div>
          </form>
        </section>

        <section class="teacher-anchor-section teacher-surface-section">
          <div>
            <div class="journal-eyebrow teacher-eyebrow--soft">Instance List</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">实例列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              按当前筛选结果查看学生实例，并执行销毁等处置动作。
            </p>
          </div>

          <div v-if="loadingInstances" class="mt-5 space-y-3">
          <div
            v-for="index in 6"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]"
          />
          </div>

          <AppEmpty
            v-else-if="instances.length === 0"
            class="teacher-surface-empty mt-5"
            icon="Inbox"
            title="当前没有匹配到实例"
            description="可以调整筛选条件，或等待学员创建新的训练环境后再查看。"
          />

          <div v-else class="mt-5">
            <ElTable
              :data="instances"
              row-key="id"
              class="teacher-instance-table teacher-surface-table"
              empty-text="没有匹配实例"
            >
            <ElTableColumn label="学生 / 班级" min-width="220">
              <template #default="{ row }">
                <div class="py-1">
                  <div class="font-semibold text-text-primary">
                    {{ row.student_name || row.student_username }}
                  </div>
                  <div class="mt-1 text-sm text-text-secondary">
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
                  <div class="font-semibold text-text-primary">{{ row.challenge_title }}</div>
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
                <span class="text-sm text-text-secondary">{{
                  formatDateTime(row.expires_at)
                }}</span>
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
                <span class="text-sm text-text-secondary">{{
                  formatDateTime(row.created_at)
                }}</span>
              </template>
            </ElTableColumn>

            <ElTableColumn label="操作" width="160" align="right">
              <template #default="{ row }">
                <ElButton
                  plain
                  type="danger"
                  :disabled="destroyingId === row.id"
                  :data-instance-id="row.id"
                  @click="emit('destroy', row.id)"
                >
                  <Trash2 class="mr-1 h-4 w-4" />
                  {{ destroyingId === row.id ? '销毁中...' : '销毁实例' }}
                </ElButton>
              </template>
            </ElTableColumn>
            </ElTable>
          </div>
        </section>

        <div v-if="error" class="teacher-error-card teacher-surface-error" role="alert" aria-live="polite">
          {{ error }}
          <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
:deep(.teacher-filter-field) {
  color: var(--journal-ink);
  border-color: var(--journal-border) !important;
  background: var(--journal-surface) !important;
}

:deep(.teacher-filter-field option) {
  background-color: var(--journal-surface);
  color: var(--journal-ink);
}

:deep(.teacher-filter-field select),
:deep(.teacher-filter-field input) {
  color: var(--journal-ink);
}

:deep(.teacher-filter-field:focus-within) {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent) !important;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

:deep(.teacher-instance-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-expanded-cell-bg-color: transparent;
  --el-table-header-bg-color: var(--journal-surface);
  --el-table-border-color: var(--journal-border);
  --el-table-row-hover-bg-color: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  --el-table-text-color: var(--journal-ink);
  --el-table-header-text-color: var(--journal-muted);
}

:deep(.teacher-instance-table th.el-table__cell) {
  background: var(--journal-surface);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

:deep(.teacher-instance-table td.el-table__cell),
:deep(.teacher-instance-table th.el-table__cell) {
  border-bottom-color: var(--journal-border);
}

:deep(.teacher-instance-table.el-table),
:deep(.teacher-instance-table .el-table__inner-wrapper),
:deep(.teacher-instance-table .el-table__body-wrapper),
:deep(.teacher-instance-table .el-table__header-wrapper),
:deep(.teacher-instance-table .el-table__empty-block) {
  background: var(--journal-surface);
}

:deep(.teacher-instance-table .el-table__inner-wrapper::before) {
  display: none;
}

.teacher-management-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --color-primary: #4f46e5;
  --color-primary-hover: #4338ca;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.2rem 0.72rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
}

.teacher-eyebrow--soft {
  opacity: 0.88;
}

.teacher-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 14%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-base))
    );
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.teacher-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.teacher-brief-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.5rem;
  border-radius: 0.9rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.55rem 1.1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-btn:hover {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 12px 24px rgba(79, 70, 229, 0.18);
}

.teacher-btn--primary:hover {
  background: var(--journal-accent-strong);
  border-color: transparent;
}

.teacher-eyebrow-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.teacher-class-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.3rem 0.75rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.teacher-badge-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.teacher-badge-card {
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: var(--journal-surface);
  padding: 0.9rem 0.95rem;
}

.teacher-badge-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-badge-value {
  margin-top: 0.55rem;
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-tip-block {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
  padding-top: 1rem;
}

.teacher-tip-title {
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-tip-list {
  display: grid;
  gap: 0.6rem;
}

.teacher-tip-item {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  font-size: 0.83rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.teacher-tip-index {
  display: inline-flex;
  min-width: 1.2rem;
  justify-content: center;
  margin-top: 0.04rem;
  font-family:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',
    monospace;
  font-size: 0.76rem;
  font-weight: 700;
  color: var(--journal-accent);
}

.teacher-metric-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin-top: 1.5rem;
}

.teacher-metric-card {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.teacher-metric-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-metric-value {
  margin-top: 0.55rem;
  font-size: 1.18rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-metric-hint {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-board {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 92%, transparent);
  padding-top: 1.25rem;
  margin-top: 1.25rem;
}

.teacher-board > * + * {
  margin-top: 1.25rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
  padding-top: 1.25rem;
}

.teacher-anchor-section {
  scroll-margin-top: 84px;
}

.teacher-error-card {
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--color-danger) 22%, var(--journal-border));
  background: color-mix(in srgb, var(--color-danger) 6%, transparent);
  padding: 1rem 1rem 1.1rem;
  color: var(--journal-muted);
}

@media (max-width: 1279px) {
  .teacher-metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 639px) {
  .teacher-badge-grid,
  .teacher-metric-grid {
    grid-template-columns: 1fr;
  }
}
</style>
