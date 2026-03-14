<script setup lang="ts">
import { computed } from 'vue'
import { Search, Trash2 } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherInstanceItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
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
        <div v-for="index in 6" :key="index" class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]" />
      </div>

      <AppEmpty
        v-else-if="instances.length === 0"
        icon="Inbox"
        title="当前没有匹配到实例"
        description="可以调整筛选条件，或等待学员创建新的训练环境后再查看。"
      />

      <div v-else>
        <ElTable
          :data="instances"
          row-key="id"
          class="teacher-instance-table"
          empty-text="没有匹配实例"
        >
          <ElTableColumn label="学生 / 班级" min-width="220">
            <template #default="{ row }">
              <div class="py-1">
                <div class="font-semibold text-text-primary">{{ row.student_name || row.student_username }}</div>
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
                  <span class="rounded-full border px-3 py-1 text-xs font-semibold" :class="statusMeta(row.status).chipClass">
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
              <span class="text-sm text-text-secondary">{{ formatDateTime(row.expires_at) }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="剩余时间" width="130" align="center">
            <template #default="{ row }">
              <span class="font-mono text-sm font-medium text-text-primary">{{ formatRemainingTime(row.remaining_time) }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="延期" width="120" align="center">
            <template #default="{ row }">
              <span class="text-sm font-medium text-text-primary">{{ remainingExtends(row) }} / {{ row.max_extends }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="创建时间" width="180">
            <template #default="{ row }">
              <span class="text-sm text-text-secondary">{{ formatDateTime(row.created_at) }}</span>
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

:deep(.teacher-instance-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-expanded-cell-bg-color: transparent;
  --el-table-header-bg-color: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --el-table-border-color: var(--color-border-default);
  --el-table-row-hover-bg-color: color-mix(
    in srgb,
    var(--color-primary) 8%,
    var(--color-bg-surface)
  );
  --el-table-text-color: var(--color-text-primary);
  --el-table-header-text-color: var(--color-text-secondary);
}

:deep(.teacher-instance-table th.el-table__cell) {
  background: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

:deep(.teacher-instance-table td.el-table__cell),
:deep(.teacher-instance-table th.el-table__cell) {
  border-bottom-color: var(--color-border-default);
}

:deep(.teacher-instance-table .el-table__inner-wrapper::before) {
  display: none;
}
</style>
