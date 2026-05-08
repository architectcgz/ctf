<script setup lang="ts">
import { computed } from 'vue'
import { AlignLeft, X } from 'lucide-vue-next'

import ModalTemplateShell from './ModalTemplateShell.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    subtitle?: string
    eyebrow?: string
    width?: string
    bodyPadding?: string
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    eyebrow: '',
    width: '26.25rem',
    bodyPadding: 'var(--modal-template-drawer-default-body-padding, 0 var(--space-7))',
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()

const overlayClass = 'modal-template-shell--drawer'
// 将布局参数和视觉参数统一通过 CSS 变量传递
const panelStyle = computed<Record<string, string>>(() => ({
  '--modal-template-drawer-width': props.width,
  '--modal-template-drawer-body-padding': props.bodyPadding,
}))

function forwardOpen(value: boolean): void {
  emit('update:open', value)
}

function forwardClose(): void {
  emit('close')
}

function handleCloseClick(): void {
  forwardClose()
  forwardOpen(false)
}
</script>

<template>
  <ModalTemplateShell
    :open="open"
    :frosted="true"
    :panel-style="panelStyle"
    :overlay-class="overlayClass"
    panel-class="modal-template-panel--drawer"
    panel-tag="aside"
    :aria-label="title"
    :close-on-backdrop="closeOnBackdrop"
    :close-on-escape="closeOnEscape"
    :style="{
      '--modal-shell-justify': 'flex-end',
      '--modal-shell-align': 'stretch',
      '--modal-shell-padding': '0',
      '--modal-shell-blur': '12px',
      '--modal-template-shell-overlay': 'color-mix(in srgb, var(--color-bg-base) 18%, transparent)',
    }"
    @update:open="forwardOpen"
    @close="forwardClose"
  >
    <div class="modal-template-drawer">
      <!-- 关闭按钮：移至最右上角，圆形，带背景 -->
      <button
        type="button"
        class="modal-template-drawer__close"
        aria-label="关闭抽屉"
        @click="handleCloseClick"
      >
        <X class="modal-template-drawer__close-glyph" />
      </button>

      <header class="modal-template-drawer__header">
        <div class="modal-template-drawer__head-row">
          <div class="modal-template-drawer__head-main">
            <!-- 头部图标：圆形背景 + 细边框 -->
            <div class="modal-template-drawer__icon">
              <slot name="icon">
                <AlignLeft class="modal-template-drawer__icon-glyph" />
              </slot>
            </div>
            <div class="modal-template-drawer__title-block">
              <p v-if="eyebrow" class="modal-template-drawer__eyebrow">
                {{ eyebrow }}
              </p>
              <h2 class="modal-template-drawer__title">
                {{ title }}
              </h2>
            </div>
          </div>
        </div>

        <!-- 副标题区域：常用于展示“未读/总计” -->
        <div v-if="subtitle || $slots.subtitle" class="modal-template-drawer__subtitle-area">
          <slot name="subtitle">
            <p class="modal-template-drawer__subtitle">
              {{ subtitle }}
            </p>
          </slot>
        </div>

        <!-- 额外头部内容：如 Tab 切换 -->
        <div v-if="$slots['header-extra']" class="modal-template-drawer__header-extra">
          <slot name="header-extra" />
        </div>
      </header>

      <!-- 分割线 -->
      <div class="modal-template-drawer__divider" />

      <!-- 主体列表 -->
      <div class="modal-template-drawer__body">
        <slot />
      </div>

      <!-- 底部操作：通常是一个浮动风格的按钮/卡片 -->
      <footer v-if="$slots.footer" class="modal-template-drawer__footer">
        <slot name="footer" />
      </footer>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.modal-template-shell--drawer {
  --modal-template-drawer-accent: var(--drawer-accent, var(--color-primary));
  --modal-template-drawer-line: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --modal-template-drawer-line-strong: color-mix(
    in srgb,
    var(--color-border-default) 92%,
    transparent
  );
  --modal-template-drawer-surface: color-mix(
    in srgb,
    var(--color-bg-surface) 96%,
    var(--color-bg-base)
  );
  --modal-template-drawer-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 88%,
    var(--color-bg-surface)
  );
  --modal-template-drawer-surface-muted: color-mix(
    in srgb,
    var(--color-bg-surface) 88%,
    var(--color-bg-base)
  );
  --modal-template-drawer-text: color-mix(in srgb, var(--color-text-primary) 96%, transparent);
  --modal-template-drawer-muted: color-mix(in srgb, var(--color-text-secondary) 94%, transparent);
  --modal-template-drawer-faint: color-mix(in srgb, var(--color-text-muted) 94%, transparent);
  --modal-template-drawer-radius: calc(var(--ui-dialog-radius-wide) + var(--space-0-5));
  --modal-template-drawer-header-padding-block-start: var(--space-8);
  --modal-template-drawer-header-padding-inline: var(--space-7);
  --modal-template-drawer-header-padding-block-end: var(--space-5);
  --modal-template-drawer-header-extra-margin-top: var(--space-6);
  --modal-template-drawer-divider-margin-inline: var(--space-7);
  --modal-template-drawer-footer-padding: var(--space-6) var(--space-7) var(--space-7);
  --modal-template-drawer-icon-size: calc(var(--space-5-5) * 2);
  --modal-template-drawer-icon-glyph-size: var(--space-5);
  --modal-template-drawer-close-size: calc(var(--space-4-5) * 2);
  --modal-template-drawer-close-glyph-size: calc(var(--space-4-5) * 0.9);
  --modal-template-drawer-close-offset: var(--space-7);
  --modal-template-drawer-title-size: var(--font-size-1-80);
  --modal-template-drawer-title-line-height: 1.15;
  --modal-template-drawer-title-spacing: -0.04em;
  --modal-template-drawer-eyebrow-size: var(--font-size-11);
  --modal-template-drawer-subtitle-size: var(--font-size-15);
  --modal-template-drawer-panel-shadow:
    calc(var(--space-6) * -1) 0 var(--space-12)
      color-mix(in srgb, var(--color-shadow-strong) 18%, transparent),
    calc(var(--space-0-5) * -1) 0 0 color-mix(in srgb, var(--color-border-subtle) 72%, transparent);
  --modal-template-drawer-panel-border: 1px solid var(--modal-template-drawer-line-strong);
  --modal-template-drawer-header-surface: var(--modal-template-drawer-surface);
  --modal-template-drawer-body-surface: var(--modal-template-drawer-surface);
  --modal-template-drawer-footer-surface: var(--modal-template-drawer-surface);
  --modal-template-drawer-icon-surface: linear-gradient(
    180deg,
    color-mix(
      in srgb,
      var(--modal-template-drawer-accent) 12%,
      var(--modal-template-drawer-surface)
    ),
    color-mix(
      in srgb,
      var(--modal-template-drawer-accent) 6%,
      var(--modal-template-drawer-surface-subtle)
    )
  );
  --modal-template-drawer-icon-border: 1px solid
    color-mix(in srgb, var(--modal-template-drawer-accent) 18%, var(--modal-template-drawer-line));
  --modal-template-drawer-icon-color: color-mix(
    in srgb,
    var(--modal-template-drawer-accent) 88%,
    var(--modal-template-drawer-text)
  );
  --modal-template-drawer-icon-shadow: 0 var(--space-3) var(--space-7)
    color-mix(in srgb, var(--modal-template-drawer-accent) 12%, transparent);
  --modal-template-drawer-close-surface: color-mix(
    in srgb,
    var(--modal-template-drawer-surface-muted) 96%,
    transparent
  );
  --modal-template-drawer-close-border: 1px solid transparent;
  --modal-template-drawer-close-color: var(--modal-template-drawer-faint);
  --modal-template-drawer-close-hover-surface: color-mix(
    in srgb,
    var(--modal-template-drawer-line) 24%,
    var(--modal-template-drawer-surface-subtle)
  );
  --modal-template-drawer-close-hover-color: var(--modal-template-drawer-text);
  --modal-template-drawer-close-hover-transform: rotate(90deg);
}

