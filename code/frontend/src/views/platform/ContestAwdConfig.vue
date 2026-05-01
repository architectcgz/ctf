<script setup lang="ts">
import {
  AlertTriangle,
  ArrowLeft,
  CheckCircle2,
  ChevronDown,
  Code2,
  Play,
  RefreshCw,
  Save,
  ShieldCheck,
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import ContestAwdConfigFooter from '@/components/platform/contest/ContestAwdConfigFooter.vue'
import ContestAwdConfigTopbar from '@/components/platform/contest/ContestAwdConfigTopbar.vue'
import ContestAwdDebugStation from '@/components/platform/contest/ContestAwdDebugStation.vue'
import ContestAwdEditorHeader from '@/components/platform/contest/ContestAwdEditorHeader.vue'
import ContestAwdScoreWeights from '@/components/platform/contest/ContestAwdScoreWeights.vue'
import ContestAwdServiceDirectory from '@/components/platform/contest/ContestAwdServiceDirectory.vue'
import { useContestAwdConfigPage } from '@/features/contest-awd-config'

const {
  AWD_HTTP_METHOD_OPTIONS,
  AWD_HTTP_STANDARD_PRESETS,
  addTCPCheckerStep,
  applyHTTPPreset,
  canAttachPreviewToken,
  checkerConfigJSON,
  contest,
  expandedTCPCheckerStepIndex,
  fieldErrors,
  form,
  getCheckStatusLabel,
  getCheckerTypeLabel,
  getProtocolLabel,
  getValidationLabel,
  goBackToStudio,
  handlePreview,
  handleSave,
  httpActionSections,
  httpStandardDraft,
  legacyProbeDraft,
  loadError,
  loading,
  loadPage,
  previewAccessURL,
  previewError,
  previewForm,
  previewResult,
  previewSummary,
  previewing,
  refreshing,
  removeTCPCheckerStep,
  saving,
  scriptCheckerDraft,
  selectService,
  selectedCheckerType,
  selectedService,
  selectedServiceId,
  sortedServices,
  summarizeTCPCheckerStep,
  tcpStandardDraft,
  toggleTCPCheckerStep,
} = useContestAwdConfigPage()
</script>

<template>
  <section class="awd-config-page workspace-shell journal-shell journal-shell-admin">
    <div v-if="loading" class="awd-config-page__loading">
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
        <button type="button" class="ui-btn ui-btn--primary" @click="loadPage(true)">
          重试
        </button>
      </template>
    </AppEmpty>

    <main v-else class="awd-config-page__body">
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

          <div v-if="fieldErrors.checker_type" class="awd-config-alert">
            <AlertTriangle class="h-4 w-4" />
            <span>{{ fieldErrors.checker_type }}，请先在 AWD 题库修正题目包协议与 checker 契约。</span>
          </div>

          <ContestAwdScoreWeights
            v-model:sla-score="form.sla_score"
            v-model:defense-score="form.defense_score"
            :sla-error="fieldErrors.sla_score"
            :defense-error="fieldErrors.defense_score"
          />

          <section class="awd-config-form-section awd-config-card awd-config-card--canvas">
            <header class="list-heading awd-config-section-head">
              <div>
                <div class="journal-note-label">Checker Parameters</div>
                <h3 class="list-heading__title">{{ getCheckerTypeLabel(selectedCheckerType) }}</h3>
              </div>
              <span class="awd-config-section-tag">配置画布</span>
            </header>

            <label v-if="selectedCheckerType === 'legacy_probe'" class="ui-field">
              <span class="ui-field__label">健康检查路径</span>
              <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.legacy_health_path }">
                <input v-model="legacyProbeDraft.health_path" type="text" class="ui-control" placeholder="/healthz" />
              </span>
              <span v-if="fieldErrors.legacy_health_path" class="ui-field__error">{{ fieldErrors.legacy_health_path }}</span>
            </label>

            <template v-else-if="selectedCheckerType === 'http_standard'">
              <div class="checker-preset-strip checker-preset-strip--compact">
                <button
                  v-for="preset in AWD_HTTP_STANDARD_PRESETS"
                  :key="preset.id"
                  type="button"
                  class="ui-btn ui-btn--secondary checker-preset-button"
                  @click="applyHTTPPreset(preset.id)"
                >
                  {{ preset.label }}
                </button>
              </div>

              <section
                v-for="action in httpActionSections"
                :key="action.key"
                class="checker-action-section checker-action-section--panel"
              >
                <header class="list-heading checker-action-section__head">
                  <div class="checker-action-section__heading">
                    <h4 class="list-heading__title checker-action-section__title">{{ action.title }}</h4>
                    <span class="checker-action-section__hint">动作配置</span>
                  </div>
                </header>
                <div class="checker-action-grid checker-action-grid--http">
                  <label class="ui-field checker-field checker-field--method">
                    <span class="ui-field__label">Method</span>
                    <span class="ui-control-wrap">
                      <select v-model="httpStandardDraft[action.key].method" class="ui-control">
                        <option v-for="method in AWD_HTTP_METHOD_OPTIONS" :key="method" :value="method">{{ method }}</option>
                      </select>
                    </span>
                  </label>
                  <label class="ui-field checker-field checker-field--path">
                    <span class="ui-field__label">Path</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': action.pathErrorKey ? !!fieldErrors[action.pathErrorKey] : false }">
                      <input v-model="httpStandardDraft[action.key].path" type="text" class="ui-control" />
                    </span>
                    <span v-if="action.pathErrorKey && fieldErrors[action.pathErrorKey]" class="ui-field__error">{{ fieldErrors[action.pathErrorKey] }}</span>
                  </label>
                  <label class="ui-field checker-field checker-field--status">
                    <span class="ui-field__label">状态码</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors[action.statusErrorKey] }">
                      <input v-model.number="httpStandardDraft[action.key].expected_status" type="number" min="1" step="1" class="ui-control" />
                    </span>
                    <span v-if="fieldErrors[action.statusErrorKey]" class="ui-field__error">{{ fieldErrors[action.statusErrorKey] }}</span>
                  </label>
                </div>
                <div class="checker-action-extra-grid checker-action-extra-grid--http">
                  <label class="ui-field checker-field checker-field--wide">
                    <span class="ui-field__label">Body Template</span>
                    <span class="ui-control-wrap">
                      <textarea v-model="httpStandardDraft[action.key].body_template" rows="2" class="ui-control awd-config-control--mono" />
                    </span>
                  </label>
                  <label class="ui-field">
                    <span class="ui-field__label">Expected Substring</span>
                    <span class="ui-control-wrap">
                      <input v-model="httpStandardDraft[action.key].expected_substring" type="text" class="ui-control awd-config-control--mono" />
                    </span>
                  </label>
                  <label class="ui-field checker-action-extra-grid__wide">
                    <span class="ui-field__label">Headers JSON</span>
                    <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors[action.headersErrorKey] }">
                      <textarea v-model="httpStandardDraft[action.key].headers_text" rows="2" class="ui-control awd-config-control--mono" />
                    </span>
                    <span v-if="fieldErrors[action.headersErrorKey]" class="ui-field__error">{{ fieldErrors[action.headersErrorKey] }}</span>
                  </label>
                </div>
              </section>
            </template>

            <template v-else-if="selectedCheckerType === 'tcp_standard'">
              <div class="checker-toolbar">
                <label class="ui-field awd-config-small-field">
                  <span class="ui-field__label">总超时</span>
                  <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.tcp_timeout }">
                    <input v-model.number="tcpStandardDraft.timeout_ms" type="number" min="1" max="60000" step="100" class="ui-control" />
                  </span>
                  <span v-if="fieldErrors.tcp_timeout" class="ui-field__error">{{ fieldErrors.tcp_timeout }}</span>
                </label>
                <button type="button" class="ui-btn ui-btn--secondary" @click="addTCPCheckerStep">添加步骤</button>
              </div>
              <span v-if="fieldErrors.tcp_steps" class="ui-field__error">{{ fieldErrors.tcp_steps }}</span>
              <section
                v-for="(step, index) in tcpStandardDraft.steps"
                :key="index"
                class="checker-action-section checker-action-section--panel checker-action-section--tcp"
                :class="{ 'is-collapsed': expandedTCPCheckerStepIndex !== index }"
              >
                <header class="list-heading checker-action-section__head">
                  <button
                    type="button"
                    class="checker-step-toggle"
                    :aria-expanded="expandedTCPCheckerStepIndex === index"
                    @click="toggleTCPCheckerStep(index)"
                  >
                    <span class="checker-action-section__heading">
                      <span class="list-heading__title checker-action-section__title">Step {{ index + 1 }}</span>
                      <span class="checker-action-section__hint">{{ summarizeTCPCheckerStep(step) }}</span>
                    </span>
                    <ChevronDown class="h-4 w-4 checker-step-toggle__icon" />
                  </button>
                  <button v-if="tcpStandardDraft.steps.length > 1" type="button" class="ui-btn ui-btn--secondary" @click="removeTCPCheckerStep(index)">删除</button>
                </header>
                <div v-show="expandedTCPCheckerStepIndex === index" class="checker-action-extra-grid checker-action-extra-grid--tcp">
                  <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Send</span><span class="ui-control-wrap"><textarea v-model="step.send" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Send Template</span><span class="ui-control-wrap"><textarea v-model="step.send_template" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field"><span class="ui-field__label">Send Hex</span><span class="ui-control-wrap"><textarea v-model="step.send_hex" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Expect Contains</span><span class="ui-control-wrap"><textarea v-model="step.expect_contains" rows="2" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field"><span class="ui-field__label">Expect Regex</span><span class="ui-control-wrap"><input v-model="step.expect_regex" type="text" class="ui-control awd-config-control--mono" /></span></label>
                  <label class="ui-field"><span class="ui-field__label">Step Timeout</span><span class="ui-control-wrap"><input v-model.number="step.timeout_ms" type="number" min="0" max="60000" step="100" class="ui-control" /></span></label>
                </div>
              </section>
            </template>

            <template v-else-if="selectedCheckerType === 'script_checker'">
              <div class="checker-action-grid checker-action-grid--script-meta">
                <label class="ui-field"><span class="ui-field__label">Runtime</span><span class="ui-control-wrap"><select v-model="scriptCheckerDraft.runtime" class="ui-control"><option value="python3">python3</option></select></span></label>
                <label class="ui-field"><span class="ui-field__label">输出格式</span><span class="ui-control-wrap"><select v-model="scriptCheckerDraft.output" class="ui-control"><option value="exit_code">Exit Code</option><option value="json">JSON</option></select></span></label>
                <label class="ui-field"><span class="ui-field__label">超时时间</span><span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_timeout }"><input v-model.number="scriptCheckerDraft.timeout_sec" type="number" min="1" max="60" step="1" class="ui-control" /></span><span v-if="fieldErrors.script_timeout" class="ui-field__error">{{ fieldErrors.script_timeout }}</span></label>
              </div>
              <label class="ui-field">
                <span class="ui-field__label">入口文件</span>
                <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_entry }"><input v-model="scriptCheckerDraft.entry" type="text" class="ui-control" /></span>
                <span v-if="fieldErrors.script_entry" class="ui-field__error">{{ fieldErrors.script_entry }}</span>
              </label>
              <div class="checker-action-extra-grid checker-action-extra-grid--script">
                <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Args</span><span class="ui-control-wrap"><textarea v-model="scriptCheckerDraft.args_text" rows="3" class="ui-control awd-config-control--mono" /></span></label>
                <label class="ui-field checker-field checker-field--wide"><span class="ui-field__label">Env JSON</span><span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.script_env_text }"><textarea v-model="scriptCheckerDraft.env_text" rows="3" class="ui-control awd-config-control--mono" /></span><span v-if="fieldErrors.script_env_text" class="ui-field__error">{{ fieldErrors.script_env_text }}</span></label>
              </div>
            </template>
          </section>

          <ContestAwdDebugStation
            v-model:access-url="previewForm.access_url"
            v-model:preview-flag="previewForm.preview_flag"
            :checker-config-json="checkerConfigJSON"
            :previewing="previewing"
            :preview-result="previewResult"
            :preview-error="previewError"
            :preview-access-url="previewAccessURL"
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
  border-radius: var(--ui-control-radius);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-13);
}

