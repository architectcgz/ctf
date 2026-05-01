<script setup lang="ts">
import type { AdminChallengeListItem, FlagType } from '@/api/contracts'
import ChallengeDescriptionPanel from '@/components/platform/challenge/ChallengeDescriptionPanel.vue'
import {
  ChallengeProfileMetaGrid,
  ChallengeProfileSummaryStrip,
  getChallengeStatusLabel,
} from '@/entities/challenge'

interface Props {
  challenge: AdminChallengeListItem
  downloadingAttachment: boolean
  flagConfigSummary: string
  flagDraftSummary: string
  flagType: FlagType
  flagValue: string
  flagRegex: string
  flagPrefix: string
  saving: boolean
  isSharedInstanceChallenge: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'download-attachment': []
  'save-flag-config': []
  'update:flagType': [value: FlagType]
  'update:flagValue': [value: string]
  'update:flagRegex': [value: string]
  'update:flagPrefix': [value: string]
}>()

function updateFlagType(event: Event): void {
  emit('update:flagType', (event.target as HTMLSelectElement).value as FlagType)
}

function updateFlagValue(event: Event): void {
  emit('update:flagValue', (event.target as HTMLInputElement).value)
}

function updateFlagRegex(event: Event): void {
  emit('update:flagRegex', (event.target as HTMLInputElement).value)
}

