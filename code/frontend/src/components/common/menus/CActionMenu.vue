<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch, type ComponentPublicInstance } from 'vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    menuLabel?: string
    width?: string
    minWidth?: string
    align?: 'start' | 'end'
    gap?: number
    viewportPadding?: number
    closeOnBackdrop?: boolean
    closeOnEscape?: boolean
  }>(),
  {
    title: '',
    menuLabel: '更多操作',
    width: '11rem',
    minWidth: '',
    align: 'end',
    gap: 8,
    viewportPadding: 12,
    closeOnBackdrop: true,
    closeOnEscape: true,
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
}>()

const triggerRef = ref<HTMLElement | null>(null)
const panelRef = ref<HTMLElement | null>(null)
const panelPositionStyle = ref<Record<string, string>>({})

const resolvedPanelStyle = computed<Record<string, string>>(() => {
  const style: Record<string, string> = {
    ...panelPositionStyle.value,
  }

  if (props.width) {
    style.width = props.width
  }
  if (props.minWidth) {
    style.minWidth = props.minWidth
  }

  return style
})

function setTriggerRef(element: Element | ComponentPublicInstance | null): void {
  if (element instanceof HTMLElement) {
    triggerRef.value = element
    return
  }

  triggerRef.value = null
}

function closeMenu(): void {
  emit('update:open', false)
  emit('close')
}

function toggleMenu(): void {
  emit('update:open', !props.open)
}

function handleBackdropClick(): void {
  if (!props.closeOnBackdrop) {
    return
  }

  closeMenu()
}

function handleWindowKeydown(event: KeyboardEvent): void {
  if (!props.open || !props.closeOnEscape || event.key !== 'Escape') {
    return
  }

  closeMenu()
}

function updatePanelPosition(): void {
  if (!props.open || !triggerRef.value) {
    return
  }

  const rect = triggerRef.value.getBoundingClientRect()
  const panelWidth = panelRef.value?.offsetWidth ?? 176
  const panelHeight = panelRef.value?.offsetHeight ?? 132
  const maxLeft = Math.max(
    props.viewportPadding,
    window.innerWidth - panelWidth - props.viewportPadding
  )
  const baseLeft = props.align === 'start' ? rect.left : rect.right - panelWidth
  const left = Math.min(Math.max(props.viewportPadding, baseLeft), maxLeft)
  const spaceBelow = window.innerHeight - rect.bottom - props.viewportPadding
  const spaceAbove = rect.top - props.viewportPadding
  const shouldOpenUpward = spaceBelow < panelHeight + props.gap && spaceAbove > spaceBelow
  const maxTop = Math.max(
    props.viewportPadding,
    window.innerHeight - panelHeight - props.viewportPadding
  )
  const top = shouldOpenUpward
    ? Math.max(props.viewportPadding, rect.top - panelHeight - props.gap)
    : Math.min(rect.bottom + props.gap, maxTop)

  panelPositionStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
  }
}

watch(
  () => props.open,
  async (open, _previousOpen, onCleanup) => {
    if (!open) {
      panelPositionStyle.value = {}
      return
    }

    await nextTick()
    updatePanelPosition()

    const handleViewportChange = () => {
      updatePanelPosition()
    }

    window.addEventListener('resize', handleViewportChange)
    window.addEventListener('scroll', handleViewportChange, true)
    window.addEventListener('keydown', handleWindowKeydown)

    onCleanup(() => {
      window.removeEventListener('resize', handleViewportChange)
      window.removeEventListener('scroll', handleViewportChange, true)
      window.removeEventListener('keydown', handleWindowKeydown)
    })
  }
)

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleWindowKeydown)
})
</script>

<template>
  <div class="c-action-menu">
    <slot
      name="trigger"
      :open="open"
      :toggle="toggleMenu"
      :close="closeMenu"
      :set-trigger-ref="setTriggerRef"
    />

    <Teleport to="body">
      <div
        v-if="open"
        class="c-action-menu__layer"
        data-action-menu-layer
        @click="handleBackdropClick"
      >
        <div
          ref="panelRef"
          class="c-action-menu__panel"
          data-action-menu-panel
          :style="resolvedPanelStyle"
          role="menu"
          :aria-label="menuLabel"
          @click.stop
        >
          <div v-if="title" class="c-action-menu__title">{{ title }}</div>
          <div class="c-action-menu__content">
            <slot :close="closeMenu" />
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.c-action-menu,
.c-action-menu__layer {
  --c-action-menu-surface: var(
    --action-menu-surface,
    var(
      --journal-surface,
      color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base))
    )
  );
  --c-action-menu-surface-subtle: var(
    --action-menu-surface-subtle,
    var(
      --journal-surface-subtle,
      color-mix(in srgb, var(--color-bg-elevated) 88%, var(--color-bg-surface))
    )
  );
  --c-action-menu-line: var(
    --action-menu-line,
    color-mix(in srgb, var(--color-border-default) 88%, transparent)
  );
  --c-action-menu-line-strong: var(
    --action-menu-line-strong,
    color-mix(in srgb, var(--color-border-default) 94%, transparent)
  );
  --c-action-menu-text: var(
    --action-menu-text,
    color-mix(in srgb, var(--color-text-primary) 94%, transparent)
  );
  --c-action-menu-muted: var(
    --action-menu-muted,
    color-mix(in srgb, var(--color-text-secondary) 92%, transparent)
  );
  --c-action-menu-accent: var(--action-menu-accent, var(--color-primary));
  --c-action-menu-accent-soft: color-mix(
    in srgb,
    var(--c-action-menu-accent) 10%,
    var(--c-action-menu-surface)
  );
}

