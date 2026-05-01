<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ChallengeImportPreviewWorkspacePanel from '@/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue'
import { useChallengePackageImport } from '@/features/challenge-package-import'

const route = useRoute()
const router = useRouter()

const importId = computed(() => {
  const raw = route.params.importId
  if (typeof raw === 'string') {
    return raw
  }
  if (Array.isArray(raw)) {
    return raw[0] ?? ''
  }
  return ''
})

const { preview, uploading, committing, hasPreview, loadPreview, resetPreview, commitPreview } =
  useChallengePackageImport({
    onCommitted: async () => {
      await router.push({ name: 'ChallengeManage' })
    },
  })

async function syncPreviewByRoute(): Promise<void> {
  resetPreview()
  const id = importId.value.trim()
  if (!id) {
    return
  }
  await loadPreview(id)
}

async function backToImportPanel(): Promise<void> {
  await router.push({ name: 'PlatformChallengeImportManage' })
}

async function backToQueuePanel(): Promise<void> {
  await router.push({ name: 'PlatformChallengeImportManage', hash: '#challenge-queue-workspace' })
}

async function handleCommitPreview(): Promise<void> {
  await commitPreview()
}

onMounted(() => {
  void syncPreviewByRoute()
})

watch(importId, () => {
  void syncPreviewByRoute()
})
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
