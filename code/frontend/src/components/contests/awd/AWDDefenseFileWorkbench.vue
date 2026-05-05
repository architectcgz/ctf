<script setup lang="ts">
import { FileText, Folder, RefreshCw } from 'lucide-vue-next'

import type { AWDDefenseDirectoryData, AWDDefenseFileData } from '@/api/contracts'

defineProps<{
  serviceTitle: string
  directory?: AWDDefenseDirectoryData | null
  file?: AWDDefenseFileData | null
  loading: boolean
  fileLoading?: boolean
  error: string
  fileError?: string
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

function parentDirectoryPath(currentPath?: string | null): string {
  if (!currentPath || currentPath === '.') return '.'
  const parts = currentPath.split('/').filter(Boolean)
  if (parts.length <= 1) return '.'
  return parts.slice(0, -1).join('/')
}

function canOpenParent(currentPath?: string | null): boolean {
  return Boolean(currentPath && currentPath !== '.')
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
    <div v-else class="defense-file-workbench__body">
      <aside class="defense-file-workbench__sidebar">
        <div class="defense-file-workbench__path">
          <div class="defense-file-workbench__path-actions">
            <button
              class="defense-file-workbench__path-button"
              type="button"
              @click="emit('openDirectory', '.')"
            >
              根目录
            </button>
            <button
              v-if="canOpenParent(directory?.path)"
              class="defense-file-workbench__path-button"
              type="button"
              @click="emit('openDirectory', parentDirectoryPath(directory?.path))"
            >
              上一级
            </button>
          </div>
          <span class="font-mono">{{ directory?.path || '.' }}</span>
        </div>
        <div v-if="loading && !directory" class="defense-file-workbench__notice">正在读取目录...</div>
        <div v-else-if="!directory" class="defense-file-workbench__notice">暂无文件列表。</div>
        <div v-else-if="directory.entries.length === 0" class="defense-file-workbench__notice">
          当前目录为空。
        </div>
        <div v-else class="defense-file-workbench__entries">
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
      </aside>

      <article class="defense-file-workbench__file">
        <div v-if="file" class="defense-file-workbench__file-head">
          <span class="font-mono">{{ file.path }}</span>
          <span>{{ formatSize(file.size) }}</span>
        </div>
        <div v-if="fileLoading" class="defense-file-workbench__empty">正在读取文件...</div>
        <div v-else-if="fileError" class="defense-file-workbench__empty">{{ fileError }}</div>
        <div v-else-if="!file" class="defense-file-workbench__empty">点击左侧文件后在这里查看内容。</div>
        <pre v-else class="defense-file-workbench__content"><code>{{ file.content }}</code></pre>
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
  display: grid;
  gap: var(--space-4);
}

.defense-file-workbench__sidebar {
  min-width: 0;
}

.defense-file-workbench__path {
  display: grid;
  gap: var(--space-2);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  margin-bottom: var(--space-2);
}

.defense-file-workbench__path-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.defense-file-workbench__path-button {
  min-height: var(--ui-control-height-sm);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 var(--space-3);
  border-radius: var(--ui-control-radius-sm);
  border: 1px solid var(--color-border-default);
  background: color-mix(in srgb, var(--color-bg-elevated) 82%, transparent);
  color: var(--color-text-secondary);
  font-size: var(--font-size-11);
  font-weight: 800;
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
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-bg-elevated) 82%, transparent);
  min-height: 18rem;
  padding: var(--space-3);
}

.defense-file-workbench__file-head {
  color: var(--color-text-muted);
  display: flex;
  font-size: var(--font-size-11);
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
}

.defense-file-workbench__empty {
  min-height: 100%;
  display: grid;
  place-items: center;
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  text-align: center;
  line-height: 1.7;
}

.defense-file-workbench__content {
  background: var(--color-bg-page);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  line-height: 1.6;
  min-height: 18rem;
  max-height: 38rem;
  overflow: auto;
  padding: var(--space-3);
  white-space: pre-wrap;
}

@media (min-width: 900px) {
  .defense-file-workbench__body {
    grid-template-columns: minmax(18rem, 22rem) minmax(0, 1fr);
    align-items: start;
  }

  .defense-file-workbench__file {
    min-height: 24rem;
  }

  .defense-file-workbench__content {
    min-height: 24rem;
  }
}
</style>
