<script setup lang="ts">
import { CircleDot, Gauge, Tags, Trophy } from 'lucide-vue-next'

import type { AdminChallengeListItem } from '@/api/contracts'
import ChallengeCategoryText from './ChallengeCategoryText.vue'
import ChallengeDifficultyText from './ChallengeDifficultyText.vue'

interface Props {
  challenge: Pick<AdminChallengeListItem, 'category' | 'difficulty' | 'points' | 'status'>
  statusLabel: string
}

defineProps<Props>()
</script>

<template>
  <div
    class="challenge-overview-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
  >
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>分类</span>
        <Tags class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        <ChallengeCategoryText v-if="challenge.category" :category="challenge.category" />
        <span v-else>未分类</span>
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前题目的所属方向
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>难度</span>
        <Gauge class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        <ChallengeDifficultyText v-if="challenge.difficulty" :difficulty="challenge.difficulty" />
        <span v-else>未设置</span>
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        学员侧展示的题目难度
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>分值</span>
        <Trophy class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ challenge.points }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前题目的基础得分
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>状态</span>
        <CircleDot class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ statusLabel }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前发布与维护状态
      </div>
    </article>
  </div>
</template>
