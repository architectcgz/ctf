<script setup lang="ts">
import StudentOverviewStyleEditorial from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue'
import type { StudentOverviewProps } from './overviewProps'

defineProps<StudentOverviewProps>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()
</script>

<template>
  <StudentOverviewStyleEditorial
    v-bind="$props"
    @open-challenge="emit('openChallenge', $event)"
    @open-challenges="emit('openChallenges')"
    @open-skill-profile="emit('openSkillProfile')"
  />
</template>
