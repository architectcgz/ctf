<script setup lang="ts">
interface ReviewHighlight {
  key: string
  title: string
  detail: string
  chips: string[]
  tone: 'ready' | 'warning' | 'danger'
}

defineProps<{
  reviewHighlights: ReviewHighlight[]
}>()
</script>

<template>
  <section class="overview-panel workspace-directory-section teacher-directory-section">
    <header class="list-heading">
      <div>
        <h2 class="list-heading__title">教学复盘摘要</h2>
      </div>
    </header>

    <div
      v-if="reviewHighlights.length > 0"
      class="teacher-dashboard-panel-body review-highlight-list workspace-directory-list"
    >
      <article
        v-for="item in reviewHighlights"
        :key="item.key"
        class="review-highlight-item"
        :class="`review-highlight-item--${item.tone}`"
      >
        <div class="review-highlight-item__main">
          <h3 class="review-highlight-item__title">{{ item.title }}</h3>
          <p class="review-highlight-item__detail">{{ item.detail }}</p>
          <div class="review-highlight-item__chips">
            <span
              v-for="chip in item.chips"
              :key="chip"
              class="workspace-directory-status-pill workspace-directory-status-pill--muted"
            >
              {{ chip }}
            </span>
          </div>
        </div>
      </article>
    </div>
    <div v-else class="workspace-directory-empty portrait-empty">暂无可展示的复盘摘要</div>
  </section>
</template>

<style scoped>
.overview-panel {
  --workspace-directory-section-padding: 0;
  --workspace-directory-section-gap: var(--space-5);
  --workspace-directory-shell-radius: 16px;
  --workspace-directory-shell-padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
}

.overview-panel > .list-heading {
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.teacher-dashboard-panel-body,
.review-highlight-list {
  display: grid;
  gap: var(--space-5);
}

.portrait-empty {
  padding: var(--space-5);
  font-size: var(--font-size-13);
  color: var(--journal-muted);
}

.review-highlight-list {
  overflow: hidden;
}

.review-highlight-item {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-4-5) var(--space-5);
  border-bottom: 1px solid var(--workspace-directory-row-divider);
  border-left: 3px solid transparent;
}

.review-highlight-item:last-child {
  border-bottom: 0;
}

.review-highlight-item--ready {
  border-left-color: color-mix(in srgb, var(--color-success) 42%, transparent);
}

.review-highlight-item--warning {
  border-left-color: color-mix(in srgb, var(--color-warning) 42%, transparent);
}

.review-highlight-item--danger {
  border-left-color: color-mix(in srgb, var(--color-danger) 42%, transparent);
}

.review-highlight-item__main {
  min-width: 0;
}

.review-highlight-item__title {
  margin: 0;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--journal-ink);
}

.review-highlight-item__detail {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.review-highlight-item__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-3);
}
</style>
