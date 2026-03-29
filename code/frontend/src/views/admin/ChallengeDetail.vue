<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">靶场详情</h1>
      <div class="flex items-center gap-3">
        <button
          v-if="route.params.id"
          class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary)]/90"
          @click="router.push(`/admin/challenges/${String(route.params.id)}/writeup`)"
        >
          题解管理
        </button>
        <button
          v-if="route.params.id"
          class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary)]/90"
          @click="router.push(`/admin/challenges/${String(route.params.id)}/topology`)"
        >
          拓扑编排
        </button>
        <button
          class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[var(--color-bg-elevated)]"
          @click="$router.back()"
        >
          返回
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"
      ></div>
    </div>

    <div v-else-if="challenge" class="space-y-4">
      <div
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6"
      >
        <h2 class="mb-4 text-xl font-semibold text-[var(--color-text-primary)]">
          {{ challenge.title }}
        </h2>
        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="text-[var(--color-text-secondary)]">分类：</span>
            <span class="text-[var(--color-text-primary)]">{{ challenge.category }}</span>
          </div>
          <div>
            <span class="text-[var(--color-text-secondary)]">难度：</span>
            <span class="text-[var(--color-text-primary)]">{{ challenge.difficulty }}</span>
          </div>
          <div>
            <span class="text-[var(--color-text-secondary)]">分值：</span>
            <span class="text-[var(--color-text-primary)]">{{ challenge.points }}</span>
          </div>
          <div>
            <span class="text-[var(--color-text-secondary)]">状态：</span>
            <span class="text-[var(--color-text-primary)]">{{ challenge.status }}</span>
          </div>
          <div v-if="challenge.image_id" class="col-span-2">
            <span class="text-[var(--color-text-secondary)]">镜像：</span>
            <span class="font-mono text-[var(--color-text-primary)]"
              >ID #{{ challenge.image_id }}</span
            >
          </div>
          <div v-if="challenge.flag_config" class="col-span-2">
            <span class="text-[var(--color-text-secondary)]">Flag 配置：</span>
            <span class="font-mono text-[var(--color-text-primary)]">
              {{
                challenge.flag_config.configured
                  ? `${challenge.flag_config.flag_type || 'unknown'} / ${challenge.flag_config.flag_prefix || 'flag'}`
                  : '未配置'
              }}
            </span>
          </div>
          <div v-if="challenge.attachment_url" class="col-span-2">
            <span class="text-[var(--color-text-secondary)]">附件：</span>
            <a
              :href="challenge.attachment_url"
              target="_blank"
              rel="noreferrer"
              class="break-all text-[var(--color-primary)] underline"
            >
              {{ challenge.attachment_url }}
            </a>
          </div>
        </div>
        <div v-if="challenge.description" class="mt-4">
          <div class="text-sm text-[var(--color-text-secondary)]">描述：</div>
          <div class="mt-2 text-sm text-[var(--color-text-primary)]">
            {{ challenge.description }}
          </div>
        </div>
        <div v-if="challenge.hints?.length" class="mt-4">
          <div class="text-sm text-[var(--color-text-secondary)]">提示：</div>
          <div class="mt-2 space-y-3">
            <div
              v-for="hint in challenge.hints"
              :key="hint.id || hint.level"
              class="rounded-lg border border-[var(--color-border-default)] p-3"
            >
              <div class="text-sm font-medium text-[var(--color-text-primary)]">
                Level {{ hint.level }}{{ hint.title ? ` · ${hint.title}` : '' }}
              </div>
              <div v-if="hint.cost_points" class="mt-1 text-xs text-[var(--color-text-secondary)]">
                解锁消耗：{{ hint.cost_points }} 分
              </div>
              <div class="mt-2 text-sm text-[var(--color-text-primary)]">{{ hint.content }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getChallengeDetail } from '@/api/admin'
import { useToast } from '@/composables/useToast'
import type { AdminChallengeListItem } from '@/api/contracts'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const loading = ref(true)
const challenge = ref<AdminChallengeListItem | null>(null)

onMounted(async () => {
  try {
    challenge.value = await getChallengeDetail(route.params.id as string)
  } catch (error) {
    toast.error('加载失败')
    setTimeout(() => router.back(), 1500)
  } finally {
    loading.value = false
  }
})
</script>