.c-action-menu {
  display: inline-flex;
}

.c-action-menu__layer {
  position: fixed;
  inset: 0;
  z-index: 120;
}

.c-action-menu__panel {
  position: fixed;
  z-index: 130;
  overflow: hidden;
  border: 1px solid var(--c-action-menu-line);
  border-radius: 1rem;
  background-color: var(--c-action-menu-surface);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--c-action-menu-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--c-action-menu-surface-subtle) 96%, var(--color-bg-base))
  );
  box-shadow:
    0 24px 60px color-mix(in srgb, var(--color-shadow-strong) 18%, transparent),
    0 10px 24px color-mix(in srgb, var(--color-shadow-soft) 16%, transparent);
}

.c-action-menu__title {
  padding: 0.78rem 1rem 0.55rem;
  border-bottom: 1px solid color-mix(in srgb, var(--c-action-menu-line) 78%, transparent);
  background: color-mix(in srgb, var(--c-action-menu-accent) 5%, var(--c-action-menu-surface-subtle));
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--c-action-menu-muted);
}

.c-action-menu__content {
  display: grid;
}

.c-action-menu__content :deep(.c-action-menu__item) {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.5rem;
  justify-content: flex-start;
  border: 0;
  background: transparent;
  padding: 0.72rem 1rem;
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: var(--c-action-menu-text);
  transition:
    background-color 160ms ease,
    color 160ms ease;
}

.c-action-menu__content :deep(.c-action-menu__item:hover) {
  background: color-mix(in srgb, var(--c-action-menu-accent) 7%, var(--c-action-menu-surface-subtle));
  color: var(--c-action-menu-accent);
}

.c-action-menu__content :deep(.c-action-menu__item--success) {
  color: color-mix(in srgb, var(--color-success) 84%, var(--c-action-menu-text));
}

.c-action-menu__content :deep(.c-action-menu__item--success:hover) {
  background: color-mix(in srgb, var(--color-success) 10%, var(--c-action-menu-surface-subtle));
  color: color-mix(in srgb, var(--color-success) 92%, var(--c-action-menu-text));
}

.c-action-menu__content :deep(.c-action-menu__item--danger) {
  color: color-mix(in srgb, var(--color-danger) 88%, var(--c-action-menu-text));
}

.c-action-menu__content :deep(.c-action-menu__item--danger:hover) {
  background: color-mix(in srgb, var(--color-danger) 10%, var(--c-action-menu-surface-subtle));
  color: color-mix(in srgb, var(--color-danger) 96%, var(--c-action-menu-text));
}

.c-action-menu :deep(.c-action-menu__trigger) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 1px solid var(--c-action-menu-line);
  border-radius: 0.75rem;
  background: var(--c-action-menu-surface);
  color: var(--c-action-menu-muted);
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    color 160ms ease,
    box-shadow 160ms ease;
}

.c-action-menu :deep(.c-action-menu__trigger--icon) {
  width: 1.95rem;
  height: 1.95rem;
  padding: 0;
  flex: 0 0 auto;
}

.c-action-menu :deep(.c-action-menu__trigger:hover),
.c-action-menu :deep(.c-action-menu__trigger[aria-expanded='true']),
.c-action-menu :deep(.c-action-menu__trigger--active) {
  border-color: color-mix(in srgb, var(--c-action-menu-accent) 26%, var(--c-action-menu-line-strong));
  background: var(--c-action-menu-accent-soft);
  color: var(--c-action-menu-accent);
  box-shadow: 0 12px 26px color-mix(in srgb, var(--c-action-menu-accent) 10%, transparent);
}

.c-action-menu :deep(.c-action-menu__trigger:focus-visible),
.c-action-menu__content :deep(.c-action-menu__item:focus-visible) {
  outline: 2px solid color-mix(in srgb, var(--c-action-menu-accent) 42%, white);
  outline-offset: 2px;
}
</style>
