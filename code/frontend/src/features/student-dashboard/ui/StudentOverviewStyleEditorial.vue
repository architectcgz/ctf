<template>
  <StudentOverviewContent
    v-bind="$props"
    @open-challenge="emit('openChallenge', $event)"
    @open-challenges="emit('openChallenges')"
    @open-skill-profile="emit('openSkillProfile')"
  />
</template>

<script setup lang="ts">
import StudentOverviewContent from './StudentOverviewContent.vue'
import type { StudentOverviewProps } from './studentOverviewProps'

defineProps<StudentOverviewProps>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()
</script>
