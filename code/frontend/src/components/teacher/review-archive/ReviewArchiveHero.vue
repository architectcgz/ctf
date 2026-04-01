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
  <section class="archive-hero">
    <div class="archive-hero__backdrop" />
    <div class="archive-hero__content">
      <div class="archive-hero__meta">
        <div>
          <div class="archive-hero__eyebrow">Teaching Review Archive</div>
          <h1 class="archive-hero__title">教学复盘归档</h1>
          <p class="archive-hero__description">
            将学生训练摘要、攻防证据、Writeup 与评阅记录收束为一份可讲解、可导出的课堂复盘视图。
          </p>
        </div>
        <div class="archive-hero__actions">
          <ElButton plain @click="emit('back')">返回学生列表</ElButton>
          <ElButton plain @click="emit('openAnalysis')">返回学员分析</ElButton>
          <ElButton type="primary" :loading="exporting" @click="emit('exportArchive')">
            导出复盘归档
          </ElButton>
        </div>
      </div>

      <div class="archive-hero__grid">
        <article class="archive-hero__profile">
          <div class="archive-hero__label">当前学员</div>
          <div class="archive-hero__student">{{ archive?.student.name || archive?.student.username || '--' }}</div>
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
            class="archive-hero__stat"
          >
            <div class="archive-hero__stat-label">{{ item.label }}</div>
            <div class="archive-hero__stat-value">{{ archive?.summary[item.field] ?? 0 }}</div>
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
  border: 1px solid color-mix(in srgb, #1e40af 18%, var(--color-border-default));
  border-radius: 28px;
  background:
    radial-gradient(circle at top right, rgba(245, 158, 11, 0.16), transparent 34%),
    linear-gradient(135deg, rgba(30, 64, 175, 0.08), rgba(248, 250, 252, 0.98));
}

.archive-hero__backdrop {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(59, 130, 246, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(59, 130, 246, 0.08) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.8), transparent);
}

.archive-hero__content {
  position: relative;
  z-index: 1;
  padding: 1.5rem;
}

.archive-hero__meta {
  display: flex;
  gap: 1rem;
  justify-content: space-between;
  align-items: flex-start;
}

.archive-hero__eyebrow {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: #1d4ed8;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.archive-hero__title {
  margin-top: 0.85rem;
  font-size: clamp(2rem, 4vw, 3rem);
  line-height: 1.05;
  color: #172554;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.archive-hero__description {
  max-width: 48rem;
  margin-top: 0.85rem;
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.archive-hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  justify-content: flex-end;
}

.archive-hero__grid {
  display: grid;
  gap: 1rem;
  margin-top: 1.5rem;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
}

.archive-hero__profile,
.archive-hero__stat {
  border: 1px solid color-mix(in srgb, #1e40af 14%, var(--color-border-default));
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.76);
  backdrop-filter: blur(10px);
}

.archive-hero__profile {
  padding: 1.1rem 1.15rem;
}

.archive-hero__label,
.archive-hero__stat-label {
  font-size: 0.75rem;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: #475569;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.archive-hero__student {
  margin-top: 0.8rem;
  font-size: 1.9rem;
  font-weight: 700;
  color: #0f172a;
}

.archive-hero__student-subline {
  display: flex;
  flex-wrap: wrap;
  gap: 0.85rem;
  margin-top: 0.4rem;
  color: #334155;
}

.archive-hero__stamp {
  display: inline-flex;
  flex-direction: column;
  gap: 0.3rem;
  margin-top: 1.1rem;
  padding: 0.8rem 0.9rem;
  border-radius: 18px;
  background: rgba(15, 23, 42, 0.04);
  color: #334155;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.82rem;
}

.archive-hero__stamp strong {
  color: #0f172a;
  font-size: 0.95rem;
}

.archive-hero__stats {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.archive-hero__stat {
  padding: 1rem;
}

.archive-hero__stat-value {
  margin-top: 0.7rem;
  font-size: 1.8rem;
  font-weight: 700;
  color: #172554;
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
    padding: 1.1rem;
  }

  .archive-hero__stats {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
