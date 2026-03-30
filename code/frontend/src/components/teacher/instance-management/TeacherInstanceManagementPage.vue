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
  <div class="teacher-management-shell space-y-6">
    <section class="teacher-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-eyebrow">Teacher Instance Ops</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            实例管理
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            聚焦教师对学生训练实例的查看与处置，先筛班级与学员，再快速定位异常或即将到期的实例。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="teacher-btn" @click="emit('openDashboard')">
              返回教学概览
            </button>
          </div>
        </div>

        <article class="teacher-brief rounded-[24px] border px-5 py-5">
          <div class="teacher-brief-title">当前实例概况</div>
          <div class="teacher-kpi-grid mt-5 grid gap-3 sm:grid-cols-3">
            <article class="teacher-kpi-card teacher-kpi-card--primary">
              <div class="teacher-kpi-label">当前可见</div>
              <div class="teacher-kpi-value">{{ totalCount }}</div>
              <div class="teacher-kpi-hint">符合当前筛选条件的实例数量</div>
            </article>
            <article class="teacher-kpi-card teacher-kpi-card--success">
              <div class="teacher-kpi-label">运行中</div>
              <div class="teacher-kpi-value">{{ runningCount }}</div>
              <div class="teacher-kpi-hint">仍在占用环境资源的实例数量</div>
            </article>
            <article class="teacher-kpi-card teacher-kpi-card--warning">
              <div class="teacher-kpi-label">即将到期</div>
              <div class="teacher-kpi-value">{{ expiringSoonCount }}</div>
              <div class="teacher-kpi-hint">剩余时间不足 10 分钟的实例数量</div>
            </article>
          </div>
        </article>
      </div>

      <div class="teacher-hero-divider" />

      <div class="teacher-hero-section">
        <div class="teacher-hero-section-head">
          <div>
            <div class="teacher-eyebrow teacher-eyebrow--soft">Instance Filters</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">实例筛选与列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              当前范围：{{ selectedClassLabel }}。支持按班级、用户名关键字、学号精确筛选。
            </p>
          </div>
        </div>

        <form class="mt-5 grid gap-4 md:grid-cols-[220px_1fr_1fr]" @submit.prevent="emit('submit')">
          <label class="space-y-2">
            <span class="text-sm text-text-secondary">班级</span>
            <select
              :value="className"
              class="teacher-filter-field w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-text-primary outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
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
              class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
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
              class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
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
            <button type="button" class="teacher-btn" @click="emit('reset')">重置筛选</button>
            <button type="submit" class="teacher-btn teacher-btn--primary">查询实例</button>
          </div>
        </form>

        <div class="teacher-hero-divider teacher-hero-divider--inner" />

        <div v-if="loadingInstances" class="space-y-3">
          <div
            v-for="index in 6"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]"
          />
        </div>

        <AppEmpty
          v-else-if="instances.length === 0"
          class="mt-5"
          icon="Inbox"
          title="当前没有匹配到实例"
          description="可以调整筛选条件，或等待学员创建新的训练环境后再查看。"
        />

        <div v-else class="mt-5">
          <ElTable
            :data="instances"
            row-key="id"
            class="teacher-instance-table"
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
      </div>
    </section>

    <div
      v-if="error"
      class="rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 px-5 py-4 text-sm text-[var(--color-danger)]"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>
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

:deep(.teacher-instance-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-expanded-cell-bg-color: transparent;
  --el-table-header-bg-color: var(--journal-surface);
  --el-table-border-color: var(--journal-border);
  --el-table-row-hover-bg-color: rgba(99, 102, 241, 0.06);
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

:deep(.teacher-instance-table .el-table__inner-wrapper::before) {
  display: none;
}

.teacher-management-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  --color-primary: #4f46e5;
  --color-primary-hover: #4338ca;
  --color-text-primary: var(--journal-ink);
  --color-text-secondary: var(--journal-muted);
  --color-text-muted: #94a3b8;
  --color-border-default: var(--journal-border);
  --color-border-subtle: rgba(226, 232, 240, 0.74);
  --color-bg-surface: var(--journal-surface);
  --color-bg-base: #f8fafc;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.teacher-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.teacher-eyebrow--soft {
  opacity: 0.88;
}

.teacher-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.teacher-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.teacher-brief-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-hero-divider {
  margin-top: 1.5rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.teacher-hero-divider--inner {
  margin-top: 1.25rem;
}

.teacher-hero-section {
  margin-top: 1.5rem;
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
    background 0.18s ease;
}

.teacher-btn:hover {
  border-color: var(--journal-accent);
  background: rgba(99, 102, 241, 0.06);
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 12px 24px rgba(79, 70, 229, 0.18);
}

.teacher-btn--primary:hover {
  border-color: transparent;
  background: var(--color-primary-hover);
}

.teacher-kpi-grid {
  align-items: stretch;
}

.teacher-kpi-card {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: var(--journal-surface-subtle);
  padding: 0.95rem 1rem;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.teacher-kpi-card--primary {
  border-top: 3px solid rgba(79, 70, 229, 0.42);
}

.teacher-kpi-card--success {
  border-top: 3px solid rgba(16, 185, 129, 0.36);
}

.teacher-kpi-card--warning {
  border-top: 3px solid rgba(245, 158, 11, 0.38);
}

.teacher-kpi-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-kpi-value {
  margin-top: 0.45rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-kpi-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}
</style>