:deep(.modal-template-panel--drawer) {
  width: min(100%, var(--modal-template-drawer-width));
  max-width: 100%;
  height: 100%;
  background-color: var(--modal-template-drawer-surface);
  border: var(--modal-template-drawer-panel-border);
  border-right: none;
  border-top-left-radius: var(--modal-template-drawer-radius);
  border-bottom-left-radius: var(--modal-template-drawer-radius);
  box-shadow: var(--modal-template-drawer-panel-shadow);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-template-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  position: relative;
}

.modal-template-drawer__header {
  padding: var(--modal-template-drawer-header-padding-block-start)
    calc(
      var(--modal-template-drawer-header-padding-inline) + var(--modal-template-drawer-close-size) +
        var(--space-3)
    )
    var(--modal-template-drawer-header-padding-block-end)
    var(--modal-template-drawer-header-padding-inline);
  position: relative;
  background: var(--modal-template-drawer-header-surface);
}

.modal-template-drawer__head-row {
  display: flex;
  align-items: center;
  margin-bottom: var(--space-5);
}

.modal-template-drawer__head-main {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  min-width: 0;
}

.modal-template-drawer__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--modal-template-drawer-icon-size);
  height: var(--modal-template-drawer-icon-size);
  border-radius: 999px;
  flex-shrink: 0;
  background: var(--modal-template-drawer-icon-surface);
  border: var(--modal-template-drawer-icon-border);
  color: var(--modal-template-drawer-icon-color);
  box-shadow: var(--modal-template-drawer-icon-shadow);
}

.modal-template-drawer__icon-glyph,
.modal-template-drawer__close-glyph {
  width: var(--modal-template-drawer-icon-glyph-size);
  height: var(--modal-template-drawer-icon-glyph-size);
}

.modal-template-drawer__close-glyph {
  width: var(--modal-template-drawer-close-glyph-size);
  height: var(--modal-template-drawer-close-glyph-size);
}

