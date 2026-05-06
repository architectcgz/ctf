<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { FileText, Folder } from 'lucide-vue-next'

import type { AWDDefenseDirectoryData, AWDDefenseFileData } from '@/api/contracts'

const props = defineProps<{
  serviceTitle: string
  directory?: AWDDefenseDirectoryData | null
  file?: AWDDefenseFileData | null
  loading: boolean
  fileLoading?: boolean
  error: string
  fileError?: string
  saving?: boolean
  saveError?: string
  editablePaths?: string[]
  currentDirectoryPath?: string
}>()

const emit = defineEmits<{
  openDirectory: [path: string]
  openFile: [path: string]
  saveFile: [path: string, content: string]
}>()

const draftContent = ref('')
const gutterRef = ref<HTMLElement | null>(null)

function formatSize(size?: number): string {
  if (!size) return '0 B'
  if (size < 1024) return `${size} B`
  return `${Math.ceil(size / 1024)} KiB`
}

function normalizePath(path?: string | null): string {
  if (!path) return '.'
  const normalized = path.replace(/^\.?\//, '').replace(/\/+/g, '/').replace(/\/$/, '')
  return normalized.length > 0 ? normalized : '.'
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

function isPathWithin(basePath: string, targetPath?: string | null): boolean {
  const normalizedBase = normalizePath(basePath)
  const normalizedTarget = normalizePath(targetPath)
  return (
    normalizedTarget === normalizedBase || normalizedTarget.startsWith(`${normalizedBase}/`)
  )
}

function detectFileLanguage(path?: string | null): string {
  const extension = path?.split('.').pop()?.toLowerCase()
  switch (extension) {
    case 'py':
      return 'Python'
    case 'sh':
      return 'Shell'
    case 'js':
    case 'ts':
      return 'Script'
    case 'json':
      return 'JSON'
    case 'yaml':
    case 'yml':
      return 'YAML'
    default:
      return 'Text'
  }
}

function syncGutterScroll(event: Event): void {
  const target = event.target
  if (!(target instanceof HTMLElement) || !gutterRef.value) return
  gutterRef.value.scrollTop = target.scrollTop
}

const activeDirectoryPath = computed(() => props.currentDirectoryPath || props.directory?.path || '.')
const isEditableFile = computed(() =>
  Boolean(
    props.file &&
      (props.editablePaths || []).some(
        (path) =>
          isPathWithin(path, props.file?.path) || isPathWithin(path, activeDirectoryPath.value)
      )
  )
)
const isDirty = computed(() =>
  Boolean(props.file && isEditableFile.value && draftContent.value !== (props.file.content || ''))
)
const editorText = computed(() =>
  isEditableFile.value ? draftContent.value : props.file?.content || ''
)
const editorLineNumbers = computed(() => {
  const lineCount = Math.max(editorText.value.split('\n').length, 1)
  return Array.from({ length: lineCount }, (_, index) => index + 1)
})
const fileLanguageLabel = computed(() => detectFileLanguage(props.file?.path))
const currentFileName = computed(() => props.file?.path?.split('/').pop() || '未选择文件')
const editorStateText = computed(() => {
  if (!isEditableFile.value) return '只读'
  if (props.saving) return '保存中'
  if (props.saveError && !isDirty.value) return '保存失败'
  if (isDirty.value) return '未保存'
  return '已保存'
})
const editorStateClass = computed(() => {
  if (!isEditableFile.value) return 'defense-file-workbench__editor-state--readonly'
  if (props.saving) return 'defense-file-workbench__editor-state--saving'
  if (props.saveError && !isDirty.value) return 'defense-file-workbench__editor-state--error'
  if (isDirty.value) return 'defense-file-workbench__editor-state--dirty'
  return 'defense-file-workbench__editor-state--saved'
})

function handleEditorKeydown(event: KeyboardEvent): void {
  if (event.key.toLowerCase() !== 's' || (!event.ctrlKey && !event.metaKey)) {
    return
  }
  event.preventDefault()
  triggerSave()
}

function triggerSave(): void {
  if (!props.file || !isEditableFile.value || props.saving || !isDirty.value) {
    return
  }
  emit('saveFile', props.file.path, draftContent.value)
}

// Use a window-level shortcut so Ctrl+S still works after focus drifts out of the textarea.
function handleWindowKeydown(event: KeyboardEvent): void {
  if (event.key.toLowerCase() !== 's' || (!event.ctrlKey && !event.metaKey)) {
    return
  }
  event.preventDefault()
  triggerSave()
}

watch(
  () => [props.file?.path, props.file?.content],
  () => {
    draftContent.value = props.file?.content || ''
  },
  { immediate: true }
)

onMounted(() => {
  window.addEventListener('keydown', handleWindowKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleWindowKeydown)
})
</script>

<template>
  <section
    class="defense-file-workbench metric-panel-default-surface metric-panel-workspace-surface"
    aria-label="防守文件"
  >
    <div v-if="error" class="defense-file-workbench__notice">{{ error }}</div>
    <div v-else class="defense-file-workbench__body">
      <aside class="defense-file-workbench__sidebar defense-file-workbench__sidebar--nav">
        <div class="defense-file-workbench__sidebar-head">
          <div class="defense-file-workbench__section-label">目录入口</div>
          <div class="defense-file-workbench__sidebar-title">{{ serviceTitle || '未选择服务' }}</div>
        </div>
        <div class="defense-file-workbench__path">
          <div class="defense-file-workbench__path-actions">
            <button
              class="defense-file-workbench__path-button ui-btn ui-btn--sm ui-btn--ghost"
              type="button"
              @click="emit('openDirectory', '.')"
            >
              根目录
            </button>
            <button
              v-if="canOpenParent(activeDirectoryPath)"
              class="defense-file-workbench__path-button ui-btn ui-btn--sm ui-btn--ghost"
              type="button"
              @click="emit('openDirectory', parentDirectoryPath(activeDirectoryPath))"
            >
              上一级
            </button>
          </div>
          <span class="font-mono">{{ activeDirectoryPath }}</span>
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
            class="defense-file-workbench__entry ui-btn ui-btn--ghost"
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
        <div v-if="fileLoading" class="defense-file-workbench__empty">正在读取文件...</div>
        <div v-else-if="fileError" class="defense-file-workbench__empty">{{ fileError }}</div>
        <div v-else-if="!file" class="defense-file-workbench__empty">点击左侧文件后在这里查看或修改内容。</div>
        <div v-else class="defense-file-workbench__editor-shell">
          <div class="defense-file-workbench__editor-bar">
            <div class="defense-file-workbench__editor-bar-main">
              <span class="defense-file-workbench__editor-name font-mono">{{ currentFileName }}</span>
              <span class="defense-file-workbench__editor-path font-mono">{{ file.path }}</span>
              <span class="defense-file-workbench__editor-badge">{{ fileLanguageLabel }}</span>
            </div>
            <span
              class="defense-file-workbench__editor-state"
              :class="editorStateClass"
              :title="saveError || ''"
            >
              <span class="defense-file-workbench__editor-state-dot" />
              <span>{{ formatSize(file.size) }} · {{ editorStateText }}</span>
            </span>
          </div>
          <div class="defense-file-workbench__editor-frame">
            <div ref="gutterRef" class="defense-file-workbench__editor-gutter" aria-hidden="true">
              <span v-for="line in editorLineNumbers" :key="line">{{ line }}</span>
            </div>
            <textarea
              v-if="isEditableFile"
              v-model="draftContent"
              class="defense-file-workbench__editor"
              spellcheck="false"
              @keydown="handleEditorKeydown"
              @scroll="syncGutterScroll"
            />
            <pre v-else class="defense-file-workbench__content" @scroll="syncGutterScroll"><code>{{ file.content }}</code></pre>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.defense-file-workbench {
  --defense-file-workbench-preview-top: calc(var(--space-6) * 3);
  display: grid;
  gap: 0;
  overflow: hidden;
  border: 1px solid var(--metric-panel-border, color-mix(in srgb, var(--color-border-default) 84%, transparent));
  border-radius: var(--metric-panel-radius, var(--radius-xl));
  background: color-mix(in srgb, var(--color-bg-page) 92%, var(--color-bg-surface));
  box-shadow: var(--metric-panel-shadow, none);
}

.defense-file-workbench__notice {
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  padding: var(--space-4);
}

.defense-file-workbench__body {
  display: grid;
  min-height: clamp(30rem, 70vh, 48rem);
}

.defense-file-workbench__sidebar {
  min-width: 0;
}

.defense-file-workbench__sidebar--nav {
  display: grid;
  align-content: start;
  gap: var(--space-3);
  padding: var(--space-4);
  background: color-mix(in srgb, var(--color-bg-page) 96%, var(--color-bg-surface));
}

.defense-file-workbench__sidebar-head {
  display: grid;
  gap: var(--space-1);
}

.defense-file-workbench__section-label {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.defense-file-workbench__sidebar-title {
  color: var(--color-text-primary);
  font-size: var(--font-size-14);
  font-weight: 800;
}

.defense-file-workbench__path {
  display: grid;
  gap: var(--space-2);
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
}

.defense-file-workbench__path-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.defense-file-workbench__path-button {
  --ui-btn-color: var(--color-text-secondary);
  --ui-btn-hover-background: color-mix(in srgb, var(--color-bg-elevated) 82%, transparent);
  --ui-btn-hover-border: color-mix(in srgb, var(--color-border-default) 88%, transparent);
}

.defense-file-workbench__entries {
  display: grid;
  gap: var(--space-2);
  align-content: start;
  min-height: 0;
  overflow: auto;
}

.defense-file-workbench__entry {
  align-items: center;
  display: flex;
  justify-content: flex-start;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  gap: var(--space-2);
  min-width: 0;
  padding: var(--space-2) var(--space-3);
  text-align: left;
  --ui-btn-justify: flex-start;
  --ui-btn-color: var(--color-text-secondary);
  --ui-btn-background: color-mix(in srgb, var(--color-bg-page) 84%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--color-bg-surface) 90%, transparent);
  --ui-btn-hover-color: var(--color-text-primary);
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
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.defense-file-workbench__editor {
  width: 100%;
  min-height: 0;
  flex: 1;
  resize: none;
  overflow: auto;
  border: 0;
  background: transparent;
  color: var(--color-text-primary);
  padding: var(--space-3);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  line-height: 1.6;
  outline: none;
}

.defense-file-workbench__empty {
  min-height: 100%;
  flex: 1;
  display: grid;
  place-items: center;
  color: var(--color-text-muted);
  font-size: var(--font-size-12);
  text-align: center;
  line-height: 1.7;
}

.defense-file-workbench__content {
  background: transparent;
  border: 0;
  border-radius: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-family: var(--font-family-mono);
  line-height: 1.6;
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: var(--space-3);
  white-space: pre-wrap;
}

.defense-file-workbench__editor-shell {
  display: flex;
  flex: 1;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  border-radius: var(--radius-lg);
  background: color-mix(in srgb, var(--color-bg-page) 88%, var(--color-bg-surface));
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--color-shadow-soft) 18%, transparent);
}

.defense-file-workbench__editor-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  min-height: var(--ui-control-height-sm);
  padding: 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
}

