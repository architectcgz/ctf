<script setup lang="ts">
import type { ReviewArchiveData } from '@/api/contracts'

defineProps<{
  archive: ReviewArchiveData | null
  exporting: boolean
}>()

const emit = defineEmits<{
  back: []
  openAnalysis: []
  exportArchive: []
}>()

const statItems = [
  { key: 'solved', label: '完成题目', field: 'total_solved' },
  { key: 'attempts', label: '总提交', field: 'total_attempts' },
  { key: 'evidence', label: '证据事件', field: 'evidence_event_count' },
  { key: 'writeups', label: '复盘材料', field: 'writeup_count' },
] as const
</script>

<template>
  <section class="archive-hero teacher-surface-hero">
    <div class="archive-hero__backdrop" />
    <div class="archive-hero__content">
      <div class="archive-hero__meta">
        <div>
          <div class="archive-hero__eyebrow">
            Teaching Review Archive
          </div>
          <h1 class="archive-hero__title workspace-page-title">
            教学复盘归档
          </h1>
          <p class="archive-hero__description workspace-page-copy">
            将学生训练摘要、攻防证据、Writeup 与评阅记录收束为一份可讲解、可导出的课堂复盘视图。
          </p>
        </div>
        <div class="archive-hero__actions">
          <button
            type="button"
            class="ui-btn ui-btn--secondary"
            @click="emit('back')"
          >
            返回学生列表
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--secondary"
            @click="emit('openAnalysis')"
          >
            返回学员分析
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--primary"
            :disabled="exporting"
            @click="emit('exportArchive')"
          >
            {{ exporting ? '导出中...' : '导出复盘归档' }}
          </button>
        </div>
      </div>

      <div class="archive-hero__grid">
        <article class="archive-hero__profile">
          <div class="archive-hero__label">
            当前学员
          </div>
          <div class="archive-hero__student">
            {{ archive?.student.name || archive?.student.username || '--' }}
          </div>
          <div class="archive-hero__student-subline">
            <span>@{{ archive?.student.username || '--' }}</span>
            <span>{{ archive?.student.class_name || '--' }}</span>
          </div>
          <div class="archive-hero__stamp">
            <span>last activity</span>
            <strong>{{ archive?.summary.last_activity_at || '--' }}</strong>
          </div>
        </article>

        <div class="archive-hero__stats">
          <article
            v-for="item in statItems"
            :key="item.key"
            class="archive-hero__stat teacher-surface-metric"
          >
            <div class="archive-hero__stat-label">
              {{ item.label }}
            </div>
            <div class="archive-hero__stat-value">
              {{ archive?.summary[item.field] ?? 0 }}
            </div>
          </article>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.archive-hero {
  position: relative;
  overflow: hidden;
  border-radius: 28px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 12%, transparent),
      transparent 34%
    ),
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--journal-accent) 10%, var(--journal-surface)),
      color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base))
    );
}

.archive-hero__backdrop {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(color-mix(in srgb, var(--journal-accent) 8%, transparent) 1px, transparent 1px),
    linear-gradient(
      90deg,
      color-mix(in srgb, var(--journal-accent) 8%, transparent) 1px,
      transparent 1px
    );
  background-size: 28px 28px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.8), transparent);
}

.archive-hero__content {
  position: relative;
  z-index: 1;
  padding: var(--space-6);
}

.archive-hero__meta {
  display: flex;
  gap: var(--space-4);
  justify-content: space-between;
  align-items: flex-start;
}

.archive-hero__eyebrow {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
  font-family: var(--font-family-sans);
}

.archive-hero__title {
  margin-top: var(--space-3);
  color: var(--journal-ink);
  font-family: var(--font-family-sans);
}

.archive-hero__description {
  max-width: 48rem;
  margin-top: var(--space-3);
  color: var(--color-text-secondary);
}

.archive-hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  justify-content: flex-end;
}

.archive-hero__grid {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-6);
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
}

.archive-hero__profile,
.archive-hero__stat {
  border-radius: 22px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
}

.archive-hero__profile {
  padding: var(--space-4-5) var(--space-5);
}

.archive-hero__label,
.archive-hero__stat-label {
  font-size: var(--font-size-0-75);
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.archive-hero__student {
  margin-top: var(--space-3);
  font-size: var(--font-size-1-90);
  font-weight: 700;
  color: var(--journal-ink);
}

.archive-hero__student-subline {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-1-5);
  color: color-mix(in srgb, var(--journal-muted) 82%, var(--journal-ink));
}

.archive-hero__stamp {
  display: inline-flex;
  flex-direction: column;
  gap: var(--space-1);
  margin-top: var(--space-4-5);
  padding: var(--space-3) var(--space-3-5);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-accent) 6%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-muted) 82%, var(--journal-ink));
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-82);
}

.archive-hero__stamp strong {
  color: var(--journal-ink);
  font-size: var(--font-size-0-95);
}

.archive-hero__stats {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.archive-hero__stat {
  padding: var(--space-4);
}

.archive-hero__stat-value {
  margin-top: var(--space-3);
  font-size: var(--font-size-1-80);
  font-weight: 700;
  color: var(--journal-ink);
}

@media (max-width: 1023px) {
  .archive-hero__meta,
  .archive-hero__grid {
    grid-template-columns: 1fr;
    flex-direction: column;
  }

  .archive-hero__actions {
    justify-content: flex-start;
  }
}

@media (max-width: 767px) {
  .archive-hero {
    border-radius: 24px;
  }

  .archive-hero__content {
    padding: var(--space-4-5);
  }

  .archive-hero__stats {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
