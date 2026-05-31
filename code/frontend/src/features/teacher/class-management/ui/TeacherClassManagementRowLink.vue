<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'
import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'

interface TeacherClassRouteTarget {
  name: string
  params: {
    className: string
  }
}

const props = defineProps<{
  className: string
  studentCount: number
  buildClassRoute: (className: string) => TeacherClassRouteTarget
}>()
</script>

<template>
  <AppRouteLink
    :to="props.buildClassRoute(className)"
    class="ui-btn ui-btn--primary ui-btn--xs"
    :aria-label="`${className}，${studentCount} 名学生，进入班级`"
  >
    进入班级
    <ArrowRight class="h-4 w-4" />
  </AppRouteLink>
</template>
