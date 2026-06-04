<template>
  <section
    class="writeup-editor-page-shell workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col"
  >
    <ChallengeWriteupEditorContent :challenge-id="challengeId" @back="emit('back')" />
  </section>
</template>

<script setup lang="ts">
import ChallengeWriteupEditorContent from './ChallengeWriteupEditorContent.vue'
import './challengeWriteupEditorPage.css'

defineProps<{
  challengeId: string
}>()

const emit = defineEmits<{
  back: []
}>()
</script>
