<template>
  <StudentDifficultyContent
    v-bind="$props"
    @open-challenges="emit('openChallenges')"
    @open-difficulty-challenges="emit('openDifficultyChallenges', $event)"
  />
</template>

<script setup lang="ts">
import StudentDifficultyContent from './StudentDifficultyContent.vue'

defineProps<{
  difficultyStats: {
    difficulty: string
    total: number
    solved: number
  }[]
}>()

const emit = defineEmits<{
  openChallenges: []
  openDifficultyChallenges: [difficulty: string]
}>()
</script>
