<script setup lang="ts">
import { RouterLink } from 'vue-router'

import type { AwdPageDefinition } from '@/modules/awd/types'

const props = defineProps<{
  items: Array<AwdPageDefinition<string>>
  currentPage: string
  resolvePath?: (pageKey: string) => string
}>()

function itemClasses(pageKey: string): string[] {
  return [
    'awd-page-nav__item',
    pageKey === props.currentPage ? 'awd-page-nav__item--active' : '',
  ]
}
</script>

<template>
  <nav
    class="awd-page-nav"
    aria-label="AWD 页面导航"
  >
    <div class="awd-page-nav__label">
      页面导航
    </div>
    <div class="awd-page-nav__list">
      <component
        :is="resolvePath ? RouterLink : 'button'"
        v-for="item in items"
        :key="item.key"
        :to="resolvePath ? resolvePath(item.key) : undefined"
        :type="resolvePath ? undefined : 'button'"
        class="awd-page-nav__link"
        :class="itemClasses(item.key)"
        data-testid="awd-page-nav-item"
      >
        <strong>{{ item.label }}</strong>
        <span>{{ item.description }}</span>
      </component>
    </div>
  </nav>
</template>

<style scoped>
.awd-page-nav {
  display: grid;
  gap: 0.9rem;
}

.awd-page-nav__label {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--color-text-secondary);
}

.awd-page-nav__list {
  display: grid;
  gap: 0.75rem;
}

.awd-page-nav__link {
  display: grid;
  gap: 0.28rem;
  padding: 0.95rem 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  text-align: left;
  color: inherit;
  text-decoration: none;
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.awd-page-nav__link strong {
  font-size: var(--font-size-14);
}

.awd-page-nav__link span {
  font-size: var(--font-size-12);
  color: var(--color-text-secondary);
  line-height: 1.45;
}

.awd-page-nav__item--active {
  border-color: color-mix(in srgb, var(--color-primary) 36%, transparent);
  background: color-mix(in srgb, var(--color-primary) 9%, var(--color-bg-surface));
  transform: translateX(2px);
}
</style>
