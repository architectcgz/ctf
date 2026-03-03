<template>
  <div class="mx-auto max-w-xl px-4 py-16">
    <h1 class="text-2xl font-semibold">403 无权限</h1>
    <p class="mt-3 text-sm text-text-secondary">你没有访问该页面的权限。</p>
    <div class="mt-6">
      <RouterLink class="text-sm text-primary hover:underline" to="/dashboard">返回仪表盘</RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

