<template>
  <aside class="detail-aside tool-pane">
    <div class="tool-pane-inner">
      <section class="tool-group">
        <div>
          <div class="workspace-overline">
            Primary Action
          </div>
          <h2 class="tool-title">
            {{ submitPanelTitle }}
          </h2>
          <p class="tool-copy">
            {{ submitPanelCopy }}
          </p>
        </div>
        <span
          v-if="challengeSolved"
          class="writeup-status-pill writeup-status-pill--success"
        >
          已通过
        </span>
        <div class="flag-field">
          <label
            for="challenge-flag-input"
            class="flag-label"
          >
            {{ submitFieldLabel }}
          </label>
          <div class="flag-row">
            <div
              class="ui-control-wrap flag-input-wrap"
              :class="[submitInputClass, { 'is-disabled': submitting }]"
            >
              <input
                id="challenge-flag-input"
                :value="flagInput"
                type="text"
                aria-label="Flag"
                :placeholder="submitPlaceholder"
                class="ui-control challenge-input flag-input disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="submitting"
                @input="updateFlagInput"
                @keyup.enter="emit('submit-flag')"
              >
            </div>
            <button
              type="button"
              :disabled="submitting"
              class="ui-btn ui-btn--primary disabled:cursor-not-allowed disabled:opacity-50"
              @click="emit('submit-flag')"
            >
              {{ submitting ? '提交中...' : '提交' }}
            </button>
          </div>
        </div>
        <div
          v-if="submitResult"
          class="status-inline"
          :class="`status-inline--${submitResult.variant}`"
        >
          <span class="status-dot" />
          {{ submitResult.message }}
        </div>
      </section>

      <ChallengeInstanceCard
        v-if="needTarget"
        :instance="instance"
        :instance-sharing="instanceSharing"
        :loading="instanceLoading"
        :creating="instanceCreating"
        :opening="instanceOpening"
        :extending="instanceExtending"
        :destroying="instanceDestroying"
        :challenge-solved="challengeSolved"
        @start="emit('start-instance')"
        @open="emit('open-instance')"
        @extend="emit('extend-instance')"
        @destroy="emit('destroy-instance')"
      />
      <section
        v-else
        class="tool-group detail-aside-empty text-sm text-[var(--color-success)]"
      >
        该题目不需要靶机，可直接分析题面并提交 Flag。
      </section>
    </div>
  </aside>
</template>

<script setup lang="ts">
import type { InstanceData, InstanceSharing } from '@/api/contracts'
import ChallengeInstanceCard from '@/components/challenge/ChallengeInstanceCard.vue'

interface SubmitResultState {
  variant: 'success' | 'error' | 'pending'
  message: string
}

interface Props {
  needTarget: boolean
  challengeSolved: boolean
  submitPanelTitle: string
  submitPanelCopy: string
  submitFieldLabel: string
  submitInputClass: string
  submitPlaceholder: string
  submitting: boolean
  flagInput: string
  submitResult: SubmitResultState | null
  instance: InstanceData | null
  instanceSharing: InstanceSharing
  instanceLoading: boolean
  instanceCreating: boolean
  instanceOpening: boolean
  instanceExtending: boolean
  instanceDestroying: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'update:flagInput': [value: string]
  'submit-flag': []
  'start-instance': []
  'open-instance': []
  'extend-instance': []
  'destroy-instance': []
}>()

function updateFlagInput(event: Event): void {
  emit('update:flagInput', (event.target as HTMLInputElement).value)
}
</script>

<style scoped>
.tool-pane-inner {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 100%;
  position: sticky;
  top: var(--space-7);
}

.flag-label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-12);
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--text-faint);
}

.flag-input-wrap {
  border-color: var(--line-strong);
  background: var(--bg-panel);
  --ui-control-height: 3.125rem;
}

.flag-input-wrap > input {
  font: 500 15px/1 var(--font-mono);
}

.flag-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-3);
}

.tool-group + .tool-group,
:deep(.instance-shell) {
  margin-top: var(--space-6);
}

.tool-group + .tool-group {
  margin-top: var(--space-6);
  padding-top: var(--space-6);
  border-top: 1px solid var(--line-soft);
}

.tool-title {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-18);
  line-height: 1.25;
  color: var(--text-main);
}

.tool-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-14);
  line-height: 1.75;
  color: var(--text-subtle);
}

.status-inline {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-3);
  font-size: var(--font-size-14);
  color: var(--text-subtle);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: var(--warning);
}

.status-inline--success {
  color: var(--journal-success-ink);
}

.status-inline--success .status-dot {
  background: var(--color-success);
}

.status-inline--pending {
  color: var(--journal-warning-ink);
}

.status-inline--pending .status-dot {
  background: var(--color-warning);
}

.status-inline--error {
  color: var(--journal-danger-ink);
}

.status-inline--error .status-dot {
  background: var(--color-danger);
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

.writeup-status-pill--success {
  border-color: color-mix(in srgb, var(--color-success) 18%, transparent);
  background: var(--journal-success-soft);
  color: var(--journal-success-ink);
}

@media (max-width: 1080px) {
  .tool-pane-inner {
    min-height: 0;
    position: static;
  }
}

@media (max-width: 760px) {
  .flag-row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
