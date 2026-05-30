<script setup lang="ts">
import { TeacherDashboardPage } from '@/features/teacher/dashboard'
import { useDashboardPage } from '@/features/teacher/dashboard/model/useDashboardPage'

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
