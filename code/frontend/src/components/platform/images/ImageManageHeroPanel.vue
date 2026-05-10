<script setup lang="ts">
defineProps<{
  loading: boolean
  refreshHint: string
  statusSummary: Array<{
    key: string
    label: string
    value: number
    tone: 'success' | 'warning' | 'danger' | 'muted'
  }>
}>()

const emit = defineEmits<{
  refresh: []
  create: []
}>()

function handleRefresh(): void {
  emit('refresh')
}

function handleCreate(): void {
  emit('create')
}
</script>

<template>
  <header class="workspace-page-header image-header">
    <div class="image-header__intro">
      <div class="image-header__copy">
        <div class="workspace-overline">
          Image Registry
        </div>
        <h1 class="image-title">
          镜像管理
        </h1>
        <p class="image-copy">
          集中查看镜像构建状态、描述与创建时间。
        </p>
      </div>

      <div
        class="image-status-strip"
        aria-label="镜像状态摘要"
      >
        <div
          v-if="statusSummary.length > 0"
          class="image-status-strip__row"
        >
          <div
            v-for="item in statusSummary"
            :key="item.key"
            :class="['image-status-pill', `image-status-pill--${item.tone}`]"
            data-testid="image-status-pill"
          >
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>
        <div class="image-status-strip__note">{{ refreshHint }}</div>
      </div>
    </div>

    <div class="image-header__side">
      <div
        class="header-actions image-header__actions"
        role="group"
        aria-label="镜像列表操作"
      >
        <button
          :disabled="loading"
          class="header-btn header-btn--ghost"
          data-testid="image-refresh-button"
          @click="handleRefresh"
        >
          立即刷新
        </button>
        <button
          class="header-btn header-btn--primary"
          @click="handleCreate"
        >
          创建镜像
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.image-copy {
  max-width: 48rem;
}

.image-header__intro {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(18rem, auto);
  align-items: start;
  gap: var(--space-5);
}

.image-header__copy {
  min-width: 0;
}

.image-header__side {
  display: grid;
  gap: var(--space-3);
  justify-items: start;
}

.image-status-strip {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3) var(--space-4);
  justify-self: end;
  max-width: 34rem;
}

.image-status-strip__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
}

.image-status-strip__note {
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
}

.image-status-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
  line-height: 1;
}

.image-status-pill strong {
  font-size: var(--font-size-0-9);
  font-weight: 700;
  color: var(--journal-ink);
}

.image-status-pill--success {
  border-color: color-mix(in srgb, var(--color-success) 22%, transparent);
  background: color-mix(in srgb, var(--color-success) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-success) 82%, var(--journal-ink));
}

.image-status-pill--warning {
  border-color: color-mix(in srgb, var(--color-warning) 24%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-warning) 84%, var(--journal-ink));
}

.image-status-pill--danger {
  border-color: color-mix(in srgb, var(--color-danger) 24%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.image-status-pill--muted {
  border-color: color-mix(in srgb, var(--journal-muted) 18%, transparent);
  background: color-mix(in srgb, var(--journal-muted) 10%, var(--journal-surface));
  color: var(--journal-muted);
}

@media (max-width: 720px) {
  .image-header__intro {
    grid-template-columns: minmax(0, 1fr);
  }

  .image-status-strip {
    align-items: flex-start;
    justify-content: flex-start;
    justify-self: stretch;
  }

  .image-status-strip__note {
    width: 100%;
  }
}
</style>
