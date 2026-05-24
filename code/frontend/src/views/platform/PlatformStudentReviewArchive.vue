<script setup lang="ts">
import { useStudentReviewArchivePage } from '@/features/student-review-archive-workspace'
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
} = useStudentReviewArchivePage()
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
