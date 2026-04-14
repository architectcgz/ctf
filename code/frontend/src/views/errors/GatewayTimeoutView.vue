<script setup lang="ts">
import { ArrowLeft, LayoutDashboard, TimerOff } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="504"
    kicker="504 · Gateway Timeout"
    title="服务响应超时"
    description="请求已经发出，但依赖服务在限定时间内没有返回结果。建议先返回上一页稍后重试；若仍然超时，再回工作台。"
    accent="warning"
    :icon="TimerOff"
    :primary-icon="ArrowLeft"
    :secondary-icon="LayoutDashboard"
    primary-action="back"
    secondary-to-home
  />
</template>
