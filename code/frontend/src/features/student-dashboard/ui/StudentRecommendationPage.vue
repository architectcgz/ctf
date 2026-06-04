<template>
  <StudentRecommendationContent
    v-bind="$props"
    @open-challenge="emit('openChallenge', $event)"
    @open-challenges="emit('openChallenges')"
    @open-skill-profile="emit('openSkillProfile')"
  />
</template>

<script setup lang="ts">
import StudentRecommendationContent from './StudentRecommendationContent.vue'

defineProps<{
  weakDimensions: string[]
  recommendations: import('@/api/contracts').RecommendationItem[]
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
  openChallenges: []
  openSkillProfile: []
}>()
</script>
