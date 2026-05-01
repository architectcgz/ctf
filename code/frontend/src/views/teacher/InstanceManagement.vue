<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import TeacherInstanceManagementPage from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useTeacherInstances } from '@/features/teacher-instances'
import { useAuthStore } from '@/stores/auth'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { resolveTeachingDashboardRouteName } from '@/utils/teachingWorkspaceRouting'

const router = useRouter()
const authStore = useAuthStore()
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)

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
  updateFilter,
  removeInstance,
} = useTeacherInstances()

const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalCount.value / Math.max(pageSize.value, 1)))
)
const paginatedInstances = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return instances.value.slice(start, start + pageSize.value)
})

function handlePageChange(nextPage: number): void {
  const normalizedPage = Math.max(1, Math.floor(nextPage))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }
  page.value = normalizedPage
}

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

watch(
  () => [filters.className, filters.keyword, filters.studentNo],
  () => {
    page.value = 1
  }
)

watch(totalCount, () => {
  if (page.value > totalPages.value) {
    page.value = totalPages.value
  }
})
</script>

<template>
  <TeacherInstanceManagementPage
    :classes="classes"
    :class-name="filters.className"
    :keyword="filters.keyword"
    :student-no="filters.studentNo"
    :page="page"
    :total-pages="totalPages"
    :loading-classes="loadingClasses"
    :loading-instances="loadingInstances"
    :destroying-id="destroyingId"
    :error="error"
    :is-admin="isAdmin"
    :total-count="totalCount"
    :running-count="runningCount"
    :expiring-soon-count="expiringSoonCount"
    :instances="paginatedInstances"
    @retry="initialize"
    @update-class-name="updateFilter('className', $event)"
    @update-keyword="updateFilter('keyword', $event)"
    @update-student-no="updateFilter('studentNo', $event)"
    @change-page="handlePageChange"
    @destroy="handleDestroy"
    @open-dashboard="router.push({ name: resolveTeachingDashboardRouteName(authStore.user?.role) })"
  />
</template>
