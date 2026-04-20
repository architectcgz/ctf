<script setup lang="ts">
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps, refreshReadiness } = useAdminAwdWorkspacePage('readiness')
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-actions">
        <button type="button" class="ui-btn ui-btn--ghost" @click="refreshReadiness">
          刷新 Readiness
        </button>
      </section>

      <section class="awd-admin-metric-grid">
        <article class="awd-admin-metric-card">
          <span>通过状态</span>
          <strong>{{ pageModel.readiness.ready ? '通过' : '待处理' }}</strong>
          <small>当前阻塞 {{ pageModel.readiness.blockingCount }} 项</small>
        </article>
        <article class="awd-admin-metric-card">
          <span>阻塞动作</span>
          <strong>{{ pageModel.readiness.actions.length }}</strong>
          <small>{{ pageModel.readiness.actions.join('、') || '暂无' }}</small>
        </article>
      </section>

      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Readiness</div>
            <h2>检查项列表</h2>
          </div>
          <div class="awd-admin-section__hint">{{ pageModel.readiness.items.length }} 项</div>
        </div>

        <div v-if="pageModel.readiness.items.length === 0" class="awd-admin-note">
          当前没有待处理的 Readiness 项。
        </div>

        <div v-else class="awd-admin-table">
          <article
            v-for="item in pageModel.readiness.items"
            :key="item.id"
            class="awd-admin-table__row"
          >
            <div class="awd-admin-table__main">
              <strong>{{ item.title }}</strong>
              <span class="awd-admin-table__meta">
                {{ item.statusLabel }} · {{ item.reasonLabel }} · {{ item.accessUrl }}
              </span>
            </div>
            <div class="awd-admin-table__value">{{ item.updatedAt }}</div>
          </article>
        </div>
      </section>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
