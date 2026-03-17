<script setup lang="ts">
import { RefreshCw, Save, Trash2 } from 'lucide-vue-next'

import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useChallengeWriteupEditorPage } from '@/composables/useChallengeWriteupEditorPage'

const props = defineProps<{
  challengeId: string
}>()

const emit = defineEmits<{
  back: []
}>()

const {
  loading,
  saving,
  deleting,
  challenge,
  writeup,
  form,
  hasWriteup,
  visibilityLabel,
  loadPage,
  handleSave,
  handleDelete,
  restoreExistingWriteup,
} = useChallengeWriteupEditorPage(props.challengeId)
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      title="题解管理"
      eyebrow="Admin Writeup"
      :description="
        challenge
          ? `为《${challenge.title}》维护管理员题解，控制公开范围与发布时间。`
          : '为挑战维护管理员题解，控制公开范围与发布时间。'
      "
    >
      <button
        class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[var(--color-bg-surface)]"
        @click="emit('back')"
      >
        返回挑战
      </button>
      <button
        class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[var(--color-bg-surface)]"
        @click="void loadPage()"
      >
        <RefreshCw class="mr-2 inline-flex h-4 w-4" />
        刷新
      </button>
    </PageHeader>

    <AppLoading v-if="loading">
      正在加载题解数据...
    </AppLoading>

    <div
      v-else
      class="grid gap-6 xl:grid-cols-[320px_minmax(0,1fr)]"
    >
      <AppCard
        accent="neutral"
        title="挑战信息"
        subtitle="当前编辑对象"
        eyebrow="Challenge"
      >
        <div
          v-if="challenge"
          class="space-y-4 text-sm text-[var(--color-text-secondary)]"
        >
          <div>
            <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
              标题
            </div>
            <div class="mt-2 text-base font-semibold text-[var(--color-text-primary)]">
              {{ challenge.title }}
            </div>
          </div>
          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
            <div>
              <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
                分类
              </div>
              <div class="mt-2 text-[var(--color-text-primary)]">
                {{ challenge.category }}
              </div>
            </div>
            <div>
              <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
                状态
              </div>
              <div class="mt-2 text-[var(--color-text-primary)]">
                {{ challenge.status }}
              </div>
            </div>
            <div>
              <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
                难度
              </div>
              <div class="mt-2 text-[var(--color-text-primary)]">
                {{ challenge.difficulty }}
              </div>
            </div>
            <div>
              <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
                分值
              </div>
              <div class="mt-2 text-[var(--color-text-primary)]">
                {{ challenge.points }}
              </div>
            </div>
          </div>
        </div>
      </AppCard>

      <AppCard
        variant="hero"
        accent="primary"
        title="题解编辑器"
        subtitle="支持 private / public / scheduled 三种可见性。"
        eyebrow="Writeup"
      >
        <template #header>
          <div class="flex items-center gap-2">
            <span
              class="rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-[0.22em]"
              :class="
                hasWriteup
                  ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'
                  : 'border-amber-500/30 bg-amber-500/10 text-amber-300'
              "
            >
              {{ hasWriteup ? '已存在题解' : '尚未创建' }}
            </span>
          </div>
        </template>

        <div class="space-y-5">
          <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_220px]">
            <label class="space-y-2">
              <span class="text-sm font-medium text-[var(--color-text-primary)]">题解标题</span>
              <input
                v-model="form.title"
                type="text"
                class="w-full rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-primary)]"
                placeholder="例如：官方解题思路 / 赛后复盘"
              >
            </label>

            <label class="space-y-2">
              <span class="text-sm font-medium text-[var(--color-text-primary)]">可见性</span>
              <select
                v-model="form.visibility"
                class="w-full rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-primary)]"
              >
                <option value="private">private</option>
                <option value="public">public</option>
                <option value="scheduled">scheduled</option>
              </select>
            </label>
          </div>

          <div
            class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)]/70 p-4 text-sm text-[var(--color-text-secondary)]"
          >
            {{ visibilityLabel }}
          </div>

          <label
            v-if="form.visibility === 'scheduled'"
            class="block space-y-2"
          >
            <span class="text-sm font-medium text-[var(--color-text-primary)]">发布时间</span>
            <input
              v-model="form.releaseAt"
              type="datetime-local"
              class="w-full rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-primary)]"
            >
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-medium text-[var(--color-text-primary)]">题解正文</span>
            <textarea
              v-model="form.content"
              rows="16"
              class="min-h-[360px] w-full rounded-3xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-4 py-4 text-sm leading-7 text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-primary)]"
              placeholder="输入官方题解、赛后复盘或教学讲解内容。"
            />
          </label>

          <div class="flex flex-wrap items-center gap-3">
            <button
              :disabled="saving"
              class="rounded-2xl bg-[var(--color-primary)] px-5 py-3 text-sm font-semibold text-white transition-colors hover:bg-[var(--color-primary)]/90 disabled:cursor-not-allowed disabled:opacity-60"
              @click="void handleSave()"
            >
              <Save class="mr-2 inline-flex h-4 w-4" />
              {{ saving ? '保存中...' : '保存题解' }}
            </button>
            <button
              v-if="hasWriteup"
              class="rounded-2xl border border-[var(--color-border-default)] px-5 py-3 text-sm font-medium text-[var(--color-text-primary)] transition-colors hover:bg-[var(--color-bg-surface)]"
              @click="restoreExistingWriteup"
            >
              恢复已保存版本
            </button>
            <button
              v-if="hasWriteup"
              :disabled="deleting"
              class="rounded-2xl border border-red-500/30 bg-red-500/10 px-5 py-3 text-sm font-medium text-red-300 transition-colors hover:bg-red-500/20 disabled:cursor-not-allowed disabled:opacity-60"
              @click="void handleDelete()"
            >
              <Trash2 class="mr-2 inline-flex h-4 w-4" />
              {{ deleting ? '删除中...' : '删除题解' }}
            </button>
          </div>
        </div>
      </AppCard>

      <AppCard
        v-if="writeup"
        accent="success"
        title="当前已保存版本"
        subtitle="保存成功后的元数据会显示在这里。"
        eyebrow="Snapshot"
        class="xl:col-span-2"
      >
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-2xl border border-border-subtle bg-[var(--color-bg-surface)]/70 p-4">
            <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
              标题
            </div>
            <div class="mt-2 text-sm font-semibold text-[var(--color-text-primary)]">
              {{ writeup.title }}
            </div>
          </div>
          <div class="rounded-2xl border border-border-subtle bg-[var(--color-bg-surface)]/70 p-4">
            <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
              可见性
            </div>
            <div class="mt-2 text-sm font-semibold text-[var(--color-text-primary)]">
              {{ writeup.visibility }}
            </div>
          </div>
          <div class="rounded-2xl border border-border-subtle bg-[var(--color-bg-surface)]/70 p-4">
            <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
              创建时间
            </div>
            <div class="mt-2 text-sm font-semibold text-[var(--color-text-primary)]">
              {{ writeup.created_at }}
            </div>
          </div>
          <div class="rounded-2xl border border-border-subtle bg-[var(--color-bg-surface)]/70 p-4">
            <div class="text-xs uppercase tracking-[0.2em] text-[var(--color-text-muted)]">
              更新时间
            </div>
            <div class="mt-2 text-sm font-semibold text-[var(--color-text-primary)]">
              {{ writeup.updated_at }}
            </div>
          </div>
        </div>
      </AppCard>

      <AppCard
        v-else
        accent="warning"
        class="xl:col-span-2"
      >
        <AppEmpty
          title="当前还没有管理员题解"
          description="填写表单后点击保存，即可创建题解并控制公开范围。"
          icon="BookOpen"
        />
      </AppCard>
    </div>
  </div>
</template>
