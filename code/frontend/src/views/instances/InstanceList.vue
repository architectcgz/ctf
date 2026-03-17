<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">
        我的实例
      </h1>
      <span class="text-sm text-[var(--color-text-secondary)]">运行中: {{ runningCount }}/{{ maxInstances }}</span>
    </div>

    <div
      v-if="loading"
      class="flex justify-center py-12"
    >
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"
      />
    </div>

    <div
      v-else
      class="space-y-4"
    >
      <div
        v-for="instance in instances"
        :key="instance.id"
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-5"
      >
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
              {{ instance.challenge_title }}
            </h3>
            <div class="flex gap-2">
              <span class="rounded bg-[#06b6d4]/10 px-2 py-0.5 text-xs font-medium text-[#06b6d4]">
                {{ instance.category }}
              </span>
              <span class="rounded bg-[#34d399]/10 px-2 py-0.5 text-xs font-medium text-[#34d399]">
                {{ instance.difficulty }}
              </span>
            </div>
          </div>

          <div class="flex items-center gap-2 text-sm">
            <span :class="getInstanceStatusClass(instance.status)">●</span>
            <span class="text-[var(--color-text-secondary)]">{{
              getInstanceStatusLabel(instance.status)
            }}</span>
          </div>

          <div
            v-if="instance.status === 'running'"
            class="space-y-2 text-sm"
          >
            <div class="flex items-center justify-between">
              <span class="text-[var(--color-text-secondary)]">地址:</span>
              <div class="flex items-center gap-2">
                <span class="font-mono text-[var(--color-text-primary)]">{{
                  instance.access_url ||
                    (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '')
                }}</span>
                <button
                  class="rounded px-2 py-1 text-xs text-[var(--color-primary)] hover:bg-[var(--color-primary)]/10"
                  @click="
                    copyAddress(
                      instance.access_url ||
                        (instance.ssh_info
                          ? `${instance.ssh_info.host}:${instance.ssh_info.port}`
                          : '')
                    )
                  "
                >
                  复制
                </button>
                <button
                  v-if="instance.access_url"
                  class="rounded px-2 py-1 text-xs text-[var(--color-primary)] hover:bg-[var(--color-primary)]/10"
                  @click="openTarget(instance.id)"
                >
                  打开目标
                </button>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-[var(--color-text-secondary)]">剩余:</span>
              <span
                class="font-mono"
                :class="
                  instance.remaining < WARNING_THRESHOLD_SECONDS
                    ? 'text-[#f59e0b] font-semibold'
                    : 'text-[var(--color-text-primary)]'
                "
              >
                {{ formatRemainingTime(instance.remaining) }}
              </span>
            </div>
          </div>

          <div class="flex gap-3">
            <button
              v-if="instance.status === 'running'"
              :disabled="instance.remaining_extends <= 0"
              class="rounded-lg border border-sky-400/40 bg-sky-500/12 px-4 py-2 text-sm font-medium text-sky-950 transition-colors duration-150 hover:bg-sky-500/20 disabled:cursor-not-allowed disabled:border-[var(--color-border-default)] disabled:bg-[var(--color-bg-surface)] disabled:text-[var(--color-text-muted)]"
              @click="extendTime(instance.id)"
            >
              延时 +{{ EXTEND_DURATION_SECONDS / 60 }}min ({{ instance.remaining_extends }})
            </button>
            <button
              class="rounded-lg border border-[#ef4444]/20 bg-[#ef4444]/10 px-4 py-2 text-sm font-medium text-[#f87171] transition-colors duration-150 hover:bg-[#ef4444]/20"
              @click="destroyInstance(instance.id)"
            >
              销毁
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="instances.length === 0"
        class="flex flex-col items-center justify-center py-12 text-center"
      >
        <div class="text-[var(--color-text-muted)] mb-4">
          暂无运行中的实例
        </div>
        <router-link
          to="/challenges"
          class="text-[var(--color-primary)] hover:underline"
        >
          前往靶场列表创建实例
        </router-link>
      </div>
    </div>

    <!-- 超时提醒弹窗 -->
    <div
      v-if="showWarning"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click.self="closeWarning"
    >
      <div
        class="w-full max-w-md rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-xl"
      >
        <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
          实例即将过期
        </h3>
        <p class="mt-2 text-sm text-[var(--color-text-secondary)]">
          实例 "{{ warningInstance?.challenge_title }}" 剩余时间不足 5 分钟，是否延长？
        </p>
        <div class="mt-6 flex justify-end gap-3">
          <button
            class="rounded-lg px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-hover)]"
            @click="closeWarning"
          >
            取消
          </button>
          <button
            class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--color-primary-hover)]"
            @click="extendFromWarning"
          >
            延长 {{ EXTEND_DURATION_SECONDS / 60 }} 分钟
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  EXTEND_DURATION_SECONDS,
  WARNING_THRESHOLD_SECONDS,
  formatRemainingTime,
  getInstanceStatusClass,
  getInstanceStatusLabel,
  useInstanceListPage,
} from '@/composables/useInstanceListPage'

const {
  loading,
  maxInstances,
  instances,
  runningCount,
  showWarning,
  warningInstance,
  copyAddress,
  extendTime,
  openTarget,
  destroyInstance,
  extendFromWarning,
  closeWarning,
} = useInstanceListPage()
</script>
