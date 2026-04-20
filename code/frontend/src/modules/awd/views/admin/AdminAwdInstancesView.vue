<script setup lang="ts">
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps } = useAdminAwdWorkspacePage('instances')
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Instances</div>
            <h2>队伍实例健康</h2>
          </div>
          <div class="awd-admin-section__hint">{{ pageModel.instances.rows.length }} 支队伍</div>
        </div>

        <div v-if="pageModel.instances.rows.length === 0" class="awd-admin-note">
          当前没有可展示的实例摘要。
        </div>

        <div v-else class="awd-admin-table">
          <article
            v-for="row in pageModel.instances.rows"
            :key="row.teamId"
            class="awd-admin-table__row"
          >
            <div class="awd-admin-table__main">
              <strong>{{ row.teamName }}</strong>
              <span class="awd-admin-table__meta">
                总服务 {{ row.totalServices }} · 正常 {{ row.upCount }} · 离线 {{ row.downCount }} · 失陷 {{ row.compromisedCount }}
              </span>
            </div>
            <div class="awd-admin-table__value">{{ row.latestUpdatedAt }}</div>
          </article>
        </div>
      </section>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
