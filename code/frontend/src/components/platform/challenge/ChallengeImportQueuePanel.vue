<script setup lang="ts">
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import { ChallengeCategoryDifficultyPills } from '@/entities/challenge'

type ChallengeImportQueueItem = {
  id: string
  title: string
  file_name: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  points: number
  created_at: string
}

defineProps<{
  queueLoading: boolean
  queueCount: number
  queue: ChallengeImportQueueItem[]
  formatDateTime: (value: string) => string
}>()

const emit = defineEmits<{
  inspect: [importId: string]
}>()

function handleInspect(importId: string): void {
  emit('inspect', importId)
}
</script>

<template>
  <section
    id="challenge-queue-workspace"
    class="workspace-directory-section challenge-import-directory challenge-workspace-section"
  >
    <div class="list-heading challenge-directory-head challenge-section-heading">
      <div>
        <div class="workspace-overline">
          Import Review
        </div>
        <h2 class="list-heading__title">
          待确认导入
        </h2>
        <p class="challenge-section-copy">
          这里列出已生成预览、但还没正式导入题库的题目包。确认无误后，再继续写入题库。
        </p>
      </div>
      <div class="challenge-directory-meta">
        共 {{ queueCount }} 个待处理任务
      </div>
    </div>

    <div
      v-if="queueLoading"
      class="challenge-directory-state"
    >
      正在同步导入队列...
    </div>
    <div
      v-else-if="queue.length === 0"
      class="challenge-directory-state"
    >
      当前没有待确认的导入任务。
    </div>

    <div
      v-else
      class="challenge-panel-stack"
    >
      <article
        v-for="item in queue"
        :key="item.id"
        class="challenge-plain-section challenge-queue-item"
      >
        <div class="flex min-w-0 items-start gap-4">
          <div class="challenge-queue-id">
            IMP-{{ item.id.slice(0, 6).toUpperCase() }}
          </div>
          <div class="min-w-0 flex-1">
            <h2
              class="challenge-queue-title"
              :title="item.title"
            >
              {{ item.title }}
            </h2>
            <p
              class="challenge-queue-file"
              :title="item.file_name"
            >
              {{ item.file_name }}
            </p>
            <div class="mt-3 flex flex-wrap gap-2">
              <ChallengeCategoryDifficultyPills
                :category="item.category"
                :difficulty="item.difficulty"
              />
              <span class="challenge-queue-points">{{ item.points }} pts</span>
            </div>
          </div>
        </div>

        <div class="flex flex-col items-start gap-2 md:items-end">
          <div class="challenge-queue-time">
            {{ formatDateTime(item.created_at) }}
          </div>
          <button
            type="button"
            class="ui-btn ui-btn--primary"
            @click="handleInspect(item.id)"
          >
            继续查看预览
          </button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.challenge-workspace-section {
  scroll-margin-top: 6rem;
}

.challenge-import-directory {
  display: grid;
  gap: 1.5rem;
}

.challenge-section-heading {
  align-items: flex-start;
  gap: 1rem;
}

.challenge-section-copy {
  margin: 0.5rem 0 0;
  max-width: 44rem;
  font-size: 0.92rem;
  line-height: 1.6;
  color: var(--challenge-page-muted);
}

.challenge-panel-stack {
  display: grid;
  gap: 1rem;
}

.challenge-directory-state,
.challenge-directory-meta,
.challenge-queue-file,
.challenge-queue-time {
  color: var(--challenge-page-muted);
}

.challenge-queue-title {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 1rem;
  font-weight: 700;
  color: var(--challenge-page-text);
}

.challenge-queue-file {
  margin: 0.25rem 0 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.875rem;
}

.challenge-queue-id {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 5.5rem;
  height: 2rem;
  padding: 0 0.75rem;
  border: 1px solid color-mix(in srgb, var(--workspace-brand) 18%, var(--challenge-page-line));
  border-radius: 999px;
  background: color-mix(in srgb, var(--workspace-brand) 9%, var(--challenge-page-surface));
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--challenge-page-accent);
}

.challenge-queue-points {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--challenge-page-muted);
}

</style>
