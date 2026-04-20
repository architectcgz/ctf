<script setup lang="ts">
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps, selectedRoundId, selectRound, refreshRoundDetail } =
  useAdminAwdWorkspacePage('rounds')
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Round Control</div>
            <h2>轮次切换</h2>
          </div>
          <div class="awd-admin-section__hint">点击切换当前观察轮次</div>
        </div>

        <div class="awd-admin-round-tabs">
          <button
            v-for="row in pageModel.rounds.rows"
            :key="row.id"
            type="button"
            class="awd-admin-round-pill"
            :class="{ 'awd-admin-round-pill--active': selectedRoundId === row.id }"
            @click="selectRound(row.id)"
          >
            第 {{ row.roundNumber }} 轮
          </button>
          <button type="button" class="ui-btn ui-btn--ghost" @click="refreshRoundDetail(selectedRoundId || undefined)">
            刷新轮次
          </button>
        </div>
      </section>

      <div class="awd-admin-grid awd-admin-grid--two">
        <section class="awd-admin-section">
          <div class="awd-admin-section__head">
            <div>
              <div class="workspace-overline">Rounds</div>
              <h2>轮次列表</h2>
            </div>
          </div>

          <div class="awd-admin-list">
            <article
              v-for="row in pageModel.rounds.rows"
              :key="row.id"
              class="awd-admin-list__row"
            >
              <div class="awd-admin-list__main">
                <strong>第 {{ row.roundNumber }} 轮</strong>
                <span class="awd-admin-table__meta">
                  {{ row.statusLabel }} · 更新于 {{ row.updatedAt }}
                </span>
              </div>
              <div class="awd-admin-table__value">{{ row.scoreLabel }}</div>
            </article>
          </div>
        </section>

        <section class="awd-admin-section">
          <div class="awd-admin-section__head">
            <div>
              <div class="workspace-overline">Selected Round</div>
              <h2>当前轮队伍摘要</h2>
            </div>
          </div>

          <div v-if="pageModel.rounds.summaryRows.length === 0" class="awd-admin-note">
            当前轮次还没有队伍摘要。
          </div>

          <div v-else class="awd-admin-table">
            <article
              v-for="row in pageModel.rounds.summaryRows"
              :key="row.teamId"
              class="awd-admin-table__row"
            >
              <div class="awd-admin-table__main">
                <strong>{{ row.teamName }}</strong>
                <span class="awd-admin-table__meta">{{ row.serviceHealthLabel }}</span>
              </div>
              <div class="awd-admin-table__value">{{ row.totalScore }}</div>
            </article>
          </div>
        </section>
      </div>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
