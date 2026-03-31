<script setup lang="ts">
import { BellRing, Gauge, RotateCcw } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="429"
    kicker="429 · Too Many Requests"
    title="请求过于频繁"
    description="系统在短时间内收到了过多请求。请稍后再试，或先返回工作台继续其他操作。"
    accent="warning"
    :icon="Gauge"
    :primary-icon="RotateCcw"
    :secondary-icon="BellRing"
    primary-to="/dashboard"
    primary-label="返回工作台"
  />
</template>
