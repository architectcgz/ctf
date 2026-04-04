<script setup lang="ts">
import { computed } from 'vue'
import { ChevronRight, Users } from 'lucide-vue-next'

import type { TeacherClassItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openDashboard: []
  openReportExport: []
  openClass: [className: string]
}>()

const totalStudents = computed(() =>
  props.classes.reduce((sum, item) => sum + (item.student_count || 0), 0)
)

const averageStudentsText = computed(() => {
  if (props.classes.length === 0) return '--'
  return (totalStudents.value / props.classes.length).toFixed(1)
})

const largestClassText = computed(() => {
  if (props.classes.length === 0) return '--'
  const largestClass = props.classes.reduce((largest, item) =>
    (item.student_count || 0) > (largest.student_count || 0) ? item : largest
  )
  return `${largestClass.name} · ${largestClass.student_count || 0} 人`
})

const overviewBadges = computed(() => [
  { key: 'class-count', label: '班级池', value: `${props.classes.length} 个班级` },
  { key: 'student-total', label: '学生总量', value: `${totalStudents.value} 人` },
  { key: 'largest-class', label: '最大班级', value: largestClassText.value },
])

const classTips = computed(() => {
  if (props.classes.length === 0) {
    return [
      '当前教师账号下还没有可访问班级。',
      '可先确认班级分配与教师权限状态。',
      '接入班级后可在此继续查看学生名单与训练表现。',
    ]
  }

  return [
    `当前已接入 ${props.classes.length} 个班级，优先关注学生数较多的班级。`,
    `班级学生总量为 ${totalStudents.value} 人，可结合报告导出做阶段复盘。`,
    '进入班级后可继续查看学生名单、训练表现与后续教学动作。',
  ]
})

const overviewMetrics = computed(() => [
  { key: 'class-count', label: '班级数量', value: props.classes.length, hint: '当前可管理班级总数' },
  { key: 'student-total', label: '学生总量', value: totalStudents.value, hint: '各班级学生数汇总' },
  { key: 'average-class', label: '平均班级人数', value: averageStudentsText.value, hint: '按当前班级池估算的平均规模' },
  { key: 'largest-class', label: '最大班级', value: largestClassText.value, hint: '当前学生数最多的班级' },
])
</script>

<template>
  <div class="teacher-management-shell teacher-surface space-y-6">
    <section class="teacher-hero teacher-surface-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-eyebrow-row">
            <div class="journal-eyebrow">Class Directory</div>
            <span class="teacher-class-chip">{{ props.classes.length }} 个班级</span>
          </div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            班级管理
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            统一查看当前可管理班级，并进入对应班级继续查看学生和训练表现。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="teacher-btn teacher-surface-btn" @click="emit('openDashboard')">
              教学概览
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--primary teacher-surface-btn teacher-surface-btn--primary"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>
        </div>

        <article class="teacher-brief teacher-surface-brief journal-brief rounded-[24px] border px-5 py-5">
          <div class="teacher-brief-title">当前班级概况</div>
          <div class="teacher-badge-grid mt-5">
            <div v-for="badge in overviewBadges" :key="badge.key" class="teacher-badge-card">
              <div class="teacher-badge-label">{{ badge.label }}</div>
              <div class="teacher-badge-value">{{ badge.value }}</div>
            </div>
          </div>

          <div class="teacher-tip-block mt-5">
            <div class="teacher-tip-title">今日教学建议</div>
            <ul class="teacher-tip-list mt-3">
              <li v-for="(tip, index) in classTips" :key="tip" class="teacher-tip-item">
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
            <div class="journal-eyebrow teacher-eyebrow--soft">Class List</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">班级列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              查看当前可管理的班级，并进入对应班级继续查看学生。
            </p>
          </div>

          <div v-if="loading" class="mt-5 space-y-3">
            <div
              v-for="index in 5"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]"
            />
          </div>

          <AppEmpty
            v-else-if="classes.length === 0"
            class="teacher-surface-empty mt-5"
            icon="Users"
            title="暂无班级"
            description="当前教师账号下还没有可访问的班级。"
          />

          <div v-else class="mt-5">
            <ElTable
              :data="classes"
              row-key="name"
              class="teacher-class-table teacher-surface-table"
              empty-text="暂无班级"
            >
              <ElTableColumn prop="name" label="班级名称" min-width="260">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="font-semibold text-text-primary">{{ row.name }}</div>
                    <div class="mt-1 text-sm text-text-secondary">查看班级学生名单。</div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn prop="student_count" label="学生数" width="140" align="center">
                <template #default="{ row }">
                  <span class="text-sm font-medium text-text-primary">{{
                    row.student_count || 0
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="操作" width="160" align="right">
                <template #default="{ row }">
                  <button
                    type="button"
                    class="teacher-btn teacher-surface-btn"
                    @click="emit('openClass', row.name)"
                  >
                    进入班级
                    <ChevronRight class="ml-1 h-4 w-4" />
                  </button>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        </section>

        <div
          v-if="error"
          class="teacher-error-card teacher-surface-error"
          role="alert"
          aria-live="polite"
        >
          {{ error }}
          <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
:deep(.teacher-class-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-expanded-cell-bg-color: transparent;
  --el-table-header-bg-color: var(--journal-surface);
  --el-table-border-color: var(--journal-border);
  --el-table-row-hover-bg-color: rgba(99, 102, 241, 0.06);
  --el-table-text-color: var(--journal-ink);
  --el-table-header-text-color: var(--journal-muted);
}

:deep(.teacher-class-table th.el-table__cell) {
  background: var(--journal-surface);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

:deep(.teacher-class-table td.el-table__cell),
:deep(.teacher-class-table th.el-table__cell) {
  border-bottom-color: var(--journal-border);
}

:deep(.teacher-class-table.el-table),
:deep(.teacher-class-table .el-table__inner-wrapper),
:deep(.teacher-class-table .el-table__body-wrapper),
:deep(.teacher-class-table .el-table__header-wrapper),
:deep(.teacher-class-table .el-table__empty-block) {
  background: var(--journal-surface);
}

:deep(.teacher-class-table .el-table__inner-wrapper::before) {
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

.teacher-eyebrow--soft {
  opacity: 0.88;
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

.teacher-kpi-grid {
  align-items: stretch;
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

:deep(.teacher-class-table) {
  --el-table-row-hover-bg-color: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
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
