<template>
  <div class="mx-auto max-w-xl px-4 py-16">
    <h1 class="text-2xl font-semibold">404 页面不存在</h1>
    <p class="mt-3 text-sm text-text-secondary">你访问的地址不存在，或已被移动。</p>
    <div class="mt-6">
      <RouterLink class="text-sm text-primary hover:underline" to="/dashboard">返回仪表盘</RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

