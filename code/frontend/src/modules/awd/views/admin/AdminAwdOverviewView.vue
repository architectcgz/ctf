<script setup lang="ts">
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps, refreshAll } = useAdminAwdWorkspacePage('overview')

function handleRefreshAll(): void {
  void refreshAll()
}
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-actions">
        <button type="button" class="ui-btn ui-btn--ghost" @click="handleRefreshAll">刷新数据</button>
      </section>

      <section class="awd-admin-metric-grid">
        <article
          v-for="card in pageModel.overview.cards"
          :key="card.label"
          class="awd-admin-metric-card"
        >
          <span>{{ card.label }}</span>
          <strong>{{ card.value }}</strong>
          <small>{{ card.helper }}</small>
        </article>
      </section>

      <div class="awd-admin-grid awd-admin-grid--two">
        <section class="awd-admin-section">
          <div class="awd-admin-section__head">
            <div>
              <div class="workspace-overline">Scoreboard</div>
              <h2>真实榜单</h2>
            </div>
            <div class="awd-admin-section__hint">
              {{ pageModel.overview.scoreboardRows.length }} 支队伍
            </div>
          </div>

          <div v-if="pageModel.overview.scoreboardRows.length === 0" class="awd-admin-note">
            当前还没有可展示的真实榜单。
          </div>

          <div v-else class="awd-admin-table">
            <article
              v-for="row in pageModel.overview.scoreboardRows"
              :key="row.team_id"
              class="awd-admin-table__row"
            >
              <div class="awd-admin-table__main">
                <strong>#{{ row.rank }} {{ row.team_name }}</strong>
                <span class="awd-admin-table__meta">最近提交 {{ row.last_submission_at || '暂无' }}</span>
              </div>
              <div class="awd-admin-table__value">{{ row.score }}</div>
            </article>
          </div>
        </section>

        <section class="awd-admin-section">
          <div class="awd-admin-section__head">
            <div>
              <div class="workspace-overline">Rounds</div>
              <h2>轮次摘要</h2>
            </div>
            <div class="awd-admin-section__hint">
              {{ pageModel.overview.roundRows.length }} 轮
            </div>
          </div>

          <div v-if="pageModel.overview.roundRows.length === 0" class="awd-admin-note">
            还没有轮次信息。
          </div>

          <div v-else class="awd-admin-list">
            <article
              v-for="row in pageModel.overview.roundRows"
              :key="row.id"
              class="awd-admin-list__row"
            >
              <div class="awd-admin-list__main">
                <strong>第 {{ row.roundNumber }} 轮</strong>
                <span class="awd-admin-table__meta">
                  {{ row.statusLabel }} · 服务 {{ row.serviceCount }} · 攻击 {{ row.attackCount }}
                </span>
              </div>
              <div class="awd-admin-table__value">{{ row.scoreLabel }}</div>
            </article>
          </div>
        </section>
      </div>

      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Teams</div>
            <h2>队伍轮次表现</h2>
          </div>
          <div class="awd-admin-section__hint">
            {{ pageModel.overview.summaryRows.length }} 行摘要
          </div>
        </div>

        <div v-if="pageModel.overview.summaryRows.length === 0" class="awd-admin-note">
          当前轮次还没有队伍摘要。
        </div>

        <div v-else class="awd-admin-table">
          <article
            v-for="row in pageModel.overview.summaryRows"
            :key="row.teamId"
            class="awd-admin-table__row"
          >
            <div class="awd-admin-table__main">
              <strong>{{ row.teamName }}</strong>
              <span class="awd-admin-table__meta">{{ row.serviceHealthLabel }}</span>
            </div>
            <div class="awd-admin-table__value">
              {{ row.totalScore }} / {{ row.attackScore }} / {{ row.defenseScore }}
            </div>
          </article>
        </div>
      </section>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
