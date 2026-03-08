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
const latestReportFormat = ref<'pdf' | 'excel'>('pdf')
const latestReportCreatedAt = ref<string | null>(null)
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

const reportTaskMeta = computed(() => {
  if (latestReport.value?.status === 'ready') {
    return {
      label: '可下载',
      accent: 'success' as const,
      hint: '报告已生成，可直接下载。',
      chipClass: 'bg-emerald-500/12 text-emerald-600',
    }
  }

  if (latestReport.value?.status === 'failed') {
    return {
      label: '失败',
      accent: 'danger' as const,
      hint: latestReport.value.error_message || '报告生成失败，请重新创建导出任务。',
      chipClass: 'bg-rose-500/12 text-rose-600',
    }
  }

  if (latestReport.value?.status === 'processing') {
    return {
      label: '生成中',
      accent: 'warning' as const,
      hint: polling ? '正在自动刷新导出状态。' : '报告任务正在处理中。',
      chipClass: 'bg-amber-500/12 text-amber-600',
    }
  }

  return {
    label: '待创建',
    accent: 'primary' as const,
    hint: '先选择导出格式，再创建本次个人报告。',
    chipClass: 'bg-cyan-500/12 text-cyan-600',
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
    latestReportFormat.value = reportFormat.value
    latestReportCreatedAt.value = new Date().toISOString()
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
          <div
            class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-100/75"
          >
            <span>Profile Snapshot</span>
            <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">{{
              profile.role
            }}</span>
          </div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-white">
            {{ profile.name || profile.username }}
          </h2>
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
      </SectionCard>

      <SectionCard title="个人报告导出" subtitle="生成 PDF 或 Excel 报告，便于训练复盘和归档。">
        <AppCard
          variant="hero"
          accent="primary"
          eyebrow="Personal Export"
          title="个人报告生成器"
          subtitle="选择导出格式后创建任务。系统会自动跟踪最近一次任务状态，并在就绪后提供下载。"
        >
          <template #header>
            <span
              class="rounded-full px-3 py-1 text-xs font-semibold"
              :class="reportTaskMeta.chipClass"
            >
              {{ reportTaskMeta.label }}
            </span>
          </template>

          <fieldset>
            <legend class="mb-2 text-sm font-medium text-[var(--color-text-primary)]">
              导出格式
            </legend>
            <div class="grid gap-3 sm:grid-cols-2">
              <label
                class="flex items-start gap-3 rounded-xl border px-4 py-4 transition"
                :class="
                  reportFormat === 'pdf'
                    ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/8'
                    : 'border-[var(--color-border-default)] bg-[var(--color-bg-base)] hover:border-[var(--color-primary)]/50'
                "
              >
                <input v-model="reportFormat" type="radio" value="pdf" class="mt-1" />
                <span>
                  <span class="block font-medium text-[var(--color-text-primary)]">PDF</span>
                  <span class="mt-1 block text-sm text-[var(--color-text-secondary)]"
                    >适合打印、归档和正式复盘。</span
                  >
                </span>
              </label>

              <label
                class="flex items-start gap-3 rounded-xl border px-4 py-4 transition"
                :class="
                  reportFormat === 'excel'
                    ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/8'
                    : 'border-[var(--color-border-default)] bg-[var(--color-bg-base)] hover:border-[var(--color-primary)]/50'
                "
              >
                <input v-model="reportFormat" type="radio" value="excel" class="mt-1" />
                <span>
                  <span class="block font-medium text-[var(--color-text-primary)]">Excel</span>
                  <span class="mt-1 block text-sm text-[var(--color-text-secondary)]"
                    >适合继续整理、筛选和二次分析。</span
                  >
                </span>
              </label>
            </div>
          </fieldset>

          <AppCard variant="action" accent="neutral">
            {{ reportTaskMeta.hint }}
          </AppCard>

          <button
            type="button"
            class="inline-flex items-center rounded-xl bg-[var(--color-primary)] px-5 py-3 text-sm font-medium text-white transition hover:bg-[var(--color-primary-hover)] disabled:cursor-not-allowed disabled:opacity-60"
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
        </AppCard>

        <AppCard
          v-if="latestReport"
          variant="hero"
          :accent="reportTaskMeta.accent"
          eyebrow="Latest Task"
          :title="String(latestReport.report_id)"
          :subtitle="reportTaskMeta.hint"
        >
          <template #header>
            <span
              class="rounded-full px-3 py-1 text-xs font-semibold"
              :class="reportTaskMeta.chipClass"
            >
              {{ reportTaskMeta.label }}
            </span>
          </template>

          <div class="grid grid-cols-2 gap-3 text-sm">
            <AppCard
              variant="metric"
              accent="primary"
              eyebrow="格式"
              :title="latestReportFormat.toUpperCase()"
            />
            <AppCard
              variant="metric"
              :accent="reportTaskMeta.accent"
              eyebrow="任务状态"
              :title="reportTaskMeta.label"
            />
            <AppCard
              variant="metric"
              accent="neutral"
              eyebrow="创建时间"
              :title="latestReportCreatedAt ? formatDate(latestReportCreatedAt) : '刚刚创建'"
            />
            <AppCard
              variant="metric"
              accent="neutral"
              eyebrow="有效期"
              :title="
                latestReport.expires_at ? formatDate(latestReport.expires_at) : '待生成完成后返回'
              "
            />
          </div>

          <p v-if="latestReport.error_message" class="text-sm text-rose-400">
            {{ latestReport.error_message }}
          </p>

          <button
            type="button"
            class="inline-flex items-center rounded-xl border border-[var(--color-border-default)] px-4 py-2.5 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="latestReport.status !== 'ready'"
            @click="handleDownload"
          >
            下载最近报告
          </button>
        </AppCard>

        <AppEmpty
          v-else
          title="还没有创建个人报告"
          description="生成一次个人报告后，这里会展示最近一次任务状态和下载入口。"
          icon="FileChartColumnIncreasing"
        />
      </SectionCard>
    </div>
  </div>
</template>
