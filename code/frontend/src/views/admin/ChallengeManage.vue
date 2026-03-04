<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[#c9d1d9]">挑战管理</h1>
      <button class="rounded-lg bg-[#0891b2] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[#0891b2]/90" @click="openDialog()">
        创建挑战
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[#30363d]">
      <table class="w-full">
        <thead class="bg-[#161b22]">
          <tr>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">标题</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">分类</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">难度</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">分值</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">状态</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#30363d]">
          <tr v-for="row in list" :key="row.id" class="transition-colors hover:bg-[#1c2128]">
            <td class="px-4 py-3 text-sm text-[#c9d1d9]">{{ row.title }}</td>
            <td class="px-4 py-3">
              <span class="rounded px-2 py-1 text-xs font-medium" :style="{ backgroundColor: getCategoryColor(row.category) + '20', color: getCategoryColor(row.category) }">
                {{ getCategoryLabel(row.category) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span class="rounded px-2 py-1 text-xs font-medium" :style="{ backgroundColor: getDifficultyColor(row.difficulty) + '20', color: getDifficultyColor(row.difficulty) }">
                {{ getDifficultyLabel(row.difficulty) }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-[#c9d1d9]">{{ row.base_score }}</td>
            <td class="px-4 py-3">
              <span class="rounded px-2 py-1 text-xs font-medium" :style="{ backgroundColor: getStatusColor(row.status) + '20', color: getStatusColor(row.status) }">
                {{ getStatusLabel(row.status) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button class="rounded bg-[#0891b2] px-3 py-1 text-xs text-white transition-colors hover:bg-[#0891b2]/90" @click="openDialog(row)">
                  编辑
                </button>
                <button class="rounded bg-red-500/20 px-3 py-1 text-xs text-red-500 transition-colors hover:bg-red-500/30" @click="handleDelete(row.id)">
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="!loading && total > 0" class="flex items-center justify-between">
      <span class="text-sm text-[#8b949e]">共 {{ total }} 条</span>
      <div class="flex items-center gap-2">
        <button
          :disabled="page === 1"
          class="rounded-lg border border-[#30363d] px-3 py-1.5 text-sm text-[#c9d1d9] transition-colors hover:border-[#0891b2] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page - 1)"
        >
          上一页
        </button>
        <span class="text-sm text-[#8b949e]">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button
          :disabled="page >= Math.ceil(total / pageSize)"
          class="rounded-lg border border-[#30363d] px-3 py-1.5 text-sm text-[#c9d1d9] transition-colors hover:border-[#0891b2] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page + 1)"
        >
          下一页
        </button>
      </div>
    </div>

    <ElDialog v-model="dialogVisible" :title="editingId ? '编辑挑战' : '创建挑战'" width="600px">
      <ElForm :model="form" label-width="100px">
        <ElFormItem label="标题" required>
          <ElInput v-model="form.title" />
        </ElFormItem>
        <ElFormItem label="分类" required>
          <ElSelect v-model="form.category">
            <ElOption label="Web" value="web" />
            <ElOption label="Pwn" value="pwn" />
            <ElOption label="逆向" value="reverse" />
            <ElOption label="密码" value="crypto" />
            <ElOption label="杂项" value="misc" />
            <ElOption label="取证" value="forensics" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="难度" required>
          <ElSelect v-model="form.difficulty">
            <ElOption label="入门" value="beginner" />
            <ElOption label="简单" value="easy" />
            <ElOption label="中等" value="medium" />
            <ElOption label="困难" value="hard" />
            <ElOption label="地狱" value="hell" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="分值" required>
          <ElInputNumber v-model="form.base_score" :min="10" :max="1000" />
        </ElFormItem>
        <ElFormItem label="状态" required>
          <ElSelect v-model="form.status">
            <ElOption label="草稿" value="draft" />
            <ElOption label="审核中" value="review" />
            <ElOption label="已发布" value="active" />
            <ElOption label="已归档" value="archived" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <button class="rounded-lg border border-[#30363d] px-4 py-2 text-sm text-[#c9d1d9] transition-colors hover:bg-[#21262d]" @click="dialogVisible = false">
          取消
        </button>
        <button
          :disabled="saving"
          class="ml-2 rounded-lg bg-[#0891b2] px-4 py-2 text-sm text-white transition-colors hover:bg-[#0891b2]/90 disabled:cursor-not-allowed disabled:opacity-50"
          @click="handleSave"
        >
          {{ saving ? '保存中...' : '保存' }}
        </button>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import { getChallenges, createChallenge, updateChallenge, deleteChallenge } from '@/api/admin'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import type { ChallengeCategory, ChallengeDifficulty, ChallengeStatus } from '@/api/contracts'

const toast = useToast()
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)
const form = reactive({
  title: '',
  category: 'web' as ChallengeCategory,
  difficulty: 'easy' as ChallengeDifficulty,
  base_score: 100,
  status: 'draft' as ChallengeStatus,
})

const { list, total, page, pageSize, loading, changePage, changePageSize, refresh } = usePagination(getChallenges)

function openDialog(row?: any) {
  if (row) {
    editingId.value = row.id
    Object.assign(form, {
      title: row.title,
      category: row.category,
      difficulty: row.difficulty,
      base_score: row.base_score,
      status: row.status,
    })
  } else {
    editingId.value = null
    Object.assign(form, {
      title: '',
      category: 'web',
      difficulty: 'easy',
      base_score: 100,
      status: 'draft',
    })
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.title) {
    toast.error('请填写标题')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await updateChallenge(editingId.value, form)
      toast.success('更新成功')
    } else {
      await createChallenge(form)
      toast.success('创建成功')
    }
    dialogVisible.value = false
    refresh()
  } catch (error) {
    toast.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await ElMessageBox.confirm('确定要删除此挑战吗？', '确认', { type: 'warning' })
    await deleteChallenge(id)
    toast.success('删除成功')
    refresh()
  } catch (error) {
    if (error !== 'cancel') {
      toast.error('删除失败')
    }
  }
}

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
  return { web: '#3b82f6', pwn: '#ef4444', reverse: '#8b5cf6', crypto: '#f59e0b', misc: '#10b981', forensics: '#06b6d4' }[category]
}

function getDifficultyLabel(difficulty: ChallengeDifficulty): string {
  const labels: Record<ChallengeDifficulty, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    hell: '地狱',
  }
  return labels[difficulty]
}

function getDifficultyColor(difficulty: ChallengeDifficulty): string {
  return { beginner: '#10b981', easy: '#3b82f6', medium: '#f59e0b', hard: '#ef4444', hell: '#7c3aed' }[difficulty]
}

function getStatusLabel(status: ChallengeStatus): string {
  return { draft: '草稿', review: '审核中', active: '已发布', archived: '已归档' }[status]
}

function getStatusColor(status: ChallengeStatus): string {
  return { draft: '#8b949e', review: '#f59e0b', active: '#10b981', archived: '#6e7681' }[status]
}

onMounted(() => {
  refresh()
})
</script>
