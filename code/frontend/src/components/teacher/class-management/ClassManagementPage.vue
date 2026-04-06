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
  <div class="teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <section
      class="teacher-hero teacher-surface-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
    >
      <div class="teacher-page">
        <header class="teacher-topbar">
          <div class="teacher-heading">
            <div class="teacher-surface-eyebrow journal-eyebrow">Class Directory</div>
            <h2 class="teacher-title">班级管理</h2>
            <p class="teacher-copy">查看当前可管理班级，并进入对应班级继续查看学生和训练表现。</p>
          </div>

          <div class="teacher-actions">
            <button type="button" class="teacher-btn teacher-btn--primary" @click="emit('openDashboard')">
              教学概览
            </button>
            <button type="button" class="teacher-btn teacher-btn--ghost" @click="emit('openReportExport')">
              导出报告
            </button>
          </div>
        </header>

        <section class="teacher-summary">
          <div class="teacher-summary-title">
            <FolderKanban class="h-4 w-4" />
            <span>Directory Snapshot</span>
          </div>
          <div class="teacher-summary-grid">
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">班级数量</div>
              <div class="teacher-summary-value">{{ classes.length }}</div>
              <div class="teacher-summary-helper">当前可管理班级总数</div>
            </div>
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">学生总量</div>
              <div class="teacher-summary-value">
                {{ classes.reduce((sum, item) => sum + (item.student_count || 0), 0) }}
              </div>
              <div class="teacher-summary-helper">各班级学生数汇总</div>
            </div>
            <div class="teacher-summary-item">
              <div class="teacher-summary-label">当前状态</div>
              <div class="teacher-summary-value">{{ loading ? '同步中' : '已就绪' }}</div>
              <div class="teacher-summary-helper">班级目录与入口操作已同步</div>
            </div>
          </div>
        </section>

        <div class="teacher-divider" />

        <div v-if="loading" class="teacher-skeleton-list">
          <div
            v-for="index in 5"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[var(--journal-surface-subtle)]"
          />
        </div>

        <AppEmpty
          v-else-if="classes.length === 0"
          class="teacher-empty-state"
          icon="Users"
          title="暂无班级"
          description="当前教师账号下还没有可访问的班级。"
        />

        <section v-else class="teacher-directory" aria-label="班级目录">
          <div class="teacher-directory-top">
            <h3 class="teacher-directory-title">班级目录</h3>
            <div class="teacher-directory-meta">共 {{ classes.length }} 个班级</div>
          </div>

          <div class="teacher-directory-head">
            <span>班级</span>
            <span>标签</span>
            <span>状态</span>
            <span>数据</span>
            <span>操作</span>
          </div>

          <button
            v-for="(item, index) in classes"
            :key="item.name"
            type="button"
            class="teacher-directory-row"
            :aria-label="`${item.name}，${item.student_count || 0} 名学生，进入班级`"
            @click="emit('openClass', item.name)"
          >
            <div class="teacher-directory-row-main">
              <div class="teacher-directory-row-index">CL-{{ String(index + 1).padStart(2, '0') }}</div>
              <div class="teacher-directory-row-title-group">
                <h4 class="teacher-directory-row-title">{{ item.name }}</h4>
                <div class="teacher-directory-row-points">{{ item.student_count || 0 }} Students</div>
              </div>
              <div class="teacher-directory-row-copy">查看班级学生名单与训练表现。</div>
            </div>

            <div class="teacher-directory-row-tags">
              <span class="teacher-directory-chip">Teaching Class</span>
              <span class="teacher-directory-chip teacher-directory-chip-muted">
                {{
                  (item.student_count || 0) >= 40
                    ? 'Large'
                    : (item.student_count || 0) >= 20
                      ? 'Standard'
                      : 'Compact'
                }}
              </span>
            </div>

            <div class="teacher-directory-row-status">
              <span
                class="teacher-directory-state-chip"
                :class="
                  (item.student_count || 0) > 0
                    ? 'teacher-directory-state-chip-ready'
                    : 'teacher-directory-state-chip-empty'
                "
              >
                {{ (item.student_count || 0) > 0 ? '可查看' : '待入班' }}
              </span>
            </div>

            <div class="teacher-directory-row-metrics">
              <span>{{ item.student_count || 0 }} 名学生</span>
              <span>{{ (item.student_count || 0) > 0 ? '可继续查看训练趋势' : '当前还没有学生加入' }}</span>
            </div>

            <div class="teacher-directory-row-cta">
              <span>进入班级</span>
              <ArrowRight class="h-4 w-4" />
            </div>
          </button>
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
  margin-top: 0.8rem;
  font-size: clamp(2rem, 4vw, 2.85rem);
  font-weight: 700;
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

.teacher-divider {
  margin-top: 1.5rem;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.teacher-skeleton-list {
  margin-top: 1.5rem;
  display: grid;
  gap: 0.75rem;
}

.teacher-empty-state {
  margin-top: 1.5rem;
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
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
  grid-template-columns: minmax(0, 1.35fr) minmax(220px, 0.85fr) 7rem 10rem 7rem;
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
  grid-template-columns: minmax(0, 1.35fr) minmax(220px, 0.85fr) 7rem 10rem 7rem;
  gap: 1rem;
  align-items: center;
  width: 100%;
  padding: 1.1rem 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.teacher-directory-row:hover,
.teacher-directory-row:focus-visible {
  background: color-mix(in srgb, var(--journal-accent) 5%, transparent);
  box-shadow: inset 2px 0 0 color-mix(in srgb, var(--journal-accent) 62%, transparent);
  outline: none;
}

.teacher-directory-row-main {
  display: grid;
  gap: 0.5rem;
  min-width: 0;
}

.teacher-directory-row-index,
.teacher-directory-row-points {
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

.teacher-directory-state-chip-ready {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

.teacher-directory-state-chip-empty {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.teacher-directory-row-metrics {
  display: grid;
  gap: 0.25rem;
  font-size: 0.81rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.teacher-directory-row-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.38rem;
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--journal-accent-strong);
}

@media (max-width: 960px) {
  .teacher-topbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .teacher-summary-grid {
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
