<script setup lang="ts">
import { computed } from 'vue'

export type AppCardVariant = 'panel' | 'hero' | 'metric' | 'action'
export type AppCardAccent = 'primary' | 'success' | 'warning' | 'danger' | 'violet' | 'neutral'
export type AppCardTag = 'section' | 'button' | 'article'

interface AccentMeta {
  color: string
  softColor: string
  borderColor: string
  tintColor: string
}

const props = withDefaults(
  defineProps<{
    variant?: AppCardVariant
    accent?: AppCardAccent
    as?: AppCardTag
    eyebrow?: string
    title?: string
    subtitle?: string
    interactive?: boolean
  }>(),
  {
    variant: 'panel',
    accent: 'primary',
    as: 'section',
    eyebrow: '',
    title: '',
    subtitle: '',
    interactive: false,
  },
)

const accentMap: Record<AppCardAccent, AccentMeta> = {
  primary: {
    color: 'var(--color-primary)',
    softColor: 'var(--color-primary-soft)',
    borderColor: 'color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default))',
    tintColor: 'rgba(8,145,178,0.16)',
  },
  success: {
    color: 'var(--color-success)',
    softColor: 'rgba(63,185,80,0.12)',
    borderColor: 'rgba(63,185,80,0.22)',
    tintColor: 'rgba(63,185,80,0.15)',
  },
  warning: {
    color: 'var(--color-warning)',
    softColor: 'rgba(210,153,34,0.12)',
    borderColor: 'rgba(210,153,34,0.22)',
    tintColor: 'rgba(210,153,34,0.15)',
  },
  danger: {
    color: 'var(--color-danger)',
    softColor: 'rgba(248,81,73,0.12)',
    borderColor: 'rgba(248,81,73,0.22)',
    tintColor: 'rgba(248,81,73,0.15)',
  },
  violet: {
    color: '#8b5cf6',
    softColor: 'rgba(139,92,246,0.12)',
    borderColor: 'rgba(139,92,246,0.22)',
    tintColor: 'rgba(139,92,246,0.15)',
  },
  neutral: {
    color: 'var(--color-text-secondary)',
    softColor: 'rgba(139,148,158,0.08)',
    borderColor: 'var(--color-border-default)',
    tintColor: 'rgba(139,148,158,0.08)',
  },
}

const accentMeta = computed(() => accentMap[props.accent])
const rootTag = computed(() => props.as)

const hasHeader = computed(() => Boolean(props.eyebrow || props.title || props.subtitle))

const shellClass = computed(() => {
  const base =
    'relative overflow-hidden border shadow-[0_18px_40px_var(--color-shadow-soft)] backdrop-blur-sm'
  const hover = props.interactive ? 'transition duration-200 hover:-translate-y-0.5' : ''

  if (props.variant === 'hero') {
    return `${base} ${hover} rounded-[30px] p-6 shadow-[0_24px_64px_var(--color-shadow-soft)]`
  }

  if (props.variant === 'metric') {
    return `${base} ${hover} rounded-[24px] p-5`
  }

  if (props.variant === 'action') {
    return `${base} ${hover} rounded-[24px] p-4`
  }

  return `${base} ${hover} rounded-[26px] p-5`
})

const shellStyle = computed<Record<string, string>>(() => {
  const accent = accentMeta.value
  const baseStyle: Record<string, string> = {
    borderColor: 'var(--color-border-default)',
    background: 'transparent',
    backgroundColor: 'transparent',
  }

  if (props.variant === 'hero') {
    return {
      ...baseStyle,
      borderColor: accent.borderColor,
      background: `radial-gradient(circle at top left, ${accent.tintColor}, transparent 44%), linear-gradient(145deg, color-mix(in srgb, var(--color-bg-surface) 78%, black), color-mix(in srgb, var(--color-bg-base) 88%, black))`,
    }
  }

  if (props.variant === 'metric') {
    return {
      ...baseStyle,
      borderColor: accent.borderColor,
      background: `linear-gradient(180deg, color-mix(in srgb, ${accent.tintColor} 42%, transparent), rgba(15,23,42,0.08)) , color-mix(in srgb, var(--color-bg-surface) 88%, transparent)`,
    }
  }

  if (props.variant === 'action') {
    return {
      ...baseStyle,
      borderColor: accent.borderColor,
      backgroundColor: 'color-mix(in srgb, var(--color-bg-base) 62%, transparent)',
    }
  }

  return {
    ...baseStyle,
    backgroundColor: 'color-mix(in srgb, var(--color-bg-surface) 88%, transparent)',
  }
})

