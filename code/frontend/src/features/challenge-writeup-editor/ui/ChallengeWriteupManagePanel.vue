<script setup lang="ts">
import { ref, toRef } from 'vue'

import AppLoading from '@/components/common/AppLoading.vue'

import { useChallengeWriteupManagement } from '../model'

import ChallengeWriteupDirectorySection from './ChallengeWriteupDirectorySection.vue'
import ChallengeWriteupManageHeader from './ChallengeWriteupManageHeader.vue'
import ChallengeWriteupSummaryStrip from './ChallengeWriteupSummaryStrip.vue'
import './challengeWriteupManagePanel.css'

const props = defineProps<{
  challengeId: string
  challengeTitle?: string
}>()

const emit = defineEmits<{
  openWriteup: [mode: 'view' | 'edit']
}>()

const actionMenuOpen = ref(false)
const {
  loading,
  deleting,
  submissionLoading,
  submissionPage,
  submissionTotal,
  submissionTotalPages,
  officialWriteupCount,
  hasAnyWriteups,
  directoryRows,
  changeSubmissionPage,
  deleteOfficialWriteup,
} = useChallengeWriteupManagement({
  challengeId: toRef(props, 'challengeId'),
})

function openWriteup(mode: 'view' | 'edit') {
  if (!props.challengeId) return
  actionMenuOpen.value = false
  emit('openWriteup', mode)
}

function closeActionMenu() {
  actionMenuOpen.value = false
}

function setActionMenuOpen(nextOpen: boolean) {
  actionMenuOpen.value = nextOpen
}

async function handleDelete() {
  const deleted = await deleteOfficialWriteup()
  if (deleted) {
    closeActionMenu()
  }
}
</script>

<template>
  <section class="writeup-manage-panel">
    <ChallengeWriteupManageHeader :open-writeup="openWriteup" />

    <ChallengeWriteupSummaryStrip
      :official-writeup-count="officialWriteupCount"
      :submission-total="submissionTotal"
    />

    <AppLoading
      v-if="loading && submissionLoading"
      class="writeup-manage-loading"
    >
      正在加载题解内容...
    </AppLoading>

    <ChallengeWriteupDirectorySection
      v-else
      :submission-loading="submissionLoading"
      :has-any-writeups="hasAnyWriteups"
      :challenge-title="challengeTitle"
      :official-writeup-count="officialWriteupCount"
      :submission-total="submissionTotal"
      :directory-rows="directoryRows"
      :submission-page="submissionPage"
      :submission-total-pages="submissionTotalPages"
      :deleting="deleting"
      :action-menu-open="actionMenuOpen"
      :open-writeup="openWriteup"
      :set-action-menu-open="setActionMenuOpen"
      :handle-delete="handleDelete"
      :change-submission-page="changeSubmissionPage"
    />
  </section>
</template>
