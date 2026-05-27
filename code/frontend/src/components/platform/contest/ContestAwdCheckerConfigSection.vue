<script setup lang="ts">
import { AlertTriangle } from 'lucide-vue-next'

import type { AWDCheckerType } from '@/api/contracts'
import ContestAwdHttpStandardFields from './ContestAwdHttpStandardFields.vue'
import ContestAwdLegacyProbeFields from './ContestAwdLegacyProbeFields.vue'
import ContestAwdScriptCheckerFields from './ContestAwdScriptCheckerFields.vue'
import ContestAwdTcpStandardFields from './ContestAwdTcpStandardFields.vue'
import type {
  AwdConfigFieldErrors,
  AwdHttpActionDraft,
  AwdHttpActionSection,
  AwdHttpStandardPreset,
  AwdLegacyProbeDraft,
  AwdScriptCheckerDraft,
  AwdTcpCheckerStepDraft,
  AwdTcpStandardDraft,
} from './contestAwdConfigTypes'

defineProps<{
  addTcpCheckerStep: () => void
  applyHttpPreset: (presetId: string) => void
  awdHttpMethodOptions: readonly string[]
  awdHttpStandardPresets: readonly AwdHttpStandardPreset[]
  expandedTcpCheckerStepIndex: number | null
  fieldErrors: AwdConfigFieldErrors
  getCheckerTypeLabel: (value?: AWDCheckerType) => string
  httpActionSections: readonly AwdHttpActionSection[]
  httpStandardDraft: Record<string, AwdHttpActionDraft>
  legacyProbeDraft: AwdLegacyProbeDraft
  removeTcpCheckerStep: (index: number) => void
  scriptCheckerDraft: AwdScriptCheckerDraft
  selectedCheckerType: AWDCheckerType | undefined
  summarizeTcpCheckerStep: (step: AwdTcpCheckerStepDraft) => string
  tcpStandardDraft: AwdTcpStandardDraft
  toggleTcpCheckerStep: (index: number) => void
}>()
</script>

<template>
  <section class="awd-config-form-section awd-config-card awd-config-card--canvas">
    <header class="list-heading awd-config-section-head">
      <div>
        <div class="journal-note-label">Checker Parameters</div>
        <h3 class="list-heading__title">{{ getCheckerTypeLabel(selectedCheckerType) }}</h3>
      </div>
      <span class="awd-config-section-tag">配置画布</span>
    </header>

    <div
      v-if="fieldErrors.checker_type"
      class="awd-config-alert"
    >
      <AlertTriangle class="h-4 w-4" />
      <span>{{ fieldErrors.checker_type }}，请先在 AWD 题库修正题目包协议与 checker 契约。</span>
    </div>

    <ContestAwdLegacyProbeFields
      v-if="selectedCheckerType === 'legacy_probe'"
      :field-errors="fieldErrors"
      :legacy-probe-draft="legacyProbeDraft"
    />

    <ContestAwdHttpStandardFields
      v-else-if="selectedCheckerType === 'http_standard'"
      :awd-http-method-options="awdHttpMethodOptions"
      :awd-http-standard-presets="awdHttpStandardPresets"
      :apply-http-preset="applyHttpPreset"
      :field-errors="fieldErrors"
      :http-action-sections="httpActionSections"
      :http-standard-draft="httpStandardDraft"
    />

    <ContestAwdTcpStandardFields
      v-else-if="selectedCheckerType === 'tcp_standard'"
      :add-tcp-checker-step="addTcpCheckerStep"
      :expanded-tcp-checker-step-index="expandedTcpCheckerStepIndex"
      :field-errors="fieldErrors"
      :remove-tcp-checker-step="removeTcpCheckerStep"
      :summarize-tcp-checker-step="summarizeTcpCheckerStep"
      :tcp-standard-draft="tcpStandardDraft"
      :toggle-tcp-checker-step="toggleTcpCheckerStep"
    />

    <ContestAwdScriptCheckerFields
      v-else-if="selectedCheckerType === 'script_checker'"
      :field-errors="fieldErrors"
      :script-checker-draft="scriptCheckerDraft"
    />
  </section>
</template>

<style scoped>
.awd-config-form-section {
  display: grid;
  gap: var(--space-3);
}

.awd-config-card {
  padding: var(--space-4);
  border: 1px solid var(--awd-card-border);
  border-radius: var(--awd-card-radius);
  background: var(--awd-card-surface);
  box-shadow: var(--awd-card-shadow);
}

.awd-config-card--canvas {
  gap: var(--space-4);
}

.awd-config-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}

.awd-config-alert {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-4);
  border-radius: var(--ui-control-radius);
  background: var(--color-warning-soft);
  padding: var(--space-2) var(--space-3);
  color: var(--color-warning);
  font-size: var(--font-size-13);
}

