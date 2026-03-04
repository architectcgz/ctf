<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">挑战管理</h1>
      <ElButton type="primary" @click="openDialog()">创建挑战</ElButton>
    </div>

    <ElTable v-loading="loading" :data="list" stripe>
      <ElTableColumn prop="title" label="标题" />
      <ElTableColumn prop="category" label="分类">
        <template #default="{ row }">
          <ElTag size="small">{{ getCategoryLabel(row.category) }}</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="difficulty" label="难度">
        <template #default="{ row }">
          <ElTag :type="getDifficultyColor(row.difficulty)" size="small">{{ getDifficultyLabel(row.difficulty) }}</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="base_score" label="分值" width="80" />
      <ElTableColumn prop="status" label="状态">
        <template #default="{ row }">
          <ElTag :type="getStatusColor(row.status)" size="small">{{ getStatusLabel(row.status) }}</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" width="180">
        <template #default="{ row }">
          <ElButton size="small" @click="openDialog(row)">编辑</ElButton>
          <ElButton size="small" type="danger" @click="handleDelete(row.id)">删除</ElButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <ElPagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      layout="total, prev, pager, next"
      @current-change="changePage"
      @size-change="changePageSize"
    />

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
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
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
  const colors: Record<ChallengeDifficulty, string> = {
    beginner: 'info',
    easy: 'success',
    medium: 'warning',
    hard: 'danger',
    hell: 'danger',
  }
  return colors[difficulty]
}

function getStatusLabel(status: ChallengeStatus): string {
  const labels: Record<ChallengeStatus, string> = {
    draft: '草稿',
    review: '审核中',
    active: '已发布',
    archived: '已归档',
  }
  return labels[status]
}

function getStatusColor(status: ChallengeStatus): string {
  const colors: Record<ChallengeStatus, string> = {
    draft: 'info',
    review: 'warning',
    active: 'success',
    archived: '',
  }
  return colors[status]
}

onMounted(() => {
  refresh()
})
</script>


