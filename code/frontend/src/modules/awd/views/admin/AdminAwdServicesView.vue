<script setup lang="ts">
import { computed } from 'vue'

import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps } = useAdminAwdWorkspacePage('services')

function statusClass(status: string): string {
  if (status === 'compromised') return 'awd-admin-badge awd-admin-badge--critical'
  if (status === 'down') return 'awd-admin-badge awd-admin-badge--warning'
  return 'awd-admin-badge awd-admin-badge--info'
}

const items = computed(() => pageModel.value.services.items)
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Service Matrix</div>
            <h2>队伍与服务运行状态</h2>
          </div>
          <div class="awd-admin-section__hint">{{ items.length }} 行</div>
        </div>

        <div v-if="items.length === 0" class="awd-admin-note">当前轮次还没有服务矩阵数据。</div>

        <div v-else class="awd-admin-table">
          <article v-for="item in items" :key="item.id" class="awd-admin-table__row">
            <div class="awd-admin-table__main">
              <strong>{{ item.teamName }} / {{ item.challengeTitle }}</strong>
              <span class="awd-admin-table__meta">
                {{ item.serviceLabel }} · 被打 {{ item.attackReceived }} 次 · SLA {{ item.slaScore }}
              </span>
            </div>
            <div class="awd-admin-table__value">
              <span :class="statusClass(item.status)">{{ item.statusLabel }}</span>
            </div>
          </article>
        </div>
      </section>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