.defense-file-workbench__editor-bar-main {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.defense-file-workbench__editor-name {
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 700;
}

.defense-file-workbench__editor-path {
  color: var(--color-text-muted);
  font-size: var(--font-size-11);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.defense-file-workbench__editor-badge,
.defense-file-workbench__editor-state {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  min-height: var(--ui-control-height-xs);
  padding: 0 var(--space-2);
  border-radius: 999px;
  font-size: var(--font-size-11);
  font-weight: 700;
}

.defense-file-workbench__editor-badge {
  color: var(--color-text-secondary);
  background: color-mix(in srgb, var(--color-bg-page) 78%, transparent);
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
}

.defense-file-workbench__editor-state {
  color: var(--color-text-secondary);
}

.defense-file-workbench__editor-state-dot {
  width: var(--space-2);
  height: var(--space-2);
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-primary) 72%, var(--color-success));
  flex: 0 0 auto;
}

.defense-file-workbench__editor-state--saved .defense-file-workbench__editor-state-dot {
  background: color-mix(in srgb, var(--color-success) 80%, var(--color-primary));
}

.defense-file-workbench__editor-state--dirty .defense-file-workbench__editor-state-dot,
.defense-file-workbench__editor-state--readonly .defense-file-workbench__editor-state-dot {
  background: color-mix(in srgb, var(--color-warning) 78%, var(--color-text-muted));
}

