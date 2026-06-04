<template>
  <StudentCategoryProgressContent
    v-bind="$props"
    @open-category-challenges="emit('openCategoryChallenges', $event)"
    @open-challenges="emit('openChallenges')"
    @open-skill-profile="emit('openSkillProfile')"
  />
</template>

<script setup lang="ts">
import StudentCategoryProgressContent from './StudentCategoryProgressContent.vue'

defineProps<{
  categoryStats: {
    category: string
    total: number
    solved: number
  }[]
  completionRate: number
}>()

const emit = defineEmits<{
  openChallenges: []
  openCategoryChallenges: [category: string]
  openSkillProfile: []
}>()
</script>
