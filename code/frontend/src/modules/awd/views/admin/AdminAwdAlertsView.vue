<script setup lang="ts">
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps } = useAdminAwdWorkspacePage('alerts')

function severityClass(severity: string): string {
  if (severity === 'critical') return 'awd-admin-badge awd-admin-badge--critical'
  if (severity === 'warning') return 'awd-admin-badge awd-admin-badge--warning'
  return 'awd-admin-badge awd-admin-badge--info'
}
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Alerts</div>
            <h2>高优先级告警</h2>
          </div>
          <div class="awd-admin-section__hint">{{ pageModel.alerts.items.length }} 条</div>
        </div>

        <div v-if="pageModel.alerts.items.length === 0" class="awd-admin-note">
          当前没有新的告警。
        </div>

        <div v-else class="awd-admin-table">
          <article
            v-for="item in pageModel.alerts.items"
            :key="item.id"
            class="awd-admin-table__row"
          >
            <div class="awd-admin-table__main">
              <strong>{{ item.title }}</strong>
              <span class="awd-admin-table__meta">{{ item.description }}</span>
            </div>
            <div class="awd-admin-table__value">
              <span :class="severityClass(item.severity)">{{ item.time || item.severity }}</span>
            </div>
          </article>
        </div>
      </section>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
