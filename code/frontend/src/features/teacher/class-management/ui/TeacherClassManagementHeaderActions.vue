<script setup lang="ts">
import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'

interface DashboardRouteTarget {
  name: string
}

defineProps<{
  dashboardRoute: DashboardRouteTarget
}>()

const emit = defineEmits<{
  openReportExport: []
}>()
</script>

<template>
  <div class="header-actions">
    <AppRouteLink :to="dashboardRoute" class="header-btn header-btn--primary">
      教学概览
    </AppRouteLink>
    <button type="button" class="header-btn header-btn--ghost" @click="emit('openReportExport')">
      导出班级报告
    </button>
  </div>
</template>
