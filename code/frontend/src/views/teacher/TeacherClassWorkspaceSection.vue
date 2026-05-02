<script setup lang="ts">
import { useTeacherClassWorkspaceSection } from '@/features/teacher-class-workspace'
import TeacherClassStudents from '@/views/teacher/TeacherClassStudents.vue'

useTeacherClassWorkspaceSection()
</script>

<template>
  <TeacherClassStudents />
</template>
