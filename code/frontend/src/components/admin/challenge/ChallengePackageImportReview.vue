<script setup lang="ts">
import { computed } from 'vue'

import type { AdminChallengeImportPreview } from '@/api/contracts'

const props = defineProps<{
  preview: AdminChallengeImportPreview
  committing: boolean
}>()

const emit = defineEmits<{
  confirm: []
  reset: []
}>()

const metadata = computed(() => [
  { label: 'Slug', value: props.preview.slug },
  { label: '分类', value: props.preview.category },
  { label: '难度', value: props.preview.difficulty },
  { label: '分值', value: `${props.preview.points} pts` },
])
</script>

<template>
  <section class="import-review">
    <header class="import-review__header">
      <div>
        <div class="import-review__eyebrow">Preview</div>
        <h2 class="import-review__title">{{ preview.title }}</h2>
      </div>
      <div class="import-review__actions">
        <button class="import-review__ghost" type="button" @click="emit('reset')">重新选择</button>
        <button
          class="import-review__primary"
          type="button"
          :disabled="committing"
          @click="emit('confirm')"
        >
          {{ committing ? '导入中...' : '确认导入' }}
        </button>
      </div>
    </header>

    <div class="import-review__grid">
      <article class="import-review__section">
        <div class="import-review__section-title">题目概览</div>
        <div class="import-review__meta">
          <div v-for="item in metadata" :key="item.label" class="import-review__meta-item">
            <span class="import-review__meta-label">{{ item.label }}</span>
            <strong class="import-review__meta-value">{{ item.value }}</strong>
          </div>
        </div>
        <p class="import-review__description">{{ preview.description }}</p>
      </article>

      <article class="import-review__section">
        <div class="import-review__section-title">运行时与 Flag</div>
        <dl class="import-review__definition">
          <div>
            <dt>Flag</dt>
            <dd>{{ preview.flag.type }} / {{ preview.flag.prefix || 'flag' }}</dd>
          </div>
          <div>
            <dt>Runtime</dt>
            <dd>{{ preview.runtime.image_ref || '无镜像引用' }}</dd>
          </div>
          <div>
            <dt>Topology</dt>
            <dd>
              {{ preview.extensions.topology.source || '未声明' }}
              <span v-if="preview.extensions.topology.enabled"> / 已启用</span>
            </dd>
          </div>
        </dl>
      </article>
    </div>

    <div class="import-review__grid">
      <article class="import-review__section">
        <div class="import-review__section-title">附件</div>
        <div v-if="preview.attachments?.length" class="import-review__list">
          <div v-for="attachment in preview.attachments" :key="attachment.path" class="import-review__list-item">
            <strong>{{ attachment.name }}</strong>
            <span>{{ attachment.path }}</span>
          </div>
        </div>
        <div v-else class="import-review__empty">当前题目包未包含附件。</div>
      </article>

      <article class="import-review__section">
        <div class="import-review__section-title">提示系统</div>
        <div v-if="preview.hints?.length" class="import-review__list">
          <div v-for="hint in preview.hints" :key="hint.level" class="import-review__list-item">
            <strong>Level {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}</strong>
            <span>{{ hint.content }}</span>
          </div>
        </div>
        <div v-else class="import-review__empty">当前题目包未声明提示。</div>
      </article>
    </div>

    <article v-if="preview.warnings?.length" class="import-review__warning">
      <div class="import-review__section-title">导入提醒</div>
      <ul class="import-review__warnings">
        <li v-for="warning in preview.warnings" :key="warning">{{ warning }}</li>
      </ul>
    </article>
  </section>
</template>

<style scoped>
.import-review {
  display: grid;
  gap: 1rem;
  padding-block: 0.5rem 1rem;
}

.import-review__header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.import-review__eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.import-review__title {
  margin: 0.3rem 0 0;
  font-size: clamp(1.35rem, 1.8vw, 1.7rem);
  font-weight: 700;
  color: var(--journal-ink);
}

.import-review__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.import-review__primary,
.import-review__ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.75rem;
  padding: 0.65rem 1rem;
  border-radius: 999px;
  font-size: 0.9rem;
  font-weight: 700;
  transition: all 150ms ease;
}

.import-review__primary {
  border: 1px solid rgba(37, 99, 235, 0.18);
  background: var(--journal-accent);
  color: #fff;
}

.import-review__primary:disabled {
  opacity: 0.6;
  cursor: progress;
}

.import-review__ghost {
  border: 1px solid var(--journal-border);
  background: rgba(255, 255, 255, 0.78);
  color: var(--journal-ink);
}

.import-review__grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.import-review__section,
.import-review__warning {
  display: grid;
  gap: 0.85rem;
  padding: 1rem 0;
  border-top: 1px solid rgba(148, 163, 184, 0.26);
}

.import-review__section-title {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.import-review__meta {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.import-review__meta-item {
  display: grid;
  gap: 0.2rem;
}

.import-review__meta-label {
  font-size: 0.78rem;
  color: var(--journal-muted);
}

.import-review__meta-value {
  font-size: 0.96rem;
  color: var(--journal-ink);
}

.import-review__description {
  margin: 0;
  color: var(--journal-muted);
  line-height: 1.75;
  white-space: pre-wrap;
}

.import-review__definition {
  display: grid;
  gap: 0.85rem;
  margin: 0;
}

.import-review__definition dt {
  font-size: 0.78rem;
  color: var(--journal-muted);
}

.import-review__definition dd {
  margin: 0.22rem 0 0;
  color: var(--journal-ink);
  line-height: 1.65;
}

.import-review__list {
  display: grid;
  gap: 0.65rem;
}

.import-review__list-item {
  display: grid;
  gap: 0.2rem;
  padding-bottom: 0.65rem;
  border-bottom: 1px dashed rgba(148, 163, 184, 0.4);
}

.import-review__list-item strong {
  color: var(--journal-ink);
  font-size: 0.9rem;
}

.import-review__list-item span,
.import-review__empty {
  color: var(--journal-muted);
  font-size: 0.88rem;
  line-height: 1.7;
}

.import-review__warnings {
  margin: 0;
  padding-left: 1rem;
  display: grid;
  gap: 0.45rem;
  color: #92400e;
  font-size: 0.88rem;
  line-height: 1.7;
}

@media (max-width: 960px) {
  .import-review__grid,
  .import-review__meta {
    grid-template-columns: 1fr;
  }
}
</style>
