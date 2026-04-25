<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import type { TeacherInstanceItem } from '@/api/contracts'
import { destroyTeacherInstance, getTeacherInstances } from '@/api/teacher'
import InstanceManageHeroPanel from '@/components/platform/instance/InstanceManageHeroPanel.vue'
import InstanceManageWorkspacePanel from '@/components/platform/instance/InstanceManageWorkspacePanel.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'

interface InstanceManageTableRow {
  id: string
  challenge: string
  user: string
  user_meta: string
  ip_address: string
  status: string
  status_label: string
  created_at: string
  actions: string
}

const router = useRouter()
const list = ref<TeacherInstanceItem[]>([])
const page = ref(1)
const pageSize = ref(15)
const loading = ref(false)
const destroyingId = ref('')
const error = ref<string | null>(null)

const total = computed(() => list.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const pageRows = computed<InstanceManageTableRow[]>(() => {
  const start = (page.value - 1) * pageSize.value
  return list.value.slice(start, start + pageSize.value).map((item) => ({
    id: item.id,
    challenge: item.challenge_title,
    user: item.student_name || item.student_username,
    user_meta: `${item.student_username} · ${item.class_name}`,
    ip_address: item.access_url || '暂未分配',
    status: item.status,
    status_label: formatStatus(item.status),
    created_at: formatDateTime(item.created_at),
    actions: '销毁',
  }))
})
const runningCount = computed(() => list.value.filter((item) => item.status === 'running').length)
const warningCount = computed(
  () => list.value.filter((item) => item.status !== 'running' || item.remaining_time <= 600).length
)

function formatStatus(status: string): string {
  switch (status) {
    case 'running':
      return '运行中'
    case 'creating':
      return '创建中'
    case 'expired':
      return '已过期'
    case 'failed':
      return '异常'
    default:
      return status
  }
}

function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '--'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

async function loadInstances(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    list.value = await getTeacherInstances({
      class_name: undefined,
      keyword: undefined,
      student_no: undefined,
    })
    if (page.value > totalPages.value) {
      page.value = totalPages.value
    }
  } catch (err) {
    console.error('加载实例列表失败:', err)
    error.value = '加载实例列表失败，请稍后重试'
    list.value = []
  } finally {
    loading.value = false
  }
}

async function handleDestroyInstance(instance: TeacherInstanceItem): Promise<void> {
  const confirmed = await confirmDestructiveAction({
    title: '强制销毁实例',
    message: `您确定要强制销毁实例 ${instance.id} 吗？此操作不可逆，用户当前的运行状态将丢失。`,
    confirmButtonText: '强制销毁',
    cancelButtonText: '取消',
  })

  if (!confirmed) return

  try {
    destroyingId.value = instance.id
    await destroyTeacherInstance(instance.id)
    list.value = list.value.filter((item) => item.id !== instance.id)
    if (page.value > totalPages.value) {
      page.value = totalPages.value
    }
  } catch (err) {
    console.error('销毁实例失败:', err)
    error.value = '销毁实例失败，请稍后重试'
  } finally {
    destroyingId.value = ''
  }
}

function requestDestroyById(id: string): void {
  const instance = list.value.find((item) => item.id === id)
  if (!instance) {
    return
  }

  void handleDestroyInstance(instance)
}

function handlePageChange(p: number): void {
  page.value = p
}

onMounted(() => {
  void loadInstances()
})
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero admin-instance-manage-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <InstanceManageHeroPanel
          :running-count="runningCount"
          :total="total"
          :warning-count="warningCount"
          @back="void router.push({ name: 'PlatformOverview' })"
          @refresh="void loadInstances()"
        />

        <InstanceManageWorkspacePanel
          :loading="loading"
          :has-instances="list.length > 0"
          :rows="pageRows"
          :page="page"
          :total-pages="totalPages"
          :total="total"
          :destroying-id="destroyingId"
          :error="error"
          @destroy-instance="requestDestroyById"
          @change-page="handlePageChange"
        />
      </main>
    </div>
  </div>
</template>

<style scoped>
</style>
