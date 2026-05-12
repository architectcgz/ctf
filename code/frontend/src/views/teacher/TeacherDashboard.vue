<script setup lang="ts">
import TeacherDashboardPage from '@/components/teacher/dashboard/TeacherDashboardPage.vue'
import { useTeacherOverviewPage } from '@/features/teacher-dashboard'

const { overview, error, initialize, openClassManagement } = useTeacherOverviewPage()
</script>

<template>
  <TeacherDashboardPage
    :overview="overview"
    :error="error"
    @retry="initialize"
    @open-class-management="openClassManagement"
  />
</template>
