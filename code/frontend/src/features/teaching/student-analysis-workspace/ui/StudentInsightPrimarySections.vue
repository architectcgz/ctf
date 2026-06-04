<template>
  <template v-if="isSectionVisible('overview')">
    <StudentInsightOverviewSection :profile="profile" :loading="loading" />
  </template>

  <StudentInsightRecommendationsSection
    v-if="isSectionVisible('recommendations')"
    :recommendations="recommendations"
    :loading="recommendationsLoading"
    @open-challenge="emit('openChallenge', $event)"
  />

  <template v-if="isSectionVisible('timeline')">
    <StudentInsightTimelineSection :timeline="timeline" :loading="loading" />
  </template>
</template>

<script setup lang="ts">
import type {
  RecommendationItem,
  SkillProfileData,
  TimelineEvent,
} from '@/api/contracts'
import type { StudentInsightSection } from '@/features/teaching/student-analysis-review'
import StudentInsightOverviewSection from './StudentInsightOverviewSection.vue'
import StudentInsightRecommendationsSection from './StudentInsightRecommendationsSection.vue'
import StudentInsightTimelineSection from './StudentInsightTimelineSection.vue'

const props = defineProps<{
  profile: SkillProfileData | null
  recommendations: RecommendationItem[]
  recommendationsLoading?: boolean
  timeline: TimelineEvent[]
  loading?: boolean
  activeSection?: StudentInsightSection
}>()

const emit = defineEmits<{
  openChallenge: [challengeId: string]
}>()

function isSectionVisible(section: Exclude<StudentInsightSection, 'all' | 'writeups' | 'manual-review' | 'evidence'>): boolean {
  return !props.activeSection || props.activeSection === 'all' || props.activeSection === section
}
</script>
