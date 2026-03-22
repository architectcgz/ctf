<script setup lang="ts">
import { computed } from 'vue'

import StudentOverviewStyleCommand from '@/components/dashboard/student/StudentOverviewStyleCommand.vue'
import StudentOverviewStyleEditorial from '@/components/dashboard/student/StudentOverviewStyleEditorial.vue'
import StudentOverviewStyleSignal from '@/components/dashboard/student/StudentOverviewStyleSignal.vue'

import type { StudentOverviewProps } from './overviewProps'

const props = defineProps<StudentOverviewProps & {
  variant: '1' | '2' | '3'
}>()

const emit = defineEmits<{
  openChallenges: []
  openSkillProfile: []
  openChallenge: [challengeId: string]
}>()

const activeComponent = computed(() => {
  if (props.variant === '2') return StudentOverviewStyleEditorial
  if (props.variant === '3') return StudentOverviewStyleSignal
  return StudentOverviewStyleCommand
})

const overviewProps = computed<StudentOverviewProps>(() => ({
  displayName: props.displayName,
  className: props.className,
  progress: props.progress,
  completionRate: props.completionRate,
  highlightItems: props.highlightItems,
  recommendations: props.recommendations,
  timeline: props.timeline,
  weakDimensions: props.weakDimensions,
  skillDimensions: props.skillDimensions,
}))
</script>

<template>
  <component
    :is="activeComponent"
    v-bind="overviewProps"
    @open-challenge="emit('openChallenge', $event)"
    @open-challenges="emit('openChallenges')"
    @open-skill-profile="emit('openSkillProfile')"
  />
</template>
