<script setup lang="ts">
import { Home, RotateCcw, TimerOff } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="504"
    kicker="504 · Gateway Timeout"
    title="服务响应超时"
    description="请求已经发出，但依赖服务在限定时间内没有返回结果。你可以先刷新当前页面重试，若仍然超时，再返回工作台。"
    accent="warning"
    :icon="TimerOff"
    :primary-icon="RotateCcw"
    :secondary-icon="Home"
    primary-action="reload"
    secondary-to-home
  />
</template>
