<script setup lang="ts">
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
</script>

<template>
  <div class="teacher-management-shell space-y-6">
    <section class="teacher-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-eyebrow">Class Directory</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            班级管理
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            统一查看当前可管理班级，并进入对应班级继续查看学生和训练表现。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="teacher-btn" @click="emit('openDashboard')">
              教学概览
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>
        </div>

        <article class="teacher-brief rounded-[24px] border px-5 py-5">
          <div class="teacher-brief-title">当前班级概况</div>
          <div class="teacher-kpi-grid mt-5 grid gap-3 sm:grid-cols-2">
            <article class="teacher-kpi-card teacher-kpi-card--primary">
              <div class="teacher-kpi-label">班级数量</div>
              <div class="teacher-kpi-value">{{ classes.length }}</div>
              <div class="teacher-kpi-hint">当前可管理班级总数</div>
            </article>
            <article class="teacher-kpi-card teacher-kpi-card--success">
              <div class="teacher-kpi-label">学生总量</div>
              <div class="teacher-kpi-value">
                {{ classes.reduce((sum, item) => sum + (item.student_count || 0), 0) }}
              </div>
              <div class="teacher-kpi-hint">各班级学生数汇总</div>
            </article>
          </div>
        </article>
      </div>

      <div class="teacher-hero-divider" />

      <div class="teacher-hero-section">
        <div class="teacher-hero-section-head">
          <div>
            <div class="teacher-eyebrow teacher-eyebrow--soft">Class List</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">班级列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              查看当前可管理的班级，并进入对应班级继续查看学生。
            </p>
          </div>
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
          class="mt-5"
          icon="Users"
          title="暂无班级"
          description="当前教师账号下还没有可访问的班级。"
        />

        <div v-else class="mt-5">
          <ElTable :data="classes" row-key="name" class="teacher-class-table" empty-text="暂无班级">
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
                <ElButton type="primary" plain @click="emit('openClass', row.name)">
                  进入班级
                  <ChevronRight class="ml-1 h-4 w-4" />
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

:deep(.teacher-class-table .el-table__inner-wrapper::before) {
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

.teacher-hero-section {
  margin-top: 1.5rem;
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
