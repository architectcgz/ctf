<template>
  <section class="journal-shell journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Challenge Authoring</div>
          <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            挑战管理
          </h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            在这里查看题目状态，并继续进入详情、编排和题解。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button class="admin-btn admin-btn-primary" @click="void openDialog()">
              创建挑战
            </button>
          </div>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="journal-note-label">题库概况</div>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div class="journal-note">
              <div class="journal-note-label">题目总量</div>
              <div class="journal-note-value">{{ total }}</div>
              <div class="journal-note-helper">当前题库中可管理的题目</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">当前页</div>
              <div class="journal-note-value">{{ list.length }}</div>
              <div class="journal-note-helper">当前分页中的题目数量</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">已发布</div>
              <div class="journal-note-value">{{ publishedCount }}</div>
              <div class="journal-note-helper">当前页已开放训练的题目</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">草稿</div>
              <div class="journal-note-value">{{ draftCount }}</div>
              <div class="journal-note-helper">仍待发布或继续补充的题目</div>
            </div>
          </div>
        </article>
      </div>
      <div class="journal-divider" />

      <div class="space-y-3">
      <div
        v-if="loading"
        class="flex items-center justify-center py-12"
      >
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        />
      </div>

      <template v-else>
        <div v-if="list.length === 0" class="admin-empty">
          当前还没有题目。
        </div>

        <div v-else class="space-y-3">
          <article v-for="row in list" :key="row.id" class="challenge-row">
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="text-base font-semibold text-[var(--journal-ink)]">{{ row.title }}</h2>
                  <span
                    class="admin-status-chip"
                    :style="{
                      backgroundColor: getStatusColor(row.status) + '18',
                      color: getStatusColor(row.status),
                    }"
                  >
                    {{ getStatusLabel(row.status) }}
                  </span>
                </div>

                <div class="mt-3 flex flex-wrap gap-2">
                  <span
                    class="admin-inline-chip"
                    :style="{
                      backgroundColor: getCategoryColor(row.category) + '16',
                      color: getCategoryColor(row.category),
                    }"
                  >
                    {{ getCategoryLabel(row.category) }}
                  </span>
                  <span
                    class="admin-inline-chip"
                    :style="{
                      backgroundColor: getDifficultyColor(row.difficulty) + '16',
                      color: getDifficultyColor(row.difficulty),
                    }"
                  >
                    {{ getDifficultyLabel(row.difficulty) }}
                  </span>
                  <span class="admin-inline-chip admin-inline-chip-neutral">{{ row.points }} pts</span>
                </div>
              </div>
            </div>

            <div class="journal-divider mt-4" />

            <div class="mt-4 flex flex-wrap gap-2">
              <button class="admin-btn admin-btn-ghost admin-btn-compact" @click="$router.push(`/admin/challenges/${row.id}`)">
                查看
              </button>
              <button class="admin-btn admin-btn-ghost admin-btn-compact" @click="$router.push(`/admin/challenges/${row.id}/topology`)">
                编排
              </button>
              <button class="admin-btn admin-btn-ghost admin-btn-compact" @click="$router.push(`/admin/challenges/${row.id}/writeup`)">
                题解
              </button>
              <button class="admin-btn admin-btn-primary admin-btn-compact" @click="void openDialog(row)">
                编辑
              </button>
              <button
                v-if="row.status !== 'published'"
                class="admin-btn admin-btn-success admin-btn-compact"
                @click="void publish(row)"
              >
                发布
              </button>
              <button
                class="admin-btn admin-btn-danger admin-btn-compact"
                @click="void remove(row.id)"
              >
                删除
              </button>
            </div>
          </article>
        </div>

        <div
          v-if="total > 0"
          class="admin-pagination mt-4"
        >
          <span>共 {{ total }} 条</span>
          <div class="flex items-center gap-2">
            <button
              :disabled="page === 1"
              class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
              @click="changePage(page - 1)"
            >
              上一页
            </button>
            <span>{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
            <button
              :disabled="page >= Math.ceil(total / pageSize)"
              class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
              @click="changePage(page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </template>
      </div>

    <ElDialog
      v-model="dialogVisible"
      :title="editingId ? '编辑挑战' : '创建挑战'"
      width="620px"
    >
      <ElForm
        :model="form"
        label-width="110px"
      >
        <ElFormItem
          label="标题"
          required
        >
          <ElInput v-model="form.title" />
        </ElFormItem>
        <ElFormItem
          label="分类"
          required
        >
          <ElSelect v-model="form.category">
            <ElOption
              label="Web"
              value="web"
            />
            <ElOption
              label="Pwn"
              value="pwn"
            />
            <ElOption
              label="逆向"
              value="reverse"
            />
            <ElOption
              label="密码"
              value="crypto"
            />
            <ElOption
              label="杂项"
              value="misc"
            />
            <ElOption
              label="取证"
              value="forensics"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem
          label="难度"
          required
        >
          <ElSelect v-model="form.difficulty">
            <ElOption
              label="入门"
              value="beginner"
            />
            <ElOption
              label="简单"
              value="easy"
            />
            <ElOption
              label="中等"
              value="medium"
            />
            <ElOption
              label="困难"
              value="hard"
            />
            <ElOption
              label="地狱"
              value="insane"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem
          label="分值"
          required
        >
          <ElInputNumber
            v-model="form.points"
            :min="10"
            :max="1000"
          />
        </ElFormItem>
        <ElFormItem label="镜像">
          <ElSelect
            v-model="form.image_id"
            placeholder="可选，不选则题目无需靶机"
            clearable
          >
            <ElOption
              label="不需要靶机"
              value=""
            />
            <ElOption
              v-for="img in images"
              :key="img.id"
              :label="`${img.name}:${img.tag}`"
              :value="img.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="靶机描述"
          />
        </ElFormItem>
        <ElFormItem label="附件地址">
          <ElInput
            v-model="form.attachment_url"
            placeholder="可选，例如：https://example.com/files/challenge.zip"
          />
        </ElFormItem>
        <ElFormItem label="提示系统">
          <div class="w-full space-y-3">
            <div
              v-for="(hint, index) in form.hints"
              :key="`${hint.id || 'new'}-${index}`"
              class="rounded-lg border border-[var(--color-border-default)] p-3"
            >
              <div class="mb-3 flex items-center justify-between">
                <span class="text-sm font-medium text-[var(--color-text-primary)]">提示 {{ index + 1 }}</span>
                <button
                  type="button"
                  class="text-xs text-[var(--color-danger)] transition-colors hover:text-[var(--color-danger)]"
                  @click="removeHint(index)"
                >
                  删除
                </button>
              </div>
              <div class="grid gap-3 md:grid-cols-3">
                <ElInputNumber
                  v-model="hint.level"
                  :min="1"
                  controls-position="right"
                />
                <ElInput
                  v-model="hint.title"
                  placeholder="提示标题（可选）"
                />
                <ElInputNumber
                  v-model="hint.cost_points"
                  :min="0"
                  controls-position="right"
                />
              </div>
              <ElInput
                v-model="hint.content"
                class="mt-3"
                type="textarea"
                :rows="3"
                placeholder="提示内容"
              />
            </div>
            <button
              type="button"
              class="rounded border border-dashed border-[var(--color-border-default)] px-3 py-2 text-sm text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
              @click="addHint"
            >
              添加提示
            </button>
          </div>
        </ElFormItem>
        <ElFormItem label="Flag 类型">
          <ElSelect v-model="form.flag_type">
            <ElOption
              label="静态 Flag"
              value="static"
            />
            <ElOption
              label="动态 Flag"
              value="dynamic"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem
          v-if="form.flag_type === 'static'"
          label="静态 Flag"
          required
        >
          <ElInput
            v-model="form.flag"
            :placeholder="editingId ? '留空表示保持当前静态 Flag 不变' : '例如：flag{sqli_success}'"
          />
        </ElFormItem>
        <ElFormItem label="Flag 前缀">
          <ElInput
            v-model="form.flag_prefix"
            placeholder="默认 flag"
          />
        </ElFormItem>
        <ElFormItem label="当前状态">
          <div class="text-sm text-[var(--color-text-secondary)]">
            {{ getStatusLabel(form.current_status) }}
          </div>
        </ElFormItem>
        <ElFormItem>
          <label class="flex items-center gap-2 text-sm text-[var(--color-text-primary)]">
            <input
              v-model="form.publish_after_save"
              type="checkbox"
              :disabled="form.current_status === 'published'"
            >
            保存后立即发布
          </label>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <button
          class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[var(--color-bg-elevated)]"
          @click="dialogVisible = false"
        >
          取消
        </button>
        <button
          :disabled="saving"
          class="ml-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary)]/90 disabled:cursor-not-allowed disabled:opacity-50"
          @click="void saveChallenge()"
        >
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </template>
    </ElDialog>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAdminChallenges } from '@/composables/useAdminChallenges'
