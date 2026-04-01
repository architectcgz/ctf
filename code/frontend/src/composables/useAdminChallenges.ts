import { onMounted } from 'vue'
import { ElMessageBox } from 'element-plus'

import { deleteChallenge, getChallenges, publishChallenge } from '@/api/admin'
import type { AdminChallengeListItem } from '@/api/contracts'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'

export function useAdminChallenges() {
  const toast = useToast()
  const pagination = usePagination(getChallenges)

  async function publish(row: AdminChallengeListItem) {
    try {
      await publishChallenge(row.id)
      toast.success('发布成功')
      await pagination.refresh()
    } catch {
      toast.error('发布失败，请先确认题目包、镜像和 Flag 已准备完成')
    }
  }

  async function remove(id: string) {
    try {
      await ElMessageBox.confirm('确定要删除此挑战吗？', '确认', { type: 'warning' })
      await deleteChallenge(id)
      toast.success('删除成功')
      await pagination.refresh()
    } catch (error) {
      if (error !== 'cancel') {
        toast.error('删除失败')
      }
    }
  }

  onMounted(() => {
    void pagination.refresh()
  })

  return {
    ...pagination,
    publish,
    remove,
  }
}
