<script setup lang="ts">
import {
  Activity,
  Clock,
  FileJson,
  Fingerprint,
  Package,
  User,
} from 'lucide-vue-next'

import type { AuditLogItem } from '@/api/contracts'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'

defineProps<{
  open: boolean
  item: AuditLogItem | null
  formatDate: (value: string) => string
  actorDisplayName: (item: AuditLogItem) => string
  resourceDisplayName: (item: AuditLogItem) => string
  detailPreview: (detail: Record<string, unknown> | undefined) => string
}>()

const emit = defineEmits<{
  close: []
  'update:open': [value: boolean]
}>()

function handleClose(): void {
  emit('close')
}

function handleOpenUpdate(value: boolean): void {
  emit('update:open', value)
  if (!value) {
    emit('close')
  }
}
</script>

<template>
  <AdminSurfaceModal
    class="audit-actor-modal"
    :open="open"
    title="执行人详情"
    eyebrow="Audit Log"
    width="34rem"
    @close="handleClose"
    @update:open="handleOpenUpdate"
  >
    <section
      v-if="item"
      class="audit-actor-detail"
    >
      <div class="audit-actor-detail__grid">
        <article class="audit-actor-detail__item">
          <div class="audit-actor-detail__head">
            <User class="h-3.5 w-3.5" />
            <span class="audit-actor-detail__label">用户名</span>
          </div>
          <strong class="audit-actor-detail__value">
            {{ actorDisplayName(item) }}
          </strong>
        </article>

        <article class="audit-actor-detail__item">
          <div class="audit-actor-detail__head">
            <Fingerprint class="h-3.5 w-3.5" />
            <span class="audit-actor-detail__label">用户 ID</span>
          </div>
          <strong class="audit-actor-detail__value audit-actor-detail__value--mono">
            {{ item.actor_user_id || '-' }}
          </strong>
        </article>

        <article class="audit-actor-detail__item">
          <div class="audit-actor-detail__head">
            <Activity class="h-3.5 w-3.5" />
            <span class="audit-actor-detail__label">动作</span>
          </div>
          <strong class="audit-actor-detail__value">{{ item.action }}</strong>
        </article>

        <article class="audit-actor-detail__item">
          <div class="audit-actor-detail__head">
            <Clock class="h-3.5 w-3.5" />
            <span class="audit-actor-detail__label">发生时间</span>
          </div>
          <strong class="audit-actor-detail__value">
            {{ formatDate(item.created_at) }}
          </strong>
        </article>

        <article class="audit-actor-detail__item">
          <div class="audit-actor-detail__head">
            <Package class="h-3.5 w-3.5" />
            <span class="audit-actor-detail__label">目标资源</span>
          </div>
          <strong class="audit-actor-detail__value">
            {{ resourceDisplayName(item) }}
          </strong>
        </article>

        <article class="audit-actor-detail__item audit-actor-detail__item--wide">
          <div class="audit-actor-detail__head">
            <FileJson class="h-3.5 w-3.5" />
            <span class="audit-actor-detail__label">明细上下文</span>
          </div>
          <p class="audit-actor-detail__detail">
            {{ detailPreview(item.detail) }}
          </p>
        </article>
      </div>
    </section>
  </AdminSurfaceModal>
</template>

<style scoped>
.audit-actor-detail__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.75rem 2rem;
  padding: 0.25rem;
}

.audit-actor-detail__item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.audit-actor-detail__head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-text-muted);
}

.audit-actor-detail__label {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.audit-actor-detail__value {
  font-size: var(--font-size-16);
  font-weight: 800;
  line-height: 1.2;
  color: var(--color-text-primary);
}

.audit-actor-detail__item--wide {
  grid-column: 1 / -1;
}

.audit-actor-detail__value--mono {
  font-family: var(--font-family-mono);
}

.audit-actor-detail__detail {
  margin: 0;
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

@media (max-width: 720px) {
  .audit-actor-detail__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
