<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { downloadReport, exportPersonalReport } from '@/api/assessment'
import { getProfile } from '@/api/auth'
import type { AuthUser, ReportExportData } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
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
const { polling, start: startPolling, stop: stopPolling } = useReportStatusPolling()

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
    profile.value = authStore.user ? { ...authStore.user } : null
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
    if (latestReport.value.status === 'processing') {
      startPolling(String(latestReport.value.report_id), (next) => {
        latestReport.value = next
      })
    } else {
      stopPolling()
      if (latestReport.value.status === 'failed') {
        exportError.value = latestReport.value.error_message || '个人报告生成失败，请稍后重试'
      }
    }
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
    <PageHeader
      eyebrow="My Account"
      title="个人资料"
      description="查看当前账号信息，并生成你的个人训练报告用于复盘和归档。"
    />

    <div
      v-if="error"
      class="rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4 text-sm text-amber-700"
    >
      {{ error }}
    </div>

    <div v-if="loading" class="grid gap-4 lg:grid-cols-[1fr_0.9fr]">
      <AppCard variant="panel" accent="neutral">
        <div class="h-72 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]"></div>
      </AppCard>
      <AppCard variant="panel" accent="neutral">
        <div class="h-72 animate-pulse rounded-2xl bg-[var(--color-bg-surface)]"></div>
      </AppCard>
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-[1fr_0.9fr]">
      <SectionCard title="账号信息" subtitle="当前展示的是后端返回的用户资料。">
        <template #header>
          <button
            type="button"
            class="rounded-xl border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)]"
            @click="loadProfile"
          >
            刷新
          </button>
        </template>

        <AppCard
          v-if="profile"
          variant="hero"
          accent="primary"
          eyebrow="Profile Snapshot"
          :title="profile.name || profile.username"
          subtitle="账号基础信息、角色和班级都收进同一块个人视图，避免再拆成松散的信息盒。"
        >
          <template #header>
            <span
              class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
              style="border-color: color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default)); background-color: var(--color-primary-soft); color: var(--color-primary);"
            >
              {{ profile.role }}
            </span>
          </template>

          <div class="grid gap-4 sm:grid-cols-2">
            <AppCard
              v-for="item in profileFields"
              :key="item.label"
              variant="metric"
              accent="primary"
              :eyebrow="item.label"
              :title="item.value"
            />
          </div>
        </AppCard>

        <AppEmpty
          v-else
          title="暂无用户信息"
          description="当前没有可展示的用户信息。"
          icon="UsersRound"
        />

        <AppCard variant="action" accent="neutral">
          当前后端尚未开放密码修改接口，本页不展示不可提交的表单，避免产生误导操作。
        </AppCard>
      </SectionCard>

      <SectionCard title="个人报告导出" subtitle="生成 PDF 或 Excel 报告，便于训练复盘和归档。">
        <AppCard
          variant="hero"
          accent="primary"
          eyebrow="Report Export"
          title="最近一次导出任务"
          :subtitle="latestReport ? '导出状态会在这里汇总，准备就绪后可直接下载。' : '先选择导出格式，再创建本次个人报告。'"
        >
          <template #header>
            <span
              class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
              :style="latestReport?.status === 'ready'
                ? 'border-color: rgba(63,185,80,0.24); background-color: rgba(63,185,80,0.12); color: var(--color-success);'
                : latestReport?.status === 'failed'
                  ? 'border-color: rgba(248,81,73,0.24); background-color: rgba(248,81,73,0.12); color: var(--color-danger);'
                  : 'border-color: color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default)); background-color: var(--color-primary-soft); color: var(--color-primary);'"
            >
              {{
                latestReport?.status === 'ready'
                  ? '可下载'
                  : latestReport?.status === 'failed'
                    ? '失败'
                    : latestReport?.status === 'processing'
                      ? '生成中'
                      : '待创建'
              }}
            </span>
          </template>

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

          <div
            v-if="exportError"
            class="rounded-xl border border-red-200 bg-red-50 px-4 py-4 text-sm text-red-600"
          >
            {{ exportError }}
          </div>

          <AppCard v-if="latestReport" variant="panel" accent="primary">
            <div class="flex items-start justify-between gap-4">
              <div>
                <p class="font-medium text-[var(--color-text-primary)]">最近一次导出任务</p>
                <p class="mt-1 text-sm text-[var(--color-text-secondary)]">
                  报告 ID：{{ latestReport.report_id }}
                </p>
                <p
                  v-if="latestReport.expires_at"
                  class="mt-1 text-sm text-[var(--color-text-secondary)]"
                >
                  下载有效期至：{{ formatDate(latestReport.expires_at) }}
                </p>
              </div>
              <span
                class="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-semibold text-emerald-700"
              >
                {{
                  latestReport.status === 'ready'
                    ? '可下载'
                    : latestReport.status === 'failed'
                      ? '失败'
                      : '生成中'
                }}
              </span>
            </div>

            <p v-if="latestReport.error_message" class="text-sm text-rose-400">
              {{ latestReport.error_message }}
            </p>

            <button
              type="button"
              class="w-full rounded-xl border border-[var(--color-border-default)] px-4 py-3 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="latestReport.status !== 'ready'"
              @click="handleDownload"
            >
              下载最近报告
            </button>

            <p
              v-if="latestReport.status === 'processing'"
              class="text-sm text-[var(--color-text-secondary)]"
            >
              {{ polling ? '正在自动刷新导出状态...' : '等待生成完成...' }}
            </p>
          </AppCard>

          <AppCard v-else variant="action" accent="neutral">
            创建报告后，这里会显示最近一次导出任务的状态和下载入口。
          </AppCard>
        </AppCard>
      </SectionCard>
    </div>
  </div>
</template>
