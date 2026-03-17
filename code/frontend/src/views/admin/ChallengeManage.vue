<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">
        挑战管理
      </h1>
      <button
        class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[var(--color-primary)]/90"
        @click="void openDialog()"
      >
        创建挑战
      </button>
    </div>

    <div
      v-if="loading"
      class="flex items-center justify-center py-12"
    >
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"
      />
    </div>

    <div
      v-else
      class="overflow-hidden rounded-lg border border-[var(--color-border-default)]"
    >
      <table class="w-full">
        <thead class="bg-[var(--color-bg-surface)]">
          <tr>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">
              标题
            </th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">
              分类
            </th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">
              难度
            </th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">
              分值
            </th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">
              状态
            </th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">
              操作
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#30363d]">
          <tr
            v-for="row in list"
            :key="row.id"
            class="transition-colors hover:bg-[var(--color-bg-elevated)]"
          >
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">
              {{ row.title }}
            </td>
            <td class="px-4 py-3">
              <span
                class="rounded px-2 py-1 text-xs font-medium"
                :style="{
                  backgroundColor: getCategoryColor(row.category) + '20',
                  color: getCategoryColor(row.category),
                }"
              >
                {{ getCategoryLabel(row.category) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span
                class="rounded px-2 py-1 text-xs font-medium"
                :style="{
                  backgroundColor: getDifficultyColor(row.difficulty) + '20',
                  color: getDifficultyColor(row.difficulty),
                }"
              >
                {{ getDifficultyLabel(row.difficulty) }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">
              {{ row.points }}
            </td>
            <td class="px-4 py-3">
              <span
                class="rounded px-2 py-1 text-xs font-medium"
                :style="{
                  backgroundColor: getStatusColor(row.status) + '20',
                  color: getStatusColor(row.status),
                }"
              >
                {{ getStatusLabel(row.status) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button
                  class="rounded bg-[var(--color-primary)] px-3 py-1 text-xs text-white transition-colors hover:bg-[var(--color-primary)]/90"
                  @click="$router.push(`/admin/challenges/${row.id}`)"
                >
                  查看
                </button>
                <button
                  class="rounded border border-[var(--color-primary)]/40 px-3 py-1 text-xs text-[var(--color-primary)] transition-colors hover:bg-[var(--color-primary)]/10"
                  @click="$router.push(`/admin/challenges/${row.id}/topology`)"
                >
                  编排
                </button>
                <button
                  class="rounded border border-[var(--color-primary)]/40 px-3 py-1 text-xs text-[var(--color-primary)] transition-colors hover:bg-[var(--color-primary)]/10"
                  @click="$router.push(`/admin/challenges/${row.id}/writeup`)"
                >
                  题解
                </button>
                <button
                  class="rounded bg-[var(--color-primary)] px-3 py-1 text-xs text-white transition-colors hover:bg-[var(--color-primary)]/90"
                  @click="void openDialog(row)"
                >
                  编辑
                </button>
                <button
                  v-if="row.status !== 'published'"
                  class="rounded bg-emerald-500/20 px-3 py-1 text-xs text-emerald-500 transition-colors hover:bg-emerald-500/30"
                  @click="void publish(row)"
                >
                  发布
                </button>
                <button
                  class="rounded bg-red-500/20 px-3 py-1 text-xs text-red-500 transition-colors hover:bg-red-500/30"
                  @click="void remove(row.id)"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div
      v-if="!loading && total > 0"
      class="flex items-center justify-between"
    >
      <span class="text-sm text-[var(--color-text-secondary)]">共 {{ total }} 条</span>
      <div class="flex items-center gap-2">
        <button
          :disabled="page === 1"
          class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page - 1)"
        >
          上一页
        </button>
        <span class="text-sm text-[var(--color-text-secondary)]">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button
          :disabled="page >= Math.ceil(total / pageSize)"
          class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page + 1)"
        >
          下一页
        </button>
      </div>
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
                  class="text-xs text-red-500 transition-colors hover:text-red-400"
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
          class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[#21262d]"
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
  </div>
</template>

<script setup lang="ts">
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
