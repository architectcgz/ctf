<template>
  <SectionCard
    class="insight-tab-section-card"
    variant="teacher-flat"
    title="复盘工作台"
    subtitle="按攻击会话查看访问、请求、提交和复盘输出。"
  >
    <div v-if="reviewWorkspaceLoading && !attackSessions" class="evidence-glass">
      <div class="evidence-glass__filters">
        <span class="evidence-glass__filter" />
        <span class="evidence-glass__filter" />
        <span class="evidence-glass__filter" />
      </div>
      <div class="evidence-glass__summary">
        <span
          v-for="index in 4"
          :key="index"
          class="evidence-glass__summary-card"
        />
      </div>
      <div class="evidence-glass__sessions">
        <div
          v-for="index in 3"
          :key="index"
          class="evidence-glass__session"
        >
          <span class="evidence-glass__session-icon" />
          <span class="evidence-glass__session-text" />
          <span class="evidence-glass__session-meta" />
        </div>
      </div>
    </div>

    <StudentReviewWorkspace
      v-else
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

/* ── Evidence glass skeleton ── */

.evidence-glass {
  position: relative;
  overflow: hidden;
  margin-top: var(--space-5);
  border: 1px solid color-mix(in srgb, var(--teacher-card-border) 88%, transparent);
  border-radius: var(--workspace-radius-lg);
  padding: var(--space-4);
  background:
    radial-gradient(
      ellipse at top right,
      color-mix(in srgb, var(--journal-accent) 9%, transparent),
      transparent 46%
    ),
    radial-gradient(
      ellipse at bottom left,
      color-mix(in srgb, var(--color-bg-surface) 58%, transparent),
      transparent 52%
    ),
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base))
    );
  box-shadow: var(--workspace-shadow-panel);
  display: grid;
  gap: var(--space-4);
}

.evidence-glass::before {
  position: absolute;
  inset: 1px;
  pointer-events: none;
  content: '';
  border-radius: calc(var(--workspace-radius-lg) - 1px);
  background:
    linear-gradient(
      115deg,
      transparent 0%,
      color-mix(in srgb, var(--color-bg-surface) 34%, transparent) 38%,
      transparent 72%
    );
  opacity: 0.54;
}

.evidence-glass__filters {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.evidence-glass__filter {
  display: block;
  height: var(--space-10);
  overflow: hidden;
  border-radius: var(--workspace-radius-lg);
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--teacher-divider) 66%, transparent) 0%,
      color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface)) 42%,
      color-mix(in srgb, var(--teacher-divider) 58%, transparent) 76%
    );
  background-size: 220% 100%;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 0 color-mix(in srgb, var(--teacher-divider) 48%, transparent);
  animation: evidenceGlassSkeletonSweep 1.55s ease-in-out infinite;
}

.evidence-glass__summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.evidence-glass__summary-card {
  display: block;
  height: var(--space-16);
  overflow: hidden;
  border-radius: var(--workspace-radius-lg);
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--teacher-divider) 66%, transparent) 0%,
      color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface)) 42%,
      color-mix(in srgb, var(--teacher-divider) 58%, transparent) 76%
    );
  background-size: 220% 100%;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 0 color-mix(in srgb, var(--teacher-divider) 48%, transparent);
  animation: evidenceGlassSkeletonSweep 1.55s ease-in-out infinite;
}

.evidence-glass__sessions {
  display: grid;
  gap: var(--space-2);
  padding-top: var(--space-2);
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 68%, transparent);
}

.evidence-glass__session {
  display: grid;
  grid-template-columns: var(--space-10) minmax(0, 1fr) var(--space-18);
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) 0;
}

.evidence-glass__session-icon {
  display: block;
  width: var(--space-10);
  height: var(--space-10);
  overflow: hidden;
  border-radius: var(--workspace-radius-lg);
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--teacher-divider) 66%, transparent) 0%,
      color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface)) 42%,
      color-mix(in srgb, var(--teacher-divider) 58%, transparent) 76%
    );
  background-size: 220% 100%;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 0 color-mix(in srgb, var(--teacher-divider) 48%, transparent);
  animation: evidenceGlassSkeletonSweep 1.55s ease-in-out infinite;
}

.evidence-glass__session-text {
  display: block;
  width: min(28rem, 76%);
  height: var(--space-3);
  overflow: hidden;
  border-radius: 999px;
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--teacher-divider) 66%, transparent) 0%,
      color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface)) 42%,
      color-mix(in srgb, var(--teacher-divider) 58%, transparent) 76%
    );
  background-size: 220% 100%;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 0 color-mix(in srgb, var(--teacher-divider) 48%, transparent);
  animation: evidenceGlassSkeletonSweep 1.55s ease-in-out infinite;
}

.evidence-glass__session-meta {
  display: block;
  width: var(--space-18);
  height: var(--space-3);
  overflow: hidden;
  border-radius: 999px;
  background:
    linear-gradient(
      100deg,
      color-mix(in srgb, var(--teacher-divider) 66%, transparent) 0%,
      color-mix(in srgb, var(--journal-accent) 16%, var(--journal-surface)) 42%,
      color-mix(in srgb, var(--teacher-divider) 58%, transparent) 76%
    );
  background-size: 220% 100%;
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--color-bg-surface) 68%, transparent),
    0 1px 0 color-mix(in srgb, var(--teacher-divider) 48%, transparent);
  animation: evidenceGlassSkeletonSweep 1.55s ease-in-out infinite;
}

@media (max-width: 767px) {
  .evidence-glass__filters,
  .evidence-glass__summary {
    grid-template-columns: 1fr;
  }

  .evidence-glass__session-meta {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .evidence-glass__filter,
  .evidence-glass__summary-card,
  .evidence-glass__session-icon,
  .evidence-glass__session-text,
  .evidence-glass__session-meta {
    animation: none;
  }
}

@keyframes evidenceGlassSkeletonSweep {
  0% {
    background-position: 120% 0;
  }

  100% {
    background-position: -120% 0;
  }
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
