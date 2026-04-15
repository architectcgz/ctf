<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(
  defineProps<{
    title?: string
    description?: string
    width?: string
  }>(),
  {
    title: 'TLS 握手 (TLS Handshake)',
    description:
      '一种安全协议过程，用于在客户端和服务器之间建立加密通信连接。通常包含 ClientHello 和 ServerHello。',
    width: '16rem',
  }
)

const open = ref(false)
</script>

<template>
  <span
    class="c-context-tooltip relative inline-block border-b border-dashed border-slate-400 cursor-help text-slate-900 font-bold"
    @mouseenter="open = true"
    @mouseleave="open = false"
    @focusin="open = true"
    @focusout="open = false"
  >
    <slot>TLS 握手</slot>

    <div
      v-if="open"
      class="absolute bottom-full left-1/2 z-50 mb-3 -translate-x-1/2 rounded-md bg-slate-900 p-4 text-left text-[13px] font-normal leading-relaxed text-white shadow-xl"
      :style="{ width: props.width }"
    >
      <div class="mb-1 font-bold text-white">{{ props.title }}</div>
      <div class="text-slate-300">{{ props.description }}</div>
      <div
        class="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-slate-900"
      />
    </div>
  </span>
</template>
