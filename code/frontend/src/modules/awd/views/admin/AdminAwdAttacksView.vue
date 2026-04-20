<script setup lang="ts">
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps } = useAdminAwdWorkspacePage('attacks')
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Attack Log</div>
            <h2>近期攻击日志</h2>
          </div>
          <div class="awd-admin-section__hint">{{ pageModel.attacks.recent.length }} 条</div>
        </div>

        <div v-if="pageModel.attacks.recent.length === 0" class="awd-admin-note">
          当前轮次还没有攻击日志。
        </div>

        <div v-else class="awd-admin-table">
          <article
            v-for="item in pageModel.attacks.recent"
            :key="item.id"
            class="awd-admin-table__row"
          >
            <div class="awd-admin-table__main">
              <strong>{{ item.attackerTeam }} -> {{ item.victimTeam }}</strong>
              <span class="awd-admin-table__meta">
                {{ item.challengeTitle }} · {{ item.resultLabel }} · {{ item.sourceLabel }}
              </span>
            </div>
            <div class="awd-admin-table__value">{{ item.time }} / {{ item.scoreLabel }}</div>
          </article>
        </div>
      </section>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
