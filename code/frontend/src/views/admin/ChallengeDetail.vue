<template>
  <section
    class="journal-shell journal-shell-admin journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="flex flex-wrap items-start justify-between gap-4">
      <div>
        <div class="journal-eyebrow">Challenge Detail</div>
        <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)]">
          题目详情
        </h1>
        <p class="mt-3 text-sm leading-7 text-[var(--journal-muted)]">
          查看题目状态、附件、Flag 配置和提示。
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <button
          v-if="route.params.id"
          class="admin-btn admin-btn-primary"
          @click="router.push(`/platform/challenges/${String(route.params.id)}/writeup`)"
        >
          题解管理
        </button>
        <button
          v-if="route.params.id"
          class="admin-btn admin-btn-ghost"
          @click="router.push(`/platform/challenges/${String(route.params.id)}/topology`)"
        >
          拓扑编排
        </button>
        <button class="admin-btn admin-btn-ghost" @click="$router.back()">返回</button>
      </div>
    </div>
    <div class="journal-divider" />

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
      />
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
              {{ flagConfigSummary }}
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
        <ChallengeDescriptionPanel
          v-if="challenge.description"
          class="mt-4"
          :content="challenge.description"
          label="描述"
          test-id="challenge-detail-description"
        />
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
              <div class="mt-2 text-sm text-[var(--color-text-primary)]">
                {{ hint.content }}
              </div>
            </div>
          </div>
        </div>

        <div class="journal-panel mt-6 space-y-5 p-5 md:p-6">
          <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
            <div>
              <div class="journal-eyebrow">Judge Mode</div>
              <h3 class="mt-3 text-lg font-semibold text-[var(--journal-ink)]">判题模式配置</h3>
              <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
                支持静态 Flag、动态前缀、正则判题和人工审核四种模式。保存后即时刷新当前题目配置。
              </p>
            </div>
            <div class="flag-summary-chip">
              {{ flagDraftSummary }}
            </div>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <label class="flag-field">
              <span class="flag-field-label">判题模式</span>
              <select v-model="flagType" class="flag-field-input">
                <option value="static">静态 Flag</option>
                <option value="dynamic">动态前缀</option>
                <option value="regex">正则匹配</option>
                <option value="manual_review">人工审核</option>
              </select>
            </label>

            <label v-if="flagType === 'dynamic' || flagType === 'regex'" class="flag-field">
              <span class="flag-field-label">Flag 前缀</span>
              <input
                v-model="flagPrefix"
                type="text"
                placeholder="例如：flag"
                class="flag-field-input"
              />
            </label>

            <label v-if="flagType === 'static'" class="flag-field md:col-span-2">
              <span class="flag-field-label">静态 Flag</span>
              <input
                v-model="flagValue"
                type="text"
                placeholder="例如：flag{demo}"
                class="flag-field-input font-mono"
              />
            </label>

            <label v-if="flagType === 'regex'" class="flag-field md:col-span-2">
              <span class="flag-field-label">正则表达式</span>
              <input
                v-model="flagRegex"
                type="text"
                placeholder="例如：^flag\\{demo-[0-9]+\\}$"
                class="flag-field-input font-mono"
              />
            </label>
          </div>

          <div
            v-if="flagType === 'manual_review'"
            class="rounded-2xl border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 px-4 py-4 text-sm leading-6 text-[var(--color-text-primary)]"
          >
            学生提交的答案将进入教师审核队列。审核通过后才会计分并更新通过状态。
          </div>

          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="text-sm text-[var(--journal-muted)]">当前配置：{{ flagConfigSummary }}</div>
            <button :disabled="saving" class="admin-btn admin-btn-primary" @click="saveFlagConfig">
              {{ saving ? '保存中...' : '保存配置' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ChallengeDescriptionPanel from '@/components/admin/challenge/ChallengeDescriptionPanel.vue'
import { configureChallengeFlag, getChallengeDetail } from '@/api/admin'
import { useToast } from '@/composables/useToast'
import type { AdminChallengeFlagPayload } from '@/api/admin'
import type { AdminChallengeListItem, FlagType } from '@/api/contracts'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const challenge = ref<AdminChallengeListItem | null>(null)
const flagType = ref<FlagType>('static')
const flagValue = ref('')
const flagRegex = ref('')
const flagPrefix = ref('')

const flagConfigSummary = computed(() => summarizeFlagConfig(challenge.value?.flag_config))
const flagDraftSummary = computed(() =>
  summarizeFlagConfig({
    configured: true,
    flag_type: flagType.value,
    flag_regex: flagRegex.value.trim() || undefined,
    flag_prefix: flagPrefix.value.trim() || undefined,
  })
)

function summarizeFlagConfig(config?: AdminChallengeListItem['flag_config']): string {
  if (!config?.configured) return '未配置'
  switch (config.flag_type) {
    case 'static':
      return '静态 Flag'
    case 'dynamic':
      return `动态 Flag / 前缀 ${config.flag_prefix || 'flag'}`
    case 'regex':
      return `正则匹配 / ${config.flag_regex || '未填写'}`
    case 'manual_review':
      return '人工审核'
    default:
      return '未配置'
  }
}

function hydrateFlagForm(item: AdminChallengeListItem | null): void {
  const config = item?.flag_config
  flagType.value = config?.flag_type ?? 'static'
  flagValue.value = ''
  flagRegex.value = config?.flag_regex ?? ''
  flagPrefix.value = config?.flag_prefix ?? ''
}

async function loadChallenge() {
  const id = route.params.id as string
  try {
    challenge.value = await getChallengeDetail(id)
    hydrateFlagForm(challenge.value)
  } catch (error) {
    toast.error('加载失败')
    setTimeout(() => router.back(), 1500)
  } finally {
    loading.value = false
  }
}

async function saveFlagConfig() {
  const id = route.params.id as string
  const payload: AdminChallengeFlagPayload = {
    flag_type: flagType.value,
  }

  if (flagType.value === 'static') {
    if (!flagValue.value.trim()) {
      toast.error('请填写静态 Flag')
      return
    }
    payload.flag = flagValue.value.trim()
  }

  if (flagType.value === 'dynamic') {
    if (!flagPrefix.value.trim()) {
      toast.error('请填写动态 Flag 前缀')
      return
    }
    payload.flag_prefix = flagPrefix.value.trim()
  }

  if (flagType.value === 'regex') {
    if (!flagRegex.value.trim()) {
      toast.error('请填写正则表达式')
      return
    }
    payload.flag_regex = flagRegex.value.trim()
    if (flagPrefix.value.trim()) {
      payload.flag_prefix = flagPrefix.value.trim()
    }
  }

  saving.value = true
  try {
    await configureChallengeFlag(id, payload)
    toast.success('Flag 配置已保存')
    loading.value = true
    await loadChallenge()
  } catch {
    toast.error('保存 Flag 配置失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadChallenge()
})
</script>

<style scoped>
.journal-shell {
  --journal-divider-border: 1px dashed rgba(148, 163, 184, 0.7);
  --journal-shell-hero-end: color-mix(in srgb, var(--journal-surface) 90%, var(--color-bg-base));
  --journal-shell-dark-accent: #60a5fa;
  --journal-shell-dark-hero-end: color-mix(
    in srgb,
    var(--journal-surface) 92%,
    var(--color-bg-base)
  );
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
}

.admin-btn-primary {
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-ghost {
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.flag-summary-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(37, 99, 235, 0.18);
  background: rgba(37, 99, 235, 0.08);
  padding: var(--space-2) var(--space-3-5);
  font-size: var(--font-size-0-80);
  font-weight: 600;
  color: var(--journal-accent);
}

.flag-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2-5);
}

.flag-field-label {
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: var(--journal-ink);
}

.flag-field-input {
  min-height: 2.9rem;
  border: 1px solid var(--journal-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-92);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.flag-field-input:focus {
  border-color: rgba(37, 99, 235, 0.42);
  box-shadow: 0 0 0 4px rgba(37, 99, 235, 0.12);
}

</style>
