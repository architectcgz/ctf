<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">镜像管理</h1>
      <ElButton type="primary" @click="dialogVisible = true">创建镜像</ElButton>
    </div>

    <ElTable v-loading="loading" :data="list" stripe>
      <ElTableColumn prop="name" label="镜像名称" />
      <ElTableColumn prop="tag" label="标签" />
      <ElTableColumn prop="status" label="状态">
        <template #default="{ row }">
          <ElTag :type="getStatusColor(row.status)">{{ getStatusLabel(row.status) }}</ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="created_at" label="创建时间">
        <template #default="{ row }">
          {{ new Date(row.created_at).toLocaleString() }}
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" width="120">
        <template #default="{ row }">
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
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="creating" @click="handleCreate">创建</ElButton>
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
  const labels: Record<ImageStatus, string> = {
    pending: '等待中',
    building: '构建中',
    ready: '就绪',
    failed: '失败',
    deprecated: '已弃用',
  }
  return labels[status]
}

function getStatusColor(status: ImageStatus): string {
  const colors: Record<ImageStatus, string> = {
    pending: 'info',
    building: 'warning',
    ready: 'success',
    failed: 'danger',
    deprecated: 'info',
  }
  return colors[status]
}

onMounted(() => {
  refresh()
  pollTimer = window.setInterval(refresh, 10000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

