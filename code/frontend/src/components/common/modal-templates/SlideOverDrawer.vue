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
    bodyPadding: '0 var(--space-8)',
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
      '--modal-template-shell-overlay': 'color-mix(in srgb, var(--color-bg-base) 18%, transparent)'
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
        @click="forwardClose(), forwardOpen(false)"
      >
        <X class="h-[18px] w-[18px]" />
      </button>

      <header class="modal-template-drawer__header">
        <div class="modal-template-drawer__head-row">
          <div class="modal-template-drawer__head-main">
            <!-- 头部图标：圆形背景 + 细边框 -->
            <div class="modal-template-drawer__icon">
              <slot name="icon">
                <AlignLeft class="h-5 w-5" />
              </slot>
            </div>
            <div class="modal-template-drawer__title-block">
              <p
                v-if="eyebrow"
                class="modal-template-drawer__eyebrow"
              >
                {{ eyebrow }}
              </p>
              <h2 class="modal-template-drawer__title">
                {{ title }}
              </h2>
            </div>
          </div>
        </div>

        <!-- 副标题区域：常用于展示“未读/总计” -->
        <div
          v-if="subtitle || $slots.subtitle"
          class="modal-template-drawer__subtitle-area"
        >
          <slot name="subtitle">
            <p class="modal-template-drawer__subtitle">
              {{ subtitle }}
            </p>
          </slot>
        </div>

        <!-- 额外头部内容：如 Tab 切换 -->
        <div
          v-if="$slots['header-extra']"
          class="modal-template-drawer__header-extra"
        >
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
      <footer
        v-if="$slots.footer"
        class="modal-template-drawer__footer"
      >
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
}

:deep(.modal-template-panel--drawer) {
  width: var(--modal-template-drawer-width);
  max-width: 100%;
  height: 100%;
  background-color: var(--modal-template-drawer-surface);
  border-top-left-radius: 28px;
  border-bottom-left-radius: 28px;
  box-shadow: -20px 0 60px color-mix(in srgb, var(--color-shadow-strong) 18%, transparent);
  display: flex;
  flex-direction: column;
}

.modal-template-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  position: relative;
}

.modal-template-drawer__header {
  padding: 36px 32px 20px;
  position: relative;
}

.modal-template-drawer__head-row {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.modal-template-drawer__head-main {
  display: flex;
  align-items: center;
  gap: 16px;
}

.modal-template-drawer__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 999px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--modal-template-drawer-accent) 12%, var(--modal-template-drawer-surface)),
    color-mix(in srgb, var(--modal-template-drawer-accent) 6%, var(--modal-template-drawer-surface-subtle))
  );
  border: 1px solid
    color-mix(in srgb, var(--modal-template-drawer-accent) 18%, var(--modal-template-drawer-line));
  color: color-mix(in srgb, var(--modal-template-drawer-accent) 88%, var(--modal-template-drawer-text));
  box-shadow: 0 10px 24px color-mix(in srgb, var(--modal-template-drawer-accent) 12%, transparent);
}

.modal-template-drawer__close {
  position: absolute;
  right: 28px;
  top: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 999px;
  background: color-mix(in srgb, var(--modal-template-drawer-surface-muted) 96%, transparent);
  color: var(--modal-template-drawer-faint);
  transition:
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
  cursor: pointer;
  z-index: 20;
}

.modal-template-drawer__close:hover {
  background: color-mix(in srgb, var(--modal-template-drawer-line) 24%, var(--modal-template-drawer-surface-subtle));
  color: var(--modal-template-drawer-text);
  transform: rotate(90deg);
}

.modal-template-drawer__close:focus-visible {
  outline: 2px solid
    color-mix(in srgb, var(--modal-template-drawer-accent) 44%, var(--modal-template-drawer-line-strong));
  outline-offset: 2px;
}

.modal-template-drawer__title-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.modal-template-drawer__title {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--modal-template-drawer-text);
  line-height: 1.2;
}

.modal-template-drawer__eyebrow {
  margin: 0;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--modal-template-drawer-faint);
}

.modal-template-drawer__subtitle-area {
  margin-top: 4px;
}

.modal-template-drawer__subtitle {
  margin: 0;
  font-size: 15px;
  font-weight: 500;
  color: var(--modal-template-drawer-muted);
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.modal-template-drawer__header-extra {
  margin-top: 24px;
}

.modal-template-drawer__divider {
  margin: 0 32px;
  height: 1px;
  background-color: var(--modal-template-drawer-line);
}

.modal-template-drawer__body {
  flex: 1;
  overflow-y: auto;
  padding: var(--modal-template-drawer-body-padding);
}

.modal-template-drawer__body::-webkit-scrollbar {
  width: 4px;
}

.modal-template-drawer__body::-webkit-scrollbar-thumb {
  background: var(--modal-template-drawer-line);
  border-radius: 10px;
}

.modal-template-drawer__footer {
  padding: 32px;
  background: var(--modal-template-drawer-surface);
  position: relative;
}

/* 针对 footer 中常见的浮动按钮卡片样式 */
:deep(.drawer-footer-action) {
  display: flex;
  width: 100%;
  height: 58px;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: var(--modal-template-drawer-surface);
  border: 1px solid var(--modal-template-drawer-line);
  border-radius: 16px;
  font-size: 15px;
  font-weight: 600;
  color: var(--modal-template-drawer-muted);
  box-shadow: 0 8px 24px color-mix(in srgb, var(--color-shadow-soft) 16%, transparent);
  transition: all 0.2s;
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
  color: color-mix(in srgb, var(--modal-template-drawer-accent) 82%, var(--modal-template-drawer-text));
}

:deep(.drawer-footer-action:focus-visible) {
  outline: 2px solid
    color-mix(in srgb, var(--modal-template-drawer-accent) 44%, var(--modal-template-drawer-line-strong));
  outline-offset: 2px;
}
</style>