import type { ChallengeCategory, ChallengeDifficulty, ChallengeStatus } from '@/api/contracts'
const {
  list,
  total,
  page,
  pageSize,
  loading,
  changePage,
  dialogVisible,
  saving,
  editingId,
  images,
  form,
  openDialog,
  addHint,
  removeHint,
  saveChallenge,
  publish,
  remove,
} = useAdminChallenges()

const publishedCount = computed(() => list.value.filter((item) => item.status === 'published').length)
const draftCount = computed(() => list.value.filter((item) => item.status === 'draft').length)

function getCategoryLabel(category: ChallengeCategory): string {
  const labels: Record<ChallengeCategory, string> = {
    web: 'Web',
    pwn: 'Pwn',
    reverse: '逆向',
    crypto: '密码',
    misc: '杂项',
    forensics: '取证',
  }
  return labels[category]
}

function getCategoryColor(category: ChallengeCategory): string {
  return {
    web: '#3b82f6',
    pwn: '#ef4444',
    reverse: '#8b5cf6',
    crypto: '#f59e0b',
    misc: '#10b981',
    forensics: '#06b6d4',
  }[category]
}

function getDifficultyLabel(difficulty: ChallengeDifficulty): string {
  const labels: Record<ChallengeDifficulty, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    insane: '地狱',
  }
  return labels[difficulty]
}

