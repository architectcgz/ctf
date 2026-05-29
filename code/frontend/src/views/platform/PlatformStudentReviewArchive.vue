<script setup lang="ts">
import { toRef } from 'vue'

import { useStudentReviewArchivePage } from '@/features/student-review-archive-workspace'
import { ReviewArchiveWorkspace } from '@/widgets/review-archive-workspace'

const props = defineProps<{
  className: string
  studentId: string
}>()

const {
  archive,
  loading,
  error,
  reload,
  exporting,
  analysisRoute,
  backRoute,
  exportArchive,
} = useStudentReviewArchivePage({
  className: toRef(props, 'className'),
  studentId: toRef(props, 'studentId'),
})
</script>

<template>
  <ReviewArchiveWorkspace
    :archive="archive"
    :loading="loading"
    :error="error"
    :exporting="exporting"
    :analysis-route="analysisRoute"
    :back-route="backRoute"
    @reload="reload"
    @export-archive="exportArchive"
  />
</template>
