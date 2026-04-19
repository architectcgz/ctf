<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import TeacherClassStudents from '@/views/teacher/TeacherClassStudents.vue'

const route = useRoute()
const router = useRouter()

const panelByRouteName = {
  TeacherClassTrend: 'trend',
  TeacherClassReview: 'review',
  TeacherClassInsights: 'insight',
  TeacherClassIntervention: 'action',
} as const

const targetPanel = computed(() => {
  const routeName = route.name as keyof typeof panelByRouteName | undefined
  return routeName ? panelByRouteName[routeName] : null
})

async function redirectToCanonicalWorkspace(): Promise<void> {
  if (!targetPanel.value) return

  await router.replace({
    name: 'TeacherClassStudents',
    params: {
      className: route.params.className,
    },
    query: {
      ...route.query,
      panel: targetPanel.value,
    },
  })
}

watch(
  () => [route.name, route.params.className, route.query.panel] as const,
  () => {
    void redirectToCanonicalWorkspace()
  },
  { immediate: true }
)
</script>

<template>
  <TeacherClassStudents />
</template>
