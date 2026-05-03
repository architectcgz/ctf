<script setup lang="ts">
import { useTeacherStudentReviewArchivePage } from '@/features/teacher-student-review-archive'
import { TeacherReviewArchiveWorkspace } from '@/widgets/teacher-review-archive'

const {
  archive,
  loading,
  error,
  reload,
  exporting,
  openStudentAnalysis,
  goBack,
  exportArchive,
} = useTeacherStudentReviewArchivePage()
</script>

<template>
  <TeacherReviewArchiveWorkspace
    :archive="archive"
    :loading="loading"
    :error="error"
    :exporting="exporting"
    @reload="reload"
    @back="goBack"
    @open-analysis="openStudentAnalysis"
    @export-archive="exportArchive"
  />
</template>
