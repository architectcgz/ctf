<script setup lang="ts">
import { useStudentReviewArchivePage } from '@/features/student-review-archive-workspace'
import { ReviewArchiveWorkspace } from '@/widgets/review-archive-workspace'

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
  <ReviewArchiveWorkspace
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
