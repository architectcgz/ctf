<script setup lang="ts">
import ChallengeImportPreviewWorkspacePanel from '@/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue'
import { useChallengeImportPreviewPage } from '@/features/challenge-package-import'

const {
  preview,
  uploading,
  committing,
  hasPreview,
  backToImportPanel,
  backToQueuePanel,
  handleCommitPreview,
} = useChallengeImportPreviewPage()
</script>

<template>
  <ChallengeImportPreviewWorkspacePanel
    :preview="preview"
    :uploading="uploading"
    :committing="committing"
    :has-preview="hasPreview"
    @back="void backToImportPanel()"
    @back-queue="void backToQueuePanel()"
    @confirm="void handleCommitPreview()"
  />
</template>
