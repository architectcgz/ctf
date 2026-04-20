<script setup lang="ts">
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps } = useAdminAwdWorkspacePage('traffic')
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-metric-grid">
        <article
          v-for="card in pageModel.traffic.headline"
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
              <div class="workspace-overline">Top Challenges</div>
              <h2>热点 Service</h2>
            </div>
          </div>

          <div v-if="pageModel.traffic.topChallenges.length === 0" class="awd-admin-note">
            暂无流量热点。
          </div>

          <div v-else class="awd-admin-table">
            <article
              v-for="item in pageModel.traffic.topChallenges"
              :key="item.challengeTitle"
              class="awd-admin-table__row"
            >
              <div class="awd-admin-table__main">
                <strong>{{ item.challengeTitle }}</strong>
                <span class="awd-admin-table__meta">错误 {{ item.errorCount }}</span>
              </div>
              <div class="awd-admin-table__value">{{ item.requestCount }}</div>
            </article>
          </div>
        </section>

        <section class="awd-admin-section">
          <div class="awd-admin-section__head">
            <div>
              <div class="workspace-overline">Top Paths</div>
              <h2>热点路径</h2>
            </div>
          </div>

          <div v-if="pageModel.traffic.topPaths.length === 0" class="awd-admin-note">
            暂无路径统计。
          </div>

          <div v-else class="awd-admin-table">
            <article
              v-for="item in pageModel.traffic.topPaths"
              :key="item.path"
              class="awd-admin-table__row"
            >
              <div class="awd-admin-table__main">
                <strong>{{ item.path }}</strong>
                <span class="awd-admin-table__meta">错误 {{ item.errorCount }}</span>
              </div>
              <div class="awd-admin-table__value">{{ item.lastStatusCode }}</div>
            </article>
          </div>
        </section>
      </div>

      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Recent Events</div>
            <h2>最近流量事件</h2>
          </div>
        </div>

        <div v-if="pageModel.traffic.events.length === 0" class="awd-admin-note">
          当前没有最近流量事件。
        </div>

        <div v-else class="awd-admin-table">
          <article
            v-for="item in pageModel.traffic.events"
            :key="item.id"
            class="awd-admin-table__row"
          >
            <div class="awd-admin-table__main">
              <strong>{{ item.challengeTitle }}</strong>
              <span class="awd-admin-table__meta">
                {{ item.attackerTeam }} -> {{ item.victimTeam }} · {{ item.routeLabel }}
              </span>
            </div>
            <div class="awd-admin-table__value">{{ item.time }} / {{ item.statusLabel }}</div>
          </article>
        </div>
      </section>
    </div>
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
