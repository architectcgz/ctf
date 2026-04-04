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
  <div class="teacher-management-shell teacher-surface space-y-6">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-surface-eyebrow">Teacher Instance Ops</div>
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

        <article class="teacher-brief teacher-surface-brief px-5 py-5">
          <div class="text-sm font-medium text-[var(--journal-ink)]">当前实例概况</div>
          <div class="teacher-metric-grid mt-5 grid gap-3 sm:grid-cols-3">
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--accent px-4 py-4"
            >
              <div class="teacher-metric-label">当前可见</div>
              <div class="teacher-metric-value">{{ totalCount }}</div>
              <div class="teacher-metric-hint">符合当前筛选条件的实例数量</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--calm px-4 py-4"
            >
              <div class="teacher-metric-label">运行中</div>
              <div class="teacher-metric-value">{{ runningCount }}</div>
              <div class="teacher-metric-hint">仍在占用环境资源的实例数量</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--soft px-4 py-4"
            >
              <div class="teacher-metric-label">即将到期</div>
              <div class="teacher-metric-value">{{ expiringSoonCount }}</div>
              <div class="teacher-metric-hint">剩余时间不足 10 分钟的实例数量</div>
            </article>
          </div>
        </article>
      </div>

      <div class="teacher-surface-board mt-6">
        <section class="teacher-surface-section teacher-surface-filter">
          <div>
            <div class="teacher-surface-eyebrow">Instance Filters</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">实例筛选</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              当前范围：{{ selectedClassLabel }}。支持按班级、用户名关键字、学号精确筛选。
            </p>
          </div>

          <form
            class="mt-5 grid gap-4 md:grid-cols-[220px_1fr_1fr]"
            @submit.prevent="emit('submit')"
          >
            <label class="space-y-2">
              <span class="text-sm text-text-secondary">班级</span>
              <select
                :value="className"
                class="teacher-filter-field w-full rounded-xl px-4 py-3 text-sm outline-none transition disabled:cursor-not-allowed disabled:opacity-60"
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
              <div class="teacher-filter-field teacher-filter-control px-4 py-3">
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
              <div class="teacher-filter-field teacher-filter-control px-4 py-3">
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
        </section>

        <section class="teacher-surface-section">
          <div>
            <div class="teacher-surface-eyebrow">Instance List</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">实例列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              保留当前筛选范围，并直接在列表中定位运行中、异常或即将到期的实例。
            </p>
          </div>

          <div v-if="loadingInstances" class="mt-5 space-y-3">
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

          <div v-else class="mt-5">
            <ElTable
              :data="instances"
              row-key="id"
              class="teacher-surface-table teacher-instance-table"
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
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
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
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 12px 24px var(--color-shadow-soft);
}

.teacher-btn--primary:hover {
  background: var(--journal-accent-strong);
  border-color: transparent;
  color: #fff;
}

.teacher-filter-field {
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  color: var(--journal-ink);
}

.teacher-filter-control {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition:
    border-color 0.18s ease,
    background 0.18s ease;
}

.teacher-filter-control:focus-within {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 5%, var(--journal-surface));
}

.teacher-metric-grid {
  align-items: stretch;
}

.teacher-metric-card {
  min-height: 100%;
  border-top: 3px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
}

.teacher-metric-card--accent {
  border-top-color: color-mix(in srgb, var(--journal-accent) 28%, var(--journal-border));
}

.teacher-metric-card--calm {
  border-top-color: color-mix(in srgb, var(--color-success) 22%, var(--journal-border));
}

.teacher-metric-card--soft {
  border-top-color: color-mix(in srgb, var(--color-warning) 22%, var(--journal-border));
}

.teacher-metric-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-metric-value {
  margin-top: 0.45rem;
  font-size: 1.18rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-metric-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}
</style>
