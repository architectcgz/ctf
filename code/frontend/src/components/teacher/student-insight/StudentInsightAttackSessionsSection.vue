<script setup lang="ts">
import { StudentReviewWorkspace } from '@/widgets/teacher-student-review-workspace'
import type { AttackSessionQuery } from '@/api/teacher'
import type { TeacherAttackSessionResponseData, TeacherEvidenceData } from '@/api/contracts'
import SectionCard from '@/components/common/SectionCard.vue'

defineProps<{
  attackSessions: TeacherAttackSessionResponseData | null
  evidence: TeacherEvidenceData | null
  reviewChallengeOptions: Array<{ value: string; label: string }>
  reviewWorkspaceLoading: boolean
  reviewWorkspaceQuery: AttackSessionQuery
}>()

const emit = defineEmits<{
  updateReviewWorkspaceFilters: [payload: Partial<AttackSessionQuery>]
}>()
</script>

<template>
  <SectionCard
    class="insight-tab-section-card"
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
