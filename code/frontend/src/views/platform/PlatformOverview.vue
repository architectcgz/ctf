<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getDashboard } from '@/api/admin/platform'
import type { AdminDashboardData } from '@/api/contracts'
import PlatformOverviewPage from '@/components/platform/dashboard/PlatformOverviewPage.vue'

const router = useRouter()
const loading = ref(false)
const error = ref<string | null>(null)
const dashboard = ref<AdminDashboardData | null>(null)

async function loadDashboard(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    dashboard.value = await getDashboard()
  } catch (err) {
    console.error('加载系统概览失败:', err)
    error.value = '加载系统概览失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDashboard()
})
</script>

<template>
  <PlatformOverviewPage
    :dashboard="dashboard"
    :loading="loading"
    :error="error"
    @retry="loadDashboard"
    @open-audit-log="router.push({ name: 'AuditLog' })"
    @open-cheat-detection="router.push({ name: 'CheatDetection' })"
  />
</template>
