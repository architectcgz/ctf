<script setup lang="ts">
import AwdEventTimeline from '@/modules/awd/components/AwdEventTimeline.vue'
import AdminAwdWorkspaceLayout from '@/modules/awd/layouts/AdminAwdWorkspaceLayout.vue'

import { useAdminAwdWorkspacePage } from './useAdminAwdWorkspacePage'

const { pageModel, layoutProps } = useAdminAwdWorkspacePage('replay')
</script>

<template>
  <AdminAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-admin-page">
      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Replay Timeline</div>
            <h2>整场时间线</h2>
          </div>
          <div class="awd-admin-section__hint">{{ pageModel.replay.timeline.length }} 条事件</div>
        </div>
        <AwdEventTimeline :items="pageModel.replay.timeline" empty-text="当前没有可回放的时间线。" />
      </section>

      <section class="awd-admin-section">
        <div class="awd-admin-section__head">
          <div>
            <div class="workspace-overline">Rounds</div>
            <h2>轮次索引</h2>
          </div>
        </div>

        <div v-if="pageModel.replay.rounds.length === 0" class="awd-admin-note">
          当前没有轮次索引。
        </div>

        <div v-else class="awd-admin-table">
          <article
            v-for="row in pageModel.replay.rounds"
            :key="row.id"
            class="awd-admin-table__row"
          >
            <div class="awd-admin-table__main">
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
  </AdminAwdWorkspaceLayout>
</template>

<style scoped src="./admin-awd-view.css"></style>
