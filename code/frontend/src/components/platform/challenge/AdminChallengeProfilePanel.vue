<script setup lang="ts">
import type { AdminChallengeListItem } from '@/api/contracts'
import ChallengeDescriptionPanel from '@/components/platform/challenge/ChallengeDescriptionPanel.vue'
import {
  ChallengeProfileMetaGrid,
  ChallengeProfileSummaryStrip,
  getChallengeStatusLabel,
} from '@/entities/challenge'
import {
  PlatformChallengeFlagConfigPanel,
  type PlatformChallengeFlagDraft,
  type PlatformChallengeFlagDraftPatch,
} from '@/features/platform-challenge-detail'

interface Props {
  challenge: AdminChallengeListItem
  downloadingAttachment: boolean
  flagDraft: PlatformChallengeFlagDraft
}

defineProps<Props>()

const emit = defineEmits<{
  'download-attachment': []
  'save-flag-config': []
  'update:flag-draft': [value: PlatformChallengeFlagDraftPatch]
}>()
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
        :flag-config-summary="flagDraft.flagConfigSummary"
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

      <PlatformChallengeFlagConfigPanel
        :draft="flagDraft"
        @save="emit('save-flag-config')"
        @update:draft="emit('update:flag-draft', $event)"
      />
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

@media (max-width: 900px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
