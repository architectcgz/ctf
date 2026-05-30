<script setup lang="ts">
import { ChallengeCategoryPill, ChallengeDifficultyText } from '@/entities/challenge'
import type { AdminChallengeListItem } from '@/api/contracts'

defineProps<{
  challenge: AdminChallengeListItem | null
}>()
</script>

<template>
  <div class="writeup-rail-card">
    <div class="journal-note-label">
      Challenge
    </div>
    <h2 class="writeup-rail-title">
      题目信息
    </h2>

    <dl
      v-if="challenge"
      class="writeup-rail-meta"
    >
      <div>
        <dt>标题</dt>
        <dd>{{ challenge.title }}</dd>
      </div>
      <div>
        <dt>分类</dt>
        <dd>
          <ChallengeCategoryPill :category="challenge.category" />
        </dd>
      </div>
      <div>
        <dt>状态</dt>
        <dd>{{ challenge.status }}</dd>
      </div>
      <div>
        <dt>难度</dt>
        <dd>
          <ChallengeDifficultyText :difficulty="challenge.difficulty" />
        </dd>
      </div>
      <div>
        <dt>分值</dt>
        <dd>{{ challenge.points }}</dd>
      </div>
    </dl>
    <div
      v-else
      class="writeup-rail-empty"
    >
      题目信息加载中。
    </div>
  </div>
</template>
