<script setup lang="ts">
import AwdEventTimeline from '@/modules/awd/components/AwdEventTimeline.vue'
import TeacherAwdWorkspaceLayout from '@/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue'

import { useTeacherAwdWorkspacePage } from './useTeacherAwdWorkspacePage'

const { pageModel, layoutProps, selectedRoundNumber, setRound } = useTeacherAwdWorkspacePage('replay')
</script>

<template>
  <TeacherAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-teacher-page">
      <section class="awd-teacher-section">
        <div class="awd-teacher-section__head">
          <div>
            <div class="workspace-overline">Replay Filter</div>
            <h2>轮次回放筛选</h2>
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
            <div class="workspace-overline">Replay Timeline</div>
            <h2>关键时间线</h2>
          </div>
        </div>
        <AwdEventTimeline :items="pageModel.replay.timeline" empty-text="当前没有可回放的关键事件。" />
      </section>
    </div>
  </TeacherAwdWorkspaceLayout>
</template>

<style scoped src="./teacher-awd-view.css"></style>