function updateFlagPrefix(event: Event): void {
  emit('update:flagPrefix', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <div class="admin-challenge-profile-panel">
    <header class="challenge-detail-header">
      <div class="challenge-detail-header__intro workspace-tab-heading__main">
        <div class="workspace-overline">
          Challenge Profile
        </div>
        <h1 class="workspace-page-title">
          题目管理
        </h1>
      </div>
      <p class="workspace-page-copy">
        聚合《{{ challenge.title }}》的基础信息、附件与判题模式配置，便于和拓扑、题解工作区来回切换。
      </p>
    </header>

    <ChallengeProfileSummaryStrip
      :challenge="challenge"
      :status-label="getChallengeStatusLabel(challenge.status)"
    />

    <section class="workspace-directory-section challenge-section challenge-profile-section">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">
            Challenge Directory
          </div>
          <h2 class="list-heading__title">
            基础信息
          </h2>
        </div>
      </header>

      <ChallengeProfileMetaGrid
        :challenge="challenge"
        :downloading-attachment="downloadingAttachment"
        :flag-config-summary="flagConfigSummary"
        @download-attachment="emit('download-attachment')"
      />
    </section>

    <section
      v-if="challenge.description"
      class="workspace-directory-section challenge-section"
    >
      <header class="list-heading">
        <div>
          <div class="journal-note-label">
            Challenge Description
          </div>
          <h2 class="list-heading__title">
            题目描述
          </h2>
        </div>
      </header>

      <ChallengeDescriptionPanel
        :content="challenge.description"
        label="描述"
        test-id="challenge-detail-description"
      />
    </section>

    <section
      v-if="challenge.hints?.length"
      class="workspace-directory-section challenge-section"
    >
      <div class="list-heading">
        <div>
          <div class="journal-note-label">
            Hints
          </div>
          <h2 class="list-heading__title">
            提示管理
          </h2>
        </div>
      </div>

      <div class="hint-list">
        <article
          v-for="hint in challenge.hints"
          :key="hint.id || hint.level"
          class="hint-card"
        >
          <div class="hint-card__title">
            Level {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}
          </div>
          <div class="hint-card__content">
            {{ hint.content }}
          </div>
        </article>
      </div>
    </section>

    <section class="workspace-directory-section challenge-section">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">
            Judge Mode
          </div>
          <h2 class="list-heading__title">
            判题模式配置
          </h2>
        </div>
      </header>

      <section class="journal-panel challenge-flag-panel p-5 md:p-6">
        <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <p class="challenge-flag-panel__copy">
            支持静态 Flag、动态前缀、正则判题和人工审核四种模式。保存后即时刷新当前题目配置。
          </p>
          <div class="flag-summary-chip">
            {{ flagDraftSummary }}
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="flag-field">
            <span class="flag-field-label">判题模式</span>
            <select
              :value="flagType"
              class="flag-field-input"
              @change="updateFlagType"
            >
              <option value="static">静态 Flag</option>
              <option value="dynamic">动态前缀</option>
              <option value="regex">正则匹配</option>
              <option value="manual_review">人工审核</option>
            </select>
          </label>

          <label
            v-if="flagType === 'dynamic' || flagType === 'regex'"
            class="flag-field"
          >
            <span class="flag-field-label">Flag 前缀</span>
            <input
              :value="flagPrefix"
              type="text"
              placeholder="例如：flag"
              class="flag-field-input"
              @input="updateFlagPrefix"
            >
          </label>

          <label
            v-if="flagType === 'static'"
            class="flag-field md:col-span-2"
          >
            <span class="flag-field-label">静态 Flag</span>
            <input
              :value="flagValue"
              type="text"
              placeholder="例如：flag{demo}"
              class="flag-field-input font-mono"
              @input="updateFlagValue"
            >
          </label>

          <label
            v-if="flagType === 'regex'"
            class="flag-field md:col-span-2"
          >
            <span class="flag-field-label">正则表达式</span>
            <input
              :value="flagRegex"
              type="text"
              placeholder="例如：^flag\{demo-[0-9]+\}$"
              class="flag-field-input font-mono"
              @input="updateFlagRegex"
            >
          </label>
        </div>

        <div
          v-if="isSharedInstanceChallenge"
          class="challenge-flag-panel__warning"
        >
          共享实例只适用于无状态题。该模式不提供用户级答案隔离，静态/正则答案可能被转发；若需隔离答案，请使用
          per_user 或 per_team。
        </div>

        <div
          v-if="flagType === 'manual_review'"
          class="challenge-flag-panel__warning"
        >
          学生提交的答案将进入教师审核队列。审核通过后才会计分并更新通过状态。
        </div>

        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="text-sm text-[var(--journal-muted)]">
            当前配置：{{ flagConfigSummary }}
          </div>
          <button
            :disabled="saving"
            class="ui-btn ui-btn--primary"
            type="button"
            @click="emit('save-flag-config')"
          >
            {{ saving ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </section>
    </section>
  </div>
</template>

<style scoped>
.workspace-overline {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.challenge-detail-header,
.challenge-profile-section {
  display: grid;
  gap: var(--space-2-5);
}

.challenge-detail-header {
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--workspace-line-soft, color-mix(in srgb, var(--journal-border) 88%, transparent));
}

.challenge-section {
  display: grid;
  gap: var(--space-4);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.hint-list {
  display: grid;
  gap: var(--space-3);
}

.hint-card {
  display: grid;
  gap: var(--space-2);
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 26%, transparent);
  padding: var(--space-3) 0 var(--space-3) var(--space-4);
  background: color-mix(in srgb, var(--journal-surface-subtle) 85%, transparent);
}

.hint-card__title {
  font-size: var(--font-size-0-90);
  font-weight: 700;
  color: var(--journal-ink);
}

.hint-card__content {
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-ink);
}

.challenge-flag-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-flag-panel__copy {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-muted);
}

.challenge-flag-panel__warning {
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  padding: var(--space-4);
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-ink);
}

.challenge-flag-panel .ui-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.flag-summary-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 20%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: var(--space-2) var(--space-3-5);
  font-size: var(--font-size-0-80);
  font-weight: 600;
  color: var(--journal-accent);
}

.flag-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2-5);
}

.flag-field-label {
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: var(--journal-ink);
}

.flag-field-input {
  min-height: 2.9rem;
  border: 1px solid var(--journal-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-92);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.flag-field-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

@media (max-width: 900px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .challenge-meta-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .challenge-meta-item,
  .challenge-meta-item:nth-child(odd),
  .challenge-meta-item:nth-child(even) {
    padding-inline: 0;
  }
}
</style>