function getDifficultyColor(difficulty: ChallengeDifficulty): string {
  return {
    beginner: '#10b981',
    easy: '#3b82f6',
    medium: '#f59e0b',
    hard: '#ef4444',
    insane: '#7c3aed',
  }[difficulty]
}

function getStatusLabel(status: ChallengeStatus): string {
  return { draft: '草稿', published: '已发布', archived: '已归档' }[status]
}

function getStatusColor(status: ChallengeStatus): string {
  return { draft: '#8b949e', published: '#10b981', archived: '#6e7681' }[status]
}
</script>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #2563eb;
  --journal-border: rgba(226, 232, 240, 0.84);
  --journal-surface: rgba(248, 250, 252, 0.92);
  --journal-surface-subtle: rgba(241, 245, 249, 0.72);
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

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.journal-divider {
  margin-block: 1rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.7);
}

.challenge-row {
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.74);
  padding: 1rem;
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-compact {
  min-height: 2.35rem;
  padding: 0.5rem 0.85rem;
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

.admin-btn-success {
  border: 1px solid rgba(16, 185, 129, 0.18);
  background: rgba(236, 253, 245, 0.9);
  color: #059669;
}

.admin-btn-danger {
  border: 1px solid rgba(239, 68, 68, 0.2);
  background: rgba(254, 242, 242, 0.9);
  color: #dc2626;
}

.admin-status-chip,
.admin-inline-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-radius: 999px;
  padding: 0.34rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 600;
}

.admin-inline-chip {
  border: 1px solid transparent;
}

.admin-inline-chip-neutral {
  background: rgba(241, 245, 249, 0.95);
  color: var(--journal-muted);
}

.admin-empty {
  border: 1px dashed rgba(148, 163, 184, 0.72);
  border-radius: 16px;
  padding: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

.admin-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.72);
  padding-top: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #e2e8f0;
  --journal-muted: #94a3b8;
  --journal-accent: #60a5fa;
  --journal-border: rgba(71, 85, 105, 0.78);
  --journal-surface: rgba(15, 23, 42, 0.7);
  --journal-surface-subtle: rgba(15, 23, 42, 0.78);
}

:global([data-theme='dark']) .journal-hero,
:global([data-theme='dark']) .journal-panel {
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}
</style>
