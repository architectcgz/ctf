<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { ArrowLeft, Shield } from 'lucide-vue-next'

const route = useRoute()
const contestId = computed(() => String(route.params.id ?? ''))
</script>

<template>
  <section class="workspace-shell journal-shell journal-shell-user awd-defense-page">
    <RouterLink
      class="defense-back-link"
      :to="{ name: 'ContestDetail', params: { id: contestId }, query: { panel: 'challenges' } }"
    >
      <ArrowLeft class="defense-icon" />
      <span>返回战场</span>
    </RouterLink>

    <div class="defense-disabled">
      <Shield class="defense-disabled__icon" />
      <h1>防守入口已迁移</h1>
      <p>浏览器文件工作台不可用，请在战场服务卡片中生成 SSH 连接后使用本地编辑器连接。</p>
    </div>
  </section>
</template>

<style scoped>
.awd-defense-page {
  min-height: 100%;
  display: grid;
  align-content: start;
  gap: var(--space-8);
  padding: var(--space-6);
  background: var(--color-bg-surface);
}

.defense-back-link {
  width: fit-content;
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: var(--control-height-sm);
  padding: 0 var(--space-3);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
  font-weight: 700;
}

.defense-back-link:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.defense-icon {
  width: 1rem;
  height: 1rem;
}

.defense-disabled {
  max-width: 42rem;
  display: grid;
  gap: var(--space-3);
}

.defense-disabled__icon {
  width: 2rem;
  height: 2rem;
  color: var(--color-primary);
}

.defense-disabled h1 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
}

.defense-disabled p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-15);
  line-height: 1.7;
}
</style>
