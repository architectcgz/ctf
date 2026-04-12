<script setup lang="ts">
import StudentOverviewPage from '@/components/dashboard/student/StudentOverviewPage.vue'
import type { StudentOverviewProps } from './overviewProps'

defineProps<StudentOverviewProps>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()
</script>

<template>
  <StudentOverviewPage
    v-bind="$props"
    @open-challenge="emit('openChallenge', $event)"
    @open-challenges="emit('openChallenges')"
    @open-skill-profile="emit('openSkillProfile')"
  />
</template>
