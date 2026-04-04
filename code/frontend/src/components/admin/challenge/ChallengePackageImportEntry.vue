<script setup lang="ts">
import { useTemplateRef } from 'vue'

const props = defineProps<{
  uploading: boolean
  selectedFileName?: string
}>()

const emit = defineEmits<{
  select: [file: File]
}>()

const fileInput = useTemplateRef<HTMLInputElement>('fileInput')

function openPicker() {
  fileInput.value?.click()
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement | null
  const file = target?.files?.[0]
  if (!file) {
    return
  }

  emit('select', file)
  target.value = ''
}
</script>

<template>
  <section class="import-entry">
    <div class="import-entry__lead">
      <div class="import-entry__eyebrow">Challenge Package</div>
      <h2 class="import-entry__title">导入题目包</h2>
      <p class="import-entry__copy">
        使用 `challenge.yml` 作为唯一主规范导入题目。题面、附件、Flag、提示和运行时镜像会在这里一次进入平台。
      </p>
    </div>

    <div class="import-entry__panel">
      <button
        class="import-entry__dropzone"
        type="button"
        :disabled="uploading"
        @click="openPicker"
      >
        <span class="import-entry__drop-kicker">{{ uploading ? '解析中' : 'ZIP Package' }}</span>
        <strong class="import-entry__drop-title">
          {{ uploading ? '正在解析 challenge.yml 与题目内容' : '点击选择或重新上传题目包' }}
        </strong>
        <span class="import-entry__drop-copy">
          支持单目录 Zip 或根目录直接包含 `challenge.yml`
        </span>
        <span v-if="selectedFileName" class="import-entry__file">{{ selectedFileName }}</span>
      </button>

      <input
        ref="fileInput"
        class="sr-only"
        type="file"
        accept=".zip,application/zip"
        @change="handleFileChange"
      >

      <div class="import-entry__rail">
        <div class="import-entry__rail-title">首版导入范围</div>
        <ul class="import-entry__rail-list">
          <li>题目元数据与题面 Markdown</li>
          <li>附件与提示系统</li>
          <li>静态或动态 Flag 配置</li>
          <li>运行时镜像引用与拓扑扩展提示</li>
        </ul>
      </div>
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

.import-entry__copy {
  margin: 0;
  max-width: 42rem;
  font-size: 0.95rem;
  line-height: 1.8;
  color: var(--journal-muted);
}

.import-entry__panel {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 1.2fr) minmax(16rem, 0.8fr);
  align-items: stretch;
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

.import-entry__rail {
  display: grid;
  gap: 0.65rem;
  padding: 1.1rem 1rem;
  border-left: 3px solid rgba(37, 99, 235, 0.18);
  background: var(--journal-surface);
}

.import-entry__rail-title {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.import-entry__rail-list {
  display: grid;
  gap: 0.55rem;
  margin: 0;
  padding-left: 1rem;
  color: var(--journal-ink);
  font-size: 0.9rem;
  line-height: 1.7;
}

@media (max-width: 960px) {
  .import-entry__panel {
    grid-template-columns: 1fr;
  }
}
</style>
