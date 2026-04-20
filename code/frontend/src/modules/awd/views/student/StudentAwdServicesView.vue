<script setup lang="ts">
import StudentAwdWorkspaceLayout from '@/modules/awd/layouts/StudentAwdWorkspaceLayout.vue'

import { useStudentAwdWorkspacePage } from './useStudentAwdWorkspacePage'

const { pageModel, layoutProps, startService, startingServiceKey } = useStudentAwdWorkspacePage('services')
</script>

<template>
  <StudentAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-student-page">
      <article
        v-for="item in pageModel.services.items"
        :key="item.challengeId"
        class="awd-service-card"
      >
        <div class="awd-service-card__main">
          <div class="awd-service-card__head">
            <div>
              <h2>{{ item.challengeTitle }}</h2>
              <p>{{ item.category }} · {{ item.points }} pts</p>
            </div>
            <div class="awd-service-card__status">
              {{ item.statusLabel }}
            </div>
          </div>

          <div class="awd-service-card__meta">
            <span>{{ item.serviceId ? `Service #${item.serviceId}` : '未生成实例' }}</span>
            <span>{{ item.accessUrl || '当前还没有访问地址' }}</span>
          </div>

          <div class="awd-service-card__scores">
            <span>SLA {{ item.slaScore }}</span>
            <span>防守 {{ item.defenseScore }}</span>
            <span>攻击 {{ item.attackScore }}</span>
            <span>被打 {{ item.attackReceived }}</span>
          </div>
        </div>

        <button
          v-if="item.serviceId"
          type="button"
          class="ui-btn ui-btn--secondary"
          :disabled="startingServiceKey === item.serviceId"
          @click="startService(item.serviceId)"
        >
          {{ startingServiceKey === item.serviceId ? '启动中...' : '启动服务实例' }}
        </button>
      </article>
    </div>
  </StudentAwdWorkspaceLayout>
</template>

<style scoped>
.awd-student-page {
  display: grid;
  gap: 0.9rem;
}

.awd-service-card {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  padding: 1rem 1.1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 95%, var(--color-bg-base));
}

.awd-service-card__main {
  display: grid;
  gap: 0.65rem;
}

.awd-service-card__head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: start;
}

.awd-service-card__head h2,
.awd-service-card__head p {
  margin: 0;
}

.awd-service-card__head p,
.awd-service-card__meta,
.awd-service-card__scores {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
}

.awd-service-card__meta,
.awd-service-card__scores {
  display: flex;
  flex-wrap: wrap;
  gap: 0.9rem;
}

.awd-service-card__status {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-13);
  font-weight: 700;
}

@media (max-width: 900px) {
  .awd-service-card {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