.awd-config-section-tag {
  flex: none;
  border-radius: var(--ui-badge-radius-soft);
  padding: var(--space-1) var(--space-2);
  background: color-mix(in srgb, var(--color-primary-soft) 55%, var(--color-bg-surface));
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 700;
}

:deep(.checker-action-grid) {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

:deep(.checker-toolbar) {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

:deep(.checker-action-section) {
  display: grid;
  gap: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid color-mix(in srgb, var(--color-border-default) 70%, transparent);
}

:deep(.checker-action-section--panel) {
  padding: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: calc(var(--awd-card-radius) - 0.125rem);
  background: var(--awd-card-subtle);
  box-shadow: 0 0.45rem 1rem color-mix(in srgb, var(--color-shadow-soft) 12%, transparent);
}

:deep(.checker-action-section--tcp.is-collapsed) {
  gap: 0;
  padding-block: var(--space-2);
}

:deep(.checker-action-section--panel .ui-control-wrap) {
  border: 1px solid var(--ui-control-border);
  background: var(--ui-control-background);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--color-text-primary) 5%, transparent);
}

:deep(.checker-action-section--panel .ui-control-wrap:focus-within) {
  border-color: var(--ui-control-focus-border);
  background: var(--ui-control-focus-background);
  box-shadow:
    var(--ui-control-focus-shadow),
    inset 0 1px 0 color-mix(in srgb, var(--color-text-primary) 7%, transparent);
}

:deep(.checker-action-section--panel .ui-control) {
  min-height: 2.25rem;
  background: transparent;
}

:deep(.checker-action-section--panel textarea.ui-control) {
  min-height: 3.5rem;
  line-height: 1.4;
}

:deep(.checker-action-section__head) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

:deep(.checker-step-toggle) {
  min-width: 0;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border: 0;
  background: transparent;
  padding: 0;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

:deep(.checker-step-toggle:focus-visible) {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 42%, transparent);
  outline-offset: var(--space-1);
}

:deep(.checker-step-toggle__icon) {
  flex: none;
  color: var(--color-text-secondary);
  transition: transform var(--ui-motion-fast);
}

:deep(.checker-step-toggle[aria-expanded='true'] .checker-step-toggle__icon) {
  transform: rotate(180deg);
}

:deep(.checker-action-section__heading) {
  min-width: 0;
  display: grid;
  gap: 0;
}

:deep(.checker-action-section__title) {
  font-size: var(--font-size-14);
}

:deep(.checker-action-section__hint) {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
}

:deep(.checker-action-extra-grid) {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

:deep(.checker-action-extra-grid__wide) {
  grid-column: span 2;
}

:deep(.checker-action-grid--http) {
  grid-template-columns: minmax(6.5rem, 8rem) minmax(0, 1fr) minmax(7rem, 8.5rem);
}

:deep(.checker-action-extra-grid--http) {
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr) minmax(0, 1fr);
}

:deep(.checker-action-extra-grid--tcp) {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

:deep(.checker-action-grid--script-meta) {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

:deep(.checker-action-extra-grid--script) {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

:deep(.checker-field--method),
:deep(.checker-field--status),
:deep(.checker-field--path) {
  min-width: 0;
}

:deep(.checker-field--wide) {
  grid-column: span 2;
}

:deep(.checker-preset-strip) {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

:deep(.checker-preset-strip--compact) {
  margin-bottom: var(--space-1);
}

:deep(.awd-config-small-field) {
  max-width: 18rem;
}

:deep(.ui-field) {
  gap: var(--space-1);
}

:deep(.ui-field__label) {
  font-size: var(--font-size-12);
}

:deep(.ui-control) {
  min-height: 2.5rem;
}

:deep(textarea.ui-control) {
  min-height: 4.5rem;
  resize: vertical;
}

:deep(.awd-config-control--mono) {
  font-family: var(--font-family-mono);
}

@media (max-width: 1023px) {
  :deep(.checker-action-grid),
  :deep(.checker-action-extra-grid) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  :deep(.checker-action-grid--http),
  :deep(.checker-action-extra-grid--http),
  :deep(.checker-action-extra-grid--tcp),
  :deep(.checker-action-extra-grid--script) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  :deep(.checker-action-extra-grid__wide),
  :deep(.checker-field--wide) {
    grid-column: 1 / -1;
  }
}

@media (max-width: 767px) {
  :deep(.checker-action-grid),
  :deep(.checker-action-extra-grid),
  :deep(.checker-action-grid--http),
  :deep(.checker-action-grid--script-meta),
  :deep(.checker-action-extra-grid--http),
  :deep(.checker-action-extra-grid--tcp),
  :deep(.checker-action-extra-grid--script) {
    grid-template-columns: 1fr;
  }

  :deep(.checker-action-extra-grid__wide),
  :deep(.checker-field--wide) {
    grid-column: auto;
  }
}
</style>
