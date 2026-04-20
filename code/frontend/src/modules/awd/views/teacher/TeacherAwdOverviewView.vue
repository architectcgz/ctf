<script setup lang="ts">
import TeacherAwdWorkspaceLayout from '@/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue'

import { useTeacherAwdWorkspacePage } from './useTeacherAwdWorkspacePage'

const { pageModel, layoutProps, selectedRoundNumber, setRound } = useTeacherAwdWorkspacePage('overview')
</script>

<template>
  <TeacherAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-teacher-page">
      <section class="awd-teacher-card-grid">
        <article
          v-for="card in pageModel.overview.cards"
          :key="card.label"
          class="awd-teacher-card"
        >
          <span>{{ card.label }}</span>
          <strong>{{ card.value }}</strong>
          <small>{{ card.helper }}</small>
        </article>
      </section>

      <section class="awd-teacher-section">
        <div class="awd-teacher-section__head">
          <div>
            <div class="workspace-overline">Rounds</div>
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
            v-for="row in pageModel.overview.roundRows"
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
            <div class="workspace-overline">Teaching Insights</div>
            <h2>教师提示</h2>
          </div>
        </div>

        <div class="awd-teacher-list">
          <article
            v-for="(insight, index) in pageModel.overview.insights"
            :key="`${index}-${insight}`"
            class="awd-teacher-list__row"
          >
            <div class="awd-teacher-list__main">
              <strong>提示 {{ index + 1 }}</strong>
              <span class="awd-teacher-note">{{ insight }}</span>
            </div>
          </article>
        </div>
      </section>
    </div>
  </TeacherAwdWorkspaceLayout>
</template>

<style scoped src="./teacher-awd-view.css"></style>
