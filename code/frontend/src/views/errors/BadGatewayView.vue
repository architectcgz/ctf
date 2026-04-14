<script setup lang="ts">
import { ArrowLeft, LayoutDashboard, Network } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="502"
    kicker="502 · Bad Gateway"
    title="上游服务响应异常"
    description="当前服务依赖的上游节点返回了无效响应。建议先返回上一页重新进入；若问题持续，再回工作台查看其他内容。"
    accent="danger"
    :icon="Network"
    :primary-icon="ArrowLeft"
    :secondary-icon="LayoutDashboard"
    primary-action="back"
    secondary-to-home
  />
</template>
