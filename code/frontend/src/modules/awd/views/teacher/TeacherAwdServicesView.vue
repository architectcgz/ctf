<script setup lang="ts">
import TeacherAwdWorkspaceLayout from '@/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue'

import { useTeacherAwdWorkspacePage } from './useTeacherAwdWorkspacePage'

const { pageModel, layoutProps, selectedRoundNumber, setRound } =
  useTeacherAwdWorkspacePage('services')
</script>

<template>
  <TeacherAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-teacher-page">
      <section class="awd-teacher-section">
        <div class="awd-teacher-section__head">
          <div>
            <div class="workspace-overline">Service Filter</div>
            <h2>轮次筛选</h2>
          </div>
        </div>

        <div class="awd-teacher-round-tabs">
          <button
            type="button"
            class="awd-teacher-round-pill"
            :class="{ 'awd-teacher-round-pill--active': !selectedRoundNumber }"
            @click="setRound(undefined)"
          >
            整场总览
          </button>
          <button
            v-for="row in pageModel.replay.roundRows"
            :key="row.id"
            type="button"
            class="awd-teacher-round-pill"
            :class="{ 'awd-teacher-round-pill--active': selectedRoundNumber === row.roundNumber }"
            @click="setRound(row.roundNumber)"
          >
            第 {{ row.roundNumber }} 轮
          </button>
        </div>
      </section>

      <section class="awd-teacher-section">
        <div class="awd-teacher-section__head">
          <div>
            <div class="workspace-overline">Service Review</div>
            <h2>Service 聚合</h2>
          </div>
          <div class="awd-teacher-section__hint">{{ pageModel.services.cards.length }} 项</div>
        </div>

        <div v-if="pageModel.services.cards.length === 0" class="awd-teacher-note">
          当前轮次没有 Service 聚合数据。
        </div>

        <div v-else class="awd-teacher-card-grid">
          <article
            v-for="card in pageModel.services.cards"
            :key="card.challengeId"
            class="awd-teacher-card"
          >
            <span>{{ card.challengeTitle }}</span>
            <strong>{{ card.teamCount }} 队</strong>
            <small>
              正常 {{ card.healthyCount }} · 异常 {{ card.degradedCount }} · 被打 {{ card.attackReceived }} · 攻击分 {{ card.attackScore }}
            </small>
            <small>更新于 {{ card.updatedAt }}</small>
          </article>
        </div>
      </section>
    </div>
  </TeacherAwdWorkspaceLayout>
</template>

<style scoped src="./teacher-awd-view.css"></style>
