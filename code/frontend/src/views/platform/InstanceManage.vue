<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { TeacherInstanceItem } from '@/api/contracts'
import { destroyTeacherInstance, getTeacherInstances } from '@/api/teacher'
import InstanceManageHeroPanel from '@/components/platform/instance/InstanceManageHeroPanel.vue'
import InstanceManageWorkspacePanel from '@/components/platform/instance/InstanceManageWorkspacePanel.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'

interface InstanceManageTableRow {
  id: string
  challenge: string
  student_id: string
  user: string
  username: string
  class_name: string
  ip_address: string
  status: string
  status_label: string
  created_at: string
  actions: string
}

type InstanceStatusFilter = 'running' | 'creating' | 'expired' | 'failed' | 'inactive' | ''

const router = useRouter()
const list = ref<TeacherInstanceItem[]>([])
const page = ref(1)
const pageSize = ref(15)
const loading = ref(false)
const destroyingId = ref('')
const error = ref<string | null>(null)
const keyword = ref('')
const statusFilter = ref<InstanceStatusFilter>('')

const totalInstances = computed(() => list.value.length)
const filteredInstances = computed(() => {
  const query = keyword.value.trim().toLowerCase()

  return list.value.filter((item) => {
    const statusGroup: Exclude<InstanceStatusFilter, ''> =
      item.status === 'running' || item.status === 'creating' || item.status === 'expired' || item.status === 'failed'
        ? item.status
        : 'inactive'
    const searchableText = [
      item.id,
      item.challenge_title,
      item.student_name,
      item.student_username,
      item.student_no,
      item.class_name,
      item.access_url,
      item.status,
    ]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()

    const matchesKeyword = !query || searchableText.includes(query)
    const matchesStatus = !statusFilter.value || statusGroup === statusFilter.value
    return matchesKeyword && matchesStatus
  })
})
const filteredTotal = computed(() => filteredInstances.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(filteredTotal.value / pageSize.value)))
const pageRows = computed<InstanceManageTableRow[]>(() => {
  const start = (page.value - 1) * pageSize.value
  return filteredInstances.value.slice(start, start + pageSize.value).map((item) => ({
    id: item.id,
    challenge: item.challenge_title,
    student_id: String(item.student_id),
    user: item.student_name || item.student_username,
    username: item.student_username,
    class_name: item.class_name,
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

function openStudent(studentId: string, className: string): void {
  void router.push({
    name: 'PlatformStudentAnalysis',
    params: { className, studentId },
  })
}

function handlePageChange(p: number): void {
  const normalizedPage = Math.max(1, Math.floor(p))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }

  page.value = normalizedPage
}

function resetFilters(): void {
  keyword.value = ''
  statusFilter.value = ''
}

watch([keyword, statusFilter], () => {
  page.value = 1
})

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
          :total="totalInstances"
          :warning-count="warningCount"
          @back="void router.push({ name: 'PlatformOverview' })"
          @refresh="void loadInstances()"
        />

        <InstanceManageWorkspacePanel
          :loading="loading"
          :has-instances="list.length > 0"
          :rows="pageRows"
          :keyword="keyword"
          :status-filter="statusFilter"
          :page="page"
          :total-pages="totalPages"
          :total="filteredTotal"
          :destroying-id="destroyingId"
          :error="error"
          @update:keyword="keyword = $event"
          @change:status-filter="statusFilter = $event"
          @reset-filters="resetFilters"
          @open-student="openStudent"
          @destroy-instance="requestDestroyById"
          @change-page="handlePageChange"
        />
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-instance-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}
</style>
