<script setup lang="ts">
interface InterventionTarget {
  id: string
  title: string
  detail: string
  meta: string[]
}

defineProps<{
  interventionTargets: InterventionTarget[]
}>()
</script>

<template>
  <section class="overview-panel workspace-directory-section teacher-directory-section">
    <header class="list-heading">
      <div>
        <h2 class="list-heading__title">优先介入学生</h2>
      </div>
    </header>

    <div
      v-if="interventionTargets.length > 0"
      class="teacher-dashboard-panel-body intervention-target-list workspace-directory-list"
    >
      <article
        v-for="item in interventionTargets"
        :key="item.id"
        class="intervention-target-row"
      >
        <div class="intervention-target-row__main">
          <h3 class="intervention-target-row__title">{{ item.title }}</h3>
          <p class="intervention-target-row__detail">{{ item.detail }}</p>
          <div class="intervention-target-row__meta">
            <span
              v-for="meta in item.meta"
              :key="meta"
              class="workspace-directory-status-pill workspace-directory-status-pill--muted"
            >
              {{ meta }}
            </span>
          </div>
        </div>
      </article>
    </div>
    <div v-else class="workspace-directory-empty portrait-empty">暂无需要优先介入的学生</div>
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
.intervention-target-list {
  display: grid;
  gap: var(--space-5);
}

.portrait-empty {
  padding: var(--space-5);
  font-size: var(--font-size-13);
  color: var(--journal-muted);
}

.intervention-target-list {
  overflow: hidden;
}

.intervention-target-row {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-4-5) var(--space-5);
  border-bottom: 1px solid var(--workspace-directory-row-divider);
}

.intervention-target-row:last-child {
  border-bottom: 0;
}

.intervention-target-row__main {
  min-width: 0;
}

.intervention-target-row__title {
  margin: 0;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--journal-ink);
}

.intervention-target-row__detail {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.intervention-target-row__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-3);
}
</style>
