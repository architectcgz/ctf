<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppLoading from '@/components/common/AppLoading.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const contestId = computed(() => String(route.params.contestId ?? ''))

watch(
  contestId,
  (nextContestId) => {
    if (!nextContestId) {
      return
    }

    if (authStore.user?.role === 'admin') {
      void router.replace({
        name: 'AdminAwdReplay',
        params: { id: nextContestId },
      })
      return
    }

    void router.replace({
      name: 'TeacherAwdOverview',
      params: { contestId: nextContestId },
    })
  },
  { immediate: true }
)
</script>

<template>
  <div class="teacher-management-shell teacher-surface flex min-h-full flex-1 items-center justify-center">
    <AppLoading>正在进入 AWD 复盘工作台...</AppLoading>
  </div>
</template>
