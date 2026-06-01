import { request } from '../request'

import type {
  AdminContestAnnouncementCreatePayload,
  RawContestAnnouncement,
  ContestAnnouncement,
} from './contest-support'

// Re-export types for barrel consumers
export type { AdminContestAnnouncementCreatePayload } from './contest-support'

import { normalizeContestAnnouncement } from './contest-support'

export async function getAdminContestAnnouncements(
  contestId: string
): Promise<ContestAnnouncement[]> {
  const response = await request<RawContestAnnouncement[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/announcements`,
  })

  return response.map(normalizeContestAnnouncement)
}

export async function createAdminContestAnnouncement(
  contestId: string,
  data: AdminContestAnnouncementCreatePayload
): Promise<ContestAnnouncement> {
  const response = await request<RawContestAnnouncement>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/announcements`,
    data: {
      title: data.title,
      content: data.content,
    },
  })

  return normalizeContestAnnouncement(response)
}

export async function deleteAdminContestAnnouncement(
  contestId: string,
  announcementId: string
): Promise<void> {
  await request<void>({
    method: 'DELETE',
    url: `/admin/contests/${encodeURIComponent(contestId)}/announcements/${encodeURIComponent(announcementId)}`,
  })
}
