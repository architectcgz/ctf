<script setup lang="ts">
import { FileText, Folder, RefreshCw } from 'lucide-vue-next'

import type { AWDDefenseDirectoryData, AWDDefenseFileData } from '@/api/contracts'

defineProps<{
  serviceTitle: string
  directory?: AWDDefenseDirectoryData | null
  file?: AWDDefenseFileData | null
  loading: boolean
  error: string
}>()

const emit = defineEmits<{
  refresh: []
  openDirectory: [path: string]
  openFile: [path: string]
}>()

function formatSize(size?: number): string {
  if (!size) return '0 B'
  if (size < 1024) return `${size} B`
  return `${Math.ceil(size / 1024)} KiB`
}
</script>

<template>
  <section class="defense-file-workbench" aria-label="防守文件">
    <header class="defense-file-workbench__header">
      <div>
        <div class="defense-file-workbench__eyebrow">防守文件</div>
        <h4 class="defense-file-workbench__title">{{ serviceTitle || '未选择服务' }}</h4>
      </div>
      <button
        class="defense-file-workbench__refresh"
        type="button"
        :disabled="loading || !serviceTitle"
        @click="emit('refresh')"
      >
        <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
        <span>刷新</span>
      </button>
    </header>

    <div v-if="error" class="defense-file-workbench__notice">{{ error }}</div>
    <div v-else-if="!directory" class="defense-file-workbench__notice">暂无文件列表。</div>
    <div v-else class="defense-file-workbench__body">
      <div class="defense-file-workbench__path font-mono">{{ directory.path }}</div>
      <div class="defense-file-workbench__entries">
        <button
          v-for="entry in directory.entries"
          :key="entry.path"
          class="defense-file-workbench__entry"
          type="button"
          @click="entry.type === 'dir' ? emit('openDirectory', entry.path) : emit('openFile', entry.path)"
        >
          <Folder v-if="entry.type === 'dir'" class="h-3.5 w-3.5" />
          <FileText v-else class="h-3.5 w-3.5" />
          <span class="defense-file-workbench__entry-name">{{ entry.name }}</span>
          <span class="defense-file-workbench__entry-size font-mono">{{ formatSize(entry.size) }}</span>
        </button>
      </div>

      <article v-if="file" class="defense-file-workbench__file">
        <div class="defense-file-workbench__file-head">
          <span class="font-mono">{{ file.path }}</span>
          <span>{{ formatSize(file.size) }}</span>
        </div>
        <pre class="defense-file-workbench__content"><code>{{ file.content }}</code></pre>
      </article>
    </div>
  </section>
</template>

<style scoped>
.defense-file-workbench {
  margin-top: var(--space-4);
  border-top: 1px solid var(--color-border-default);
  padding-top: var(--space-4);
}

.defense-file-workbench__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.defense-file-workbench__eyebrow {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
}

.defense-file-workbench__title {
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: 800;
  margin: var(--space-1) 0 0;
}

.defense-file-workbench__refresh,
.defense-file-workbench__entry {
  align-items: center;
  border: 1px solid var(--color-border-default);
  display: flex;
}

.defense-file-workbench__refresh {
  background: var(--color-bg-elevated);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 800;
  gap: var(--space-1);
  padding: var(--space-2) var(--space-3);
}

.defense-file-workbench__refresh:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.defense-file-workbench__notice {
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  margin-top: var(--space-3);
}

.defense-file-workbench__body {
  margin-top: var(--space-3);
}

.defense-file-workbench__path {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  margin-bottom: var(--space-2);
}

.defense-file-workbench__entries {
  display: grid;
  gap: var(--space-2);
}

.defense-file-workbench__entry {
  background: color-mix(in srgb, var(--color-bg-elevated) 78%, transparent);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  gap: var(--space-2);
  min-width: 0;
  padding: var(--space-2);
  text-align: left;
}

.defense-file-workbench__entry-name {
  color: var(--color-text-primary);
  flex: 1;
  font-size: var(--font-size-12);
  font-weight: 700;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.defense-file-workbench__entry-size {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
}

.defense-file-workbench__file {
  margin-top: var(--space-3);
}

.defense-file-workbench__file-head {
  color: var(--color-text-muted);
  display: flex;
  font-size: var(--font-size-11);
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
}

.defense-file-workbench__content {
  background: var(--color-bg-page);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  line-height: 1.6;
  max-height: 18rem;
  overflow: auto;
  padding: var(--space-3);
  white-space: pre-wrap;
}
</style>
