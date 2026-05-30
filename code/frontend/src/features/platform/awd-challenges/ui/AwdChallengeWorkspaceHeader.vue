<template>
  <header class="workspace-page-header">
    <div class="workspace-tab-heading__main">
      <div class="workspace-overline">AWD Service Authoring</div>
      <h1 class="hero-title">
        {{ mode === 'import' ? '导入 AWD 题目包' : 'AWD 题目库' }}
      </h1>
      <p class="hero-summary">
        {{
          mode === 'import'
            ? '上传符合规范的 AWD 题目包，确认后生成可用于编排的 AWD 题目。'
            : '管理 AWD 赛事使用的题目。'
        }}
      </p>

      <div v-if="mode === 'import'" class="awd-import-page-note">
        上传题目包并确认导入后，系统会生成可用于 AWD 编排的题目。
      </div>
    </div>

    <div class="awd-library-hero-actions">
      <div class="header-actions quick-actions">
        <button
          v-if="mode === 'library'"
          type="button"
          class="header-btn header-btn--ghost"
          @click="emit('refresh')"
        >
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <AppRouteLink
          v-if="mode === 'library'"
          id="awd-challenge-open-import"
          :to="importRoute || { name: 'PlatformAwdChallengeImport' }"
          class="header-btn header-btn--primary"
        >
          <Upload class="h-4 w-4" />
          导入题目包
        </AppRouteLink>
        <button
          v-if="mode === 'import'"
          type="button"
          class="header-btn header-btn--ghost"
          @click="emit('refreshImportQueue')"
        >
          <RefreshCw class="h-4 w-4" />
          刷新队列
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--journal-ink);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
}

.awd-library-hero-actions {
  display: flex;
  align-items: flex-end;
  padding-bottom: 0.5rem;
}

.quick-actions {
  gap: var(--space-3);
}

.awd-import-page-note {
  max-width: 46rem;
  margin-top: var(--space-4);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}
</style>

<script setup lang="ts">
import { RefreshCw, Upload } from 'lucide-vue-next'

import AppRouteLink from '@/components/navigation/AppRouteLink.vue'
import type { AppRouteTarget } from '@/components/navigation/routeTarget'

defineProps<{
  mode: 'library' | 'import'
  importRoute?: AppRouteTarget | null
}>()

const emit = defineEmits<{
  refresh: []
  refreshImportQueue: []
}>()
</script>
