import {
  exportAwdReviewArchive,
  exportAwdReviewReport,
  getAwdReview,
  listAwdReviews,
} from '../teaching/awd-reviews'

import type {
  AwdReviewArchiveData,
  AwdReviewContestItemData,
  ReportExportData,
} from '../contracts'

export async function listPlatformAWDReviews(
  params?: {
    status?: AwdReviewContestItemData['status']
    keyword?: string
    page?: number
    page_size?: number
  },
  options?: { signal?: AbortSignal }
) {
  return listAwdReviews(params, options)
}

export async function getPlatformAWDReview(
  contestId: string,
  params?: {
    round?: number
    team_id?: string
  }
): Promise<AwdReviewArchiveData> {
  return getAwdReview(contestId, params)
}

export async function exportPlatformAWDReviewArchive(
  contestId: string,
  data?: { round_number?: number }
): Promise<ReportExportData> {
  return exportAwdReviewArchive(contestId, data)
}

export async function exportPlatformAWDReviewReport(
  contestId: string,
  data?: { round_number?: number }
): Promise<ReportExportData> {
  return exportAwdReviewReport(contestId, data)
}
