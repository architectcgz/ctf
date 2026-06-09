<script setup lang="ts">
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import AppLoading from '@/shared/ui/common/AppLoading.vue'
import ContestAwdCheckerConfigSection from './ContestAwdCheckerConfigSection.vue'
import ContestAwdConfigFooter from './ContestAwdConfigFooter.vue'
import ContestAwdConfigTopbar from './ContestAwdConfigTopbar.vue'
import ContestAwdDebugStation from './ContestAwdDebugStation.vue'
import ContestAwdEditorHeader from './ContestAwdEditorHeader.vue'
import ContestAwdScoreWeights from './ContestAwdScoreWeights.vue'
import ContestAwdServiceDirectory from './ContestAwdServiceDirectory.vue'
import type {
  AdminContestAWDServiceData,
  AWDCheckerPreviewData,
  AWDCheckerType,
} from '@/api/contracts'
import type {
  AwdCheckerFormDraft,
  AwdHttpActionDraft,
  AwdHttpActionSection,
  AwdHttpStandardPreset,
  AwdLegacyProbeDraft,
  AwdPreviewFormDraft,
  AwdScriptCheckerDraft,
  AwdTcpCheckerStepDraft,
  AwdTcpStandardDraft,
} from './contestAwdConfigTypes'

interface Props {
  awdHttpMethodOptions: readonly string[]
  awdHttpStandardPresets: AwdHttpStandardPreset[]
  canAttachPreviewToken: boolean
  checkerConfigJson: string
  contest: { title?: string | null } | null
  expandedTcpCheckerStepIndex: number | null
  fieldErrors: Record<string, string | undefined>
  form: AwdCheckerFormDraft
  getCheckStatusLabel: (value: string) => string
  getCheckerTypeLabel: (value?: AWDCheckerType) => string
  getProtocolLabel: (value?: AWDCheckerType) => string
  getValidationLabel: (value?: AdminContestAWDServiceData['validation_state']) => string
  goBackToStudio: () => void
  handlePreview: () => Promise<void> | void
  handleSave: () => Promise<void> | void
  httpActionSections: readonly AwdHttpActionSection[]
  httpStandardDraft: Record<string, AwdHttpActionDraft>
  legacyProbeDraft: AwdLegacyProbeDraft
  loadError: string | null
  loading: boolean
  loadPage: (showLoading?: boolean) => Promise<void> | void
  previewAccessUrl: string
  previewError: string
  previewForm: AwdPreviewFormDraft
  previewResult: AWDCheckerPreviewData | null
  previewSummary: string
  previewing: boolean
  refreshing: boolean
  saving: boolean
  scriptCheckerDraft: AwdScriptCheckerDraft
  selectService: (service: AdminContestAWDServiceData) => void
  selectedCheckerType: AWDCheckerType | undefined
  selectedService: AdminContestAWDServiceData | null
  selectedServiceId: string
  sortedServices: AdminContestAWDServiceData[]
  summarizeTcpCheckerStep: (step: AwdTcpCheckerStepDraft) => string
  tcpStandardDraft: AwdTcpStandardDraft
  toggleTcpCheckerStep: (index: number) => void
  addTcpCheckerStep: () => void
  applyHttpPreset: (presetId: string) => void
  removeTcpCheckerStep: (index: number) => void
}

defineProps<Props>()
</script>

