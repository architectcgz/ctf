<script setup lang="ts">
import type { ClassDirectoryItem, TeacherInstanceItem } from '@/api/contracts'

import TeacherInstanceDirectorySection from './TeacherInstanceDirectorySection.vue'
import TeacherInstanceHeroPanel from './TeacherInstanceHeroPanel.vue'

defineProps<{
  classes: ClassDirectoryItem[]
  instances: TeacherInstanceItem[]
  className: string
  keyword: string
  studentNo: string
  loadingClasses: boolean
  loadingInstances: boolean
  destroyingId: string
  error: string | null
  isAdmin: boolean
  totalCount: number
  runningCount: number
  expiringSoonCount: number
  page: number
  totalPages: number
}>()

const emit = defineEmits<{
  retry: []
  openDashboard: []
  updateClassName: [value: string]
  updateKeyword: [value: string]
  updateStudentNo: [value: string]
  destroy: [id: string]
  changePage: [page: number]
}>()
</script>

<template>
  <div class="workspace-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col">
    <main class="content-pane">
      <div class="teacher-page">
        <TeacherInstanceHeroPanel
          :total-count="totalCount"
          :running-count="runningCount"
          :expiring-soon-count="expiringSoonCount"
          @open-dashboard="emit('openDashboard')"
        />

        <TeacherInstanceDirectorySection
          :classes="classes"
          :instances="instances"
          :class-name="className"
          :keyword="keyword"
          :student-no="studentNo"
          :loading-classes="loadingClasses"
          :loading-instances="loadingInstances"
          :destroying-id="destroyingId"
          :error="error"
          :is-admin="isAdmin"
          :total-count="totalCount"
          :page="page"
          :total-pages="totalPages"
          @retry="emit('retry')"
          @update-class-name="emit('updateClassName', $event)"
          @update-keyword="emit('updateKeyword', $event)"
          @update-student-no="emit('updateStudentNo', $event)"
          @destroy="emit('destroy', $event)"
          @change-page="emit('changePage', $event)"
        />
      </div>
    </main>
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --teacher-management-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --teacher-management-accent-strong: color-mix(
    in srgb,
    var(--color-primary) 74%,
    var(--journal-ink)
  );
  --teacher-management-hero-border: var(--teacher-card-border);
}

.teacher-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.teacher-badge-card {
  border: 1px solid var(--teacher-card-border);
}

.teacher-tip-block {
  border-top: 1px dashed var(--teacher-divider);
}
</style>
