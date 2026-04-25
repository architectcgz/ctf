<script setup lang="ts">
type ChallengeImportUploadResult = {
  id: string
  status: 'success' | 'error'
  fileName: string
  message: string
  createdAt: string
  code?: number
  requestId?: string
}

defineProps<{
  uploadResults: ChallengeImportUploadResult[]
  formatDateTime: (value: string) => string
}>()
</script>

<template>
  <section class="workspace-directory-section challenge-import-directory challenge-plain-section">
    <div class="list-heading">
      <div>
        <div class="workspace-overline">
          Upload Receipt
        </div>
        <h2 class="list-heading__title">
          最近上传结果
        </h2>
      </div>
    </div>

    <div
      v-if="uploadResults.length === 0"
      class="challenge-directory-state"
    >
      还没有新的上传回执，选择题目包后会在这里显示解析结果。
    </div>

    <div
      v-else
      class="challenge-panel-stack"
    >
      <article
        v-for="result in uploadResults"
        :key="result.id"
        class="challenge-upload-result"
        :class="
          result.status === 'success'
            ? 'challenge-upload-result--success'
            : 'challenge-upload-result--error'
        "
      >
        <div class="challenge-upload-result__head">
          <span
            class="challenge-upload-result__status"
            :class="
              result.status === 'success'
                ? 'challenge-upload-result__status--success'
                : 'challenge-upload-result__status--error'
            "
          >
            {{ result.status === 'success' ? '成功' : '失败' }}
          </span>
          <strong
            class="challenge-upload-result__title"
            :title="result.fileName"
          >
            {{ result.fileName }}
          </strong>
        </div>
        <p class="challenge-upload-result__copy">
          {{ result.message }}
        </p>
        <div class="challenge-upload-result__meta">
          <span>{{ formatDateTime(result.createdAt) }}</span>
          <span v-if="result.code !== undefined">错误码 {{ result.code }}</span>
          <span v-if="result.requestId">请求ID {{ result.requestId }}</span>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.challenge-import-directory {
  display: grid;
  gap: 1.5rem;
}

.challenge-panel-stack {
  display: grid;
  gap: 1rem;
}

.challenge-directory-state,
.challenge-upload-result__copy,
.challenge-upload-result__meta {
  color: var(--challenge-page-muted);
}

.challenge-upload-result {
  display: grid;
  gap: 0.5rem;
  padding: 1rem;
  border: 1px solid var(--challenge-page-line);
  border-radius: 1rem;
  background: var(--challenge-page-surface);
}

.challenge-upload-result--success {
  border-color: color-mix(in srgb, var(--color-success) 24%, var(--challenge-page-line));
  background: color-mix(in srgb, var(--color-success) 9%, var(--challenge-page-surface));
}

.challenge-upload-result--error {
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--challenge-page-line));
  background: color-mix(in srgb, var(--color-danger) 8%, var(--challenge-page-surface));
}

.challenge-upload-result__head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.challenge-upload-result__status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.challenge-upload-result__status--success {
  background: color-mix(in srgb, var(--color-success) 16%, transparent);
  color: color-mix(in srgb, var(--color-success) 92%, var(--challenge-page-text));
}

.challenge-upload-result__status--error {
  background: color-mix(in srgb, var(--color-danger) 14%, transparent);
  color: color-mix(in srgb, var(--color-danger) 92%, var(--challenge-page-text));
}

.challenge-upload-result__title {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 1rem;
  font-weight: 700;
  color: var(--challenge-page-text);
}

.challenge-upload-result__copy {
  margin: 0;
  font-size: 0.8rem;
}

.challenge-upload-result__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.65rem;
  font-weight: 600;
}
</style>
