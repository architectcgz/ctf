<template>
  <AppRouteRedirect :to="commitSuccessRedirectRoute" />

  <ChallengeImportPreviewWorkspacePanel
    :preview="preview"
    :uploading="uploading"
    :committing="committing"
    :has-preview="hasPreview"
    :back-to-import-route="backToImportRoute"
    :back-to-queue-route="backToQueueRoute"
    @confirm="void handleCommitPreview()"
  />
</template>

<script setup lang="ts">
import { toRef } from 'vue'

import AppRouteRedirect from '@/components/navigation/AppRouteRedirect.vue'
import {
  ChallengeImportPreviewWorkspacePanel,
  useChallengeImportPreviewPage,
} from '@/features/platform/challenge-package-import'

const props = defineProps<{
  importId: string
}>()

const {
  preview,
  uploading,
  committing,
  hasPreview,
  backToImportRoute,
  backToQueueRoute,
  commitSuccessRedirectRoute,
  handleCommitPreview,
} = useChallengeImportPreviewPage(toRef(props, 'importId'))
</script>
