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
    <section class="teacher-directory-shell workspace-directory-list">
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

.teacher-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 18px 34px color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
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
