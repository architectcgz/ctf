<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { Aperture, BookOpenText, ScanLine } from 'lucide-vue-next'

const props = defineProps<{
  activeVariant: '1' | '2' | '3'
}>()

type VariantKey = '1' | '2' | '3'

const variants: Array<{
  key: VariantKey
  title: string
  summary: string
  icon: typeof Aperture
}> = [
  {
    key: '1',
    title: '战情指挥台',
    summary: '锐利网格、态势面板、偏作战室气质，适合 CTF 平台的攻防氛围。',
    icon: Aperture,
  },
  {
    key: '2',
    title: 'Clean SaaS',
    summary: '极简专业、白底 Slate 配色、高可读与柔和阴影，接近现代 SaaS 控制台。',
    icon: BookOpenText,
  },
  {
    key: '3',
    title: '终端信号台',
    summary: '终端日志、控制台栅格、技术感更强，偏 SOC / 运维控制面板。',
    icon: ScanLine,
  },
]

const activeVariantSummary = computed(() => variants.find((item) => item.key === props.activeVariant))

function dashboardPath(key: VariantKey): string {
  return `/dashboard/${key}`
}
</script>

<template>
  <section class="style-bar rounded-[28px] border border-white/10 px-5 py-5 md:px-6">
    <div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
      <div class="max-w-2xl">
        <div class="style-eyebrow">Preview Styles</div>
        <h2 class="mt-2 text-xl font-semibold text-white">同一份学生面板，提供 3 套可直接挑选的 UI 风格</h2>
        <p class="mt-2 text-sm leading-7 text-[var(--color-text-secondary)]/82">
          当前预览：
          <span class="font-medium text-white">{{ activeVariantSummary?.title }}</span>
          。模块、数据、交互一致，只调整视觉语言与信息编排密度。
        </p>
      </div>
      <div class="text-xs uppercase tracking-[0.28em] text-[var(--color-text-muted)]">URL: `dashboard/1..3`</div>
    </div>

    <div class="mt-5 grid gap-3 lg:grid-cols-3">
      <RouterLink
        v-for="item in variants"
        :key="item.key"
        :to="dashboardPath(item.key)"
        class="style-option group relative overflow-hidden rounded-[24px] border px-4 py-4 transition duration-200"
        :class="item.key === activeVariant ? 'style-option-active' : 'style-option-idle'"
      >
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="text-[11px] font-semibold uppercase tracking-[0.28em] text-[var(--color-text-muted)]">方案 {{ item.key }}</div>
            <div class="mt-2 text-lg font-semibold text-white">{{ item.title }}</div>
          </div>
          <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-white/12 bg-white/6 text-[var(--color-text-primary)]">
            <component :is="item.icon" class="h-5 w-5" />
          </div>
        </div>
        <p class="mt-3 text-sm leading-6 text-[var(--color-text-secondary)]/78">{{ item.summary }}</p>
      </RouterLink>
    </div>
  </section>
</template>

<style scoped>
.style-bar {
  background:
    radial-gradient(circle at top left, rgba(56, 189, 248, 0.16), transparent 18rem),
    linear-gradient(135deg, rgba(15, 23, 42, 0.88), rgba(8, 47, 73, 0.72));
}

.style-eyebrow {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 0.88);
}

.style-option {
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.55), rgba(2, 6, 23, 0.65));
}

.style-option-idle {
  border-color: rgba(255, 255, 255, 0.08);
}

.style-option-idle:hover {
  border-color: rgba(56, 189, 248, 0.32);
  transform: translateY(-1px);
}

.style-option-active {
  border-color: rgba(34, 211, 238, 0.42);
  box-shadow: inset 0 0 0 1px rgba(34, 211, 238, 0.18);
}
</style>
