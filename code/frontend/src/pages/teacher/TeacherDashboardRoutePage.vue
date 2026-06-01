<script setup lang="ts">
import { TeacherDashboardPage, useDashboardPage } from '@/features/teacher/dashboard'

const { overview, error, classManagementRoute, activePanel, switchPanel, initialize } =
  useDashboardPage()
</script>

<template>
  <TeacherDashboardPage
    :overview="overview"
    :error="error"
    :class-management-route="classManagementRoute"
    :active-panel="activePanel"
    @retry="initialize"
    @switch-panel="switchPanel"
  />
</template>
