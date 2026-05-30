<script setup lang="ts">
import type { PlatformChallengeFlagDraft } from '../model'

interface Props {
  draft: PlatformChallengeFlagDraft
}

defineProps<Props>()
</script>

<template>
  <div
    v-if="draft.isSharedInstanceChallenge"
    class="challenge-flag-panel__warning"
  >
    共享实例只适用于无状态题。该模式不提供用户级答案隔离，静态/正则答案可能被转发；若需隔离答案，请使用
    per_user 或 per_team。
  </div>

  <div
    v-if="draft.flagType === 'manual_review'"
    class="challenge-flag-panel__warning"
  >
    学生提交的答案将进入教师审核队列。审核通过后才会计分并更新通过状态。
  </div>
</template>

<style scoped>
.challenge-flag-panel__warning {
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  padding: var(--space-4);
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-ink);
}
</style>