<template>
  <section class="awd-config-page workspace-shell journal-shell journal-shell-admin">
    <div
      v-if="loading"
      class="awd-config-page__loading"
    >
      <AppLoading>正在同步 AWD 配置...</AppLoading>
    </div>

    <ContestAwdConfigTopbar
      :contest-title="contest?.title || 'AWD 赛事'"
      :service-name="selectedService?.display_name || '请选择服务'"
      :refreshing="refreshing"
      @back="goBackToStudio"
      @refresh="loadPage(false)"
    />

    <AppEmpty
      v-if="loadError && !contest"
      title="AWD 配置加载失败"
      :description="loadError"
      icon="AlertTriangle"
      class="awd-config-page__empty"
    >
      <template #action>
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          @click="loadPage(true)"
        >
          重试
        </button>
      </template>
    </AppEmpty>

    <main
      v-else
      class="awd-config-page__body"
    >
      <ContestAwdServiceDirectory
        :loading="loading"
        :services="sortedServices"
        :selected-service-id="selectedServiceId"
        :get-checker-type-label="getCheckerTypeLabel"
        :get-validation-label="getValidationLabel"
        @select="selectService"
      />

      <section class="awd-config-page__editor">
        <AppEmpty
          v-if="!selectedService"
          title="请选择服务"
          description="从左侧目录选择一个 AWD 服务后继续配置。"
          icon="ShieldCheck"
          class="awd-config-page__empty"
        />

        <template v-else>
          <ContestAwdEditorHeader
            :display-name="selectedService.display_name"
            :title="selectedService.title || selectedService.display_name"
            :protocol-label="getProtocolLabel(selectedCheckerType)"
            :checker-type-label="getCheckerTypeLabel(selectedCheckerType)"
          />

          <ContestAwdScoreWeights
            v-model:sla-score="form.sla_score"
            v-model:defense-score="form.defense_score"
            :sla-error="fieldErrors.sla_score || ''"
            :defense-error="fieldErrors.defense_score || ''"
          />

          <ContestAwdCheckerConfigSection
            :add-tcp-checker-step="addTcpCheckerStep"
            :apply-http-preset="applyHttpPreset"
            :awd-http-method-options="awdHttpMethodOptions"
            :awd-http-standard-presets="awdHttpStandardPresets"
            :expanded-tcp-checker-step-index="expandedTcpCheckerStepIndex"
            :field-errors="fieldErrors"
            :get-checker-type-label="getCheckerTypeLabel"
            :http-action-sections="httpActionSections"
            :http-standard-draft="httpStandardDraft"
            :legacy-probe-draft="legacyProbeDraft"
            :remove-tcp-checker-step="removeTcpCheckerStep"
            :script-checker-draft="scriptCheckerDraft"
            :selected-checker-type="selectedCheckerType"
            :summarize-tcp-checker-step="summarizeTcpCheckerStep"
            :tcp-standard-draft="tcpStandardDraft"
            :toggle-tcp-checker-step="toggleTcpCheckerStep"
          />

          <ContestAwdDebugStation
            v-model:access-url="previewForm.access_url"
            v-model:preview-flag="previewForm.preview_flag"
            :checker-config-json="checkerConfigJson"
            :previewing="previewing"
            :preview-result="previewResult"
            :preview-error="previewError"
            :preview-access-url="previewAccessUrl"
            :preview-summary="previewSummary"
            :get-check-status-label="getCheckStatusLabel"
          />

          <ContestAwdConfigFooter
            :previewing="previewing"
            :saving="saving"
            :preview-error="previewError"
            :preview-result="previewResult"
            :can-attach-preview-token="canAttachPreviewToken"
            @preview="handlePreview"
            @save="handleSave"
          />
        </template>
      </section>
    </main>
  </section>
</template>

<style scoped>
.awd-config-page {
  --awd-card-radius: 0.75rem;
  --awd-card-border: color-mix(in srgb, var(--color-border-default) 80%, transparent);
  --awd-card-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --awd-card-subtle: color-mix(in srgb, var(--color-bg-surface) 72%, var(--color-bg-base));
  --awd-card-shadow: 0 0.85rem 2rem color-mix(in srgb, var(--color-shadow-soft) 22%, transparent);
  --ui-control-background: color-mix(
    in srgb,
    var(--color-bg-elevated) 62%,
    var(--color-bg-surface)
  );
  --ui-control-border: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --ui-control-color: var(--color-text-primary);
  --ui-control-placeholder: color-mix(in srgb, var(--color-text-muted) 86%, transparent);
  --ui-control-focus-border: color-mix(in srgb, var(--color-primary) 58%, var(--color-border-default));
  --ui-control-focus-background: color-mix(
    in srgb,
    var(--color-bg-surface) 76%,
    var(--color-bg-elevated)
  );
  --ui-control-focus-shadow: 0 0 0 0.2rem color-mix(in srgb, var(--color-primary) 16%, transparent);
  position: relative;
  min-height: calc(100vh - var(--app-header-height, 4rem));
  max-height: calc(100vh - var(--app-header-height, 4rem));
  display: flex;
  flex-direction: column;
  background: var(--color-bg-base);
  overflow: hidden;
}

.awd-config-page__loading {
  position: absolute;
  inset: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: center;
  background: color-mix(in srgb, var(--color-bg-base) 82%, transparent);
}

.awd-config-page__body {
  min-height: 0;
  height: calc(100vh - var(--app-header-height, 4rem) - 3.5rem);
  flex: 1;
  display: grid;
  grid-template-columns: minmax(17rem, 20rem) minmax(0, 1fr);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 54%, transparent),
      transparent 42%
    ),
    var(--color-bg-base);
}

.awd-config-page__editor {
  min-width: 0;
  min-height: 0;
  overflow: auto;
  padding: var(--space-5);
}

.awd-config-page__editor {
  display: grid;
  align-content: start;
  gap: var(--space-5);
}

.awd-config-page__empty {
  margin: var(--space-8);
}

@media (max-width: 1023px) {
  .awd-config-page__body {
    grid-template-columns: 1fr;
  }

}

@media (max-width: 767px) {
}
</style>
