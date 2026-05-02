import { runContestAWDCheckerPreview } from '@/api/admin/contests'
import type { AWDCheckerPreviewData, AWDCheckerType } from '@/api/contracts'

interface RunAwdCheckerPreviewOptions {
  contestId: string
  serviceId?: number
  awdChallengeId: number
  checkerType: AWDCheckerType
  checkerConfig: Record<string, unknown>
  accessURL?: string
  previewFlag?: string
  previewRequestId: string
}

export async function runAwdCheckerPreview(
  options: RunAwdCheckerPreviewOptions
): Promise<AWDCheckerPreviewData> {
  return runContestAWDCheckerPreview(options.contestId, {
    ...(options.serviceId && options.serviceId > 0 ? { service_id: options.serviceId } : {}),
    awd_challenge_id: options.awdChallengeId,
    checker_type: options.checkerType,
    checker_config: options.checkerConfig,
    ...(options.accessURL ? { access_url: options.accessURL } : {}),
    preview_flag: options.previewFlag?.trim() || undefined,
    preview_request_id: options.previewRequestId,
  })
}
