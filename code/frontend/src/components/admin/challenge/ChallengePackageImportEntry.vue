<script setup lang="ts">
import { useTemplateRef } from 'vue'

const props = defineProps<{
  uploading: boolean
  selectedFileName?: string
  hideHeader?: boolean
}>()

const emit = defineEmits<{
  select: [files: File[]]
}>()

const fileInput = useTemplateRef<HTMLInputElement>('fileInput')

function openPicker() {
  fileInput.value?.click()
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) {
    return
  }

  const files = target?.files ? Array.from(target.files) : []
  if (files.length === 0) {
    return
  }

  emit('select', files)
  target.value = ''
}
</script>

<template>
  <section class="import-entry">
    <div v-if="!props.hideHeader" class="import-entry__lead">
      <div class="import-entry__eyebrow">Challenge Package</div>
      <h2 class="import-entry__title">导入题目包</h2>
    </div>

    <div class="import-entry__panel import-entry__panel--single">
      <slot name="before-dropzone" />

      <button
        class="import-entry__dropzone"
        type="button"
        :disabled="uploading"
        @click="openPicker"
      >
        <span class="import-entry__drop-kicker">{{ uploading ? '解析中' : 'ZIP Package' }}</span>
        <strong class="import-entry__drop-title">
          {{ uploading ? '正在解析 challenge.yml 与题目内容' : '点击选择或重新上传题目包（支持多选）' }}
        </strong>
        <span class="import-entry__drop-copy">
          支持一次选择多个 Zip；每个包都需要单目录 Zip 或根目录直接包含 `challenge.yml`
        </span>
        <span v-if="selectedFileName" class="import-entry__file">{{ selectedFileName }}</span>
      </button>

      <input
        ref="fileInput"
        class="sr-only"
        type="file"
        multiple
        accept=".zip,application/zip"
        @change="handleFileChange"
      >
    </div>
  </section>
</template>

<style scoped>
.import-entry {
  display: grid;
  gap: 1.25rem;
  padding-block: 0.35rem 0.75rem;
}

.import-entry__lead {
  display: grid;
  gap: 0.55rem;
  max-width: 44rem;
}

.import-entry__eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.import-entry__title {
  margin: 0;
  font-size: clamp(1.5rem, 2vw, 1.95rem);
  font-weight: 700;
  color: var(--journal-ink);
}

.import-entry__panel {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 1.2fr) minmax(16rem, 0.8fr);
  align-items: stretch;
}

.import-entry__panel--single {
  grid-template-columns: minmax(0, 1fr);
}

.import-entry__dropzone {
  display: grid;
  gap: 0.55rem;
  min-height: 15rem;
  width: 100%;
  padding: 1.25rem;
  text-align: left;
  border: 1px dashed color-mix(in srgb, var(--journal-accent) 36%, transparent);
  border-radius: 1.25rem;
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)),
      color-mix(in srgb, #2563eb 8%, var(--journal-surface-subtle))
    ),
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.12), transparent 45%);
  transition: transform 160ms ease, border-color 160ms ease, box-shadow 160ms ease;
  cursor: pointer;
}

.import-entry__dropzone:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: color-mix(in srgb, var(--journal-accent) 58%, transparent);
  box-shadow: 0 18px 32px var(--color-shadow-soft);
}

.import-entry__dropzone:disabled {
  cursor: progress;
  opacity: 0.82;
}

.import-entry__drop-kicker {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.import-entry__drop-title {
  font-size: 1.2rem;
  line-height: 1.45;
  color: var(--journal-ink);
}

.import-entry__drop-copy {
  font-size: 0.9rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.import-entry__file {
  margin-top: auto;
  display: inline-flex;
  width: fit-content;
  align-items: center;
  padding: 0.38rem 0.72rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
  font-size: 0.78rem;
  font-weight: 700;
}

@media (max-width: 960px) {
  .import-entry__panel {
    grid-template-columns: 1fr;
  }
}
</style>
