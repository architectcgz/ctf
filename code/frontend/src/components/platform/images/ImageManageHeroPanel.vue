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
  <header class="image-header">
    <div class="image-header__intro">
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

    <div class="image-header__side">
      <div
        class="image-header__actions"
        role="group"
        aria-label="镜像列表操作"
      >
        <button
          :disabled="loading"
          class="ui-btn ui-btn--ghost"
          data-testid="image-refresh-button"
          @click="handleRefresh"
        >
          立即刷新
        </button>
        <button
          class="ui-btn ui-btn--primary"
          @click="handleCreate"
        >
          创建镜像
        </button>
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
  </header>
</template>

<style scoped>
.image-header__actions > .ui-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
  --ui-btn-primary-hover-shadow: 0 10px 24px color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
}

.image-header {
  --image-toolbar-control-border: var(--color-border-default);
  --image-toolbar-control-border-strong: color-mix(in srgb, var(--color-border-default) 80%, var(--color-text-primary));
  --image-toolbar-control-background: var(--color-bg-surface);
  --image-toolbar-control-shadow: var(--color-shadow-soft);
}

:global([data-theme='dark']) .image-header {
  --image-toolbar-control-border: color-mix(in srgb, var(--color-border-default) 72%, transparent);
  --image-toolbar-control-border-strong: color-mix(in srgb, var(--color-primary) 32%, var(--color-border-default));
  --image-toolbar-control-background: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
}

.image-header__actions > [data-testid='image-refresh-button'] {
  --ui-btn-border: var(--image-toolbar-control-border);
  --ui-btn-background: var(--image-toolbar-control-background);
  --ui-btn-color: var(--color-text-primary);
  --ui-btn-hover-border: var(--image-toolbar-control-border-strong);
  --ui-btn-hover-background: var(--image-toolbar-control-background);
  --ui-btn-hover-color: var(--color-primary);
  box-shadow: var(--image-toolbar-control-shadow);
}

.image-header {
  display: grid;
  gap: var(--space-6);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.image-copy {
  max-width: 48rem;
}

.image-header__side {
  display: grid;
  gap: var(--space-3);
  justify-items: start;
}

.image-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.image-status-strip {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3) var(--space-4);
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
  .image-status-strip {
    align-items: flex-start;
  }

  .image-status-strip__note {
    width: 100%;
  }
}
</style>
