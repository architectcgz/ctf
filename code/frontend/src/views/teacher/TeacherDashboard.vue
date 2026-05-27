<script setup lang="ts">
import { TeacherDashboardPage, useDashboardPage } from '@/features/teacher-dashboard'

const { overview, error, initialize, openClassManagement } = useDashboardPage()
</script>

<template>
  <TeacherDashboardPage
    :overview="overview"
    :error="error"
    @retry="initialize"
    @open-class-management="openClassManagement"
  />
</template>
