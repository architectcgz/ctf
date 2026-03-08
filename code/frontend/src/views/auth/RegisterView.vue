<template>
  <div class="relative min-h-screen overflow-hidden bg-base text-text-primary">
    <div class="pointer-events-none absolute inset-0">
      <div class="absolute right-[12%] top-[8%] h-72 w-72 rounded-full bg-primary/12 blur-3xl" />
      <div class="absolute bottom-[10%] left-[8%] h-72 w-72 rounded-full bg-sky-400/8 blur-3xl" />
      <div class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.025)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.025)_1px,transparent_1px)] bg-[size:28px_28px] opacity-25" />
    </div>
    <div class="relative mx-auto flex min-h-screen max-w-6xl items-center px-4 py-8 lg:px-8">
      <div class="grid w-full overflow-hidden rounded-[32px] border border-border bg-surface/88 shadow-[0_28px_80px_var(--color-shadow-strong)] lg:grid-cols-[0.92fr_1.08fr]">
        <section class="border-b border-border px-6 py-7 lg:border-b-0 lg:border-r lg:px-8 lg:py-8">
          <div class="mx-auto max-w-md">
            <div class="space-y-2">
              <div class="text-[11px] font-semibold uppercase tracking-[0.24em] text-text-muted">New Account</div>
              <div class="text-2xl font-semibold text-text-primary">注册账号</div>
              <div class="text-sm text-text-secondary">创建你的平台账号，进入靶场训练或竞赛协作流程。</div>
            </div>

            <ElForm class="mt-8" :model="form" label-position="top" @submit.prevent="onSubmit">
              <ElFormItem label="用户名">
                <ElInput v-model="form.username" autocomplete="username" size="large" />
              </ElFormItem>
              <ElFormItem label="密码">
                <ElInput
                  v-model="form.password"
                  type="password"
                  autocomplete="new-password"
                  show-password
                  size="large"
                />
              </ElFormItem>
              <ElFormItem label="班级（可选）">
                <ElInput v-model="form.class_name" size="large" />
              </ElFormItem>

              <div class="mt-6 flex items-center justify-between gap-3">
                <RouterLink class="text-sm text-primary hover:text-primary-hover" to="/login">已有账号，去登录</RouterLink>
                <ElButton type="primary" size="large" :loading="loading" @click="onSubmit">注册</ElButton>
              </div>
            </ElForm>
          </div>
        </section>

        <section class="flex flex-col justify-between px-6 py-7 lg:px-8 lg:py-8">
          <div class="space-y-6">
            <div class="inline-flex w-fit items-center gap-2 rounded-full border border-primary/30 bg-primary/10 px-4 py-1.5 text-[11px] font-semibold uppercase tracking-[0.26em] text-primary">
              Guided Onboarding
            </div>
            <div class="space-y-3">
              <h1 class="max-w-lg text-4xl font-semibold tracking-tight text-text-primary">从训练到班级协作，都在同一套界面完成接入。</h1>
              <p class="max-w-xl text-sm leading-7 text-text-secondary">
                账号创建后可以直接进入靶场、竞赛和能力评估链路。教师与管理员角色也共享同一套壳层与通知体系。
              </p>
            </div>
          </div>

          <div class="mt-8 space-y-3">
            <div
              v-for="step in onboardingSteps"
              :key="step.title"
              class="flex items-start gap-4 rounded-2xl border border-border bg-elevated/72 px-4 py-4"
            >
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-primary/12 text-primary">
                <component :is="step.icon" class="h-5 w-5" />
              </div>
              <div>
                <div class="text-sm font-semibold text-text-primary">{{ step.title }}</div>
                <div class="mt-1 text-xs leading-6 text-text-secondary">{{ step.description }}</div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { BarChart3, ShieldCheck, Swords } from 'lucide-vue-next'
import { reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { useAuth } from '@/composables/useAuth'

const { register } = useAuth()

const loading = ref(false)
const form = reactive({ username: '', password: '', class_name: '' })
const onboardingSteps = [
  {
    title: '进入训练环境',
    description: '创建账号后可立即访问靶场、实例和个人进度。',
    icon: Swords,
  },
  {
    title: '查看能力画像',
    description: '系统会基于训练结果形成可导出的个人与班级报告。',
    icon: BarChart3,
  },
  {
    title: '统一权限体系',
    description: '管理员、教师和学员共享一致的认证与风控链路。',
    icon: ShieldCheck,
  },
]

async function onSubmit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    await register({ username: form.username, password: form.password, class_name: form.class_name || undefined })
  } finally {
    loading.value = false
  }
}
</script>
