<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, ShieldAlert } from 'lucide-vue-next'

import ModalTemplateShell from './ModalTemplateShell.vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    description?: string
    note?: string
    width?: string
    cancelLabel?: string
    confirmLabel?: string
  }>(),
  {
    title: '确认重建靶机环境？',
    description:
      '此操作将永久销毁您当前的实例并重新分配新的服务器资源。您在靶机内产生的所有交互数据都将丢失。',
    note: '当前剩余可用延时/重启次数：2 次。重建后将消耗 1 次额度。',
    width: '32rem',
    cancelLabel: '保留当前环境',
    confirmLabel: '确认销毁重建',
  }
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  close: []
  cancel: []
  confirm: []
}>()

const panelStyle = computed<Record<string, string>>(() => ({
  '--c-immersive-confirm-width': props.width,
}))

function forwardOpen(value: boolean): void {
  emit('update:open', value)
}

function handleClose(): void {
  emit('close')
}

function handleCancel(): void {
  emit('cancel')
  emit('update:open', false)
}

function handleConfirm(): void {
  emit('confirm')
}
</script>

<template>
  <ModalTemplateShell
    :open="open"
    :panel-style="panelStyle"
    panel-class="c-immersive-confirm-panel"
    overlay-class="c-immersive-confirm-shell backdrop-blur-sm"
    :aria-label="props.title"
    @update:open="forwardOpen"
    @close="handleClose"
  >
    <div class="c-immersive-confirm">
      <div class="c-immersive-confirm__icon">
        <slot name="icon">
          <AlertTriangle :size="32" :stroke-width="2" />
        </slot>
      </div>

      <h2 class="c-immersive-confirm__title">{{ props.title }}</h2>
      <p class="c-immersive-confirm__description">{{ props.description }}</p>

      <div class="c-immersive-confirm__note">
        <slot name="note-icon">
          <ShieldAlert :size="16" class="shrink-0 text-slate-500" />
        </slot>
        <span>{{ props.note }}</span>
      </div>

      <div class="c-immersive-confirm__actions">
        <slot name="actions" :cancel="handleCancel" :confirm="handleConfirm">
          <button type="button" class="c-immersive-confirm__button c-immersive-confirm__button--ghost" @click="handleCancel">
            {{ props.cancelLabel }}
          </button>
          <button type="button" class="c-immersive-confirm__button c-immersive-confirm__button--danger" @click="handleConfirm">
            {{ props.confirmLabel }}
          </button>
        </slot>
      </div>
    </div>
  </ModalTemplateShell>
</template>

<style scoped>
.c-immersive-confirm-shell {
  background: rgba(15, 23, 42, 0.3);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.c-immersive-confirm-panel {
  width: min(var(--c-immersive-confirm-width, 32rem), 100%);
  overflow: hidden;
  border-radius: 0.25rem;
  background: #ffffff;
  box-shadow: 0 25px 50px rgba(15, 23, 42, 0.28);
}

.c-immersive-confirm {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem;
  text-align: center;
}

.c-immersive-confirm__icon {
  margin-bottom: 1.5rem;
  display: inline-flex;
  width: 4rem;
  height: 4rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #fef2f2;
  color: #ef4444;
}

.c-immersive-confirm__title {
  margin: 0 0 1rem;
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
}

.c-immersive-confirm__description {
  margin: 0 0 2rem;
  font-size: 15px;
  line-height: 1.75;
  color: #475569;
}

.c-immersive-confirm__note {
  display: flex;
  width: 100%;
  align-items: flex-start;
  gap: 0.75rem;
  margin-bottom: 2rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  background: #f8fafc;
  padding: 1rem;
  text-align: left;
  font-size: 13px;
  line-height: 1.65;
  color: #475569;
}

.c-immersive-confirm__actions {
  display: flex;
  width: 100%;
  gap: 1rem;
}

.c-immersive-confirm__button {
  flex: 1 1 0;
  border-radius: 0.25rem;
  padding: 0.875rem 1rem;
  font-size: 14px;
  font-weight: 500;
  transition: background-color 0.18s ease, color 0.18s ease;
}

.c-immersive-confirm__button--ghost {
  background: #f1f5f9;
  color: #475569;
}

.c-immersive-confirm__button--ghost:hover {
  background: #e2e8f0;
}

.c-immersive-confirm__button--danger {
  background: #dc2626;
  color: #ffffff;
}

.c-immersive-confirm__button--danger:hover {
  background: #b91c1c;
}
</style>
