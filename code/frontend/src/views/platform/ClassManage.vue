<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import ClassManageHeroPanel from '@/components/platform/class/ClassManageHeroPanel.vue'
import ClassManageWorkspacePanel from '@/components/platform/class/ClassManageWorkspacePanel.vue'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

const router = useRouter()
const list = ref<TeacherClassItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)
const loading = ref(false)
const error = ref<string | null>(null)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))

async function loadClasses(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await getClasses({
      page: page.value,
      page_size: pageSize.value,
    })
    if (Array.isArray(data)) {
      list.value = data
      total.value = data.length
      return
    }

    list.value = data.list
    total.value = data.total
    page.value = data.page
    pageSize.value = data.page_size
  } catch (err) {
    console.error('加载班级列表失败:', err)
    error.value = '加载班级列表失败，请稍后重试'
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number): void {
  const normalizedPage = Math.max(1, Math.floor(p))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }

  page.value = normalizedPage
  void loadClasses()
}

function openClass(className: string): void {
  void router.push({
    name: 'PlatformClassStudents',
    params: { className },
  })
}

const totalStudents = computed(() =>
  list.value.reduce((sum, item) => sum + (item.student_count || 0), 0)
)

const rows = computed(() =>
  list.value.map((item, index) => ({
    id: item.name,
    name: item.name,
    student_count: item.student_count || 0,
    teacher_name: '--',
    created_at: '--',
    actions: '查看班级',
    rowIndex: index,
  }))
)

onMounted(() => {
  void loadClasses()
})
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero admin-class-manage-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <ClassManageHeroPanel
          :total="total"
          :total-students="totalStudents"
          @refresh="void loadClasses()"
        />

        <ClassManageWorkspacePanel
          :loading="loading"
          :has-classes="list.length > 0"
          :rows="rows"
          :page="page"
          :total-pages="totalPages"
          :total="total"
          :error="error"
          @open-class="openClass"
          @change-page="handlePageChange"
        />
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-class-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}
</style>
