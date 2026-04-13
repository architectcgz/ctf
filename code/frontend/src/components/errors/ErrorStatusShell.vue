<script setup lang="ts">
import { computed, type Component } from 'vue'
import { RouterLink } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { redirectTo, reloadPage } from '@/utils/browser'
import { resolveErrorStatusRetryTarget } from '@/utils/errorStatusPage'
import { getRoleDashboardPath } from '@/utils/roleRoutes'

interface Props {
  statusCode: string
  kicker: string
  title: string
  description: string
  icon: Component
  primaryIcon: Component
  secondaryIcon: Component
  accent?: 'warning' | 'danger' | 'primary'
  primaryTo?: string
  primaryLabel?: string
  primaryAction?: 'back' | 'reload'
  secondaryTo?: string
  secondaryLabel?: string
  secondaryAction?: 'back' | 'reload'
  secondaryToHome?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  accent: 'primary',
  primaryTo: '',
  primaryLabel: '',
  primaryAction: undefined,
  secondaryTo: '',
  secondaryLabel: '',
  secondaryAction: undefined,
  secondaryToHome: false,
})

const authStore = useAuthStore()

const accentValueMap: Record<NonNullable<Props['accent']>, string> = {
  primary: 'var(--color-primary)',
  warning: 'var(--color-warning)',
  danger: 'var(--color-danger)',
}

const dynamicHomePath = computed(() => {
  if (!authStore.isLoggedIn) return '/login'
  return getRoleDashboardPath(authStore.user?.role)
})

const dynamicHomeLabel = computed(() => {
  if (!authStore.isLoggedIn) return '返回登录页'
  if (authStore.isAdmin) return '返回管理工作台'
  if (authStore.isTeacher) return '返回教师工作台'
  return '返回学习工作台'
})

const resolvedPrimaryTo = computed(() => props.primaryTo || dynamicHomePath.value)
const resolvedPrimaryLabel = computed(() => {
  if (props.primaryLabel) return props.primaryLabel
  if (props.primaryAction === 'reload') return '刷新页面'
  if (props.primaryAction === 'back') return '返回上一页'
  return dynamicHomeLabel.value
})
const resolvedSecondaryTo = computed(() => {
  if (props.secondaryToHome) return dynamicHomePath.value
  return props.secondaryTo
})
const resolvedSecondaryLabel = computed(() => {
  if (props.secondaryLabel) return props.secondaryLabel
  if (props.secondaryAction === 'reload') return '刷新页面'
  if (props.secondaryAction === 'back') return '返回上一页'
  if (props.secondaryToHome) return dynamicHomeLabel.value
  return ''
})
const showSecondaryAction = computed(
  () =>
    Boolean(props.secondaryAction) ||
    Boolean(resolvedSecondaryTo.value && resolvedSecondaryLabel.value)
)
const accentVars = computed(() => ({
  '--error-accent': accentValueMap[props.accent],
}))

function navigateBack() {
  if (window.history.length > 1) {
    window.history.back()
    return
  }
  redirectTo(resolvedPrimaryTo.value)
}

function executeAction(action: NonNullable<Props['primaryAction'] | Props['secondaryAction']>) {
  if (action === 'reload') {
    const retryTarget = resolveErrorStatusRetryTarget(window.location.search)
    if (retryTarget && retryTarget !== window.location.pathname) {
      redirectTo(retryTarget)
      return
    }
    reloadPage()
    return
  }

  navigateBack()
}
</script>

<template>
  <section class="error-status-view" :style="accentVars">
    <div class="error-status-kicker">
      <component :is="icon" class="h-4 w-4" />
      <span>{{ kicker }}</span>
    </div>

    <div class="error-status-grid">
      <div class="error-status-copy">
        <h1 class="error-status-title workspace-page-title">
          {{ title }}
        </h1>
        <p class="error-status-text workspace-page-copy">
          {{ description }}
        </p>

        <div class="error-status-actions">
          <button
            v-if="primaryAction"
            type="button"
            class="error-status-action error-status-action-primary"
            @click="executeAction(primaryAction)"
          >
            <component :is="primaryIcon" class="h-4 w-4" />
            {{ resolvedPrimaryLabel }}
          </button>
          <RouterLink
            v-else
            :to="resolvedPrimaryTo"
            class="error-status-action error-status-action-primary"
          >
            <component :is="primaryIcon" class="h-4 w-4" />
            {{ resolvedPrimaryLabel }}
          </RouterLink>
          <button
            v-if="secondaryAction"
            type="button"
            class="error-status-action error-status-action-secondary"
            @click="executeAction(secondaryAction)"
          >
            <component :is="secondaryIcon" class="h-4 w-4" />
            {{ resolvedSecondaryLabel }}
          </button>
          <RouterLink
            v-else-if="showSecondaryAction"
            :to="resolvedSecondaryTo"
            class="error-status-action error-status-action-secondary"
          >
            <component :is="secondaryIcon" class="h-4 w-4" />
            {{ resolvedSecondaryLabel }}
          </RouterLink>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.error-status-view {
  min-height: calc(100vh - 11rem);
  display: flex;
  flex-direction: column;
  justify-content: center;
  margin-inline: auto;
  width: min(64rem, 100%);
  padding: 2.75rem 1rem 3.25rem;
}

.error-status-kicker {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  align-self: flex-start;
  padding-left: 0.72rem;
  border-left: 2px solid color-mix(in srgb, var(--error-accent) 55%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--error-accent) 82%, var(--color-text-primary));
}

.error-status-grid {
  margin-top: 1rem;
  display: block;
}

.error-status-copy {
  min-width: 0;
}

.error-status-title {
  font-weight: 700;
  color: var(--color-text-primary);
}

.error-status-text {
  margin-top: 0.8rem;
  max-width: 56ch;
  color: var(--color-text-secondary);
}

.error-status-actions {
  margin-top: 1.3rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.7rem;
}

.error-status-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border-radius: 10px;
  border: 1px solid transparent;
  padding: 0.58rem 0.9rem;
  font-size: var(--font-size-0-86);
  font-weight: 600;
  transition: all 180ms ease;
}

.error-status-action-primary {
  border-color: color-mix(in srgb, var(--color-primary) 45%, transparent);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--color-primary) 94%, var(--color-bg-surface)),
      color-mix(in srgb, var(--color-primary-hover) 78%, var(--color-bg-surface))
    );
  color: var(--color-text-primary);
}

.error-status-action-primary:hover {
  transform: translateY(-1px);
  filter: brightness(1.03);
}

.error-status-action-secondary {
  border-color: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 78%, transparent);
  color: var(--color-text-primary);
}

.error-status-action-secondary:hover {
  border-color: color-mix(in srgb, var(--error-accent) 28%, var(--color-primary) 22%, transparent);
}

@media (max-width: 767px) {
  .error-status-view {
    min-height: calc(100vh - 9.5rem);
    justify-content: flex-start;
    padding-top: 1.9rem;
  }
}

:global([data-theme='light']) .error-status-action-secondary {
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-elevated));
}
</style>
