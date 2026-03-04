<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">挑战管理</h1>
      <button class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[var(--color-primary)]/90" @click="openDialog()">
        创建挑战
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[var(--color-border-default)]">
      <table class="w-full">
        <thead class="bg-[var(--color-bg-surface)]">
          <tr>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">标题</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">分类</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">难度</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">分值</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">状态</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#30363d]">
          <tr v-for="row in list" :key="row.id" class="transition-colors hover:bg-[var(--color-bg-elevated)]">
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ row.title }}</td>
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
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ row.base_score }}</td>
            <td class="px-4 py-3">
              <span class="rounded px-2 py-1 text-xs font-medium" :style="{ backgroundColor: getStatusColor(row.status) + '20', color: getStatusColor(row.status) }">
                {{ getStatusLabel(row.status) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div class="flex gap-2">
                <button class="rounded bg-[var(--color-primary)] px-3 py-1 text-xs text-white transition-colors hover:bg-[var(--color-primary)]/90" @click="$router.push(`/admin/challenges/${row.id}`)">
                  查看
                </button>
                <button class="rounded bg-[var(--color-primary)] px-3 py-1 text-xs text-white transition-colors hover:bg-[var(--color-primary)]/90" @click="openDialog(row)">
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
        <ElFormItem label="描述">
          <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="靶机描述" />
        </ElFormItem>
        <ElFormItem label="镜像">
          <ElSelect v-model="form.image_id" placeholder="选择镜像" clearable>
            <ElOption v-for="img in images" :key="img.id" :label="`${img.name}:${img.tag}`" :value="img.id" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Flag">
          <ElInput v-model="form.flag" placeholder="静态 Flag 或动态 Flag 模板" />
        </ElFormItem>
        <ElFormItem label="标签">
          <ElInput v-model="form.tags" placeholder="用逗号分隔，如：SQL注入,WAF绕过" />
        </ElFormItem>
        <ElFormItem label="提示">
          <ElInput v-model="form.hints" type="textarea" :rows="2" placeholder="每行一条提示" />
        </ElFormItem>
        <ElFormItem label="CPU 限制">
          <ElInputNumber v-model="form.cpu" :min="0.1" :max="4" :step="0.1" :precision="1" />
          <span class="ml-2 text-xs text-[var(--color-text-secondary)]">核数</span>
        </ElFormItem>
        <ElFormItem label="内存限制">
          <ElInputNumber v-model="form.memory" :min="128" :max="4096" :step="128" />
          <span class="ml-2 text-xs text-[var(--color-text-secondary)]">MB</span>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <button class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[#21262d]" @click="dialogVisible = false">
          取消
        </button>
        <button
          :disabled="saving"
          class="ml-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary)]/90 disabled:cursor-not-allowed disabled:opacity-50"
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
import { getChallenges, createChallenge, updateChallenge, deleteChallenge, getImages } from '@/api/admin'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import type { ChallengeCategory, ChallengeDifficulty, ChallengeStatus, AdminImageListItem } from '@/api/contracts'

const toast = useToast()
const dialogVisible = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)
const images = ref<AdminImageListItem[]>([])
const form = reactive({
  title: '',
  category: 'web' as ChallengeCategory,
  difficulty: 'easy' as ChallengeDifficulty,
  base_score: 100,
  status: 'draft' as ChallengeStatus,
  description: '',
  image_id: '',
  flag: '',
  tags: '',
  hints: '',
  cpu: 1,
  memory: 512,
})

const { list, total, page, pageSize, loading, changePage, changePageSize, refresh } = usePagination(getChallenges)

function openDialog(row?: AdminChallengeListItem) {
  if (row) {
    editingId.value = row.id
    Object.assign(form, {
      title: row.title,
      category: row.category,
      difficulty: row.difficulty,
      base_score: row.base_score,
      status: row.status,
      description: row.description || '',
      image_id: row.image_id || '',
      flag: row.flag || '',
      tags: row.tags?.join(',') || '',
      hints: row.hints?.join('\n') || '',
      cpu: row.resource_limits?.cpu || 1,
      memory: row.resource_limits?.memory || 512,
    })
  } else {
    editingId.value = null
    Object.assign(form, {
      title: '',
      category: 'web',
      difficulty: 'easy',
      base_score: 100,
      status: 'draft',
      description: '',
      image_id: '',
      flag: '',
      tags: '',
      hints: '',
      cpu: 1,
      memory: 512,
    })
  }
  dialogVisible.value = true
}

async function handleSave() {
  if (!form.title) {
    toast.error('请填写标题')
    return
  }

  // 发布前校验
  if (form.status === 'active') {
    if (!form.image_id) {
      toast.error('发布前必须选择镜像')
      return
    }
    if (!form.flag) {
      toast.error('发布前必须配置 Flag')
      return
    }
    if (!form.tags) {
      toast.error('发布前必须添加标签')
      return
    }
  }

  saving.value = true
  try {
    const data = {
      ...form,
      tags: form.tags ? form.tags.split(',').map(t => t.trim()).filter(Boolean) : [],
      hints: form.hints ? form.hints.split('\n').map(h => h.trim()).filter(Boolean) : [],
      resource_limits: { cpu: form.cpu, memory: form.memory },
    }

    if (editingId.value) {
      await updateChallenge(editingId.value, data)
      toast.success('更新成功')
    } else {
      await createChallenge(data)
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

async function loadImages() {
  try {
    const res = await getImages({ page: 1, page_size: 100 })
    images.value = res.list.filter(img => img.status === 'ready')
  } catch (error) {
    toast.error('加载镜像列表失败')
  }
}

onMounted(() => {
  refresh()
  loadImages()
})
</script>