.awd-config-alert {
  margin-top: var(--space-4);
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

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

.awd-config-card--compact {
  gap: var(--space-2);
}

.awd-config-card--canvas {
  gap: var(--space-4);
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

.checker-action-grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-toolbar {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.checker-action-section {
  display: grid;
  gap: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid color-mix(in srgb, var(--color-border-default) 70%, transparent);
}

.checker-action-section--panel {
  padding: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: calc(var(--awd-card-radius) - 0.125rem);
  background: var(--awd-card-subtle);
  box-shadow: 0 0.45rem 1rem color-mix(in srgb, var(--color-shadow-soft) 12%, transparent);
}

.checker-action-section--tcp.is-collapsed {
  gap: 0;
  padding-block: var(--space-2);
}

.checker-action-section--panel :deep(.ui-control-wrap) {
  border: 1px solid var(--ui-control-border);
  background: var(--ui-control-background);
  box-shadow: inset 0 1px 0 color-mix(in srgb, var(--color-text-primary) 5%, transparent);
}

.checker-action-section--panel :deep(.ui-control-wrap:focus-within) {
  border-color: var(--ui-control-focus-border);
  background: var(--ui-control-focus-background);
  box-shadow:
    var(--ui-control-focus-shadow),
    inset 0 1px 0 color-mix(in srgb, var(--color-text-primary) 7%, transparent);
}

.checker-action-section--panel :deep(.ui-control) {
  background: transparent;
}

.checker-action-section--panel :deep(.ui-control) {
  min-height: 2.25rem;
}

.checker-action-section--panel textarea.ui-control {
  min-height: 3.5rem;
  line-height: 1.4;
}

.checker-action-section__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.checker-step-toggle {
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

.checker-step-toggle:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 42%, transparent);
  outline-offset: var(--space-1);
}

.checker-step-toggle__icon {
  flex: none;
  color: var(--color-text-secondary);
  transition: transform var(--ui-motion-fast);
}

.checker-step-toggle[aria-expanded='true'] .checker-step-toggle__icon {
  transform: rotate(180deg);
}

.checker-action-section__heading {
  min-width: 0;
  display: grid;
  gap: 0;
}

.checker-action-section__title {
  font-size: var(--font-size-14);
}

.checker-action-section__hint {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
}

.checker-action-extra-grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-action-extra-grid__wide {
  grid-column: span 2;
}

.checker-action-grid--http {
  grid-template-columns: minmax(6.5rem, 8rem) minmax(0, 1fr) minmax(7rem, 8.5rem);
}

.checker-action-extra-grid--http {
  grid-template-columns: minmax(0, 1.15fr) minmax(0, 0.85fr) minmax(0, 1fr);
}

.checker-action-extra-grid--tcp {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.checker-action-grid--script-meta,
.checker-action-grid--preview {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.checker-action-extra-grid--script {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.checker-field--method,
.checker-field--status {
  min-width: 0;
}

.checker-field--path {
  min-width: 0;
}

.checker-field--wide {
  grid-column: span 2;
}

.checker-preset-strip {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.checker-preset-strip--compact {
  margin-bottom: var(--space-1);
}

.awd-config-small-field {
  max-width: 18rem;
}

.awd-config-page :deep(.ui-field) {
  gap: var(--space-1);
}

.awd-config-page :deep(.ui-field__label) {
  font-size: var(--font-size-12);
}

.awd-config-page :deep(.ui-control) {
  min-height: 2.5rem;
}

.awd-config-page textarea.ui-control {
  min-height: 4.5rem;
  resize: vertical;
}

.awd-config-control--mono {
  font-family: var(--font-family-mono);
}

.awd-config-page__empty {
  margin: var(--space-8);
}

@media (max-width: 1023px) {
  .awd-config-page__body {
    grid-template-columns: 1fr;
  }

  .checker-action-grid,
  .checker-action-extra-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .checker-action-grid--http,
  .checker-action-grid--preview,
  .checker-action-extra-grid--http,
  .checker-action-extra-grid--tcp,
  .checker-action-extra-grid--script {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .checker-action-extra-grid__wide {
    grid-column: 1 / -1;
  }

  .checker-field--wide {
    grid-column: 1 / -1;
  }
}

@media (max-width: 767px) {
  .checker-action-grid,
  .checker-action-extra-grid {
    grid-template-columns: 1fr;
  }

  .checker-action-extra-grid__wide {
    grid-column: auto;
  }

  .checker-field--wide {
    grid-column: auto;
  }

  .checker-action-grid--http,
  .checker-action-grid--script-meta,
  .checker-action-grid--preview,
  .checker-action-extra-grid--http,
  .checker-action-extra-grid--tcp,
  .checker-action-extra-grid--script {
    grid-template-columns: 1fr;
  }
}
</style>
