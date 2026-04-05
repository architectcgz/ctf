<script setup lang="ts">
import { ArrowRight, FolderKanban } from 'lucide-vue-next'

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
  <div class="teacher-management-shell teacher-surface">
    <section class="teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <header class="teacher-header">
        <div class="teacher-header__main">
          <div class="teacher-surface-eyebrow journal-eyebrow">Class Directory</div>
          <h2 class="teacher-title">班级管理</h2>
          <p class="teacher-copy">查看当前可管理班级，并进入对应班级继续查看学生和训练表现。</p>

          <div class="teacher-actions">
            <button type="button" class="teacher-btn teacher-surface-btn" @click="emit('openDashboard')">
              教学概览
            </button>
            <button
              type="button"
              class="teacher-btn teacher-surface-btn teacher-btn--primary"
              @click="emit('openReportExport')"
            >
              导出报告
            </button>
          </div>
        </div>

        <div class="teacher-badge-grid">
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">班级数量</div>
            <div class="teacher-badge-value">{{ classes.length }}</div>
            <div class="teacher-badge-hint">当前可管理班级总数</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">学生总量</div>
            <div class="teacher-badge-value">
              {{ classes.reduce((sum, item) => sum + (item.student_count || 0), 0) }}
            </div>
            <div class="teacher-badge-hint">各班级学生数汇总</div>
          </article>
          <article class="teacher-badge-card teacher-surface-metric journal-brief journal-metric">
            <div class="teacher-badge-label">当前状态</div>
            <div class="teacher-badge-value">{{ loading ? '同步中' : '已就绪' }}</div>
            <div class="teacher-badge-hint">班级目录与入口操作已同步</div>
          </article>
        </div>
      </header>

      <div class="teacher-hero-divider" />

      <div class="teacher-tip-block">
        <div class="teacher-tip-label">当前焦点</div>
        <div class="teacher-tip-copy">先进入班级，再继续查看学生名单、训练趋势与学员分析。</div>
      </div>

      <div class="teacher-surface-board">
        <section class="teacher-surface-section">
          <div class="teacher-section-head">
            <div>
              <div class="teacher-surface-eyebrow journal-eyebrow">Class List</div>
              <h3 class="teacher-section-title">班级列表</h3>
            </div>
            <div class="teacher-section-meta">共 {{ classes.length }} 个班级</div>
          </div>

          <div v-if="loading" class="teacher-skeleton-list">
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

          <div v-else class="mt-5 teacher-table-shell">
            <ElTable
              :data="classes"
              row-key="name"
              class="teacher-surface-table teacher-class-table"
              empty-text="暂无班级"
            >
              <ElTableColumn prop="name" label="班级名称" min-width="260">
                <template #default="{ row }">
                  <div class="py-1">
                    <div class="teacher-class-name">{{ row.name }}</div>
                    <div class="teacher-class-copy">查看班级学生名单。</div>
                  </div>
                </template>
              </ElTableColumn>

              <ElTableColumn prop="student_count" label="学生数" width="140" align="center">
                <template #default="{ row }">
                  <span class="teacher-class-count">{{ row.student_count || 0 }}</span>
                </template>
              </ElTableColumn>

              <ElTableColumn label="操作" width="180" align="right">
                <template #default="{ row }">
                  <button
                    type="button"
                    class="teacher-row-btn"
                    @click="emit('openClass', row.name)"
                  >
                    进入班级
                    <ArrowRight class="h-4 w-4" />
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

.teacher-skeleton-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.75rem;
}

.teacher-table-shell {
  border: 1px solid var(--teacher-card-border);
  border-radius: 18px;
}

.teacher-class-name {
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 88%, var(--journal-muted));
}

.teacher-class-copy {
  margin-top: 0.3rem;
  font-size: 0.84rem;
  color: color-mix(in srgb, var(--journal-muted) 92%, transparent);
}

.teacher-class-count {
  font-size: 0.95rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-ink) 84%, var(--journal-muted));
}

.teacher-row-btn {
  display: inline-flex;
  align-items: center;
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

:deep(.teacher-class-table.el-table),
:deep(.teacher-class-table .el-table__inner-wrapper),
:deep(.teacher-class-table .el-scrollbar),
:deep(.teacher-class-table .el-scrollbar__view),
:deep(.teacher-class-table .el-table__body-wrapper),
:deep(.teacher-class-table .el-table__header-wrapper),
:deep(.teacher-class-table .el-table__empty-block) {
  background: var(--journal-surface);
}

@media (max-width: 960px) {
  .teacher-badge-grid {
    grid-template-columns: 1fr;
  }
}
</style>
