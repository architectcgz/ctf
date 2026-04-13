<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import TeacherInstanceManagementPage from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useTeacherInstances } from '@/composables/useTeacherInstances'

const router = useRouter()

const {
  classes,
  instances,
  filters,
  loadingClasses,
  loadingInstances,
  destroyingId,
  error,
  isAdmin,
  totalCount,
  runningCount,
  expiringSoonCount,
  initialize,
  submitFilters,
  resetFilters,
  updateFilter,
  removeInstance,
} = useTeacherInstances()

async function handleDestroy(id: string): Promise<void> {
  const confirmed = await confirmDestructiveAction({
    title: '确认销毁实例',
    message: '确定要销毁该实例吗？此操作不可恢复。',
    confirmButtonText: '确认销毁',
    cancelButtonText: '取消',
  })
  if (!confirmed) {
    return
  }
  await removeInstance(id)
}

onMounted(() => {
  initialize()
})
</script>

<template>
  <TeacherInstanceManagementPage
    :classes="classes"
    :instances="instances"
    :class-name="filters.className"
    :keyword="filters.keyword"
    :student-no="filters.studentNo"
    :loading-classes="loadingClasses"
    :loading-instances="loadingInstances"
    :destroying-id="destroyingId"
    :error="error"
    :is-admin="isAdmin"
    :total-count="totalCount"
    :running-count="runningCount"
    :expiring-soon-count="expiringSoonCount"
    @retry="initialize"
    @submit="submitFilters"
    @reset="resetFilters"
    @update-class-name="updateFilter('className', $event)"
    @update-keyword="updateFilter('keyword', $event)"
    @update-student-no="updateFilter('studentNo', $event)"
    @destroy="handleDestroy"
    @open-dashboard="router.push({ name: 'TeacherDashboard' })"
  />
</template>
