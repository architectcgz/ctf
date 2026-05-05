<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { ArrowLeft, FolderTree, RefreshCw, ShieldCheck } from 'lucide-vue-next'

import AWDDefenseFileWorkbench from '@/components/contests/awd/AWDDefenseFileWorkbench.vue'
import { useContestAwdDefenseWorkbenchPage } from '@/features/contest-awd-workspace'

const {
  backLink,
  currentDirectoryPath,
  directory,
  error,
  file,
  fileError,
  fileLoading,
  loading,
  openFile,
  loadDirectory,
  refreshDirectory,
  serviceCard,
  serviceTitle,
} = useContestAwdDefenseWorkbenchPage()
</script>

<template>
  <section class="workspace-shell journal-shell journal-shell-user awd-defense-page">
    <RouterLink class="defense-back-link" :to="backLink">
      <ArrowLeft class="defense-icon" />
      <span>返回战场</span>
    </RouterLink>

    <header class="defense-hero">
      <div class="defense-hero__main">
        <div class="defense-hero__eyebrow">防守内容</div>
        <h1>{{ serviceTitle }}</h1>
        <div class="defense-hero__meta">
          <span class="defense-chip">
            <ShieldCheck class="h-3.5 w-3.5" />
            <span>{{ serviceCard?.serviceStatusLabel || '待确认' }}</span>
          </span>
          <span class="defense-chip defense-chip--path">
            <FolderTree class="h-3.5 w-3.5" />
            <span class="font-mono">{{ currentDirectoryPath }}</span>
          </span>
        </div>
      </div>
      <button class="defense-refresh" type="button" :disabled="loading" @click="refreshDirectory">
        <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
        <span>刷新当前目录</span>
      </button>
    </header>

    <AWDDefenseFileWorkbench
      :service-title="serviceTitle"
      :directory="directory"
      :file="file"
      :loading="loading"
      :file-loading="fileLoading"
      :error="error"
      :file-error="fileError"
      @refresh="refreshDirectory"
      @open-directory="loadDirectory"
      @open-file="openFile"
    />
  </section>
</template>

<style scoped>
.awd-defense-page {
  min-height: 100%;
  display: grid;
  align-content: start;
  gap: var(--space-6);
  padding: var(--space-6);
  background:
    radial-gradient(circle at top left, color-mix(in srgb, var(--color-primary) 14%, transparent), transparent 28rem),
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-surface) 92%, black), var(--color-bg-page));
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
  background: color-mix(in srgb, var(--color-bg-elevated) 90%, transparent);
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

.defense-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-5);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  border-radius: var(--radius-xl);
  background: color-mix(in srgb, var(--color-bg-elevated) 86%, transparent);
}

.defense-hero__eyebrow {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.defense-hero h1 {
  margin: var(--space-2) 0 0;
  color: var(--color-text-primary);
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
}

.defense-hero__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  margin-top: var(--space-3);
}

.defense-chip,
.defense-refresh {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  font-size: var(--font-size-12);
  font-weight: 700;
}

.defense-chip {
  background: color-mix(in srgb, var(--color-bg-page) 78%, transparent);
  color: var(--color-text-secondary);
  padding: var(--space-2) var(--space-3);
}

.defense-chip--path {
  max-width: 100%;
}

.defense-refresh {
  min-height: var(--control-height-sm);
  background: color-mix(in srgb, var(--color-bg-page) 68%, transparent);
  color: var(--color-text-primary);
  padding: 0 var(--space-3);
}

.defense-refresh:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 899px) {
  .awd-defense-page {
    padding: var(--space-4);
    gap: var(--space-4);
  }

  .defense-hero {
    flex-direction: column;
  }

  .defense-refresh {
    width: 100%;
    justify-content: center;
    min-height: var(--control-height-md);
  }
}
</style>
