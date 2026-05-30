<script setup lang="ts">
import {
  Clock,
  Database,
  FileText,
  Fingerprint,
  Info,
  Maximize2,
  Tag,
} from 'lucide-vue-next'

import type { AdminImageListItem, ImageStatus } from '@/api/contracts'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'

const props = defineProps<{
  open: boolean
  image: AdminImageListItem | null
  formatSize: (bytes?: number) => string
  formatDateTime: (value: string) => string
  getStatusLabel: (status: ImageStatus) => string
  getStatusStyle: (status: ImageStatus) => Record<string, string>
}>()

const emit = defineEmits<{
  close: []
  'update:open': [value: boolean]
}>()
</script>

<template>
  <AdminSurfaceModal
    class="image-detail-modal"
    :open="open"
    :frosted="true"
    title="镜像详情"
    eyebrow="Image Registry"
    width="34rem"
    @close="emit('close')"
    @update:open="emit('update:open', $event)"
  >
    <section
      v-if="image"
      class="image-detail"
    >
      <div class="image-detail__grid">
        <article class="image-detail__item">
          <div class="image-detail__head">
            <Database class="h-3.5 w-3.5" />
            <span class="image-detail__label">镜像名称</span>
          </div>
          <strong class="image-detail__value">{{ image.name }}</strong>
        </article>

        <article class="image-detail__item">
          <div class="image-detail__head">
            <Tag class="h-3.5 w-3.5" />
            <span class="image-detail__label">标签版本</span>
          </div>
          <strong class="image-detail__value">{{ image.tag }}</strong>
        </article>

        <article class="image-detail__item">
          <div class="image-detail__head">
            <Fingerprint class="h-3.5 w-3.5" />
            <span class="image-detail__label">镜像 ID</span>
          </div>
          <strong class="image-detail__value image-detail__value--mono">
            {{ image.id }}
          </strong>
        </article>

        <article class="image-detail__item">
          <div class="image-detail__head">
            <Info class="h-3.5 w-3.5" />
            <span class="image-detail__label">状态</span>
          </div>
          <div class="image-detail__value">
            <span
              class="admin-status-chip"
              :style="getStatusStyle(image.status)"
            >
              {{ getStatusLabel(image.status) }}
            </span>
          </div>
        </article>

        <article class="image-detail__item">
          <div class="image-detail__head">
            <Maximize2 class="h-3.5 w-3.5" />
            <span class="image-detail__label">占用空间</span>
          </div>
          <strong class="image-detail__value">{{ formatSize(image.size_bytes) }}</strong>
        </article>

        <article class="image-detail__item">
          <div class="image-detail__head">
            <Clock class="h-3.5 w-3.5" />
            <span class="image-detail__label">最后更新</span>
          </div>
          <strong class="image-detail__value">
            {{ formatDateTime(image.updated_at || image.created_at) }}
          </strong>
        </article>

        <article class="image-detail__item image-detail__item--wide">
          <div class="image-detail__head">
            <FileText class="h-3.5 w-3.5" />
            <span class="image-detail__label">描述信息</span>
          </div>
          <p class="image-detail__description">
            {{ image.description || '未提供详细描述' }}
          </p>
        </article>
      </div>
    </section>
  </AdminSurfaceModal>
</template>

<style scoped>
.admin-status-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 0.5rem;
  padding: var(--space-1) var(--space-2-5);
  font-size: var(--font-size-0-72);
  font-weight: 600;
}

.image-detail__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.75rem 2rem;
  padding: 0.25rem;
}

.image-detail__item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.image-detail__item--wide {
  grid-column: 1 / -1;
}

.image-detail__head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--journal-muted);
}

.image-detail__label {
  font-size: var(--font-size-0-625);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.image-detail__value {
  font-size: var(--font-size-1-00);
  font-weight: 800;
  line-height: 1.2;
  color: var(--journal-ink);
}

.image-detail__value--mono {
  font-family: var(--font-family-mono);
}

.image-detail__description {
  margin: 0;
  font-size: var(--font-size-0-875);
  line-height: 1.7;
  color: var(--journal-muted);
}
</style>
