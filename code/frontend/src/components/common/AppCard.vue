<script setup lang="ts">
import { computed } from 'vue'

export type AppCardVariant = 'panel' | 'hero' | 'metric' | 'action'
export type AppCardAccent = 'primary' | 'success' | 'warning' | 'danger' | 'violet' | 'neutral'
export type AppCardTag = 'section' | 'button' | 'article'

interface AccentMeta {
  color: string
  softColor: string
  borderColor: string
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
  }
)

const accentMap: Record<AppCardAccent, AccentMeta> = {
  primary: {
    color: 'var(--color-primary)',
    softColor: 'var(--color-primary-soft)',
    borderColor: 'color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default))',
  },
  success: {
    color: 'var(--color-success)',
    softColor: 'color-mix(in srgb, var(--color-success) 12%, transparent)',
    borderColor: 'color-mix(in srgb, var(--color-success) 22%, transparent)',
  },
  warning: {
    color: 'var(--color-warning)',
    softColor: 'color-mix(in srgb, var(--color-warning) 12%, transparent)',
    borderColor: 'color-mix(in srgb, var(--color-warning) 22%, transparent)',
  },
  danger: {
    color: 'var(--color-danger)',
    softColor: 'color-mix(in srgb, var(--color-danger) 12%, transparent)',
    borderColor: 'color-mix(in srgb, var(--color-danger) 22%, transparent)',
  },
  violet: {
    color: 'var(--color-cat-reverse)',
    softColor: 'color-mix(in srgb, var(--color-cat-reverse) 12%, transparent)',
    borderColor: 'color-mix(in srgb, var(--color-cat-reverse) 22%, transparent)',
  },
  neutral: {
    color: 'var(--color-text-secondary)',
    softColor: 'color-mix(in srgb, var(--color-text-secondary) 8%, transparent)',
    borderColor: 'var(--color-border-default)',
  },
}

const accentMeta = computed(() => accentMap[props.accent])
const rootTag = computed(() => props.as)

const hasHeader = computed(() => Boolean(props.eyebrow || props.title || props.subtitle))

const shellClass = computed(() => {
  const base = 'relative w-full border-b border-border-subtle text-left pl-3 pr-1'
  const hover = props.interactive
    ? 'transition-colors duration-150 hover:bg-[var(--color-primary-soft)]'
    : ''

  if (props.variant === 'hero') {
    return `${base} ${hover} py-4`
  }

  if (props.variant === 'metric') {
    return `${base} ${hover} py-3`
  }

  if (props.variant === 'action') {
    return `${base} ${hover} py-3`
  }

  return `${base} ${hover} py-3`
})

const shellStyle = computed<Record<string, string>>(() => {
  const accent = accentMeta.value
  const baseStyle: Record<string, string> = {
    borderBottomColor: `color-mix(in srgb, var(--color-border-default) 84%, ${accent.softColor})`,
    borderLeftColor: accent.color,
    borderLeftWidth: '2px',
    borderLeftStyle: 'solid',
    background: 'transparent',
  }

  if (props.variant === 'hero') {
    return {
      ...baseStyle,
      background: `linear-gradient(90deg, color-mix(in srgb, ${accent.softColor} 48%, transparent), transparent 68%)`,
    }
  }

  if (props.variant === 'metric') {
    return {
      ...baseStyle,
      background: `linear-gradient(90deg, color-mix(in srgb, ${accent.softColor} 36%, transparent), transparent 64%)`,
    }
  }

  return baseStyle
})

const eyebrowStyle = computed<Record<string, string>>(() => ({
  color: accentMeta.value.color,
}))

const headerClass = computed(() => {
  if (props.variant === 'metric') return 'mb-1 flex items-start justify-between gap-3'
  if (props.variant === 'hero') return 'mb-3 flex items-start justify-between gap-3'
  return 'mb-2 flex items-start justify-between gap-3'
})

const titleClass = computed(() => {
  if (props.variant === 'hero') return 'app-card__title app-card__title--hero'
  if (props.variant === 'metric') return 'app-card__title app-card__title--metric'
  if (props.variant === 'action') return 'app-card__title app-card__title--action'
  return 'app-card__title app-card__title--panel'
})

const subtitleClass = computed(() => 'app-card__subtitle mt-1')

const bodyClass = computed(() => (props.variant === 'metric' ? 'space-y-1.5' : 'space-y-2'))

const accentChipStyle = computed<Record<string, string>>(() => ({
  backgroundColor: accentMeta.value.color,
}))
</script>

<template>
  <component
    :is="rootTag"
    :class="shellClass"
    :style="shellStyle"
    :type="props.as === 'button' ? 'button' : undefined"
  >
    <div>
      <header
        v-if="hasHeader || $slots.header"
        :class="headerClass"
      >
        <div class="min-w-0">
          <div
            v-if="eyebrow"
            class="app-card__eyebrow"
            :style="eyebrowStyle"
          >
            {{ eyebrow }}
          </div>
          <h2
            v-if="title"
            :class="[titleClass, eyebrow ? 'mt-3' : '']"
          >
            {{ title }}
          </h2>
          <p
            v-if="subtitle"
            :class="subtitleClass"
          >
            {{ subtitle }}
          </p>
        </div>

        <div class="shrink-0">
          <slot name="header">
            <div
              v-if="props.variant === 'metric'"
              class="app-card__metric-accent"
              :style="accentChipStyle"
            />
          </slot>
        </div>
      </header>

      <div :class="bodyClass">
        <slot />
      </div>

      <footer
        v-if="$slots.footer"
        class="mt-3 border-t border-border-subtle pt-3"
      >
        <slot name="footer" />
      </footer>
    </div>
  </component>
</template>

<style scoped>
.app-card__eyebrow {
  font-size: var(--font-size-12);
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
}

.app-card__title {
  color: var(--color-text-primary);
  font-weight: 600;
}

.app-card__title--hero,
.app-card__title--metric {
  font-size: 1.5rem;
  letter-spacing: -0.025em;
  line-height: 1.2;
}

.app-card__title--action {
  font-size: 0.9375rem;
}

.app-card__title--panel {
  font-size: 1rem;
}

.app-card__subtitle {
  font-size: 0.8125rem;
  line-height: 1.5rem;
  color: var(--color-text-secondary);
}

.app-card__metric-accent {
  height: 2rem;
  width: 3px;
  border-radius: 999px;
}
</style>