.defense-file-workbench__editor-state--saving .defense-file-workbench__editor-state-dot {
  background: var(--color-primary);
}

.defense-file-workbench__editor-state--error .defense-file-workbench__editor-state-dot {
  background: var(--color-danger);
}

.defense-file-workbench__editor-frame {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  flex: 1;
  min-height: 0;
  background: color-mix(in srgb, var(--color-bg-base) 90%, var(--color-bg-surface));
}

.defense-file-workbench__editor-gutter {
  display: grid;
  align-content: start;
  gap: 0;
  min-width: calc(var(--space-5) + var(--space-4));
  padding: var(--space-3) 0;
  overflow: hidden;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-11);
  line-height: 1.6;
  text-align: right;
  background: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
  border-right: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
}

.defense-file-workbench__editor-gutter span {
  display: block;
  padding: 0 var(--space-2);
}

@media (min-width: 900px) {
  .defense-file-workbench__body {
    grid-template-columns: minmax(17rem, 18rem) minmax(0, 1fr);
    align-items: start;
  }

  .defense-file-workbench__sidebar {
    position: sticky;
    top: var(--space-4);
    align-self: start;
    max-height: calc(clamp(30rem, 70vh, 48rem) - var(--space-8));
    border-right: 1px solid var(--color-border-default);
  }

  .defense-file-workbench__file {
    position: sticky;
    top: var(--defense-file-workbench-preview-top);
    align-self: start;
    max-height: calc(100vh - var(--defense-file-workbench-preview-top) - var(--space-4));
    overflow: hidden;
  }
}

@media (max-width: 899px) {
  .defense-file-workbench__editor-bar {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
