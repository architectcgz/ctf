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

const reportStateMeta = computed(() => {
  if (latestReport.value?.status === 'ready') {
    return {
      label: '可下载',
      borderColor: 'rgba(63,185,80,0.24)',
      badgeBackground: 'rgba(63,185,80,0.12)',
      badgeColor: 'var(--color-success)',
      background: 'linear-gradient(145deg,rgba(20,83,45,0.52),rgba(15,23,42,0.94))',
    }
  }

  if (latestReport.value?.status === 'failed') {
    return {
      label: '失败',
      borderColor: 'rgba(248,81,73,0.24)',
      badgeBackground: 'rgba(248,81,73,0.12)',
      badgeColor: 'var(--color-danger)',
      background: 'linear-gradient(145deg,rgba(127,29,29,0.56),rgba(15,23,42,0.94))',
    }
  }

  if (latestReport.value?.status === 'processing') {
    return {
      label: '生成中',
      borderColor: 'rgba(210,153,34,0.24)',
      badgeBackground: 'rgba(210,153,34,0.12)',
      badgeColor: 'var(--color-warning)',
      background: 'linear-gradient(145deg,rgba(120,53,15,0.48),rgba(15,23,42,0.94))',
    }
  }

  return {
    label: '待创建',
    borderColor: 'color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default))',
    badgeBackground: 'var(--color-primary-soft)',
    badgeColor: 'var(--color-primary)',
    background: 'linear-gradient(145deg,rgba(8,47,73,0.82),rgba(15,23,42,0.94))',
  }
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

        <div
          v-if="profile"
          class="rounded-[30px] border border-cyan-500/20 bg-[linear-gradient(145deg,rgba(8,47,73,0.82),rgba(15,23,42,0.94))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]"
        >
          <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-100/75">
            <span>Profile Snapshot</span>
            <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">{{ profile.role }}</span>
          </div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-white">{{ profile.name || profile.username }}</h2>
          <p class="mt-3 text-sm leading-7 text-cyan-50/80">
            这里聚合你的账号身份、班级归属和基础资料，方便快速确认当前训练账号的信息状态。
          </p>

          <div class="mt-6 grid gap-4 sm:grid-cols-2">
            <div
              v-for="item in profileFields"
              :key="item.label"
              class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4"
            >
              <div class="text-[11px] font-semibold uppercase tracking-[0.18em] text-cyan-100/60">
                {{ item.label }}
              </div>
              <div class="mt-2 text-xl font-semibold text-white">{{ item.value }}</div>
            </div>
          </div>
        </div>

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
        <div
          class="rounded-[30px] border p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]"
          :style="{ borderColor: reportStateMeta.borderColor, background: reportStateMeta.background }"
        >
          <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-white/72">
            <span>Report Export</span>
            <span
              class="rounded-full border px-2 py-1"
              :style="{ borderColor: reportStateMeta.borderColor, backgroundColor: reportStateMeta.badgeBackground, color: reportStateMeta.badgeColor }"
            >
              {{ reportStateMeta.label }}
            </span>
          </div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-white">个人报告导出</h2>
          <p class="mt-3 text-sm leading-7 text-white/80">
            {{ latestReport ? '最近一次导出状态会在这里汇总，准备就绪后可直接下载。' : '先选择导出格式，再创建本次个人报告。' }}
          </p>

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
        </div>
      </SectionCard>
    </div>
  </div>
</template>
