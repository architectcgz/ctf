<script setup lang="ts">
import { useTemplateRef } from 'vue'
import { UploadCloud } from 'lucide-vue-next'

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
    <div
      v-if="!props.hideHeader"
      class="import-entry__lead"
    >
      <div class="import-entry__eyebrow">
        Challenge Package
      </div>
      <h2 class="import-entry__title">
        导入题目包
      </h2>
    </div>

    <div class="import-entry__panel import-entry__panel--single">
      <slot name="before-dropzone" />

      <div class="import-entry__upload-card">
        <div class="import-entry__upload-copy">
          <span class="import-entry__drop-kicker">
            {{ uploading ? '解析中' : 'ZIP Package' }}
          </span>
          <strong class="import-entry__drop-title">
            {{ uploading ? '正在解析 challenge.yml 与题目内容' : '选择题目包并进入预览' }}
          </strong>
          <span class="import-entry__drop-copy">
            支持一次选择多个 Zip；每个包都需要单目录 Zip 或根目录直接包含 `challenge.yml`
          </span>
        </div>
        <button
          class="ui-btn ui-btn--primary challenge-import-action challenge-import-action--primary import-entry__upload-action"
          type="button"
          :disabled="uploading"
          @click="openPicker"
        >
          <UploadCloud class="h-4 w-4" />
          {{ uploading ? '解析中' : '导入题目包' }}
        </button>
        <span
          v-if="selectedFileName"
          class="import-entry__file"
        >
          {{ selectedFileName }}
        </span>
      </div>

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
  gap: var(--space-5);
  padding-block: var(--space-1-5) var(--space-3);
}

.import-entry__lead {
  display: grid;
  gap: var(--space-2);
  max-width: 44rem;
}

.import-entry__eyebrow {
  font-size: var(--font-size-0-70);
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
  gap: var(--space-4);
  grid-template-columns: minmax(0, 1.2fr) minmax(16rem, 0.8fr);
  align-items: stretch;
}

.import-entry__panel--single {
  grid-template-columns: minmax(0, 1fr);
}

.import-entry__upload-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-5);
  text-align: left;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, var(--journal-border));
  border-radius: var(--workspace-radius-lg, var(--ui-control-radius-lg));
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface-subtle))
    ),
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 12%, transparent),
      transparent 45%
    );
}

.import-entry__upload-copy {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.import-entry__drop-kicker {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.import-entry__drop-title {
  font-size: var(--font-size-1-20);
  line-height: 1.45;
  color: var(--journal-ink);
}

.import-entry__drop-copy {
  font-size: var(--font-size-0-90);
  line-height: 1.7;
  color: var(--journal-muted);
}

.import-entry__file {
  grid-column: 1 / -1;
  display: inline-flex;
  width: fit-content;
  align-items: center;
  padding: var(--space-1-5) var(--space-3);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
  font-size: var(--font-size-0-78);
  font-weight: 700;
}

.challenge-import-action {
  --ui-btn-height: 2.5rem;
  --ui-btn-padding: 0 var(--space-5);
  --ui-btn-radius: var(--ui-control-radius-md);
  --ui-btn-font-size: var(--font-size-12);
  --ui-btn-font-weight: 700;
  --ui-btn-hover-transform: translateY(-1px);
  box-shadow: 0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
}

.challenge-import-action--primary {
  --ui-btn-primary-border: color-mix(in srgb, var(--workspace-brand) 42%, transparent);
  --ui-btn-primary-background: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  --ui-btn-primary-hover-background: color-mix(
    in srgb,
    var(--workspace-brand-ink) 92%,
    var(--challenge-page-text)
  );
  --ui-btn-primary-hover-border: color-mix(in srgb, var(--workspace-brand-ink) 62%, transparent);
  --ui-btn-primary-hover-shadow: 0 10px 24px color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}

.import-entry__upload-action {
  justify-self: end;
  white-space: nowrap;
}

@media (max-width: 960px) {
  .import-entry__panel {
    grid-template-columns: 1fr;
  }

  .import-entry__upload-card {
    grid-template-columns: 1fr;
  }

  .import-entry__upload-action {
    justify-self: start;
  }
}
</style>
