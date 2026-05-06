<script setup lang="ts">
import { computed } from 'vue'
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
  saveError,
  saveFile,
  saveLoading,
  serviceCard,
  serviceTitle,
} = useContestAwdDefenseWorkbenchPage()

const editablePaths = computed(() => serviceCard.value?.defenseScope?.editable_paths || [])
</script>

<template>
  <section class="workspace-shell journal-shell journal-shell-user awd-defense-page">
    <header class="defense-hero">
      <div class="defense-hero__main">
        <div class="defense-hero__nav">
          <RouterLink class="defense-hero__back ui-btn ui-btn--sm ui-btn--ghost" :to="backLink">
            <ArrowLeft class="defense-icon" />
            <span>返回战场</span>
          </RouterLink>
          <div class="defense-hero__eyebrow">防守工作台</div>
        </div>
        <h1>{{ serviceTitle }}</h1>
      </div>
      <div class="defense-hero__side">
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
        <button
          class="defense-refresh ui-btn ui-btn--sm ui-btn--secondary"
          type="button"
          :disabled="loading"
          @click="refreshDirectory"
        >
          <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
          <span>刷新当前目录</span>
        </button>
      </div>
    </header>

    <AWDDefenseFileWorkbench
      :service-title="serviceTitle"
      :directory="directory"
      :file="file"
      :loading="loading"
      :file-loading="fileLoading"
      :error="error"
      :file-error="fileError"
      :save-error="saveError"
      :saving="saveLoading"
      :editable-paths="editablePaths"
      :current-directory-path="currentDirectoryPath"
      @open-directory="loadDirectory"
      @open-file="openFile"
      @save-file="saveFile"
    />
  </section>
</template>

<style scoped>
.awd-defense-page {
  min-height: 100%;
  display: grid;
  align-content: start;
  gap: var(--space-5);
  padding: var(--space-6);
}

.defense-icon {
  width: 1rem;
  height: 1rem;
}

.defense-hero {
  position: sticky;
  top: 0;
  z-index: 10;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4);
  padding: 0 0 var(--space-4);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: transparent;
}

.defense-hero__nav {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.defense-hero__back {
  --ui-btn-color: var(--color-text-secondary);
  --ui-btn-hover-background: color-mix(in srgb, var(--color-bg-page) 84%, var(--color-bg-surface));
  --ui-btn-hover-color: var(--color-text-primary);
}

.defense-hero__eyebrow {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.defense-hero h1 {
  margin: var(--space-3) 0 0;
  color: var(--color-text-primary);
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
}

.defense-hero__side {
  display: grid;
  justify-items: end;
  gap: var(--space-3);
}

.defense-hero__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  justify-content: flex-end;
}

.defense-chip {
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
  --ui-btn-background: color-mix(in srgb, var(--color-bg-page) 68%, transparent);
  --ui-btn-color: var(--color-text-primary);
  --ui-btn-hover-background: color-mix(in srgb, var(--color-bg-page) 82%, transparent);
  --ui-btn-hover-color: var(--color-text-primary);
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
    grid-template-columns: minmax(0, 1fr);
    padding: 0 0 var(--space-4);
  }

  .defense-hero__nav {
    align-items: flex-start;
    flex-direction: column;
    gap: var(--space-2);
  }

  .defense-hero__side {
    justify-items: start;
  }

  .defense-hero__meta {
    justify-content: flex-start;
  }

  .defense-refresh {
    width: 100%;
    justify-content: center;
    min-height: var(--control-height-md);
  }
}
</style>
