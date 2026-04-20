<script setup lang="ts">
import StudentAwdWorkspaceLayout from '@/modules/awd/layouts/StudentAwdWorkspaceLayout.vue'

import { useStudentAwdWorkspacePage } from './useStudentAwdWorkspacePage'

const { pageModel, layoutProps } = useStudentAwdWorkspacePage('attacks')
</script>

<template>
  <StudentAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-student-page">
      <div
        v-if="pageModel.attacks.submitResultMessage"
        class="awd-attack-result"
      >
        {{ pageModel.attacks.submitResultMessage }}
      </div>

      <article
        v-for="item in pageModel.attacks.recent"
        :key="item.id"
        class="awd-attack-card"
      >
        <div class="awd-attack-card__head">
          <strong>{{ item.challengeTitle }}</strong>
          <span>{{ item.time }}</span>
        </div>
        <div class="awd-attack-card__meta">
          <span>{{ item.directionLabel }}</span>
          <span>{{ item.peerTeamName }}</span>
          <span>{{ item.resultLabel }}</span>
          <span>{{ item.scoreGained }} 分</span>
          <span>{{ item.serviceRef }}</span>
        </div>
      </article>

      <div
        v-if="pageModel.attacks.recent.length === 0"
        class="awd-attack-empty"
      >
        当前没有攻击记录。
      </div>
    </div>
  </StudentAwdWorkspaceLayout>
</template>

<style scoped>
.awd-student-page {
  display: grid;
  gap: 0.9rem;
}

.awd-attack-result {
  padding: 0.9rem 1rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 24%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  font-size: var(--font-size-13);
  color: var(--color-text-primary);
}

.awd-attack-card {
  display: grid;
  gap: 0.5rem;
  padding: 0.95rem 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
}

.awd-attack-card__head,
.awd-attack-card__meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 0.75rem;
}

.awd-attack-card__meta,
.awd-attack-empty {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}
</style>
