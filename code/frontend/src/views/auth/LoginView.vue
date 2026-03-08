<template>
  <div class="relative min-h-screen overflow-hidden bg-base text-text-primary">
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute left-[12%] top-[10%] h-64 w-64 rounded-full bg-primary/14 blur-3xl" />
      <div class="absolute bottom-[12%] right-[8%] h-72 w-72 rounded-full bg-sky-500/10 blur-3xl" />
      <div class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.03)_1px,transparent_1px)] bg-[size:28px_28px] opacity-30" />
    </div>
    <div class="relative mx-auto flex min-h-screen max-w-6xl items-center px-4 py-8 lg:px-8">
      <div class="grid w-full overflow-hidden rounded-[32px] border border-border bg-surface/88 shadow-[0_28px_80px_var(--color-shadow-strong)] lg:grid-cols-[1.05fr_0.95fr]">
        <section class="flex flex-col justify-between border-b border-border px-6 py-7 lg:border-b-0 lg:border-r lg:px-8 lg:py-8">
          <div class="space-y-6">
            <div class="inline-flex w-fit items-center gap-2 rounded-full border border-primary/30 bg-primary/10 px-4 py-1.5 text-[11px] font-semibold uppercase tracking-[0.26em] text-primary">
              CTF Training
            </div>
            <div class="space-y-3">
              <h1 class="max-w-lg text-4xl font-semibold tracking-tight text-text-primary">在同一套训练平台里完成靶场、竞赛与教学联动。</h1>
              <p class="max-w-xl text-sm leading-7 text-text-secondary">
                面向学员、教师与管理员的统一攻防实训工作台。实时通知、能力画像和班级数据在一个界面内完成闭环。
              </p>
            </div>
            <div class="grid gap-3 sm:grid-cols-3">
              <div
                v-for="item in highlights"
                :key="item.label"
                class="rounded-2xl border border-border bg-elevated/72 px-4 py-4"
              >
                <div class="text-xs font-semibold uppercase tracking-[0.18em] text-text-muted">{{ item.label }}</div>
                <div class="mt-2 text-sm leading-6 text-text-primary">{{ item.value }}</div>
              </div>
            </div>
          </div>
          <div class="mt-8 grid gap-3 sm:grid-cols-3">
            <div
              v-for="feature in features"
              :key="feature.title"
              class="rounded-2xl border border-border-subtle bg-base/55 px-4 py-4"
            >
              <component :is="feature.icon" class="h-5 w-5 text-primary" />
              <div class="mt-3 text-sm font-semibold text-text-primary">{{ feature.title }}</div>
              <div class="mt-1 text-xs leading-6 text-text-secondary">{{ feature.description }}</div>
            </div>
          </div>
        </section>

        <section class="px-6 py-7 lg:px-8 lg:py-8">
          <div class="mx-auto max-w-md">
            <div class="space-y-2">
              <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-text-muted">Account Access</div>
              <div class="text-2xl font-semibold text-text-primary">登录平台</div>
              <div class="text-sm text-text-secondary">使用你的训练账号进入工作台。</div>
            </div>

            <ElForm class="mt-8" :model="form" label-position="top" @submit.prevent="onSubmit">
              <ElFormItem label="用户名">
                <ElInput v-model="form.username" autocomplete="username" size="large" />
              </ElFormItem>
              <ElFormItem label="密码">
                <ElInput
                  v-model="form.password"
                  type="password"
                  autocomplete="current-password"
                  show-password
                  size="large"
                />
              </ElFormItem>

              <div class="mt-6 flex items-center justify-between gap-3">
                <RouterLink class="text-sm text-primary hover:text-primary-hover" to="/register">创建新账号</RouterLink>
                <ElButton type="primary" size="large" :loading="loading" @click="onSubmit">登录</ElButton>
              </div>
            </ElForm>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { BellRing, GraduationCap, ShieldCheck } from 'lucide-vue-next'
import { reactive, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { useAuth } from '@/composables/useAuth'

const { login } = useAuth()
const route = useRoute()

const loading = ref(false)
const form = reactive({ username: '', password: '' })
const highlights = [
  { label: '场景', value: '训练 / 竞赛 / 教学三条链路统一接入' },
  { label: '通知', value: '实时推送关键事件，减少轮询感知成本' },
  { label: '评估', value: '个人与班级报告可直接导出留档' },
]
const features = [
  {
    title: '实时感知',
    description: '通知和关键状态在统一壳层里即时可见。',
    icon: BellRing,
  },
  {
    title: '教学闭环',
    description: '教师能直接查看班级、学生与报告导出进度。',
    icon: GraduationCap,
  },
  {
    title: '风控可见',
    description: '管理员可从仪表盘进入审计和风险研判。',
    icon: ShieldCheck,
  },
]

function sanitizeRedirectPath(input: unknown): string {
  if (typeof input !== 'string') return '/dashboard'
  if (!input.startsWith('/')) return '/dashboard'
  if (input.startsWith('//')) return '/dashboard'
  return input
}

async function onSubmit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    const redirectTo = sanitizeRedirectPath(route.query.redirect)
    await login({ username: form.username, password: form.password }, redirectTo)
  } finally {
    loading.value = false
  }
}
</script>
