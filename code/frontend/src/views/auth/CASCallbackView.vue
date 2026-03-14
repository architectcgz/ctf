<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { sanitizeRedirectPath } from '@/router/guards'
import { useCASAuth } from '@/composables/useCASAuth'

const route = useRoute()
const router = useRouter()
const { finishCASLogin } = useCASAuth()

const loading = ref(true)
const errorMessage = ref('')
const ticket = computed(() =>
  typeof route.query.ticket === 'string' ? route.query.ticket.trim() : ''
)
const redirectTo = computed(() => {
  const value = sanitizeRedirectPath(route.query.redirect)
  return value === '/' ? undefined : value
})

async function handleCallback() {
  if (!ticket.value) {
    errorMessage.value = '未收到 CAS ticket，无法完成统一认证登录。'
    loading.value = false
    return
  }

  try {
    await finishCASLogin(ticket.value, redirectTo.value)
  } catch (error) {
    errorMessage.value =
      error instanceof Error ? error.message : 'CAS 登录失败，请返回登录页后重试。'
    loading.value = false
  }
}

onMounted(() => {
  void handleCallback()
})
</script>

<template>
  <div class="relative min-h-screen overflow-hidden bg-base text-text-primary">
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute left-[10%] top-[8%] h-64 w-64 rounded-full bg-primary/14 blur-3xl" />
      <div class="absolute bottom-[10%] right-[8%] h-72 w-72 rounded-full bg-sky-500/10 blur-3xl" />
      <div class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.03)_1px,transparent_1px)] bg-[size:28px_28px] opacity-30" />
    </div>

    <div class="relative mx-auto flex min-h-screen max-w-3xl items-center px-4 py-8">
      <section class="w-full rounded-[32px] border border-border bg-surface/88 p-8 shadow-[0_28px_80px_var(--color-shadow-strong)]">
        <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-text-muted">
          Campus SSO
        </div>
        <h1 class="mt-3 text-3xl font-semibold tracking-tight text-text-primary">
          正在完成 CAS 登录
        </h1>
        <p class="mt-3 text-sm leading-7 text-text-secondary">
          平台正在处理学校统一认证返回的票据，并尝试恢复你之前要访问的页面。
        </p>

        <div class="mt-8">
          <div
            v-if="loading"
            class="rounded-2xl border border-border bg-elevated/72 px-6 py-6 text-sm text-text-secondary"
          >
            <AppLoading>正在校验 CAS ticket 并恢复登录状态...</AppLoading>
          </div>

          <AppEmpty
            v-else
            title="CAS 登录未完成"
            :description="errorMessage"
            icon="AlertTriangle"
          >
            <template #action>
              <div class="flex flex-wrap justify-center gap-3">
                <button
                  type="button"
                  class="rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                  @click="router.replace('/login')"
                >
                  返回登录页
                </button>
                <button
                  type="button"
                  class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
                  @click="void handleCallback()"
                >
                  重试回调
                </button>
              </div>
            </template>
          </AppEmpty>
        </div>
      </section>
    </div>
  </div>
</template>
