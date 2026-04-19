<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ClassWorkspaceSectionPage from '@/components/teacher/class-management/ClassWorkspaceSectionPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useTeacherClassWorkspacePage } from '@/composables/useTeacherClassWorkspacePage'
import { useAuthStore } from '@/stores/auth'
import {
  resolveClassManagementRouteName,
  resolveClassWorkspaceSectionKeyFromRouteName,
  resolveClassWorkspaceSectionRouteName,
  resolveClassStudentsRouteName,
  resolveTeachingDashboardRouteName,
} from '@/utils/teachingWorkspaceRouting'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const reportDialogVisible = ref(false)
const {
  classes,
  review,
  summary,
  trend,
  selectedClassName,
  students,
  error,
  initialize,
} = useTeacherClassWorkspacePage()

const activeSectionKey = computed(
  () => resolveClassWorkspaceSectionKeyFromRouteName(route.name) ?? 'trend'
)

function pushClassRoute(routeName: string, className: string): void {
  if (!className) {
    return
  }

  if (className === selectedClassName.value && route.name === routeName) {
    return
  }

  const { panel: _panel, ...nextQuery } = route.query
  router.push({
    name: routeName,
    params: { className },
    query: nextQuery,
  })
}

function selectClass(className: string): void {
  pushClassRoute(
    resolveClassWorkspaceSectionRouteName(authStore.user?.role, activeSectionKey.value),
    className
  )
}

function openClassReportDialog(): void {
  reportDialogVisible.value = true
}
</script>

<template>
  <ClassWorkspaceSectionPage
    :section-key="activeSectionKey"
    :classes="classes"
    :selected-class-name="selectedClassName"
    :students="students"
    :review="review"
    :summary="summary"
    :trend="trend"
    :error="error"
    @retry="initialize"
    @open-class-overview="
      router.push({
        name: resolveClassStudentsRouteName(authStore.user?.role),
        params: { className: selectedClassName },
      })
    "
    @open-class-management="
      router.push({ name: resolveClassManagementRouteName(authStore.user?.role) })
    "
    @open-dashboard="router.push({ name: resolveTeachingDashboardRouteName(authStore.user?.role) })"
    @open-report-export="openClassReportDialog"
    @select-class="selectClass"
  />
  <TeacherClassReportExportDialog
    v-model="reportDialogVisible"
    :default-class-name="selectedClassName"
  />
</template>