.modal-template-drawer__close {
  position: absolute;
  right: var(--modal-template-drawer-close-offset);
  top: var(--modal-template-drawer-close-offset);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--modal-template-drawer-close-size);
  height: var(--modal-template-drawer-close-size);
  border: var(--modal-template-drawer-close-border);
  border-radius: 999px;
  background: var(--modal-template-drawer-close-surface);
  color: var(--modal-template-drawer-close-color);
  transition:
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast),
    transform var(--ui-motion-fast),
    border-color var(--ui-motion-fast);
  cursor: pointer;
  z-index: 20;
}

.modal-template-drawer__close:hover {
  background: var(--modal-template-drawer-close-hover-surface);
  color: var(--modal-template-drawer-close-hover-color);
  transform: var(--modal-template-drawer-close-hover-transform);
}

.modal-template-drawer__close:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(
      in srgb,
      var(--modal-template-drawer-accent) 44%,
      var(--modal-template-drawer-line-strong)
    );
  outline-offset: var(--space-0-5);
}

.modal-template-drawer__title-block {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
  flex: 1;
}

.modal-template-drawer__title {
  margin: 0;
  font-size: var(--modal-template-drawer-title-size);
  font-weight: 800;
  letter-spacing: var(--modal-template-drawer-title-spacing);
  color: var(--modal-template-drawer-text);
  line-height: var(--modal-template-drawer-title-line-height);
}

.modal-template-drawer__eyebrow {
  margin: 0;
  font-size: var(--modal-template-drawer-eyebrow-size);
  font-weight: 800;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--modal-template-drawer-faint);
}

.modal-template-drawer__subtitle-area {
  margin-top: var(--space-1);
}

.modal-template-drawer__subtitle {
  margin: 0;
  font-size: var(--modal-template-drawer-subtitle-size);
  font-weight: 500;
  color: var(--modal-template-drawer-muted);
  display: flex;
  align-items: baseline;
  gap: var(--space-1);
  line-height: 1.6;
}

.modal-template-drawer__header-extra {
  margin-top: var(--modal-template-drawer-header-extra-margin-top);
}

.modal-template-drawer__divider {
  margin: 0 var(--modal-template-drawer-divider-margin-inline);
  height: var(--space-0-5);
  background-color: var(--modal-template-drawer-line);
}

.modal-template-drawer__body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: var(--modal-template-drawer-body-padding);
  background: var(--modal-template-drawer-body-surface);
}

.modal-template-drawer__body::-webkit-scrollbar {
  width: var(--space-1);
}

.modal-template-drawer__body::-webkit-scrollbar-thumb {
  background: var(--modal-template-drawer-line);
  border-radius: var(--ui-badge-radius-pill);
}

.modal-template-drawer__footer {
  padding: var(--modal-template-drawer-footer-padding);
  background: var(--modal-template-drawer-footer-surface);
  position: relative;
}

/* 针对 footer 中常见的浮动按钮卡片样式 */
:deep(.drawer-footer-action) {
  display: flex;
  width: 100%;
  min-height: var(--ui-control-height-lg);
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-5);
  background: var(--modal-template-drawer-surface);
  border: 1px solid var(--modal-template-drawer-line);
  border-radius: var(--ui-control-radius-lg);
  font-size: var(--font-size-15);
  font-weight: 600;
  color: var(--modal-template-drawer-muted);
  box-shadow: 0 var(--space-2) var(--space-6)
    color-mix(in srgb, var(--color-shadow-soft) 16%, transparent);
  transition:
    border-color var(--ui-motion-fast),
    background-color var(--ui-motion-fast),
    color var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
  cursor: pointer;
}

:deep(.drawer-footer-action:hover) {
  border-color: color-mix(
    in srgb,
    var(--modal-template-drawer-accent) 24%,
    var(--modal-template-drawer-line-strong)
  );
  background: color-mix(
    in srgb,
    var(--modal-template-drawer-accent) 6%,
    var(--modal-template-drawer-surface-subtle)
  );
  color: color-mix(
    in srgb,
    var(--modal-template-drawer-accent) 82%,
    var(--modal-template-drawer-text)
  );
}

:deep(.drawer-footer-action:focus-visible) {
  outline: var(--ui-focus-ring-width) solid
    color-mix(
      in srgb,
      var(--modal-template-drawer-accent) 44%,
      var(--modal-template-drawer-line-strong)
    );
  outline-offset: var(--space-0-5);
}

@media (max-width: 768px) {
  .modal-template-shell--drawer {
    --modal-template-drawer-header-padding-block-start: var(--space-6);
    --modal-template-drawer-header-padding-inline: var(--space-5);
    --modal-template-drawer-header-padding-block-end: var(--space-4);
    --modal-template-drawer-header-extra-margin-top: var(--space-4);
    --modal-template-drawer-divider-margin-inline: var(--space-5);
    --modal-template-drawer-footer-padding: var(--space-5);
    --modal-template-drawer-close-offset: var(--space-5);
    --modal-template-drawer-title-size: var(--font-size-1-45);
  }

  :deep(.modal-template-panel--drawer) {
    border-top-left-radius: var(--ui-dialog-radius-wide);
    border-bottom-left-radius: var(--ui-dialog-radius-wide);
  }
}
</style>
