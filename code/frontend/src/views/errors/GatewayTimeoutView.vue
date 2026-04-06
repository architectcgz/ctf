<script setup lang="ts">
import { BellRing, Home, TimerOff } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="504"
    kicker="504 · Gateway Timeout"
    title="服务响应超时"
    description="请求已经发出，但依赖服务在限定时间内没有返回结果。请稍后重试，或先返回工作台。"
    accent="warning"
    :icon="TimerOff"
    :primary-icon="Home"
    :secondary-icon="BellRing"
    primary-to="/student/dashboard"
    primary-label="返回工作台"
  />
</template>
