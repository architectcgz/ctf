<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">镜像管理</h1>
      <button class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[var(--color-primary)]/90" @click="dialogVisible = true">
        创建镜像
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[var(--color-border-default)]">
      <table class="w-full">
        <thead class="bg-[var(--color-bg-surface)]">
          <tr>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">镜像名称</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">标签</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">状态</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">创建时间</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[var(--color-text-primary)]">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#30363d]">
          <tr v-for="row in list" :key="row.id" class="transition-colors hover:bg-[var(--color-bg-elevated)]">
            <td class="px-4 py-3 font-mono text-sm text-[var(--color-text-primary)]">{{ row.name }}</td>
            <td class="px-4 py-3 font-mono text-sm text-[var(--color-text-secondary)]">{{ row.tag }}</td>
            <td class="px-4 py-3">
              <span class="rounded px-2 py-1 text-xs font-medium" :style="{ backgroundColor: getStatusColor(row.status) + '20', color: getStatusColor(row.status) }">
                {{ getStatusLabel(row.status) }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-secondary)]">{{ new Date(row.created_at).toLocaleString() }}</td>
            <td class="px-4 py-3">
              <button class="rounded bg-red-500/20 px-3 py-1 text-xs text-red-500 transition-colors hover:bg-red-500/30" @click="handleDelete(row.id)">
                删除
              </button>
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

    <ElDialog v-model="dialogVisible" title="创建镜像" width="500px">
      <ElForm :model="form" label-width="100px">
        <ElFormItem label="镜像名称" required>
          <ElInput v-model="form.name" placeholder="例如：ubuntu" />
        </ElFormItem>
        <ElFormItem label="标签" required>
          <ElInput v-model="form.tag" placeholder="例如：22.04" />
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="镜像说明（可选）" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <button class="rounded-lg border border-[var(--color-border-default)] px-4 py-2 text-sm text-[var(--color-text-primary)] transition-colors hover:bg-[#21262d]" @click="dialogVisible = false">
          取消
        </button>
        <button
          :disabled="creating"
          class="ml-2 rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm text-white transition-colors hover:bg-[var(--color-primary)]/90 disabled:cursor-not-allowed disabled:opacity-50"
          @click="handleCreate"
        >
          {{ creating ? '创建中...' : '创建' }}
        </button>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import { getImages, createImage, deleteImage } from '@/api/admin'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import type { ImageStatus } from '@/api/contracts'

const toast = useToast()
const dialogVisible = ref(false)
const creating = ref(false)
const form = reactive({
  name: '',
  tag: '',
  description: '',
})

const { list, total, page, pageSize, loading, changePage, changePageSize, refresh } = usePagination(getImages)

let pollTimer: number | null = null

async function handleCreate() {
  if (!form.name || !form.tag) {
    toast.error('请填写完整信息')
    return
  }
  creating.value = true
  try {
    await createImage(form)
    toast.success('镜像创建成功')
    dialogVisible.value = false
    Object.assign(form, { name: '', tag: '', description: '' })
    refresh()
  } catch (error) {
    toast.error('创建失败')
  } finally {
    creating.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await ElMessageBox.confirm('确定要删除此镜像吗？', '确认', { type: 'warning' })
    await deleteImage(id)
    toast.success('删除成功')
    refresh()
  } catch (error) {
    if (error !== 'cancel') {
      toast.error('删除失败')
    }
  }
}

function getStatusLabel(status: ImageStatus): string {
  return { pending: '等待中', building: '构建中', available: '可用', failed: '失败' }[status]
}

function getStatusColor(status: ImageStatus): string {
  return { pending: '#8b949e', building: '#f59e0b', available: '#10b981', failed: '#ef4444' }[status]
}

onMounted(() => {
  refresh()
  pollTimer = window.setInterval(refresh, 10000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>
