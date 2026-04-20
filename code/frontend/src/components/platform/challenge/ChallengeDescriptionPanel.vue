<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'

import { useSanitize } from '@/composables/useSanitize'

const props = withDefaults(
  defineProps<{
    content?: string
    label?: string
    testId?: string
  }>(),
  {
    label: '题头',
    testId: 'import-review-description',
  }
)

const { sanitizeHtml } = useSanitize()

const renderedContent = computed(() => {
  if (!props.content) return ''
  const html = marked.parse(props.content, {
    gfm: true,
    breaks: true,
  })
  return sanitizeHtml(typeof html === 'string' ? html : props.content)
})
</script>

<template>
  <span class="import-review__statement-label">{{ label }}</span>
  <div class="import-review__statement">
    <!-- eslint-disable-next-line vue/no-v-html -->
    <div
      class="import-review__description"
      :data-testid="testId"
      v-html="renderedContent"
    />
  </div>
</template>

<style scoped>
.import-review__statement {
  display: grid;
  gap: var(--space-2);
  max-height: clamp(15rem, 36vh, 24rem);
  overflow: auto;
  border: 1px solid
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  border-radius: 0.75rem;
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 95%,
    var(--color-bg-base)
  );
  scrollbar-gutter: stable;
}

.import-review__statement-label {
  display: inline-flex;
  margin-top: var(--space-1);
  font-size: var(--font-size-0-78);
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--journal-muted);
}

.import-review__description {
  padding: 0 var(--space-3-5) var(--space-3);
  color: var(--journal-ink);
  line-height: 1.72;
  font-size: var(--font-size-0-90);
}

.import-review__description :deep(*:first-child) {
  margin-top: 0;
}

.import-review__description :deep(*:last-child) {
  margin-bottom: 0;
}

.import-review__description :deep(h1),
.import-review__description :deep(h2),
.import-review__description :deep(h3),
.import-review__description :deep(h4) {
  margin-top: var(--space-2-5);
  margin-bottom: var(--space-2);
  font-weight: 700;
  line-height: 1.35;
}

.import-review__description :deep(h1) {
  font-size: var(--font-size-1-16);
}

.import-review__description :deep(h2) {
  font-size: var(--font-size-1-05);
}

.import-review__description :deep(p),
.import-review__description :deep(ul),
.import-review__description :deep(ol),
.import-review__description :deep(pre),
.import-review__description :deep(blockquote) {
  margin-top: 0;
  margin-bottom: var(--space-2-5);
}

.import-review__description :deep(ul),
.import-review__description :deep(ol) {
  padding-left: var(--space-4);
}

.import-review__description :deep(pre),
.import-review__description :deep(code) {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.import-review__description :deep(pre) {
  padding: var(--space-2-5);
  border-radius: 0.75rem;
  border: 1px solid
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 82%, transparent);
  background: color-mix(
    in srgb,
    var(--journal-surface-subtle, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
  overflow: auto;
}

.import-review__description :deep(a) {
  color: var(--journal-accent);
  text-decoration: underline;
  text-underline-offset: 2px;
}
</style>
