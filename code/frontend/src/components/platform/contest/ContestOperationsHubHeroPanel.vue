<script setup lang="ts">
import { Activity, PauseCircle, Star, Trophy } from 'lucide-vue-next'

defineProps<{
  operableContestCount: number
  runningContestCount: number
  frozenContestCount: number
  preferredContestTitle: string
}>()

const emit = defineEmits<{
  back: []
}>()

function handleBack(): void {
  emit('back')
}
</script>

<template>
  <header class="workspace-page-header contest-ops-hero">
    <div class="contest-ops-hero__main">
      <div class="workspace-overline">
        Event Operations
      </div>
      <h1 class="workspace-page-title">
        赛事运维
      </h1>
    </div>

    <div class="header-actions contest-ops-hero__actions">
      <button
        type="button"
        class="header-btn header-btn--ghost"
        @click="handleBack"
      >
        返回竞赛目录
      </button>
    </div>
  </header>

  <div
    class="progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface contest-ops-summary"
  >
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>可运维赛事</span>
        <Trophy class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ operableContestCount }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前可直接进入运维台的 AWD 赛事
      </div>
    </article>
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>进行中</span>
        <Activity class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ runningContestCount }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        正在运行的赛事数量
      </div>
    </article>
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>已冻结</span>
        <PauseCircle class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ frozenContestCount }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        已进入封榜阶段的赛事数量
      </div>
    </article>
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>推荐赛事</span>
        <Star class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ preferredContestTitle }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        优先显示进行中，其次冻结中的赛事
      </div>
    </article>
  </div>
</template>

<style scoped>
.contest-ops-hero {
  align-items: flex-start;
}

.contest-ops-hero__main {
  display: grid;
  gap: var(--space-3);
  max-width: 52rem;
}

.contest-ops-hero__actions {
  padding-top: var(--space-1);
}

.contest-ops-summary {
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

@media (max-width: 860px) {
  .contest-ops-summary {
    --metric-panel-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .contest-ops-hero {
    align-items: flex-start;
  }

  .contest-ops-hero__actions {
    justify-content: flex-start;
    width: 100%;
  }

  .contest-ops-summary {
    --metric-panel-columns: 1;
  }
}
</style>
