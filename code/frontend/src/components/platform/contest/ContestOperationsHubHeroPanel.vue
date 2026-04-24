<script setup lang="ts">
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
  <header class="list-heading contest-ops-hero workspace-directory-section">
    <div class="contest-ops-hero__main">
      <div class="workspace-overline">
        Event Operations
      </div>
      <h1 class="workspace-page-title">
        赛事运维
      </h1>
      <p class="workspace-page-copy">
        先从运维赛事目录中选择具体比赛，再进入单场运维台查看轮次、流量、大屏和榜单。
      </p>
    </div>

    <div class="contest-ops-hero__actions">
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
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
        可运维赛事
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
        进行中
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
        已冻结
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
        推荐赛事
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
