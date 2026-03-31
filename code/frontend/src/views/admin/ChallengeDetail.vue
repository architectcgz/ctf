<template>
  <div class="journal-shell">
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div class="journal-eyebrow">Challenge Detail</div>
          <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">靶场详情</h1>
          <p class="mt-3 text-sm leading-7 text-[var(--journal-muted)]">查看题目状态、附件、Flag 配置和提示。</p>
        </div>
        <div class="flex flex-wrap items-center gap-3">
        <button
          v-if="route.params.id"
          class="admin-btn admin-btn-primary"
          @click="router.push(`/admin/challenges/${String(route.params.id)}/writeup`)"
        >
          题解管理
        </button>
        <button
          v-if="route.params.id"
          class="admin-btn admin-btn-ghost"
          @click="router.push(`/admin/challenges/${String(route.params.id)}/topology`)"
        >
          拓扑编排
        </button>
        <button
          class="admin-btn admin-btn-ghost"
          @click="$router.back()"
        >
          返回
        </button>
        </div>
      </div>
      <div class="journal-divider" />

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
      ></div>
    </div>

    <div v-else-if="challenge" class="space-y-3">
      <div class="space-y-3">
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
              class="rounded-[16px] border border-[var(--journal-border)] bg-[var(--journal-surface)] p-3"
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
    </section>
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

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #2563eb;
  --journal-border: rgba(226, 232, 240, 0.84);
  --journal-surface: rgba(248, 250, 252, 0.92);
}

.journal-hero,
.journal-panel {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-radius: 16px !important;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.admin-btn-primary {
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-ghost {
  border: 1px solid var(--journal-border);
  background: rgba(255, 255, 255, 0.75);
  color: var(--journal-ink);
}
:global([data-theme='dark']) .journal-hero,
:global([data-theme='dark']) .journal-panel {
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}

.journal-divider {
  margin-block: 1rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.7);
}
:global([data-theme='dark']) .journal-shell {
  --journal-ink: #e2e8f0;
  --journal-muted: #94a3b8;
  --journal-accent: #60a5fa;
  --journal-border: rgba(71, 85, 105, 0.78);
  --journal-surface: rgba(15, 23, 42, 0.7);
}

</style>
