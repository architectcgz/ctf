<script setup lang="ts">
import type { AppRouteTarget } from '@/shared/lib/navigation/routeTarget'
import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'

interface StudentInsightRow {
  key: string
  status: string
  title: string
  detail: string
  chips: string[]
  tone: 'ready' | 'warning' | 'danger'
  classRoute?: AppRouteTarget | null
  studentTargets?: Array<{
    id: string
    title: string
    className: string
    detail: string
    chips: string[]
    route: AppRouteTarget | null
  }>
}

defineProps<{
  studentInsightRows: StudentInsightRow[]
}>()

const emit = defineEmits<{
  openStudentList: [row: StudentInsightRow]
}>()
</script>

<template>
  <section class="overview-panel workspace-directory-section teacher-directory-section">
    <header class="list-heading">
      <div>
        <h2 class="list-heading__title">学生洞察</h2>
      </div>
    </header>

    <div class="teacher-dashboard-panel-body">
      <div class="student-insight-list workspace-directory-list">
        <article
          v-for="row in studentInsightRows"
          :key="row.key"
          class="student-insight-row"
          :class="`student-insight-row--${row.tone}`"
        >
          <div class="student-insight-row__status">
            {{ row.status }}
          </div>
          <div class="student-insight-row__main">
            <div class="student-insight-row__title-line">
              <AppRouteLink
                v-if="row.classRoute"
                :to="row.classRoute"
                class="student-insight-row__title-link"
              >
                <span class="student-insight-row__title">
                  {{ row.title }}
                </span>
              </AppRouteLink>
              <button
                v-else-if="row.studentTargets && row.studentTargets.length > 0"
                type="button"
                class="student-insight-row__title-trigger"
                :title="row.title"
                @click="emit('openStudentList', row)"
              >
                <span class="student-insight-row__title">
                  {{ row.title }}
                </span>
              </button>
              <h3 v-else class="student-insight-row__title" :title="row.title">
                {{ row.title }}
              </h3>
              <div v-if="row.chips.length > 0" class="student-insight-row__chips">
                <span
                  v-for="chip in row.chips"
                  :key="chip"
                  class="student-insight-chip workspace-directory-status-pill workspace-directory-status-pill--muted"
                >
                  {{ chip }}
                </span>
              </div>
            </div>
            <p class="student-insight-row__detail">
              {{ row.detail }}
            </p>
          </div>
        </article>
      </div>
    </div>
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

.teacher-dashboard-panel-body {
  min-width: 0;
}

.student-insight-list {
  display: grid;
}

.student-insight-row {
  display: grid;
  grid-template-columns: minmax(7rem, 0.18fr) minmax(0, 1fr);
  gap: var(--space-5);
  padding: var(--space-4-5) var(--space-5);
  border-bottom: 1px solid var(--workspace-directory-row-divider);
}

.student-insight-row:last-child {
  border-bottom: 0;
}

.student-insight-row__status {
  align-self: start;
  justify-self: start;
  display: inline-flex;
  align-items: center;
  min-height: 1.875rem;
  padding: 0 var(--space-3);
  border: 1px solid var(--teacher-card-border);
  border-radius: 999px;
  font-size: var(--font-size-12);
  font-weight: 800;
  color: var(--journal-muted);
}

.student-insight-row--ready .student-insight-row__status {
  border-color: color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
  color: color-mix(in srgb, var(--color-success) 78%, var(--journal-ink));
}

.student-insight-row--warning .student-insight-row__status {
  border-color: color-mix(in srgb, var(--color-warning) 32%, transparent);
  background: color-mix(in srgb, var(--color-warning) 9%, transparent);
  color: color-mix(in srgb, var(--color-warning) 78%, var(--journal-ink));
}

.student-insight-row--danger .student-insight-row__status {
  border-color: color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  color: color-mix(in srgb, var(--color-danger) 78%, var(--journal-ink));
}

.student-insight-row__main {
  min-width: 0;
}

.student-insight-row__title {
  margin: 0;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-16);
  font-weight: 800;
  color: var(--journal-ink);
}

.student-insight-row__title-trigger,
.student-insight-row__title-link,
.student-insight-row__title {
  flex: 1 1 16rem;
}

.student-insight-row__title-link,
.student-insight-row__title-trigger {
  min-width: 0;
}

.student-insight-row__title-link {
  text-decoration: none;
}

.student-insight-row__title-trigger {
  padding: 0;
  border: 0;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.student-insight-row__title-link:hover .student-insight-row__title,
.student-insight-row__title-link:focus-visible .student-insight-row__title,
.student-insight-row__title-trigger:hover .student-insight-row__title,
.student-insight-row__title-trigger:focus-visible .student-insight-row__title {
  color: var(--workspace-brand-ink);
  text-decoration: underline;
}

.student-insight-row__title-line {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: var(--space-3);
}

.student-insight-row__detail {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.student-insight-row__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  justify-content: flex-end;
  flex: 0 0 auto;
}

.student-insight-chip {
  background: color-mix(in srgb, var(--journal-surface) 78%, transparent);
  color: var(--journal-muted);
}

@media (max-width: 760px) {
  .student-insight-row {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }

  .student-insight-row__title-line {
    flex-direction: column;
  }

  .student-insight-row__chips {
    justify-content: flex-start;
  }
}
</style>
