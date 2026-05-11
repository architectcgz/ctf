<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, type Component } from 'vue'
import { RouterLink } from 'vue-router'

import { useProbeEasterEggs } from '@/composables/useProbeEasterEggs'
import { useAuthStore } from '@/stores/auth'
import { getNavigationType, redirectTo, reloadPage } from '@/utils/browser'
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
const { track } = useProbeEasterEggs()
const probeMessage = ref('')
let probeMessageTimer: number | null = null

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

function getReloadRetryTarget(): string | null {
  const retryTarget = resolveErrorStatusRetryTarget(window.location.search)
  if (!retryTarget || retryTarget === window.location.pathname) return null
  return retryTarget
}

onMounted(() => {
  if (getNavigationType() !== 'reload') return
  if (props.primaryAction !== 'reload' && props.secondaryAction !== 'reload') return

  const retryTarget = getReloadRetryTarget()
  if (retryTarget) {
    redirectTo(retryTarget)
  }
})

onBeforeUnmount(() => {
  if (probeMessageTimer) {
    window.clearTimeout(probeMessageTimer)
  }
})

function showProbeMessage(message: string) {
  probeMessage.value = message
  if (probeMessageTimer) {
    window.clearTimeout(probeMessageTimer)
  }
  probeMessageTimer = window.setTimeout(() => {
    probeMessage.value = ''
    probeMessageTimer = null
  }, 3200)
}

function handleProbeClick() {
  const result = track('error-status', 4)
  if (!result.unlocked) {
    return
  }
  showProbeMessage('路径枚举记录已写入：热情可嘉，命中率一般。')
}

function navigateBack() {
  if (window.history.length > 1) {
    window.history.back()
    return
  }
  redirectTo(resolvedPrimaryTo.value)
}

function executeAction(action: NonNullable<Props['primaryAction'] | Props['secondaryAction']>) {
  if (action === 'reload') {
    const retryTarget = getReloadRetryTarget()
    if (retryTarget) {
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
    <div class="error-status-kicker" @click="handleProbeClick">
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
        <p v-if="probeMessage" class="error-status-probe">
          {{ probeMessage }}
        </p>

        <div class="error-status-actions">
          <button
            v-if="primaryAction"
            type="button"
            class="ui-btn ui-btn--primary error-status-action"
            @click="executeAction(primaryAction)"
          >
            <component :is="primaryIcon" class="h-4 w-4" />
            {{ resolvedPrimaryLabel }}
          </button>
          <RouterLink
            v-else
            :to="resolvedPrimaryTo"
            class="ui-btn ui-btn--primary error-status-action"
          >
            <component :is="primaryIcon" class="h-4 w-4" />
            {{ resolvedPrimaryLabel }}
          </RouterLink>
          <button
            v-if="secondaryAction"
            type="button"
            class="ui-btn ui-btn--secondary error-status-action"
            @click="executeAction(secondaryAction)"
          >
            <component :is="secondaryIcon" class="h-4 w-4" />
            {{ resolvedSecondaryLabel }}
          </button>
          <RouterLink
            v-else-if="showSecondaryAction"
            :to="resolvedSecondaryTo"
            class="ui-btn ui-btn--secondary error-status-action"
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
  --ui-btn-height: 2.35rem;
  --ui-btn-radius: 10px;
  --ui-btn-padding: 0.58rem 0.9rem;
  --ui-btn-font-size: var(--font-size-0-86);
  --ui-btn-gap: 0.45rem;
  --ui-btn-primary-border: color-mix(in srgb, var(--error-accent) 34%, var(--color-border-default));
  --ui-btn-primary-hover-border: color-mix(
    in srgb,
    var(--error-accent) 28%,
    var(--color-border-default)
  );
  --ui-btn-primary-hover-shadow: none;
  --ui-btn-secondary-hover-border: color-mix(
    in srgb,
    var(--error-accent) 24%,
    var(--color-border-default)
  );
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

.error-status-probe {
  margin-top: 0.9rem;
  max-width: 56ch;
  font-size: var(--font-size-0-82);
  line-height: 1.7;
  color: color-mix(in srgb, var(--error-accent) 84%, var(--color-text-secondary));
}

.error-status-actions {
  margin-top: 1.3rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.7rem;
}

.error-status-action {
  text-decoration: none;
}

@media (max-width: 767px) {
  .error-status-view {
    min-height: calc(100vh - 9.5rem);
    justify-content: flex-start;
    padding-top: 1.9rem;
  }
}

:global([data-theme='dark']) .error-status-view {
  --ui-btn-primary-hover-border: color-mix(
    in srgb,
    var(--error-accent) 20%,
    var(--color-border-default)
  );
  --ui-btn-secondary-hover-border: color-mix(
    in srgb,
    var(--error-accent) 18%,
    var(--color-border-default)
  );
}
</style>
