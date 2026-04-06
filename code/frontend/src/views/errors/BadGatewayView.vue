<script setup lang="ts">
import { BellRing, Home, Network } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="502"
    kicker="502 · Bad Gateway"
    title="上游服务响应异常"
    description="当前服务依赖的上游节点返回了无效响应。请稍后重试，或先返回工作台查看其他内容。"
    accent="danger"
    :icon="Network"
    :primary-icon="Home"
    :secondary-icon="BellRing"
    primary-to="/student/dashboard"
    primary-label="返回工作台"
  />
</template>