const topLineStyle = computed<Record<string, string>>(() => ({
  background: `linear-gradient(90deg, transparent, ${accentMeta.value.color}, transparent)`,
  opacity: props.variant === 'hero' ? '0.65' : '0.42',
}))

const glowStyle = computed<Record<string, string>>(() => ({
  background: `radial-gradient(circle, ${accentMeta.value.tintColor}, transparent 72%)`,
}))

const eyebrowStyle = computed<Record<string, string>>(() => ({
  color: props.variant === 'hero' ? 'color-mix(in srgb, white 76%, var(--color-text-muted))' : accentMeta.value.color,
}))

const headerClass = computed(() => {
  if (props.variant === 'metric') return 'mb-4 flex items-start justify-between gap-4'
  if (props.variant === 'hero') return 'mb-5 flex items-start justify-between gap-4'
  return 'mb-4 flex items-start justify-between gap-4 border-b border-border-subtle pb-4'
})

const titleClass = computed(() => {
  if (props.variant === 'hero') return 'text-[30px] font-semibold tracking-tight text-text-primary'
  if (props.variant === 'metric') return 'text-3xl font-semibold tracking-tight text-text-primary'
  if (props.variant === 'action') return 'text-base font-semibold text-text-primary'
  return 'text-lg font-semibold text-text-primary'
})

const subtitleClass = computed(() => {
  if (props.variant === 'hero') return 'mt-3 text-sm leading-7 text-text-secondary'
  return 'mt-1 text-sm leading-6 text-text-secondary'
})

const bodyClass = computed(() => (props.variant === 'metric' ? 'space-y-3' : 'space-y-4'))

const accentChipStyle = computed<Record<string, string>>(() => ({
  backgroundColor: accentMeta.value.softColor,
  borderColor: accentMeta.value.borderColor,
  color: accentMeta.value.color,
}))
</script>

<template>
  <component :is="rootTag" :class="shellClass" :style="shellStyle" :type="props.as === 'button' ? 'button' : undefined">
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute -right-10 top-0 h-28 w-28 blur-3xl" :style="glowStyle" />
      <div class="absolute inset-x-0 top-0 h-px" :style="topLineStyle" />
    </div>

    <div class="relative">
      <header v-if="hasHeader || $slots.header" :class="headerClass">
        <div class="min-w-0">
          <div
            v-if="eyebrow"
            class="text-[11px] font-semibold uppercase tracking-[0.22em]"
            :style="eyebrowStyle"
          >
            {{ eyebrow }}
          </div>
          <h2 v-if="title" :class="[titleClass, eyebrow ? 'mt-3' : '']">
            {{ title }}
          </h2>
          <p v-if="subtitle" :class="subtitleClass">
            {{ subtitle }}
          </p>
        </div>

        <div class="shrink-0">
          <slot name="header">
            <div v-if="props.variant === 'metric'" class="h-12 w-1.5 rounded-full border" :style="accentChipStyle" />
          </slot>
        </div>
      </header>

      <div :class="bodyClass">
        <slot />
      </div>

      <footer v-if="$slots.footer" class="relative mt-4 border-t border-border-subtle pt-4">
        <slot name="footer" />
      </footer>
    </div>
  </component>
</template>
