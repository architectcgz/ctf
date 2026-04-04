<script setup lang="ts">
import { ChevronRight } from 'lucide-vue-next'

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
  <div class="teacher-management-shell teacher-surface space-y-6">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-surface-eyebrow">Class Directory</div>
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

        <article class="teacher-brief teacher-surface-brief px-5 py-5">
          <div class="text-sm font-medium text-[var(--journal-ink)]">当前班级概况</div>
          <div class="teacher-metric-grid mt-5 grid gap-3 sm:grid-cols-2">
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--accent px-4 py-4"
            >
              <div class="teacher-metric-label">班级数量</div>
              <div class="teacher-metric-value">{{ classes.length }}</div>
              <div class="teacher-metric-hint">当前可管理班级总数</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--calm px-4 py-4"
            >
              <div class="teacher-metric-label">学生总量</div>
              <div class="teacher-metric-value">
                {{ classes.reduce((sum, item) => sum + (item.student_count || 0), 0) }}
              </div>
              <div class="teacher-metric-hint">各班级学生数汇总</div>
            </article>
          </div>
        </article>
      </div>

      <div class="teacher-surface-board mt-6">
        <section class="teacher-surface-section">
          <div>
            <div class="teacher-surface-eyebrow">Class List</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">班级列表</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              查看当前可管理的班级，并进入对应班级继续查看学生。
            </p>
          </div>

          <div v-if="loading" class="mt-5 space-y-3">
            <div
              v-for="index in 5"
              :key="index"
              class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
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
            <ElTable
              :data="classes"
              row-key="name"
              class="teacher-surface-table teacher-class-table"
              empty-text="暂无班级"
            >
              <ElTableColumn prop="name" label="班级名称" min-width="260">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="teacher-class-name font-semibold">{{ row.name }}</div>
                    <div class="teacher-class-copy mt-1 text-sm">查看班级学生名单。</div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn prop="student_count" label="学生数" width="140" align="center">
                <template #default="{ row }">
                  <span class="teacher-class-count text-sm font-medium">{{
                    row.student_count || 0
                  }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="操作" width="160" align="right">
                <template #default="{ row }">
                  <ElButton class="teacher-class-action" @click="emit('openClass', row.name)">
                    进入班级
                    <ChevronRight class="ml-1 h-4 w-4" />
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
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
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
  border-color: transparent;
  background: var(--journal-accent-strong);
  color: #fff;
}

.teacher-metric-grid {
  align-items: stretch;
}

.teacher-metric-card {
  min-height: 100%;
  border-top: 3px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
}

.teacher-metric-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.teacher-metric-card--accent {
  border-top-color: color-mix(in srgb, var(--journal-accent) 28%, var(--journal-border));
}

.teacher-metric-card--calm {
  border-top-color: color-mix(in srgb, var(--color-success) 22%, var(--journal-border));
}

.teacher-metric-value {
  margin-top: 0.45rem;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-metric-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.teacher-class-name {
  color: color-mix(in srgb, var(--journal-ink) 88%, var(--journal-muted));
}

.teacher-class-copy {
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-class-count {
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

:deep(.teacher-class-action.el-button) {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 24%, var(--journal-border));
  background: color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 78%, var(--journal-ink));
  border-radius: 0.9rem;
  box-shadow: none;
}

:deep(.teacher-class-action.el-button:hover),
:deep(.teacher-class-action.el-button:focus-visible) {
  border-color: color-mix(in srgb, var(--journal-accent) 38%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface));
  color: var(--journal-accent);
}
</style>
