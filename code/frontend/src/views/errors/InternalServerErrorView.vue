<script setup lang="ts">
import { ArrowLeft, LayoutDashboard, ServerCrash } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="500"
    kicker="500 · Internal Server Error"
    title="系统内部错误"
    description="服务端在处理当前请求时发生异常。建议先返回上一页确认入口；如果当前链路暂时走不通，再回工作台继续其他操作。"
    accent="danger"
    :icon="ServerCrash"
    :primary-icon="ArrowLeft"
    :secondary-icon="LayoutDashboard"
    primary-action="back"
    secondary-to-home
  />
</template>
