<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[#c9d1d9]">镜像管理</h1>
      <button class="rounded-lg bg-[#0891b2] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[#0891b2]/90" @click="dialogVisible = true">
        创建镜像
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[#30363d]">
      <table class="w-full">
        <thead class="bg-[#161b22]">
          <tr>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">镜像名称</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">标签</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">状态</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">创建时间</th>
            <th class="px-4 py-3 text-left text-sm font-medium text-[#c9d1d9]">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#30363d]">
          <tr v-for="row in list" :key="row.id" class="transition-colors hover:bg-[#1c2128]">
            <td class="px-4 py-3 font-mono text-sm text-[#c9d1d9]">{{ row.name }}</td>
            <td class="px-4 py-3 font-mono text-sm text-[#8b949e]">{{ row.tag }}</td>
            <td class="px-4 py-3">
              <span class="rounded px-2 py-1 text-xs font-medium" :style="{ backgroundColor: getStatusColor(row.status) + '20', color: getStatusColor(row.status) }">
                {{ getStatusLabel(row.status) }}
              </span>
            </td>
            <td class="px-4 py-3 text-sm text-[#8b949e]">{{ new Date(row.created_at).toLocaleString() }}</td>
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

    <ElDialog v-model="dialogVisible" title="创建镜像" width="500px">
      <ElForm :model="form" label-width="100px">
        <ElFormItem label="镜像名称" required>
          <ElInput v-model="form.name" placeholder="例如：ubuntu" />
        </ElFormItem>
        <ElFormItem label="标签" required>
          <ElInput v-model="form.tag" placeholder="例如：22.04" />
        </ElFormItem>
        <ElFormItem label="来源类型" required>
          <ElSelect v-model="form.source_type" placeholder="选择来源类型">
            <ElOption label="镜像仓库" value="registry" />
            <ElOption label="Dockerfile" value="dockerfile" />
            <ElOption label="上传" value="upload" />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <button class="rounded-lg border border-[#30363d] px-4 py-2 text-sm text-[#c9d1d9] transition-colors hover:bg-[#21262d]" @click="dialogVisible = false">
          取消
        </button>
        <button
          :disabled="creating"
          class="ml-2 rounded-lg bg-[#0891b2] px-4 py-2 text-sm text-white transition-colors hover:bg-[#0891b2]/90 disabled:cursor-not-allowed disabled:opacity-50"
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
  source_type: 'registry' as 'registry' | 'dockerfile' | 'upload',
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
    Object.assign(form, { name: '', tag: '', source_type: 'registry' })
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
  return { pending: '等待中', building: '构建中', ready: '就绪', failed: '失败', deprecated: '已弃用' }[status]
}

function getStatusColor(status: ImageStatus): string {
  return { pending: '#8b949e', building: '#f59e0b', ready: '#10b981', failed: '#ef4444', deprecated: '#6e7681' }[status]
}

onMounted(() => {
  refresh()
  pollTimer = window.setInterval(refresh, 10000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>
