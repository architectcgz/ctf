<template>
  <SectionCard
    class="insight-tab-section-card"
    variant="teacher-flat"
    title="复盘工作台"
    subtitle="按攻击会话查看访问、请求、提交和复盘输出。"
  >
    <StudentReviewWorkspace
      :evidence="evidence"
      :attack-sessions="attackSessions"
      :challenge-options="reviewChallengeOptions"
      :loading="reviewWorkspaceLoading"
      :query="reviewWorkspaceQuery"
      @update-filters="emit('updateReviewWorkspaceFilters', $event)"
    />
  </SectionCard>
</template>

<style scoped>
.insight-tab-section-card {
  --section-card-border-top-width: 0;
  --section-card-header-border-bottom: 0;
}
</style>

<script setup lang="ts">
import type { AttackSessionQuery, AttackSessionResponseData, StudentEvidenceData } from '@/api/contracts'
import SectionCard from '@/shared/ui/common/SectionCard.vue'
import StudentReviewWorkspace from './StudentReviewWorkspace.vue'

defineProps<{
  attackSessions: AttackSessionResponseData | null
  evidence: StudentEvidenceData | null
  reviewChallengeOptions: Array<{ value: string; label: string }>
  reviewWorkspaceLoading: boolean
  reviewWorkspaceQuery: AttackSessionQuery
}>()

const emit = defineEmits<{
  updateReviewWorkspaceFilters: [payload: Partial<AttackSessionQuery>]
}>()
</script>
