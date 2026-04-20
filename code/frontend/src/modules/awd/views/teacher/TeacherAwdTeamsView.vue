<script setup lang="ts">
import TeacherAwdWorkspaceLayout from '@/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue'

import { useTeacherAwdWorkspacePage } from './useTeacherAwdWorkspacePage'

const { pageModel, layoutProps, selectedRoundNumber, setRound } = useTeacherAwdWorkspacePage('teams')
</script>

<template>
  <TeacherAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-teacher-page">
      <section class="awd-teacher-section">
        <div class="awd-teacher-section__head">
          <div>
            <div class="workspace-overline">Teams</div>
            <h2>按轮次查看队伍复盘</h2>
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
            <div class="workspace-overline">Team Review</div>
            <h2>队伍表现</h2>
          </div>
          <div class="awd-teacher-section__hint">{{ pageModel.teams.rows.length }} 支队伍</div>
        </div>

        <div v-if="pageModel.teams.rows.length === 0" class="awd-teacher-note">
          当前轮次没有可展示的队伍复盘数据。
        </div>

        <div v-else class="awd-teacher-table">
          <article
            v-for="row in pageModel.teams.rows"
            :key="row.teamId"
            class="awd-teacher-table__row"
          >
            <div class="awd-teacher-table__main">
              <strong>{{ row.teamName }}</strong>
              <span class="awd-teacher-table__meta">
                成员 {{ row.memberCount }} · 异常服务 {{ row.serviceIssueCount }} · 攻防事件 {{ row.attackCount }}
              </span>
            </div>
            <div class="awd-teacher-table__value">{{ row.totalScore }} / {{ row.lastSolveAt }}</div>
          </article>
        </div>
      </section>
    </div>
  </TeacherAwdWorkspaceLayout>
</template>

<style scoped src="./teacher-awd-view.css"></style>
