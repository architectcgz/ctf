<script setup lang="ts">
import { computed } from 'vue'
import { RefreshCcw, Search, Trash2 } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherInstanceItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppCard from '@/components/common/AppCard.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

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
      return { label: '运行中', chipClass: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-300' }
    case 'creating':
      return { label: '创建中', chipClass: 'border-cyan-500/25 bg-cyan-500/10 text-cyan-300' }
    case 'expired':
      return { label: '已过期', chipClass: 'border-amber-500/25 bg-amber-500/10 text-amber-300' }
    case 'failed':
      return { label: '异常', chipClass: 'border-rose-500/25 bg-rose-500/10 text-rose-300' }
    default:
      return { label: status, chipClass: 'border-border bg-elevated/70 text-text-secondary' }
  }
}

function remainingExtends(item: TeacherInstanceItem): number {
  return Math.max(0, item.max_extends - item.extend_count)
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Teacher Instance Ops"
      title="实例管理"
      description="聚焦教师/管理员对学生训练实例的查看与处置。先筛班级与学员，再快速定位长时间占用、异常或即将到期的实例。"
    >
      <button
        type="button"
        class="rounded-2xl border border-border bg-elevated px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary/40"
        @click="emit('openDashboard')"
      >
        返回教学概览
      </button>
    </PageHeader>

    <section class="grid gap-4 md:grid-cols-3">
      <MetricCard
        label="当前可见"
        :value="totalCount"
        hint="符合当前筛选条件的实例数量"
        accent="primary"
      />
      <MetricCard
        label="运行中"
        :value="runningCount"
        hint="仍在占用环境资源的实例数量"
        accent="success"
      />
      <MetricCard
        label="即将到期"
        :value="expiringSoonCount"
        hint="剩余时间不足 10 分钟的实例数量"
        accent="warning"
      />
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <SectionCard title="实例筛选" :subtitle="`当前范围：${selectedClassLabel}。支持按班级、用户名关键字、学号精确筛选。`">
      <form class="grid gap-4 md:grid-cols-[220px_1fr_1fr]" @submit.prevent="emit('submit')">
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
          <div class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3">
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
          <div class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3">
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
          <button
            type="button"
            class="rounded-xl border border-border bg-surface px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary/40"
            @click="emit('reset')"
          >
            重置筛选
          </button>
          <button
            type="submit"
            class="rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-white transition hover:opacity-90"
          >
            查询实例
          </button>
        </div>
      </form>
    </SectionCard>

    <SectionCard title="实例列表" subtitle="按创建时间倒序展示，便于直接处理最近产生的问题实例。">
      <div v-if="loadingInstances" class="space-y-3">
        <div v-for="index in 4" :key="index" class="h-36 animate-pulse rounded-[24px] bg-[var(--color-bg-base)]" />
      </div>

      <AppEmpty
        v-else-if="instances.length === 0"
        icon="Inbox"
        title="当前没有匹配到实例"
        description="可以调整筛选条件，或等待学员创建新的训练环境后再查看。"
      />

      <div v-else class="grid gap-4">
        <AppCard
          v-for="item in instances"
          :key="item.id"
          variant="action"
          accent="neutral"
          class="overflow-visible"
        >
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="space-y-4">
              <div class="flex flex-wrap items-center gap-3">
                <span class="rounded-full border px-3 py-1 text-xs font-semibold" :class="statusMeta(item.status).chipClass">
                  {{ statusMeta(item.status).label }}
                </span>
                <span class="text-xs font-medium uppercase tracking-[0.18em] text-text-muted">
                  {{ item.class_name }}
                </span>
              </div>

              <div class="space-y-1">
                <h3 class="text-lg font-semibold text-text-primary">{{ item.challenge_title }}</h3>
                <p class="text-sm text-text-secondary">
                  {{ item.student_name || item.student_username }}
                  <span class="mx-1 text-text-muted">·</span>
                  @{{ item.student_username }}
                  <span v-if="item.student_no" class="mx-1 text-text-muted">·</span>
                  <span v-if="item.student_no">学号 {{ item.student_no }}</span>
                </p>
              </div>

              <div class="grid gap-3 text-sm text-text-secondary md:grid-cols-2 xl:grid-cols-4">
                <div class="rounded-2xl border border-border-subtle bg-elevated/60 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.18em] text-text-muted">访问地址</div>
                  <div class="mt-2 break-all font-mono text-text-primary">{{ item.access_url || '暂未分配' }}</div>
                </div>
                <div class="rounded-2xl border border-border-subtle bg-elevated/60 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.18em] text-text-muted">到期时间</div>
                  <div class="mt-2 font-medium text-text-primary">{{ formatDateTime(item.expires_at) }}</div>
                </div>
                <div class="rounded-2xl border border-border-subtle bg-elevated/60 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.18em] text-text-muted">剩余时间</div>
                  <div class="mt-2 font-mono text-text-primary">{{ formatRemainingTime(item.remaining_time) }}</div>
                </div>
                <div class="rounded-2xl border border-border-subtle bg-elevated/60 px-4 py-3">
                  <div class="text-[11px] uppercase tracking-[0.18em] text-text-muted">剩余延期</div>
                  <div class="mt-2 font-medium text-text-primary">{{ remainingExtends(item) }} / {{ item.max_extends }}</div>
                </div>
              </div>
            </div>

            <div class="flex shrink-0 items-center gap-3 lg:flex-col lg:items-end">
              <button
                type="button"
                class="inline-flex items-center gap-2 rounded-2xl border border-red-500/20 bg-red-500/10 px-4 py-2.5 text-sm font-medium text-red-300 transition hover:bg-red-500/15 disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="destroyingId === item.id"
                :data-instance-id="item.id"
                @click="emit('destroy', item.id)"
              >
                <Trash2 class="h-4 w-4" />
                {{ destroyingId === item.id ? '销毁中...' : '销毁实例' }}
              </button>

              <div class="inline-flex items-center gap-2 text-xs text-text-muted">
                <RefreshCcw class="h-3.5 w-3.5" />
                创建于 {{ formatDateTime(item.created_at) }}
              </div>
            </div>
          </div>
        </AppCard>
      </div>
    </SectionCard>
  </div>
</template>

<style scoped>
:deep(.teacher-filter-field) {
  color: var(--color-text-primary);
}

:deep(.teacher-filter-field option) {
  background-color: var(--color-bg-surface);
  color: var(--color-text-primary);
}

:deep(.teacher-filter-field select),
:deep(.teacher-filter-field input) {
  color: var(--color-text-primary);
}
</style>
