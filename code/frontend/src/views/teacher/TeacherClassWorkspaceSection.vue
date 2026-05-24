<script setup lang="ts">
import { useClassWorkspaceSection } from '@/features/class-workspace-redirect'
import TeacherClassStudents from '@/views/teacher/TeacherClassStudents.vue'

useClassWorkspaceSection()
</script>

<template>
  <TeacherClassStudents />
</template>
