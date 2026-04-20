<script setup lang="ts">
import StudentAwdWorkspaceLayout from '@/modules/awd/layouts/StudentAwdWorkspaceLayout.vue'

import { useStudentAwdWorkspacePage } from './useStudentAwdWorkspacePage'

const { pageModel, layoutProps } = useStudentAwdWorkspacePage('targets')
</script>

<template>
  <StudentAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-student-page">
      <section class="awd-target-toolbar">
        <div>
          <div class="workspace-overline">Attack Surface</div>
          <h2>当前攻击题目</h2>
        </div>
        <div class="awd-target-toolbar__chips">
          <span
            v-for="option in pageModel.targets.challengeOptions"
            :key="option.key"
            class="awd-target-toolbar__chip"
            :class="{ 'awd-target-toolbar__chip--active': option.key === pageModel.targets.activeChallengeId }"
          >
            {{ option.label }}
          </span>
        </div>
      </section>

      <div
        v-if="pageModel.targets.rows.length === 0"
        class="awd-target-empty"
      >
        当前没有可用的目标目录。
      </div>

      <article
        v-for="row in pageModel.targets.rows"
        :key="row.teamId"
        class="awd-target-card"
      >
        <div class="awd-target-card__main">
          <strong>{{ row.teamName }}</strong>
          <span>{{ row.challengeTitle }}</span>
        </div>
        <div class="awd-target-card__meta">
          <span>{{ row.serviceId ? `Service #${row.serviceId}` : '未关联实例' }}</span>
          <span>{{ row.accessUrl || '当前没有可攻击地址' }}</span>
        </div>
      </article>
    </div>
  </StudentAwdWorkspaceLayout>
</template>

<style scoped>
.awd-student-page {
  display: grid;
  gap: 0.9rem;
}

.awd-target-toolbar {
  display: grid;
  gap: 0.75rem;
}

.awd-target-toolbar h2 {
  margin: 0.3rem 0 0;
  font-size: var(--font-size-18);
}

.awd-target-toolbar__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.awd-target-toolbar__chip {
  display: inline-flex;
  align-items: center;
  min-height: 2rem;
  padding: 0 0.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  font-size: var(--font-size-12);
  color: var(--color-text-secondary);
}

.awd-target-toolbar__chip--active {
  border-color: color-mix(in srgb, var(--color-primary) 34%, transparent);
  color: var(--color-primary);
}

.awd-target-card {
  display: grid;
  gap: 0.45rem;
  padding: 0.95rem 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
}

.awd-target-card__main,
.awd-target-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}

.awd-target-card__meta,
.awd-target-empty {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}
</style>
