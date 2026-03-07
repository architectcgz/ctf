<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { downloadReport, exportPersonalReport } from '@/api/assessment'
import { getProfile } from '@/api/auth'
import type { AuthUser, ReportExportData } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'

const authStore = useAuthStore()

const loading = ref(false)
const error = ref<string | null>(null)
const profile = ref<AuthUser | null>(null)
const exportLoading = ref(false)
const exportError = ref<string | null>(null)
const reportFormat = ref<'pdf' | 'excel'>('pdf')
const latestReport = ref<ReportExportData | null>(null)

const profileFields = computed(() => {
  const current = profile.value
  if (!current) return []
  return [
    { label: '用户名', value: current.username },
    { label: '角色', value: current.role },
    { label: '班级', value: current.class_name || '未分配' },
    { label: '姓名', value: current.name || '未填写' },
  ]
})

async function loadProfile(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    profile.value = await getProfile()
  } catch (err) {
    console.error('加载个人资料失败:', err)
    profile.value = authStore.user
      ? { ...authStore.user }
      : null
    error.value = '加载个人资料失败，以下展示的是本地缓存信息'
  } finally {
    loading.value = false
  }
}

async function createReport(): Promise<void> {
  exportLoading.value = true
  exportError.value = null
  try {
    latestReport.value = await exportPersonalReport({ format: reportFormat.value })
  } catch (err) {
    console.error('导出个人报告失败:', err)
    exportError.value = '创建个人报告失败，请稍后重试'
  } finally {
    exportLoading.value = false
  }
}

async function handleDownload(): Promise<void> {
  if (!latestReport.value) return

  const file = await downloadReport(latestReport.value.report_id)
  const url = window.URL.createObjectURL(file.blob)
  const link = document.createElement('a')
  link.href = url
  link.download = file.filename
  link.click()
  window.URL.revokeObjectURL(url)
}

onMounted(() => {
  loadProfile()
})
</script>

<template>
  <div class="space-y-6">
    <section class="rounded-[28px] border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(14,116,144,0.10),rgba(59,130,246,0.10))] p-7 shadow-sm">
      <p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--color-primary)]/85">My Account</p>
      <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--color-text-primary)]">个人资料</h1>
      <p class="mt-2 max-w-3xl text-sm leading-6 text-[var(--color-text-secondary)]">
        查看当前账号信息，并生成你的个人训练报告用于复盘和归档。
      </p>
    </section>

    <div v-if="error" class="rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4 text-sm text-amber-700">
      {{ error }}
    </div>

    <div v-if="loading" class="grid gap-4 lg:grid-cols-[1fr_0.9fr]">
      <div class="h-72 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]"></div>
      <div class="h-72 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]"></div>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-[1fr_0.9fr]">
      <section class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">账号信息</h2>
            <p class="mt-1 text-sm text-[var(--color-text-secondary)]">当前展示的是后端返回的用户资料。</p>
          </div>
          <button
            type="button"
            class="rounded-xl border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)]"
            @click="loadProfile"
          >
            刷新
          </button>
        </div>

        <div v-if="!profile" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
          当前没有可展示的用户信息。
        </div>

        <div v-else class="mt-5 grid gap-4 sm:grid-cols-2">
          <div v-for="item in profileFields" :key="item.label" class="rounded-xl bg-[var(--color-bg-base)] px-4 py-4">
            <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">{{ item.label }}</p>
            <p class="mt-2 text-base font-medium text-[var(--color-text-primary)]">{{ item.value }}</p>
          </div>
        </div>

        <div class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-4 text-sm leading-6 text-[var(--color-text-secondary)]">
          当前后端尚未开放密码修改接口，本页不展示不可提交的表单，避免产生误导操作。
        </div>
      </section>

      <section class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">个人报告导出</h2>
        <p class="mt-1 text-sm text-[var(--color-text-secondary)]">生成 PDF 或 Excel 报告，便于训练复盘和归档。</p>

        <div class="mt-5 space-y-4">
          <label class="block text-sm font-medium text-[var(--color-text-primary)]">
            导出格式
            <select
              v-model="reportFormat"
              class="mt-2 w-full rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
            >
              <option value="pdf">PDF</option>
              <option value="excel">Excel</option>
            </select>
          </label>

          <button
            type="button"
            class="w-full rounded-xl bg-[var(--color-primary)] px-4 py-3 text-sm font-medium text-white transition hover:bg-[var(--color-primary-hover)] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="exportLoading"
            @click="createReport"
          >
            {{ exportLoading ? '正在创建报告...' : '生成个人报告' }}
          </button>
        </div>

        <div v-if="exportError" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-4 text-sm text-red-600">
          {{ exportError }}
        </div>

        <div v-if="latestReport" class="mt-5 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-4">
          <div class="flex items-start justify-between gap-4">
            <div>
              <p class="font-medium text-[var(--color-text-primary)]">最近一次导出任务</p>
              <p class="mt-1 text-sm text-[var(--color-text-secondary)]">报告 ID：{{ latestReport.report_id }}</p>
              <p v-if="latestReport.expires_at" class="mt-1 text-sm text-[var(--color-text-secondary)]">
                下载有效期至：{{ formatDate(latestReport.expires_at) }}
              </p>
            </div>
            <span class="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-semibold text-emerald-700">
              {{ latestReport.status === 'ready' ? '可下载' : latestReport.status }}
            </span>
          </div>

          <button
            type="button"
            class="mt-4 w-full rounded-xl border border-[var(--color-border-default)] px-4 py-3 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="latestReport.status !== 'ready'"
            @click="handleDownload"
          >
            下载最近报告
          </button>
        </div>
      </section>
    </div>
  </div>
</template>
