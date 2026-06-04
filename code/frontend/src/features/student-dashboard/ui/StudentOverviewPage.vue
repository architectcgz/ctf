<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <div class="content-pane">
      <StudentOverviewStyleEditorial
        v-bind="$props"
        @open-challenge="emit('openChallenge', $event)"
        @open-challenges="emit('openChallenges')"
        @open-skill-profile="emit('openSkillProfile')"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import StudentOverviewStyleEditorial from './StudentOverviewStyleEditorial.vue'
import type { StudentOverviewProps } from './studentOverviewProps'

defineProps<StudentOverviewProps>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()
</script>
