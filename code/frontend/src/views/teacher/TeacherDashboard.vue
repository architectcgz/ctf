<script setup lang="ts">
import { TeacherDashboardPage, useDashboardPage } from '@/features/teacher-dashboard'

const { overview, error, classManagementRoute, initialize } = useDashboardPage()
</script>

<template>
  <TeacherDashboardPage
    :overview="overview"
    :error="error"
    :class-management-route="classManagementRoute"
    @retry="initialize"
  />
</template>
