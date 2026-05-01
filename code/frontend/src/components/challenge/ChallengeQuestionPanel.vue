<template>
  <section
    id="challenge-workspace-panel-question"
    class="workspace-panel panel panel--question"
    role="tabpanel"
    aria-labelledby="challenge-workspace-tab-question"
  >
    <div class="question-hero">
      <div class="question-hero-main">
        <div class="workspace-overline">
          Question
        </div>
        <h1 class="question-title workspace-page-title">
          {{ challenge.title }}
        </h1>
        <ChallengeMetaStrip :challenge="challenge" />
      </div>

      <aside
        class="score-rail"
        @click="emit('score-rail-probe')"
      >
        <div class="score-label">
          分值
        </div>
        <div class="score-value">
          {{ challenge.points }} <small>pts</small>
        </div>
        <div class="score-note">
          {{ challenge.attachment_url ? '当前题目包含附件。' : '当前题目无附件。' }}
        </div>
        <div
          v-if="scoreRailProbeMessage"
          class="score-probe-note"
        >
          {{ scoreRailProbeMessage }}
        </div>
      </aside>
    </div>

    <section class="section">
      <div class="section-head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Statement
          </div>
          <h2 class="section-title workspace-tab-heading__title">
            题目描述
          </h2>
        </div>
        <button
          v-if="challenge.attachment_url"
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="emit('download-attachment')"
        >
          下载附件
        </button>
      </div>
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div
        class="prose challenge-prose description max-w-none"
        v-html="sanitizedDescription"
      />
    </section>

    <section
      v-if="challenge.hints.length > 0"
      class="section"
    >
      <div class="section-head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Hints
          </div>
          <h2 class="section-title workspace-tab-heading__title">
            提示
          </h2>
        </div>
        <div class="section-hint">
          共 {{ challenge.hints.length }} 条
        </div>
      </div>
      <div class="hint-list">
        <div
          v-for="hint in challenge.hints"
          :key="hint.id"
          class="hint-line"
        >
          <div>
            <div class="hint-label">
              提示 {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}
            </div>
            <div
              v-if="isHintExpanded(hint.level)"
              :id="`challenge-hint-panel-${hint.id}`"
              class="hint-copy"
            >
              {{ hint.content || '暂无提示内容' }}
            </div>
          </div>
          <button
            type="button"
            class="ui-btn ui-btn--sm ui-btn--ghost hint-toggle"
            :aria-expanded="isHintExpanded(hint.level)"
            :aria-controls="`challenge-hint-panel-${hint.id}`"
            @click="emit('toggle-hint', hint.level)"
          >
            {{ isHintExpanded(hint.level) ? '收起提示' : '展开提示' }}
          </button>
        </div>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import type { ChallengeDetailData } from '@/api/contracts'
import { ChallengeMetaStrip } from '@/entities/challenge'

interface Props {
  challenge: ChallengeDetailData
  sanitizedDescription: string
  scoreRailProbeMessage: string
  isHintExpanded: (level: number) => boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'download-attachment': []
  'toggle-hint': [level: number]
  'score-rail-probe': []
}>()
</script>

<style scoped>
.question-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 200px;
  gap: var(--space-6);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--line-soft);
}

.question-hero-main {
  min-width: 0;
}

.question-title {
  margin: var(--space-3) 0 0;
  color: var(--text-main);
}

.score-rail {
  padding-left: var(--space-5-5);
  border-left: 1px solid var(--line-soft);
}

.score-label {
  font-size: var(--font-size-11);
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.score-value {
  margin-top: var(--space-2);
  color: var(--text-main);
  font: 700 34px/1 var(--font-mono);
}

.score-value small {
  font-size: var(--font-size-16);
  color: var(--text-faint);
}

.score-note {
  margin-top: var(--space-4);
  padding-top: var(--space-4);
  border-top: 1px solid var(--line-soft);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--text-subtle);
}

.score-probe-note {
  margin-top: var(--space-3);
  font-size: var(--font-size-12);
  font-weight: 700;
  line-height: 1.7;
  color: color-mix(in srgb, var(--journal-accent) 80%, var(--text-subtle));
}

.section {
  padding-top: var(--space-6);
  border-top: 1px solid var(--line-soft);
}

.section:first-of-type {
  border-top: 0;
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.section-hint {
  font-size: var(--font-size-13);
  line-height: 1.75;
  color: var(--text-faint);
}

.description {
  font-size: var(--font-size-15);
  line-height: 1.92;
  color: var(--text-subtle);
}

.challenge-prose :deep(p),
.challenge-prose :deep(ul),
.challenge-prose :deep(ol) {
  margin-bottom: var(--space-4);
}

.challenge-prose :deep(pre) {
  overflow: auto;
  margin: var(--space-5) 0;
  padding: var(--space-4-5) var(--space-5);
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--bg-panel) 72%, var(--color-bg-base));
  color: var(--text-main);
  font: 13px/1.7 var(--font-mono);
}

.challenge-prose :deep(h1),
.challenge-prose :deep(h2),
.challenge-prose :deep(h3),
.challenge-prose :deep(strong),
.challenge-prose :deep(code) {
  color: var(--journal-ink);
}

.hint-list {
  display: flex;
  flex-direction: column;
}

.hint-line {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-3-5);
  padding: var(--space-4) 0;
  border-top: 1px dashed var(--line-soft);
}

.hint-line:first-of-type {
  padding-top: 0;
  border-top: 0;
}

.hint-label {
  font-size: var(--font-size-14);
  font-weight: 600;
  color: var(--text-main);
}

.hint-copy {
  margin-top: var(--space-2-5);
  font-size: var(--font-size-14);
  line-height: 1.8;
  color: var(--text-subtle);
}

.hint-toggle {
  --ui-btn-height: 2.5rem;
}

@media (max-width: 1080px) {
  .question-hero {
    grid-template-columns: minmax(0, 1fr);
  }

  .score-rail {
    padding-left: 0;
    padding-top: var(--space-4-5);
    border-left: 0;
    border-top: 1px solid var(--line-soft);
  }
}
</style>
