<template>
  <section
    id="challenge-workspace-panel-writeup"
    class="workspace-panel panel"
    role="tabpanel"
    aria-labelledby="challenge-workspace-tab-writeup"
  >
    <section class="section section--flat">
      <div class="section-head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            My Writeup
          </div>
          <h2 class="section-title workspace-tab-heading__title">
            编写题解
          </h2>
        </div>
        <div class="section-hint">
          解题过程复盘 · {{ challengeSolved ? '可发布到社区' : '仅可保存草稿' }}
        </div>
      </div>

      <div
        v-if="myWriteup?.visibility_status === 'hidden'"
        class="inline-note inline-note--warning"
      >
        当前题解已被教师或管理员隐藏，仅你自己可见。
      </div>

      <div class="meta-strip meta-strip--compact">
        <span class="writeup-status-pill writeup-status-pill--primary">
          {{ submissionStatusLabel(myWriteup?.submission_status) }}
        </span>
        <span
          v-if="myWriteup?.visibility_status === 'hidden'"
          class="writeup-status-pill writeup-status-pill--warning"
        >
          已隐藏
        </span>
        <span
          v-else-if="
            myWriteup?.submission_status === 'published' ||
              myWriteup?.submission_status === 'submitted'
          "
          class="writeup-status-pill writeup-status-pill--success"
        >
          社区可见
        </span>
        <span
          v-if="myWriteup?.is_recommended"
          class="writeup-status-pill writeup-status-pill--primary"
        >
          推荐题解
        </span>
      </div>

      <form
        class="writeup-form"
        @submit.prevent
      >
        <div class="field">
          <label for="challenge-writeup-title">标题</label>
          <div class="ui-control-wrap">
            <input
              id="challenge-writeup-title"
              :value="writeupTitle"
              type="text"
              maxlength="256"
              placeholder="例如：从回显异常到拿到 flag 的完整链路"
              class="ui-control challenge-input"
              @input="updateWriteupTitle"
            >
          </div>
        </div>

        <div class="field">
          <label for="challenge-writeup-content">正文</label>
          <div class="ui-control-wrap writeup-textarea-wrap">
            <textarea
              id="challenge-writeup-content"
              :value="writeupContent"
              rows="10"
              placeholder="建议按『题目理解 → 利用过程 → 核心 payload / 证据 → 踩坑点』组织。"
              class="ui-control challenge-input writeup-textarea"
              @input="updateWriteupContent"
            />
          </div>
        </div>

        <div class="writeup-foot">
          <div class="writeup-footnote">
            {{
              submissionLoading
                ? '正在同步你的题解...'
                : myWriteup?.updated_at
                  ? `最近更新：${formatWriteupTime(myWriteup.updated_at)}`
                  : '还没有提交记录，可以先保存草稿。'
            }}
          </div>
          <div class="writeup-actions">
            <button
              type="button"
              :disabled="submissionLoading || submissionSaving !== null"
              class="ui-btn ui-btn--secondary disabled:cursor-not-allowed disabled:opacity-50"
              @click="emit('save', 'draft')"
            >
              {{ submissionSaving === 'draft' ? '保存中...' : '保存草稿' }}
            </button>
            <button
              type="button"
              :disabled="submissionLoading || submissionSaving !== null || !challengeSolved"
              class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
              @click="emit('save', 'published')"
            >
              {{ submissionSaving === 'published' ? '发布中...' : '发布题解' }}
            </button>
          </div>
        </div>
      </form>
    </section>
  </section>
</template>

<script setup lang="ts">
import type { SubmissionWriteupData, SubmissionWriteupStatus } from '@/api/contracts'

interface Props {
  challengeSolved: boolean
  myWriteup: SubmissionWriteupData | null
  submissionLoading: boolean
  submissionSaving: 'draft' | 'published' | null
  writeupTitle: string
  writeupContent: string
  formatWriteupTime: (value?: string) => string
  submissionStatusLabel: (status?: SubmissionWriteupStatus) => string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:writeupTitle': [value: string]
  'update:writeupContent': [value: string]
  save: [status: 'draft' | 'published']
}>()

function updateWriteupTitle(event: Event): void {
  emit('update:writeupTitle', (event.target as HTMLInputElement).value)
}

function updateWriteupContent(event: Event): void {
  emit('update:writeupContent', (event.target as HTMLTextAreaElement).value)
}
</script>

<style scoped>
.section--flat {
  padding-top: 0;
  border-top: 0;
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.section-hint,
.writeup-footnote {
  font-size: var(--font-size-13);
  line-height: 1.75;
  color: var(--text-faint);
}

.inline-note {
  padding-left: var(--space-4);
  border-left: 2px solid var(--line-soft);
  font-size: var(--font-size-0-90);
  line-height: 1.8;
  color: var(--text-subtle);
}

.inline-note--warning {
  border-left-color: color-mix(in srgb, var(--color-warning) 34%, transparent);
  color: var(--journal-warning-ink);
}

.meta-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
  margin-top: var(--space-4);
}

.meta-strip--compact {
  margin-top: 0;
  margin-bottom: var(--space-4);
}

.writeup-status-pill {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 var(--space-3-5);
  border: 1px solid var(--line-soft);
  border-radius: 999px;
  background: color-mix(in srgb, var(--bg-panel) 72%, transparent);
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--text-subtle);
}

.writeup-status-pill--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: var(--journal-accent-soft);
  color: var(--journal-accent-strong);
}

.writeup-status-pill--success {
  border-color: color-mix(in srgb, var(--color-success) 18%, transparent);
  background: var(--journal-success-soft);
  color: var(--journal-success-ink);
}

.writeup-status-pill--warning {
  border-color: color-mix(in srgb, var(--color-warning) 18%, transparent);
  background: var(--journal-warning-soft);
  color: var(--journal-warning-ink);
}

.writeup-form {
  display: grid;
  gap: var(--space-4);
}

.field label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-12);
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.field .ui-control-wrap {
  border-color: var(--line-strong);
  background: var(--bg-panel);
}

.field .ui-control-wrap:not(.writeup-textarea-wrap) {
  --ui-control-height: 3.125rem;
}

.field .ui-control-wrap > input,
.writeup-textarea-wrap > textarea {
  font: 500 14px/1.6 var(--font-sans);
}

.writeup-textarea-wrap > textarea {
  min-height: 260px;
  resize: vertical;
}

.writeup-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
}

.writeup-actions {
  display: flex;
  gap: var(--space-3);
  justify-content: flex-end;
}

@media (max-width: 760px) {
  .writeup-foot {
    flex-direction: column;
    align-items: stretch;
  }

  .writeup-actions {
    flex-direction: column;
  }
}
</style>
