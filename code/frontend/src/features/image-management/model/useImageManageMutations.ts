import { ref, type Ref } from 'vue'

import { createImage, deleteImage } from '@/api/admin/authoring'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'
import type { ImageCreateForm } from '@/entities/image'
import { createEmptyImageCreateForm } from '@/entities/image'

interface UseImageManageMutationsOptions {
  form: ImageCreateForm
  dialogVisible: Ref<boolean>
  refresh: () => Promise<void>
}

export function useImageManageMutations(options: UseImageManageMutationsOptions) {
  const { form, dialogVisible, refresh } = options
  const toast = useToast()

  const creating = ref(false)

  async function handleCreate() {
    if (creating.value) {
      return
    }

    if (!form.name || !form.tag) {
      toast.error('请填写完整信息')
      return
    }

    creating.value = true
    try {
      await createImage(form)
      toast.success('镜像创建成功')
      dialogVisible.value = false
      Object.assign(form, createEmptyImageCreateForm())
      await refresh()
    } catch {
      toast.error('创建失败')
    } finally {
      creating.value = false
    }
  }

  async function handleDelete(id: string) {
    const confirmed = await confirmDestructiveAction({
      message: '确定要删除此镜像吗？',
    })
    if (!confirmed) {
      return
    }

    try {
      await deleteImage(id)
      toast.success('删除成功')
      await refresh()
    } catch (error) {
      const message = error instanceof Error && error.message.trim() ? error.message : '删除失败'
      toast.error(message)
    }
  }

  return {
    creating,
    handleCreate,
    handleDelete,
  }
}
