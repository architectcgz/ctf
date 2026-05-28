import {
  exportPlatformAWDReviewArchive,
  exportPlatformAWDReviewReport,
  getPlatformAWDReview,
  listPlatformAWDReviews,
} from '@/api/admin'
import type {
  AwdReviewContestItemData,
  ReportExportData,
  TeacherAWDReviewArchiveData,
} from '@/api/contracts'
import {
  exportTeacherAWDReviewArchive,
  exportTeacherAWDReviewReport,
  getTeacherAWDReview,
  listTeacherAWDReviews,
} from '@/api/teacher'
import type { UserRole } from '@/utils/constants'

type AwdReviewAccessRole = UserRole | null | undefined

interface ListAwdReviewsParams {
  status?: AwdReviewContestItemData['status']
  keyword?: string
  page?: number
  page_size?: number
}

interface GetAwdReviewParams {
  round?: number
  team_id?: string
}

interface ExportAwdReviewParams {
  round_number?: number
}

export async function listAwdReviewsByRole(
  role: AwdReviewAccessRole,
  params?: ListAwdReviewsParams,
  options?: { signal?: AbortSignal }
) {
  return role === 'admin'
    ? listPlatformAWDReviews(params, options)
    : listTeacherAWDReviews(params, options)
}

export async function getAwdReviewByRole(
  role: AwdReviewAccessRole,
  contestId: string,
  params?: GetAwdReviewParams
): Promise<TeacherAWDReviewArchiveData> {
  return role === 'admin'
    ? getPlatformAWDReview(contestId, params)
    : getTeacherAWDReview(contestId, params)
}

export async function exportAwdReviewArchiveByRole(
  role: AwdReviewAccessRole,
  contestId: string,
  data?: ExportAwdReviewParams
): Promise<ReportExportData> {
  return role === 'admin'
    ? exportPlatformAWDReviewArchive(contestId, data)
    : exportTeacherAWDReviewArchive(contestId, data)
}

export async function exportAwdReviewReportByRole(
  role: AwdReviewAccessRole,
  contestId: string,
  data?: ExportAwdReviewParams
): Promise<ReportExportData> {
  return role === 'admin'
    ? exportPlatformAWDReviewReport(contestId, data)
    : exportTeacherAWDReviewReport(contestId, data)
}
