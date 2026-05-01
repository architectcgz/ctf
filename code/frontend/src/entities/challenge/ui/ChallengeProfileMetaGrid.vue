<script setup lang="ts">
import type { AdminChallengeListItem } from '@/api/contracts'
import {
  formatChallengeDateTime,
  getChallengeInstanceSharingLabel,
} from '../model'

interface Props {
  challenge: Pick<
    AdminChallengeListItem,
    'attachment_url' | 'created_at' | 'image_id' | 'instance_sharing' | 'title' | 'updated_at'
  >
  downloadingAttachment: boolean
  flagConfigSummary: string
}

defineProps<Props>()

const emit = defineEmits<{
  downloadAttachment: []
}>()
</script>

<template>
  <div class="journal-panel challenge-profile-card">
    <dl class="challenge-meta-grid">
      <div class="challenge-meta-item">
        <dt>题目名称</dt>
        <dd>{{ challenge.title }}</dd>
      </div>
      <div class="challenge-meta-item">
        <dt>Flag 配置</dt>
        <dd class="challenge-meta-item__mono">
          {{ flagConfigSummary }}
        </dd>
      </div>
      <div
        v-if="challenge.image_id"
        class="challenge-meta-item"
      >
        <dt>镜像</dt>
        <dd class="challenge-meta-item__mono">
          ID #{{ challenge.image_id }}
        </dd>
      </div>
      <div class="challenge-meta-item">
        <dt>实例模式</dt>
        <dd>{{ getChallengeInstanceSharingLabel(challenge.instance_sharing) }}</dd>
      </div>
      <div class="challenge-meta-item">
        <dt>创建时间</dt>
        <dd>{{ formatChallengeDateTime(challenge.created_at) }}</dd>
      </div>
      <div class="challenge-meta-item">
        <dt>最近更新</dt>
        <dd>{{ formatChallengeDateTime(challenge.updated_at) }}</dd>
      </div>
      <div
        v-if="challenge.attachment_url"
        class="challenge-meta-item challenge-meta-item--full challenge-meta-item--action"
      >
        <dt>附件</dt>
        <dd>
          <button
            type="button"
            class="challenge-link challenge-link-button"
            :disabled="downloadingAttachment"
            @click="emit('downloadAttachment')"
          >
            {{ downloadingAttachment ? '下载中...' : '下载附件' }}
          </button>
        </dd>
      </div>
    </dl>
  </div>
</template>

<style scoped>
.challenge-profile-card {
  padding: var(--space-5);
}

.challenge-profile-card .challenge-meta-grid {
  border-top: 0;
}

.challenge-meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-meta-item {
  display: grid;
  gap: var(--space-1-5);
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-meta-item:nth-child(odd) {
  padding-right: var(--space-4);
}

.challenge-meta-item:nth-child(even) {
  padding-left: var(--space-4);
}

.challenge-meta-item dt {
  font-size: var(--font-size-0-74);
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-meta-item dd {
  margin: 0;
  font-size: var(--font-size-0-92);
  font-weight: 600;
  color: var(--journal-ink);
  word-break: break-word;
}

.challenge-meta-item__mono {
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
}

.challenge-meta-item--full {
  grid-column: 1 / -1;
  padding-inline: 0;
}

.challenge-meta-item--action dd {
  display: flex;
  align-items: center;
}

.challenge-link {
  color: var(--journal-accent);
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.15em;
}

.challenge-link-button {
  padding: 0;
  border: 0;
  background: transparent;
  font: inherit;
  cursor: pointer;
}
</style>
