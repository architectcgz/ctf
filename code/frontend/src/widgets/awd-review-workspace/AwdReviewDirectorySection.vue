<script setup lang="ts">
import { AWD_REVIEW_DIRECTORY_COLUMNS } from './model/directory'

defineProps<{
  totalCount: number
}>()
</script>

<template>
  <section
    class="workspace-directory-section teacher-directory-section"
    aria-label="AWD 赛事目录"
    :style="{ '--awd-review-directory-columns': AWD_REVIEW_DIRECTORY_COLUMNS }"
  >
    <section
      class="teacher-directory-shell workspace-directory-list workspace-directory-list--catalog"
    >
      <header class="list-heading">
        <div>
          <div class="workspace-overline">Review Directory</div>
          <h3 class="list-heading__title">赛事目录</h3>
        </div>
        <div class="teacher-directory-meta">共 {{ totalCount }} 场赛事</div>
      </header>

      <slot name="filters" />
      <slot />
    </section>
  </section>
</template>

<style scoped>
.teacher-directory-section {
  margin-top: var(--workspace-directory-page-block-gap, var(--space-5));
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.teacher-directory-meta {
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

@media (max-width: 1080px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
