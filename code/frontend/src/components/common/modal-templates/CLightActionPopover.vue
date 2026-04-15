<script setup lang="ts">
import { MessageSquare, SendHorizonal } from 'lucide-vue-next'
import { onBeforeUnmount, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    description?: string
    width?: string
    closeOnOutside?: boolean
  }>(),
  {
    title: '发现题目问题？',
    description: '请简要描述您遇到的环境或描述错误。',
    width: '18rem',
    closeOnOutside: true,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  close: []
  submit: []
}>()

const rootRef = ref<HTMLDivElement | null>(null)

function closePopover(): void {
  emit('update:modelValue', false)
  emit('close')
}

function togglePopover(): void {
  emit('update:modelValue', !props.modelValue)
}

function handleWindowClick(event: MouseEvent): void {
  if (!props.modelValue || !props.closeOnOutside) return
  const target = event.target
  if (!(target instanceof Node)) return
  if (rootRef.value?.contains(target)) return
  closePopover()
}

watch(
  () => props.modelValue,
  (open) => {
    window.removeEventListener('click', handleWindowClick)
    if (open && props.closeOnOutside) {
      window.addEventListener('click', handleWindowClick)
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  window.removeEventListener('click', handleWindowClick)
})
</script>

<template>
  <div ref="rootRef" class="relative inline-flex">
    <slot name="trigger" :open="props.modelValue" :toggle="togglePopover">
      <button
        type="button"
        class="flex items-center gap-2 rounded border border-slate-200 bg-white px-5 py-2.5 text-[13px] font-medium text-slate-700 transition-colors hover:bg-slate-50"
        :class="{ 'bg-slate-200 text-slate-900': props.modelValue }"
        @click.stop="togglePopover"
      >
        <MessageSquare :size="16" />
        题目反馈
      </button>
    </slot>

    <div
      v-if="props.modelValue"
      class="absolute left-1/2 top-full z-50 mt-3 w-72 -translate-x-1/2 rounded border border-slate-200/60 bg-white shadow-[0_12px_40px_rgba(0,0,0,0.12)]"
      :style="{ width: props.width }"
    >
      <div class="p-5">
        <h3 class="mb-1 text-[14px] font-bold text-slate-900">{{ props.title }}</h3>
        <p class="mb-4 text-[12px] text-slate-500">{{ props.description }}</p>

        <slot>
          <textarea
            class="mb-3 h-20 w-full resize-none rounded border border-slate-200 bg-slate-50 px-3 py-2 text-[13px] outline-none transition-colors placeholder:text-slate-400 focus:border-[#2a7a58]"
            placeholder="例如：实例无法连接，或附件下载 404..."
          />
        </slot>

        <div class="flex justify-end gap-2">
          <slot name="actions" :close="closePopover">
            <button
              type="button"
              class="px-4 py-2 text-[12px] font-medium text-slate-500 hover:text-slate-800"
              @click="closePopover"
            >
              取消
            </button>
            <button
              type="button"
              class="flex items-center gap-1.5 rounded bg-[#2a7a58] px-4 py-2 text-[12px] font-medium text-white transition-colors hover:bg-[#206346]"
              @click="emit('submit')"
            >
              发送反馈
              <SendHorizonal :size="12" />
            </button>
          </slot>
        </div>
      </div>
    </div>
  </div>
</template>
