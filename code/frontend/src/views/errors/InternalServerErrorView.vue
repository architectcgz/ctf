<script setup lang="ts">
import { BellRing, Home, ServerCrash } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="500"
    kicker="500 · Internal Server Error"
    title="系统内部错误"
    description="服务端在处理当前请求时发生异常。请稍后重试，或先返回工作台继续其他操作。"
    accent="danger"
    :icon="ServerCrash"
    :primary-icon="Home"
    :secondary-icon="BellRing"
    primary-to="/dashboard"
    primary-label="返回工作台"
  />
</template>
